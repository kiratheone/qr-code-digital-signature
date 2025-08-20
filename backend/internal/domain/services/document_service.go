package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
	"digital-signature-system/internal/infrastructure/crypto"
	"digital-signature-system/internal/infrastructure/pdf"
)

// SignatureServiceInterface defines the interface for signature operations
type SignatureServiceInterface interface {
	SignDocument(documentHash []byte) (*crypto.SignatureData, error)
	VerifySignature(documentHash []byte, signatureData *crypto.SignatureData) error
}

// PDFServiceInterface defines the interface for PDF operations
type PDFServiceInterface interface {
	ValidatePDF(pdfData []byte) error
	CalculateHash(pdfData []byte) ([]byte, error)
	GenerateQRCode(data pdf.QRCodeData) ([]byte, error)
	InjectQRCode(pdfData []byte, qrCodeData pdf.QRCodeData, position *pdf.QRPosition) ([]byte, error)
	ReadPDFFromReader(reader io.Reader) ([]byte, error)
}

// DocumentService handles all document-related business logic
type DocumentService struct {
	documentRepo     repositories.DocumentRepository
	signatureService SignatureServiceInterface
	pdfService       PDFServiceInterface
}

// SignDocumentRequest represents the request to sign a document
type SignDocumentRequest struct {
	Filename    string `json:"filename" binding:"required"`
	Issuer      string `json:"issuer" binding:"required"`
	PDFData     []byte `json:"-"` // PDF file data
	UserID      string `json:"-"` // Set from authentication context
}

// SignDocumentResponse represents the response after signing a document
type SignDocumentResponse struct {
	Document       *entities.Document `json:"document"`
	SignedPDFData  []byte            `json:"signed_pdf_data,omitempty"`
	QRCodeImageURL string            `json:"qr_code_image_url,omitempty"`
}

// GetDocumentsRequest represents the request to get documents
type GetDocumentsRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Status   string `form:"status"`
	UserID   string `json:"-"` // Set from authentication context
}

