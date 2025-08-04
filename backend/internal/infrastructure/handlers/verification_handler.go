package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"digital-signature-system/internal/domain/services"
	"digital-signature-system/internal/infrastructure/logging"
	"digital-signature-system/internal/infrastructure/validation"
)

// VerificationHandler handles HTTP requests for document verification
type VerificationHandler struct {
	verificationService *services.VerificationService
	validator           *validation.Validator
}

// NewVerificationHandler creates a new verification handler
func NewVerificationHandler(verificationService *services.VerificationService) *VerificationHandler {
	return &VerificationHandler{
		verificationService: verificationService,
		validator:           validation.NewValidator(),
	}
}

// GetVerificationInfo handles GET /api/verify/:docId
func (h *VerificationHandler) GetVerificationInfo(c *gin.Context) {
	// Get and validate document ID from URL parameter
	documentID := c.Param("docId")
	if _, validationErr := h.validator.ValidateUUID("document_id", documentID, true); validationErr != nil {
		RespondWithValidationError(c, "Invalid document ID", validationErr.Error())
		return
	}

	// Get verification info
	info, err := h.verificationService.GetVerificationInfo(c.Request.Context(), documentID)
	if err != nil {
		if err.Error() == "document is not active" {
			RespondWithError(c, http.StatusGone, 
				NewStandardError(ErrCodeNotFound, "Document is no longer active"))
			return
		}
		MapServiceErrorToHTTP(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"verification_info": info})
}

// VerifyDocument handles POST /api/verify/:docId/upload
func (h *VerificationHandler) VerifyDocument(c *gin.Context) {
	// Get and validate document ID from URL parameter
	documentID := c.Param("docId")
	if _, validationErr := h.validator.ValidateUUID("document_id", documentID, true); validationErr != nil {
		RespondWithValidationError(c, "Invalid document ID", validationErr.Error())
		return
	}

	// Get file from form (form parsing is handled by middleware)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		RespondWithValidationError(c, "File is required", err.Error())
		return
	}
	defer file.Close()

	// Validate filename if provided
	if header.Filename != "" {
		if _, validationErr := h.validator.ValidateFilename("filename", header.Filename, false); validationErr != nil {
			RespondWithValidationError(c, "Invalid filename", validationErr.Error())
			return
		}
	}

	// Read file data
	pdfData, err := io.ReadAll(file)
	if err != nil {
		RespondWithInternalError(c, "Failed to read file data", err.Error())
		return
	}

	// Validate file size
	if validationErr := h.validator.ValidateFileSize("file", int64(len(pdfData)), 50<<20); validationErr != nil {
		RespondWithValidationError(c, "Invalid file size", validationErr.Error())
		return
	}

	// Get and validate client IP for logging
	clientIP := c.ClientIP()
	if sanitizedIP, validationErr := h.validator.ValidateAndSanitizeString("client_ip", clientIP, 0, 45, false); validationErr != nil {
		// If IP validation fails, use a default value for logging
		clientIP = "unknown"
	} else {
		clientIP = sanitizedIP
	}

	// Create verification request
	req := &services.VerificationRequest{
		DocumentID: documentID,
		PDFData:    pdfData,
		VerifierIP: clientIP,
	}

	// Verify document
	result, err := h.verificationService.VerifyDocument(c.Request.Context(), req)
	if err != nil {
		// Log failed verification attempt
		logging.LogVerificationAttempt(
			logging.AuditEventVerificationFailure,
			documentID,
			clientIP,
			c.GetHeader("User-Agent"),
			"FAILURE",
			map[string]interface{}{
				"error": err.Error(),
				"file_size": len(pdfData),
				"endpoint": "/api/verify/" + documentID + "/upload",
			},
		)
		MapServiceErrorToHTTP(c, err)
		return
	}

	// Determine the audit event based on verification result
	var auditEvent logging.AuditEvent
	var auditResult string
	if result.IsValid {
		auditEvent = logging.AuditEventVerificationSuccess
		auditResult = "SUCCESS"
	} else {
		auditEvent = logging.AuditEventVerificationFailure
		auditResult = "INVALID"
	}

	// Log verification attempt
	logging.LogVerificationAttempt(
		auditEvent,
		documentID,
		clientIP,
		c.GetHeader("User-Agent"),
		auditResult,
		map[string]interface{}{
			"verification_status": result.Status,
			"hash_matches": result.HashMatches,
			"signature_valid": result.SignatureValid,
			"qr_code_valid": result.QRCodeValid,
			"file_size": len(pdfData),
			"endpoint": "/api/verify/" + documentID + "/upload",
		},
	)

	// Return verification result
	c.JSON(http.StatusOK, gin.H{"verification_result": result})
}

// GetVerificationHistory handles GET /api/verify/:docId/history (optional endpoint for audit)
func (h *VerificationHandler) GetVerificationHistory(c *gin.Context) {
	// Get and validate document ID from URL parameter
	documentID := c.Param("docId")
	if _, validationErr := h.validator.ValidateUUID("document_id", documentID, true); validationErr != nil {
		RespondWithValidationError(c, "Invalid document ID", validationErr.Error())
		return
	}

	// Get verification history
	history, err := h.verificationService.GetVerificationHistory(c.Request.Context(), documentID)
	if err != nil {
		MapServiceErrorToHTTP(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"document_id": documentID,
		"history":     history,
		"total":       len(history),
	})
}