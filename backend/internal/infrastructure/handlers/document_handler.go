package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"digital-signature-system/internal/domain/services"
	"digital-signature-system/internal/infrastructure/logging"
	"digital-signature-system/internal/infrastructure/validation"
)

// DocumentHandler handles HTTP requests for document operations
type DocumentHandler struct {
	documentService *services.DocumentService
	validator       *validation.Validator
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(documentService *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
		validator:       validation.NewValidator(),
	}
}

// Helper function to convert nullable LetterNumber for logging
func getLetterNumberForLogging(letterNumber *string) string {
	if letterNumber == nil {
		return ""
	}
	return *letterNumber
}

// Helper function to convert nullable Title for logging
func getTitleForLogging(title *string) string {
	if title == nil {
		return ""
	}
	return *title
}

// SignDocument handles POST /api/documents/sign
func (h *DocumentHandler) SignDocument(c *gin.Context) {
	// Get user ID from authentication context
	userID, exists := c.Get("user_id")
	if !exists {
		RespondWithUnauthorizedError(c, "User not authenticated")
		return
	}

	// Validate user ID format
	if _, validationErr := h.validator.ValidateUUID("user_id", userID.(string), true); validationErr != nil {
		RespondWithValidationError(c, "Invalid user ID", validationErr.Error())
		return
	}

	// Get file from form (form parsing is handled by middleware)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		RespondWithValidationError(c, "File is required", err.Error())
		return
	}
	defer file.Close()

	// Validate and sanitize filename (use sanitized version from middleware if available)
	filename := header.Filename
	if sanitizedFilename, exists := c.Get("sanitized_filename_file"); exists {
		filename = sanitizedFilename.(string)
	} else {
		// Fallback validation if middleware didn't process it
		if sanitized, validationErr := h.validator.ValidateFilename("filename", header.Filename, true); validationErr != nil {
			RespondWithValidationError(c, "Invalid filename", validationErr.Error())
			return
		} else {
			filename = sanitized
		}
	}

	// Get and validate issuer from form
	issuer := c.Request.FormValue("issuer")
	sanitizedIssuer, validationErr := h.validator.ValidateAndSanitizeString("issuer", issuer, 1, 100, true)
	if validationErr != nil {
		RespondWithValidationError(c, "Invalid issuer", validationErr.Error())
		return
	}

	// Get and validate title from form
	title := c.Request.FormValue("title")
	sanitizedTitle, titleValidationErr := h.validator.ValidateAndSanitizeString("title", title, 1, 200, true)
	if titleValidationErr != nil {
		RespondWithValidationError(c, "Invalid title", titleValidationErr.Error())
		return
	}

	// Get and validate letter number from form
	letterNumber := c.Request.FormValue("letter_number")
	sanitizedLetterNumber, letterValidationErr := h.validator.ValidateAndSanitizeString("letter_number", letterNumber, 1, 50, true)
	if letterValidationErr != nil {
		RespondWithValidationError(c, "Invalid letter number", letterValidationErr.Error())
		return
	}

	// Use streaming to read PDF data with size limit for better performance
	pdfData, err := h.documentService.ReadPDFFromStream(file)
	if err != nil {
		RespondWithValidationError(c, "Failed to process PDF file", err.Error())
		return
	}

	// Create request
	req := &services.SignDocumentRequest{
		Filename:     filename,
		Issuer:       sanitizedIssuer,
		Title:        sanitizedTitle,
		LetterNumber: sanitizedLetterNumber,
		PDFData:      pdfData,
		UserID:       userID.(string),
	}

	// Get user info for logging
	user, _ := c.Get("user")
	authUser := user.(*services.AuthenticatedUser)

	// Sign document
	response, err := h.documentService.SignDocument(c.Request.Context(), req)
	if err != nil {
		// Log failed document signing attempt
		logging.LogDocumentOperation(
			logging.AuditEventDocumentSign,
			authUser.ID,
			authUser.Username,
			"", // No document ID for failed signing
			c.ClientIP(),
			"FAILURE",
			map[string]interface{}{
				"filename":      filename,
				"issuer":        sanitizedIssuer,
				"letter_number": sanitizedLetterNumber,
				"file_size":     len(pdfData),
				"error":         err.Error(),
				"endpoint":      "/api/documents/sign",
			},
		)
		MapServiceErrorToHTTP(c, err)
		return
	}

	// Log successful document signing
	logging.LogDocumentOperation(
		logging.AuditEventDocumentSign,
		authUser.ID,
		authUser.Username,
		response.Document.ID,
		c.ClientIP(),
		"SUCCESS",
		map[string]interface{}{
			"filename":      response.Document.Filename,
			"issuer":        response.Document.Issuer,
			"title":         getTitleForLogging(response.Document.Title),
			"letter_number": getLetterNumberForLogging(response.Document.LetterNumber),
			"file_size":     len(pdfData),
			"endpoint":      "/api/documents/sign",
		},
	)

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
		RespondWithUnauthorizedError(c, "User not authenticated")
		return
	}

	// Validate user ID format
	if _, validationErr := h.validator.ValidateUUID("user_id", userID.(string), true); validationErr != nil {
		RespondWithValidationError(c, "Invalid user ID", validationErr.Error())
		return
	}

	// Parse query parameters with validation
	req := &services.GetDocumentsRequest{
		UserID: userID.(string),
	}

	// Parse and validate page parameter
	if pageStr := c.Query("page"); pageStr != "" {
		if sanitizedPage, validationErr := h.validator.ValidateAndSanitizeString("page", pageStr, 1, 10, false); validationErr != nil {
			RespondWithValidationError(c, "Invalid page parameter", validationErr.Error())
			return
		} else if sanitizedPage != "" {
			if page, err := strconv.Atoi(sanitizedPage); err == nil && page > 0 && page <= 1000 {
				req.Page = page
			} else {
				RespondWithValidationError(c, "Page must be a positive integer between 1 and 1000")
				return
			}
		}
	} else {
		req.Page = 1
	}

	// Parse and validate page_size parameter
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if sanitizedPageSize, validationErr := h.validator.ValidateAndSanitizeString("page_size", pageSizeStr, 1, 10, false); validationErr != nil {
			RespondWithValidationError(c, "Invalid page_size parameter", validationErr.Error())
			return
		} else if sanitizedPageSize != "" {
			if pageSize, err := strconv.Atoi(sanitizedPageSize); err == nil && pageSize > 0 && pageSize <= 100 {
				req.PageSize = pageSize
			} else {
				RespondWithValidationError(c, "Page size must be a positive integer between 1 and 100")
				return
			}
		}
	} else {
		req.PageSize = 10
	}

	// Parse and validate status parameter
	if status := c.Query("status"); status != "" {
		if sanitizedStatus, validationErr := h.validator.ValidateAndSanitizeString("status", status, 1, 20, false); validationErr != nil {
			RespondWithValidationError(c, "Invalid status parameter", validationErr.Error())
			return
		} else {
			// Only allow specific status values
			allowedStatuses := []string{"active", "inactive", "deleted"}
			validStatus := false
			for _, allowedStatus := range allowedStatuses {
				if sanitizedStatus == allowedStatus {
					validStatus = true
					break
				}
			}
			if !validStatus {
				RespondWithValidationError(c, "Status must be one of: active, inactive, deleted")
				return
			}
			req.Status = sanitizedStatus
		}
	} else {
		// Default to showing only active documents (hide deleted documents)
		req.Status = "active"
	}

	// Get user info for logging
	user, _ := c.Get("user")
	authUser := user.(*services.AuthenticatedUser)

	// Get documents
	response, err := h.documentService.GetDocuments(c.Request.Context(), req)
	if err != nil {
		MapServiceErrorToHTTP(c, err)
		return
	}

	// Log document list access
	logging.LogDocumentOperation(
		logging.AuditEventDocumentList,
		authUser.ID,
		authUser.Username,
		"", // No specific document ID for list operation
		c.ClientIP(),
		"SUCCESS",
		map[string]interface{}{
			"page":            req.Page,
			"page_size":       req.PageSize,
			"status":          req.Status,
			"total_documents": len(response.Documents),
			"endpoint":        "/api/documents",
		},
	)

	c.JSON(http.StatusOK, response)
}

