package usecases

import (
	"context"
	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
	"digital-signature-system/internal/domain/services"
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

func TestSignDocument_Success(t *testing.T) {
	// Create mocks
	mockDocRepo := new(MockDocumentRepository)
	mockSignatureService := new(MockSignatureService)
	mockPDFService := new(MockPDFService)
	mockQRService := new(MockQRService)

	// Create use case
	useCase := NewDocumentUseCase(mockDocRepo, mockSignatureService, mockPDFService, mockQRService)

	// Test data
	ctx := context.Background()
	userID := "user123"
	filename := "test.pdf"
	issuer := "Test Issuer"
	pdfData := []byte("fake pdf data")
	docHash := []byte("document hash")
	signature := []byte("signature")
	qrCode := []byte("qr code")
	signedPDF := []byte("signed pdf")
	verifyURL := "https://example.com"

	// Mock expectations
	mockPDFService.On("ValidatePDF", ctx, pdfData).Return(nil)
	mockPDFService.On("CalculateHash", ctx, pdfData).Return(docHash, nil)
	mockSignatureService.On("SignDocument", docHash).Return(signature, nil)
	mockQRService.On("GetDefaultPosition", ctx, pdfData).Return(&services.QRCodePosition{
		Page:   -1,
		X:      10,
		Y:      10,
		Width:  100,
		Height: 100,
	}, nil)
	mockQRService.On("GenerateQRCode", ctx, mock.AnythingOfType("*services.QRCodeData"), 256).Return(qrCode, nil)
	mockQRService.On("InjectQRCode", ctx, pdfData, qrCode, mock.AnythingOfType("*services.QRCodePosition")).Return(signedPDF, nil)
	mockPDFService.On("GetPDFInfo", ctx, pdfData).Return(&services.PDFInfo{
		PageCount:   1,
		FileSize:    1024,
		IsEncrypted: false,
	}, nil)
	mockDocRepo.On("Create", ctx, mock.AnythingOfType("*entities.Document")).Return(nil)

	// Call use case
	req := SignDocumentRequest{
		UserID:    userID,
		Filename:  filename,
		Issuer:    issuer,
		PDFData:   pdfData,
		VerifyURL: verifyURL,
	}
	resp, err := useCase.SignDocument(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, filename, resp.Filename)
	assert.Equal(t, issuer, resp.Issuer)
	assert.Equal(t, signedPDF, resp.SignedPDF)
	assert.NotEmpty(t, resp.DocumentID)
	assert.NotEmpty(t, resp.DocumentHash)
	assert.NotEmpty(t, resp.QRCodeData)

	// Verify mocks
	mockDocRepo.AssertExpectations(t)
	mockSignatureService.AssertExpectations(t)
	mockPDFService.AssertExpectations(t)
	mockQRService.AssertExpectations(t)
}

func TestSignDocument_InvalidPDF(t *testing.T) {
	// Create mocks
	mockDocRepo := new(MockDocumentRepository)
	mockSignatureService := new(MockSignatureService)
	mockPDFService := new(MockPDFService)
	mockQRService := new(MockQRService)

	// Create use case
	useCase := NewDocumentUseCase(mockDocRepo, mockSignatureService, mockPDFService, mockQRService)

	// Test data
	ctx := context.Background()
	userID := "user123"
	filename := "test.pdf"
	issuer := "Test Issuer"
	pdfData := []byte("fake pdf data")
	verifyURL := "https://example.com"

	// Mock expectations
	mockPDFService.On("ValidatePDF", ctx, pdfData).Return(errors.New("invalid PDF"))

	// Call use case
	req := SignDocumentRequest{
		UserID:    userID,
		Filename:  filename,
		Issuer:    issuer,
		PDFData:   pdfData,
		VerifyURL: verifyURL,
	}
	resp, err := useCase.SignDocument(ctx, req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid PDF")

	// Verify mocks
	mockPDFService.AssertExpectations(t)
}

func TestGetDocuments_Success(t *testing.T) {
	// Create mocks
	mockDocRepo := new(MockDocumentRepository)
	mockSignatureService := new(MockSignatureService)
	mockPDFService := new(MockPDFService)
	mockQRService := new(MockQRService)

	// Create use case
	useCase := NewDocumentUseCase(mockDocRepo, mockSignatureService, mockPDFService, mockQRService)

	// Test data
	ctx := context.Background()
	userID := "user123"
	now := time.Now()
	
	documents := []*entities.Document{
		{
			ID:            "doc1",
			UserID:        userID,
			Filename:      "test1.pdf",
			Issuer:        "Issuer 1",
			DocumentHash:  base64.StdEncoding.EncodeToString([]byte("hash1")),
			SignatureData: base64.StdEncoding.EncodeToString([]byte("sig1")),
			QRCodeData:    "qr1",
			CreatedAt:     now,
			UpdatedAt:     now,
			FileSize:      1024,
			Status:        "active",
		},
		{
			ID:            "doc2",
			UserID:        userID,
			Filename:      "test2.pdf",
			Issuer:        "Issuer 2",
			DocumentHash:  base64.StdEncoding.EncodeToString([]byte("hash2")),
			SignatureData: base64.StdEncoding.EncodeToString([]byte("sig2")),
			QRCodeData:    "qr2",
			CreatedAt:     now,
			UpdatedAt:     now,
			FileSize:      2048,
			Status:        "active",
		},
	}
	
	total := int64(2)

	// Mock expectations
	mockDocRepo.On("GetByUserID", ctx, userID, mock.AnythingOfType("repositories.DocumentFilter")).Return(documents, total, nil)

	// Call use case
	req := GetDocumentsRequest{
		Page:     1,
		PageSize: 10,
	}
	resp, err := useCase.GetDocuments(ctx, userID, req)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Documents))
	assert.Equal(t, total, resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 10, resp.PageSize)
	assert.Equal(t, "doc1", resp.Documents[0].ID)
	assert.Equal(t, "test1.pdf", resp.Documents[0].Filename)
	assert.Equal(t, "doc2", resp.Documents[1].ID)
	assert.Equal(t, "test2.pdf", resp.Documents[1].Filename)

	// Verify mocks
	mockDocRepo.AssertExpectations(t)
}

