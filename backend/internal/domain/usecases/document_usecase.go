package usecases

import (
	"context"
	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
	"digital-signature-system/internal/domain/services"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DocumentUseCase defines the interface for document operations
type DocumentUseCase interface {
	// SignDocument signs a PDF document and returns the signed document information
	SignDocument(ctx context.Context, req SignDocumentRequest) (*SignDocumentResponse, error)
	
	// GetDocuments retrieves a list of documents for a user
	GetDocuments(ctx context.Context, userID string, req GetDocumentsRequest) (*GetDocumentsResponse, error)
	
	// GetDocumentByID retrieves a document by its ID
	GetDocumentByID(ctx context.Context, userID, docID string) (*DocumentResponse, error)
	
	// DeleteDocument deletes a document by its ID
	DeleteDocument(ctx context.Context, userID, docID string) error
}

// SignDocumentRequest represents a request to sign a document
type SignDocumentRequest struct {
	UserID      string
	Filename    string
	Issuer      string
	PDFData     []byte
	QRPosition  *services.QRCodePosition // Optional, if nil, default position will be used
	VerifyURL   string                   // Base URL for verification
}

// SignDocumentResponse represents the response to a sign document request
type SignDocumentResponse struct {
	DocumentID   string
	Filename     string
	Issuer       string
	DocumentHash string
	SignedPDF    []byte
	QRCodeData   string
	CreatedAt    time.Time
}

// GetDocumentsRequest represents a request to get documents
type GetDocumentsRequest struct {
	Search   string
	Page     int
	PageSize int
	SortBy   string
	SortDesc bool
}

// GetDocumentsResponse represents the response to a get documents request
type GetDocumentsResponse struct {
	Documents []DocumentResponse
	Total     int64
	Page      int
	PageSize  int
}

// DocumentResponse represents a document in responses
type DocumentResponse struct {
	ID           string    `json:"id"`
	Filename     string    `json:"filename"`
	Issuer       string    `json:"issuer"`
	DocumentHash string    `json:"document_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	FileSize     int64     `json:"file_size"`
	Status       string    `json:"status"`
}

// DocumentUseCaseImpl implements DocumentUseCase
type DocumentUseCaseImpl struct {
	documentRepo    repositories.DocumentRepository
	signatureService services.SignatureService
	pdfService      services.PDFService
	qrService       services.QRService
}

// NewDocumentUseCase creates a new document use case
func NewDocumentUseCase(
	documentRepo repositories.DocumentRepository,
	signatureService services.SignatureService,
	pdfService services.PDFService,
	qrService services.QRService,
) *DocumentUseCaseImpl {
	return &DocumentUseCaseImpl{
		documentRepo:    documentRepo,
		signatureService: signatureService,
		pdfService:      pdfService,
		qrService:       qrService,
	}
}

// SignDocument signs a PDF document and returns the signed document information
func (uc *DocumentUseCaseImpl) SignDocument(ctx context.Context, req SignDocumentRequest) (*SignDocumentResponse, error) {
	// Validate request
	if req.UserID == "" || req.Filename == "" || req.Issuer == "" || len(req.PDFData) == 0 {
		return nil, errors.New("all fields are required")
	}

	// Validate PDF
	err := uc.pdfService.ValidatePDF(ctx, req.PDFData)
	if err != nil {
		return nil, fmt.Errorf("invalid PDF: %w", err)
	}

	// Calculate document hash
	docHash, err := uc.pdfService.CalculateHash(ctx, req.PDFData)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate document hash: %w", err)
	}

	// Sign the document hash
	signature, err := uc.signatureService.SignDocument(docHash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign document: %w", err)
	}

	// Generate document ID
	documentID := uuid.New().String()

	// Create QR code data
	qrData := &services.QRCodeData{
		DocumentID:      documentID,
		Hash:            base64.StdEncoding.EncodeToString(docHash),
		Signature:       base64.StdEncoding.EncodeToString(signature),
		Timestamp:       time.Now().Unix(),
		Issuer:          req.Issuer,
		VerificationURL: fmt.Sprintf("%s/verify/%s", req.VerifyURL, documentID),
	}

	// Generate QR code
	qrCode, err := uc.qrService.GenerateQRCode(ctx, qrData, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Determine QR code position
	position := req.QRPosition
	if position == nil {
		position, err = uc.qrService.GetDefaultPosition(ctx, req.PDFData)
		if err != nil {
			return nil, fmt.Errorf("failed to get default QR position: %w", err)
		}
	}

	// Inject QR code into PDF
	signedPDF, err := uc.qrService.InjectQRCode(ctx, req.PDFData, qrCode, position)
	if err != nil {
		return nil, fmt.Errorf("failed to inject QR code: %w", err)
	}

	// Get PDF info for file size
	pdfInfo, err := uc.pdfService.GetPDFInfo(ctx, req.PDFData)
	if err != nil {
		return nil, fmt.Errorf("failed to get PDF info: %w", err)
	}

	// Encode QR data as JSON string for storage
	qrCodeDataJSON, err := services.EncodeQRCodeData(qrData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR code data: %w", err)
	}

	// Create document entity
	document := entities.NewDocument(
		req.UserID,
		req.Filename,
		req.Issuer,
		base64.StdEncoding.EncodeToString(docHash),
		base64.StdEncoding.EncodeToString(signature),
		qrCodeDataJSON,
		pdfInfo.FileSize,
	)

	// Save document to repository
	err = uc.documentRepo.Create(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	// Return response
	return &SignDocumentResponse{
		DocumentID:   document.ID,
		Filename:     document.Filename,
		Issuer:       document.Issuer,
		DocumentHash: document.DocumentHash,
		SignedPDF:    signedPDF,
		QRCodeData:   document.QRCodeData,
		CreatedAt:    document.CreatedAt,
	}, nil
}

// GetDocuments retrieves a list of documents for a user
func (uc *DocumentUseCaseImpl) GetDocuments(ctx context.Context, userID string, req GetDocumentsRequest) (*GetDocumentsResponse, error) {
	// Validate request
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	// Set default values if not provided
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// Calculate offset
	offset := (req.Page - 1) * req.PageSize

	// Create filter
	filter := repositories.DocumentFilter{
		Search:   req.Search,
		Limit:    req.PageSize,
		Offset:   offset,
		SortBy:   req.SortBy,
		SortDesc: req.SortDesc,
	}

	// Get documents from repository
	documents, total, err := uc.documentRepo.GetByUserID(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}

	// Convert to response format
	documentResponses := make([]DocumentResponse, len(documents))
	for i, doc := range documents {
		documentResponses[i] = DocumentResponse{
			ID:           doc.ID,
			Filename:     doc.Filename,
			Issuer:       doc.Issuer,
			DocumentHash: doc.DocumentHash,
			CreatedAt:    doc.CreatedAt,
			UpdatedAt:    doc.UpdatedAt,
			FileSize:     doc.FileSize,
			Status:       doc.Status,
		}
	}

	// Return response
	return &GetDocumentsResponse{
		Documents: documentResponses,
		Total:     total,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}, nil
}

// GetDocumentByID retrieves a document by its ID
func (uc *DocumentUseCaseImpl) GetDocumentByID(ctx context.Context, userID, docID string) (*DocumentResponse, error) {
	// Validate request
	if userID == "" || docID == "" {
		return nil, errors.New("user ID and document ID are required")
	}

	// Get document from repository
	document, err := uc.documentRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	if document == nil {
		return nil, errors.New("document not found")
	}

	// Check if document belongs to user
	if document.UserID != userID {
		return nil, errors.New("document not found")
	}

	// Return response
	return &DocumentResponse{
		ID:           document.ID,
		Filename:     document.Filename,
		Issuer:       document.Issuer,
		DocumentHash: document.DocumentHash,
		CreatedAt:    document.CreatedAt,
		UpdatedAt:    document.UpdatedAt,
		FileSize:     document.FileSize,
		Status:       document.Status,
	}, nil
}

// DeleteDocument deletes a document by its ID
func (uc *DocumentUseCaseImpl) DeleteDocument(ctx context.Context, userID, docID string) error {
	// Validate request
	if userID == "" || docID == "" {
		return errors.New("user ID and document ID are required")
	}

	// Get document from repository
	document, err := uc.documentRepo.GetByID(ctx, docID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}
	if document == nil {
		return errors.New("document not found")
	}

	// Check if document belongs to user
	if document.UserID != userID {
		return errors.New("document not found")
	}

	// Delete document
	err = uc.documentRepo.Delete(ctx, docID)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}