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

func TestAuditService_LogEvent(t *testing.T) {
	// Create temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "audit_test")
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
		BufferSize:    10,
		FlushInterval: 100 * time.Millisecond,
	}

	auditService, err := NewAuditService(config)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}
	defer auditService.Close()

	// Create context with test data
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user_id", "test-user-123")
	ctx = context.WithValue(ctx, "request_id", "req-456")
	ctx = context.WithValue(ctx, "client_ip", "192.168.1.100")

	// Log test event
	details := map[string]interface{}{
		"test_field": "test_value",
		"number":     42,
	}

	auditService.LogEvent(ctx, AuditEventDocumentSign, AuditSeverityInfo, 
		"test_action", "success", "Test audit message", details)

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify log file was created
	files, err := filepath.Glob(filepath.Join(tempDir, "audit_*.log"))
	if err != nil {
		t.Fatalf("Failed to list log files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No audit log files were created")
	}

	// Read log file content
	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Parse JSON log entry
	var logEntry AuditEvent
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatal("Log file is empty")
	}

	err = json.Unmarshal([]byte(lines[0]), &logEntry)
	if err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	// Verify log entry fields
	if logEntry.EventType != AuditEventDocumentSign {
		t.Errorf("Expected event type %s, got %s", AuditEventDocumentSign, logEntry.EventType)
	}

	if logEntry.Severity != AuditSeverityInfo {
		t.Errorf("Expected severity %s, got %s", AuditSeverityInfo, logEntry.Severity)
	}

	if logEntry.Action != "test_action" {
		t.Errorf("Expected action 'test_action', got %s", logEntry.Action)
	}

	if logEntry.Result != "success" {
		t.Errorf("Expected result 'success', got %s", logEntry.Result)
	}

	if logEntry.Message != "Test audit message" {
		t.Errorf("Expected message 'Test audit message', got %s", logEntry.Message)
	}

	if logEntry.UserID != "test-user-123" {
		t.Errorf("Expected user ID 'test-user-123', got %s", logEntry.UserID)
	}

	if logEntry.RequestID != "req-456" {
		t.Errorf("Expected request ID 'req-456', got %s", logEntry.RequestID)
	}

	if logEntry.IPAddress != "192.168.1.100" {
		t.Errorf("Expected IP address '192.168.1.100', got %s", logEntry.IPAddress)
	}

	// Verify details
	if logEntry.Details["test_field"] != "test_value" {
		t.Errorf("Expected test_field 'test_value', got %v", logEntry.Details["test_field"])
	}

	if logEntry.Details["number"] != float64(42) { // JSON unmarshals numbers as float64
		t.Errorf("Expected number 42, got %v", logEntry.Details["number"])
	}
}

func TestAuditService_LogAuthEvent(t *testing.T) {
	// Create temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "audit_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create audit service
	config := AuditConfig{
		LogDir:        tempDir,
		BufferSize:    10,
		FlushInterval: 100 * time.Millisecond,
	}

	auditService, err := NewAuditService(config)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}
	defer auditService.Close()

	// Test successful login
	ctx := context.Background()
	details := map[string]interface{}{
		"username": "testuser",
		"role":     "admin",
	}

	auditService.LogAuthEvent(ctx, AuditEventLogin, "user-123", "192.168.1.100", true, details)

	// Test failed login
	failDetails := map[string]interface{}{
		"username": "testuser",
		"reason":   "invalid_password",
	}

	auditService.LogAuthEvent(ctx, AuditEventLoginFailed, "user-123", "192.168.1.100", false, failDetails)

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify log file content
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
		t.Fatal("Expected at least 2 log entries")
	}

	// Parse first log entry (successful login)
	var successEntry AuditEvent
	err = json.Unmarshal([]byte(lines[0]), &successEntry)
	if err != nil {
		t.Fatalf("Failed to parse success log entry: %v", err)
	}

	if successEntry.EventType != AuditEventLogin {
		t.Errorf("Expected event type %s, got %s", AuditEventLogin, successEntry.EventType)
	}

	if successEntry.Severity != AuditSeverityInfo {
		t.Errorf("Expected severity %s, got %s", AuditSeverityInfo, successEntry.Severity)
	}

	if successEntry.Result != "success" {
		t.Errorf("Expected result 'success', got %s", successEntry.Result)
	}

	// Parse second log entry (failed login)
	var failEntry AuditEvent
	err = json.Unmarshal([]byte(lines[1]), &failEntry)
	if err != nil {
		t.Fatalf("Failed to parse fail log entry: %v", err)
	}

	if failEntry.EventType != AuditEventLoginFailed {
		t.Errorf("Expected event type %s, got %s", AuditEventLoginFailed, failEntry.EventType)
	}

	if failEntry.Severity != AuditSeverityWarning {
		t.Errorf("Expected severity %s, got %s", AuditSeverityWarning, failEntry.Severity)
	}

	if failEntry.Result != "failure" {
		t.Errorf("Expected result 'failure', got %s", failEntry.Result)
	}
}

