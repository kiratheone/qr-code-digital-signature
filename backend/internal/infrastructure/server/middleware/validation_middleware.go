package middleware

import (
	"digital-signature-system/internal/infrastructure/validation"
	"fmt"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ValidationMiddleware provides comprehensive input validation
type ValidationMiddleware struct {
	validator *validation.Validator
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{
		validator: validation.NewValidator(),
	}
}

// ValidateJSON validates JSON request bodies
func (vm *ValidationMiddleware) ValidateJSON(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(model); err != nil {
			AbortWithValidationError(c, "Invalid JSON data", err.Error())
			return
		}
		
		if err := vm.validator.ValidateStruct(model); err != nil {
			AbortWithValidationError(c, "Validation failed", err.Error())
			return
		}
		
		c.Set("validated_data", model)
		c.Next()
	}
}

// ValidateForm validates form data
func (vm *ValidationMiddleware) ValidateForm(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBind(model); err != nil {
			AbortWithValidationError(c, "Invalid form data", err.Error())
			return
		}
		
		if err := vm.validator.ValidateStruct(model); err != nil {
			AbortWithValidationError(c, "Validation failed", err.Error())
			return
		}
		
		c.Set("validated_data", model)
		c.Next()
	}
}

// ValidateMultipartForm validates multipart form data with file uploads
func (vm *ValidationMiddleware) ValidateMultipartForm(maxMemory int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.Request.ParseMultipartForm(maxMemory); err != nil {
			AbortWithValidationError(c, "Failed to parse multipart form", err.Error())
			return
		}
		
		// Validate and sanitize form values
		for key, values := range c.Request.MultipartForm.Value {
			for i, value := range values {
				sanitized := vm.validator.SanitizeString(value)
				c.Request.MultipartForm.Value[key][i] = sanitized
			}
		}
		
		c.Next()
	}
}

// ValidatePDFUpload validates PDF file uploads
func (vm *ValidationMiddleware) ValidatePDFUpload(fieldName string, maxSize int64, required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile(fieldName)
		if err != nil {
			if required {
				AbortWithValidationError(c, "PDF file is required", fmt.Sprintf("Missing required file field: %s", fieldName))
				return
			}
			c.Next()
			return
		}
		
		if err := vm.validator.ValidatePDFFile(file, maxSize); err != nil {
			AbortWithValidationError(c, "Invalid PDF file", err.Error())
			return
		}
		
		c.Set("validated_file", file)
		c.Next()
	}
}

// ValidateFileUpload validates general file uploads
func (vm *ValidationMiddleware) ValidateFileUpload(fieldName string, allowedTypes []string, maxSize int64, required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile(fieldName)
		if err != nil {
			if required {
				AbortWithValidationError(c, "File is required", fmt.Sprintf("Missing required file field: %s", fieldName))
				return
			}
			c.Next()
			return
		}
		
		if err := vm.validator.ValidateFile(file, allowedTypes, maxSize); err != nil {
			AbortWithValidationError(c, "Invalid file", err.Error())
			return
		}
		
		c.Set("validated_file", file)
		c.Next()
	}
}

// ValidateQueryParams validates query parameters
func (vm *ValidationMiddleware) ValidateQueryParams(params map[string]QueryParamConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		validatedParams := make(map[string]interface{})
		
		for paramName, config := range params {
			value := c.Query(paramName)
			
			// Check if required parameter is missing
			if config.Required && value == "" {
				AbortWithValidationError(c, "Missing required parameter", fmt.Sprintf("Parameter '%s' is required", paramName))
				return
			}
			
			// Skip validation if parameter is not provided and not required
			if value == "" {
				if config.DefaultValue != nil {
					validatedParams[paramName] = config.DefaultValue
				}
				continue
			}
			
			// Sanitize the value
			sanitized := vm.validator.SanitizeString(value)
			
			// Validate based on type
			validated, err := vm.validateQueryParam(sanitized, config)
			if err != nil {
				AbortWithValidationError(c, "Invalid query parameter", fmt.Sprintf("Parameter '%s': %s", paramName, err.Error()))
				return
			}
			
			validatedParams[paramName] = validated
		}
		
		c.Set("validated_query_params", validatedParams)
		c.Next()
	}
}

