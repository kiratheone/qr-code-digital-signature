package handlers

import (
	"digital-signature-system/internal/domain/usecases"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// VerificationHandler handles verification-related HTTP requests
type VerificationHandler struct {
	verificationUC usecases.VerificationUseCase
}

// NewVerificationHandler creates a new verification handler
func NewVerificationHandler(verificationUC usecases.VerificationUseCase) *VerificationHandler {
	return &VerificationHandler{
		verificationUC: verificationUC,
	}
}

// GetVerificationInfo handles GET /api/verify/:docId
func (h *VerificationHandler) GetVerificationInfo(c *gin.Context) {
	// Get validated path parameters (set by validation middleware)
	validatedParams, exists := c.Get("validated_path_params")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Path validation middleware not applied",
		})
		return
	}
	
	params, ok := validatedParams.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid validated path params type",
		})
		return
	}
	
	docIDInterface, exists := params["docId"]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Document ID is required",
		})
		return
	}
	
	docID, ok := docIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid document ID type",
		})
		return
	}

	// Get verification info
	info, err := h.verificationUC.GetVerificationInfo(c.Request.Context(), docID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Document not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get verification info",
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

// VerifyDocument handles POST /api/verify/:docId/upload
func (h *VerificationHandler) VerifyDocument(c *gin.Context) {
	// Get validated path parameters (set by validation middleware)
	validatedParams, exists := c.Get("validated_path_params")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Path validation middleware not applied",
		})
		return
	}
	
	params, ok := validatedParams.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid validated path params type",
		})
		return
	}
	
	docIDInterface, exists := params["docId"]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Document ID is required",
		})
		return
	}
	
	docID, ok := docIDInterface.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid document ID type",
		})
		return
	}

	// Get validated file (set by validation middleware)
	validatedFile, exists := c.Get("validated_file")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Document file is required",
		})
		return
	}
	
	file, ok := validatedFile.(*multipart.FileHeader)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid validated file type",
		})
		return
	}

	// Open and read the validated file
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to open document file",
		})
		return
	}
	defer f.Close()

	// Read file content
	buf := make([]byte, file.Size)
	_, err = f.Read(buf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read document file",
		})
		return
	}

	// Get client IP
	clientIP := c.ClientIP()

	// Verify document
	result, err := h.verificationUC.VerifyDocument(c.Request.Context(), docID, buf, clientIP)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Document not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify document",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}