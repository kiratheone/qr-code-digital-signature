package validation

import (
	"fmt"
	"html"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("validation failed with %d errors", len(e))
}

// Validator provides input validation and sanitization
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateAndSanitizeString validates and sanitizes a string input
func (v *Validator) ValidateAndSanitizeString(field, value string, minLen, maxLen int, required bool) (string, *ValidationError) {
	// Check if required
	if required && strings.TrimSpace(value) == "" {
		return "", &ValidationError{
			Field:   field,
			Message: "field is required",
			Code:    "REQUIRED",
		}
	}

	// If not required and empty, return empty string
	if !required && strings.TrimSpace(value) == "" {
		return "", nil
	}

	// Sanitize the input
	sanitized := v.SanitizeString(value)

	// Check length after sanitization
	if len(sanitized) < minLen {
		return "", &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("minimum length is %d characters", minLen),
			Code:    "MIN_LENGTH",
		}
	}

	if maxLen > 0 && len(sanitized) > maxLen {
		return "", &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("maximum length is %d characters", maxLen),
			Code:    "MAX_LENGTH",
		}
	}

	// Check for suspicious content
	if v.containsSuspiciousContent(sanitized) {
		return "", &ValidationError{
			Field:   field,
			Message: "contains invalid characters or patterns",
			Code:    "SUSPICIOUS_CONTENT",
		}
	}

	return sanitized, nil
}

// ValidateEmail validates and sanitizes an email address
func (v *Validator) ValidateEmail(field, email string, required bool) (string, *ValidationError) {
	// Check if required
	if required && strings.TrimSpace(email) == "" {
		return "", &ValidationError{
			Field:   field,
			Message: "email is required",
			Code:    "REQUIRED",
		}
	}

	// If not required and empty, return empty string
	if !required && strings.TrimSpace(email) == "" {
		return "", nil
	}

	// Sanitize and normalize
	sanitized := strings.TrimSpace(strings.ToLower(email))

	// Validate email format
	if _, err := mail.ParseAddress(sanitized); err != nil {
		return "", &ValidationError{
			Field:   field,
			Message: "invalid email format",
			Code:    "INVALID_EMAIL",
		}
	}

	// Check length
	if len(sanitized) > 254 { // RFC 5321 limit
		return "", &ValidationError{
			Field:   field,
			Message: "email address too long",
			Code:    "MAX_LENGTH",
		}
	}

	return sanitized, nil
}

// ValidateUUID validates a UUID string
func (v *Validator) ValidateUUID(field, value string, required bool) (string, *ValidationError) {
	// Check if required
	if required && strings.TrimSpace(value) == "" {
		return "", &ValidationError{
			Field:   field,
			Message: "UUID is required",
			Code:    "REQUIRED",
		}
	}

	// If not required and empty, return empty string
	if !required && strings.TrimSpace(value) == "" {
		return "", nil
	}

	// Validate UUID format
	if _, err := uuid.Parse(value); err != nil {
		return "", &ValidationError{
			Field:   field,
			Message: "invalid UUID format",
			Code:    "INVALID_UUID",
		}
	}

	return value, nil
}

// ValidatePassword validates password strength
func (v *Validator) ValidatePassword(field, password string) *ValidationError {
	if len(password) < 8 {
		return &ValidationError{
			Field:   field,
			Message: "password must be at least 8 characters long",
			Code:    "MIN_LENGTH",
		}
	}

	if len(password) > 128 {
		return &ValidationError{
			Field:   field,
			Message: "password must be less than 128 characters",
			Code:    "MAX_LENGTH",
		}
	}

	// Check for at least one uppercase, lowercase, digit, and special character
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return &ValidationError{
			Field:   field,
			Message: "password must contain at least one uppercase letter, lowercase letter, digit, and special character",
			Code:    "WEAK_PASSWORD",
		}
	}

	return nil
}

