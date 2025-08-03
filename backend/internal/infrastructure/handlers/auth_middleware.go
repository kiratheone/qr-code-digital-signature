package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"digital-signature-system/internal/domain/services"
)

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// RequireAuth is a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractTokenFromHeader(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "missing_token",
				Message: "Authorization token is required",
			})
			c.Abort()
			return
		}

		user, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			switch err {
			case services.ErrInvalidToken:
				c.JSON(http.StatusUnauthorized, ErrorResponse{
					Error:   "invalid_token",
					Message: "Invalid or malformed token",
				})
			case services.ErrSessionExpired:
				c.JSON(http.StatusUnauthorized, ErrorResponse{
					Error:   "token_expired",
					Message: "Token has expired",
				})
			case services.ErrUserNotFound:
				c.JSON(http.StatusUnauthorized, ErrorResponse{
					Error:   "user_not_found",
					Message: "User not found",
				})
			case services.ErrUserInactive:
				c.JSON(http.StatusForbidden, ErrorResponse{
					Error:   "user_inactive",
					Message: "User account is inactive",
				})
			default:
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "internal_error",
					Message: "Authentication failed",
				})
			}
			c.Abort()
			return
		}

		// Convert to AuthenticatedUser and store in context
		authUser := &services.AuthenticatedUser{
			ID:       user.ID,
			Username: user.Username,
			FullName: user.FullName,
			Email:    user.Email,
			Role:     user.Role,
		}

		c.Set("user", authUser)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Next()
	}
}

// RequireRole is a middleware that requires a specific role
func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Message: "User not authenticated",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok || role != requiredRole {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "insufficient_permissions",
				Message: "Insufficient permissions for this operation",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole is a middleware that requires any of the specified roles
func (m *AuthMiddleware) RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Message: "User not authenticated",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "insufficient_permissions",
				Message: "Insufficient permissions for this operation",
			})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		for _, requiredRole := range roles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   "insufficient_permissions",
			Message: "Insufficient permissions for this operation",
		})
		c.Abort()
	}
}

// OptionalAuth is a middleware that optionally authenticates the user
// If a token is provided, it validates it, but doesn't require authentication
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractTokenFromHeader(c)
		if token == "" {
			c.Next()
			return
		}

		user, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err == nil && user != nil {
			// Convert to AuthenticatedUser and store in context
			authUser := &services.AuthenticatedUser{
				ID:       user.ID,
				Username: user.Username,
				FullName: user.FullName,
				Email:    user.Email,
				Role:     user.Role,
			}

			c.Set("user", authUser)
			c.Set("user_id", user.ID)
			c.Set("user_role", user.Role)
		}

		c.Next()
	}
}

// CORS middleware for handling cross-origin requests
func (m *AuthMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimiting middleware (simple implementation)
func (m *AuthMiddleware) RateLimiting() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Simple rate limiting can be implemented here
		// For now, we'll just pass through
		c.Next()
	})
}

// RequestLogging middleware for logging requests
func (m *AuthMiddleware) RequestLogging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}