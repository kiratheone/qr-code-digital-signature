package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ErrorType represents different types of errors
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "VALIDATION_ERROR"
	ErrorTypeAuthentication ErrorType = "AUTHENTICATION_ERROR"
	ErrorTypeAuthorization  ErrorType = "AUTHORIZATION_ERROR"
	ErrorTypeNotFound      ErrorType = "NOT_FOUND"
	ErrorTypeConflict      ErrorType = "CONFLICT"
	ErrorTypeRateLimit     ErrorType = "RATE_LIMIT_EXCEEDED"
	ErrorTypeInternal      ErrorType = "INTERNAL_ERROR"
	ErrorTypeBadRequest    ErrorType = "BAD_REQUEST"
	ErrorTypeServiceUnavailable ErrorType = "SERVICE_UNAVAILABLE"
)

// APIError represents a standardized API error response
type APIError struct {
	Type        ErrorType `json:"type"`
	Code        string    `json:"code"`
	Message     string    `json:"message"`
	Details     string    `json:"details,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	RequestID   string    `json:"request_id"`
	Path        string    `json:"path"`
	Method      string    `json:"method"`
	StatusCode  int       `json:"status_code"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Type, e.Message)
}

// NewAPIError creates a new API error
func NewAPIError(errorType ErrorType, code, message, details string, statusCode int) *APIError {
	return &APIError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		Details:    details,
		Timestamp:  time.Now(),
		StatusCode: statusCode,
	}
}

// Common error constructors
func NewValidationError(message, details string) *APIError {
	return NewAPIError(ErrorTypeValidation, "VALIDATION_FAILED", message, details, http.StatusBadRequest)
}

func NewAuthenticationError(message string) *APIError {
	return NewAPIError(ErrorTypeAuthentication, "AUTHENTICATION_FAILED", message, "", http.StatusUnauthorized)
}

func NewAuthorizationError(message string) *APIError {
	return NewAPIError(ErrorTypeAuthorization, "AUTHORIZATION_FAILED", message, "", http.StatusForbidden)
}

func NewNotFoundError(resource string) *APIError {
	return NewAPIError(ErrorTypeNotFound, "RESOURCE_NOT_FOUND", fmt.Sprintf("%s not found", resource), "", http.StatusNotFound)
}

func NewConflictError(message, details string) *APIError {
	return NewAPIError(ErrorTypeConflict, "RESOURCE_CONFLICT", message, details, http.StatusConflict)
}

func NewRateLimitError(retryAfter int) *APIError {
	return &APIError{
		Type:       ErrorTypeRateLimit,
		Code:       "RATE_LIMIT_EXCEEDED",
		Message:    "Too many requests",
		Details:    fmt.Sprintf("Retry after %d seconds", retryAfter),
		Timestamp:  time.Now(),
		StatusCode: http.StatusTooManyRequests,
	}
}

func NewInternalError(message string) *APIError {
	return NewAPIError(ErrorTypeInternal, "INTERNAL_SERVER_ERROR", "Internal server error", message, http.StatusInternalServerError)
}

func NewBadRequestError(message, details string) *APIError {
	return NewAPIError(ErrorTypeBadRequest, "BAD_REQUEST", message, details, http.StatusBadRequest)
}

func NewServiceUnavailableError(message string) *APIError {
	return NewAPIError(ErrorTypeServiceUnavailable, "SERVICE_UNAVAILABLE", message, "", http.StatusServiceUnavailable)
}

// ErrorMiddleware provides comprehensive error handling
type ErrorMiddleware struct {
	logger Logger
}

// Logger interface for structured logging
type Logger interface {
	Error(ctx context.Context, msg string, fields map[string]interface{})
	Warn(ctx context.Context, msg string, fields map[string]interface{})
	Info(ctx context.Context, msg string, fields map[string]interface{})
}

// DefaultLogger implements a simple logger
type DefaultLogger struct{}

func (l *DefaultLogger) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	log.Printf("[ERROR] %s - %v", msg, fields)
}

func (l *DefaultLogger) Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	log.Printf("[WARN] %s - %v", msg, fields)
}

func (l *DefaultLogger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	log.Printf("[INFO] %s - %v", msg, fields)
}

