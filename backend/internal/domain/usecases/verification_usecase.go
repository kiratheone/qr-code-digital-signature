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
)

// VerificationResult represents the result of a document verification
type VerificationResult struct {
	Status           string    `json:"status"`           // "valid", "modified", "invalid"
	Message          string    `json:"message"`          // Human-readable message
	DocumentID       string    `json:"document_id"`      // Document ID
	Issuer           string    `json:"issuer,omitempty"` // Document issuer
	CreatedAt        time.Time `json:"created_at"`       // Document creation time
	VerificationTime time.Time `json:"verification_time"` // Time of verification
	Details          string    `json:"details,omitempty"` // Additional details (JSON)
}

// VerificationInfo represents basic information about a document for verification
type VerificationInfo struct {
	DocumentID string    `json:"document_id"`
	Filename   string    `json:"filename"`
	Issuer     string    `json:"issuer"`
	CreatedAt  time.Time `json:"created_at"`
}

// VerificationUseCase defines the interface for document verification operations
type VerificationUseCase interface {
	// GetVerificationInfo retrieves basic document information for verification
	GetVerificationInfo(ctx context.Context, docID string) (*VerificationInfo, error)
	
	// VerifyDocument verifies a document against its stored hash and signature
	VerifyDocument(ctx context.Context, docID string, documentData []byte, verifierIP string) (*VerificationResult, error)
}

// AuditService interface for audit logging
type AuditService interface {
	LogVerificationEvent(ctx context.Context, documentID, ip, result string, duration time.Duration, details map[string]interface{})
}

// MonitoringService interface for monitoring
type MonitoringService interface {
	TrackVerificationFailure(ctx context.Context, documentID, ip, reason string)
}

// verificationUseCase implements the VerificationUseCase interface
type verificationUseCase struct {
	documentRepo       repositories.DocumentRepository
	verificationRepo   repositories.VerificationLogRepository
	signatureService   services.SignatureService
	pdfService         services.PDFService
	qrService          services.QRService
	auditService       AuditService
	monitoringService  MonitoringService
}

// NewVerificationUseCase creates a new verification use case
func NewVerificationUseCase(
	documentRepo repositories.DocumentRepository,
	verificationRepo repositories.VerificationLogRepository,
	signatureService services.SignatureService,
	pdfService services.PDFService,
	qrService services.QRService,
	auditService AuditService,
	monitoringService MonitoringService,
) VerificationUseCase {
	return &verificationUseCase{
		documentRepo:       documentRepo,
		verificationRepo:   verificationRepo,
		signatureService:   signatureService,
		pdfService:         pdfService,
		qrService:          qrService,
		auditService:       auditService,
		monitoringService:  monitoringService,
	}
}

// GetVerificationInfo retrieves basic document information for verification
func (uc *verificationUseCase) GetVerificationInfo(ctx context.Context, docID string) (*VerificationInfo, error) {
	// Get document from repository
	doc, err := uc.documentRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	
	if doc == nil {
		return nil, errors.New("document not found")
	}
	
	// Return basic document information
	return &VerificationInfo{
		DocumentID: doc.ID,
		Filename:   doc.Filename,
		Issuer:     doc.Issuer,
		CreatedAt:  doc.CreatedAt,
	}, nil
}

