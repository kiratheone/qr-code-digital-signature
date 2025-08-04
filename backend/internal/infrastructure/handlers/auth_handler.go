package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"digital-signature-system/internal/domain/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// Remove the old ErrorResponse struct as we now use StandardError from errors.go

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, "Invalid request format", err.Error())
		return
	}

	loginReq := services.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	response, err := h.authService.Login(c.Request.Context(), loginReq)
	if err != nil {
		MapServiceErrorToHTTP(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, "Invalid request format", err.Error())
		return
	}

	// Validate password strength
	if err := h.authService.ValidatePassword(req.Password); err != nil {
		RespondWithValidationError(c, "Invalid password", err.Error())
		return
	}

	registerReq := services.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		FullName: req.FullName,
		Email:    req.Email,
	}

	user, err := h.authService.Register(c.Request.Context(), registerReq)
	if err != nil {
		MapServiceErrorToHTTP(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	token := extractTokenFromHeader(c)
	if token == "" {
		RespondWithValidationError(c, "Authorization token is required")
		return
	}

	if err := h.authService.Logout(c.Request.Context(), token); err != nil {
		RespondWithInternalError(c, "Failed to logout", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		RespondWithUnauthorizedError(c, "User not authenticated")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, "Invalid request format", err.Error())
		return
	}

	// Validate new password strength
	if err := h.authService.ValidatePassword(req.NewPassword); err != nil {
		RespondWithValidationError(c, "Invalid password", err.Error())
		return
	}

	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		RespondWithUnauthorizedError(c, "User not authenticated")
		return
	}

	user, ok := userInterface.(*services.AuthenticatedUser)
	if !ok {
		RespondWithInternalError(c, "Invalid user context")
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), user.ID, req.OldPassword, req.NewPassword); err != nil {
		MapServiceErrorToHTTP(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// extractTokenFromHeader extracts the JWT token from the Authorization header
func extractTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check if the header starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}

	// Extract the token part
	return strings.TrimPrefix(authHeader, "Bearer ")
}