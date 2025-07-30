package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Standalone audit service implementation for testing
type AuditSeverity string

const (
	AuditSeverityInfo     AuditSeverity = "info"
	AuditSeverityWarning  AuditSeverity = "warning"
	AuditSeverityError    AuditSeverity = "error"
	AuditSeverityCritical AuditSeverity = "critical"
)

type AuditEventType string

const (
	AuditEventLogin               AuditEventType = "auth.login"
	AuditEventLoginFailed         AuditEventType = "auth.login_failed"
	AuditEventDocumentSign        AuditEventType = "document.sign"
	AuditEventVerificationSuccess AuditEventType = "verification.success"
	AuditEventSecurityAlert       AuditEventType = "system.security_alert"
)

type AuditEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   AuditEventType         `json:"event_type"`
	Severity    AuditSeverity          `json:"severity"`
	UserID      string                 `json:"user_id,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	Action      string                 `json:"action"`
	Result      string                 `json:"result"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Duration    *time.Duration         `json:"duration,omitempty"`
}

type AuditConfig struct {
	LogDir        string
	MaxFileSize   int64
	MaxFiles      int
	BufferSize    int
	FlushInterval time.Duration
}

type AuditService struct {
	writer      io.Writer
	logFile     *os.File
	logDir      string
	maxFileSize int64
	maxFiles    int
	mu          sync.RWMutex
	buffer      chan *AuditEvent
	bufferSize  int
	flushTicker *time.Ticker
	stopChan    chan struct{}
}

func NewAuditService(config AuditConfig) (*AuditService, error) {
	if config.LogDir == "" {
		config.LogDir = "./logs/audit"
	}
	if config.MaxFileSize == 0 {
		config.MaxFileSize = 100 * 1024 * 1024 // 100MB
	}
	if config.MaxFiles == 0 {
		config.MaxFiles = 30
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
		logDir:      config.LogDir,
		maxFileSize: config.MaxFileSize,
		maxFiles:    config.MaxFiles,
		buffer:      make(chan *AuditEvent, config.BufferSize),
		bufferSize:  config.BufferSize,
		flushTicker: time.NewTicker(config.FlushInterval),
		stopChan:    make(chan struct{}),
	}

	// Initialize log file
	if err := service.initLogFile(); err != nil {
		return nil, fmt.Errorf("failed to initialize audit log file: %w", err)
	}

	// Start background processing
	go service.processEvents()

	return service, nil
}

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
		if ip := ctx.Value("client_ip"); ip != nil {
			event.IPAddress = fmt.Sprintf("%v", ip)
		}
	}

	// Try to send to buffer
	select {
	case as.buffer <- event:
		// Successfully buffered
	default:
		// Buffer is full, log synchronously
		as.writeEvent(event)
	}
}

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

func (as *AuditService) writeEvent(event *AuditEvent) {
	as.mu.Lock()
	defer as.mu.Unlock()

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

func (as *AuditService) flush() {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.logFile != nil {
		as.logFile.Sync()
	}
}

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

	return stats
}

// Monitoring Service
type MetricType string

const (
	MetricTypeCounter MetricType = "counter"
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeTiming  MetricType = "timing"
)

