package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)



// AuditEvent represents a single audit event
type AuditEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   AuditEventType         `json:"event_type"`
	Severity    AuditSeverity          `json:"severity"`
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Resource    string                 `json:"resource,omitempty"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	Action      string                 `json:"action"`
	Result      string                 `json:"result"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Duration    *time.Duration         `json:"duration,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AuditService provides comprehensive audit logging functionality
type AuditService struct {
	writer       io.Writer
	logFile      *os.File
	logDir       string
	maxFileSize  int64
	maxFiles     int
	rotationTime time.Duration
	mu           sync.RWMutex
	buffer       chan *AuditEvent
	bufferSize   int
	flushTicker  *time.Ticker
	stopChan     chan struct{}
}

// AuditConfig represents audit service configuration
type AuditConfig struct {
	LogDir       string        // Directory for audit logs
	MaxFileSize  int64         // Maximum file size before rotation (bytes)
	MaxFiles     int           // Maximum number of log files to keep
	RotationTime time.Duration // Time-based rotation interval
	BufferSize   int           // Buffer size for async logging
	FlushInterval time.Duration // Interval for flushing buffered logs
}

// NewAuditService creates a new audit service
func NewAuditService(config AuditConfig) (*AuditService, error) {
	if config.LogDir == "" {
		config.LogDir = "./logs/audit"
	}
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 100 * 1024 * 1024 // 100MB
	}
	if config.MaxFiles == 0 {
		config.MaxFiles = 30 // Keep 30 files
	}
	if config.RotationTime == 0 {
		config.RotationTime = 24 * time.Hour // Daily rotation
	}
	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 5 * time.Second
	}

	// Create log directory
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}

	service := &AuditService{
		logDir:       config.LogDir,
		maxFileSize:  config.MaxFileSize,
		maxFiles:     config.MaxFiles,
		rotationTime: config.RotationTime,
		buffer:       make(chan *AuditEvent, config.BufferSize),
		bufferSize:   config.BufferSize,
		flushTicker:  time.NewTicker(config.FlushInterval),
		stopChan:     make(chan struct{}),
	}

	// Initialize log file
	if err := service.initLogFile(); err != nil {
		return nil, fmt.Errorf("failed to initialize audit log file: %w", err)
	}

	// Start background processing
	go service.processEvents()

	return service, nil
}

// initLogFile initializes the current log file
func (as *AuditService) initLogFile() error {
	filename := fmt.Sprintf("audit_%s.log", time.Now().Format("2006-01-02"))
	filepath := filepath.Join(as.logDir, filename)

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	as.mu.Lock()
	if as.logFile != nil {
		as.logFile.Close()
	}
	as.logFile = file
	as.writer = file
	as.mu.Unlock()

	return nil
}

// LogEvent logs an audit event asynchronously
func (as *AuditService) LogEvent(ctx context.Context, eventType AuditEventType, severity AuditSeverity, action, result, message string, details map[string]interface{}) {
	event := &AuditEvent{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		EventType: eventType,
		Severity:  severity,
		Action:    action,
		Result:    result,
		Message:   message,
		Details:   details,
	}

	// Extract context information
	if ctx != nil {
		if userID := ctx.Value("user_id"); userID != nil {
			event.UserID = fmt.Sprintf("%v", userID)
		}
		if sessionID := ctx.Value("session_id"); sessionID != nil {
			event.SessionID = fmt.Sprintf("%v", sessionID)
		}
		if requestID := ctx.Value("request_id"); requestID != nil {
			event.RequestID = fmt.Sprintf("%v", requestID)
		}
		if ip := ctx.Value("client_ip"); ip != nil {
			event.IPAddress = fmt.Sprintf("%v", ip)
		}
		if userAgent := ctx.Value("user_agent"); userAgent != nil {
			event.UserAgent = fmt.Sprintf("%v", userAgent)
		}
	}

	// Try to send to buffer, fallback to synchronous logging if buffer is full
	select {
	case as.buffer <- event:
		// Successfully buffered
	default:
		// Buffer is full, log synchronously
		as.writeEvent(event)
	}
}

// LogAuthEvent logs authentication-related events
func (as *AuditService) LogAuthEvent(ctx context.Context, eventType AuditEventType, userID, ip string, success bool, details map[string]interface{}) {
	severity := AuditSeverityInfo
	result := "success"
	
	if !success {
		severity = AuditSeverityWarning
		result = "failure"
	}

	if details == nil {
		details = make(map[string]interface{})
	}
	details["user_id"] = userID
	details["ip_address"] = ip
	details["success"] = success

	message := fmt.Sprintf("Authentication %s for user %s from IP %s", result, userID, ip)
	as.LogEvent(ctx, eventType, severity, string(eventType), result, message, details)
}

// LogDocumentEvent logs document-related events
func (as *AuditService) LogDocumentEvent(ctx context.Context, eventType AuditEventType, documentID, userID string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["document_id"] = documentID
	details["user_id"] = userID

	message := fmt.Sprintf("Document %s: %s by user %s", string(eventType), documentID, userID)
	as.LogEvent(ctx, eventType, AuditSeverityInfo, string(eventType), "success", message, details)
}

// LogVerificationEvent logs verification-related events
func (as *AuditService) LogVerificationEvent(ctx context.Context, documentID, ip, result string, duration time.Duration, details map[string]interface{}) {
	eventType := AuditEventVerificationAttempt
	severity := AuditSeverityInfo

	switch result {
	case "valid":
		eventType = AuditEventVerificationSuccess
	case "invalid", "tampered", "signature_invalid":
		eventType = AuditEventVerificationFailed
		severity = AuditSeverityWarning
	}

	if details == nil {
		details = make(map[string]interface{})
	}
	details["document_id"] = documentID
	details["ip_address"] = ip
	details["verification_result"] = result
	details["duration_ms"] = duration.Milliseconds()

	message := fmt.Sprintf("Document verification %s for document %s from IP %s", result, documentID, ip)
	
	event := &AuditEvent{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		EventType: eventType,
		Severity:  severity,
		IPAddress: ip,
		Resource:  "document",
		ResourceID: documentID,
		Action:    "verify",
		Result:    result,
		Message:   message,
		Details:   details,
		Duration:  &duration,
	}

	// Extract context information
	if ctx != nil {
		if requestID := ctx.Value("request_id"); requestID != nil {
			event.RequestID = fmt.Sprintf("%v", requestID)
		}
		if userAgent := ctx.Value("user_agent"); userAgent != nil {
			event.UserAgent = fmt.Sprintf("%v", userAgent)
		}
	}

	// Send to buffer
	select {
	case as.buffer <- event:
	default:
		as.writeEvent(event)
	}
}

// LogSecurityEvent logs security-related events
func (as *AuditService) LogSecurityEvent(ctx context.Context, eventType AuditEventType, severity AuditSeverity, message string, details map[string]interface{}) {
	as.LogEvent(ctx, eventType, severity, "security_event", "detected", message, details)
}

// LogSystemEvent logs system-related events
func (as *AuditService) LogSystemEvent(ctx context.Context, eventType AuditEventType, message string, details map[string]interface{}) {
	as.LogEvent(ctx, eventType, AuditSeverityInfo, "system_event", "completed", message, details)
}

// processEvents processes buffered events in background
func (as *AuditService) processEvents() {
	for {
		select {
		case event := <-as.buffer:
			as.writeEvent(event)
		case <-as.flushTicker.C:
			as.flush()
		case <-as.stopChan:
			// Drain remaining events
			for len(as.buffer) > 0 {
				event := <-as.buffer
				as.writeEvent(event)
			}
			return
		}
	}
}

// writeEvent writes an event to the log file
func (as *AuditService) writeEvent(event *AuditEvent) {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Check if rotation is needed
	if as.needsRotation() {
		if err := as.rotateLog(); err != nil {
			// Log rotation failed, continue with current file
			fmt.Printf("Failed to rotate audit log: %v\n", err)
		}
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Failed to marshal audit event: %v\n", err)
		return
	}

	// Write to file
	if _, err := as.writer.Write(append(data, '\n')); err != nil {
		fmt.Printf("Failed to write audit event: %v\n", err)
	}
}

// needsRotation checks if log rotation is needed
func (as *AuditService) needsRotation() bool {
	if as.logFile == nil {
		return true
	}

	// Check file size
	if stat, err := as.logFile.Stat(); err == nil {
		if stat.Size() >= as.maxFileSize {
			return true
		}
	}

	// Check time-based rotation (daily)
	filename := filepath.Base(as.logFile.Name())
	expectedFilename := fmt.Sprintf("audit_%s.log", time.Now().Format("2006-01-02"))
	
	return filename != expectedFilename
}

// rotateLog rotates the current log file
func (as *AuditService) rotateLog() error {
	// Close current file
	if as.logFile != nil {
		as.logFile.Close()
	}

	// Clean up old files
	as.cleanupOldFiles()

	// Create new file
	return as.initLogFile()
}

// cleanupOldFiles removes old log files beyond the retention limit
func (as *AuditService) cleanupOldFiles() {
	files, err := filepath.Glob(filepath.Join(as.logDir, "audit_*.log"))
	if err != nil {
		return
	}

	if len(files) <= as.maxFiles {
		return
	}

	// Sort files by modification time (oldest first)
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	var fileInfos []fileInfo
	for _, file := range files {
		if stat, err := os.Stat(file); err == nil {
			fileInfos = append(fileInfos, fileInfo{
				path:    file,
				modTime: stat.ModTime(),
			})
		}
	}

	// Sort by modification time
	for i := 0; i < len(fileInfos)-1; i++ {
		for j := i + 1; j < len(fileInfos); j++ {
			if fileInfos[i].modTime.After(fileInfos[j].modTime) {
				fileInfos[i], fileInfos[j] = fileInfos[j], fileInfos[i]
			}
		}
	}

	// Remove oldest files
	filesToRemove := len(fileInfos) - as.maxFiles
	for i := 0; i < filesToRemove; i++ {
		os.Remove(fileInfos[i].path)
	}
}

// flush flushes any buffered data to disk
func (as *AuditService) flush() {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.logFile != nil {
		as.logFile.Sync()
	}
}

// Close closes the audit service and flushes remaining events
func (as *AuditService) Close() error {
	close(as.stopChan)
	as.flushTicker.Stop()

	as.mu.Lock()
	defer as.mu.Unlock()

	if as.logFile != nil {
		return as.logFile.Close()
	}

	return nil
}

// GetAuditStats returns statistics about audit logging
func (as *AuditService) GetAuditStats() map[string]interface{} {
	as.mu.RLock()
	defer as.mu.RUnlock()

	stats := map[string]interface{}{
		"buffer_size":     as.bufferSize,
		"buffered_events": len(as.buffer),
		"log_directory":   as.logDir,
		"max_file_size":   as.maxFileSize,
		"max_files":       as.maxFiles,
	}

	if as.logFile != nil {
		if stat, err := as.logFile.Stat(); err == nil {
			stats["current_file_size"] = stat.Size()
			stats["current_file_name"] = stat.Name()
		}
	}

	return stats
}