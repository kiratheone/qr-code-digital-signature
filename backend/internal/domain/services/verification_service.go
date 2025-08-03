package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
	"digital-signature-system/internal/infrastructure/crypto"
	"digital-signature-system/internal/infrastructure/pdf"
)

// DocumentServiceInterface defines the interface for document service operations needed by verification
type DocumentServiceInterface interface {
	DecodeSignatureData(signatureDataStr string) (*crypto.SignatureData, error)
}

// VerificationService handles document verification business logic
type VerificationService struct {
	documentRepo         repositories.DocumentRepository
	verificationLogRepo  repositories.VerificationLogRepository
	signatureService     SignatureServiceInterface
	pdfService           PDFServiceInterface
	documentService      DocumentServiceInterface
}

// VerificationInfo represents information about a document for verification
type VerificationInfo struct {
	DocumentID   string    `json:"document_id"`
	Filename     string    `json:"filename"`
	Issuer       string    `json:"issuer"`
	CreatedAt    time.Time `json:"created_at"`
	FileSize     int64     `json:"file_size"`
	Status       string    `json:"status"`
	QRCodeData   string    `json:"qr_code_data,omitempty"`
}

// VerificationRequest represents a request to verify a document
type VerificationRequest struct {
	DocumentID string `json:"document_id"`
	PDFData    []byte `json:"-"` // PDF file data to verify
	VerifierIP string `json:"verifier_ip"`
}

// VerificationResult represents the result of document verification
type VerificationResult struct {
	DocumentID         string                 `json:"document_id"`
	IsValid            bool                   `json:"is_valid"`
	Status             string                 `json:"status"`
	Message            string                 `json:"message"`
	Details            map[string]interface{} `json:"details"`
	VerifiedAt         time.Time              `json:"verified_at"`
	HashMatches        bool                   `json:"hash_matches"`
	SignatureValid     bool                   `json:"signature_valid"`
	QRCodeValid        bool                   `json:"qr_code_valid"`
}

// Verification status constants
const (
	StatusValid                = "valid"
	StatusQRValidContentChanged = "qr_valid_content_changed"
	StatusInvalid              = "invalid"
	StatusError                = "error"
)

// NewVerificationService creates a new verification service
func NewVerificationService(
	documentRepo repositories.DocumentRepository,
	verificationLogRepo repositories.VerificationLogRepository,
	signatureService SignatureServiceInterface,
	pdfService PDFServiceInterface,
	documentService DocumentServiceInterface,
) *VerificationService {
	return &VerificationService{
		documentRepo:        documentRepo,
		verificationLogRepo: verificationLogRepo,
		signatureService:    signatureService,
		pdfService:          pdfService,
		documentService:     documentService,
	}
}

// GetVerificationInfo retrieves information about a document for verification
func (s *VerificationService) GetVerificationInfo(ctx context.Context, documentID string) (*VerificationInfo, error) {
	// Get document from database
	document, err := s.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if document == nil {
		return nil, fmt.Errorf("document not found")
	}

	// Check if document is active
	if document.Status != "active" {
		return nil, fmt.Errorf("document is not active")
	}

	return &VerificationInfo{
		DocumentID: document.ID,
		Filename:   document.Filename,
		Issuer:     document.Issuer,
		CreatedAt:  document.CreatedAt,
		FileSize:   document.FileSize,
		Status:     document.Status,
		QRCodeData: document.QRCodeData,
	}, nil
}

