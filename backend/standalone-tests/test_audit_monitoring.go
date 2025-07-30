package main

import (
	"context"
	"digital-signature-system/internal/infrastructure/services"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	fmt.Println("Testing Audit and Monitoring Services...")

	// Test Audit Service
	testAuditService()

	// Test Monitoring Service
	testMonitoringService()

	fmt.Println("All tests completed successfully!")
}

func testAuditService() {
	fmt.Println("\n=== Testing Audit Service ===")

	// Create temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "audit_test")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp directory: %v", err))
	}
	defer os.RemoveAll(tempDir)

	// Create audit service
	config := services.AuditConfig{
		LogDir:        tempDir,
		MaxFileSize:   1024 * 1024, // 1MB
		MaxFiles:      5,
		RotationTime:  24 * time.Hour,
		BufferSize:    100,
		FlushInterval: 50 * time.Millisecond,
	}

	auditService, err := services.NewAuditService(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create audit service: %v", err))
	}
	defer auditService.Close()

	// Create context with test data
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user_id", "test-user-123")
	ctx = context.WithValue(ctx, "session_id", "session-456")
	ctx = context.WithValue(ctx, "request_id", "req-789")
	ctx = context.WithValue(ctx, "client_ip", "192.168.1.100")
	ctx = context.WithValue(ctx, "user_agent", "TestAgent/1.0")

	// Test different types of audit events
	fmt.Println("Logging authentication events...")
	auditService.LogAuthEvent(ctx, services.AuditEventLogin, "test-user-123", "192.168.1.100", true, map[string]interface{}{
		"username": "testuser",
		"role":     "admin",
	})

	auditService.LogAuthEvent(ctx, services.AuditEventLoginFailed, "test-user-456", "192.168.1.200", false, map[string]interface{}{
		"username": "baduser",
		"reason":   "invalid_password",
	})

	fmt.Println("Logging document events...")
	auditService.LogDocumentEvent(ctx, services.AuditEventDocumentSign, "doc-123", "test-user-123", map[string]interface{}{
		"filename": "test-document.pdf",
		"issuer":   "Test Organization",
	})

	fmt.Println("Logging verification events...")
	auditService.LogVerificationEvent(ctx, "doc-123", "192.168.1.300", "valid", 250*time.Millisecond, map[string]interface{}{
		"document_filename": "test-document.pdf",
		"hash_matches":      true,
		"signature_valid":   true,
	})

	fmt.Println("Logging security events...")
	auditService.LogSecurityEvent(ctx, services.AuditEventSecurityAlert, services.AuditSeverityCritical, "Test security alert", map[string]interface{}{
		"source_ip": "192.168.1.500",
		"threat_level": "high",
	})

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify log files were created
	files, err := filepath.Glob(filepath.Join(tempDir, "audit_*.log"))
	if err != nil {
		panic(fmt.Sprintf("Failed to list log files: %v", err))
	}

	if len(files) == 0 {
		panic("No audit log files were created")
	}

	fmt.Printf("Created %d audit log file(s)\n", len(files))

	// Read and verify log content
	content, err := os.ReadFile(files[0])
	if err != nil {
		panic(fmt.Sprintf("Failed to read log file: %v", err))
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	fmt.Printf("Logged %d audit events\n", len(lines))

	// Parse and display a sample log entry
	if len(lines) > 0 {
		var logEntry services.AuditEvent
		err := json.Unmarshal([]byte(lines[0]), &logEntry)
		if err == nil {
			fmt.Printf("Sample log entry: %s - %s - %s\n", logEntry.EventType, logEntry.Severity, logEntry.Message)
		}
	}

	// Test audit statistics
	stats := auditService.GetAuditStats()
	fmt.Printf("Audit stats: buffer_size=%v, buffered_events=%v\n", stats["buffer_size"], stats["buffered_events"])

	fmt.Println("✓ Audit Service test completed successfully")
}

func testMonitoringService() {
	fmt.Println("\n=== Testing Monitoring Service ===")

	// Create monitoring service
	config := services.MonitoringConfig{
		PerformanceInterval: 100 * time.Millisecond,
		AlertCheckInterval:  50 * time.Millisecond,
		RateLimitThreshold:  5, // Low threshold for testing
	}

	monitoringService := services.NewMonitoringService(config)
	defer monitoringService.Close()

	// Test metrics recording
	fmt.Println("Recording test metrics...")
	monitoringService.RecordMetric("test_counter", services.MetricTypeCounter, 10, map[string]string{
		"endpoint": "/api/test",
	})

	monitoringService.SetGauge("test_gauge", 42.5, map[string]string{
		"service": "test",
	})

	monitoringService.RecordTiming("test_timing", 150*time.Millisecond, map[string]string{
		"operation": "test",
	})

	// Test request tracking
	fmt.Println("Tracking test requests...")
	monitoringService.TrackRequest("GET", "/api/documents", 200, 100*time.Millisecond)
	monitoringService.TrackRequest("POST", "/api/documents", 400, 50*time.Millisecond)

	// Test security event tracking
	fmt.Println("Testing security event tracking...")
	ctx := context.Background()
	ctx = context.WithValue(ctx, "client_ip", "192.168.1.100")

	// Trigger rate limit violations
	for i := 0; i < 7; i++ {
		monitoringService.TrackRateLimitViolation(ctx, "192.168.1.100")
	}

	// Trigger auth failures
	for i := 0; i < 6; i++ {
		monitoringService.TrackAuthFailure(ctx, "user-123", "192.168.1.200", "invalid_password")
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Get and display metrics
	metrics := monitoringService.GetMetrics()
	fmt.Printf("Recorded %d metrics\n", len(metrics))

	for name, metric := range metrics {
		fmt.Printf("  %s (%s): %.2f\n", name, metric.Type, metric.Value)
	}

	// Get and display alerts
	alerts := monitoringService.GetAlerts(false) // Unresolved alerts
	fmt.Printf("Generated %d security alerts\n", len(alerts))

	for _, alert := range alerts {
		fmt.Printf("  Alert: %s - %s - %s\n", alert.Type, alert.Severity, alert.Message)
	}

	// Test performance metrics
	time.Sleep(150 * time.Millisecond) // Wait for performance collection
	perfMetrics := monitoringService.GetPerformanceMetrics()
	if perfMetrics != nil {
		fmt.Printf("Performance: Goroutines=%d, HeapMB=%.2f\n", perfMetrics.GoroutineCount, perfMetrics.HeapAllocMB)
	}

	// Test health status
	health := monitoringService.GetHealthStatus()
	fmt.Printf("Health status: %s\n", health["status"])

	// Test alert resolution
	if len(alerts) > 0 {
		err := monitoringService.ResolveAlert(alerts[0].ID, "test-admin")
		if err != nil {
			fmt.Printf("Failed to resolve alert: %v\n", err)
		} else {
			fmt.Println("Successfully resolved test alert")
		}
	}

	fmt.Println("✓ Monitoring Service test completed successfully")
}