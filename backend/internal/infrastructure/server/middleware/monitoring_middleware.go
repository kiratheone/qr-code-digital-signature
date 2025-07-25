package middleware

import (
	"context"
	"digital-signature-system/internal/infrastructure/services"
	"time"

	"github.com/gin-gonic/gin"
)

// MonitoringMiddleware provides request monitoring and metrics collection
type MonitoringMiddleware struct {
	monitoringService *services.MonitoringService
}

// NewMonitoringMiddleware creates a new monitoring middleware
func NewMonitoringMiddleware(monitoringService *services.MonitoringService) *MonitoringMiddleware {
	return &MonitoringMiddleware{
		monitoringService: monitoringService,
	}
}

// Middleware returns the monitoring middleware
func (mm *MonitoringMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Add client IP to context for monitoring
		ctx := context.WithValue(c.Request.Context(), "client_ip", c.ClientIP())
		c.Request = c.Request.WithContext(ctx)
		
		// Process request
		c.Next()
		
		// Calculate duration
		duration := time.Since(start)
		
		// Track request metrics
		if mm.monitoringService != nil {
			mm.monitoringService.TrackRequest(
				c.Request.Method,
				c.Request.URL.Path,
				c.Writer.Status(),
				duration,
			)
		}
	}
}

// RateLimitViolationMiddleware tracks rate limit violations
type RateLimitViolationMiddleware struct {
	monitoringService *services.MonitoringService
}

// NewRateLimitViolationMiddleware creates a new rate limit violation middleware
func NewRateLimitViolationMiddleware(monitoringService *services.MonitoringService) *RateLimitViolationMiddleware {
	return &RateLimitViolationMiddleware{
		monitoringService: monitoringService,
	}
}

// TrackViolation tracks a rate limit violation
func (rlm *RateLimitViolationMiddleware) TrackViolation(c *gin.Context) {
	if rlm.monitoringService != nil {
		ctx := context.WithValue(c.Request.Context(), "client_ip", c.ClientIP())
		rlm.monitoringService.TrackRateLimitViolation(ctx, c.ClientIP())
	}
}