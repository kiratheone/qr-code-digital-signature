package server

import (
	"fmt"
	"os"
	"time"

	"digital-signature-system/internal/config"
	"digital-signature-system/internal/infrastructure/di"
	"digital-signature-system/internal/infrastructure/server/handlers"
	"digital-signature-system/internal/infrastructure/server/middleware"
	"digital-signature-system/internal/infrastructure/server/routes"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	config              *config.Config
	db                  *gorm.DB
	router              *gin.Engine
	container           *di.Container
	errorMiddleware     *middleware.ErrorMiddleware
	rateLimitMiddleware *middleware.RateLimitMiddleware
	loggingMiddleware   *middleware.LoggingMiddleware
	monitoringMiddleware *middleware.MonitoringMiddleware
	auditLogger         *middleware.AuditLogger
}

func New(cfg *config.Config, db *gorm.DB, container *di.Container) *Server {
	// Create structured logger
	logger := middleware.NewStructuredLogger(middleware.LogLevelInfo, os.Stdout)
	
	// Create middleware instances
	errorMiddleware := middleware.NewErrorMiddleware(logger)
	
	// Create rate limit middleware with monitoring integration
	rateLimitMiddleware := middleware.NewRateLimitMiddleware()
	
	// Create monitoring middleware
	monitoringMiddleware := middleware.NewMonitoringMiddleware(container.MonitoringService())
	
	loggingMiddleware := middleware.NewLoggingMiddleware(middleware.LoggingConfig{
		Logger:          logger,
		SkipPaths:       []string{"/health", "/metrics"},
		LogRequestBody:  false, // Enable for debugging
		LogResponseBody: false, // Enable for debugging
		MaxBodySize:     1024,  // 1KB for body logging
	})
	auditLogger := middleware.NewAuditLogger(logger)

	// Create router without default middleware
	router := gin.New()

	return &Server{
		config:               cfg,
		db:                   db,
		router:               router,
		container:            container,
		errorMiddleware:      errorMiddleware,
		rateLimitMiddleware:  rateLimitMiddleware,
		loggingMiddleware:    loggingMiddleware,
		monitoringMiddleware: monitoringMiddleware,
		auditLogger:          auditLogger,
	}
}

func (s *Server) Start() error {
	s.setupMiddleware()
	s.setupRoutes()
	
	addr := fmt.Sprintf("%s:%s", s.config.Server.Host, s.config.Server.Port)
	return s.router.Run(addr)
}

func (s *Server) setupMiddleware() {
	// Recovery and error handling (must be first)
	s.router.Use(s.errorMiddleware.ErrorHandler())
	
	// Request ID middleware
	s.router.Use(s.errorMiddleware.RequestIDMiddleware())
	
	// Enhanced security headers
	s.router.Use(s.setupSecurityHeaders())
	
	// Monitoring middleware (early to track all requests)
	s.router.Use(s.monitoringMiddleware.Middleware())
	
	// Comprehensive security validation
	s.router.Use(middleware.ValidateAPIRequest())
	
	// Request validation and sanitization (comprehensive)
	s.router.Use(s.errorMiddleware.RequestValidationMiddleware())
	s.router.Use(s.errorMiddleware.RequestSanitizationMiddleware())
	
	// Input sanitization middleware
	validationMiddleware := middleware.NewValidationMiddleware()
	s.router.Use(validationMiddleware.SanitizeAllInputs())
	
	// Logging middleware
	s.router.Use(s.loggingMiddleware.Middleware())
	
	// Global rate limiting with monitoring integration
	s.router.Use(s.createRateLimitWithMonitoring(s.rateLimitMiddleware.GlobalLimit()))
	s.router.Use(s.createRateLimitWithMonitoring(s.rateLimitMiddleware.IPLimit()))
	
	// CORS middleware
	authMiddleware := middleware.NewAuthMiddleware(s.container.AuthUseCase())
	s.router.Use(authMiddleware.CORS())
	
	// CSRF protection (skip for API endpoints)
	csrfMiddleware := middleware.NewCSRFMiddleware(middleware.CSRFConfig{
		SecureCookie: s.config.Server.Environment == "production",
		SkipFunc:     middleware.SkipCSRFForPublicEndpoints,
	})
	s.router.Use(csrfMiddleware.Middleware())
	
	// Validation error handling
	s.router.Use(s.errorMiddleware.ValidationErrorHandler())
	
	// General error handling (must be last)
	s.router.Use(s.errorMiddleware.HandleError())
}

