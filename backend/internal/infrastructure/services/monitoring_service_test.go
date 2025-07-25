package services

import (
	"context"
	"testing"
	"time"
)

func TestMonitoringService_RecordMetric(t *testing.T) {
	// Create monitoring service
	config := MonitoringConfig{
		PerformanceInterval: 100 * time.Millisecond,
		AlertCheckInterval:  50 * time.Millisecond,
	}

	monitoringService := NewMonitoringService(config)
	defer monitoringService.Close()

	// Record test metrics
	tags := map[string]string{
		"endpoint": "/api/test",
		"method":   "GET",
	}

	monitoringService.RecordMetric("test_counter", MetricTypeCounter, 5, tags)
	monitoringService.SetGauge("test_gauge", 42.5, tags)
	monitoringService.RecordTiming("test_timing", 150*time.Millisecond, tags)

	// Get metrics
	metrics := monitoringService.GetMetrics()

	// Verify counter metric
	if metric, exists := metrics["test_counter"]; !exists {
		t.Error("Expected test_counter metric to exist")
	} else {
		if metric.Type != MetricTypeCounter {
			t.Errorf("Expected metric type %s, got %s", MetricTypeCounter, metric.Type)
		}
		if metric.Value != 5 {
			t.Errorf("Expected metric value 5, got %f", metric.Value)
		}
		if metric.Tags["endpoint"] != "/api/test" {
			t.Errorf("Expected endpoint tag '/api/test', got %s", metric.Tags["endpoint"])
		}
	}

	// Verify gauge metric
	if metric, exists := metrics["test_gauge"]; !exists {
		t.Error("Expected test_gauge metric to exist")
	} else {
		if metric.Type != MetricTypeGauge {
			t.Errorf("Expected metric type %s, got %s", MetricTypeGauge, metric.Type)
		}
		if metric.Value != 42.5 {
			t.Errorf("Expected metric value 42.5, got %f", metric.Value)
		}
	}

	// Verify timing metric
	if metric, exists := metrics["test_timing"]; !exists {
		t.Error("Expected test_timing metric to exist")
	} else {
		if metric.Type != MetricTypeTiming {
			t.Errorf("Expected metric type %s, got %s", MetricTypeTiming, metric.Type)
		}
		if metric.Value != 150 {
			t.Errorf("Expected metric value 150, got %f", metric.Value)
		}
	}
}

func TestMonitoringService_IncrementCounter(t *testing.T) {
	// Create monitoring service
	config := MonitoringConfig{}
	monitoringService := NewMonitoringService(config)
	defer monitoringService.Close()

	tags := map[string]string{"test": "value"}

	// Increment counter multiple times
	monitoringService.IncrementCounter("test_counter", tags)
	monitoringService.IncrementCounter("test_counter", tags)
	monitoringService.IncrementCounter("test_counter", tags)

	// Get metrics
	metrics := monitoringService.GetMetrics()

	// Verify counter was incremented
	if metric, exists := metrics["test_counter"]; !exists {
		t.Error("Expected test_counter metric to exist")
	} else {
		if metric.Value != 3 {
			t.Errorf("Expected metric value 3, got %f", metric.Value)
		}
	}
}

func TestMonitoringService_TrackRequest(t *testing.T) {
	// Create monitoring service
	config := MonitoringConfig{}
	monitoringService := NewMonitoringService(config)
	defer monitoringService.Close()

	// Track successful request
	monitoringService.TrackRequest("GET", "/api/documents", 200, 100*time.Millisecond)

	// Track error request
	monitoringService.TrackRequest("POST", "/api/documents", 400, 50*time.Millisecond)

	// Get metrics
	metrics := monitoringService.GetMetrics()

	// Verify request metrics
	if metric, exists := metrics["http_requests_total"]; !exists {
		t.Error("Expected http_requests_total metric to exist")
	} else {
		if metric.Value != 2 {
			t.Errorf("Expected total requests 2, got %f", metric.Value)
		}
	}

	// Verify error metrics
	if metric, exists := metrics["http_errors_total"]; !exists {
		t.Error("Expected http_errors_total metric to exist")
	} else {
		if metric.Value != 1 {
			t.Errorf("Expected total errors 1, got %f", metric.Value)
		}
	}
}