// GetDocument handles GET /api/documents/:id
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	// Get user ID from authentication context
	userID, exists := c.Get("user_id")
	if !exists {
		RespondWithUnauthorizedError(c, "User not authenticated")
		return
	}

	// Validate user ID format
	if _, validationErr := h.validator.ValidateUUID("user_id", userID.(string), true); validationErr != nil {
		RespondWithValidationError(c, "Invalid user ID", validationErr.Error())
		return
	}

	// Get and validate document ID from URL parameter
	documentID := c.Param("id")
	if _, validationErr := h.validator.ValidateUUID("document_id", documentID, true); validationErr != nil {
		RespondWithValidationError(c, "Invalid document ID", validationErr.Error())
		return
	}

	// Get user info for logging
	user, _ := c.Get("user")
	authUser := user.(*services.AuthenticatedUser)

	// Get document
	document, err := h.documentService.GetDocumentByID(c.Request.Context(), userID.(string), documentID)
	if err != nil {
		// Log failed document access attempt
		logging.LogDocumentOperation(
			logging.AuditEventDocumentView,
			authUser.ID,
			authUser.Username,
			documentID,
			c.ClientIP(),
			"FAILURE",
			map[string]interface{}{
				"error":    err.Error(),
				"endpoint": "/api/documents/" + documentID,
			},
		)
		MapServiceErrorToHTTP(c, err)
		return
	}

	// Log successful document access
	logging.LogDocumentOperation(
		logging.AuditEventDocumentView,
		authUser.ID,
		authUser.Username,
		document.ID,
		c.ClientIP(),
		"SUCCESS",
		map[string]interface{}{
			"filename": document.Filename,
			"issuer":   document.Issuer,
			"endpoint": "/api/documents/" + documentID,
		},
	)

	c.JSON(http.StatusOK, gin.H{"document": document})
}

