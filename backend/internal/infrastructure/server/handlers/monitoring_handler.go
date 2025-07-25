package handlers

import (
	"digital-signature-system/internal/infrastructure/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MonitoringHandler handles monitoring and audit endpoints
type MonitoringHandler struct {
	auditService      *services.AuditService
	monitoringService *services.MonitoringService
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(auditService *services.AuditService, monitoringService *services.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{
		auditService:      auditService,
		monitoringService: monitoringService,
	}
}

// GetMetrics returns current system metrics
func (h *MonitoringHandler) GetMetrics(c *gin.Context) {
	if h.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	metrics := h.monitoringService.GetMetrics()
	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// GetPerformanceMetrics returns system performance metrics
func (h *MonitoringHandler) GetPerformanceMetrics(c *gin.Context) {
	if h.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	perfMetrics := h.monitoringService.GetPerformanceMetrics()
	if perfMetrics == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Performance metrics not available",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"performance": perfMetrics,
	})
}

// GetHealthStatus returns overall system health status
func (h *MonitoringHandler) GetHealthStatus(c *gin.Context) {
	if h.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	health := h.monitoringService.GetHealthStatus()
	
	// Return appropriate HTTP status based on health
	statusCode := http.StatusOK
	if status, ok := health["status"].(string); ok && status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

// GetSecurityAlerts returns security alerts
func (h *MonitoringHandler) GetSecurityAlerts(c *gin.Context) {
	if h.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	// Parse query parameters
	resolvedParam := c.DefaultQuery("resolved", "false")
	resolved := resolvedParam == "true"

	alerts := h.monitoringService.GetAlerts(resolved)
	
	c.JSON(http.StatusOK, gin.H{
		"alerts":   alerts,
		"resolved": resolved,
		"count":    len(alerts),
	})
}

// ResolveSecurityAlert resolves a security alert
func (h *MonitoringHandler) ResolveSecurityAlert(c *gin.Context) {
	if h.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	alertID := c.Param("alertId")
	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Alert ID is required",
		})
		return
	}

	// Get user ID from context (assuming authentication middleware sets this)
	userID := "system" // Default
	if uid, exists := c.Get("user_id"); exists {
		userID = uid.(string)
	}

	err := h.monitoringService.ResolveAlert(alertID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Alert resolved successfully",
		"alert_id":  alertID,
		"resolved_by": userID,
	})
}

// GetAuditStats returns audit service statistics
func (h *MonitoringHandler) GetAuditStats(c *gin.Context) {
	if h.auditService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Audit service not available",
		})
		return
	}

	stats := h.auditService.GetAuditStats()
	c.JSON(http.StatusOK, gin.H{
		"audit_stats": stats,
	})
}

// CreateTestAlert creates a test security alert (for testing purposes)
func (h *MonitoringHandler) CreateTestAlert(c *gin.Context) {
	if h.monitoringService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Monitoring service not available",
		})
		return
	}

	// Parse request body
	var req struct {
		Type     string                 `json:"type" binding:"required"`
		Severity string                 `json:"severity" binding:"required"`
		Message  string                 `json:"message" binding:"required"`
		Details  map[string]interface{} `json:"details"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Convert severity string to AuditSeverity
	var severity services.AuditSeverity
	switch req.Severity {
	case "info":
		severity = services.AuditSeverityInfo
	case "warning":
		severity = services.AuditSeverityWarning
	case "error":
		severity = services.AuditSeverityError
	case "critical":
		severity = services.AuditSeverityCritical
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid severity level. Must be one of: info, warning, error, critical",
		})
		return
	}

	// Create test alert
	h.monitoringService.CreateSecurityAlert(c.Request.Context(), req.Type, severity, req.Message, req.Details)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Test alert created successfully",
		"type":    req.Type,
		"severity": req.Severity,
	})
}

// GetSystemStats returns comprehensive system statistics
func (h *MonitoringHandler) GetSystemStats(c *gin.Context) {
	response := gin.H{}

	// Get monitoring metrics if available
	if h.monitoringService != nil {
		response["metrics"] = h.monitoringService.GetMetrics()
		response["performance"] = h.monitoringService.GetPerformanceMetrics()
		response["health"] = h.monitoringService.GetHealthStatus()
		
		// Get alert counts
		unresolvedAlerts := h.monitoringService.GetAlerts(false)
		resolvedAlerts := h.monitoringService.GetAlerts(true)
		response["alerts"] = gin.H{
			"unresolved_count": len(unresolvedAlerts),
			"resolved_count":   len(resolvedAlerts),
			"total_count":      len(unresolvedAlerts) + len(resolvedAlerts),
		}
	}

	// Get audit stats if available
	if h.auditService != nil {
		response["audit"] = h.auditService.GetAuditStats()
	}

	c.JSON(http.StatusOK, response)
}

// GetMetricHistory returns historical data for a specific metric (placeholder)
func (h *MonitoringHandler) GetMetricHistory(c *gin.Context) {
	metricName := c.Param("metricName")
	if metricName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Metric name is required",
		})
		return
	}

	// Parse query parameters
	hoursParam := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursParam)
	if err != nil || hours <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid hours parameter",
		})
		return
	}

	// This is a placeholder implementation
	// In a real system, you would query a time-series database
	c.JSON(http.StatusOK, gin.H{
		"metric_name": metricName,
		"hours":       hours,
		"message":     "Metric history endpoint - implementation pending",
		"note":        "This would typically query a time-series database like InfluxDB or Prometheus",
	})
}