package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// AuditEvent represents different types of audit events
type AuditEvent string

const (
	// Authentication events
	AuditEventLogin          AuditEvent = "LOGIN"
	AuditEventLogout         AuditEvent = "LOGOUT"
	AuditEventRegister       AuditEvent = "REGISTER"
	AuditEventPasswordChange AuditEvent = "PASSWORD_CHANGE"
	AuditEventAuthFailure    AuditEvent = "AUTH_FAILURE"

	// Document events
	AuditEventDocumentSign   AuditEvent = "DOCUMENT_SIGN"
	AuditEventDocumentView   AuditEvent = "DOCUMENT_VIEW"
	AuditEventDocumentDelete AuditEvent = "DOCUMENT_DELETE"
	AuditEventDocumentList   AuditEvent = "DOCUMENT_LIST"

	// Verification events
	AuditEventVerificationAttempt AuditEvent = "VERIFICATION_ATTEMPT"
	AuditEventVerificationSuccess AuditEvent = "VERIFICATION_SUCCESS"
	AuditEventVerificationFailure AuditEvent = "VERIFICATION_FAILURE"

	// Security events
	AuditEventSuspiciousActivity AuditEvent = "SUSPICIOUS_ACTIVITY"
	AuditEventRateLimitExceeded  AuditEvent = "RATE_LIMIT_EXCEEDED"
	AuditEventValidationFailure  AuditEvent = "VALIDATION_FAILURE"
)

// AuditLogEntry represents a structured audit log entry
type AuditLogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Event       AuditEvent             `json:"event"`
	UserID      string                 `json:"user_id,omitempty"`
	Username    string                 `json:"username,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Resource    string                 `json:"resource,omitempty"`
	Action      string                 `json:"action,omitempty"`
	Result      string                 `json:"result"`
	Message     string                 `json:"message,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	DocumentID  string                 `json:"document_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Duration    int64                  `json:"duration_ms,omitempty"`
	ErrorCode   string                 `json:"error_code,omitempty"`
	Severity    string                 `json:"severity"`
}

// AuditLogger handles audit logging with structured JSON format
type AuditLogger struct {
	logger *log.Logger
}

var defaultAuditLogger *AuditLogger

// InitializeAuditLogger sets up the audit logger with file-based logging and rotation
func InitializeAuditLogger(logDir string) error {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create audit log directory: %w", err)
	}

	// Rotate logs if needed
	if err := rotateAuditLogsIfNeeded(logDir); err != nil {
		return fmt.Errorf("failed to rotate audit logs: %w", err)
	}

	// Create audit log file
	auditLogFile := filepath.Join(logDir, "audit.log")

	// Open audit log file
	auditFile, err := os.OpenFile(auditLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}

	// Create multi-writer to write to both file and stdout in development
	var writer io.Writer
	if os.Getenv("ENVIRONMENT") == "development" {
		writer = io.MultiWriter(os.Stdout, auditFile)
	} else {
		writer = auditFile
	}

	// Create logger with no prefix or flags (we'll format ourselves)
	logger := log.New(writer, "", 0)

	defaultAuditLogger = &AuditLogger{
		logger: logger,
	}

	return nil
}

// rotateAuditLogsIfNeeded performs basic log rotation based on file size
func rotateAuditLogsIfNeeded(logDir string) error {
	auditLogFile := filepath.Join(logDir, "audit.log")
	
	// Check if audit log file exists
	info, err := os.Stat(auditLogFile)
	if os.IsNotExist(err) {
		return nil // No file to rotate
	}
	if err != nil {
		return err
	}

	// Rotate if file is larger than 10MB
	const maxLogSize = 10 * 1024 * 1024 // 10MB
	if info.Size() > maxLogSize {
		// Create backup filename with timestamp
		timestamp := time.Now().Format("2006-01-02-15-04-05")
		backupFile := filepath.Join(logDir, fmt.Sprintf("audit.log.%s", timestamp))
		
		// Rename current log file to backup
		if err := os.Rename(auditLogFile, backupFile); err != nil {
			return fmt.Errorf("failed to rotate audit log: %w", err)
		}

		// Clean up old backup files (keep only last 5)
		if err := cleanupOldAuditLogs(logDir, 5); err != nil {
			// Log the error but don't fail initialization
			fmt.Printf("Warning: failed to cleanup old audit logs: %v\n", err)
		}
	}

	return nil
}

