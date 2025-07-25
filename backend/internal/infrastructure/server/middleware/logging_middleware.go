package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LogLevel represents different log levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Method      string                 `json:"method,omitempty"`
	Path        string                 `json:"path,omitempty"`
	StatusCode  int                    `json:"status_code,omitempty"`
	Duration    string                 `json:"duration,omitempty"`
	IP          string                 `json:"ip,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
}

// StructuredLogger provides structured logging functionality
type StructuredLogger struct {
	level  LogLevel
	output io.Writer
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(level LogLevel, output io.Writer) *StructuredLogger {
	if output == nil {
		output = os.Stdout
	}
	return &StructuredLogger{
		level:  level,
		output: output,
	}
}

// Log writes a log entry
func (l *StructuredLogger) Log(level LogLevel, ctx context.Context, message string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level.String(),
		Message:   message,
		Fields:    fields,
	}

	// Extract context values if available
	if ctx != nil {
		if requestID := ctx.Value("request_id"); requestID != nil {
			entry.RequestID = fmt.Sprintf("%v", requestID)
		}
		if userID := ctx.Value("user_id"); userID != nil {
			entry.UserID = fmt.Sprintf("%v", userID)
		}
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	// Write to output
	fmt.Fprintln(l.output, string(jsonData))
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(ctx context.Context, message string, fields map[string]interface{}) {
	l.Log(LogLevelDebug, ctx, message, fields)
}

// Info logs an info message
func (l *StructuredLogger) Info(ctx context.Context, message string, fields map[string]interface{}) {
	l.Log(LogLevelInfo, ctx, message, fields)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(ctx context.Context, message string, fields map[string]interface{}) {
	l.Log(LogLevelWarn, ctx, message, fields)
}

// Error logs an error message
func (l *StructuredLogger) Error(ctx context.Context, message string, fields map[string]interface{}) {
	l.Log(LogLevelError, ctx, message, fields)
}

// Fatal logs a fatal message
func (l *StructuredLogger) Fatal(ctx context.Context, message string, fields map[string]interface{}) {
	l.Log(LogLevelFatal, ctx, message, fields)
}

// LoggingMiddleware provides comprehensive request logging
type LoggingMiddleware struct {
	logger        *StructuredLogger
	skipPaths     []string
	logRequestBody bool
	logResponseBody bool
	maxBodySize   int64
}

// LoggingConfig represents logging middleware configuration
type LoggingConfig struct {
	Logger          *StructuredLogger
	SkipPaths       []string
	LogRequestBody  bool
	LogResponseBody bool
	MaxBodySize     int64 // Maximum body size to log (in bytes)
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(config LoggingConfig) *LoggingMiddleware {
	if config.Logger == nil {
		config.Logger = NewStructuredLogger(LogLevelInfo, os.Stdout)
	}
	if config.MaxBodySize == 0 {
		config.MaxBodySize = 1024 * 1024 // 1MB default
	}

	return &LoggingMiddleware{
		logger:          config.Logger,
		skipPaths:       config.SkipPaths,
		logRequestBody:  config.LogRequestBody,
		logResponseBody: config.LogResponseBody,
		maxBodySize:     config.MaxBodySize,
	}
}

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Middleware returns the logging middleware
func (lm *LoggingMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for specified paths
		for _, path := range lm.skipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		start := time.Now()
		requestID := lm.getOrCreateRequestID(c)
		
		// Add request ID to context
		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)

		// Log request
		lm.logRequest(c, requestID)

		// Capture response if needed
		var respWriter *responseWriter
		if lm.logResponseBody {
			respWriter = &responseWriter{
				ResponseWriter: c.Writer,
				body:          bytes.NewBuffer([]byte{}),
			}
			c.Writer = respWriter
		}

		// Process request
		c.Next()

		// Log response
		duration := time.Since(start)
		lm.logResponse(c, requestID, duration, respWriter)
	}
}

// getOrCreateRequestID gets or creates a request ID
func (lm *LoggingMiddleware) getOrCreateRequestID(c *gin.Context) string {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		if id, exists := c.Get("request_id"); exists {
			requestID = id.(string)
		} else {
			requestID = uuid.New().String()
		}
	}
	c.Set("request_id", requestID)
	c.Header("X-Request-ID", requestID)
	return requestID
}

// logRequest logs the incoming request
func (lm *LoggingMiddleware) logRequest(c *gin.Context, requestID string) {
	fields := map[string]interface{}{
		"request_id": requestID,
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"query":      c.Request.URL.RawQuery,
		"ip":         c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"referer":    c.Request.Referer(),
		"headers":    lm.sanitizeHeaders(c.Request.Header),
	}

	// Add user ID if available
	if userID, exists := c.Get("user_id"); exists {
		fields["user_id"] = userID
	}

	// Log request body if enabled
	if lm.logRequestBody && c.Request.Body != nil {
		if body := lm.readBody(c.Request.Body, lm.maxBodySize); body != "" {
			fields["request_body"] = body
		}
	}

	lm.logger.Info(c.Request.Context(), "Request received", fields)
}

// logResponse logs the response
func (lm *LoggingMiddleware) logResponse(c *gin.Context, requestID string, duration time.Duration, respWriter *responseWriter) {
	fields := map[string]interface{}{
		"request_id":  requestID,
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"status_code": c.Writer.Status(),
		"duration":    duration.String(),
		"size":        c.Writer.Size(),
		"ip":          c.ClientIP(),
	}

	// Add user ID if available
	if userID, exists := c.Get("user_id"); exists {
		fields["user_id"] = userID
	}

	// Log response body if enabled and captured
	if lm.logResponseBody && respWriter != nil && respWriter.body.Len() > 0 {
		if respWriter.body.Len() <= int(lm.maxBodySize) {
			fields["response_body"] = respWriter.body.String()
		} else {
			fields["response_body"] = fmt.Sprintf("[TRUNCATED - Size: %d bytes]", respWriter.body.Len())
		}
	}

	// Log errors if any
	if len(c.Errors) > 0 {
		errors := make([]string, len(c.Errors))
		for i, err := range c.Errors {
			errors[i] = err.Error()
		}
		fields["errors"] = errors
	}

	// Determine log level based on status code
	statusCode := c.Writer.Status()
	message := "Request completed"

	switch {
	case statusCode >= 500:
		lm.logger.Error(c.Request.Context(), message, fields)
	case statusCode >= 400:
		lm.logger.Warn(c.Request.Context(), message, fields)
	default:
		lm.logger.Info(c.Request.Context(), message, fields)
	}
}

// readBody reads and returns the body content
func (lm *LoggingMiddleware) readBody(body io.ReadCloser, maxSize int64) string {
	if body == nil {
		return ""
	}

	// Read limited amount of data
	limitedReader := io.LimitReader(body, maxSize)
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return fmt.Sprintf("[ERROR reading body: %v]", err)
	}

	// Try to restore the body for further processing
	if seeker, ok := body.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	return string(bodyBytes)
}

// sanitizeHeaders removes sensitive headers from logging
func (lm *LoggingMiddleware) sanitizeHeaders(headers map[string][]string) map[string][]string {
	sensitiveHeaders := []string{
		"authorization",
		"cookie",
		"x-api-key",
		"x-auth-token",
		"x-session-token",
	}

	sanitized := make(map[string][]string)
	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		isSensitive := false
		
		for _, sensitive := range sensitiveHeaders {
			if lowerKey == sensitive {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			sanitized[key] = []string{"[REDACTED]"}
		} else {
			sanitized[key] = values
		}
	}

	return sanitized
}

// AuditLogger provides audit logging functionality
type AuditLogger struct {
	logger *StructuredLogger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *StructuredLogger) *AuditLogger {
	if logger == nil {
		logger = NewStructuredLogger(LogLevelInfo, os.Stdout)
	}
	return &AuditLogger{logger: logger}
}

// LogAction logs an audit action
func (al *AuditLogger) LogAction(ctx context.Context, action, resource string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	
	fields["action"] = action
	fields["resource"] = resource
	fields["audit"] = true

	al.logger.Info(ctx, fmt.Sprintf("Audit: %s on %s", action, resource), fields)
}

// LogDocumentAction logs document-related actions
func (al *AuditLogger) LogDocumentAction(ctx context.Context, action, documentID, userID string, metadata map[string]interface{}) {
	fields := map[string]interface{}{
		"document_id": documentID,
		"user_id":     userID,
	}
	
	for k, v := range metadata {
		fields[k] = v
	}
	
	al.LogAction(ctx, action, "document", fields)
}

// LogAuthAction logs authentication-related actions
func (al *AuditLogger) LogAuthAction(ctx context.Context, action, userID, ip string, success bool, metadata map[string]interface{}) {
	fields := map[string]interface{}{
		"user_id": userID,
		"ip":      ip,
		"success": success,
	}
	
	for k, v := range metadata {
		fields[k] = v
	}
	
	al.LogAction(ctx, action, "authentication", fields)
}

// LogVerificationAction logs verification-related actions
func (al *AuditLogger) LogVerificationAction(ctx context.Context, documentID, ip string, result string, metadata map[string]interface{}) {
	fields := map[string]interface{}{
		"document_id": documentID,
		"ip":          ip,
		"result":      result,
	}
	
	for k, v := range metadata {
		fields[k] = v
	}
	
	al.LogAction(ctx, "verify", "document", fields)
}