// ValidatePathParams validates path parameters
func (vm *ValidationMiddleware) ValidatePathParams(params map[string]PathParamConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		validatedParams := make(map[string]interface{})
		
		for paramName, config := range params {
			value := c.Param(paramName)
			
			if value == "" {
				AbortWithValidationError(c, "Missing path parameter", fmt.Sprintf("Path parameter '%s' is required", paramName))
				return
			}
			
			// Sanitize the value
			sanitized := vm.validator.SanitizeString(value)
			
			// Validate based on type
			validated, err := vm.validatePathParam(sanitized, config)
			if err != nil {
				AbortWithValidationError(c, "Invalid path parameter", fmt.Sprintf("Parameter '%s': %s", paramName, err.Error()))
				return
			}
			
			validatedParams[paramName] = validated
		}
		
		c.Set("validated_path_params", validatedParams)
		c.Next()
	}
}

// ValidateHeaders validates request headers
func (vm *ValidationMiddleware) ValidateHeaders(headers map[string]HeaderConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		validatedHeaders := make(map[string]string)
		
		for headerName, config := range headers {
			value := c.GetHeader(headerName)
			
			if config.Required && value == "" {
				AbortWithValidationError(c, "Missing required header", fmt.Sprintf("Header '%s' is required", headerName))
				return
			}
			
			if value == "" {
				continue
			}
			
			// Sanitize the value
			sanitized := vm.validator.SanitizeString(value)
			
			// Validate length
			if config.MaxLength > 0 && len(sanitized) > config.MaxLength {
				AbortWithValidationError(c, "Header value too long", fmt.Sprintf("Header '%s' exceeds maximum length of %d", headerName, config.MaxLength))
				return
			}
			
			// Validate pattern if provided
			if config.Pattern != "" {
				if matched, _ := regexp.MatchString(config.Pattern, sanitized); !matched {
					AbortWithValidationError(c, "Invalid header format", fmt.Sprintf("Header '%s' does not match required pattern", headerName))
					return
				}
			}
			
			validatedHeaders[headerName] = sanitized
		}
		
		c.Set("validated_headers", validatedHeaders)
		c.Next()
	}
}

// SanitizeAllInputs sanitizes all request inputs
func (vm *ValidationMiddleware) SanitizeAllInputs() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		query := c.Request.URL.Query()
		for key, values := range query {
			for i, value := range values {
				query[key][i] = vm.validator.SanitizeString(value)
			}
		}
		c.Request.URL.RawQuery = query.Encode()
		
		// Sanitize headers (except standard ones)
		skipHeaders := map[string]bool{
			"Authorization": true,
			"Content-Type":  true,
			"Content-Length": true,
			"User-Agent":    true,
			"Accept":        true,
			"Host":          true,
		}
		
		for key, values := range c.Request.Header {
			if skipHeaders[key] {
				continue
			}
			for i, value := range values {
				c.Request.Header[key][i] = vm.validator.SanitizeString(value)
			}
		}
		
		c.Next()
	}
}

// Configuration types

type QueryParamConfig struct {
	Type         string      // "string", "int", "bool", "uuid"
	Required     bool
	DefaultValue interface{}
	MinValue     *int
	MaxValue     *int
	MinLength    int
	MaxLength    int
	Pattern      string
}

type PathParamConfig struct {
	Type      string // "string", "int", "uuid"
	MinLength int
	MaxLength int
	Pattern   string
}

type HeaderConfig struct {
	Required  bool
	MaxLength int
	Pattern   string
}

// Helper methods