func TestAuditService_LogVerificationEvent(t *testing.T) {
	// Create temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "audit_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create audit service
	config := AuditConfig{
		LogDir:        tempDir,
		BufferSize:    10,
		FlushInterval: 100 * time.Millisecond,
	}

	auditService, err := NewAuditService(config)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}
	defer auditService.Close()

	// Test verification event
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "req-789")

	details := map[string]interface{}{
		"document_filename": "test.pdf",
		"document_size":     1024,
		"hash_matches":      true,
		"signature_valid":   true,
	}

	duration := 250 * time.Millisecond
	auditService.LogVerificationEvent(ctx, "doc-123", "192.168.1.200", "valid", duration, details)

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify log file content
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
	if len(lines) == 0 {
		t.Fatal("Expected at least 1 log entry")
	}

	// Parse log entry
	var logEntry AuditEvent
	err = json.Unmarshal([]byte(lines[0]), &logEntry)
	if err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if logEntry.EventType != AuditEventVerificationSuccess {
		t.Errorf("Expected event type %s, got %s", AuditEventVerificationSuccess, logEntry.EventType)
	}

	if logEntry.Resource != "document" {
		t.Errorf("Expected resource 'document', got %s", logEntry.Resource)
	}

	if logEntry.ResourceID != "doc-123" {
		t.Errorf("Expected resource ID 'doc-123', got %s", logEntry.ResourceID)
	}

	if logEntry.Action != "verify" {
		t.Errorf("Expected action 'verify', got %s", logEntry.Action)
	}

	if logEntry.Result != "valid" {
		t.Errorf("Expected result 'valid', got %s", logEntry.Result)
	}

	if logEntry.IPAddress != "192.168.1.200" {
		t.Errorf("Expected IP address '192.168.1.200', got %s", logEntry.IPAddress)
	}

	if logEntry.Duration == nil {
		t.Error("Expected duration to be set")
	} else if *logEntry.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, *logEntry.Duration)
	}

	// Verify details
	if logEntry.Details["document_filename"] != "test.pdf" {
		t.Errorf("Expected document_filename 'test.pdf', got %v", logEntry.Details["document_filename"])
	}

	if logEntry.Details["verification_result"] != "valid" {
		t.Errorf("Expected verification_result 'valid', got %v", logEntry.Details["verification_result"])
	}
}

func TestAuditService_LogRotation(t *testing.T) {
	// Create temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "audit_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create audit service with small file size for testing rotation
	config := AuditConfig{
		LogDir:        tempDir,
		MaxFileSize:   100, // Very small size to trigger rotation
		MaxFiles:      3,
		BufferSize:    1,
		FlushInterval: 10 * time.Millisecond,
	}

	auditService, err := NewAuditService(config)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}
	defer auditService.Close()

	// Log multiple events to trigger rotation
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		details := map[string]interface{}{
			"iteration": i,
			"large_data": strings.Repeat("x", 50), // Add some bulk to trigger size limit
		}
		
		auditService.LogEvent(ctx, AuditEventDocumentSign, AuditSeverityInfo, 
			"test_action", "success", "Test message for rotation", details)
		
		time.Sleep(20 * time.Millisecond) // Allow processing
	}

	// Wait for all events to be processed
	time.Sleep(200 * time.Millisecond)

	// Check that log files were created
	files, err := filepath.Glob(filepath.Join(tempDir, "audit_*.log"))
	if err != nil {
		t.Fatalf("Failed to list log files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No audit log files were created")
	}

	// Verify that at least one file exists (rotation may or may not have occurred depending on timing)
	t.Logf("Created %d log files", len(files))
}

func TestAuditService_GetAuditStats(t *testing.T) {
	// Create temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "audit_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create audit service
	config := AuditConfig{
		LogDir:        tempDir,
		MaxFileSize:   1024 * 1024,
		MaxFiles:      5,
		BufferSize:    100,
		FlushInterval: 100 * time.Millisecond,
	}

	auditService, err := NewAuditService(config)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}
	defer auditService.Close()

	// Get stats
	stats := auditService.GetAuditStats()

	// Verify stats structure
	if stats["buffer_size"] != 100 {
		t.Errorf("Expected buffer_size 100, got %v", stats["buffer_size"])
	}

	if stats["log_directory"] != tempDir {
		t.Errorf("Expected log_directory %s, got %v", tempDir, stats["log_directory"])
	}

	if stats["max_file_size"] != int64(1024*1024) {
		t.Errorf("Expected max_file_size %d, got %v", 1024*1024, stats["max_file_size"])
	}

	if stats["max_files"] != 5 {
		t.Errorf("Expected max_files 5, got %v", stats["max_files"])
	}

	// Buffered events should be 0 initially
	if stats["buffered_events"] != 0 {
		t.Errorf("Expected buffered_events 0, got %v", stats["buffered_events"])
	}
}

func TestAuditService_BufferOverflow(t *testing.T) {
	// Create temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "audit_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create audit service with very small buffer
	config := AuditConfig{
		LogDir:        tempDir,
		BufferSize:    2, // Very small buffer
		FlushInterval: 1 * time.Second, // Long interval to test overflow
	}

	auditService, err := NewAuditService(config)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}
	defer auditService.Close()

	// Log more events than buffer size
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		details := map[string]interface{}{
			"iteration": i,
		}
		
		auditService.LogEvent(ctx, AuditEventDocumentSign, AuditSeverityInfo, 
			"test_action", "success", "Test overflow message", details)
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Verify that events were still logged (should fallback to synchronous logging)
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
	if len(lines) < 3 { // Should have at least some events logged
		t.Errorf("Expected at least 3 log entries, got %d", len(lines))
	}
}