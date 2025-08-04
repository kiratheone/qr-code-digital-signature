package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"digital-signature-system/internal/domain/services"
	"digital-signature-system/internal/infrastructure/logging"
	"digital-signature-system/internal/infrastructure/validation"
)

type AuthMiddleware struct {
	authService *services.AuthService
	rateLimiter *rate.Limiter
	logger      *logging.Logger
	validator   *validation.Validator
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		rateLimiter: rate.NewLimiter(rate.Every(time.Second), 100), // 100 requests per second
		logger:      logging.GetLogger(),
		validator:   validation.NewValidator(),
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
			
			// Log security event for rate limiting
			logging.LogSecurityEvent(
				logging.AuditEventRateLimitExceeded,
				c.ClientIP(),
				c.GetHeader("User-Agent"),
				"Rate limit exceeded",
				map[string]interface{}{
					"endpoint": c.Request.URL.Path,
					"method": c.Request.Method,
				},
			)
			
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

// InputValidation middleware for comprehensive input sanitization
func (m *AuthMiddleware) InputValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate and sanitize common headers
		userAgent := c.GetHeader("User-Agent")
		if userAgent != "" {
			if _, err := m.validator.ValidateAndSanitizeString("user-agent", userAgent, 0, 500, false); err != nil {
				m.logger.Warn("Suspicious User-Agent detected from IP %s: %s", c.ClientIP(), userAgent)
				
				// Log security event for suspicious user agent
				logging.LogSecurityEvent(
					logging.AuditEventSuspiciousActivity,
					c.ClientIP(),
					userAgent,
					"Suspicious User-Agent detected",
					map[string]interface{}{
						"validation_error": err.Error(),
						"endpoint": c.Request.URL.Path,
						"method": c.Request.Method,
					},
				)
				
				RespondWithValidationError(c, "Invalid request headers", err.Error())
				c.Abort()
				return
			}
		}

		// Validate Content-Type header for security
		contentType := c.GetHeader("Content-Type")
		if contentType != "" {
			if _, err := m.validator.ValidateAndSanitizeString("content-type", contentType, 0, 100, false); err != nil {
				m.logger.Warn("Suspicious Content-Type detected from IP %s: %s", c.ClientIP(), contentType)
				RespondWithValidationError(c, "Invalid content type", err.Error())
				c.Abort()
				return
			}
		}

		// Validate and sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			// Validate parameter name
			if _, err := m.validator.ValidateAndSanitizeString("query-param-name", key, 0, 50, false); err != nil {
				m.logger.Warn("Suspicious query parameter name detected from IP %s: %s", c.ClientIP(), key)
				RespondWithValidationError(c, "Invalid query parameter name", err.Error())
				c.Abort()
				return
			}

			// Validate parameter values
			for _, value := range values {
				if _, err := m.validator.ValidateAndSanitizeString("query-param-value", value, 0, 200, false); err != nil {
					m.logger.Warn("Suspicious query parameter value detected from IP %s: %s=%s", 
						c.ClientIP(), key, value)
					
					// Log security event for suspicious query parameters
					logging.LogSecurityEvent(
						logging.AuditEventSuspiciousActivity,
						c.ClientIP(),
						c.GetHeader("User-Agent"),
						"Suspicious query parameter detected",
						map[string]interface{}{
							"parameter_name": key,
							"parameter_value": value,
							"validation_error": err.Error(),
							"endpoint": c.Request.URL.Path,
							"method": c.Request.Method,
						},
					)
					
					RespondWithValidationError(c, "Invalid query parameter value", err.Error())
					c.Abort()
					return
				}
			}
		}

		// Validate URL path parameters
		for _, param := range c.Params {
			if _, err := m.validator.ValidateAndSanitizeString("path-param", param.Value, 0, 100, false); err != nil {
				m.logger.Warn("Suspicious path parameter detected from IP %s: %s=%s", 
					c.ClientIP(), param.Key, param.Value)
				RespondWithValidationError(c, "Invalid path parameter", err.Error())
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// SecurityHeaders middleware adds comprehensive security headers
func (m *AuthMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
		c.Header("Content-Security-Policy", csp)
		
		// Strict Transport Security (HTTPS only)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		
		// Permissions Policy (formerly Feature Policy)
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		
		// Remove server information
		c.Header("Server", "")
		
		// Prevent caching of sensitive responses
		if strings.Contains(c.Request.URL.Path, "/api/") {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		
		c.Next()
	}
}

// FileValidation middleware for comprehensive file upload validation
func (m *AuthMiddleware) FileValidation(maxSize int64, allowedTypes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to multipart form requests
		if !strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
			c.Next()
			return
		}

		// Validate content length
		if err := m.validator.ValidateFileSize("request", c.Request.ContentLength, maxSize); err != nil {
			m.logger.Warn("File validation failed from IP %s: %v", c.ClientIP(), err)
			RespondWithError(c, http.StatusRequestEntityTooLarge,
				NewStandardError(ErrCodeFileTooLarge, err.Message))
			c.Abort()
			return
		}

		// Parse multipart form to validate file
		err := c.Request.ParseMultipartForm(maxSize)
		if err != nil {
			m.logger.Error("Failed to parse multipart form from IP %s: %v", c.ClientIP(), err)
			RespondWithValidationError(c, "Invalid form data", err.Error())
			c.Abort()
			return
		}

		// Comprehensive file validation if files are present
		if c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
			for fieldName, fileHeaders := range c.Request.MultipartForm.File {
				for _, fileHeader := range fileHeaders {
					// Validate filename
					if sanitizedFilename, err := m.validator.ValidateFilename("filename", fileHeader.Filename, true); err != nil {
						m.logger.Warn("Invalid filename from IP %s: %s - %v", c.ClientIP(), fileHeader.Filename, err)
						RespondWithValidationError(c, "Invalid filename", err.Error())
						c.Abort()
						return
					} else {
						// Store sanitized filename for later use
						c.Set("sanitized_filename_"+fieldName, sanitizedFilename)
					}

					// Validate content type
					contentType := fileHeader.Header.Get("Content-Type")
					if err := m.validator.ValidateContentType("content_type", contentType, allowedTypes); err != nil {
						m.logger.Warn("Invalid file type from IP %s: %s", c.ClientIP(), contentType)
						RespondWithError(c, http.StatusBadRequest,
							NewStandardError(ErrCodeInvalidFile, err.Message))
						c.Abort()
						return
					}

					// Validate file size
					if err := m.validator.ValidateFileSize("file", fileHeader.Size, maxSize); err != nil {
						m.logger.Warn("File size validation failed from IP %s: %v", c.ClientIP(), err)
						RespondWithError(c, http.StatusRequestEntityTooLarge,
							NewStandardError(ErrCodeFileTooLarge, err.Message))
						c.Abort()
						return
					}

					// Additional security checks
					if fileHeader.Size == 0 {
						m.logger.Warn("Empty file uploaded from IP %s", c.ClientIP())
						RespondWithValidationError(c, "Empty files are not allowed")
						c.Abort()
						return
					}

					// Check for suspicious file extensions in filename
					if m.hasSuspiciousExtension(fileHeader.Filename) {
						m.logger.Warn("Suspicious file extension from IP %s: %s", c.ClientIP(), fileHeader.Filename)
						RespondWithValidationError(c, "File extension not allowed")
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}

// hasSuspiciousExtension checks for potentially dangerous file extensions
func (m *AuthMiddleware) hasSuspiciousExtension(filename string) bool {
	// Convert to lowercase for case-insensitive comparison
	lowerFilename := strings.ToLower(filename)
	
	// List of suspicious extensions that should not be allowed
	suspiciousExtensions := []string{
		".exe", ".bat", ".cmd", ".com", ".pif", ".scr", ".vbs", ".js",
		".jar", ".app", ".deb", ".pkg", ".dmg", ".rpm", ".msi",
		".sh", ".bash", ".zsh", ".fish", ".ps1", ".psm1",
		".php", ".asp", ".aspx", ".jsp", ".py", ".rb", ".pl",
		".sql", ".db", ".sqlite", ".mdb",
	}
	
	for _, ext := range suspiciousExtensions {
		if strings.HasSuffix(lowerFilename, ext) {
			return true
		}
	}
	
	// Check for double extensions (e.g., file.pdf.exe)
	parts := strings.Split(lowerFilename, ".")
	if len(parts) > 2 {
		for i := 1; i < len(parts)-1; i++ {
			for _, ext := range suspiciousExtensions {
				if "."+parts[i] == ext {
					return true
				}
			}
		}
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