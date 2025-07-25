package usecases_test

import (
	"context"
	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
	"digital-signature-system/internal/domain/services"
	"digital-signature-system/internal/domain/usecases"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories and services
type MockDocumentRepository struct {
	mock.Mock
}

func (m *MockDocumentRepository) Create(ctx context.Context, doc *entities.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockDocumentRepository) GetByID(ctx context.Context, id string) (*entities.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Document), args.Error(1)
}

func (m *MockDocumentRepository) GetByUserID(ctx context.Context, userID string, filter repositories.DocumentFilter) ([]*entities.Document, int64, error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).([]*entities.Document), args.Get(1).(int64), args.Error(2)
}

func (m *MockDocumentRepository) Update(ctx context.Context, doc *entities.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockDocumentRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockVerificationLogRepository struct {
	mock.Mock
}

func (m *MockVerificationLogRepository) Create(ctx context.Context, log *entities.VerificationLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockVerificationLogRepository) GetByDocumentID(ctx context.Context, docID string) ([]*entities.VerificationLog, error) {
	args := m.Called(ctx, docID)
	return args.Get(0).([]*entities.VerificationLog), args.Error(1)
}

func (m *MockVerificationLogRepository) GetByID(ctx context.Context, id string) (*entities.VerificationLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.VerificationLog), args.Error(1)
}

type MockSignatureService struct {
	mock.Mock
}

func (m *MockSignatureService) SignDocument(docHash []byte) ([]byte, error) {
	args := m.Called(docHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSignatureService) VerifySignature(docHash []byte, signature []byte) (bool, error) {
	args := m.Called(docHash, signature)
	return args.Bool(0), args.Error(1)
}

func (m *MockSignatureService) GenerateKeyPair(bits int) (string, string, error) {
	args := m.Called(bits)
	return args.String(0), args.String(1), args.Error(2)
}

type MockPDFService struct {
	mock.Mock
}

func (m *MockPDFService) ValidatePDF(ctx context.Context, pdfData []byte) error {
	args := m.Called(ctx, pdfData)
	return args.Error(0)
}

func (m *MockPDFService) CalculateHash(ctx context.Context, pdfData []byte) ([]byte, error) {
	args := m.Called(ctx, pdfData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockPDFService) GetPDFInfo(ctx context.Context, pdfData []byte) (*services.PDFInfo, error) {
	args := m.Called(ctx, pdfData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.PDFInfo), args.Error(1)
}

type MockQRService struct {
	mock.Mock
}

func (m *MockQRService) GenerateQRCode(ctx context.Context, data *services.QRCodeData, size int) ([]byte, error) {
	args := m.Called(ctx, data, size)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockQRService) InjectQRCode(ctx context.Context, pdfData []byte, qrCodeData []byte, position *services.QRCodePosition) ([]byte, error) {
	args := m.Called(ctx, pdfData, qrCodeData, position)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockQRService) ParseQRCode(ctx context.Context, qrCodeImage []byte) (*services.QRCodeData, error) {
	args := m.Called(ctx, qrCodeImage)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.QRCodeData), args.Error(1)
}

func (m *MockQRService) GetDefaultPosition(ctx context.Context, pdfData []byte) (*services.QRCodePosition, error) {
	args := m.Called(ctx, pdfData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.QRCodePosition), args.Error(1)
}

func TestGetVerificationInfo(t *testing.T) {
	// Create mocks
	mockDocRepo := new(MockDocumentRepository)
	mockVerificationRepo := new(MockVerificationLogRepository)
	mockSignatureService := new(MockSignatureService)
	mockPDFService := new(MockPDFService)
	mockQRService := new(MockQRService)

	// Create use case
	uc := usecases.NewVerificationUseCase(
		mockDocRepo,
		mockVerificationRepo,
		mockSignatureService,
		mockPDFService,
		mockQRService,
	)

	// Test case: Document found
	t.Run("Document found", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		docID := "test-doc-id"
		createdAt := time.Now()
		
		doc := &entities.Document{
			ID:        docID,
			Filename:  "test.pdf",
			Issuer:    "Test Issuer",
			CreatedAt: createdAt,
		}
		
		mockDocRepo.On("GetByID", ctx, docID).Return(doc, nil)
		
		// Execute
		info, err := uc.GetVerificationInfo(ctx, docID)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, docID, info.DocumentID)
		assert.Equal(t, "test.pdf", info.Filename)
		assert.Equal(t, "Test Issuer", info.Issuer)
		assert.Equal(t, createdAt, info.CreatedAt)
		
		mockDocRepo.AssertExpectations(t)
	})

	// Test case: Document not found
	t.Run("Document not found", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		docID := "non-existent-id"
		
		mockDocRepo.On("GetByID", ctx, docID).Return(nil, nil)
		
		// Execute
		info, err := uc.GetVerificationInfo(ctx, docID)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "document not found")
		
		mockDocRepo.AssertExpectations(t)
	})

	// Test case: Repository error
	t.Run("Repository error", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		docID := "error-doc-id"
		repoErr := errors.New("database error")
		
		mockDocRepo.On("GetByID", ctx, docID).Return(nil, repoErr)
		
		// Execute
		info, err := uc.GetVerificationInfo(ctx, docID)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "failed to get document")
		
		mockDocRepo.AssertExpectations(t)
	})
}

