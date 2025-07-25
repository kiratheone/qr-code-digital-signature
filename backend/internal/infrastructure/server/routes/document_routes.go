package routes

import (
	"digital-signature-system/internal/infrastructure/server/handlers"
	"digital-signature-system/internal/infrastructure/server/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupDocumentRoutes sets up document routes
func SetupDocumentRoutes(router *gin.RouterGroup, documentHandler *handlers.DocumentHandler, authMiddleware *middleware.AuthMiddleware) {
	// Apply CORS middleware to all routes
	router.Use(authMiddleware.CORS())
	
	// Create validation middleware
	validationMiddleware := middleware.NewValidationMiddleware()
	
	// Document routes
	docs := router.Group("/documents")
	{
		// Protected routes that require authentication
		docs.Use(authMiddleware.Authenticate())
		{
			// Document signing with comprehensive validation
			docs.POST("/sign", 
				middleware.ValidateFileUploadRequest(), // File upload security validation
				validationMiddleware.ValidateMultipartForm(50*1024*1024), // 50MB max
				validationMiddleware.ValidatePDFUpload("pdf", 50*1024*1024, true), // Required PDF
				validationMiddleware.ValidateForm(&handlers.SignDocumentRequest{}),
				documentHandler.SignDocument)
			
			// Document management with query validation
			docs.GET("", 
				middleware.ValidateDocumentQuery(),
				documentHandler.GetDocuments)
			docs.GET("/:id", 
				middleware.ValidateDocumentID(),
				documentHandler.GetDocumentByID)
			docs.DELETE("/:id", 
				middleware.ValidateDocumentID(),
				documentHandler.DeleteDocument)
		}
	}
}