package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Requests     int           // Number of requests allowed
	Window       time.Duration // Time window
	KeyGenerator func(*gin.Context) string // Function to generate rate limit key
	SkipFunc     func(*gin.Context) bool   // Function to skip rate limiting
	OnLimitReached func(*gin.Context)      // Callback when limit is reached
	MonitoringService interface {
		TrackRateLimitViolation(ctx context.Context, ip string)
	} // Optional monitoring service
}

// RateLimitEntry represents a rate limit entry
type RateLimitEntry struct {
	Count      int
	ResetTime  time.Time
	LastAccess time.Time
	mutex      sync.RWMutex
}

// RateLimiter implements rate limiting functionality
type RateLimiter struct {
	config  RateLimitConfig
	entries map[string]*RateLimitEntry
	mutex   sync.RWMutex
	cleanup *time.Ticker
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	if config.KeyGenerator == nil {
		config.KeyGenerator = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}

	if config.SkipFunc == nil {
		config.SkipFunc = func(c *gin.Context) bool {
			return false
		}
	}

	rl := &RateLimiter{
		config:  config,
		entries: make(map[string]*RateLimitEntry),
		cleanup: time.NewTicker(time.Minute), // Cleanup every minute
	}

	// Start cleanup goroutine
	go rl.cleanupExpiredEntries()

	return rl
}

// Middleware returns the rate limiting middleware
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if configured to skip
		if rl.config.SkipFunc(c) {
			c.Next()
			return
		}

		key := rl.config.KeyGenerator(c)
		now := time.Now()

		rl.mutex.Lock()
		entry, exists := rl.entries[key]
		if !exists {
			entry = &RateLimitEntry{
				Count:      0,
				ResetTime:  now.Add(rl.config.Window),
				LastAccess: now,
			}
			rl.entries[key] = entry
		}
		rl.mutex.Unlock()

		entry.mutex.Lock()
		defer entry.mutex.Unlock()

		// Reset if window has passed
		if now.After(entry.ResetTime) {
			entry.Count = 0
			entry.ResetTime = now.Add(rl.config.Window)
		}

		entry.LastAccess = now

		// Check if limit exceeded
		if entry.Count >= rl.config.Requests {
			// Track rate limit violation in monitoring service
			if rl.config.MonitoringService != nil {
				ctx := context.WithValue(c.Request.Context(), "client_ip", c.ClientIP())
				rl.config.MonitoringService.TrackRateLimitViolation(ctx, c.ClientIP())
			}

			// Set rate limit headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.Requests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(entry.ResetTime.Unix(), 10))
			c.Header("Retry-After", strconv.Itoa(int(time.Until(entry.ResetTime).Seconds())))

			// Call callback if configured
			if rl.config.OnLimitReached != nil {
				rl.config.OnLimitReached(c)
			}

			// Return rate limit error
			apiError := NewRateLimitError(int(time.Until(entry.ResetTime).Seconds()))
			requestID := ""
			if id, exists := c.Get("request_id"); exists {
				requestID = id.(string)
			}
			apiError.RequestID = requestID
			apiError.Path = c.Request.URL.Path
			apiError.Method = c.Request.Method

			c.JSON(http.StatusTooManyRequests, apiError)
			c.Abort()
			return
		}

		// Increment counter
		entry.Count++

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(rl.config.Requests-entry.Count))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(entry.ResetTime.Unix(), 10))

		c.Next()
	}
}

// cleanupExpiredEntries removes expired entries from memory
func (rl *RateLimiter) cleanupExpiredEntries() {
	for range rl.cleanup.C {
		now := time.Now()
		rl.mutex.Lock()
		
		for key, entry := range rl.entries {
			entry.mutex.RLock()
			// Remove entries that haven't been accessed for 2x the window duration
			if now.Sub(entry.LastAccess) > rl.config.Window*2 {
				delete(rl.entries, key)
			}
			entry.mutex.RUnlock()
		}
		
		rl.mutex.Unlock()
	}
}

// Stop stops the rate limiter cleanup
func (rl *RateLimiter) Stop() {
	rl.cleanup.Stop()
}

// Predefined rate limiters