// ValidateFilename validates and sanitizes a filename
func (v *Validator) ValidateFilename(field, filename string, required bool) (string, *ValidationError) {
	// Check if required
	if required && strings.TrimSpace(filename) == "" {
		return "", &ValidationError{
			Field:   field,
			Message: "filename is required",
			Code:    "REQUIRED",
		}
	}

	// If not required and empty, return empty string
	if !required && strings.TrimSpace(filename) == "" {
		return "", nil
	}

	// Sanitize filename
	sanitized := v.SanitizeFilename(filename)

	// Check length
	if len(sanitized) > 255 {
		return "", &ValidationError{
			Field:   field,
			Message: "filename too long",
			Code:    "MAX_LENGTH",
		}
	}

	// Check for valid characters
	validFilename := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validFilename.MatchString(sanitized) {
		return "", &ValidationError{
			Field:   field,
			Message: "filename contains invalid characters",
			Code:    "INVALID_CHARACTERS",
		}
	}

	return sanitized, nil
}

// ValidateFileSize validates file size
func (v *Validator) ValidateFileSize(field string, size, maxSize int64) *ValidationError {
	if size <= 0 {
		return &ValidationError{
			Field:   field,
			Message: "file is empty",
			Code:    "EMPTY_FILE",
		}
	}

	if size > maxSize {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("file size exceeds maximum limit of %d bytes", maxSize),
			Code:    "FILE_TOO_LARGE",
		}
	}

	return nil
}

// ValidateContentType validates file content type
func (v *Validator) ValidateContentType(field, contentType string, allowedTypes []string) *ValidationError {
	if contentType == "" {
		return &ValidationError{
			Field:   field,
			Message: "content type is required",
			Code:    "REQUIRED",
		}
	}

	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return nil
		}
	}

	return &ValidationError{
		Field:   field,
		Message: fmt.Sprintf("content type '%s' is not allowed", contentType),
		Code:    "INVALID_CONTENT_TYPE",
	}
}

// SanitizeString sanitizes a string by removing/escaping dangerous content
func (v *Validator) SanitizeString(input string) string {
	// Trim whitespace
	sanitized := strings.TrimSpace(input)

	// HTML escape to prevent XSS
	sanitized = html.EscapeString(sanitized)

	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	// Normalize unicode
	if !utf8.ValidString(sanitized) {
		// If not valid UTF-8, remove invalid characters
		sanitized = strings.ToValidUTF8(sanitized, "")
	}

	return sanitized
}

// SanitizeFilename sanitizes a filename
func (v *Validator) SanitizeFilename(filename string) string {
	// Remove path separators and dangerous characters
	sanitized := strings.ReplaceAll(filename, "/", "")
	sanitized = strings.ReplaceAll(sanitized, "\\", "")
	sanitized = strings.ReplaceAll(sanitized, "..", "")
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	// Remove leading/trailing dots and spaces
	sanitized = strings.Trim(sanitized, ". ")

	// If empty after sanitization, provide a default
	if sanitized == "" {
		sanitized = "unnamed_file"
	}

	return sanitized
}

// containsSuspiciousContent checks for common attack patterns
func (v *Validator) containsSuspiciousContent(input string) bool {
	// SQL injection patterns
	sqlPatterns := []string{
		"union select", "drop table", "insert into", "delete from",
		"update set", "alter table", "create table", "exec(",
		"execute(", "sp_", "xp_", "@@", "char(", "cast(",
	}

	// XSS patterns
	xssPatterns := []string{
		"<script", "</script>", "javascript:", "vbscript:",
		"onload=", "onerror=", "onclick=", "onmouseover=",
		"eval(", "alert(", "confirm(", "prompt(",
		"document.cookie", "document.write", "window.location",
	}

	// Path traversal patterns
	pathPatterns := []string{
		"../", "..\\", "/..", "\\..", "%2e%2e",
		"%252e%252e", "..%2f", "..%5c",
	}

	lowerInput := strings.ToLower(input)

	// Check all patterns
	allPatterns := append(append(sqlPatterns, xssPatterns...), pathPatterns...)
	for _, pattern := range allPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}

	// Check for excessive control characters
	controlCharCount := 0
	for _, char := range input {
		if unicode.IsControl(char) && char != '\n' && char != '\r' && char != '\t' {
			controlCharCount++
			if controlCharCount > 5 {
				return true
			}
		}
	}

	return false
}

// ValidateMultiple validates multiple fields and returns all errors
func (v *Validator) ValidateMultiple(validations ...func() *ValidationError) ValidationErrors {
	var errors ValidationErrors
	for _, validation := range validations {
		if err := validation(); err != nil {
			errors = append(errors, *err)
		}
	}
	return errors
}