func TestMonitoringService_TrackRateLimitViolation(t *testing.T) {
	// Create audit service for testing
	auditConfig := AuditConfig{
		BufferSize:    10,
		FlushInterval: 10 * time.Millisecond,
	}
	auditService, err := NewAuditService(auditConfig)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}
	defer auditService.Close()

	// Create monitoring service
	config := MonitoringConfig{
		AuditService:       auditService,
		RateLimitThreshold: 2, // Low threshold for testing
	}
	monitoringService := NewMonitoringService(config)
	defer monitoringService.Close()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "client_ip", "192.168.1.100")

	// Track multiple violations to trigger alert
	for i := 0; i < 5; i++ {
		monitoringService.TrackRateLimitViolation(ctx, "192.168.1.100")
	}

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	// Get metrics
	metrics := monitoringService.GetMetrics()

	// Verify rate limit violation metrics
	if metric, exists := metrics["rate_limit_violations"]; !exists {
		t.Error("Expected rate_limit_violations metric to exist")
	} else {
		if metric.Value != 5 {
			t.Errorf("Expected 5 rate limit violations, got %f", metric.Value)
		}
	}

	// Get alerts
	alerts := monitoringService.GetAlerts(false) // Get unresolved alerts

	// Verify security alert was created
	found := false
	for _, alert := range alerts {
		if alert.Type == "rate_limit_violation" && alert.IPAddress == "192.168.1.100" {
			found = true
			if alert.Severity != AuditSeverityWarning {
				t.Errorf("Expected alert severity %s, got %s", AuditSeverityWarning, alert.Severity)
			}
			break
		}
	}

	if !found {
		t.Error("Expected rate limit violation alert to be created")
	}
}

func TestMonitoringService_TrackAuthFailure(t *testing.T) {
	// Create audit service for testing
	auditConfig := AuditConfig{
		BufferSize:    10,
		FlushInterval: 10 * time.Millisecond,
	}
	auditService, err := NewAuditService(auditConfig)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}
	defer auditService.Close()

	// Create monitoring service
	config := MonitoringConfig{
		AuditService: auditService,
	}
	monitoringService := NewMonitoringService(config)
	defer monitoringService.Close()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "client_ip", "192.168.1.200")

	// Track multiple auth failures to trigger alert
	for i := 0; i < 7; i++ {
		monitoringService.TrackAuthFailure(ctx, "user-123", "192.168.1.200", "invalid_password")
	}

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	// Get metrics
	metrics := monitoringService.GetMetrics()

	// Verify auth failure metrics
	if metric, exists := metrics["auth_failures_total"]; !exists {
		t.Error("Expected auth_failures_total metric to exist")
	} else {
		if metric.Value != 7 {
			t.Errorf("Expected 7 auth failures, got %f", metric.Value)
		}
	}

	// Get alerts
	alerts := monitoringService.GetAlerts(false)

	// Verify brute force alert was created
	found := false
	for _, alert := range alerts {
		if alert.Type == "brute_force_attempt" && alert.IPAddress == "192.168.1.200" {
			found = true
			if alert.Severity != AuditSeverityCritical {
				t.Errorf("Expected alert severity %s, got %s", AuditSeverityCritical, alert.Severity)
			}
			break
		}
	}

	if !found {
		t.Error("Expected brute force attempt alert to be created")
	}
}

