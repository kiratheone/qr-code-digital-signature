package services

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAuditService_Integration(t *testing.T) {
	// Create temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "audit_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create audit service
	config := AuditConfig{
		LogDir:        tempDir,
		MaxFileSize:   1024 * 1024, // 1MB
		MaxFiles:      5,
		RotationTime:  24 * time.Hour,
		BufferSize:    100,
		FlushInterval: 50 * time.Millisecond,
	}

	auditService, err := NewAuditService(config)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}
	defer auditService.Close()

	// Test comprehensive audit logging scenario
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user_id", "test-user-123")
	ctx = context.WithValue(ctx, "session_id", "session-456")
	ctx = context.WithValue(ctx, "request_id", "req-789")
	ctx = context.WithValue(ctx, "client_ip", "192.168.1.100")
	ctx = context.WithValue(ctx, "user_agent", "TestAgent/1.0")

	// Test authentication events
	auditService.LogAuthEvent(ctx, AuditEventLogin, "test-user-123", "192.168.1.100", true, map[string]interface{}{
		"username": "testuser",
		"role":     "admin",
	})

	auditService.LogAuthEvent(ctx, AuditEventLoginFailed, "test-user-456", "192.168.1.200", false, map[string]interface{}{
		"username": "baduser",
		"reason":   "invalid_password",
		"attempts": 3,
	})

	// Test document events
	auditService.LogDocumentEvent(ctx, AuditEventDocumentSign, "doc-123", "test-user-123", map[string]interface{}{
		"filename":     "test-document.pdf",
		"file_size":    1024000,
		"issuer":       "Test Organization",
		"signature_id": "sig-789",
	})

	auditService.LogDocumentEvent(ctx, AuditEventDocumentDelete, "doc-456", "test-user-123", map[string]interface{}{
		"filename": "old-document.pdf",
		"reason":   "user_requested",
	})

	// Test verification events
	auditService.LogVerificationEvent(ctx, "doc-123", "192.168.1.300", "valid", 250*time.Millisecond, map[string]interface{}{
		"document_filename": "test-document.pdf",
		"hash_matches":      true,
		"signature_valid":   true,
		"verifier_country":  "US",
	})

	auditService.LogVerificationEvent(ctx, "doc-789", "192.168.1.400", "invalid", 150*time.Millisecond, map[string]interface{}{
		"document_filename": "suspicious-document.pdf",
		"hash_matches":      false,
		"signature_valid":   false,
		"suspicious_activity": true,
	})

	// Test security events
	auditService.LogSecurityEvent(ctx, AuditEventSecurityAlert, AuditSeverityCritical, "Multiple failed login attempts detected", map[string]interface{}{
		"source_ip":      "192.168.1.500",
		"failed_attempts": 10,
		"time_window":    "5 minutes",
		"blocked":        true,
	})

	// Test system events
	auditService.LogSystemEvent(ctx, AuditEventKeyRotation, "Cryptographic keys rotated successfully", map[string]interface{}{
		"old_key_id": "key-123",
		"new_key_id": "key-456",
		"algorithm":  "RSA-2048",
	})

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify log files were created
	files, err := filepath.Glob(filepath.Join(tempDir, "audit_*.log"))
	if err != nil {
		t.Fatalf("Failed to list log files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No audit log files were created")
	}

	// Read and verify log file content
	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 7 { // We logged 7 events
		t.Fatalf("Expected at least 7 log entries, got %d", len(lines))
	}

	// Verify each type of event was logged correctly
	eventTypes := make(map[AuditEventType]int)
	for _, line := range lines {
		if line == "" {
			continue
		}

		var logEntry AuditEvent
		err := json.Unmarshal([]byte(line), &logEntry)
		if err != nil {
			t.Fatalf("Failed to parse log entry: %v", err)
		}

		eventTypes[logEntry.EventType]++

		// Verify common fields
		if logEntry.ID == "" {
			t.Error("Log entry missing ID")
		}
		if logEntry.Timestamp.IsZero() {
			t.Error("Log entry missing timestamp")
		}
		if logEntry.EventType == "" {
			t.Error("Log entry missing event type")
		}
		if logEntry.Severity == "" {
			t.Error("Log entry missing severity")
		}

		// Verify context fields were captured
		switch logEntry.EventType {
		case AuditEventLogin, AuditEventLoginFailed:
			if logEntry.UserID == "" {
				t.Error("Auth event missing user ID")
			}
			if logEntry.IPAddress == "" {
				t.Error("Auth event missing IP address")
			}
		case AuditEventDocumentSign, AuditEventDocumentDelete:
			if logEntry.Details["document_id"] == nil {
				t.Error("Document event missing document ID in details")
			}
		case AuditEventVerificationSuccess, AuditEventVerificationFailed:
			if logEntry.ResourceID == "" {
				t.Error("Verification event missing resource ID")
			}
			if logEntry.Duration == nil {
				t.Error("Verification event missing duration")
			}
		}
	}

	// Verify we have the expected event types
	expectedEvents := []AuditEventType{
		AuditEventLogin,
		AuditEventLoginFailed,
		AuditEventDocumentSign,
		AuditEventDocumentDelete,
		AuditEventVerificationSuccess,
		AuditEventVerificationFailed,
		AuditEventSecurityAlert,
		AuditEventKeyRotation,
	}

	for _, expectedEvent := range expectedEvents {
		if eventTypes[expectedEvent] == 0 {
			t.Errorf("Expected event type %s was not found in logs", expectedEvent)
		}
	}

	// Test audit statistics
	stats := auditService.GetAuditStats()
	if stats["log_directory"] != tempDir {
		t.Errorf("Expected log directory %s, got %v", tempDir, stats["log_directory"])
	}
	if stats["buffer_size"] != 100 {
		t.Errorf("Expected buffer size 100, got %v", stats["buffer_size"])
	}
}