// NewErrorMiddleware creates a new error middleware
func NewErrorMiddleware(logger Logger) *ErrorMiddleware {
	if logger == nil {
		logger = &DefaultLogger{}
	}
	return &ErrorMiddleware{
		logger: logger,
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func (m *ErrorMiddleware) RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// ErrorHandler middleware handles errors and panics
func (m *ErrorMiddleware) ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := m.getRequestID(c)
		
		// Log panic with stack trace
		m.logger.Error(c.Request.Context(), "Panic recovered", map[string]interface{}{
			"request_id": requestID,
			"panic":      recovered,
			"stack":      string(debug.Stack()),
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
			"ip":         c.ClientIP(),
		})

		// Create internal server error
		apiError := NewInternalError("An unexpected error occurred")
		apiError.RequestID = requestID
		apiError.Path = c.Request.URL.Path
		apiError.Method = c.Request.Method

		c.JSON(apiError.StatusCode, apiError)
		c.Abort()
	})
}

// HandleError processes errors and returns appropriate responses
func (m *ErrorMiddleware) HandleError() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			requestID := m.getRequestID(c)

			// Check if it's already an APIError
			if apiErr, ok := err.Err.(*APIError); ok {
				apiErr.RequestID = requestID
				apiErr.Path = c.Request.URL.Path
				apiErr.Method = c.Request.Method

				m.logError(c, apiErr)
				c.JSON(apiErr.StatusCode, apiErr)
				return
			}

			// Convert generic error to APIError
			apiError := m.convertToAPIError(err.Err)
			apiError.RequestID = requestID
			apiError.Path = c.Request.URL.Path
			apiError.Method = c.Request.Method

			m.logError(c, apiError)
			c.JSON(apiError.StatusCode, apiError)
		}
	}
}

// ValidationErrorHandler handles validation errors from gin binding
func (m *ErrorMiddleware) ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for binding errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				if err.Type == gin.ErrorTypeBind {
					requestID := m.getRequestID(c)
					
					apiError := NewValidationError("Invalid request data", err.Error())
					apiError.RequestID = requestID
					apiError.Path = c.Request.URL.Path
					apiError.Method = c.Request.Method

					m.logError(c, apiError)
					c.JSON(apiError.StatusCode, apiError)
					c.Abort()
					return
				}
			}
		}
	}
}

// SecurityHeadersMiddleware adds security headers
func (m *ErrorMiddleware) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// RequestSanitizationMiddleware sanitizes request data
func (m *ErrorMiddleware) RequestSanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize headers
		for key, values := range c.Request.Header {
			for i, value := range values {
				c.Request.Header[key][i] = m.sanitizeString(value)
			}
		}

		// Sanitize query parameters
		query := c.Request.URL.Query()
		for key, values := range query {
			for i, value := range values {
				query[key][i] = m.sanitizeString(value)
			}
		}
		c.Request.URL.RawQuery = query.Encode()

		c.Next()
	}
}

// RequestValidationMiddleware validates common request constraints
func (m *ErrorMiddleware) RequestValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate request size
		if c.Request.ContentLength > 50*1024*1024 { // 50MB limit
			apiError := NewBadRequestError("Request too large", "Maximum request size is 50MB")
			apiError.RequestID = m.getRequestID(c)
			apiError.Path = c.Request.URL.Path
			apiError.Method = c.Request.Method
			
			m.logError(c, apiError)
			c.JSON(apiError.StatusCode, apiError)
			c.Abort()
			return
		}

		// Validate content type for POST/PUT requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && !m.isValidContentType(contentType) {
				apiError := NewBadRequestError("Invalid content type", "Supported content types: application/json, multipart/form-data, application/pdf")
				apiError.RequestID = m.getRequestID(c)
				apiError.Path = c.Request.URL.Path
				apiError.Method = c.Request.Method
				
				m.logError(c, apiError)
				c.JSON(apiError.StatusCode, apiError)
				c.Abort()
				return
			}
		}

		// Validate URL path length
		if len(c.Request.URL.Path) > 2048 {
			apiError := NewBadRequestError("URL path too long", "Maximum URL path length is 2048 characters")
			apiError.RequestID = m.getRequestID(c)
			apiError.Path = c.Request.URL.Path
			apiError.Method = c.Request.Method
			
			m.logError(c, apiError)
			c.JSON(apiError.StatusCode, apiError)
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoggingMiddleware logs all requests
func (m *ErrorMiddleware) LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		requestID := ""
		if param.Keys != nil {
			if id, exists := param.Keys["request_id"]; exists {
				requestID = fmt.Sprintf("%v", id)
			}
		}

		return fmt.Sprintf("[%s] %s - %s \"%s %s %s\" %d %s \"%s\" \"%s\" %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			requestID,
			param.ClientIP,
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.Request.Referer(),
			param.ErrorMessage,
		)
	})
}