func (vm *ValidationMiddleware) validateQueryParam(value string, config QueryParamConfig) (interface{}, error) {
	switch config.Type {
	case "int":
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("must be a valid integer")
		}
		if config.MinValue != nil && intVal < *config.MinValue {
			return nil, fmt.Errorf("must be at least %d", *config.MinValue)
		}
		if config.MaxValue != nil && intVal > *config.MaxValue {
			return nil, fmt.Errorf("must be at most %d", *config.MaxValue)
		}
		return intVal, nil
		
	case "bool":
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("must be a valid boolean")
		}
		return boolVal, nil
		
	case "uuid":
		if !isValidUUID(value) {
			return nil, fmt.Errorf("must be a valid UUID")
		}
		return value, nil
		
	default: // string
		if config.MinLength > 0 && len(value) < config.MinLength {
			return nil, fmt.Errorf("must be at least %d characters", config.MinLength)
		}
		if config.MaxLength > 0 && len(value) > config.MaxLength {
			return nil, fmt.Errorf("must be at most %d characters", config.MaxLength)
		}
		if config.Pattern != "" {
			if matched, _ := regexp.MatchString(config.Pattern, value); !matched {
				return nil, fmt.Errorf("does not match required pattern")
			}
		}
		return value, nil
	}
}

func (vm *ValidationMiddleware) validatePathParam(value string, config PathParamConfig) (interface{}, error) {
	switch config.Type {
	case "int":
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("must be a valid integer")
		}
		return intVal, nil
		
	case "uuid":
		if !isValidUUID(value) {
			return nil, fmt.Errorf("must be a valid UUID")
		}
		return value, nil
		
	default: // string
		if config.MinLength > 0 && len(value) < config.MinLength {
			return nil, fmt.Errorf("must be at least %d characters", config.MinLength)
		}
		if config.MaxLength > 0 && len(value) > config.MaxLength {
			return nil, fmt.Errorf("must be at most %d characters", config.MaxLength)
		}
		if config.Pattern != "" {
			if matched, _ := regexp.MatchString(config.Pattern, value); !matched {
				return nil, fmt.Errorf("does not match required pattern")
			}
		}
		return value, nil
	}
}

// Helper functions

func isValidUUID(uuid string) bool {
	uuidRegex := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(uuidRegex, uuid)
	return matched
}

// Convenience functions for common validations

func ValidateDocumentID() gin.HandlerFunc {
	vm := NewValidationMiddleware()
	return vm.ValidatePathParams(map[string]PathParamConfig{
		"id": {
			Type:      "uuid",
			MinLength: 36,
			MaxLength: 36,
		},
	})
}

func ValidateDocumentQuery() gin.HandlerFunc {
	vm := NewValidationMiddleware()
	defaultPage := 1
	defaultPageSize := 10
	maxPageSize := 100
	
	return vm.ValidateQueryParams(map[string]QueryParamConfig{
		"search": {
			Type:      "string",
			Required:  false,
			MaxLength: 100,
		},
		"page": {
			Type:         "int",
			Required:     false,
			DefaultValue: &defaultPage,
			MinValue:     &defaultPage,
		},
		"page_size": {
			Type:         "int",
			Required:     false,
			DefaultValue: &defaultPageSize,
			MinValue:     &defaultPage,
			MaxValue:     &maxPageSize,
		},
		"sort_by": {
			Type:      "string",
			Required:  false,
			MaxLength: 50,
			Pattern:   "^[a-zA-Z_]+$",
		},
		"sort_desc": {
			Type:     "bool",
			Required: false,
		},
	})
}

func ValidatePDFUploadForm() gin.HandlerFunc {
	vm := NewValidationMiddleware()
	return vm.ValidateMultipartForm(50 * 1024 * 1024) // 50MB
}

func ValidatePDFFile() gin.HandlerFunc {
	vm := NewValidationMiddleware()
	return vm.ValidatePDFUpload("pdf", 50*1024*1024, true) // 50MB, required
}

func ValidateVerificationFile() gin.HandlerFunc {
	vm := NewValidationMiddleware()
	return vm.ValidatePDFUpload("document", 50*1024*1024, true) // 50MB, required
}