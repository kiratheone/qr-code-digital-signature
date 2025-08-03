package handlers

import (
	"fmt"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	// Get verification info
	info, err := h.verificationService.GetVerificationInfo(c.Request.Context(), documentID)
	if err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		if err.Error() == "document is not active" {
			c.JSON(http.StatusGone, gin.H{"error": "Document is no longer active"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get verification info: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"verification_info": info})
}

// VerifyDocument handles POST /api/verify/:docId/upload
func (h *VerificationHandler) VerifyDocument(c *gin.Context) {
	// Get document ID from URL parameter
	documentID := c.Param("docId")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(50 << 20) // 50MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	defer file.Close()

	// Validate file type
	if header.Header.Get("Content-Type") != "application/pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files are allowed"})
		return
	}

	// Read file data
	pdfData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read file data"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to verify document: %v", err)})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	// Get verification history
	history, err := h.verificationService.GetVerificationHistory(c.Request.Context(), documentID)
	if err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get verification history: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"document_id": documentID,
		"history":     history,
		"total":       len(history),
	})
}