func TestAuditService_LogRotation_Integration(t *testing.T) {
	// Create temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "audit_rotation_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create audit service with small file size to trigger rotation
	config := AuditConfig{
		LogDir:        tempDir,
		MaxFileSize:   500, // Very small size to trigger rotation
		MaxFiles:      3,
		RotationTime:  24 * time.Hour,
		BufferSize:    10,
		FlushInterval: 10 * time.Millisecond,
	}

	auditService, err := NewAuditService(config)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}
	defer auditService.Close()

	// Log many events to trigger rotation
	ctx := context.Background()
	for i := 0; i < 20; i++ {
		details := map[string]interface{}{
			"iteration":   i,
			"large_data":  strings.Repeat("x", 100), // Add bulk to trigger size limit
			"timestamp":   time.Now(),
			"test_field":  "test_value_with_some_length",
		}

		auditService.LogEvent(ctx, AuditEventDocumentSign, AuditSeverityInfo,
			"test_action", "success", "Test message for rotation testing", details)

		// Small delay to allow processing
		time.Sleep(20 * time.Millisecond)
	}

	// Wait for all events to be processed
	time.Sleep(300 * time.Millisecond)

	// Check that log files were created
	files, err := filepath.Glob(filepath.Join(tempDir, "audit_*.log"))
	if err != nil {
		t.Fatalf("Failed to list log files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No audit log files were created")
	}

	t.Logf("Created %d log files during rotation test", len(files))

	// Verify that events were logged across files
	totalEvents := 0
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Failed to read log file %s: %v", file, err)
		}

		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		for _, line := range lines {
			if line != "" {
				totalEvents++
			}
		}
	}

	if totalEvents < 15 { // Should have most of the 20 events
		t.Errorf("Expected at least 15 events across all files, got %d", totalEvents)
	}
}

func TestAuditService_PerformanceUnderLoad(t *testing.T) {
	// Create temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "audit_performance_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create audit service with larger buffer for performance testing
	config := AuditConfig{
		LogDir:        tempDir,
		MaxFileSize:   10 * 1024 * 1024, // 10MB
		MaxFiles:      5,
		RotationTime:  24 * time.Hour,
		BufferSize:    1000, // Large buffer
		FlushInterval: 100 * time.Millisecond,
	}

	auditService, err := NewAuditService(config)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}
	defer auditService.Close()

	// Performance test: log many events quickly
	start := time.Now()
	numEvents := 1000

	ctx := context.Background()
	ctx = context.WithValue(ctx, "user_id", "perf-test-user")
	ctx = context.WithValue(ctx, "client_ip", "192.168.1.100")

	for i := 0; i < numEvents; i++ {
		details := map[string]interface{}{
			"iteration":    i,
			"batch_id":     i / 100,
			"test_data":    "performance_test_data",
			"timestamp":    time.Now(),
		}

		auditService.LogEvent(ctx, AuditEventDocumentSign, AuditSeverityInfo,
			"performance_test", "success", "Performance test event", details)
	}

	logDuration := time.Since(start)

	// Wait for all events to be processed
	time.Sleep(500 * time.Millisecond)

	processingDuration := time.Since(start)

	// Verify performance
	eventsPerSecond := float64(numEvents) / logDuration.Seconds()
	t.Logf("Logged %d events in %v (%.2f events/sec)", numEvents, logDuration, eventsPerSecond)
	t.Logf("Total processing time: %v", processingDuration)

	// Should be able to log at least 1000 events per second
	if eventsPerSecond < 1000 {
		t.Errorf("Performance too slow: %.2f events/sec, expected at least 1000", eventsPerSecond)
	}

	// Verify all events were logged
	files, err := filepath.Glob(filepath.Join(tempDir, "audit_*.log"))
	if err != nil {
		t.Fatalf("Failed to list log files: %v", err)
	}

	totalLoggedEvents := 0
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		for _, line := range lines {
			if line != "" {
				totalLoggedEvents++
			}
		}
	}

	if totalLoggedEvents < numEvents-10 { // Allow for small loss due to timing
		t.Errorf("Expected at least %d events logged, got %d", numEvents-10, totalLoggedEvents)
	}

	// Test audit statistics
	stats := auditService.GetAuditStats()
	t.Logf("Audit stats: %+v", stats)
}

func TestAuditService_ErrorHandling(t *testing.T) {
	// Test with invalid directory
	config := AuditConfig{
		LogDir:        "/invalid/path/that/does/not/exist",
		MaxFileSize:   1024 * 1024,
		MaxFiles:      5,
		BufferSize:    10,
		FlushInterval: 100 * time.Millisecond,
	}

	_, err := NewAuditService(config)
	if err == nil {
		t.Error("Expected error when creating audit service with invalid directory")
	}

	// Test with valid directory but restricted permissions
	tempDir, err := os.MkdirTemp("", "audit_error_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create audit service
	config.LogDir = tempDir
	auditService, err := NewAuditService(config)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}
	defer auditService.Close()

	// Test logging with nil context
	auditService.LogEvent(nil, AuditEventDocumentSign, AuditSeverityInfo,
		"test_action", "success", "Test with nil context", nil)

	// Test logging with empty details
	ctx := context.Background()
	auditService.LogEvent(ctx, AuditEventDocumentSign, AuditSeverityInfo,
		"test_action", "success", "Test with nil details", nil)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify events were still logged
	files, err := filepath.Glob(filepath.Join(tempDir, "audit_*.log"))
	if err != nil {
		t.Fatalf("Failed to list log files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No audit log files were created")
	}

	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) < 2 {
		t.Errorf("Expected at least 2 log entries, got %d", len(lines))
	}
}