package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/infrastructure/crypto"
	"digital-signature-system/internal/infrastructure/pdf"
)

// Mock VerificationLogRepository
type MockVerificationLogRepository struct {
	mock.Mock
}

func (m *MockVerificationLogRepository) Create(ctx context.Context, log *entities.VerificationLog) error {
	args := m.Called(ctx, log)
	// Simulate GORM setting the ID
	if log.ID == "" {
		log.ID = "test-log-id"
	}
	return args.Error(0)
}

func (m *MockVerificationLogRepository) GetByDocumentID(ctx context.Context, docID string) ([]*entities.VerificationLog, error) {
	args := m.Called(ctx, docID)
	return args.Get(0).([]*entities.VerificationLog), args.Error(1)
}

func (m *MockVerificationLogRepository) GetByID(ctx context.Context, id string) (*entities.VerificationLog, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.VerificationLog), args.Error(1)
}

func TestVerificationService_GetVerificationInfo(t *testing.T) {
	tests := []struct {
		name          string
		documentID    string
		setupMocks    func(*MockDocumentRepository)
		expectedError string
	}{
		{
			name:       "successful verification info retrieval",
			documentID: "doc-123",
			setupMocks: func(docRepo *MockDocumentRepository) {
				document := &entities.Document{
					ID:        "doc-123",
					UserID:    "user-123",
					Filename:  "test.pdf",
					Issuer:    "John Doe",
					CreatedAt: time.Now(),
					FileSize:  1024,
					Status:    "active",
				}
				docRepo.On("GetByID", mock.Anything, "doc-123").Return(document, nil)
			},
			expectedError: "",
		},
		{
			name:       "document not found",
			documentID: "doc-123",
			setupMocks: func(docRepo *MockDocumentRepository) {
				docRepo.On("GetByID", mock.Anything, "doc-123").Return((*entities.Document)(nil), nil)
			},
			expectedError: "document not found",
		},
		{
			name:       "document not active",
			documentID: "doc-123",
			setupMocks: func(docRepo *MockDocumentRepository) {
				document := &entities.Document{
					ID:       "doc-123",
					Status:   "deleted",
					Filename: "test.pdf",
				}
				docRepo.On("GetByID", mock.Anything, "doc-123").Return(document, nil)
			},
			expectedError: "document is not active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockDocRepo := new(MockDocumentRepository)
			mockVerifyLogRepo := new(MockVerificationLogRepository)

			// Setup mocks
			tt.setupMocks(mockDocRepo)

			// Create service
			service := &VerificationService{
				documentRepo:        mockDocRepo,
				verificationLogRepo: mockVerifyLogRepo,
			}

			// Execute
			info, err := service.GetVerificationInfo(context.Background(), tt.documentID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, info)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, info)
				assert.Equal(t, tt.documentID, info.DocumentID)
			}

			// Verify mocks
			mockDocRepo.AssertExpectations(t)
		})
	}
}

