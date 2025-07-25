package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogLevel(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "DEBUG"},
		{LogLevelInfo, "INFO"},
		{LogLevelWarn, "WARN"},
		{LogLevelError, "ERROR"},
		{LogLevelFatal, "FATAL"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

func TestStructuredLogger(t *testing.T) {
	t.Run("NewStructuredLogger with default output", func(t *testing.T) {
		logger := NewStructuredLogger(LogLevelInfo, nil)
		assert.NotNil(t, logger)
		assert.Equal(t, LogLevelInfo, logger.level)
		assert.Equal(t, os.Stdout, logger.output)
	})

	t.Run("logs at appropriate levels", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelWarn, &buffer)

		ctx := context.Background()
		fields := map[string]interface{}{"test": "value"}

		// Debug and Info should be filtered out
		logger.Debug(ctx, "debug message", fields)
		logger.Info(ctx, "info message", fields)
		
		// Warn, Error, Fatal should be logged
		logger.Warn(ctx, "warn message", fields)
		logger.Error(ctx, "error message", fields)
		logger.Fatal(ctx, "fatal message", fields)

		output := buffer.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		// Should have 3 lines (warn, error, fatal)
		assert.Len(t, lines, 3)
		
		// Check that each line is valid JSON
		for _, line := range lines {
			var entry LogEntry
			err := json.Unmarshal([]byte(line), &entry)
			require.NoError(t, err)
		}
	})

	t.Run("includes context values in log entries", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)

		ctx := context.WithValue(context.Background(), "request_id", "test-request-id")
		ctx = context.WithValue(ctx, "user_id", "test-user-id")

		logger.Info(ctx, "test message", map[string]interface{}{"key": "value"})

		var entry LogEntry
		err := json.Unmarshal(buffer.Bytes(), &entry)
		require.NoError(t, err)

		assert.Equal(t, "test-request-id", entry.RequestID)
		assert.Equal(t, "test-user-id", entry.UserID)
		assert.Equal(t, "test message", entry.Message)
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "value", entry.Fields["key"])
	})

	t.Run("handles nil context gracefully", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)

		logger.Info(nil, "test message", nil)

		var entry LogEntry
		err := json.Unmarshal(buffer.Bytes(), &entry)
		require.NoError(t, err)

		assert.Equal(t, "test message", entry.Message)
		assert.Empty(t, entry.RequestID)
		assert.Empty(t, entry.UserID)
	})

	t.Run("handles marshaling errors gracefully", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)

		// Create a field that can't be marshaled to JSON
		fields := map[string]interface{}{
			"invalid": make(chan int),
		}

		// Should not panic
		logger.Info(context.Background(), "test message", fields)
		
		// The error is logged to the default logger, not our buffer
		// So we just check that it didn't panic and buffer is empty
		output := buffer.String()
		assert.Empty(t, output)
	})
}

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("NewLoggingMiddleware with default config", func(t *testing.T) {
		middleware := NewLoggingMiddleware(LoggingConfig{})
		assert.NotNil(t, middleware)
		assert.NotNil(t, middleware.logger)
		assert.Equal(t, int64(1024*1024), middleware.maxBodySize)
	})

	t.Run("logs request and response", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)
		
		middleware := NewLoggingMiddleware(LoggingConfig{
			Logger: logger,
		})

		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test?param=value", nil)
		req.Header.Set("User-Agent", "test-agent")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		output := buffer.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		// Should have 2 lines (request received, request completed)
		assert.Len(t, lines, 2)

		// Parse first log entry (request received)
		var requestEntry LogEntry
		err := json.Unmarshal([]byte(lines[0]), &requestEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "Request received", requestEntry.Message)
		assert.Equal(t, "GET", requestEntry.Fields["method"])
		assert.Equal(t, "/test", requestEntry.Fields["path"])
		assert.Equal(t, "param=value", requestEntry.Fields["query"])
		assert.Equal(t, "test-agent", requestEntry.Fields["user_agent"])

		// Parse second log entry (request completed)
		var responseEntry LogEntry
		err = json.Unmarshal([]byte(lines[1]), &responseEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "Request completed", responseEntry.Message)
		assert.Equal(t, float64(200), responseEntry.Fields["status_code"])
		assert.NotEmpty(t, responseEntry.Fields["duration"])
	})

	t.Run("skips configured paths", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)
		
		middleware := NewLoggingMiddleware(LoggingConfig{
			Logger:    logger,
			SkipPaths: []string{"/health", "/metrics"},
		})

		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// Request to /health should be skipped
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Request to /test should be logged
		req = httptest.NewRequest("GET", "/test", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		output := buffer.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		// Should have 2 lines only for /test request
		assert.Len(t, lines, 2)
		
		// Both lines should be for /test
		for _, line := range lines {
			var entry LogEntry
			err := json.Unmarshal([]byte(line), &entry)
			require.NoError(t, err)
			assert.Equal(t, "/test", entry.Fields["path"])
		}
	})

	t.Run("logs request body when enabled", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)
		
		middleware := NewLoggingMiddleware(LoggingConfig{
			Logger:         logger,
			LogRequestBody: true,
			MaxBodySize:    1024,
		})

		router := gin.New()
		router.Use(middleware.Middleware())
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		requestBody := `{"test": "data"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		output := buffer.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		// Parse request log entry
		var requestEntry LogEntry
		err := json.Unmarshal([]byte(lines[0]), &requestEntry)
		require.NoError(t, err)
		
		assert.Contains(t, requestEntry.Fields["request_body"], "test")
		assert.Contains(t, requestEntry.Fields["request_body"], "data")
	})

	t.Run("logs response body when enabled", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)
		
		middleware := NewLoggingMiddleware(LoggingConfig{
			Logger:          logger,
			LogResponseBody: true,
			MaxBodySize:     1024,
		})

		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test response"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		output := buffer.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		// Parse response log entry
		var responseEntry LogEntry
		err := json.Unmarshal([]byte(lines[1]), &responseEntry)
		require.NoError(t, err)
		
		responseBody := responseEntry.Fields["response_body"].(string)
		assert.Contains(t, responseBody, "test response")
	})

	t.Run("sanitizes sensitive headers", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)
		
		middleware := NewLoggingMiddleware(LoggingConfig{
			Logger: logger,
		})

		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer secret-token")
		req.Header.Set("Cookie", "session=secret-session")
		req.Header.Set("X-API-Key", "secret-api-key")
		req.Header.Set("User-Agent", "test-agent")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		output := buffer.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		// Parse request log entry
		var requestEntry LogEntry
		err := json.Unmarshal([]byte(lines[0]), &requestEntry)
		require.NoError(t, err)
		
		headers := requestEntry.Fields["headers"].(map[string]interface{})
		
		// Sensitive headers should be redacted
		assert.Equal(t, []interface{}{"[REDACTED]"}, headers["Authorization"])
		assert.Equal(t, []interface{}{"[REDACTED]"}, headers["Cookie"])
		assert.Equal(t, []interface{}{"[REDACTED]"}, headers["X-Api-Key"])
		
		// Non-sensitive headers should be preserved
		assert.Equal(t, []interface{}{"test-agent"}, headers["User-Agent"])
	})

	t.Run("handles errors in requests", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)
		
		middleware := NewLoggingMiddleware(LoggingConfig{
			Logger: logger,
		})

		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/error", func(c *gin.Context) {
			c.Error(NewInternalError("test error"))
			c.JSON(500, gin.H{"error": "internal error"})
		})

		req := httptest.NewRequest("GET", "/error", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		output := buffer.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		// Parse response log entry
		var responseEntry LogEntry
		err := json.Unmarshal([]byte(lines[1]), &responseEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "ERROR", responseEntry.Level) // Should log as ERROR for 5xx status
		assert.NotEmpty(t, responseEntry.Fields["errors"])
	})

	t.Run("handles concurrent requests safely", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)
		
		middleware := NewLoggingMiddleware(LoggingConfig{
			Logger: logger,
		})

		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/test", func(c *gin.Context) {
			time.Sleep(10 * time.Millisecond) // Simulate some work
			c.JSON(200, gin.H{"message": "ok"})
		})

		var wg sync.WaitGroup
		numRequests := 10

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("X-Request-Index", string(rune(index)))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
			}(i)
		}

		wg.Wait()

		output := buffer.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		// Should have 2 lines per request (request + response), but due to concurrency
		// some lines might be interleaved, so we check for at least the minimum
		assert.GreaterOrEqual(t, len(lines), numRequests)
		
		// All lines should be valid JSON
		for _, line := range lines {
			var entry LogEntry
			err := json.Unmarshal([]byte(line), &entry)
			require.NoError(t, err)
		}
	})
}

func TestAuditLogger(t *testing.T) {
	t.Run("NewAuditLogger with default logger", func(t *testing.T) {
		auditLogger := NewAuditLogger(nil)
		assert.NotNil(t, auditLogger)
		assert.NotNil(t, auditLogger.logger)
	})

	t.Run("LogAction logs audit events", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)
		auditLogger := NewAuditLogger(logger)

		ctx := context.WithValue(context.Background(), "request_id", "test-request-id")
		fields := map[string]interface{}{
			"user_id": "test-user",
			"ip":      "192.168.1.1",
		}

		auditLogger.LogAction(ctx, "CREATE", "document", fields)

		var entry LogEntry
		err := json.Unmarshal(buffer.Bytes(), &entry)
		require.NoError(t, err)

		assert.Equal(t, "Audit: CREATE on document", entry.Message)
		assert.Equal(t, "CREATE", entry.Fields["action"])
		assert.Equal(t, "document", entry.Fields["resource"])
		assert.Equal(t, true, entry.Fields["audit"])
		assert.Equal(t, "test-user", entry.Fields["user_id"])
		assert.Equal(t, "192.168.1.1", entry.Fields["ip"])
	})

	t.Run("LogDocumentAction logs document-specific events", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)
		auditLogger := NewAuditLogger(logger)

		ctx := context.Background()
		metadata := map[string]interface{}{
			"filename": "test.pdf",
			"size":     1024,
		}

		auditLogger.LogDocumentAction(ctx, "SIGN", "doc-123", "user-456", metadata)

		var entry LogEntry
		err := json.Unmarshal(buffer.Bytes(), &entry)
		require.NoError(t, err)

		assert.Equal(t, "Audit: SIGN on document", entry.Message)
		assert.Equal(t, "SIGN", entry.Fields["action"])
		assert.Equal(t, "document", entry.Fields["resource"])
		assert.Equal(t, "doc-123", entry.Fields["document_id"])
		assert.Equal(t, "user-456", entry.Fields["user_id"])
		assert.Equal(t, "test.pdf", entry.Fields["filename"])
		assert.Equal(t, float64(1024), entry.Fields["size"])
	})

	t.Run("LogAuthAction logs authentication events", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)
		auditLogger := NewAuditLogger(logger)

		ctx := context.Background()
		metadata := map[string]interface{}{
			"user_agent": "Mozilla/5.0",
			"method":     "password",
		}

		auditLogger.LogAuthAction(ctx, "LOGIN", "user-123", "192.168.1.1", true, metadata)

		var entry LogEntry
		err := json.Unmarshal(buffer.Bytes(), &entry)
		require.NoError(t, err)

		assert.Equal(t, "Audit: LOGIN on authentication", entry.Message)
		assert.Equal(t, "LOGIN", entry.Fields["action"])
		assert.Equal(t, "authentication", entry.Fields["resource"])
		assert.Equal(t, "user-123", entry.Fields["user_id"])
		assert.Equal(t, "192.168.1.1", entry.Fields["ip"])
		assert.Equal(t, true, entry.Fields["success"])
		assert.Equal(t, "Mozilla/5.0", entry.Fields["user_agent"])
		assert.Equal(t, "password", entry.Fields["method"])
	})

	t.Run("LogVerificationAction logs verification events", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)
		auditLogger := NewAuditLogger(logger)

		ctx := context.Background()
		metadata := map[string]interface{}{
			"hash_match":      true,
			"signature_valid": true,
		}

		auditLogger.LogVerificationAction(ctx, "doc-789", "10.0.0.1", "VALID", metadata)

		var entry LogEntry
		err := json.Unmarshal(buffer.Bytes(), &entry)
		require.NoError(t, err)

		assert.Equal(t, "Audit: verify on document", entry.Message)
		assert.Equal(t, "verify", entry.Fields["action"])
		assert.Equal(t, "document", entry.Fields["resource"])
		assert.Equal(t, "doc-789", entry.Fields["document_id"])
		assert.Equal(t, "10.0.0.1", entry.Fields["ip"])
		assert.Equal(t, "VALID", entry.Fields["result"])
		assert.Equal(t, true, entry.Fields["hash_match"])
		assert.Equal(t, true, entry.Fields["signature_valid"])
	})

	t.Run("handles nil metadata gracefully", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)
		auditLogger := NewAuditLogger(logger)

		ctx := context.Background()

		auditLogger.LogDocumentAction(ctx, "DELETE", "doc-123", "user-456", nil)

		var entry LogEntry
		err := json.Unmarshal(buffer.Bytes(), &entry)
		require.NoError(t, err)

		assert.Equal(t, "Audit: DELETE on document", entry.Message)
		assert.Equal(t, "doc-123", entry.Fields["document_id"])
		assert.Equal(t, "user-456", entry.Fields["user_id"])
	})
}

func TestResponseWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("captures response body correctly", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)
		
		middleware := NewLoggingMiddleware(LoggingConfig{
			Logger:          logger,
			LogResponseBody: true,
		})

		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test response", "data": []int{1, 2, 3}})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify actual response
		var actualResponse map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
		require.NoError(t, err)
		assert.Equal(t, "test response", actualResponse["message"])

		// Verify logged response
		output := buffer.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		var responseEntry LogEntry
		err = json.Unmarshal([]byte(lines[1]), &responseEntry)
		require.NoError(t, err)
		
		responseBody := responseEntry.Fields["response_body"].(string)
		var loggedResponse map[string]interface{}
		err = json.Unmarshal([]byte(responseBody), &loggedResponse)
		require.NoError(t, err)
		assert.Equal(t, "test response", loggedResponse["message"])
	})

	t.Run("truncates large response bodies", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := NewStructuredLogger(LogLevelInfo, &buffer)
		
		middleware := NewLoggingMiddleware(LoggingConfig{
			Logger:          logger,
			LogResponseBody: true,
			MaxBodySize:     10, // Very small limit
		})

		router := gin.New()
		router.Use(middleware.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "this is a very long response that should be truncated"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		output := buffer.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		var responseEntry LogEntry
		err := json.Unmarshal([]byte(lines[1]), &responseEntry)
		require.NoError(t, err)
		
		responseBody := responseEntry.Fields["response_body"].(string)
		assert.Contains(t, responseBody, "[TRUNCATED - Size:")
	})
}