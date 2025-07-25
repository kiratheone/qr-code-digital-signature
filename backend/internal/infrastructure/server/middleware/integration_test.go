package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("complete middleware stack integration", func(t *testing.T) {
		// Setup logging
		var logBuffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &logBuffer)
		
		// Create middleware instances
		errorMiddleware := NewErrorMiddleware(logger)
		rateLimitMiddleware := NewRateLimitMiddleware()
		defer rateLimitMiddleware.Stop()
		
		loggingMiddleware := NewLoggingMiddleware(LoggingConfig{
			Logger:          logger,
			SkipPaths:       []string{"/health"},
			LogRequestBody:  true,
			LogResponseBody: true,
			MaxBodySize:     1024,
		})
		
		auditLogger := NewAuditLogger(logger)

		// Setup router with complete middleware stack
		router := gin.New()
		
		// Apply middleware in correct order
		router.Use(errorMiddleware.ErrorHandler())
		router.Use(errorMiddleware.RequestIDMiddleware())
		router.Use(errorMiddleware.SecurityHeadersMiddleware())
		router.Use(errorMiddleware.RequestValidationMiddleware())
		router.Use(errorMiddleware.RequestSanitizationMiddleware())
		router.Use(loggingMiddleware.Middleware())
		router.Use(rateLimitMiddleware.IPLimit())
		router.Use(errorMiddleware.ValidationErrorHandler())
		router.Use(errorMiddleware.HandleError())

		// Test endpoints
		router.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		router.POST("/api/test", func(c *gin.Context) {
			// Simulate audit logging
			requestID, _ := c.Get("request_id")
			auditLogger.LogAction(c.Request.Context(), "TEST_ACTION", "test_resource", map[string]interface{}{
				"request_id": requestID,
				"ip":         c.ClientIP(),
			})
			
			c.JSON(200, gin.H{"message": "success", "request_id": requestID})
		})

		router.GET("/api/error", func(c *gin.Context) {
			AbortWithValidationError(c, "Test validation error", "Invalid input provided")
		})

		router.GET("/api/panic", func(c *gin.Context) {
			panic("test panic for recovery")
		})

		// Test 1: Health check (should be skipped from logging)
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))

		// Test 2: Normal API request with audit logging
		requestBody := `{"test": "data"}`
		req = httptest.NewRequest("POST", "/api/test", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "test-client")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["message"])
		assert.NotEmpty(t, response["request_id"])

		// Test 3: Error handling
		req = httptest.NewRequest("GET", "/api/error", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var errorResponse APIError
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		require.NoError(t, err)
		
		assert.Equal(t, ErrorTypeValidation, errorResponse.Type)
		assert.Equal(t, "VALIDATION_FAILED", errorResponse.Code)
		assert.Equal(t, "Test validation error", errorResponse.Message)
		assert.Equal(t, "Invalid input provided", errorResponse.Details)
		assert.NotEmpty(t, errorResponse.RequestID)
		assert.Equal(t, "/api/error", errorResponse.Path)
		assert.Equal(t, "GET", errorResponse.Method)

		// Test 4: Panic recovery
		req = httptest.NewRequest("GET", "/api/panic", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		require.NoError(t, err)
		
		assert.Equal(t, ErrorTypeInternal, errorResponse.Type)
		assert.Equal(t, "INTERNAL_SERVER_ERROR", errorResponse.Code)

		// Test 5: Request validation (too large request)
		req = httptest.NewRequest("POST", "/api/test", strings.NewReader("test"))
		req.ContentLength = 51 * 1024 * 1024 // 51MB
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		require.NoError(t, err)
		
		assert.Equal(t, ErrorTypeBadRequest, errorResponse.Type)
		assert.Contains(t, errorResponse.Message, "Request too large")

		// Test 6: Input sanitization
		req = httptest.NewRequest("GET", "/api/test?param=<script>alert('xss')</script>", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should not contain script tags in logs or response
		logOutput := logBuffer.String()
		assert.NotContains(t, logOutput, "<script>")
		assert.NotContains(t, logOutput, "</script>")

		// Verify comprehensive logging
		logLines := strings.Split(strings.TrimSpace(logOutput), "\n")
		
		// Should have multiple log entries
		assert.Greater(t, len(logLines), 5)
		
		// Check for different types of log entries
		hasRequestLog := false
		hasResponseLog := false
		hasErrorLog := false
		hasAuditLog := false
		
		for _, line := range logLines {
			if strings.Contains(line, "Request received") {
				hasRequestLog = true
			}
			if strings.Contains(line, "Request completed") {
				hasResponseLog = true
			}
			if strings.Contains(line, "error occurred") {
				hasErrorLog = true
			}
			if strings.Contains(line, "Audit:") {
				hasAuditLog = true
			}
		}
		
		assert.True(t, hasRequestLog, "Should have request logs")
		assert.True(t, hasResponseLog, "Should have response logs")
		assert.True(t, hasErrorLog, "Should have error logs")
		assert.True(t, hasAuditLog, "Should have audit logs")
	})

	t.Run("rate limiting integration", func(t *testing.T) {
		logger := NewStructuredLogger(LogLevelInfo, &bytes.Buffer{})
		errorMiddleware := NewErrorMiddleware(logger)
		rateLimitMiddleware := NewRateLimitMiddleware()
		defer rateLimitMiddleware.Stop()

		router := gin.New()
		router.Use(errorMiddleware.RequestIDMiddleware())
		router.Use(rateLimitMiddleware.AuthLimit()) // 5 requests per minute
		router.Use(errorMiddleware.HandleError())

		router.POST("/auth/login", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "login successful"})
		})

		// Make requests up to the limit
		successCount := 0
		rateLimitedCount := 0
		
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest("POST", "/auth/login", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code == http.StatusOK {
				successCount++
			} else if w.Code == http.StatusTooManyRequests {
				rateLimitedCount++
				
				// Verify rate limit error response
				var errorResponse APIError
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				require.NoError(t, err)
				
				assert.Equal(t, ErrorTypeRateLimit, errorResponse.Type)
				assert.Equal(t, "RATE_LIMIT_EXCEEDED", errorResponse.Code)
				assert.NotEmpty(t, w.Header().Get("Retry-After"))
			}
		}
		
		assert.Equal(t, 5, successCount, "Should allow exactly 5 auth requests")
		assert.Equal(t, 5, rateLimitedCount, "Should rate limit 5 requests")
	})

	t.Run("security headers integration", func(t *testing.T) {
		logger := NewStructuredLogger(LogLevelInfo, &bytes.Buffer{})
		errorMiddleware := NewErrorMiddleware(logger)

		router := gin.New()
		router.Use(errorMiddleware.SecurityHeadersMiddleware())
		router.Use(errorMiddleware.RequestSanitizationMiddleware())

		router.GET("/api/secure", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "secure endpoint"})
		})

		req := httptest.NewRequest("GET", "/api/secure", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		req.Header.Set("User-Agent", "Mozilla/5.0 <script>alert('xss')</script>")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify security headers
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
		assert.Equal(t, "default-src 'self'", w.Header().Get("Content-Security-Policy"))
	})

	t.Run("error correlation across middleware", func(t *testing.T) {
		var logBuffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &logBuffer)
		
		errorMiddleware := NewErrorMiddleware(logger)
		loggingMiddleware := NewLoggingMiddleware(LoggingConfig{
			Logger: logger,
		})

		router := gin.New()
		router.Use(errorMiddleware.RequestIDMiddleware())
		router.Use(loggingMiddleware.Middleware())
		router.Use(errorMiddleware.HandleError())

		router.GET("/api/correlated-error", func(c *gin.Context) {
			c.Error(NewInternalError("Correlated error test"))
		})

		req := httptest.NewRequest("GET", "/api/correlated-error", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		// Get request ID from response
		var errorResponse APIError
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		require.NoError(t, err)
		requestID := errorResponse.RequestID
		assert.NotEmpty(t, requestID)

		// Verify request ID appears in all log entries
		logOutput := logBuffer.String()
		logLines := strings.Split(strings.TrimSpace(logOutput), "\n")
		
		for _, line := range logLines {
			if strings.TrimSpace(line) != "" {
				var logEntry LogEntry
				err := json.Unmarshal([]byte(line), &logEntry)
				require.NoError(t, err)
				assert.Equal(t, requestID, logEntry.RequestID, "All log entries should have the same request ID")
			}
		}
	})
}