// DeleteDocument handles DELETE /api/documents/:id
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	// Get user ID from authentication context
	userID, exists := c.Get("user_id")
	if !exists {
		RespondWithUnauthorizedError(c, "User not authenticated")
		return
	}

	// Validate user ID format
	if _, validationErr := h.validator.ValidateUUID("user_id", userID.(string), true); validationErr != nil {
		RespondWithValidationError(c, "Invalid user ID", validationErr.Error())
		return
	}

	// Get and validate document ID from URL parameter
	documentID := c.Param("id")
	if _, validationErr := h.validator.ValidateUUID("document_id", documentID, true); validationErr != nil {
		RespondWithValidationError(c, "Invalid document ID", validationErr.Error())
		return
	}

	// Get user info for logging
	user, _ := c.Get("user")
	authUser := user.(*services.AuthenticatedUser)

	// Delete document
	err := h.documentService.DeleteDocument(c.Request.Context(), userID.(string), documentID)
	if err != nil {
		// Log failed document deletion attempt
		logging.LogDocumentOperation(
			logging.AuditEventDocumentDelete,
			authUser.ID,
			authUser.Username,
			documentID,
			c.ClientIP(),
			"FAILURE",
			map[string]interface{}{
				"error":    err.Error(),
				"endpoint": "/api/documents/" + documentID,
			},
		)
		MapServiceErrorToHTTP(c, err)
		return
	}

	// Log successful document deletion
	logging.LogDocumentOperation(
		logging.AuditEventDocumentDelete,
		authUser.ID,
		authUser.Username,
		documentID,
		c.ClientIP(),
		"SUCCESS",
		map[string]interface{}{
			"endpoint": "/api/documents/" + documentID,
		},
	)

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}