type Metric struct {
	Name      string            `json:"name"`
	Type      MetricType        `json:"type"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
	Tags      map[string]string `json:"tags,omitempty"`
}

type SecurityAlert struct {
	ID        string        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Severity  AuditSeverity `json:"severity"`
	Type      string        `json:"type"`
	Message   string        `json:"message"`
	IPAddress string        `json:"ip_address,omitempty"`
	Resolved  bool          `json:"resolved"`
}

type PerformanceMetrics struct {
	Timestamp      time.Time `json:"timestamp"`
	GoroutineCount int       `json:"goroutine_count"`
	HeapAllocMB    float64   `json:"heap_alloc_mb"`
	HeapSysMB      float64   `json:"heap_sys_mb"`
}

type MonitoringService struct {
	metrics     map[string]*Metric
	alerts      []*SecurityAlert
	perfMetrics *PerformanceMetrics
	mu          sync.RWMutex
	perfTicker  *time.Ticker
	stopChan    chan struct{}
}

func NewMonitoringService() *MonitoringService {
	service := &MonitoringService{
		metrics:    make(map[string]*Metric),
		alerts:     make([]*SecurityAlert, 0),
		perfTicker: time.NewTicker(30 * time.Second),
		stopChan:   make(chan struct{}),
	}

	go service.monitorPerformance()
	return service
}

func (ms *MonitoringService) RecordMetric(name string, metricType MetricType, value float64, tags map[string]string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.metrics[name] = &Metric{
		Name:      name,
		Type:      metricType,
		Value:     value,
		Timestamp: time.Now(),
		Tags:      tags,
	}
}

func (ms *MonitoringService) TrackRequest(method, path string, statusCode int, duration time.Duration) {
	ms.RecordMetric("http_requests_total", MetricTypeCounter, 1, map[string]string{
		"method": method,
		"path":   path,
		"status": fmt.Sprintf("%d", statusCode),
	})

	if statusCode >= 400 {
		ms.RecordMetric("http_errors_total", MetricTypeCounter, 1, map[string]string{
			"method": method,
			"path":   path,
		})
	}
}

func (ms *MonitoringService) CreateSecurityAlert(alertType string, severity AuditSeverity, message, ip string) {
	alert := &SecurityAlert{
		ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Severity:  severity,
		Type:      alertType,
		Message:   message,
		IPAddress: ip,
		Resolved:  false,
	}

	ms.mu.Lock()
	ms.alerts = append(ms.alerts, alert)
	ms.mu.Unlock()
}

func (ms *MonitoringService) GetMetrics() map[string]*Metric {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	metrics := make(map[string]*Metric)
	for k, v := range ms.metrics {
		metrics[k] = v
	}
	return metrics
}

func (ms *MonitoringService) GetAlerts() []*SecurityAlert {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	alerts := make([]*SecurityAlert, len(ms.alerts))
	copy(alerts, ms.alerts)
	return alerts
}

func (ms *MonitoringService) GetPerformanceMetrics() *PerformanceMetrics {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.perfMetrics == nil {
		return nil
	}

	metrics := *ms.perfMetrics
	return &metrics
}

func (ms *MonitoringService) monitorPerformance() {
	for {
		select {
		case <-ms.perfTicker.C:
			ms.collectPerformanceMetrics()
		case <-ms.stopChan:
			return
		}
	}
}

func (ms *MonitoringService) collectPerformanceMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := &PerformanceMetrics{
		Timestamp:      time.Now(),
		GoroutineCount: runtime.NumGoroutine(),
		HeapAllocMB:    float64(m.Alloc) / 1024 / 1024,
		HeapSysMB:      float64(m.Sys) / 1024 / 1024,
	}

	ms.mu.Lock()
	ms.perfMetrics = metrics
	ms.mu.Unlock()
}

func (ms *MonitoringService) Close() error {
	close(ms.stopChan)
	ms.perfTicker.Stop()
	return nil
}

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
	config := AuditConfig{
		LogDir:        tempDir,
		MaxFileSize:   1024 * 1024, // 1MB
		MaxFiles:      5,
		BufferSize:    100,
		FlushInterval: 50 * time.Millisecond,
	}

	auditService, err := NewAuditService(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create audit service: %v", err))
	}
	defer auditService.Close()

	// Create context with test data
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user_id", "test-user-123")
	ctx = context.WithValue(ctx, "client_ip", "192.168.1.100")

	// Test different types of audit events
	fmt.Println("Logging authentication events...")
	auditService.LogEvent(ctx, AuditEventLogin, AuditSeverityInfo, "login", "success", "User login successful", map[string]interface{}{
		"username": "testuser",
		"role":     "admin",
	})

	auditService.LogEvent(ctx, AuditEventLoginFailed, AuditSeverityWarning, "login", "failure", "User login failed", map[string]interface{}{
		"username": "baduser",
		"reason":   "invalid_password",
	})

	fmt.Println("Logging document events...")
	auditService.LogEvent(ctx, AuditEventDocumentSign, AuditSeverityInfo, "sign", "success", "Document signed successfully", map[string]interface{}{
		"document_id": "doc-123",
		"filename":    "test-document.pdf",
		"issuer":      "Test Organization",
	})

	fmt.Println("Logging verification events...")
	duration := 250 * time.Millisecond
	event := &AuditEvent{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		EventType: AuditEventVerificationSuccess,
		Severity:  AuditSeverityInfo,
		UserID:    "test-user-123",
		IPAddress: "192.168.1.300",
		Action:    "verify",
		Result:    "valid",
		Message:   "Document verification successful",
		Details: map[string]interface{}{
			"document_id":       "doc-123",
			"document_filename": "test-document.pdf",
			"hash_matches":      true,
			"signature_valid":   true,
		},
		Duration: &duration,
	}

	// Send verification event to buffer
	select {
	case auditService.buffer <- event:
	default:
		auditService.writeEvent(event)
	}

	fmt.Println("Logging security events...")
	auditService.LogEvent(ctx, AuditEventSecurityAlert, AuditSeverityCritical, "security_alert", "detected", "Critical security alert", map[string]interface{}{
		"source_ip":    "192.168.1.500",
		"threat_level": "high",
		"blocked":      true,
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

	// Parse and display sample log entries
	eventTypes := make(map[AuditEventType]int)
	for i, line := range lines {
		if line == "" {
			continue
		}

		var logEntry AuditEvent
		err := json.Unmarshal([]byte(line), &logEntry)
		if err != nil {
			fmt.Printf("Failed to parse log entry %d: %v\n", i, err)
			continue
		}

		eventTypes[logEntry.EventType]++

		if i < 3 { // Show first 3 entries
			fmt.Printf("  Entry %d: %s - %s - %s\n", i+1, logEntry.EventType, logEntry.Severity, logEntry.Message)
		}
	}

	fmt.Printf("Event type distribution:\n")
	for eventType, count := range eventTypes {
		fmt.Printf("  %s: %d\n", eventType, count)
	}

	// Test audit statistics
	stats := auditService.GetAuditStats()
	fmt.Printf("Audit stats: buffer_size=%v, buffered_events=%v, log_directory=%v\n", 
		stats["buffer_size"], stats["buffered_events"], stats["log_directory"])

	fmt.Println("✓ Audit Service test completed successfully")
}

func testMonitoringService() {
	fmt.Println("\n=== Testing Monitoring Service ===")

	// Create monitoring service
	monitoringService := NewMonitoringService()
	defer monitoringService.Close()

	// Test metrics recording
	fmt.Println("Recording test metrics...")
	monitoringService.RecordMetric("test_counter", MetricTypeCounter, 10, map[string]string{
		"endpoint": "/api/test",
	})

	monitoringService.RecordMetric("test_gauge", MetricTypeGauge, 42.5, map[string]string{
		"service": "test",
	})

	monitoringService.RecordMetric("test_timing", MetricTypeTiming, 150, map[string]string{
		"operation": "test",
	})

	// Test request tracking
	fmt.Println("Tracking test requests...")
	monitoringService.TrackRequest("GET", "/api/documents", 200, 100*time.Millisecond)
	monitoringService.TrackRequest("POST", "/api/documents", 400, 50*time.Millisecond)
	monitoringService.TrackRequest("GET", "/api/verify", 200, 75*time.Millisecond)

	// Test security alerts
	fmt.Println("Creating test security alerts...")
	monitoringService.CreateSecurityAlert("rate_limit_violation", AuditSeverityWarning, "Rate limit exceeded", "192.168.1.100")
	monitoringService.CreateSecurityAlert("brute_force_attempt", AuditSeverityCritical, "Multiple failed login attempts", "192.168.1.200")
	monitoringService.CreateSecurityAlert("suspicious_activity", AuditSeverityError, "Suspicious verification pattern", "192.168.1.300")

	// Wait for performance collection
	time.Sleep(100 * time.Millisecond)

	// Get and display metrics
	metrics := monitoringService.GetMetrics()
	fmt.Printf("Recorded %d metrics:\n", len(metrics))

	for name, metric := range metrics {
		fmt.Printf("  %s (%s): %.2f at %s\n", name, metric.Type, metric.Value, metric.Timestamp.Format("15:04:05"))
	}

	// Get and display alerts
	alerts := monitoringService.GetAlerts()
	fmt.Printf("Generated %d security alerts:\n", len(alerts))

	for i, alert := range alerts {
		fmt.Printf("  Alert %d: %s - %s - %s (IP: %s)\n", i+1, alert.Type, alert.Severity, alert.Message, alert.IPAddress)
	}

	// Test performance metrics
	perfMetrics := monitoringService.GetPerformanceMetrics()
	if perfMetrics != nil {
		fmt.Printf("Performance metrics:\n")
		fmt.Printf("  Goroutines: %d\n", perfMetrics.GoroutineCount)
		fmt.Printf("  Heap Alloc: %.2f MB\n", perfMetrics.HeapAllocMB)
		fmt.Printf("  Heap Sys: %.2f MB\n", perfMetrics.HeapSysMB)
		fmt.Printf("  Timestamp: %s\n", perfMetrics.Timestamp.Format("15:04:05"))
	} else {
		fmt.Println("Performance metrics not yet available")
	}

	fmt.Println("✓ Monitoring Service test completed successfully")
}