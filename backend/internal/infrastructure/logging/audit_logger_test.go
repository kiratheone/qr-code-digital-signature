package logging

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditLogger_LogEvent(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	logger := &AuditLogger{
		logger: log.New(&buf, "", 0),
	}

	// Test logging an event
	entry := AuditLogEntry{
		Event:     AuditEventLogin,
		UserID:    "user123",
		Username:  "testuser",
		IPAddress: "192.168.1.1",
		Action:    "LOGIN",
		Result:    "SUCCESS",
		Message:   "User logged in successfully",
	}

	logger.LogEvent(entry)

	// Parse the logged JSON
	var loggedEntry AuditLogEntry
	err := json.Unmarshal(buf.Bytes(), &loggedEntry)
	require.NoError(t, err)

	// Verify the logged entry
	assert.Equal(t, AuditEventLogin, loggedEntry.Event)
	assert.Equal(t, "user123", loggedEntry.UserID)
	assert.Equal(t, "testuser", loggedEntry.Username)
	assert.Equal(t, "192.168.1.1", loggedEntry.IPAddress)
	assert.Equal(t, "LOGIN", loggedEntry.Action)
	assert.Equal(t, "SUCCESS", loggedEntry.Result)
	assert.Equal(t, "User logged in successfully", loggedEntry.Message)
	assert.Equal(t, "MEDIUM", loggedEntry.Severity) // Default severity for login
	assert.False(t, loggedEntry.Timestamp.IsZero())
}

func TestAuditLogger_LogAuthentication(t *testing.T) {
	var buf bytes.Buffer
	logger := &AuditLogger{
		logger: log.New(&buf, "", 0),
	}

	details := map[string]interface{}{
		"attempt_count": float64(1), // JSON unmarshaling converts to float64
		"source":        "web",
	}

	logger.LogAuthentication(
		AuditEventLogin,
		"user123",
		"testuser",
		"192.168.1.1",
		"Mozilla/5.0",
		"SUCCESS",
		details,
	)

	var loggedEntry AuditLogEntry
	err := json.Unmarshal(buf.Bytes(), &loggedEntry)
	require.NoError(t, err)

	assert.Equal(t, AuditEventLogin, loggedEntry.Event)
	assert.Equal(t, "user123", loggedEntry.UserID)
	assert.Equal(t, "testuser", loggedEntry.Username)
	assert.Equal(t, "192.168.1.1", loggedEntry.IPAddress)
	assert.Equal(t, "Mozilla/5.0", loggedEntry.UserAgent)
	assert.Equal(t, "SUCCESS", loggedEntry.Result)
	assert.Equal(t, details, loggedEntry.Details)
}

func TestAuditLogger_LogDocumentOperation(t *testing.T) {
	var buf bytes.Buffer
	logger := &AuditLogger{
		logger: log.New(&buf, "", 0),
	}

	details := map[string]interface{}{
		"filename": "document.pdf",
		"size":     float64(1024), // JSON unmarshaling converts to float64
	}

	logger.LogDocumentOperation(
		AuditEventDocumentSign,
		"user123",
		"testuser",
		"doc456",
		"192.168.1.1",
		"SUCCESS",
		details,
	)

	var loggedEntry AuditLogEntry
	err := json.Unmarshal(buf.Bytes(), &loggedEntry)
	require.NoError(t, err)

	assert.Equal(t, AuditEventDocumentSign, loggedEntry.Event)
	assert.Equal(t, "user123", loggedEntry.UserID)
	assert.Equal(t, "testuser", loggedEntry.Username)
	assert.Equal(t, "doc456", loggedEntry.DocumentID)
	assert.Equal(t, "192.168.1.1", loggedEntry.IPAddress)
	assert.Equal(t, "document", loggedEntry.Resource)
	assert.Equal(t, "SUCCESS", loggedEntry.Result)
	assert.Equal(t, details, loggedEntry.Details)
}

