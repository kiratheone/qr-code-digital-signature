package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIError(t *testing.T) {
	t.Run("NewAPIError creates error correctly", func(t *testing.T) {
		err := NewAPIError(ErrorTypeValidation, "TEST_CODE", "Test message", "Test details", http.StatusBadRequest)
		
		assert.Equal(t, ErrorTypeValidation, err.Type)
		assert.Equal(t, "TEST_CODE", err.Code)
		assert.Equal(t, "Test message", err.Message)
		assert.Equal(t, "Test details", err.Details)
		assert.Equal(t, http.StatusBadRequest, err.StatusCode)
		assert.NotZero(t, err.Timestamp)
	})

	t.Run("Error method returns formatted string", func(t *testing.T) {
		err := NewAPIError(ErrorTypeValidation, "TEST_CODE", "Test message", "Test details", http.StatusBadRequest)
		expected := "[TEST_CODE] VALIDATION_ERROR: Test message"
		assert.Equal(t, expected, err.Error())
	})
}

func TestErrorConstructors(t *testing.T) {
	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("Invalid input", "Field is required")
		assert.Equal(t, ErrorTypeValidation, err.Type)
		assert.Equal(t, "VALIDATION_FAILED", err.Code)
		assert.Equal(t, "Invalid input", err.Message)
		assert.Equal(t, "Field is required", err.Details)
		assert.Equal(t, http.StatusBadRequest, err.StatusCode)
	})

	t.Run("NewAuthenticationError", func(t *testing.T) {
		err := NewAuthenticationError("Invalid credentials")
		assert.Equal(t, ErrorTypeAuthentication, err.Type)
		assert.Equal(t, "AUTHENTICATION_FAILED", err.Code)
		assert.Equal(t, "Invalid credentials", err.Message)
		assert.Equal(t, http.StatusUnauthorized, err.StatusCode)
	})

	t.Run("NewNotFoundError", func(t *testing.T) {
		err := NewNotFoundError("User")
		assert.Equal(t, ErrorTypeNotFound, err.Type)
		assert.Equal(t, "RESOURCE_NOT_FOUND", err.Code)
		assert.Equal(t, "User not found", err.Message)
		assert.Equal(t, http.StatusNotFound, err.StatusCode)
	})

	t.Run("NewRateLimitError", func(t *testing.T) {
		err := NewRateLimitError(60)
		assert.Equal(t, ErrorTypeRateLimit, err.Type)
		assert.Equal(t, "RATE_LIMIT_EXCEEDED", err.Code)
		assert.Equal(t, "Too many requests", err.Message)
		assert.Contains(t, err.Details, "Retry after 60 seconds")
		assert.Equal(t, http.StatusTooManyRequests, err.StatusCode)
	})
}

func TestDefaultLogger(t *testing.T) {
	logger := &DefaultLogger{}
	ctx := context.Background()
	fields := map[string]interface{}{"test": "value"}

	// These should not panic
	logger.Error(ctx, "test error", fields)
	logger.Warn(ctx, "test warning", fields)
	logger.Info(ctx, "test info", fields)
}

func TestErrorMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("RequestIDMiddleware adds request ID", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		
		router.GET("/test", func(c *gin.Context) {
			requestID, exists := c.Get("request_id")
			assert.True(t, exists)
			assert.NotEmpty(t, requestID)
			c.JSON(200, gin.H{"request_id": requestID})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	})

	t.Run("RequestIDMiddleware uses existing request ID", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		
		router.GET("/test", func(c *gin.Context) {
			requestID, _ := c.Get("request_id")
			c.JSON(200, gin.H{"request_id": requestID})
		})

		existingID := "existing-request-id"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", existingID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, existingID, w.Header().Get("X-Request-ID"))
	})

	t.Run("ErrorHandler handles panics", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		router.Use(middleware.ErrorHandler())
		
		router.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		req := httptest.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		var response APIError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, ErrorTypeInternal, response.Type)
		assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Code)
		assert.NotEmpty(t, response.RequestID)
	})

	t.Run("HandleError processes API errors", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		router.Use(middleware.HandleError())
		
		router.GET("/error", func(c *gin.Context) {
			apiError := NewValidationError("Test validation error", "Field is required")
			c.Error(apiError)
		})

		req := httptest.NewRequest("GET", "/error", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response APIError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, ErrorTypeValidation, response.Type)
		assert.Equal(t, "VALIDATION_FAILED", response.Code)
		assert.Equal(t, "Test validation error", response.Message)
		assert.Equal(t, "Field is required", response.Details)
		assert.NotEmpty(t, response.RequestID)
	})

	t.Run("SecurityHeadersMiddleware adds security headers", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.SecurityHeadersMiddleware())
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
		assert.Equal(t, "default-src 'self'", w.Header().Get("Content-Security-Policy"))
	})

	t.Run("RequestSanitizationMiddleware sanitizes input", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestSanitizationMiddleware())
		
		router.GET("/test", func(c *gin.Context) {
			query := c.Query("test")
			c.JSON(200, gin.H{"query": query})
		})

		req := httptest.NewRequest("GET", "/test?test=<script>alert('xss')</script>", nil)
		req.Header.Set("X-Test", "<script>alert('xss')</script>")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		// Should be sanitized (script tags removed)
		query := response["query"].(string)
		assert.NotContains(t, query, "<script>")
		assert.NotContains(t, query, "</script>")
	})
}

