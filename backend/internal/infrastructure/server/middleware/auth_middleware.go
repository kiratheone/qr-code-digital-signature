package middleware

import (
	"digital-signature-system/internal/domain/usecases"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware is a middleware for authentication
type AuthMiddleware struct {
	authUseCase usecases.AuthUseCase
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authUseCase usecases.AuthUseCase) *AuthMiddleware {
	return &AuthMiddleware{
		authUseCase: authUseCase,
	}
}

// Authenticate authenticates a request
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if authorization header is in the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Get token
		token := parts[1]

		// Validate token
		sessionInfo, err := m.authUseCase.ValidateSession(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", sessionInfo.UserID)
		c.Set("username", sessionInfo.Username)
		c.Set("full_name", sessionInfo.FullName)
		c.Set("email", sessionInfo.Email)
		c.Set("role", sessionInfo.Role)
		c.Set("session_token", token)

		c.Next()
	}
}

// RequireRole requires a specific role
func (m *AuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Check if user has the required role
		if userRole != role {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole requires the user to have any of the specified roles
func (m *AuthMiddleware) RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, role := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimiter implements a simple rate limiting middleware
func (m *AuthMiddleware) RateLimiter(limit int, duration time.Duration) gin.HandlerFunc {
	// Simple in-memory rate limiter
	type client struct {
		count    int
		lastSeen time.Time
	}
	clients := make(map[string]*client)

	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()
		
		// Get or create client
		cl, exists := clients[clientIP]
		if !exists {
			clients[clientIP] = &client{
				count:    0,
				lastSeen: time.Now(),
			}
			cl = clients[clientIP]
		}

		// Reset count if duration has passed
		if time.Since(cl.lastSeen) > duration {
			cl.count = 0
			cl.lastSeen = time.Now()
		}

		// Check if limit is exceeded
		if cl.count >= limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": duration.Seconds(),
			})
			c.Abort()
			return
		}

		// Increment count
		cl.count++
		cl.lastSeen = time.Now()

		c.Next()
	}
}

// CORS middleware for handling Cross-Origin Resource Sharing
func (m *AuthMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// GetUserID gets the user ID from the context
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	return userID.(string)
}

// GetUsername gets the username from the context
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	return username.(string)
}

// GetUserRole gets the user role from the context
func GetUserRole(c *gin.Context) string {
	role, exists := c.Get("role")
	if !exists {
		return ""
	}
	return role.(string)
}

// GetSessionToken gets the session token from the context
func GetSessionToken(c *gin.Context) string {
	token, exists := c.Get("session_token")
	if !exists {
		return ""
	}
	return token.(string)
}