// Helper methods

func (m *ErrorMiddleware) getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

func (m *ErrorMiddleware) logError(c *gin.Context, apiError *APIError) {
	fields := map[string]interface{}{
		"request_id":  apiError.RequestID,
		"error_type":  apiError.Type,
		"error_code":  apiError.Code,
		"message":     apiError.Message,
		"details":     apiError.Details,
		"status_code": apiError.StatusCode,
		"path":        apiError.Path,
		"method":      apiError.Method,
		"ip":          c.ClientIP(),
		"user_agent":  c.Request.UserAgent(),
	}

	if apiError.StatusCode >= 500 {
		m.logger.Error(c.Request.Context(), "Server error occurred", fields)
	} else if apiError.StatusCode >= 400 {
		m.logger.Warn(c.Request.Context(), "Client error occurred", fields)
	}
}

func (m *ErrorMiddleware) convertToAPIError(err error) *APIError {
	errMsg := err.Error()
	
	// Check for common error patterns
	switch {
	case strings.Contains(errMsg, "validation"):
		return NewValidationError("Validation failed", errMsg)
	case strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "authentication"):
		return NewAuthenticationError("Authentication failed")
	case strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "permission"):
		return NewAuthorizationError("Access denied")
	case strings.Contains(errMsg, "not found"):
		return NewNotFoundError("Resource")
	case strings.Contains(errMsg, "conflict") || strings.Contains(errMsg, "duplicate"):
		return NewConflictError("Resource conflict", errMsg)
	default:
		return NewInternalError(errMsg)
	}
}

func (m *ErrorMiddleware) sanitizeString(input string) string {
	// Remove potentially dangerous characters
	dangerous := []string{"<", ">", "\"", "'", "&", "javascript:", "data:", "vbscript:"}
	result := input
	
	for _, char := range dangerous {
		result = strings.ReplaceAll(result, char, "")
	}
	
	return strings.TrimSpace(result)
}

func (m *ErrorMiddleware) isValidContentType(contentType string) bool {
	validTypes := []string{
		"application/json",
		"multipart/form-data",
		"application/pdf",
		"application/x-www-form-urlencoded",
		"text/plain",
	}
	
	// Extract main content type (ignore charset and other parameters)
	mainType := strings.Split(contentType, ";")[0]
	mainType = strings.TrimSpace(mainType)
	
	for _, validType := range validTypes {
		if strings.HasPrefix(mainType, validType) {
			return true
		}
	}
	
	return false
}

// AbortWithError is a helper function to abort with a custom API error
func AbortWithError(c *gin.Context, apiError *APIError) {
	c.Error(apiError)
	c.Abort()
}

// AbortWithValidationError is a helper for validation errors
func AbortWithValidationError(c *gin.Context, message, details string) {
	AbortWithError(c, NewValidationError(message, details))
}

// AbortWithAuthenticationError is a helper for authentication errors
func AbortWithAuthenticationError(c *gin.Context, message string) {
	AbortWithError(c, NewAuthenticationError(message))
}

// AbortWithAuthorizationError is a helper for authorization errors
func AbortWithAuthorizationError(c *gin.Context, message string) {
	AbortWithError(c, NewAuthorizationError(message))
}

// AbortWithNotFoundError is a helper for not found errors
func AbortWithNotFoundError(c *gin.Context, resource string) {
	AbortWithError(c, NewNotFoundError(resource))
}

// AbortWithInternalError is a helper for internal errors
func AbortWithInternalError(c *gin.Context, message string) {
	AbortWithError(c, NewInternalError(message))
}