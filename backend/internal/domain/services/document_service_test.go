package services

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
	"digital-signature-system/internal/infrastructure/crypto"
	"digital-signature-system/internal/infrastructure/pdf"
)

// Helper function to create a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// Mock repositories and services
type MockDocumentRepository struct {
	mock.Mock
}

func (m *MockDocumentRepository) Create(ctx context.Context, doc *entities.Document) error {
	args := m.Called(ctx, doc)
	// Simulate GORM setting the ID
	if doc.ID == "" {
		doc.ID = "test-doc-id"
	}
	return args.Error(0)
}

func (m *MockDocumentRepository) GetByID(ctx context.Context, id string) (*entities.Document, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Document), args.Error(1)
}

func (m *MockDocumentRepository) GetByUserID(ctx context.Context, userID string, filter repositories.DocumentFilter) ([]*entities.Document, int64, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).([]*entities.Document), args.Get(1).(int64), args.Error(2)
}

func (m *MockDocumentRepository) GetByHash(ctx context.Context, hash string) (*entities.Document, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(*entities.Document), args.Error(1)
}

func (m *MockDocumentRepository) Update(ctx context.Context, doc *entities.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockDocumentRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockSignatureService struct {
	mock.Mock
}

func (m *MockSignatureService) SignDocument(documentHash []byte) (*crypto.SignatureData, error) {
	args := m.Called(documentHash)
	return args.Get(0).(*crypto.SignatureData), args.Error(1)
}

func (m *MockSignatureService) VerifySignature(documentHash []byte, signatureData *crypto.SignatureData) error {
	args := m.Called(documentHash, signatureData)
	return args.Error(0)
}

type MockPDFService struct {
	mock.Mock
}

func (m *MockPDFService) ValidatePDF(pdfData []byte) error {
	args := m.Called(pdfData)
	return args.Error(0)
}

func (m *MockPDFService) CalculateHash(pdfData []byte) ([]byte, error) {
	args := m.Called(pdfData)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockPDFService) GenerateQRCode(data pdf.QRCodeData) ([]byte, error) {
	args := m.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockPDFService) InjectQRCode(pdfData []byte, qrCodeData pdf.QRCodeData, position *pdf.QRPosition) ([]byte, error) {
	args := m.Called(pdfData, qrCodeData, position)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockPDFService) ReadPDFFromReader(reader io.Reader) ([]byte, error) {
	args := m.Called(reader)
	return args.Get(0).([]byte), args.Error(1)
}

func TestDocumentService_SignDocument(t *testing.T) {
	tests := []struct {
		name          string
		request       *SignDocumentRequest
		setupMocks    func(*MockDocumentRepository, *MockSignatureService, *MockPDFService)
		expectedError string
	}{
		{
			name: "successful document signing",
			request: &SignDocumentRequest{
				Filename:     "test.pdf",
				Issuer:       "John Doe",
				Title:        "Test Document Title",
				LetterNumber: "LN-001",
				PDFData:      []byte("%PDF-1.4 test content"),
				UserID:       "user-123",
			},
			setupMocks: func(docRepo *MockDocumentRepository, sigService *MockSignatureService, pdfService *MockPDFService) {
				// PDF validation and hash calculation
				pdfService.On("ValidatePDF", mock.AnythingOfType("[]uint8")).Return(nil)
				pdfService.On("CalculateHash", mock.AnythingOfType("[]uint8")).Return([]byte("test-hash"), nil)

				// Signature creation
				sigService.On("SignDocument", []byte("test-hash")).Return(&crypto.SignatureData{
					Signature: []byte("test-signature"),
					Hash:      []byte("test-hash"),
					Algorithm: "RSA-PSS-SHA256",
				}, nil)

				// QR code generation
				pdfService.On("GenerateQRCode", mock.AnythingOfType("pdf.QRCodeData")).Return([]byte("qr-code-image"), nil)

				// Document creation and update
				docRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Document")).Return(nil)
				docRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Document")).Return(nil)

				// QR code injection (may fail in development)
				pdfService.On("InjectQRCode", mock.AnythingOfType("[]uint8"), mock.AnythingOfType("pdf.QRCodeData"), (*pdf.QRPosition)(nil)).Return([]byte("modified-pdf"), nil)
			},
			expectedError: "",
		},
		{
			name: "invalid PDF data",
			request: &SignDocumentRequest{
				Filename:     "test.pdf",
				Issuer:       "John Doe",
				Title:        "Test Title",
				LetterNumber: "LN-002",
				PDFData:      []byte("invalid pdf data"),
				UserID:       "user-123",
			},
			setupMocks: func(docRepo *MockDocumentRepository, sigService *MockSignatureService, pdfService *MockPDFService) {
				pdfService.On("ValidatePDF", mock.AnythingOfType("[]uint8")).Return(assert.AnError)
			},
			expectedError: "invalid PDF",
		},
		{
			name: "signature creation failure",
			request: &SignDocumentRequest{
				Filename:     "test.pdf",
				Issuer:       "John Doe",
				Title:        "Test Title",
				LetterNumber: "LN-003",
				PDFData:      []byte("%PDF-1.4 test content"),
				UserID:       "user-123",
			},
			setupMocks: func(docRepo *MockDocumentRepository, sigService *MockSignatureService, pdfService *MockPDFService) {
				pdfService.On("ValidatePDF", mock.AnythingOfType("[]uint8")).Return(nil)
				pdfService.On("CalculateHash", mock.AnythingOfType("[]uint8")).Return([]byte("test-hash"), nil)
				sigService.On("SignDocument", []byte("test-hash")).Return((*crypto.SignatureData)(nil), assert.AnError)
			},
			expectedError: "failed to sign document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockDocRepo := new(MockDocumentRepository)
			mockSigService := new(MockSignatureService)
			mockPDFService := new(MockPDFService)

			// Setup mocks
			tt.setupMocks(mockDocRepo, mockSigService, mockPDFService)

			// Create service
			service := &DocumentService{
				documentRepo:     mockDocRepo,
				signatureService: mockSigService,
				pdfService:       mockPDFService,
			}

			// Execute
			response, err := service.SignDocument(context.Background(), tt.request)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotNil(t, response.Document)
				assert.Equal(t, tt.request.Filename, response.Document.Filename)
				assert.Equal(t, tt.request.Issuer, response.Document.Issuer)
				assert.Equal(t, tt.request.UserID, response.Document.UserID)
				assert.Equal(t, "active", response.Document.Status)
			}

			// Verify mocks
			mockDocRepo.AssertExpectations(t)
			mockSigService.AssertExpectations(t)
			mockPDFService.AssertExpectations(t)
		})
	}
}