func TestVerificationService_VerifyDocument(t *testing.T) {
	// Test data
	testHash := []byte("test-hash")
	testHashB64 := base64.StdEncoding.EncodeToString(testHash)
	testSignature := &crypto.SignatureData{
		Signature: []byte("test-signature"),
		Hash:      testHash,
		Algorithm: "RSA-PSS-SHA256",
	}
	testSignatureJSON := `{"algorithm":"RSA-PSS-SHA256","hash":"dGVzdC1oYXNo","signature":"dGVzdC1zaWduYXR1cmU="}`
	testQRCodeData := pdf.QRCodeData{
		DocID:     "doc-123",
		Hash:      testHashB64,
		Signature: testSignatureJSON,
		Timestamp: time.Now().Unix(),
	}
	testQRCodeJSON, _ := json.Marshal(testQRCodeData)

	tests := []struct {
		name           string
		request        *VerificationRequest
		setupMocks     func(*MockDocumentRepository, *MockVerificationLogRepository, *MockSignatureService, *MockPDFService, *MockDocumentService)
		expectedStatus string
		expectedValid  bool
	}{
		{
			name: "valid document verification",
			request: &VerificationRequest{
				DocumentID: "doc-123",
				PDFData:    []byte("%PDF-1.4 test content"),
				VerifierIP: "127.0.0.1",
			},
			setupMocks: func(docRepo *MockDocumentRepository, logRepo *MockVerificationLogRepository, sigService *MockSignatureService, pdfService *MockPDFService, docService *MockDocumentService) {
				// Document exists and is active
				document := &entities.Document{
					ID:            "doc-123",
					DocumentHash:  testHashB64,
					SignatureData: testSignatureJSON,
					QRCodeData:    string(testQRCodeJSON),
					Status:        "active",
				}
				docRepo.On("GetByID", mock.Anything, "doc-123").Return(document, nil)

				// PDF validation and hash calculation
				pdfService.On("ValidatePDF", mock.AnythingOfType("[]uint8")).Return(nil)
				pdfService.On("CalculateHash", mock.AnythingOfType("[]uint8")).Return(testHash, nil)

				// Document service methods
				docService.On("DecodeSignatureData", testSignatureJSON).Return(testSignature, nil)

				// Signature verification
				sigService.On("VerifySignature", testHash, testSignature).Return(nil)

				// Verification logging
				logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.VerificationLog")).Return(nil)
			},
			expectedStatus: StatusValid,
			expectedValid:  true,
		},
		{
			name: "document not found",
			request: &VerificationRequest{
				DocumentID: "doc-123",
				PDFData:    []byte("%PDF-1.4 test content"),
				VerifierIP: "127.0.0.1",
			},
			setupMocks: func(docRepo *MockDocumentRepository, logRepo *MockVerificationLogRepository, sigService *MockSignatureService, pdfService *MockPDFService, docService *MockDocumentService) {
				docRepo.On("GetByID", mock.Anything, "doc-123").Return((*entities.Document)(nil), nil)
				logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.VerificationLog")).Return(nil)
			},
			expectedStatus: StatusError,
			expectedValid:  false,
		},
		{
			name: "hash mismatch - content changed",
			request: &VerificationRequest{
				DocumentID: "doc-123",
				PDFData:    []byte("%PDF-1.4 different content"),
				VerifierIP: "127.0.0.1",
			},
			setupMocks: func(docRepo *MockDocumentRepository, logRepo *MockVerificationLogRepository, sigService *MockSignatureService, pdfService *MockPDFService, docService *MockDocumentService) {
				document := &entities.Document{
					ID:            "doc-123",
					DocumentHash:  testHashB64,
					SignatureData: testSignatureJSON,
					QRCodeData:    string(testQRCodeJSON),
					Status:        "active",
				}
				docRepo.On("GetByID", mock.Anything, "doc-123").Return(document, nil)

				// PDF validation and hash calculation (different hash)
				differentHash := []byte("different-hash")
				pdfService.On("ValidatePDF", mock.AnythingOfType("[]uint8")).Return(nil)
				pdfService.On("CalculateHash", mock.AnythingOfType("[]uint8")).Return(differentHash, nil)

				// Document service methods
				docService.On("DecodeSignatureData", testSignatureJSON).Return(testSignature, nil)

				// Signature verification (still valid for original hash)
				sigService.On("VerifySignature", testHash, testSignature).Return(nil)

				// Verification logging
				logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.VerificationLog")).Return(nil)
			},
			expectedStatus: StatusQRValidContentChanged,
			expectedValid:  false,
		},
		{
			name: "invalid signature",
			request: &VerificationRequest{
				DocumentID: "doc-123",
				PDFData:    []byte("%PDF-1.4 test content"),
				VerifierIP: "127.0.0.1",
			},
			setupMocks: func(docRepo *MockDocumentRepository, logRepo *MockVerificationLogRepository, sigService *MockSignatureService, pdfService *MockPDFService, docService *MockDocumentService) {
				document := &entities.Document{
					ID:            "doc-123",
					DocumentHash:  testHashB64,
					SignatureData: testSignatureJSON,
					QRCodeData:    string(testQRCodeJSON),
					Status:        "active",
				}
				docRepo.On("GetByID", mock.Anything, "doc-123").Return(document, nil)

				// PDF validation and hash calculation
				pdfService.On("ValidatePDF", mock.AnythingOfType("[]uint8")).Return(nil)
				pdfService.On("CalculateHash", mock.AnythingOfType("[]uint8")).Return(testHash, nil)

				// Document service methods
				docService.On("DecodeSignatureData", testSignatureJSON).Return(testSignature, nil)

				// Signature verification fails
				sigService.On("VerifySignature", testHash, testSignature).Return(assert.AnError)

				// Verification logging
				logRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.VerificationLog")).Return(nil)
			},
			expectedStatus: StatusInvalid,
			expectedValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockDocRepo := new(MockDocumentRepository)
			mockLogRepo := new(MockVerificationLogRepository)
			mockSigService := new(MockSignatureService)
			mockPDFService := new(MockPDFService)
			mockDocService := new(MockDocumentService)

			// Setup mocks
			tt.setupMocks(mockDocRepo, mockLogRepo, mockSigService, mockPDFService, mockDocService)

			// Create service
			service := &VerificationService{
				documentRepo:        mockDocRepo,
				verificationLogRepo: mockLogRepo,
				signatureService:    mockSigService,
				pdfService:          mockPDFService,
				documentService:     mockDocService,
			}

			// Execute
			result, err := service.VerifyDocument(context.Background(), tt.request)

			// Assert
			assert.NoError(t, err) // Service should not return errors, only verification results
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Equal(t, tt.expectedValid, result.IsValid)
			assert.Equal(t, tt.request.DocumentID, result.DocumentID)

			// Verify mocks
			mockDocRepo.AssertExpectations(t)
			mockLogRepo.AssertExpectations(t)
			mockSigService.AssertExpectations(t)
			mockPDFService.AssertExpectations(t)
			mockDocService.AssertExpectations(t)
		})
	}
}