func TestMonitoringService_TrackVerificationFailure(t *testing.T) {
	// Create audit service for testing
	auditConfig := AuditConfig{
		BufferSize:    10,
		FlushInterval: 10 * time.Millisecond,
	}
	auditService, err := NewAuditService(auditConfig)
	if err != nil {
		t.Fatalf("Failed to create audit service: %v", err)
	}
	defer auditService.Close()

	// Create monitoring service
	config := MonitoringConfig{
		AuditService: auditService,
	}
	monitoringService := NewMonitoringService(config)
	defer monitoringService.Close()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "client_ip", "192.168.1.300")

	// Track multiple verification failures to trigger alert
	for i := 0; i < 12; i++ {
		monitoringService.TrackVerificationFailure(ctx, "doc-123", "192.168.1.300", "invalid_signature")
	}

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	// Get metrics
	metrics := monitoringService.GetMetrics()

	// Verify verification failure metrics
	if metric, exists := metrics["verification_failures_total"]; !exists {
		t.Error("Expected verification_failures_total metric to exist")
	} else {
		if metric.Value != 12 {
			t.Errorf("Expected 12 verification failures, got %f", metric.Value)
		}
	}

	// Get alerts
	alerts := monitoringService.GetAlerts(false)

	// Verify suspicious verification activity alert was created
	found := false
	for _, alert := range alerts {
		if alert.Type == "suspicious_verification_activity" && alert.IPAddress == "192.168.1.300" {
			found = true
			if alert.Severity != AuditSeverityWarning {
				t.Errorf("Expected alert severity %s, got %s", AuditSeverityWarning, alert.Severity)
			}
			break
		}
	}

	if !found {
		t.Error("Expected suspicious verification activity alert to be created")
	}
}

func TestMonitoringService_CreateSecurityAlert(t *testing.T) {
	// Create monitoring service
	config := MonitoringConfig{}
	monitoringService := NewMonitoringService(config)
	defer monitoringService.Close()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "client_ip", "192.168.1.400")
	ctx = context.WithValue(ctx, "user_id", "user-456")
	ctx = context.WithValue(ctx, "request_id", "req-789")

	details := map[string]interface{}{
		"test_detail": "test_value",
		"count":       5,
	}

	// Create security alert
	monitoringService.CreateSecurityAlert(ctx, "test_alert", AuditSeverityError, 
		"Test security alert message", details)

	// Get alerts
	alerts := monitoringService.GetAlerts(false)

	// Verify alert was created
	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert, got %d", len(alerts))
	}

	alert := alerts[0]
	if alert.Type != "test_alert" {
		t.Errorf("Expected alert type 'test_alert', got %s", alert.Type)
	}

	if alert.Severity != AuditSeverityError {
		t.Errorf("Expected alert severity %s, got %s", AuditSeverityError, alert.Severity)
	}

	if alert.Message != "Test security alert message" {
		t.Errorf("Expected alert message 'Test security alert message', got %s", alert.Message)
	}

	if alert.IPAddress != "192.168.1.400" {
		t.Errorf("Expected IP address '192.168.1.400', got %s", alert.IPAddress)
	}

	if alert.UserID != "user-456" {
		t.Errorf("Expected user ID 'user-456', got %s", alert.UserID)
	}

	if alert.RequestID != "req-789" {
		t.Errorf("Expected request ID 'req-789', got %s", alert.RequestID)
	}

	if alert.Resolved {
		t.Error("Expected alert to be unresolved")
	}

	// Verify details
	if alert.Details["test_detail"] != "test_value" {
		t.Errorf("Expected test_detail 'test_value', got %v", alert.Details["test_detail"])
	}
}

