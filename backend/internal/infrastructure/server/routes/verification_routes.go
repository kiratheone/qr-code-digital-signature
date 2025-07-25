package routes

import (
	"digital-signature-system/internal/infrastructure/di"
	"digital-signature-system/internal/infrastructure/server/handlers"
	"digital-signature-system/internal/infrastructure/server/middleware"

	"github.com/gin-gonic/gin"
)

// SetupVerificationRoutes sets up verification routes
func SetupVerificationRoutes(router *gin.Engine, container *di.Container) {
	// Create verification handler
	verificationHandler := handlers.NewVerificationHandler(container.VerificationUseCase())
	
	// Create validation middleware
	validationMiddleware := middleware.NewValidationMiddleware()

	// Public verification routes (no authentication required)
	verifyGroup := router.Group("/api/verify")
	{
		// GET /api/verify/:docId - Get verification info with UUID validation
		verifyGroup.GET("/:docId", 
			validationMiddleware.ValidatePathParams(map[string]middleware.PathParamConfig{
				"docId": {
					Type:      "uuid",
					MinLength: 36,
					MaxLength: 36,
				},
			}),
			verificationHandler.GetVerificationInfo)

		// POST /api/verify/:docId/upload - Verify document with comprehensive validation
		verifyGroup.POST("/:docId/upload", 
			middleware.ValidateFileUploadRequest(), // File upload security validation
			validationMiddleware.ValidatePathParams(map[string]middleware.PathParamConfig{
				"docId": {
					Type:      "uuid",
					MinLength: 36,
					MaxLength: 36,
				},
			}),
			validationMiddleware.ValidateMultipartForm(50*1024*1024), // 50MB max
			middleware.ValidateVerificationFile(), // PDF validation
			verificationHandler.VerifyDocument)
	}
}