func TestVerificationService_GetVerificationHistory(t *testing.T) {
	tests := []struct {
		name          string
		documentID    string
		setupMocks    func(*MockDocumentRepository, *MockVerificationLogRepository)
		expectedError string
		expectedCount int
	}{
		{
			name:       "successful history retrieval",
			documentID: "doc-123",
			setupMocks: func(docRepo *MockDocumentRepository, logRepo *MockVerificationLogRepository) {
				document := &entities.Document{
					ID:     "doc-123",
					Status: "active",
				}
				docRepo.On("GetByID", mock.Anything, "doc-123").Return(document, nil)

				logs := []*entities.VerificationLog{
					{
						ID:                 "log-1",
						DocumentID:         "doc-123",
						VerificationResult: StatusValid,
						VerifiedAt:         time.Now(),
					},
					{
						ID:                 "log-2",
						DocumentID:         "doc-123",
						VerificationResult: StatusInvalid,
						VerifiedAt:         time.Now(),
					},
				}
				logRepo.On("GetByDocumentID", mock.Anything, "doc-123").Return(logs, nil)
			},
			expectedError: "",
			expectedCount: 2,
		},
		{
			name:       "document not found",
			documentID: "doc-123",
			setupMocks: func(docRepo *MockDocumentRepository, logRepo *MockVerificationLogRepository) {
				docRepo.On("GetByID", mock.Anything, "doc-123").Return((*entities.Document)(nil), nil)
			},
			expectedError: "document not found",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockDocRepo := new(MockDocumentRepository)
			mockLogRepo := new(MockVerificationLogRepository)

			// Setup mocks
			tt.setupMocks(mockDocRepo, mockLogRepo)

			// Create service
			service := &VerificationService{
				documentRepo:        mockDocRepo,
				verificationLogRepo: mockLogRepo,
			}

			// Execute
			history, err := service.GetVerificationHistory(context.Background(), tt.documentID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, history)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, history)
				assert.Len(t, history, tt.expectedCount)
			}

			// Verify mocks
			mockDocRepo.AssertExpectations(t)
			mockLogRepo.AssertExpectations(t)
		})
	}
}

// Mock DocumentService for testing
type MockDocumentService struct {
	mock.Mock
}

func (m *MockDocumentService) DecodeSignatureData(signatureDataStr string) (*crypto.SignatureData, error) {
	args := m.Called(signatureDataStr)
	return args.Get(0).(*crypto.SignatureData), args.Error(1)
}