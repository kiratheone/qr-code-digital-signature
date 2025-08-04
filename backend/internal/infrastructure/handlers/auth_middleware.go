package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"digital-signature-system/internal/domain/services"
	"digital-signature-system/internal/infrastructure/logging"
)

type AuthMiddleware struct {
	authService *services.AuthService
	rateLimiter *rate.Limiter
	logger      *logging.Logger
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		rateLimiter: rate.NewLimiter(rate.Every(time.Second), 100), // 100 requests per second
		logger:      logging.GetLogger(),
	}
}

// RequireAuth is a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractTokenFromHeader(c)
		if token == "" {
			RespondWithUnauthorizedError(c, "Authorization token is required")
			c.Abort()
			return
		}

		user, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			m.logger.Warn("Authentication failed for IP %s: %v", c.ClientIP(), err)
			MapServiceErrorToHTTP(c, err)
			c.Abort()
			return
		}

		// Convert to AuthenticatedUser and store in context
		authUser := &services.AuthenticatedUser{
			ID:       user.ID,
			Username: user.Username,
			FullName: user.FullName,
			Email:    user.Email,
			Role:     user.Role,
		}

		c.Set("user", authUser)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Next()
	}
}

// RequireRole is a middleware that requires a specific role
func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			RespondWithUnauthorizedError(c, "User not authenticated")
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok || role != requiredRole {
			RespondWithForbiddenError(c, "Insufficient permissions for this operation")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole is a middleware that requires any of the specified roles
func (m *AuthMiddleware) RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			RespondWithUnauthorizedError(c, "User not authenticated")
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			RespondWithForbiddenError(c, "Insufficient permissions for this operation")
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		for _, requiredRole := range roles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		RespondWithForbiddenError(c, "Insufficient permissions for this operation")
		c.Abort()
	}
}

// OptionalAuth is a middleware that optionally authenticates the user
// If a token is provided, it validates it, but doesn't require authentication
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractTokenFromHeader(c)
		if token == "" {
			c.Next()
			return
		}

		user, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err == nil && user != nil {
			// Convert to AuthenticatedUser and store in context
			authUser := &services.AuthenticatedUser{
				ID:       user.ID,
				Username: user.Username,
				FullName: user.FullName,
				Email:    user.Email,
				Role:     user.Role,
			}

			c.Set("user", authUser)
			c.Set("user_id", user.ID)
			c.Set("user_role", user.Role)
		}

		c.Next()
	}
}

// CORS middleware for handling cross-origin requests
func (m *AuthMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimiting middleware with simple rate limiting
func (m *AuthMiddleware) RateLimiting() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if !m.rateLimiter.Allow() {
			m.logger.Warn("Rate limit exceeded for IP: %s", c.ClientIP())
			RespondWithError(c, http.StatusTooManyRequests, 
				NewStandardError(ErrCodeRateLimitExceeded, "Too many requests"))
			c.Abort()
			return
		}
		c.Next()
	})
}

// RequestLogging middleware for logging requests
func (m *AuthMiddleware) RequestLogging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log to our custom logger
		if param.StatusCode >= 400 {
			m.logger.Error("HTTP %d - %s %s - IP: %s - Latency: %s - Error: %s",
				param.StatusCode, param.Method, param.Path, param.ClientIP, 
				param.Latency, param.ErrorMessage)
		} else {
			m.logger.Info("HTTP %d - %s %s - IP: %s - Latency: %s",
				param.StatusCode, param.Method, param.Path, param.ClientIP, param.Latency)
		}

		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// InputValidation middleware for basic input sanitization
func (m *AuthMiddleware) InputValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize common headers
		userAgent := c.GetHeader("User-Agent")
		if userAgent != "" && containsSuspiciousContent(userAgent) {
			m.logger.Warn("Suspicious User-Agent detected from IP %s: %s", c.ClientIP(), userAgent)
			RespondWithValidationError(c, "Invalid request headers")
			c.Abort()
			return
		}

		// Check for suspicious query parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if containsSuspiciousContent(value) {
					m.logger.Warn("Suspicious query parameter detected from IP %s: %s=%s", 
						c.ClientIP(), key, value)
					RespondWithValidationError(c, "Invalid query parameters")
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// SecurityHeaders middleware adds basic security headers
func (m *AuthMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		
		// Remove server information
		c.Header("Server", "")
		
		c.Next()
	}
}

// FileValidation middleware for validating file uploads
func (m *AuthMiddleware) FileValidation(maxSize int64, allowedTypes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to multipart form requests
		if !strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
			c.Next()
			return
		}

		// Check content length
		if c.Request.ContentLength > maxSize {
			m.logger.Warn("File too large from IP %s: %d bytes", c.ClientIP(), c.Request.ContentLength)
			RespondWithError(c, http.StatusRequestEntityTooLarge,
				NewStandardError(ErrCodeFileTooLarge, "File size exceeds maximum limit"))
			c.Abort()
			return
		}

		// Parse multipart form to validate file
		err := c.Request.ParseMultipartForm(maxSize)
		if err != nil {
			m.logger.Error("Failed to parse multipart form from IP %s: %v", c.ClientIP(), err)
			RespondWithValidationError(c, "Invalid form data")
			c.Abort()
			return
		}

		// Validate file types if files are present
		if c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
			for _, fileHeaders := range c.Request.MultipartForm.File {
				for _, fileHeader := range fileHeaders {
					contentType := fileHeader.Header.Get("Content-Type")
					if !isAllowedFileType(contentType, allowedTypes) {
						m.logger.Warn("Invalid file type from IP %s: %s", c.ClientIP(), contentType)
						RespondWithError(c, http.StatusBadRequest,
							NewStandardError(ErrCodeInvalidFile, "File type not allowed"))
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}

// containsSuspiciousContent checks for common attack patterns
func containsSuspiciousContent(input string) bool {
	suspiciousPatterns := []string{
		"<script",
		"javascript:",
		"vbscript:",
		"onload=",
		"onerror=",
		"eval(",
		"alert(",
		"document.cookie",
		"../",
		"..\\",
		"union select",
		"drop table",
		"insert into",
		"delete from",
		"update set",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	// Check for excessive length
	if len(input) > 1000 {
		return true
	}

	// Check for suspicious characters
	suspiciousChars := regexp.MustCompile(`[<>'"&;]`)
	if suspiciousChars.MatchString(input) {
		return true
	}

	return false
}

// isAllowedFileType checks if the file type is in the allowed list
func isAllowedFileType(contentType string, allowedTypes []string) bool {
	if len(allowedTypes) == 0 {
		return true // No restrictions
	}

	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}