func TestGetDocumentByID_Success(t *testing.T) {
	// Create mocks
	mockDocRepo := new(MockDocumentRepository)
	mockSignatureService := new(MockSignatureService)
	mockPDFService := new(MockPDFService)
	mockQRService := new(MockQRService)

	// Create use case
	useCase := NewDocumentUseCase(mockDocRepo, mockSignatureService, mockPDFService, mockQRService)

	// Test data
	ctx := context.Background()
	userID := "user123"
	docID := "doc1"
	now := time.Now()
	
	document := &entities.Document{
		ID:            docID,
		UserID:        userID,
		Filename:      "test1.pdf",
		Issuer:        "Issuer 1",
		DocumentHash:  base64.StdEncoding.EncodeToString([]byte("hash1")),
		SignatureData: base64.StdEncoding.EncodeToString([]byte("sig1")),
		QRCodeData:    "qr1",
		CreatedAt:     now,
		UpdatedAt:     now,
		FileSize:      1024,
		Status:        "active",
	}

	// Mock expectations
	mockDocRepo.On("GetByID", ctx, docID).Return(document, nil)

	// Call use case
	resp, err := useCase.GetDocumentByID(ctx, userID, docID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, docID, resp.ID)
	assert.Equal(t, "test1.pdf", resp.Filename)
	assert.Equal(t, "Issuer 1", resp.Issuer)
	assert.Equal(t, document.DocumentHash, resp.DocumentHash)

	// Verify mocks
	mockDocRepo.AssertExpectations(t)
}

func TestGetDocumentByID_NotFound(t *testing.T) {
	// Create mocks
	mockDocRepo := new(MockDocumentRepository)
	mockSignatureService := new(MockSignatureService)
	mockPDFService := new(MockPDFService)
	mockQRService := new(MockQRService)

	// Create use case
	useCase := NewDocumentUseCase(mockDocRepo, mockSignatureService, mockPDFService, mockQRService)

	// Test data
	ctx := context.Background()
	userID := "user123"
	docID := "doc1"

	// Mock expectations
	mockDocRepo.On("GetByID", ctx, docID).Return(nil, nil)

	// Call use case
	resp, err := useCase.GetDocumentByID(ctx, userID, docID)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "document not found")

	// Verify mocks
	mockDocRepo.AssertExpectations(t)
}

func TestDeleteDocument_Success(t *testing.T) {
	// Create mocks
	mockDocRepo := new(MockDocumentRepository)
	mockSignatureService := new(MockSignatureService)
	mockPDFService := new(MockPDFService)
	mockQRService := new(MockQRService)

	// Create use case
	useCase := NewDocumentUseCase(mockDocRepo, mockSignatureService, mockPDFService, mockQRService)

	// Test data
	ctx := context.Background()
	userID := "user123"
	docID := "doc1"
	now := time.Now()
	
	document := &entities.Document{
		ID:            docID,
		UserID:        userID,
		Filename:      "test1.pdf",
		Issuer:        "Issuer 1",
		DocumentHash:  base64.StdEncoding.EncodeToString([]byte("hash1")),
		SignatureData: base64.StdEncoding.EncodeToString([]byte("sig1")),
		QRCodeData:    "qr1",
		CreatedAt:     now,
		UpdatedAt:     now,
		FileSize:      1024,
		Status:        "active",
	}

	// Mock expectations
	mockDocRepo.On("GetByID", ctx, docID).Return(document, nil)
	mockDocRepo.On("Delete", ctx, docID).Return(nil)

	// Call use case
	err := useCase.DeleteDocument(ctx, userID, docID)

	// Assertions
	assert.NoError(t, err)

	// Verify mocks
	mockDocRepo.AssertExpectations(t)
}

func TestDeleteDocument_NotFound(t *testing.T) {
	// Create mocks
	mockDocRepo := new(MockDocumentRepository)
	mockSignatureService := new(MockSignatureService)
	mockPDFService := new(MockPDFService)
	mockQRService := new(MockQRService)

	// Create use case
	useCase := NewDocumentUseCase(mockDocRepo, mockSignatureService, mockPDFService, mockQRService)

	// Test data
	ctx := context.Background()
	userID := "user123"
	docID := "doc1"

	// Mock expectations
	mockDocRepo.On("GetByID", ctx, docID).Return(nil, nil)

	// Call use case
	err := useCase.DeleteDocument(ctx, userID, docID)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document not found")

	// Verify mocks
	mockDocRepo.AssertExpectations(t)
}