// cleanupOldAuditLogs removes old audit log backup files, keeping only the specified number
func cleanupOldAuditLogs(logDir string, keepCount int) error {
	// Find all audit log backup files
	pattern := filepath.Join(logDir, "audit.log.*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// If we have fewer files than the keep count, nothing to do
	if len(matches) <= keepCount {
		return nil
	}

	// Sort files by modification time (oldest first)
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	var files []fileInfo
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue // Skip files we can't stat
		}
		files = append(files, fileInfo{
			path:    match,
			modTime: info.ModTime(),
		})
	}

	// Sort by modification time (oldest first)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].modTime.After(files[j].modTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Remove oldest files, keeping only the specified count
	filesToRemove := len(files) - keepCount
	for i := 0; i < filesToRemove; i++ {
		if err := os.Remove(files[i].path); err != nil {
			fmt.Printf("Warning: failed to remove old audit log %s: %v\n", files[i].path, err)
		}
	}

	return nil
}

// GetAuditLogger returns the default audit logger instance
func GetAuditLogger() *AuditLogger {
	if defaultAuditLogger == nil {
		// Fallback to stdout if not initialized
		defaultAuditLogger = &AuditLogger{
			logger: log.New(os.Stdout, "", 0),
		}
	}
	return defaultAuditLogger
}

// LogEvent logs an audit event with structured data
func (a *AuditLogger) LogEvent(entry AuditLogEntry) {
	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	// Set default severity if not provided
	if entry.Severity == "" {
		entry.Severity = a.getSeverityForEvent(entry.Event)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple logging if JSON marshaling fails
		a.logger.Printf("AUDIT_LOG_ERROR: Failed to marshal audit entry: %v", err)
		return
	}

	// Log the JSON entry
	a.logger.Println(string(jsonData))
}

// getSeverityForEvent returns the default severity level for an event type
func (a *AuditLogger) getSeverityForEvent(event AuditEvent) string {
	switch event {
	case AuditEventAuthFailure, AuditEventSuspiciousActivity, AuditEventRateLimitExceeded:
		return "HIGH"
	case AuditEventVerificationFailure, AuditEventValidationFailure:
		return "MEDIUM"
	case AuditEventLogin, AuditEventLogout, AuditEventDocumentSign, AuditEventDocumentDelete:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// Convenience methods for common audit events

// LogAuthentication logs authentication-related events
func (a *AuditLogger) LogAuthentication(event AuditEvent, userID, username, ipAddress, userAgent string, result string, details map[string]interface{}) {
	entry := AuditLogEntry{
		Event:     event,
		UserID:    userID,
		Username:  username,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Action:    string(event),
		Result:    result,
		Details:   details,
	}
	a.LogEvent(entry)
}

// LogDocumentOperation logs document-related operations
func (a *AuditLogger) LogDocumentOperation(event AuditEvent, userID, username, documentID, ipAddress string, result string, details map[string]interface{}) {
	entry := AuditLogEntry{
		Event:      event,
		UserID:     userID,
		Username:   username,
		DocumentID: documentID,
		IPAddress:  ipAddress,
		Resource:   "document",
		Action:     string(event),
		Result:     result,
		Details:    details,
	}
	a.LogEvent(entry)
}

// LogVerificationAttempt logs document verification attempts
func (a *AuditLogger) LogVerificationAttempt(event AuditEvent, documentID, ipAddress, userAgent string, result string, details map[string]interface{}) {
	entry := AuditLogEntry{
		Event:      event,
		DocumentID: documentID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Resource:   "verification",
		Action:     string(event),
		Result:     result,
		Details:    details,
	}
	a.LogEvent(entry)
}

// LogSecurityEvent logs security-related events
func (a *AuditLogger) LogSecurityEvent(event AuditEvent, ipAddress, userAgent, message string, details map[string]interface{}) {
	entry := AuditLogEntry{
		Event:     event,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Resource:  "security",
		Action:    string(event),
		Result:    "BLOCKED",
		Message:   message,
		Details:   details,
		Severity:  "HIGH",
	}
	a.LogEvent(entry)
}

// Package-level convenience functions
func LogAuthentication(event AuditEvent, userID, username, ipAddress, userAgent string, result string, details map[string]interface{}) {
	GetAuditLogger().LogAuthentication(event, userID, username, ipAddress, userAgent, result, details)
}

func LogDocumentOperation(event AuditEvent, userID, username, documentID, ipAddress string, result string, details map[string]interface{}) {
	GetAuditLogger().LogDocumentOperation(event, userID, username, documentID, ipAddress, result, details)
}

func LogVerificationAttempt(event AuditEvent, documentID, ipAddress, userAgent string, result string, details map[string]interface{}) {
	GetAuditLogger().LogVerificationAttempt(event, documentID, ipAddress, userAgent, result, details)
}

func LogSecurityEvent(event AuditEvent, ipAddress, userAgent, message string, details map[string]interface{}) {
	GetAuditLogger().LogSecurityEvent(event, ipAddress, userAgent, message, details)
}