func TestMiddlewarePerformance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("middleware stack performance", func(t *testing.T) {
		logger := NewStructuredLogger(LogLevelError, &bytes.Buffer{}) // Only log errors to reduce overhead
		errorMiddleware := NewErrorMiddleware(logger)
		rateLimitMiddleware := NewRateLimitMiddleware()
		defer rateLimitMiddleware.Stop()
		
		loggingMiddleware := NewLoggingMiddleware(LoggingConfig{
			Logger:          logger,
			LogRequestBody:  false,
			LogResponseBody: false,
		})

		router := gin.New()
		router.Use(errorMiddleware.ErrorHandler())
		router.Use(errorMiddleware.RequestIDMiddleware())
		router.Use(errorMiddleware.SecurityHeadersMiddleware())
		router.Use(errorMiddleware.RequestValidationMiddleware())
		router.Use(errorMiddleware.RequestSanitizationMiddleware())
		router.Use(loggingMiddleware.Middleware())
		router.Use(rateLimitMiddleware.IPLimit())
		router.Use(errorMiddleware.HandleError())

		router.GET("/api/perf-test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "performance test"})
		})

		// Measure performance
		numRequests := 1000
		start := time.Now()

		for i := 0; i < numRequests; i++ {
			req := httptest.NewRequest("GET", "/api/perf-test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, http.StatusOK, w.Code)
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(numRequests)

		t.Logf("Processed %d requests in %v (avg: %v per request)", numRequests, duration, avgDuration)
		
		// Performance assertion - should handle requests reasonably fast
		// This is a rough benchmark, adjust based on requirements
		assert.Less(t, avgDuration, 10*time.Millisecond, "Average request processing should be under 10ms")
	})
}