func (s *Server) setupRoutes() {
	// Health check (no rate limiting)
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now(),
			"version":   "1.0.0",
		})
	})

	// Monitoring and metrics endpoints
	monitoringHandler := handlers.NewMonitoringHandler(s.container.AuditService(), s.container.MonitoringService())
	
	// Public monitoring endpoints (with rate limiting)
	s.router.GET("/health/detailed", s.rateLimitMiddleware.EndpointLimit(), monitoringHandler.GetHealthStatus)
	s.router.GET("/metrics", s.rateLimitMiddleware.EndpointLimit(), monitoringHandler.GetMetrics)

	// API routes
	api := s.router.Group("/api")
	{
		// Basic ping endpoint
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})
		
		// Setup auth routes with strict rate limiting
		authHandler := handlers.NewAuthHandler(s.container.AuthUseCase())
		authMiddleware := middleware.NewAuthMiddleware(s.container.AuthUseCase())
		
		// Auth routes with special rate limiting
		authGroup := api.Group("/auth")
		authGroup.Use(s.createRateLimitWithMonitoring(s.rateLimitMiddleware.AuthLimit())) // Strict rate limiting for auth
		routes.SetupAuthRoutes(authGroup, authHandler, authMiddleware)
		
		// Document routes with user-based rate limiting
		docGroup := api.Group("/documents")
		docGroup.Use(s.createRateLimitWithMonitoring(s.rateLimitMiddleware.UserLimit())) // User-based rate limiting
		documentHandler := handlers.NewDocumentHandler(s.container.DocumentUseCase())
		routes.SetupDocumentRoutes(docGroup, documentHandler, authMiddleware)
		
		// Verification routes (public, with endpoint-specific rate limiting)
		verifyGroup := s.router.Group("/verify")
		verifyGroup.Use(s.createRateLimitWithMonitoring(s.rateLimitMiddleware.EndpointLimit())) // Endpoint-specific rate limiting
		routes.SetupVerificationRoutes(verifyGroup, s.container)
		
		// Protected routes
		protected := api.Group("")
		protected.Use(authMiddleware.Authenticate())
		protected.Use(s.createRateLimitWithMonitoring(s.rateLimitMiddleware.UserLimit()))
		{
			// Additional protected routes can be added here
		}
		
		// Admin routes with strict rate limiting
		admin := api.Group("/admin")
		admin.Use(authMiddleware.Authenticate(), authMiddleware.RequireRole("admin"))
		admin.Use(s.createRateLimitWithMonitoring(s.rateLimitMiddleware.StrictLimit())) // Very strict rate limiting for admin
		{
			// Admin monitoring endpoints
			admin.GET("/monitoring/stats", monitoringHandler.GetSystemStats)
			admin.GET("/monitoring/performance", monitoringHandler.GetPerformanceMetrics)
			admin.GET("/monitoring/alerts", monitoringHandler.GetSecurityAlerts)
			admin.PUT("/monitoring/alerts/:alertId/resolve", monitoringHandler.ResolveSecurityAlert)
			admin.GET("/monitoring/audit/stats", monitoringHandler.GetAuditStats)
			admin.GET("/monitoring/metrics/:metricName/history", monitoringHandler.GetMetricHistory)
			admin.POST("/monitoring/alerts/test", monitoringHandler.CreateTestAlert)
		}
	}
}

// setupSecurityHeaders configures comprehensive security headers
func (s *Server) setupSecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: blob:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"media-src 'self'; " +
			"object-src 'none'; " +
			"child-src 'none'; " +
			"worker-src 'none'; " +
			"frame-ancestors 'none'; " +
			"form-action 'self'; " +
			"base-uri 'self'"
		c.Header("Content-Security-Policy", csp)
		
		// HSTS (only in production with HTTPS)
		if s.config.Server.Environment == "production" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		
		// Permissions Policy
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		
		// Remove server information
		c.Header("Server", "")
		
		c.Next()
	}
}

// createRateLimitWithMonitoring wraps rate limit middleware with monitoring integration
func (s *Server) createRateLimitWithMonitoring(rateLimitHandler gin.HandlerFunc) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Store original status for comparison
		originalStatus := c.Writer.Status()
		
		// Execute rate limit handler
		rateLimitHandler(c)
		
		// Check if rate limit was triggered (status changed to 429)
		if c.Writer.Status() == 429 && originalStatus != 429 {
			// Rate limit was triggered, track it in monitoring
			if s.container.MonitoringService() != nil {
				ctx := c.Request.Context()
				s.container.MonitoringService().TrackRateLimitViolation(ctx, c.ClientIP())
			}
		}
	})
}

// Stop gracefully stops the server
func (s *Server) Stop() {
	if s.rateLimitMiddleware != nil {
		s.rateLimitMiddleware.Stop()
	}
	
	// Close audit and monitoring services
	if s.container != nil {
		s.container.Close()
	}
}