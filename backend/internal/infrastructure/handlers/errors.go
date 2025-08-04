package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// StandardError represents a standardized error response
type StandardError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error codes for different types of errors
const (
	ErrCodeInvalidRequest     = "INVALID_REQUEST"
	ErrCodeUnauthorized      = "UNAUTHORIZED"
	ErrCodeForbidden         = "FORBIDDEN"
	ErrCodeNotFound          = "NOT_FOUND"
	ErrCodeConflict          = "CONFLICT"
	ErrCodeValidationFailed  = "VALIDATION_FAILED"
	ErrCodeInternalError     = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrCodeInvalidFile       = "INVALID_FILE"
	ErrCodeFileTooLarge      = "FILE_TOO_LARGE"
	ErrCodeInvalidPDF        = "INVALID_PDF"
	ErrCodeSignatureFailed   = "SIGNATURE_FAILED"
	ErrCodeVerificationFailed = "VERIFICATION_FAILED"
)

// NewStandardError creates a new standardized error
func NewStandardError(code, message string, details ...string) *StandardError {
	err := &StandardError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// Error implements the error interface
func (e *StandardError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, statusCode int, err *StandardError) {
	c.JSON(statusCode, err)
}

// RespondWithValidationError sends a validation error response
func RespondWithValidationError(c *gin.Context, message string, details ...string) {
	err := NewStandardError(ErrCodeValidationFailed, message, details...)
	RespondWithError(c, http.StatusBadRequest, err)
}

// RespondWithInternalError sends an internal server error response
func RespondWithInternalError(c *gin.Context, message string, details ...string) {
	err := NewStandardError(ErrCodeInternalError, message, details...)
	RespondWithError(c, http.StatusInternalServerError, err)
}

// RespondWithUnauthorizedError sends an unauthorized error response
func RespondWithUnauthorizedError(c *gin.Context, message string, details ...string) {
	err := NewStandardError(ErrCodeUnauthorized, message, details...)
	RespondWithError(c, http.StatusUnauthorized, err)
}

// RespondWithForbiddenError sends a forbidden error response
func RespondWithForbiddenError(c *gin.Context, message string, details ...string) {
	err := NewStandardError(ErrCodeForbidden, message, details...)
	RespondWithError(c, http.StatusForbidden, err)
}

// RespondWithNotFoundError sends a not found error response
func RespondWithNotFoundError(c *gin.Context, message string, details ...string) {
	err := NewStandardError(ErrCodeNotFound, message, details...)
	RespondWithError(c, http.StatusNotFound, err)
}

// RespondWithConflictError sends a conflict error response
func RespondWithConflictError(c *gin.Context, message string, details ...string) {
	err := NewStandardError(ErrCodeConflict, message, details...)
	RespondWithError(c, http.StatusConflict, err)
}

// MapServiceErrorToHTTP maps service layer errors to HTTP responses
func MapServiceErrorToHTTP(c *gin.Context, err error) {
	switch err.Error() {
	case "invalid credentials":
		RespondWithUnauthorizedError(c, "Invalid username or password")
	case "user already exists":
		RespondWithConflictError(c, "Username or email already exists")
	case "user not found":
		RespondWithNotFoundError(c, "User not found")
	case "user inactive":
		RespondWithForbiddenError(c, "User account is inactive")
	case "invalid token":
		RespondWithUnauthorizedError(c, "Invalid or malformed token")
	case "session expired":
		RespondWithUnauthorizedError(c, "Token has expired")
	case "document not found":
		RespondWithNotFoundError(c, "Document not found")
	case "access denied: document belongs to different user":
		RespondWithForbiddenError(c, "Access denied")
	case "invalid PDF format":
		RespondWithValidationError(c, "Invalid PDF format")
	case "file too large":
		RespondWithValidationError(c, "File size exceeds maximum limit")
	case "signature generation failed":
		RespondWithInternalError(c, "Failed to generate digital signature")
	case "verification failed":
		RespondWithInternalError(c, "Document verification failed")
	default:
		RespondWithInternalError(c, "An internal error occurred", err.Error())
	}
}