func TestVerifyDocument(t *testing.T) {
	// Create mocks
	mockDocRepo := new(MockDocumentRepository)
	mockVerificationRepo := new(MockVerificationLogRepository)
	mockSignatureService := new(MockSignatureService)
	mockPDFService := new(MockPDFService)
	mockQRService := new(MockQRService)

	// Create use case
	uc := usecases.NewVerificationUseCase(
		mockDocRepo,
		mockVerificationRepo,
		mockSignatureService,
		mockPDFService,
		mockQRService,
	)

	// Test case: Valid document
	t.Run("Valid document", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		docID := "valid-doc-id"
		verifierIP := "127.0.0.1"
		pdfData := []byte("test pdf data")
		
		// Create test hash and signature
		originalHash := []byte("original-hash")
		encodedHash := base64.StdEncoding.EncodeToString(originalHash)
		signature := []byte("signature-data")
		encodedSignature := base64.StdEncoding.EncodeToString(signature)
		
		doc := &entities.Document{
			ID:            docID,
			Filename:      "valid.pdf",
			Issuer:        "Valid Issuer",
			DocumentHash:  encodedHash,
			SignatureData: encodedSignature,
			CreatedAt:     time.Now(),
		}
		
		mockDocRepo.On("GetByID", ctx, docID).Return(doc, nil)
		mockPDFService.On("CalculateHash", ctx, pdfData).Return(originalHash, nil)
		mockSignatureService.On("VerifySignature", originalHash, signature).Return(true, nil)
		mockVerificationRepo.On("Create", ctx, mock.AnythingOfType("*entities.VerificationLog")).Return(nil)
		
		// Execute
		result, err := uc.VerifyDocument(ctx, docID, pdfData, verifierIP)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "valid", result.Status)
		assert.Equal(t, "✅ Document is valid", result.Message)
		assert.Equal(t, docID, result.DocumentID)
		assert.Equal(t, doc.Issuer, result.Issuer)
		
		mockDocRepo.AssertExpectations(t)
		mockPDFService.AssertExpectations(t)
		mockSignatureService.AssertExpectations(t)
		mockVerificationRepo.AssertExpectations(t)
	})

	// Test case: Modified document
	t.Run("Modified document", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		docID := "modified-doc-id"
		verifierIP := "127.0.0.1"
		pdfData := []byte("modified pdf data")
		
		// Create test hash and signature
		originalHash := []byte("original-hash")
		encodedHash := base64.StdEncoding.EncodeToString(originalHash)
		modifiedHash := []byte("modified-hash")
		signature := []byte("signature-data")
		encodedSignature := base64.StdEncoding.EncodeToString(signature)
		
		doc := &entities.Document{
			ID:            docID,
			Filename:      "modified.pdf",
			Issuer:        "Modified Issuer",
			DocumentHash:  encodedHash,
			SignatureData: encodedSignature,
			CreatedAt:     time.Now(),
		}
		
		mockDocRepo.On("GetByID", ctx, docID).Return(doc, nil)
		mockPDFService.On("CalculateHash", ctx, pdfData).Return(modifiedHash, nil)
		mockSignatureService.On("VerifySignature", originalHash, signature).Return(true, nil)
		mockVerificationRepo.On("Create", ctx, mock.AnythingOfType("*entities.VerificationLog")).Return(nil)
		
		// Execute
		result, err := uc.VerifyDocument(ctx, docID, pdfData, verifierIP)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "modified", result.Status)
		assert.Equal(t, "⚠️ QR valid, but file content has changed", result.Message)
		assert.Equal(t, docID, result.DocumentID)
		
		mockDocRepo.AssertExpectations(t)
		mockPDFService.AssertExpectations(t)
		mockSignatureService.AssertExpectations(t)
		mockVerificationRepo.AssertExpectations(t)
	})

	// Test case: Invalid signature
	t.Run("Invalid signature", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		docID := "invalid-sig-id"
		verifierIP := "127.0.0.1"
		pdfData := []byte("test pdf data")
		
		// Create test hash and signature
		originalHash := []byte("original-hash")
		encodedHash := base64.StdEncoding.EncodeToString(originalHash)
		signature := []byte("invalid-signature")
		encodedSignature := base64.StdEncoding.EncodeToString(signature)
		
		doc := &entities.Document{
			ID:            docID,
			Filename:      "invalid.pdf",
			Issuer:        "Invalid Issuer",
			DocumentHash:  encodedHash,
			SignatureData: encodedSignature,
			CreatedAt:     time.Now(),
		}
		
		mockDocRepo.On("GetByID", ctx, docID).Return(doc, nil)
		mockPDFService.On("CalculateHash", ctx, pdfData).Return(originalHash, nil)
		mockSignatureService.On("VerifySignature", originalHash, signature).Return(false, nil)
		mockVerificationRepo.On("Create", ctx, mock.AnythingOfType("*entities.VerificationLog")).Return(nil)
		
		// Execute
		result, err := uc.VerifyDocument(ctx, docID, pdfData, verifierIP)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "invalid", result.Status)
		assert.Equal(t, "❌ QR invalid / signature incorrect", result.Message)
		assert.Equal(t, docID, result.DocumentID)
		
		mockDocRepo.AssertExpectations(t)
		mockPDFService.AssertExpectations(t)
		mockSignatureService.AssertExpectations(t)
		mockVerificationRepo.AssertExpectations(t)
	})

	// Test case: Document not found
	t.Run("Document not found", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		docID := "non-existent-id"
		verifierIP := "127.0.0.1"
		pdfData := []byte("test pdf data")
		
		mockDocRepo.On("GetByID", ctx, docID).Return(nil, nil)
		
		// Execute
		result, err := uc.VerifyDocument(ctx, docID, pdfData, verifierIP)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "document not found")
		
		mockDocRepo.AssertExpectations(t)
	})

	// Test case: Hash calculation error
	t.Run("Hash calculation error", func(t *testing.T) {
		// Setup
		ctx := context.Background()
		docID := "hash-error-id"
		verifierIP := "127.0.0.1"
		pdfData := []byte("corrupted pdf data")
		hashErr := errors.New("failed to calculate hash")
		
		doc := &entities.Document{
			ID:            docID,
			Filename:      "corrupted.pdf",
			Issuer:        "Error Issuer",
			DocumentHash:  "hash",
			SignatureData: "signature",
			CreatedAt:     time.Now(),
		}
		
		mockDocRepo.On("GetByID", ctx, docID).Return(doc, nil)
		mockPDFService.On("CalculateHash", ctx, pdfData).Return(nil, hashErr)
		mockVerificationRepo.On("Create", ctx, mock.AnythingOfType("*entities.VerificationLog")).Return(nil)
		
		// Execute
		result, err := uc.VerifyDocument(ctx, docID, pdfData, verifierIP)
		
		// Assert
		assert.NoError(t, err) // This should not return an error, but include error in result
		assert.NotNil(t, result)
		assert.Equal(t, "error", result.Status)
		assert.Contains(t, result.Message, "Failed to calculate document hash")
		
		mockDocRepo.AssertExpectations(t)
		mockPDFService.AssertExpectations(t)
		mockVerificationRepo.AssertExpectations(t)
	})
}