// NewGlobalRateLimiter creates a global rate limiter (applies to all IPs)
func NewGlobalRateLimiter(requests int, window time.Duration) *RateLimiter {
	return NewRateLimiter(RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyGenerator: func(c *gin.Context) string {
			return "global"
		},
	})
}

// NewIPRateLimiter creates an IP-based rate limiter
func NewIPRateLimiter(requests int, window time.Duration) *RateLimiter {
	return NewRateLimiter(RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyGenerator: func(c *gin.Context) string {
			return c.ClientIP()
		},
	})
}

// NewUserRateLimiter creates a user-based rate limiter
func NewUserRateLimiter(requests int, window time.Duration) *RateLimiter {
	return NewRateLimiter(RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyGenerator: func(c *gin.Context) string {
			if userID, exists := c.Get("user_id"); exists {
				return fmt.Sprintf("user:%s", userID)
			}
			return c.ClientIP() // Fallback to IP if no user
		},
		SkipFunc: func(c *gin.Context) bool {
			// Skip for unauthenticated requests (they'll use IP-based limiting)
			_, exists := c.Get("user_id")
			return !exists
		},
	})
}

// NewEndpointRateLimiter creates an endpoint-specific rate limiter
func NewEndpointRateLimiter(requests int, window time.Duration) *RateLimiter {
	return NewRateLimiter(RateLimitConfig{
		Requests: requests,
		Window:   window,
		KeyGenerator: func(c *gin.Context) string {
			return fmt.Sprintf("%s:%s:%s", c.ClientIP(), c.Request.Method, c.Request.URL.Path)
		},
	})
}

// RateLimitMiddleware provides different rate limiting strategies
type RateLimitMiddleware struct {
	globalLimiter   *RateLimiter
	ipLimiter      *RateLimiter
	userLimiter    *RateLimiter
	endpointLimiter *RateLimiter
}

// NewRateLimitMiddleware creates a new rate limit middleware with multiple strategies
func NewRateLimitMiddleware() *RateLimitMiddleware {
	return &RateLimitMiddleware{
		globalLimiter:   NewGlobalRateLimiter(10000, time.Hour),        // 10k requests per hour globally
		ipLimiter:      NewIPRateLimiter(1000, time.Hour),             // 1k requests per hour per IP
		userLimiter:    NewUserRateLimiter(5000, time.Hour),           // 5k requests per hour per user
		endpointLimiter: NewEndpointRateLimiter(100, time.Minute),     // 100 requests per minute per endpoint per IP
	}
}

// GlobalLimit applies global rate limiting
func (rlm *RateLimitMiddleware) GlobalLimit() gin.HandlerFunc {
	return rlm.globalLimiter.Middleware()
}

// IPLimit applies IP-based rate limiting
func (rlm *RateLimitMiddleware) IPLimit() gin.HandlerFunc {
	return rlm.ipLimiter.Middleware()
}

// UserLimit applies user-based rate limiting
func (rlm *RateLimitMiddleware) UserLimit() gin.HandlerFunc {
	return rlm.userLimiter.Middleware()
}

// EndpointLimit applies endpoint-specific rate limiting
func (rlm *RateLimitMiddleware) EndpointLimit() gin.HandlerFunc {
	return rlm.endpointLimiter.Middleware()
}

// StrictLimit applies strict rate limiting for sensitive endpoints
func (rlm *RateLimitMiddleware) StrictLimit() gin.HandlerFunc {
	strictLimiter := NewIPRateLimiter(10, time.Minute) // 10 requests per minute per IP
	return strictLimiter.Middleware()
}

// AuthLimit applies rate limiting for authentication endpoints
func (rlm *RateLimitMiddleware) AuthLimit() gin.HandlerFunc {
	authLimiter := NewIPRateLimiter(5, time.Minute) // 5 auth attempts per minute per IP
	return authLimiter.Middleware()
}

// Stop stops all rate limiters
func (rlm *RateLimitMiddleware) Stop() {
	rlm.globalLimiter.Stop()
	rlm.ipLimiter.Stop()
	rlm.userLimiter.Stop()
	rlm.endpointLimiter.Stop()
}