func TestDocumentService_GetDocuments(t *testing.T) {
	tests := []struct {
		name          string
		request       *GetDocumentsRequest
		setupMocks    func(*MockDocumentRepository)
		expectedError string
		expectedCount int
	}{
		{
			name: "successful documents retrieval",
			request: &GetDocumentsRequest{
				Page:     1,
				PageSize: 10,
				Status:   "active",
				UserID:   "user-123",
			},
			setupMocks: func(docRepo *MockDocumentRepository) {
				documents := []*entities.Document{
					{
						ID:           "doc-1",
						UserID:       "user-123",
						Filename:     "test1.pdf",
						Status:       "active",
						Title:        stringPtr("Test Document 1"),
						LetterNumber: stringPtr("LN-001"),
					},
					{
						ID:           "doc-2",
						UserID:       "user-123",
						Filename:     "test2.pdf",
						Status:       "active",
						Title:        stringPtr("Test Document 2"),
						LetterNumber: stringPtr("LN-002"),
					},
				}
				docRepo.On("GetByUserID", mock.Anything, "user-123", repositories.DocumentFilter{
					Page:     1,
					PageSize: 10,
					Status:   "active",
				}).Return(documents, int64(2), nil)
			},
			expectedError: "",
			expectedCount: 2,
		},
		{
			name: "repository error",
			request: &GetDocumentsRequest{
				Page:     1,
				PageSize: 10,
				UserID:   "user-123",
			},
			setupMocks: func(docRepo *MockDocumentRepository) {
				docRepo.On("GetByUserID", mock.Anything, "user-123", mock.AnythingOfType("repositories.DocumentFilter")).Return(([]*entities.Document)(nil), int64(0), assert.AnError)
			},
			expectedError: "failed to get documents",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockDocRepo := new(MockDocumentRepository)

			// Setup mocks
			tt.setupMocks(mockDocRepo)

			// Create service
			service := &DocumentService{
				documentRepo: mockDocRepo,
			}

			// Execute
			response, err := service.GetDocuments(context.Background(), tt.request)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Len(t, response.Documents, tt.expectedCount)
				assert.Equal(t, tt.request.Page, response.Page)
				assert.Equal(t, tt.request.PageSize, response.PageSize)
			}

			// Verify mocks
			mockDocRepo.AssertExpectations(t)
		})
	}
}