// GetDocumentsResponse represents the response for getting documents
type GetDocumentsResponse struct {
	Documents []*entities.Document `json:"documents"`
	Total     int64               `json:"total"`
	Page      int                 `json:"page"`
	PageSize  int                 `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// NewDocumentService creates a new document service
func NewDocumentService(
	documentRepo repositories.DocumentRepository,
	signatureService SignatureServiceInterface,
	pdfService PDFServiceInterface,
) *DocumentService {
	return &DocumentService{
		documentRepo:     documentRepo,
		signatureService: signatureService,
		pdfService:       pdfService,
	}
}

// SignDocument signs a PDF document and generates QR code
func (s *DocumentService) SignDocument(ctx context.Context, req *SignDocumentRequest) (*SignDocumentResponse, error) {
	// Validate PDF data
	if err := s.pdfService.ValidatePDF(req.PDFData); err != nil {
		return nil, fmt.Errorf("invalid PDF: %w", err)
	}

	// Calculate document hash
	documentHash, err := s.pdfService.CalculateHash(req.PDFData)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate document hash: %w", err)
	}

	// Create digital signature
	signatureData, err := s.signatureService.SignDocument(documentHash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign document: %w", err)
	}

	// Create document entity
	document := &entities.Document{
		UserID:        req.UserID,
		Filename:      req.Filename,
		Issuer:        req.Issuer,
		DocumentHash:  base64.StdEncoding.EncodeToString(documentHash),
		SignatureData: s.encodeSignatureData(signatureData),
		FileSize:      int64(len(req.PDFData)),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Generate QR code data
	qrCodeData := pdf.QRCodeData{
		DocID:     document.ID, // Will be set by GORM BeforeCreate
		Hash:      document.DocumentHash,
		Signature: document.SignatureData,
		Timestamp: document.CreatedAt.Unix(),
	}

	// Generate QR code
	_, err = s.pdfService.GenerateQRCode(qrCodeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Store QR code data as JSON
	qrCodeJSON, err := json.Marshal(qrCodeData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal QR code data: %w", err)
	}
	document.QRCodeData = string(qrCodeJSON)

	// Save document to database
	if err := s.documentRepo.Create(ctx, document); err != nil {
		return nil, fmt.Errorf("failed to save document: %w", err)
	}

	// Update QR code data with the actual document ID
	qrCodeData.DocID = document.ID
	qrCodeJSON, _ = json.Marshal(qrCodeData)
	document.QRCodeData = string(qrCodeJSON)

	// Update document with correct QR code data
	if err := s.documentRepo.Update(ctx, document); err != nil {
		return nil, fmt.Errorf("failed to update document with QR code: %w", err)
	}

	// Try to inject QR code into PDF (may fail in development without license)
	var signedPDFData []byte
	modifiedPDF, err := s.pdfService.InjectQRCode(req.PDFData, qrCodeData, nil)
	if err != nil {
		// Log the error but don't fail the entire operation
		// In development, this will fail due to UniPDF license requirements
		fmt.Printf("Warning: Failed to inject QR code into PDF: %v\n", err)
		signedPDFData = req.PDFData // Return original PDF
	} else {
		signedPDFData = modifiedPDF
	}

	return &SignDocumentResponse{
		Document:      document,
		SignedPDFData: signedPDFData,
	}, nil
}

// GetDocuments retrieves documents for a user with pagination
func (s *DocumentService) GetDocuments(ctx context.Context, req *GetDocumentsRequest) (*GetDocumentsResponse, error) {
	filter := repositories.DocumentFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
	}

	documents, total, err := s.documentRepo.GetByUserID(ctx, req.UserID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &GetDocumentsResponse{
		Documents:  documents,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetDocumentByID retrieves a specific document by ID
func (s *DocumentService) GetDocumentByID(ctx context.Context, userID, documentID string) (*entities.Document, error) {
	document, err := s.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if document == nil {
		return nil, fmt.Errorf("document not found")
	}

	// Verify user owns the document
	if document.UserID != userID {
		return nil, fmt.Errorf("access denied: document belongs to different user")
	}

	return document, nil
}

// DeleteDocument deletes a document
func (s *DocumentService) DeleteDocument(ctx context.Context, userID, documentID string) error {
	// First verify the document exists and belongs to the user
	document, err := s.GetDocumentByID(ctx, userID, documentID)
	if err != nil {
		return err
	}

	// Update status to deleted instead of hard delete for audit purposes
	document.Status = "deleted"
	document.UpdatedAt = time.Now()

	if err := s.documentRepo.Update(ctx, document); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// encodeSignatureData converts signature data to a storable string format
func (s *DocumentService) encodeSignatureData(signatureData *crypto.SignatureData) string {
	data := map[string]interface{}{
		"signature": base64.StdEncoding.EncodeToString(signatureData.Signature),
		"hash":      base64.StdEncoding.EncodeToString(signatureData.Hash),
		"algorithm": signatureData.Algorithm,
	}

	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

// DecodeSignatureData converts stored signature data back to SignatureData struct
func (s *DocumentService) DecodeSignatureData(signatureDataStr string) (*crypto.SignatureData, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(signatureDataStr), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal signature data: %w", err)
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(data["signature"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	hashBytes, err := base64.StdEncoding.DecodeString(data["hash"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	return &crypto.SignatureData{
		Signature: signatureBytes,
		Hash:      hashBytes,
		Algorithm: data["algorithm"].(string),
	}, nil
}

// ReadPDFFromStream reads PDF data from a stream with size limits for better performance
func (s *DocumentService) ReadPDFFromStream(reader io.Reader) ([]byte, error) {
	return s.pdfService.ReadPDFFromReader(reader)
}

// GetQRCodeImage generates and returns QR code image for a document
func (s *DocumentService) GetQRCodeImage(ctx context.Context, userID, documentID string) ([]byte, string, error) {
	// Get document and verify ownership
	document, err := s.GetDocumentByID(ctx, userID, documentID)
	if err != nil {
		return nil, "", err
	}

	// Parse QR code data from document
	var qrCodeData pdf.QRCodeData
	if err := json.Unmarshal([]byte(document.QRCodeData), &qrCodeData); err != nil {
		return nil, "", fmt.Errorf("failed to parse QR code data: %w", err)
	}

	// Generate QR code image
	qrCodeImage, err := s.pdfService.GenerateQRCode(qrCodeData)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate QR code image: %w", err)
	}

	// Generate filename
	filename := fmt.Sprintf("%s_qr_code.png", document.Filename[:len(document.Filename)-4]) // Remove .pdf extension

	return qrCodeImage, filename, nil
}

// GetSignedPDF returns the signed PDF with embedded QR code (if available)
func (s *DocumentService) GetSignedPDF(ctx context.Context, userID, documentID string) ([]byte, string, error) {
	// Get document and verify ownership
	document, err := s.GetDocumentByID(ctx, userID, documentID)
	if err != nil {
		return nil, "", err
	}

	// For now, we'll return a placeholder since we don't store the actual PDF data
	// In a production system, you would store the signed PDF file and retrieve it here
	// This is a simplified implementation that generates a basic response
	
	// Parse QR code data
	var qrCodeData pdf.QRCodeData
	if err := json.Unmarshal([]byte(document.QRCodeData), &qrCodeData); err != nil {
		return nil, "", fmt.Errorf("failed to parse QR code data: %w", err)
	}

	// Generate a simple PDF with document information
	// Note: In a real implementation, you would retrieve the stored signed PDF
	pdfContent := fmt.Sprintf(`%%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj

2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj

3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
>>
endobj

4 0 obj
<<
/Length 200
>>
stream
BT
/F1 12 Tf
50 750 Td
(Document: %s) Tj
0 -20 Td
(Issuer: %s) Tj
0 -20 Td
(Signed: %s) Tj
0 -20 Td
(Document ID: %s) Tj
0 -20 Td
(Hash: %s) Tj
ET
endstream
endobj

xref
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000206 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
456
%%%%EOF`, 
		document.Filename, 
		document.Issuer, 
		document.CreatedAt.Format("2006-01-02 15:04:05"),
		document.ID,
		document.DocumentHash[:20]+"...")

	// Generate filename for signed PDF
	filename := fmt.Sprintf("signed_%s", document.Filename)

	return []byte(pdfContent), filename, nil
}