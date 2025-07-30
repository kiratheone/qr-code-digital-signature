package middleware

import (
	"digital-signature-system/internal/infrastructure/validation"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityMiddleware provides comprehensive security validation
type SecurityMiddleware struct {
	validator *validation.Validator
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware() *SecurityMiddleware {
	return &SecurityMiddleware{
		validator: validation.NewValidator(),
	}
}

// ValidateSecurityHeaders validates and sanitizes security-sensitive headers
func (sm *SecurityMiddleware) ValidateSecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate User-Agent
		userAgent := c.GetHeader("User-Agent")
		if userAgent != "" && !sm.validator.ValidateUserAgent(userAgent) {
			AbortWithValidationError(c, "Invalid User-Agent header", "Suspicious User-Agent detected")
			return
		}

		// Validate Referer
		referer := c.GetHeader("Referer")
		if referer != "" && !sm.validator.ValidateReferer(referer) {
			AbortWithValidationError(c, "Invalid Referer header", "Suspicious Referer detected")
			return
		}

		// Validate Content-Length
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if !sm.validator.ValidateContentLength(c.Request.ContentLength, 50*1024*1024) { // 50MB max
				AbortWithValidationError(c, "Invalid Content-Length", "Content-Length exceeds maximum allowed size")
				return
			}
		}

		// Sanitize custom headers
		for headerName, headerValues := range c.Request.Header {
			// Skip standard headers
			if sm.isStandardHeader(headerName) {
				continue
			}

			for i, headerValue := range headerValues {
				sanitized := sm.validator.SanitizeHeader(headerValue)
				c.Request.Header[headerName][i] = sanitized
			}
		}

		c.Next()
	}
}

// DetectMaliciousInput detects various types of malicious input
func (sm *SecurityMiddleware) DetectMaliciousInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check query parameters for malicious content
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if sm.isMaliciousInput(value) {
					sm.logSecurityEvent(c, "Malicious query parameter detected", map[string]interface{}{
						"parameter": key,
						"value":     value,
					})
					AbortWithValidationError(c, "Malicious input detected", "Request contains potentially dangerous content")
					return
				}
			}
		}

		// Check headers for malicious content (except standard ones)
		for headerName, headerValues := range c.Request.Header {
			if sm.isStandardHeader(headerName) {
				continue
			}

			for _, headerValue := range headerValues {
				if sm.isMaliciousInput(headerValue) {
					sm.logSecurityEvent(c, "Malicious header detected", map[string]interface{}{
						"header": headerName,
						"value":  headerValue,
					})
					AbortWithValidationError(c, "Malicious input detected", "Request headers contain potentially dangerous content")
					return
				}
			}
		}

		c.Next()
	}
}

// ValidateRequestSize validates overall request size
func (sm *SecurityMiddleware) ValidateRequestSize(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			AbortWithValidationError(c, "Request too large", "Request exceeds maximum allowed size")
			return
		}
		c.Next()
	}
}

// ValidateHTTPMethod validates HTTP methods
func (sm *SecurityMiddleware) ValidateHTTPMethod() gin.HandlerFunc {
	allowedMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
	}

	return func(c *gin.Context) {
		if !allowedMethods[c.Request.Method] {
			AbortWithValidationError(c, "Method not allowed", "HTTP method not supported")
			return
		}
		c.Next()
	}
}

// ValidateContentType validates content types for specific methods
func (sm *SecurityMiddleware) ValidateContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType == "" {
				// Allow empty content type for some endpoints
				c.Next()
				return
			}

			// Extract main content type (ignore charset and other parameters)
			mainType := strings.Split(contentType, ";")[0]
			mainType = strings.TrimSpace(strings.ToLower(mainType))

			allowedTypes := map[string]bool{
				"application/json":                  true,
				"application/x-www-form-urlencoded": true,
				"multipart/form-data":               true,
				"application/pdf":                   true,
				"text/plain":                        true,
			}

			if !allowedTypes[mainType] {
				AbortWithValidationError(c, "Invalid content type", "Content type not supported")
				return
			}
		}
		c.Next()
	}
}

// PreventClickjacking adds anti-clickjacking headers
func (sm *SecurityMiddleware) PreventClickjacking() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "frame-ancestors 'none'")
		c.Next()
	}
}

// Helper methods

func (sm *SecurityMiddleware) isMaliciousInput(input string) bool {
	// Check for SQL injection
	if sm.validator.DetectSQLInjection(input) {
		return true
	}

	// Check for XSS
	if sm.validator.DetectXSS(input) {
		return true
	}

	// Check for path traversal
	if sm.validator.DetectPathTraversal(input) {
		return true
	}

	return false
}

func (sm *SecurityMiddleware) isStandardHeader(headerName string) bool {
	standardHeaders := map[string]bool{
		"Accept":             true,
		"Accept-Encoding":    true,
		"Accept-Language":    true,
		"Authorization":      true,
		"Cache-Control":      true,
		"Connection":         true,
		"Content-Length":     true,
		"Content-Type":       true,
		"Cookie":             true,
		"Host":               true,
		"If-Modified-Since":  true,
		"If-None-Match":      true,
		"Origin":             true,
		"Referer":            true,
		"User-Agent":         true,
		"X-Forwarded-For":    true,
		"X-Forwarded-Proto":  true,
		"X-Real-IP":          true,
		"X-Request-ID":       true,
		"X-CSRF-Token":       true,
	}

	return standardHeaders[headerName]
}

func (sm *SecurityMiddleware) logSecurityEvent(c *gin.Context, message string, details map[string]interface{}) {
	// Add request context to details
	details["ip"] = c.ClientIP()
	details["method"] = c.Request.Method
	details["path"] = c.Request.URL.Path
	details["user_agent"] = c.GetHeader("User-Agent")
	
	if requestID, exists := c.Get("request_id"); exists {
		details["request_id"] = requestID
	}

	// Log security event (in a real implementation, this might go to a security log)
	// For now, we'll use the standard logging mechanism
	// logger.Warn(c.Request.Context(), message, details)
}

// Convenience functions for common security validations

// ValidateAPIRequest provides comprehensive validation for API requests
func ValidateAPIRequest() gin.HandlerFunc {
	sm := NewSecurityMiddleware()
	return gin.HandlerFunc(func(c *gin.Context) {
		// Apply multiple security validations
		sm.ValidateHTTPMethod()(c)
		if c.IsAborted() {
			return
		}

		sm.ValidateRequestSize(50 * 1024 * 1024)(c) // 50MB max
		if c.IsAborted() {
			return
		}

		sm.ValidateContentType()(c)
		if c.IsAborted() {
			return
		}

		sm.ValidateSecurityHeaders()(c)
		if c.IsAborted() {
			return
		}

		sm.DetectMaliciousInput()(c)
		if c.IsAborted() {
			return
		}

		c.Next()
	})
}

// ValidateFileUploadRequest provides validation specifically for file uploads
func ValidateFileUploadRequest() gin.HandlerFunc {
	sm := NewSecurityMiddleware()
	return gin.HandlerFunc(func(c *gin.Context) {
		// Validate content type for file uploads
		contentType := c.GetHeader("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			AbortWithValidationError(c, "Invalid content type for file upload", "Expected multipart/form-data")
			return
		}

		// Apply other security validations
		sm.ValidateRequestSize(50 * 1024 * 1024)(c) // 50MB max for file uploads
		if c.IsAborted() {
			return
		}

		sm.ValidateSecurityHeaders()(c)
		if c.IsAborted() {
			return
		}

		c.Next()
	})
}