package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"digital-signature-system/internal/config"
	"digital-signature-system/internal/domain/services"
	"digital-signature-system/internal/infrastructure/database"
)

type Server struct {
	config         *config.Config
	db             *gorm.DB
	router         *gin.Engine
	authService    *services.AuthService
	authHandler    *AuthHandler
	authMiddleware *AuthMiddleware
}

func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	// Initialize repositories
	userRepo := database.NewUserRepository(db)
	sessionRepo := database.NewSessionRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, sessionRepo, cfg.JWTSecret)

	// Initialize handlers and middleware
	authHandler := NewAuthHandler(authService)
	authMiddleware := NewAuthMiddleware(authService)

	server := &Server{
		config:         cfg,
		db:             db,
		router:         gin.Default(),
		authService:    authService,
		authHandler:    authHandler,
		authMiddleware: authMiddleware,
	}

	server.setupMiddleware()
	server.setupRoutes()
	return server
}

func (s *Server) setupMiddleware() {
	// Add CORS middleware
	s.router.Use(s.authMiddleware.CORS())

	// Add request logging middleware
	s.router.Use(s.authMiddleware.RequestLogging())

	// Add rate limiting middleware
	s.router.Use(s.authMiddleware.RateLimiting())
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := s.router.Group("/api")
	{
		// Public routes (no authentication required)
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", s.authHandler.Login)
			auth.POST("/register", s.authHandler.Register)
			auth.POST("/logout", s.authHandler.Logout)
		}

		// Protected routes (authentication required)
		protected := api.Group("/")
		protected.Use(s.authMiddleware.RequireAuth())
		{
			// User profile routes
			protected.GET("/profile", s.authHandler.GetProfile)
			protected.POST("/change-password", s.authHandler.ChangePassword)

			// Document routes will be added in subsequent tasks
			// documents := protected.Group("/documents")
			// {
			//     documents.POST("/sign", documentHandler.SignDocument)
			//     documents.GET("/", documentHandler.GetDocuments)
			//     documents.GET("/:id", documentHandler.GetDocument)
			//     documents.DELETE("/:id", documentHandler.DeleteDocument)
			// }
		}

		// Public verification routes (no authentication required)
		// verify := api.Group("/verify")
		// {
		//     verify.GET("/:docId", verificationHandler.GetVerificationInfo)
		//     verify.POST("/:docId/upload", verificationHandler.VerifyDocument)
		// }
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}