func TestAuditLogger_LogVerificationAttempt(t *testing.T) {
	var buf bytes.Buffer
	logger := &AuditLogger{
		logger: log.New(&buf, "", 0),
	}

	details := map[string]interface{}{
		"verification_type": "hash_comparison",
		"hash_match":        true,
	}

	logger.LogVerificationAttempt(
		AuditEventVerificationSuccess,
		"doc456",
		"192.168.1.1",
		"Mozilla/5.0",
		"SUCCESS",
		details,
	)

	var loggedEntry AuditLogEntry
	err := json.Unmarshal(buf.Bytes(), &loggedEntry)
	require.NoError(t, err)

	assert.Equal(t, AuditEventVerificationSuccess, loggedEntry.Event)
	assert.Equal(t, "doc456", loggedEntry.DocumentID)
	assert.Equal(t, "192.168.1.1", loggedEntry.IPAddress)
	assert.Equal(t, "Mozilla/5.0", loggedEntry.UserAgent)
	assert.Equal(t, "verification", loggedEntry.Resource)
	assert.Equal(t, "SUCCESS", loggedEntry.Result)
	assert.Equal(t, details, loggedEntry.Details)
}

func TestAuditLogger_LogSecurityEvent(t *testing.T) {
	var buf bytes.Buffer
	logger := &AuditLogger{
		logger: log.New(&buf, "", 0),
	}

	details := map[string]interface{}{
		"attack_type": "sql_injection",
		"pattern":     "'; DROP TABLE",
	}

	logger.LogSecurityEvent(
		AuditEventSuspiciousActivity,
		"192.168.1.1",
		"Mozilla/5.0",
		"Suspicious SQL injection attempt detected",
		details,
	)

	var loggedEntry AuditLogEntry
	err := json.Unmarshal(buf.Bytes(), &loggedEntry)
	require.NoError(t, err)

	assert.Equal(t, AuditEventSuspiciousActivity, loggedEntry.Event)
	assert.Equal(t, "192.168.1.1", loggedEntry.IPAddress)
	assert.Equal(t, "Mozilla/5.0", loggedEntry.UserAgent)
	assert.Equal(t, "security", loggedEntry.Resource)
	assert.Equal(t, "BLOCKED", loggedEntry.Result)
	assert.Equal(t, "Suspicious SQL injection attempt detected", loggedEntry.Message)
	assert.Equal(t, "HIGH", loggedEntry.Severity)
	assert.Equal(t, details, loggedEntry.Details)
}

func TestAuditLogger_getSeverityForEvent(t *testing.T) {
	logger := &AuditLogger{}

	tests := []struct {
		event    AuditEvent
		expected string
	}{
		{AuditEventAuthFailure, "HIGH"},
		{AuditEventSuspiciousActivity, "HIGH"},
		{AuditEventRateLimitExceeded, "HIGH"},
		{AuditEventVerificationFailure, "MEDIUM"},
		{AuditEventValidationFailure, "MEDIUM"},
		{AuditEventLogin, "MEDIUM"},
		{AuditEventLogout, "MEDIUM"},
		{AuditEventDocumentSign, "MEDIUM"},
		{AuditEventDocumentDelete, "MEDIUM"},
		{AuditEventDocumentView, "LOW"},
		{AuditEventRegister, "LOW"},
	}

	for _, tt := range tests {
		t.Run(string(tt.event), func(t *testing.T) {
			severity := logger.getSeverityForEvent(tt.event)
			assert.Equal(t, tt.expected, severity)
		})
	}
}

func TestInitializeAuditLogger(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Initialize audit logger
	err := InitializeAuditLogger(tempDir)
	require.NoError(t, err)

	// Verify that the audit logger was initialized
	auditLogger := GetAuditLogger()
	assert.NotNil(t, auditLogger)

	// Test logging to verify file creation
	auditLogger.LogEvent(AuditLogEntry{
		Event:   AuditEventLogin,
		UserID:  "test",
		Result:  "SUCCESS",
		Message: "Test log entry",
	})

	// Verify that the audit log file was created
	auditLogPath := tempDir + "/audit.log"
	_, err = os.Stat(auditLogPath)
	assert.NoError(t, err, "Audit log file should be created")
}

func TestAuditPackageLevelFunctions(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Initialize audit logger
	err := InitializeAuditLogger(tempDir)
	require.NoError(t, err)

	// Test package-level functions (these should not panic)
	LogAuthentication(AuditEventLogin, "user123", "testuser", "192.168.1.1", "Mozilla/5.0", "SUCCESS", nil)
	LogDocumentOperation(AuditEventDocumentSign, "user123", "testuser", "doc456", "192.168.1.1", "SUCCESS", nil)
	LogVerificationAttempt(AuditEventVerificationSuccess, "doc456", "192.168.1.1", "Mozilla/5.0", "SUCCESS", nil)
	LogSecurityEvent(AuditEventSuspiciousActivity, "192.168.1.1", "Mozilla/5.0", "Test security event", nil)

	// If we reach here without panicking, the test passes
	assert.True(t, true)
}