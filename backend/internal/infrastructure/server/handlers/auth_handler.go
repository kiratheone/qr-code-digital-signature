package handlers

import (
	"digital-signature-system/internal/domain/usecases"
	"digital-signature-system/internal/infrastructure/server/middleware"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authUseCase usecases.AuthUseCase
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authUseCase usecases.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// RegisterRequest represents a request to register a new user
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" validate:"username,safe_string,no_xss"`
	Password string `json:"password" binding:"required,min=8,max=100" validate:"password"`
	FullName string `json:"full_name" binding:"required,min=3,max=100" validate:"safe_string,no_xss"`
	Email    string `json:"email" binding:"required,email" validate:"safe_string"`
}

// LoginRequest represents a request to authenticate a user
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" validate:"username,safe_string,no_sql_injection"`
	Password string `json:"password" binding:"required,min=8,max=100" validate:"safe_string"`
}

// RefreshRequest represents a request to refresh a session
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,min=10,max=500" validate:"safe_string"`
}

// ChangePasswordRequest represents a request to change a user's password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=8,max=100" validate:"safe_string"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=100" validate:"password"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithValidationError(c, "Invalid registration data", err.Error())
		return
	}

	// Convert to use case request
	ucReq := usecases.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		FullName: req.FullName,
		Email:    req.Email,
	}

	// Call use case
	resp, err := h.authUseCase.Register(c.Request.Context(), ucReq)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate") {
			middleware.AbortWithError(c, middleware.NewConflictError("User already exists", err.Error()))
		} else {
			middleware.AbortWithValidationError(c, "Registration failed", err.Error())
		}
		return
	}

	// Return response
	c.JSON(http.StatusCreated, gin.H{
		"user_id":   resp.UserID,
		"username":  resp.Username,
		"full_name": resp.FullName,
		"email":     resp.Email,
		"message":   "User registered successfully",
	})
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithValidationError(c, "Invalid login data", err.Error())
		return
	}

	// Convert to use case request
	ucReq := usecases.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	// Call use case
	resp, err := h.authUseCase.Login(c.Request.Context(), ucReq)
	if err != nil {
		middleware.AbortWithAuthenticationError(c, "Invalid credentials")
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"user_id":       resp.UserID,
		"username":      resp.Username,
		"full_name":     resp.FullName,
		"email":         resp.Email,
		"role":          resp.Role,
		"token":         resp.SessionToken,
		"refresh_token": resp.RefreshToken,
		"expires_at":    resp.ExpiresAt,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from context (set by auth middleware)
	token := middleware.GetSessionToken(c)
	if token == "" {
		// Fallback to authorization header if not in context
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			middleware.AbortWithValidationError(c, "Authorization header is required", "Missing Authorization header")
			return
		}

		// Extract token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			middleware.AbortWithValidationError(c, "Invalid authorization header format", "Expected 'Bearer <token>'")
			return
		}
		token = parts[1]
	}

	// Call use case
	err := h.authUseCase.Logout(c.Request.Context(), token)
	if err != nil {
		middleware.AbortWithInternalError(c, "Failed to logout")
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// RefreshToken handles session refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call use case
	resp, err := h.authUseCase.RefreshSession(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"user_id":       resp.UserID,
		"username":      resp.Username,
		"full_name":     resp.FullName,
		"email":         resp.Email,
		"role":          resp.Role,
		"token":         resp.SessionToken,
		"refresh_token": resp.RefreshToken,
		"expires_at":    resp.ExpiresAt,
	})
}

// Me returns the current user's information
func (h *AuthHandler) Me(c *gin.Context) {
	// Get user info from context
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	fullName, _ := c.Get("full_name")
	email, _ := c.Get("email")
	role, _ := c.Get("role")

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"user_id":   userID,
		"username":  username,
		"full_name": fullName,
		"email":     email,
		"role":      role,
	})
}

// ChangePassword handles password change requests
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Call use case (this would need to be implemented in the auth use case)
	// For now, we'll return a not implemented response
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Password change functionality not implemented yet"})
}

// LogoutAll logs out all sessions for the current user
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// This would need to be implemented in the auth use case
	// For now, we'll return a not implemented response
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Logout all sessions functionality not implemented yet"})
}

// ValidateToken validates a token without requiring authentication
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	// Get token from query parameter
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token parameter is required"})
		return
	}

	// Call use case
	sessionInfo, err := h.authUseCase.ValidateSession(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token", "valid": false})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"valid":     true,
		"user_id":   sessionInfo.UserID,
		"username":  sessionInfo.Username,
		"full_name": sessionInfo.FullName,
		"email":     sessionInfo.Email,
		"role":      sessionInfo.Role,
	})
}