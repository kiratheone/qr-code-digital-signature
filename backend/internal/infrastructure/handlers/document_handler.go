package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"digital-signature-system/internal/domain/services"
)

// DocumentHandler handles HTTP requests for document operations
type DocumentHandler struct {
	documentService *services.DocumentService
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(documentService *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

// SignDocument handles POST /api/documents/sign
func (h *DocumentHandler) SignDocument(c *gin.Context) {
	// Get user ID from authentication context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
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

	// Get issuer from form
	issuer := c.Request.FormValue("issuer")
	if issuer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Issuer is required"})
		return
	}

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

	// Create request
	req := &services.SignDocumentRequest{
		Filename: header.Filename,
		Issuer:   issuer,
		PDFData:  pdfData,
		UserID:   userID.(string),
	}

	// Sign document
	response, err := h.documentService.SignDocument(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to sign document: %v", err)})
		return
	}

	// Return response without PDF data in JSON (too large)
	c.JSON(http.StatusCreated, gin.H{
		"document": response.Document,
		"message":  "Document signed successfully",
	})
}

// GetDocuments handles GET /api/documents
func (h *DocumentHandler) GetDocuments(c *gin.Context) {
	// Get user ID from authentication context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse query parameters
	req := &services.GetDocumentsRequest{
		UserID: userID.(string),
	}

	// Parse page parameter
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		} else {
			req.Page = 1
		}
	} else {
		req.Page = 1
	}

	// Parse page_size parameter
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			req.PageSize = pageSize
		} else {
			req.PageSize = 10
		}
	} else {
		req.PageSize = 10
	}

	// Parse status parameter
	req.Status = c.Query("status")

	// Get documents
	response, err := h.documentService.GetDocuments(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get documents: %v", err)})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDocument handles GET /api/documents/:id
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	// Get user ID from authentication context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get document ID from URL parameter
	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	// Get document
	document, err := h.documentService.GetDocumentByID(c.Request.Context(), userID.(string), documentID)
	if err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		if err.Error() == "access denied: document belongs to different user" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get document: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"document": document})
}

// DeleteDocument handles DELETE /api/documents/:id
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	// Get user ID from authentication context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get document ID from URL parameter
	documentID := c.Param("id")
	if documentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document ID is required"})
		return
	}

	// Delete document
	err := h.documentService.DeleteDocument(c.Request.Context(), userID.(string), documentID)
	if err != nil {
		if err.Error() == "document not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		if err.Error() == "access denied: document belongs to different user" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete document: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}