// VerifyDocument verifies a document against its stored hash and signature
func (uc *verificationUseCase) VerifyDocument(ctx context.Context, docID string, documentData []byte, verifierIP string) (*VerificationResult, error) {
	startTime := time.Now()
	
	// Get document from repository
	doc, err := uc.documentRepo.GetByID(ctx, docID)
	if err != nil {
		// Log audit event for failed document lookup
		if uc.auditService != nil {
			uc.auditService.LogVerificationEvent(ctx, docID, verifierIP, "document_not_found", 
				time.Since(startTime), map[string]interface{}{
					"error": err.Error(),
				})
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	
	if doc == nil {
		// Log audit event for document not found
		if uc.auditService != nil {
			uc.auditService.LogVerificationEvent(ctx, docID, verifierIP, "document_not_found", 
				time.Since(startTime), map[string]interface{}{
					"reason": "document does not exist",
				})
		}
		return nil, errors.New("document not found")
	}
	
	// Calculate hash of the uploaded document
	uploadedHash, err := uc.pdfService.CalculateHash(ctx, documentData)
	if err != nil {
		duration := time.Since(startTime)
		
		// Log audit event for hash calculation failure
		if uc.auditService != nil {
			uc.auditService.LogVerificationEvent(ctx, docID, verifierIP, "hash_calculation_failed", 
				duration, map[string]interface{}{
					"error": err.Error(),
					"document_size": len(documentData),
				})
		}
		
		// Track monitoring failure
		if uc.monitoringService != nil {
			uc.monitoringService.TrackVerificationFailure(ctx, docID, verifierIP, "hash_calculation_failed")
		}
		
		return &VerificationResult{
			Status:           "error",
			Message:          "Failed to calculate document hash",
			DocumentID:       docID,
			Issuer:           doc.Issuer,
			CreatedAt:        doc.CreatedAt,
			VerificationTime: time.Now(),
			Details:          fmt.Sprintf("Error: %s", err.Error()),
		}, nil
	}
	
	// Decode stored signature
	storedSignature, err := base64.StdEncoding.DecodeString(doc.SignatureData)
	if err != nil {
		duration := time.Since(startTime)
		
		// Log audit event for signature decode failure
		if uc.auditService != nil {
			uc.auditService.LogVerificationEvent(ctx, docID, verifierIP, "signature_decode_failed", 
				duration, map[string]interface{}{
					"error": err.Error(),
				})
		}
		
		// Track monitoring failure
		if uc.monitoringService != nil {
			uc.monitoringService.TrackVerificationFailure(ctx, docID, verifierIP, "signature_decode_failed")
		}
		
		return &VerificationResult{
			Status:           "error",
			Message:          "Failed to decode stored signature",
			DocumentID:       docID,
			Issuer:           doc.Issuer,
			CreatedAt:        doc.CreatedAt,
			VerificationTime: time.Now(),
			Details:          fmt.Sprintf("Error: %s", err.Error()),
		}, nil
	}
	
	// Decode stored hash
	storedHash, err := base64.StdEncoding.DecodeString(doc.DocumentHash)
	if err != nil {
		duration := time.Since(startTime)
		
		// Log audit event for hash decode failure
		if uc.auditService != nil {
			uc.auditService.LogVerificationEvent(ctx, docID, verifierIP, "hash_decode_failed", 
				duration, map[string]interface{}{
					"error": err.Error(),
				})
		}
		
		// Track monitoring failure
		if uc.monitoringService != nil {
			uc.monitoringService.TrackVerificationFailure(ctx, docID, verifierIP, "hash_decode_failed")
		}
		
		return &VerificationResult{
			Status:           "error",
			Message:          "Failed to decode stored hash",
			DocumentID:       docID,
			Issuer:           doc.Issuer,
			CreatedAt:        doc.CreatedAt,
			VerificationTime: time.Now(),
			Details:          fmt.Sprintf("Error: %s", err.Error()),
		}, nil
	}
	
	// Verify signature against stored hash
	signatureValid, err := uc.signatureService.VerifySignature(storedHash, storedSignature)
	if err != nil {
		duration := time.Since(startTime)
		
		// Log audit event for signature verification failure
		if uc.auditService != nil {
			uc.auditService.LogVerificationEvent(ctx, docID, verifierIP, "signature_verification_failed", 
				duration, map[string]interface{}{
					"error": err.Error(),
				})
		}
		
		// Track monitoring failure
		if uc.monitoringService != nil {
			uc.monitoringService.TrackVerificationFailure(ctx, docID, verifierIP, "signature_verification_failed")
		}
		
		return &VerificationResult{
			Status:           "error",
			Message:          "Failed to verify signature",
			DocumentID:       docID,
			Issuer:           doc.Issuer,
			CreatedAt:        doc.CreatedAt,
			VerificationTime: time.Now(),
			Details:          fmt.Sprintf("Error: %s", err.Error()),
		}, nil
	}
	
	// Compare uploaded hash with stored hash
	hashMatches := base64.StdEncoding.EncodeToString(uploadedHash) == doc.DocumentHash
	
	var result *VerificationResult
	duration := time.Since(startTime)
	
	// Determine verification result
	if signatureValid && hashMatches {
		// Document is valid
		result = &VerificationResult{
			Status:           "valid",
			Message:          "✅ Document is valid",
			DocumentID:       docID,
			Issuer:           doc.Issuer,
			CreatedAt:        doc.CreatedAt,
			VerificationTime: time.Now(),
			Details:          "The document is authentic and has not been modified",
		}
	} else if signatureValid && !hashMatches {
		// QR code is valid but document has been modified
		result = &VerificationResult{
			Status:           "modified",
			Message:          "⚠️ QR valid, but file content has changed",
			DocumentID:       docID,
			Issuer:           doc.Issuer,
			CreatedAt:        doc.CreatedAt,
			VerificationTime: time.Now(),
			Details:          "The QR code is authentic but the document content has been modified",
		}
		
		// Track monitoring failure for modified document
		if uc.monitoringService != nil {
			uc.monitoringService.TrackVerificationFailure(ctx, docID, verifierIP, "document_modified")
		}
	} else {
		// Invalid signature
		result = &VerificationResult{
			Status:           "invalid",
			Message:          "❌ QR invalid / signature incorrect",
			DocumentID:       docID,
			Issuer:           doc.Issuer,
			CreatedAt:        doc.CreatedAt,
			VerificationTime: time.Now(),
			Details:          "The document signature is invalid or has been tampered with",
		}
		
		// Track monitoring failure for invalid signature
		if uc.monitoringService != nil {
			uc.monitoringService.TrackVerificationFailure(ctx, docID, verifierIP, "invalid_signature")
		}
	}
	
	// Log comprehensive audit event
	if uc.auditService != nil {
		auditDetails := map[string]interface{}{
			"document_id":        docID,
			"document_filename":  doc.Filename,
			"document_issuer":    doc.Issuer,
			"document_created":   doc.CreatedAt,
			"verification_result": result.Status,
			"signature_valid":    signatureValid,
			"hash_matches":       hashMatches,
			"document_size":      len(documentData),
			"stored_hash":        doc.DocumentHash,
			"uploaded_hash":      base64.StdEncoding.EncodeToString(uploadedHash),
		}
		
		uc.auditService.LogVerificationEvent(ctx, docID, verifierIP, result.Status, duration, auditDetails)
	}
	
	// Log verification attempt to database
	verificationLog := entities.NewVerificationLog(
		docID,
		result.Status,
		verifierIP,
		result.Details,
	)
	
	// Store verification log
	err = uc.verificationRepo.Create(ctx, verificationLog)
	if err != nil {
		// Log error but don't fail the verification
		fmt.Printf("Failed to log verification: %v\n", err)
		
		// Log audit event for database logging failure
		if uc.auditService != nil {
			uc.auditService.LogVerificationEvent(ctx, docID, verifierIP, "database_log_failed", 
				time.Since(startTime), map[string]interface{}{
					"error": err.Error(),
				})
		}
	}
	
	return result, nil
}