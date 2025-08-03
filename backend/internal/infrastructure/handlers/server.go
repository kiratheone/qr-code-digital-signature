package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"digital-signature-system/internal/config"
	"digital-signature-system/internal/domain/services"
	"digital-signature-system/internal/infrastructure/crypto"
	"digital-signature-system/internal/infrastructure/database"
	"digital-signature-system/internal/infrastructure/pdf"
)

type Server struct {
	config              *config.Config
	db                  *gorm.DB
	router              *gin.Engine
	authService         *services.AuthService
	documentService     *services.DocumentService
	verificationService *services.VerificationService
	authHandler         *AuthHandler
	documentHandler     *DocumentHandler
	verificationHandler *VerificationHandler
	authMiddleware      *AuthMiddleware
}

func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	// Initialize repositories
	userRepo := database.NewUserRepository(db)
	sessionRepo := database.NewSessionRepository(db)
	documentRepo := database.NewDocumentRepository(db)
	verificationLogRepo := database.NewVerificationLogRepository(db)

	// Initialize crypto services
	keyManager, err := crypto.NewKeyManager()
	if err != nil {
		panic("Failed to initialize key manager: " + err.Error())
	}

	signatureService, err := crypto.NewSignatureServiceFromKeyManager(keyManager)
	if err != nil {
		panic("Failed to initialize signature service: " + err.Error())
	}

	pdfService := pdf.NewPDFService()

	// Initialize services
	authService := services.NewAuthService(userRepo, sessionRepo, cfg.JWTSecret)
	documentService := services.NewDocumentService(documentRepo, signatureService, pdfService)
	verificationService := services.NewVerificationService(documentRepo, verificationLogRepo, signatureService, pdfService, documentService)

	// Initialize handlers and middleware
	authHandler := NewAuthHandler(authService)
	documentHandler := NewDocumentHandler(documentService)
	verificationHandler := NewVerificationHandler(verificationService)
	authMiddleware := NewAuthMiddleware(authService)

	server := &Server{
		config:              cfg,
		db:                  db,
		router:              gin.Default(),
		authService:         authService,
		documentService:     documentService,
		verificationService: verificationService,
		authHandler:         authHandler,
		documentHandler:     documentHandler,
		verificationHandler: verificationHandler,
		authMiddleware:      authMiddleware,
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

			// Document routes
			documents := protected.Group("/documents")
			{
				documents.POST("/sign", s.documentHandler.SignDocument)
				documents.GET("/", s.documentHandler.GetDocuments)
				documents.GET("/:id", s.documentHandler.GetDocument)
				documents.DELETE("/:id", s.documentHandler.DeleteDocument)
			}
		}

		// Public verification routes (no authentication required)
		verify := api.Group("/verify")
		{
			verify.GET("/:docId", s.verificationHandler.GetVerificationInfo)
			verify.POST("/:docId/upload", s.verificationHandler.VerifyDocument)
			verify.GET("/:docId/history", s.verificationHandler.GetVerificationHistory)
		}
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}