// VerifyDocument verifies a document against its stored signature and hash
func (s *VerificationService) VerifyDocument(ctx context.Context, req *VerificationRequest) (*VerificationResult, error) {
	result := &VerificationResult{
		DocumentID: req.DocumentID,
		VerifiedAt: time.Now(),
		Details:    make(map[string]interface{}),
	}

	// Get original document from database
	document, err := s.documentRepo.GetByID(ctx, req.DocumentID)
	if err != nil {
		result.Status = StatusError
		result.Message = "Failed to retrieve document information"
		result.Details["error"] = err.Error()
		s.logVerification(ctx, req.DocumentID, result, req.VerifierIP)
		return result, nil
	}

	if document == nil {
		result.Status = StatusError
		result.Message = "Document not found"
		s.logVerification(ctx, req.DocumentID, result, req.VerifierIP)
		return result, nil
	}

	// Check if document is active
	if document.Status != "active" {
		result.Status = StatusError
		result.Message = "Document is not active"
		s.logVerification(ctx, req.DocumentID, result, req.VerifierIP)
		return result, nil
	}

	// Validate uploaded PDF
	if err := s.pdfService.ValidatePDF(req.PDFData); err != nil {
		result.Status = StatusError
		result.Message = "Invalid PDF file"
		result.Details["validation_error"] = err.Error()
		s.logVerification(ctx, req.DocumentID, result, req.VerifierIP)
		return result, nil
	}

	// Calculate hash of uploaded document
	uploadedHash, err := s.pdfService.CalculateHash(req.PDFData)
	if err != nil {
		result.Status = StatusError
		result.Message = "Failed to calculate document hash"
		result.Details["hash_error"] = err.Error()
		s.logVerification(ctx, req.DocumentID, result, req.VerifierIP)
		return result, nil
	}

	// Parse stored QR code data
	var qrCodeData pdf.QRCodeData
	if err := json.Unmarshal([]byte(document.QRCodeData), &qrCodeData); err != nil {
		result.Status = StatusError
		result.Message = "Failed to parse QR code data"
		result.Details["qr_parse_error"] = err.Error()
		s.logVerification(ctx, req.DocumentID, result, req.VerifierIP)
		return result, nil
	}

	// Verify QR code data matches document
	result.QRCodeValid = (qrCodeData.DocID == document.ID && qrCodeData.Hash == document.DocumentHash)
	result.Details["qr_doc_id"] = qrCodeData.DocID
	result.Details["qr_hash"] = qrCodeData.Hash
	result.Details["stored_hash"] = document.DocumentHash

	// Compare hashes (encode uploaded hash in same format as stored hash)
	uploadedHashStr := encodeHashForComparison(uploadedHash)
	result.HashMatches = (uploadedHashStr == document.DocumentHash)
	result.Details["uploaded_hash"] = uploadedHashStr
	result.Details["hash_matches"] = result.HashMatches

	// Decode and verify signature
	signatureData, err := s.documentService.DecodeSignatureData(document.SignatureData)
	if err != nil {
		result.Status = StatusError
		result.Message = "Failed to decode signature data"
		result.Details["signature_decode_error"] = err.Error()
		s.logVerification(ctx, req.DocumentID, result, req.VerifierIP)
		return result, nil
	}

	// Verify signature against original hash (from database)
	err = s.signatureService.VerifySignature(signatureData.Hash, signatureData)
	result.SignatureValid = (err == nil)
	result.Details["signature_valid"] = result.SignatureValid
	if err != nil {
		result.Details["signature_error"] = err.Error()
	}

	// Determine final verification result
	if !result.QRCodeValid {
		result.Status = StatusInvalid
		result.Message = "❌ QR invalid / signature incorrect"
		result.IsValid = false
	} else if !result.SignatureValid {
		result.Status = StatusInvalid
		result.Message = "❌ QR invalid / signature incorrect"
		result.IsValid = false
	} else if !result.HashMatches {
		result.Status = StatusQRValidContentChanged
		result.Message = "⚠️ QR valid, but file content has changed"
		result.IsValid = false
	} else {
		result.Status = StatusValid
		result.Message = "✅ Document is valid"
		result.IsValid = true
	}

	// Log verification attempt
	s.logVerification(ctx, req.DocumentID, result, req.VerifierIP)

	return result, nil
}

// logVerification logs the verification attempt
func (s *VerificationService) logVerification(ctx context.Context, documentID string, result *VerificationResult, verifierIP string) {
	details, _ := json.Marshal(result.Details)

	log := &entities.VerificationLog{
		DocumentID:         documentID,
		VerificationResult: result.Status,
		VerifiedAt:         result.VerifiedAt,
		VerifierIP:         verifierIP,
		Details:            string(details),
	}

	// Log error if logging fails, but don't fail the verification
	if err := s.verificationLogRepo.Create(ctx, log); err != nil {
		fmt.Printf("Warning: Failed to log verification attempt: %v\n", err)
	}
}

// GetVerificationHistory retrieves verification history for a document
func (s *VerificationService) GetVerificationHistory(ctx context.Context, documentID string) ([]*entities.VerificationLog, error) {
	// Verify document exists
	document, err := s.documentRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if document == nil {
		return nil, fmt.Errorf("document not found")
	}

	// Get verification logs
	logs, err := s.verificationLogRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification logs: %w", err)
	}

	return logs, nil
}

// encodeHashForComparison encodes a hash in the same format used for storage
func encodeHashForComparison(hash []byte) string {
	return base64.StdEncoding.EncodeToString(hash)
}