func TestDocumentService_GetDocumentByID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		documentID    string
		setupMocks    func(*MockDocumentRepository)
		expectedError string
	}{
		{
			name:       "successful document retrieval",
			userID:     "user-123",
			documentID: "doc-123",
			setupMocks: func(docRepo *MockDocumentRepository) {
				document := &entities.Document{
					ID:       "doc-123",
					UserID:   "user-123",
					Filename: "test.pdf",
					Status:   "active",
				}
				docRepo.On("GetByID", mock.Anything, "doc-123").Return(document, nil)
			},
			expectedError: "",
		},
		{
			name:       "document not found",
			userID:     "user-123",
			documentID: "doc-123",
			setupMocks: func(docRepo *MockDocumentRepository) {
				docRepo.On("GetByID", mock.Anything, "doc-123").Return((*entities.Document)(nil), nil)
			},
			expectedError: "document not found",
		},
		{
			name:       "access denied - different user",
			userID:     "user-123",
			documentID: "doc-123",
			setupMocks: func(docRepo *MockDocumentRepository) {
				document := &entities.Document{
					ID:       "doc-123",
					UserID:   "user-456", // Different user
					Filename: "test.pdf",
					Status:   "active",
				}
				docRepo.On("GetByID", mock.Anything, "doc-123").Return(document, nil)
			},
			expectedError: "access denied: document belongs to different user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockDocRepo := new(MockDocumentRepository)

			// Setup mocks
			tt.setupMocks(mockDocRepo)

			// Create service
			service := &DocumentService{
				documentRepo: mockDocRepo,
			}

			// Execute
			document, err := service.GetDocumentByID(context.Background(), tt.userID, tt.documentID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, document)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, document)
				assert.Equal(t, tt.documentID, document.ID)
				assert.Equal(t, tt.userID, document.UserID)
			}

			// Verify mocks
			mockDocRepo.AssertExpectations(t)
		})
	}
}

func TestDocumentService_DeleteDocument(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		documentID    string
		setupMocks    func(*MockDocumentRepository)
		expectedError string
	}{
		{
			name:       "successful document deletion",
			userID:     "user-123",
			documentID: "doc-123",
			setupMocks: func(docRepo *MockDocumentRepository) {
				document := &entities.Document{
					ID:       "doc-123",
					UserID:   "user-123",
					Filename: "test.pdf",
					Status:   "active",
				}
				docRepo.On("GetByID", mock.Anything, "doc-123").Return(document, nil)
				docRepo.On("Update", mock.Anything, mock.MatchedBy(func(doc *entities.Document) bool {
					return doc.Status == "deleted"
				})).Return(nil)
			},
			expectedError: "",
		},
		{
			name:       "document not found",
			userID:     "user-123",
			documentID: "doc-123",
			setupMocks: func(docRepo *MockDocumentRepository) {
				docRepo.On("GetByID", mock.Anything, "doc-123").Return((*entities.Document)(nil), nil)
			},
			expectedError: "document not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockDocRepo := new(MockDocumentRepository)

			// Setup mocks
			tt.setupMocks(mockDocRepo)

			// Create service
			service := &DocumentService{
				documentRepo: mockDocRepo,
			}

			// Execute
			err := service.DeleteDocument(context.Background(), tt.userID, tt.documentID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			mockDocRepo.AssertExpectations(t)
		})
	}
}

func TestDocumentService_EncodeDecodeSignatureData(t *testing.T) {
	service := &DocumentService{}

	originalData := &crypto.SignatureData{
		Signature: []byte("test-signature"),
		Hash:      []byte("test-hash"),
		Algorithm: "RSA-PSS-SHA256",
	}

	// Test encoding
	encoded := service.encodeSignatureData(originalData)
	assert.NotEmpty(t, encoded)

	// Verify it's valid JSON
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(encoded), &jsonData)
	assert.NoError(t, err)

	// Test decoding
	decoded, err := service.DecodeSignatureData(encoded)
	assert.NoError(t, err)
	assert.NotNil(t, decoded)
	assert.Equal(t, originalData.Signature, decoded.Signature)
	assert.Equal(t, originalData.Hash, decoded.Hash)
	assert.Equal(t, originalData.Algorithm, decoded.Algorithm)
}