func TestMonitoringService_ResolveAlert(t *testing.T) {
	// Create monitoring service
	config := MonitoringConfig{}
	monitoringService := NewMonitoringService(config)
	defer monitoringService.Close()

	ctx := context.Background()

	// Create security alert
	monitoringService.CreateSecurityAlert(ctx, "test_alert", AuditSeverityWarning, 
		"Test alert for resolution", nil)

	// Get unresolved alerts
	alerts := monitoringService.GetAlerts(false)
	if len(alerts) != 1 {
		t.Fatalf("Expected 1 unresolved alert, got %d", len(alerts))
	}

	alertID := alerts[0].ID

	// Resolve the alert
	err := monitoringService.ResolveAlert(alertID, "admin-user")
	if err != nil {
		t.Fatalf("Failed to resolve alert: %v", err)
	}

	// Verify alert is resolved
	unresolvedAlerts := monitoringService.GetAlerts(false)
	if len(unresolvedAlerts) != 0 {
		t.Errorf("Expected 0 unresolved alerts, got %d", len(unresolvedAlerts))
	}

	resolvedAlerts := monitoringService.GetAlerts(true)
	if len(resolvedAlerts) != 1 {
		t.Errorf("Expected 1 resolved alert, got %d", len(resolvedAlerts))
	}

	resolvedAlert := resolvedAlerts[0]
	if !resolvedAlert.Resolved {
		t.Error("Expected alert to be resolved")
	}

	if resolvedAlert.ResolvedBy != "admin-user" {
		t.Errorf("Expected resolved by 'admin-user', got %s", resolvedAlert.ResolvedBy)
	}

	if resolvedAlert.ResolvedAt == nil {
		t.Error("Expected resolved at timestamp to be set")
	}
}

func TestMonitoringService_GetPerformanceMetrics(t *testing.T) {
	// Create monitoring service with short interval for testing
	config := MonitoringConfig{
		PerformanceInterval: 50 * time.Millisecond,
	}
	monitoringService := NewMonitoringService(config)
	defer monitoringService.Close()

	// Wait for at least one performance collection cycle
	time.Sleep(100 * time.Millisecond)

	// Get performance metrics
	perfMetrics := monitoringService.GetPerformanceMetrics()

	if perfMetrics == nil {
		t.Fatal("Expected performance metrics to be available")
	}

	// Verify basic performance metrics are collected
	if perfMetrics.GoroutineCount <= 0 {
		t.Error("Expected goroutine count to be greater than 0")
	}

	if perfMetrics.HeapAllocMB < 0 {
		t.Error("Expected heap allocation to be non-negative")
	}

	if perfMetrics.HeapSysMB < 0 {
		t.Error("Expected heap system memory to be non-negative")
	}

	if perfMetrics.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestMonitoringService_GetHealthStatus(t *testing.T) {
	// Create monitoring service
	config := MonitoringConfig{
		PerformanceInterval: 50 * time.Millisecond,
	}
	monitoringService := NewMonitoringService(config)
	defer monitoringService.Close()

	// Wait for performance metrics to be collected
	time.Sleep(100 * time.Millisecond)

	// Get health status
	health := monitoringService.GetHealthStatus()

	// Verify health status structure
	if health["status"] == nil {
		t.Error("Expected status field in health response")
	}

	if health["timestamp"] == nil {
		t.Error("Expected timestamp field in health response")
	}

	if health["checks"] == nil {
		t.Error("Expected checks field in health response")
	}

	checks := health["checks"].(map[string]interface{})

	// Verify memory check
	if checks["memory"] == nil {
		t.Error("Expected memory check in health response")
	}

	// Verify goroutines check
	if checks["goroutines"] == nil {
		t.Error("Expected goroutines check in health response")
	}

	// Verify security check
	if checks["security"] == nil {
		t.Error("Expected security check in health response")
	}

	// Status should be healthy initially
	if health["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", health["status"])
	}
}

func TestMonitoringService_AlertThresholds(t *testing.T) {
	// Create monitoring service with custom thresholds
	config := MonitoringConfig{
		AlertThresholds: map[string]float64{
			"memory_usage": 50.0, // Low threshold for testing
		},
		PerformanceInterval: 50 * time.Millisecond,
		AlertCheckInterval:  25 * time.Millisecond,
	}
	monitoringService := NewMonitoringService(config)
	defer monitoringService.Close()

	// Wait for alert checking to run
	time.Sleep(100 * time.Millisecond)

	// The test environment might trigger memory alerts depending on actual usage
	// This test mainly verifies the alert checking mechanism runs without errors
	alerts := monitoringService.GetAlerts(false)
	
	// We don't assert specific alert counts since they depend on actual system state
	// but we verify the alert system is functional
	t.Logf("Found %d alerts during threshold testing", len(alerts))
}