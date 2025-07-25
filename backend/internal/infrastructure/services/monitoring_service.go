package services

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// MetricType represents different types of metrics
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeTiming    MetricType = "timing"
)

// Metric represents a single metric
type Metric struct {
	Name      string                 `json:"name"`
	Type      MetricType             `json:"type"`
	Value     float64                `json:"value"`
	Timestamp time.Time              `json:"timestamp"`
	Tags      map[string]string      `json:"tags,omitempty"`
	Labels    map[string]interface{} `json:"labels,omitempty"`
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Severity    AuditSeverity          `json:"severity"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy  string                 `json:"resolved_by,omitempty"`
}

// PerformanceMetrics represents system performance metrics
type PerformanceMetrics struct {
	Timestamp           time.Time `json:"timestamp"`
	CPUUsagePercent     float64   `json:"cpu_usage_percent"`
	MemoryUsageMB       float64   `json:"memory_usage_mb"`
	MemoryUsagePercent  float64   `json:"memory_usage_percent"`
	GoroutineCount      int       `json:"goroutine_count"`
	HeapAllocMB         float64   `json:"heap_alloc_mb"`
	HeapSysMB           float64   `json:"heap_sys_mb"`
	GCPauseMS           float64   `json:"gc_pause_ms"`
	RequestsPerSecond   float64   `json:"requests_per_second"`
	ActiveConnections   int       `json:"active_connections"`
	DatabaseConnections int       `json:"database_connections"`
}

// AuditServiceInterface defines the interface for audit logging
type AuditServiceInterface interface {
	LogSecurityEvent(ctx context.Context, eventType AuditEventType, severity AuditSeverity, message string, details map[string]interface{})
}

// MonitoringService provides comprehensive monitoring functionality
type MonitoringService struct {
	auditService    AuditServiceInterface
	metrics         map[string]*Metric
	alerts          []*SecurityAlert
	perfMetrics     *PerformanceMetrics
	mu              sync.RWMutex
	alertThresholds map[string]float64
	
	// Rate limiting tracking
	rateLimitViolations map[string]int
	rateLimitWindow     time.Duration
	rateLimitThreshold  int
	
	// Request tracking
	requestCounts       map[string]int
	requestCountWindow  time.Duration
	lastRequestReset    time.Time
	
	// Performance monitoring
	perfTicker          *time.Ticker
	alertTicker         *time.Ticker
	stopChan            chan struct{}
}

// MonitoringConfig represents monitoring service configuration
type MonitoringConfig struct {
	AuditService           AuditServiceInterface
	PerformanceInterval    time.Duration
	AlertCheckInterval     time.Duration
	RateLimitWindow        time.Duration
	RateLimitThreshold     int
	RequestCountWindow     time.Duration
	AlertThresholds        map[string]float64
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(config MonitoringConfig) *MonitoringService {
	if config.PerformanceInterval == 0 {
		config.PerformanceInterval = 30 * time.Second
	}
	if config.AlertCheckInterval == 0 {
		config.AlertCheckInterval = 10 * time.Second
	}
	if config.RateLimitWindow == 0 {
		config.RateLimitWindow = time.Minute
	}
	if config.RateLimitThreshold == 0 {
		config.RateLimitThreshold = 100
	}
	if config.RequestCountWindow == 0 {
		config.RequestCountWindow = time.Minute
	}
	if config.AlertThresholds == nil {
		config.AlertThresholds = map[string]float64{
			"cpu_usage":           80.0,
			"memory_usage":        85.0,
			"error_rate":          5.0,
			"response_time_p95":   2000.0, // 2 seconds
			"failed_logins":       10.0,
			"verification_failures": 20.0,
		}
	}

	service := &MonitoringService{
		auditService:        config.AuditService,
		metrics:             make(map[string]*Metric),
		alerts:              make([]*SecurityAlert, 0),
		alertThresholds:     config.AlertThresholds,
		rateLimitViolations: make(map[string]int),
		rateLimitWindow:     config.RateLimitWindow,
		rateLimitThreshold:  config.RateLimitThreshold,
		requestCounts:       make(map[string]int),
		requestCountWindow:  config.RequestCountWindow,
		lastRequestReset:    time.Now(),
		perfTicker:          time.NewTicker(config.PerformanceInterval),
		alertTicker:         time.NewTicker(config.AlertCheckInterval),
		stopChan:            make(chan struct{}),
	}

	// Start background monitoring
	go service.monitorPerformance()
	go service.checkAlerts()

	return service
}

// RecordMetric records a metric value
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

// IncrementCounter increments a counter metric
func (ms *MonitoringService) IncrementCounter(name string, tags map[string]string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if metric, exists := ms.metrics[name]; exists && metric.Type == MetricTypeCounter {
		metric.Value++
		metric.Timestamp = time.Now()
	} else {
		ms.metrics[name] = &Metric{
			Name:      name,
			Type:      MetricTypeCounter,
			Value:     1,
			Timestamp: time.Now(),
			Tags:      tags,
		}
	}
}

// SetGauge sets a gauge metric value
func (ms *MonitoringService) SetGauge(name string, value float64, tags map[string]string) {
	ms.RecordMetric(name, MetricTypeGauge, value, tags)
}

// RecordTiming records a timing metric
func (ms *MonitoringService) RecordTiming(name string, duration time.Duration, tags map[string]string) {
	ms.RecordMetric(name, MetricTypeTiming, float64(duration.Milliseconds()), tags)
}

// TrackRequest tracks request metrics
func (ms *MonitoringService) TrackRequest(method, path string, statusCode int, duration time.Duration) {
	// Record basic metrics
	ms.IncrementCounter("http_requests_total", map[string]string{
		"method": method,
		"path":   path,
		"status": fmt.Sprintf("%d", statusCode),
	})

	ms.RecordTiming("http_request_duration", duration, map[string]string{
		"method": method,
		"path":   path,
	})

	// Track error rates
	if statusCode >= 400 {
		ms.IncrementCounter("http_errors_total", map[string]string{
			"method": method,
			"path":   path,
			"status": fmt.Sprintf("%d", statusCode),
		})
	}

	// Update request counts for rate calculation
	ms.mu.Lock()
	key := fmt.Sprintf("%s:%s", method, path)
	ms.requestCounts[key]++
	
	// Reset counts if window has passed
	if time.Since(ms.lastRequestReset) > ms.requestCountWindow {
		for k := range ms.requestCounts {
			ms.requestCounts[k] = 0
		}
		ms.lastRequestReset = time.Now()
	}
	ms.mu.Unlock()
}

// TrackRateLimitViolation tracks rate limit violations
func (ms *MonitoringService) TrackRateLimitViolation(ctx context.Context, ip string) {
	ms.mu.Lock()
	ms.rateLimitViolations[ip]++
	violations := ms.rateLimitViolations[ip]
	ms.mu.Unlock()

	// Create security alert if threshold exceeded
	if violations > ms.rateLimitThreshold {
		ms.CreateSecurityAlert(ctx, "rate_limit_violation", AuditSeverityWarning, 
			fmt.Sprintf("Rate limit violations from IP %s: %d", ip, violations),
			map[string]interface{}{
				"ip_address":  ip,
				"violations":  violations,
				"threshold":   ms.rateLimitThreshold,
				"window":      ms.rateLimitWindow.String(),
			})
	}

	ms.IncrementCounter("rate_limit_violations", map[string]string{"ip": ip})
}

// TrackAuthFailure tracks authentication failures
func (ms *MonitoringService) TrackAuthFailure(ctx context.Context, userID, ip string, reason string) {
	ms.IncrementCounter("auth_failures_total", map[string]string{
		"user_id": userID,
		"reason":  reason,
	})

	// Check for brute force attempts
	ms.mu.RLock()
	key := fmt.Sprintf("auth_failures:%s", ip)
	failures := ms.requestCounts[key]
	ms.mu.RUnlock()

	if failures > 5 { // More than 5 failures from same IP
		ms.CreateSecurityAlert(ctx, "brute_force_attempt", AuditSeverityCritical,
			fmt.Sprintf("Multiple authentication failures from IP %s", ip),
			map[string]interface{}{
				"ip_address":     ip,
				"failure_count":  failures,
				"user_id":        userID,
				"failure_reason": reason,
			})
	}
}

// TrackVerificationFailure tracks verification failures
func (ms *MonitoringService) TrackVerificationFailure(ctx context.Context, documentID, ip, reason string) {
	ms.IncrementCounter("verification_failures_total", map[string]string{
		"document_id": documentID,
		"reason":      reason,
	})

	// Track suspicious verification patterns
	ms.mu.Lock()
	key := fmt.Sprintf("verification_failures:%s", ip)
	ms.requestCounts[key]++
	failures := ms.requestCounts[key]
	ms.mu.Unlock()

	if failures > 10 { // More than 10 verification failures from same IP
		ms.CreateSecurityAlert(ctx, "suspicious_verification_activity", AuditSeverityWarning,
			fmt.Sprintf("Multiple verification failures from IP %s", ip),
			map[string]interface{}{
				"ip_address":     ip,
				"failure_count":  failures,
				"document_id":    documentID,
				"failure_reason": reason,
			})
	}
}

// CreateSecurityAlert creates a new security alert
func (ms *MonitoringService) CreateSecurityAlert(ctx context.Context, alertType string, severity AuditSeverity, message string, details map[string]interface{}) {
	alert := &SecurityAlert{
		ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Severity:  severity,
		Type:      alertType,
		Source:    "monitoring_service",
		Message:   message,
		Details:   details,
		Resolved:  false,
	}

	// Extract context information
	if ctx != nil {
		if ip := ctx.Value("client_ip"); ip != nil {
			alert.IPAddress = fmt.Sprintf("%v", ip)
		}
		if userID := ctx.Value("user_id"); userID != nil {
			alert.UserID = fmt.Sprintf("%v", userID)
		}
		if requestID := ctx.Value("request_id"); requestID != nil {
			alert.RequestID = fmt.Sprintf("%v", requestID)
		}
	}

	ms.mu.Lock()
	ms.alerts = append(ms.alerts, alert)
	ms.mu.Unlock()

	// Log to audit service
	if ms.auditService != nil {
		ms.auditService.LogSecurityEvent(ctx, AuditEventSecurityAlert, severity, message, details)
	}
}

// GetMetrics returns current metrics
func (ms *MonitoringService) GetMetrics() map[string]*Metric {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	metrics := make(map[string]*Metric)
	for k, v := range ms.metrics {
		metrics[k] = v
	}
	return metrics
}

// GetAlerts returns current security alerts
func (ms *MonitoringService) GetAlerts(resolved bool) []*SecurityAlert {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var alerts []*SecurityAlert
	for _, alert := range ms.alerts {
		if alert.Resolved == resolved {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// ResolveAlert resolves a security alert
func (ms *MonitoringService) ResolveAlert(alertID, resolvedBy string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, alert := range ms.alerts {
		if alert.ID == alertID {
			alert.Resolved = true
			now := time.Now()
			alert.ResolvedAt = &now
			alert.ResolvedBy = resolvedBy
			return nil
		}
	}

	return fmt.Errorf("alert not found: %s", alertID)
}

// GetPerformanceMetrics returns current performance metrics
func (ms *MonitoringService) GetPerformanceMetrics() *PerformanceMetrics {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.perfMetrics == nil {
		return nil
	}

	// Return a copy
	metrics := *ms.perfMetrics
	return &metrics
}

// monitorPerformance monitors system performance in background
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

// collectPerformanceMetrics collects current performance metrics
func (ms *MonitoringService) collectPerformanceMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := &PerformanceMetrics{
		Timestamp:      time.Now(),
		GoroutineCount: runtime.NumGoroutine(),
		HeapAllocMB:    float64(m.Alloc) / 1024 / 1024,
		HeapSysMB:      float64(m.Sys) / 1024 / 1024,
		GCPauseMS:      float64(m.PauseNs[(m.NumGC+255)%256]) / 1000000,
	}

	// Calculate memory usage percentage (approximate)
	metrics.MemoryUsagePercent = (metrics.HeapAllocMB / metrics.HeapSysMB) * 100

	// Calculate requests per second
	ms.mu.RLock()
	totalRequests := 0
	for _, count := range ms.requestCounts {
		totalRequests += count
	}
	ms.mu.RUnlock()

	windowSeconds := ms.requestCountWindow.Seconds()
	metrics.RequestsPerSecond = float64(totalRequests) / windowSeconds

	ms.mu.Lock()
	ms.perfMetrics = metrics
	ms.mu.Unlock()

	// Record metrics
	ms.SetGauge("system_goroutines", float64(metrics.GoroutineCount), nil)
	ms.SetGauge("system_heap_alloc_mb", metrics.HeapAllocMB, nil)
	ms.SetGauge("system_heap_sys_mb", metrics.HeapSysMB, nil)
	ms.SetGauge("system_memory_usage_percent", metrics.MemoryUsagePercent, nil)
	ms.SetGauge("system_requests_per_second", metrics.RequestsPerSecond, nil)
}

// checkAlerts checks for alert conditions
func (ms *MonitoringService) checkAlerts() {
	for {
		select {
		case <-ms.alertTicker.C:
			ms.evaluateAlertConditions()
		case <-ms.stopChan:
			return
		}
	}
}

// evaluateAlertConditions evaluates current conditions for alerts
func (ms *MonitoringService) evaluateAlertConditions() {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.perfMetrics == nil {
		return
	}

	ctx := context.Background()

	// Check memory usage
	if threshold, exists := ms.alertThresholds["memory_usage"]; exists {
		if ms.perfMetrics.MemoryUsagePercent > threshold {
			ms.CreateSecurityAlert(ctx, "high_memory_usage", AuditSeverityWarning,
				fmt.Sprintf("High memory usage: %.2f%%", ms.perfMetrics.MemoryUsagePercent),
				map[string]interface{}{
					"memory_usage_percent": ms.perfMetrics.MemoryUsagePercent,
					"threshold":           threshold,
					"heap_alloc_mb":       ms.perfMetrics.HeapAllocMB,
				})
		}
	}

	// Check goroutine count
	if ms.perfMetrics.GoroutineCount > 1000 {
		ms.CreateSecurityAlert(ctx, "high_goroutine_count", AuditSeverityWarning,
			fmt.Sprintf("High goroutine count: %d", ms.perfMetrics.GoroutineCount),
			map[string]interface{}{
				"goroutine_count": ms.perfMetrics.GoroutineCount,
			})
	}

	// Check error rates
	if errorMetric, exists := ms.metrics["http_errors_total"]; exists {
		if requestMetric, exists := ms.metrics["http_requests_total"]; exists {
			errorRate := (errorMetric.Value / requestMetric.Value) * 100
			if threshold, exists := ms.alertThresholds["error_rate"]; exists && errorRate > threshold {
				ms.CreateSecurityAlert(ctx, "high_error_rate", AuditSeverityError,
					fmt.Sprintf("High error rate: %.2f%%", errorRate),
					map[string]interface{}{
						"error_rate": errorRate,
						"threshold":  threshold,
						"errors":     errorMetric.Value,
						"requests":   requestMetric.Value,
					})
			}
		}
	}
}

// GetHealthStatus returns overall system health status
func (ms *MonitoringService) GetHealthStatus() map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	status := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"checks":    make(map[string]interface{}),
	}

	checks := status["checks"].(map[string]interface{})

	// Check performance metrics
	if ms.perfMetrics != nil {
		checks["memory"] = map[string]interface{}{
			"status":      "ok",
			"usage_percent": ms.perfMetrics.MemoryUsagePercent,
		}

		if ms.perfMetrics.MemoryUsagePercent > 90 {
			checks["memory"].(map[string]interface{})["status"] = "critical"
			status["status"] = "unhealthy"
		} else if ms.perfMetrics.MemoryUsagePercent > 80 {
			checks["memory"].(map[string]interface{})["status"] = "warning"
		}

		checks["goroutines"] = map[string]interface{}{
			"status": "ok",
			"count":  ms.perfMetrics.GoroutineCount,
		}

		if ms.perfMetrics.GoroutineCount > 1000 {
			checks["goroutines"].(map[string]interface{})["status"] = "warning"
		}
	}

	// Check for unresolved critical alerts
	criticalAlerts := 0
	for _, alert := range ms.alerts {
		if !alert.Resolved && alert.Severity == AuditSeverityCritical {
			criticalAlerts++
		}
	}

	checks["security"] = map[string]interface{}{
		"status":          "ok",
		"critical_alerts": criticalAlerts,
	}

	if criticalAlerts > 0 {
		checks["security"].(map[string]interface{})["status"] = "critical"
		status["status"] = "unhealthy"
	}

	return status
}

// Close stops the monitoring service
func (ms *MonitoringService) Close() error {
	close(ms.stopChan)
	ms.perfTicker.Stop()
	ms.alertTicker.Stop()
	return nil
}