func TestConvertToAPIError(t *testing.T) {
	middleware := NewErrorMiddleware(nil)

	tests := []struct {
		name           string
		inputError     string
		expectedType   ErrorType
		expectedCode   string
	}{
		{
			name:         "validation error",
			inputError:   "validation failed for field",
			expectedType: ErrorTypeValidation,
			expectedCode: "VALIDATION_FAILED",
		},
		{
			name:         "authentication error",
			inputError:   "unauthorized access",
			expectedType: ErrorTypeAuthentication,
			expectedCode: "AUTHENTICATION_FAILED",
		},
		{
			name:         "authorization error",
			inputError:   "forbidden - insufficient permissions",
			expectedType: ErrorTypeAuthorization,
			expectedCode: "AUTHORIZATION_FAILED",
		},
		{
			name:         "not found error",
			inputError:   "resource not found",
			expectedType: ErrorTypeNotFound,
			expectedCode: "RESOURCE_NOT_FOUND",
		},
		{
			name:         "conflict error",
			inputError:   "duplicate entry conflict",
			expectedType: ErrorTypeConflict,
			expectedCode: "RESOURCE_CONFLICT",
		},
		{
			name:         "generic error",
			inputError:   "some random error",
			expectedType: ErrorTypeInternal,
			expectedCode: "INTERNAL_SERVER_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := middleware.convertToAPIError(errors.New(tt.inputError))
			assert.Equal(t, tt.expectedType, err.Type)
			assert.Equal(t, tt.expectedCode, err.Code)
		})
	}
}

