package routes

import (
	"digital-signature-system/internal/infrastructure/server/handlers"
	"digital-signature-system/internal/infrastructure/server/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes sets up authentication routes
func SetupAuthRoutes(router *gin.RouterGroup, authHandler *handlers.AuthHandler, authMiddleware *middleware.AuthMiddleware) {
	// Apply CORS middleware to all routes
	router.Use(authMiddleware.CORS())
	
	// Create validation middleware
	validationMiddleware := middleware.NewValidationMiddleware()
	
	auth := router.Group("/auth")
	{
		// Public routes with rate limiting
		public := auth.Group("")
		public.Use(authMiddleware.RateLimiter(100, time.Minute)) // 100 requests per minute
		{
			public.POST("/register", 
				validationMiddleware.ValidateJSON(&handlers.RegisterRequest{}),
				authHandler.Register)
			public.POST("/login", 
				validationMiddleware.ValidateJSON(&handlers.LoginRequest{}),
				authHandler.Login)
			public.POST("/refresh", 
				validationMiddleware.ValidateJSON(&handlers.RefreshRequest{}),
				authHandler.RefreshToken)
			public.GET("/validate", authHandler.ValidateToken)
		}
		
		// Protected routes
		protected := auth.Group("")
		protected.Use(authMiddleware.Authenticate())
		{
			protected.POST("/logout", authHandler.Logout)
			protected.GET("/me", authHandler.Me)
			protected.POST("/change-password", authHandler.ChangePassword)
			protected.POST("/logout-all", authHandler.LogoutAll)
		}
		
		// Admin routes
		admin := auth.Group("/admin")
		admin.Use(authMiddleware.Authenticate(), authMiddleware.RequireRole("admin"))
		{
			// Admin-specific routes would go here
			// For example: user management, system settings, etc.
		}
	}
}