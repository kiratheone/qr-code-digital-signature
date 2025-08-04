package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"digital-signature-system/internal/domain/services"
)

// VerificationHandler handles HTTP requests for document verification
type VerificationHandler struct {
	verificationService *services.VerificationService
}

// NewVerificationHandler creates a new verification handler
func NewVerificationHandler(verificationService *services.VerificationService) *VerificationHandler {
	return &VerificationHandler{
		verificationService: verificationService,
	}
}

// GetVerificationInfo handles GET /api/verify/:docId
func (h *VerificationHandler) GetVerificationInfo(c *gin.Context) {
	// Get document ID from URL parameter
	documentID := c.Param("docId")
	if documentID == "" {
		RespondWithValidationError(c, "Document ID is required")
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
	// Get document ID from URL parameter
	documentID := c.Param("docId")
	if documentID == "" {
		RespondWithValidationError(c, "Document ID is required")
		return
	}

	// Get file from form (form parsing is handled by middleware)
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		RespondWithValidationError(c, "File is required", err.Error())
		return
	}
	defer file.Close()

	// Read file data
	pdfData, err := io.ReadAll(file)
	if err != nil {
		RespondWithInternalError(c, "Failed to read file data", err.Error())
		return
	}

	// Get client IP for logging
	clientIP := c.ClientIP()

	// Create verification request
	req := &services.VerificationRequest{
		DocumentID: documentID,
		PDFData:    pdfData,
		VerifierIP: clientIP,
	}

	// Verify document
	result, err := h.verificationService.VerifyDocument(c.Request.Context(), req)
	if err != nil {
		MapServiceErrorToHTTP(c, err)
		return
	}

	// Return verification result
	c.JSON(http.StatusOK, gin.H{"verification_result": result})
}

// GetVerificationHistory handles GET /api/verify/:docId/history (optional endpoint for audit)
func (h *VerificationHandler) GetVerificationHistory(c *gin.Context) {
	// Get document ID from URL parameter
	documentID := c.Param("docId")
	if documentID == "" {
		RespondWithValidationError(c, "Document ID is required")
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