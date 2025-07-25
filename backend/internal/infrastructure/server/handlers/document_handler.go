package handlers

import (
	"digital-signature-system/internal/domain/services"
	"digital-signature-system/internal/domain/usecases"
	"digital-signature-system/internal/infrastructure/server/middleware"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// DocumentHandler handles document-related requests
type DocumentHandler struct {
	documentUseCase usecases.DocumentUseCase
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(documentUseCase usecases.DocumentUseCase) *DocumentHandler {
	return &DocumentHandler{
		documentUseCase: documentUseCase,
	}
}

// SignDocumentRequest represents a request to sign a document
type SignDocumentRequest struct {
	Filename   string                   `form:"filename" binding:"required,min=1,max=255" validate:"filename,safe_string"`
	Issuer     string                   `form:"issuer" binding:"required,min=2,max=100" validate:"issuer,safe_string,no_xss"`
	QRPosition *services.QRCodePosition `form:"qr_position"`
}

// SignDocument handles document signing
func (h *DocumentHandler) SignDocument(c *gin.Context) {
	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		middleware.AbortWithAuthenticationError(c, "Authentication required")
		return
	}

	// Get validated form data (set by validation middleware)
	validatedData, exists := c.Get("validated_data")
	if !exists {
		middleware.AbortWithInternalError(c, "Validation middleware not applied")
		return
	}
	
	req, ok := validatedData.(*SignDocumentRequest)
	if !ok {
		middleware.AbortWithInternalError(c, "Invalid validated data type")
		return
	}

	// Get validated file (set by validation middleware)
	validatedFile, exists := c.Get("validated_file")
	if !exists {
		middleware.AbortWithValidationError(c, "PDF file is required", "No validated file found")
		return
	}
	
	header, ok := validatedFile.(*multipart.FileHeader)
	if !ok {
		middleware.AbortWithInternalError(c, "Invalid validated file type")
		return
	}

	// Open and read the validated file
	file, err := header.Open()
	if err != nil {
		middleware.AbortWithInternalError(c, "Failed to open file")
		return
	}
	defer file.Close()

	// Read file content
	pdfData, err := io.ReadAll(file)
	if err != nil {
		middleware.AbortWithInternalError(c, "Failed to read file")
		return
	}

	// Get base URL for verification
	baseURL := c.Request.Host
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	verifyURL := fmt.Sprintf("%s://%s", scheme, baseURL)

	// Create use case request
	ucReq := usecases.SignDocumentRequest{
		UserID:     userID,
		Filename:   req.Filename,
		Issuer:     req.Issuer,
		PDFData:    pdfData,
		QRPosition: req.QRPosition,
		VerifyURL:  verifyURL,
	}

	// Call use case
	resp, err := h.documentUseCase.SignDocument(c.Request.Context(), ucReq)
	if err != nil {
		if strings.Contains(err.Error(), "invalid PDF") || strings.Contains(err.Error(), "PDF") {
			middleware.AbortWithValidationError(c, "Invalid PDF file", err.Error())
		} else {
			middleware.AbortWithInternalError(c, "Failed to sign document")
		}
		return
	}

	// Set response headers for file download
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"signed_%s\"", req.Filename))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", strconv.Itoa(len(resp.SignedPDF)))

	// Return the signed PDF
	c.Data(http.StatusOK, "application/pdf", resp.SignedPDF)
}

// GetDocuments handles retrieving a list of documents
func (h *DocumentHandler) GetDocuments(c *gin.Context) {
	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		middleware.AbortWithAuthenticationError(c, "Authentication required")
		return
	}

	// Get validated query parameters (set by validation middleware)
	validatedParams, exists := c.Get("validated_query_params")
	if !exists {
		middleware.AbortWithInternalError(c, "Query validation middleware not applied")
		return
	}
	
	params, ok := validatedParams.(map[string]interface{})
	if !ok {
		middleware.AbortWithInternalError(c, "Invalid validated query params type")
		return
	}

	// Extract validated parameters with defaults
	search := ""
	if val, exists := params["search"]; exists && val != nil {
		search = val.(string)
	}
	
	page := 1
	if val, exists := params["page"]; exists && val != nil {
		page = val.(int)
	}
	
	pageSize := 10
	if val, exists := params["page_size"]; exists && val != nil {
		pageSize = val.(int)
	}
	
	sortBy := "created_at"
	if val, exists := params["sort_by"]; exists && val != nil {
		sortBy = val.(string)
	}
	
	sortDesc := true
	if val, exists := params["sort_desc"]; exists && val != nil {
		sortDesc = val.(bool)
	}

	// Create use case request
	req := usecases.GetDocumentsRequest{
		Search:   search,
		Page:     page,
		PageSize: pageSize,
		SortBy:   sortBy,
		SortDesc: sortDesc,
	}

	// Call use case
	resp, err := h.documentUseCase.GetDocuments(c.Request.Context(), userID, req)
	if err != nil {
		middleware.AbortWithInternalError(c, "Failed to get documents")
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"documents": resp.Documents,
		"total":     resp.Total,
		"page":      resp.Page,
		"page_size": resp.PageSize,
	})
}

// GetDocumentByID handles retrieving a document by ID
func (h *DocumentHandler) GetDocumentByID(c *gin.Context) {
	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		middleware.AbortWithAuthenticationError(c, "Authentication required")
		return
	}

	// Get validated path parameters (set by validation middleware)
	validatedParams, exists := c.Get("validated_path_params")
	if !exists {
		middleware.AbortWithInternalError(c, "Path validation middleware not applied")
		return
	}
	
	params, ok := validatedParams.(map[string]interface{})
	if !ok {
		middleware.AbortWithInternalError(c, "Invalid validated path params type")
		return
	}
	
	docID, exists := params["id"]
	if !exists {
		middleware.AbortWithValidationError(c, "Document ID is required", "Missing document ID in URL path")
		return
	}
	
	docIDStr, ok := docID.(string)
	if !ok {
		middleware.AbortWithInternalError(c, "Invalid document ID type")
		return
	}

	// Call use case
	doc, err := h.documentUseCase.GetDocumentByID(c.Request.Context(), userID, docIDStr)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			middleware.AbortWithNotFoundError(c, "Document")
		} else {
			middleware.AbortWithInternalError(c, "Failed to get document")
		}
		return
	}

	// Return response
	c.JSON(http.StatusOK, doc)
}

// DeleteDocument handles deleting a document
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		middleware.AbortWithAuthenticationError(c, "Authentication required")
		return
	}

	// Get validated path parameters (set by validation middleware)
	validatedParams, exists := c.Get("validated_path_params")
	if !exists {
		middleware.AbortWithInternalError(c, "Path validation middleware not applied")
		return
	}
	
	params, ok := validatedParams.(map[string]interface{})
	if !ok {
		middleware.AbortWithInternalError(c, "Invalid validated path params type")
		return
	}
	
	docID, exists := params["id"]
	if !exists {
		middleware.AbortWithValidationError(c, "Document ID is required", "Missing document ID in URL path")
		return
	}
	
	docIDStr, ok := docID.(string)
	if !ok {
		middleware.AbortWithInternalError(c, "Invalid document ID type")
		return
	}

	// Call use case
	err := h.documentUseCase.DeleteDocument(c.Request.Context(), userID, docIDStr)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			middleware.AbortWithNotFoundError(c, "Document")
		} else {
			middleware.AbortWithInternalError(c, "Failed to delete document")
		}
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}