func TestSanitizeString(t *testing.T) {
	middleware := NewErrorMiddleware(nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes script tags",
			input:    "<script>alert('xss')</script>",
			expected: "scriptalert(xss)/script",
		},
		{
			name:     "removes quotes",
			input:    `"test"`,
			expected: "test",
		},
		{
			name:     "removes javascript protocol",
			input:    "javascript:alert('xss')",
			expected: "alert(xss)",
		},
		{
			name:     "removes data protocol",
			input:    "data:text/html,<script>alert('xss')</script>",
			expected: "text/html,scriptalert(xss)/script",
		},
		{
			name:     "normal string unchanged",
			input:    "normal text",
			expected: "normal text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.sanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAbortHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("AbortWithValidationError", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		router.Use(middleware.HandleError())
		router.GET("/test", func(c *gin.Context) {
			AbortWithValidationError(c, "Invalid input", "Field is required")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.True(t, strings.Contains(w.Body.String(), "VALIDATION_FAILED"))
	})

	t.Run("AbortWithAuthenticationError", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		router.Use(middleware.HandleError())
		router.GET("/test", func(c *gin.Context) {
			AbortWithAuthenticationError(c, "Invalid credentials")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, strings.Contains(w.Body.String(), "AUTHENTICATION_FAILED"))
	})

	t.Run("AbortWithNotFoundError", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		router.Use(middleware.HandleError())
		router.GET("/test", func(c *gin.Context) {
			AbortWithNotFoundError(c, "User")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.True(t, strings.Contains(w.Body.String(), "RESOURCE_NOT_FOUND"))
	})
}

// Mock error for testing
type mockError struct {
	message string
}

func (e mockError) Error() string {
	return e.message
}

func TestRequestValidationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("rejects requests that are too large", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		router.Use(middleware.RequestValidationMiddleware())
		
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// Create a request with large content length
		req := httptest.NewRequest("POST", "/test", strings.NewReader("test"))
		req.ContentLength = 51 * 1024 * 1024 // 51MB
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response APIError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, ErrorTypeBadRequest, response.Type)
		assert.Contains(t, response.Message, "Request too large")
	})

	t.Run("rejects invalid content types", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		router.Use(middleware.RequestValidationMiddleware())
		
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("POST", "/test", strings.NewReader("test"))
		req.Header.Set("Content-Type", "application/xml")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response APIError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, ErrorTypeBadRequest, response.Type)
		assert.Contains(t, response.Message, "Invalid content type")
	})

	t.Run("accepts valid content types", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		router.Use(middleware.RequestValidationMiddleware())
		
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		validContentTypes := []string{
			"application/json",
			"application/json; charset=utf-8",
			"multipart/form-data",
			"application/pdf",
			"application/x-www-form-urlencoded",
			"text/plain",
		}

		for _, contentType := range validContentTypes {
			req := httptest.NewRequest("POST", "/test", strings.NewReader("test"))
			req.Header.Set("Content-Type", contentType)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Content type %s should be valid", contentType)
		}
	})

	t.Run("rejects URLs that are too long", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		router.Use(middleware.RequestValidationMiddleware())
		
		router.GET("/*path", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// Create a very long URL path
		longPath := "/" + strings.Repeat("a", 2050)
		req := httptest.NewRequest("GET", longPath, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response APIError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, ErrorTypeBadRequest, response.Type)
		assert.Contains(t, response.Message, "URL path too long")
	})

	t.Run("allows GET requests without content type validation", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		router.Use(middleware.RequestValidationMiddleware())
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestIsValidContentType(t *testing.T) {
	middleware := NewErrorMiddleware(nil)

	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{
			name:        "valid json",
			contentType: "application/json",
			expected:    true,
		},
		{
			name:        "valid json with charset",
			contentType: "application/json; charset=utf-8",
			expected:    true,
		},
		{
			name:        "valid multipart",
			contentType: "multipart/form-data; boundary=something",
			expected:    true,
		},
		{
			name:        "valid pdf",
			contentType: "application/pdf",
			expected:    true,
		},
		{
			name:        "invalid xml",
			contentType: "application/xml",
			expected:    false,
		},
		{
			name:        "invalid image",
			contentType: "image/jpeg",
			expected:    false,
		},
		{
			name:        "empty content type",
			contentType: "",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.isValidContentType(tt.contentType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorHandlingScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handles concurrent errors correctly", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		router.Use(middleware.HandleError())
		
		router.GET("/error", func(c *gin.Context) {
			c.Error(NewInternalError("Concurrent error"))
		})

		// Make concurrent requests
		var wg sync.WaitGroup
		results := make([]int, 10)
		
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				req := httptest.NewRequest("GET", "/error", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				results[index] = w.Code
			}(i)
		}
		
		wg.Wait()
		
		// All requests should return 500
		for i, code := range results {
			assert.Equal(t, http.StatusInternalServerError, code, "Request %d should return 500", i)
		}
	})

	t.Run("handles multiple errors in single request", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		router.Use(middleware.HandleError())
		
		router.GET("/multiple-errors", func(c *gin.Context) {
			c.Error(NewValidationError("First error", "Details 1"))
			c.Error(NewAuthenticationError("Second error"))
			c.Error(NewInternalError("Third error"))
		})

		req := httptest.NewRequest("GET", "/multiple-errors", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return the last error (internal error)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		var response APIError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, ErrorTypeInternal, response.Type)
		assert.Equal(t, "Third error", response.Details)
	})

	t.Run("handles errors with special characters", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		router.Use(middleware.HandleError())
		
		router.GET("/special-chars", func(c *gin.Context) {
			c.Error(NewValidationError("Error with <script>alert('xss')</script>", "Details with \"quotes\" and 'apostrophes'"))
		})

		req := httptest.NewRequest("GET", "/special-chars", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response APIError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		// Response should be properly JSON encoded
		assert.Contains(t, response.Message, "script")
		assert.Contains(t, response.Details, "quotes")
	})

	t.Run("handles very long error messages", func(t *testing.T) {
		middleware := NewErrorMiddleware(nil)
		router := gin.New()
		router.Use(middleware.RequestIDMiddleware())
		router.Use(middleware.HandleError())
		
		longMessage := strings.Repeat("This is a very long error message. ", 100)
		router.GET("/long-error", func(c *gin.Context) {
			c.Error(NewValidationError(longMessage, "Short details"))
		})

		req := httptest.NewRequest("GET", "/long-error", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response APIError
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, longMessage, response.Message)
		assert.Equal(t, "Short details", response.Details)
	})
}

func TestLoggingIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	var logBuffer bytes.Buffer
	logger := NewStructuredLogger(LogLevelInfo, &logBuffer)
	middleware := NewErrorMiddleware(logger)

	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.HandleError())
	
	router.GET("/error", func(c *gin.Context) {
		c.Error(NewInternalError("Test internal error"))
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	// Check that error was logged
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Server error occurred")
	assert.Contains(t, logOutput, "INTERNAL_SERVER_ERROR")
}