// DownloadQRCode handles GET /api/documents/:id/qr-code
func (h *DocumentHandler) DownloadQRCode(c *gin.Context) {
	// Get user ID from authentication context
	userID, exists := c.Get("user_id")
	if !exists {
		RespondWithUnauthorizedError(c, "User not authenticated")
		return
	}

	// Validate user ID format
	if _, validationErr := h.validator.ValidateUUID("user_id", userID.(string), true); validationErr != nil {
		RespondWithValidationError(c, "Invalid user ID", validationErr.Error())
		return
	}

	// Get and validate document ID from URL parameter
	documentID := c.Param("id")
	if _, validationErr := h.validator.ValidateUUID("document_id", documentID, true); validationErr != nil {
		RespondWithValidationError(c, "Invalid document ID", validationErr.Error())
		return
	}

	// Get user info for logging
	user, _ := c.Get("user")
	authUser := user.(*services.AuthenticatedUser)

	// Get QR code image
	qrCodeImage, filename, err := h.documentService.GetQRCodeImage(c.Request.Context(), userID.(string), documentID)
	if err != nil {
		// Log failed QR code download attempt
		logging.LogDocumentOperation(
			logging.AuditEventDocumentView,
			authUser.ID,
			authUser.Username,
			documentID,
			c.ClientIP(),
			"FAILURE",
			map[string]interface{}{
				"operation": "qr_code_download",
				"error":     err.Error(),
				"endpoint":  "/api/documents/" + documentID + "/qr-code",
			},
		)
		MapServiceErrorToHTTP(c, err)
		return
	}

	// Log successful QR code download
	logging.LogDocumentOperation(
		logging.AuditEventDocumentView,
		authUser.ID,
		authUser.Username,
		documentID,
		c.ClientIP(),
		"SUCCESS",
		map[string]interface{}{
			"operation": "qr_code_download",
			"filename":  filename,
			"endpoint":  "/api/documents/" + documentID + "/qr-code",
		},
	)

	// Set headers for file download
	c.Header("Content-Type", "image/png")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(qrCodeImage)))

	// Return QR code image
	c.Data(http.StatusOK, "image/png", qrCodeImage)
}

// DownloadSignedPDF handles GET /api/documents/:id/download
func (h *DocumentHandler) DownloadSignedPDF(c *gin.Context) {
	// Get user ID from authentication context
	userID, exists := c.Get("user_id")
	if !exists {
		RespondWithUnauthorizedError(c, "User not authenticated")
		return
	}

	// Validate user ID format
	if _, validationErr := h.validator.ValidateUUID("user_id", userID.(string), true); validationErr != nil {
		RespondWithValidationError(c, "Invalid user ID", validationErr.Error())
		return
	}

	// Get and validate document ID from URL parameter
	documentID := c.Param("id")
	if _, validationErr := h.validator.ValidateUUID("document_id", documentID, true); validationErr != nil {
		RespondWithValidationError(c, "Invalid document ID", validationErr.Error())
		return
	}

	// Get user info for logging
	user, _ := c.Get("user")
	authUser := user.(*services.AuthenticatedUser)

	// Get signed PDF
	pdfData, filename, err := h.documentService.GetSignedPDF(c.Request.Context(), userID.(string), documentID)
	if err != nil {
		// Log failed PDF download attempt
		logging.LogDocumentOperation(
			logging.AuditEventDocumentView,
			authUser.ID,
			authUser.Username,
			documentID,
			c.ClientIP(),
			"FAILURE",
			map[string]interface{}{
				"operation": "pdf_download",
				"error":     err.Error(),
				"endpoint":  "/api/documents/" + documentID + "/download",
			},
		)
		MapServiceErrorToHTTP(c, err)
		return
	}

	// Log successful PDF download
	logging.LogDocumentOperation(
		logging.AuditEventDocumentView,
		authUser.ID,
		authUser.Username,
		documentID,
		c.ClientIP(),
		"SUCCESS",
		map[string]interface{}{
			"operation": "pdf_download",
			"filename":  filename,
			"file_size": len(pdfData),
			"endpoint":  "/api/documents/" + documentID + "/download",
		},
	)

	// Set headers for file download
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfData)))

	// Return PDF file
	c.Data(http.StatusOK, "application/pdf", pdfData)
}
