package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"digital-signature-system/internal/domain/services"
	"digital-signature-system/internal/infrastructure/logging"
	"digital-signature-system/internal/infrastructure/validation"
)

type AuthHandler struct {
	authService *services.AuthService
	validator   *validation.Validator
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
		validator:   validation.NewValidator(),
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, "Invalid request format", err.Error())
		return
	}

	// Validate and sanitize username
	sanitizedUsername, validationErr := h.validator.ValidateAndSanitizeString("username", req.Username, 1, 100, true)
	if validationErr != nil {
		RespondWithValidationError(c, "Invalid username", validationErr.Error())
		return
	}

	// Validate password (don't sanitize passwords as they may contain special characters)
	if strings.TrimSpace(req.Password) == "" {
		RespondWithValidationError(c, "Password is required")
		return
	}

	if len(req.Password) > 128 {
		RespondWithValidationError(c, "Password too long")
		return
	}

	loginReq := services.LoginRequest{
		Username: sanitizedUsername,
		Password: req.Password,
	}

	response, err := h.authService.Login(c.Request.Context(), loginReq)
	if err != nil {
		// Log failed authentication attempt
		logging.LogAuthentication(
			logging.AuditEventAuthFailure,
			"", // No user ID for failed login
			sanitizedUsername,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
			"FAILURE",
			map[string]interface{}{
				"error": err.Error(),
				"endpoint": "/api/auth/login",
			},
		)
		MapServiceErrorToHTTP(c, err)
		return
	}

	// Log successful authentication
	logging.LogAuthentication(
		logging.AuditEventLogin,
		response.User.ID,
		response.User.Username,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		"SUCCESS",
		map[string]interface{}{
			"endpoint": "/api/auth/login",
		},
	)

	c.JSON(http.StatusOK, response)
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, "Invalid request format", err.Error())
		return
	}

	// Validate and sanitize all input fields
	sanitizedUsername, validationErr := h.validator.ValidateAndSanitizeString("username", req.Username, 3, 50, true)
	if validationErr != nil {
		RespondWithValidationError(c, "Invalid username", validationErr.Error())
		return
	}

	sanitizedFullName, validationErr := h.validator.ValidateAndSanitizeString("full_name", req.FullName, 1, 100, true)
	if validationErr != nil {
		RespondWithValidationError(c, "Invalid full name", validationErr.Error())
		return
	}

	sanitizedEmail, validationErr := h.validator.ValidateEmail("email", req.Email, true)
	if validationErr != nil {
		RespondWithValidationError(c, "Invalid email", validationErr.Error())
		return
	}

	// Validate password strength using our validator
	if validationErr := h.validator.ValidatePassword("password", req.Password); validationErr != nil {
		RespondWithValidationError(c, "Invalid password", validationErr.Error())
		return
	}

	registerReq := services.RegisterRequest{
		Username: sanitizedUsername,
		Password: req.Password, // Don't sanitize password
		FullName: sanitizedFullName,
		Email:    sanitizedEmail,
	}

	user, err := h.authService.Register(c.Request.Context(), registerReq)
	if err != nil {
		// Log failed registration attempt
		logging.LogAuthentication(
			logging.AuditEventRegister,
			"", // No user ID for failed registration
			sanitizedUsername,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
			"FAILURE",
			map[string]interface{}{
				"error": err.Error(),
				"email": sanitizedEmail,
				"endpoint": "/api/auth/register",
			},
		)
		MapServiceErrorToHTTP(c, err)
		return
	}

	// Log successful registration
	logging.LogAuthentication(
		logging.AuditEventRegister,
		user.ID,
		user.Username,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		"SUCCESS",
		map[string]interface{}{
			"email": user.Email,
			"endpoint": "/api/auth/register",
		},
	)

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

	// Get user info for logging (if available)
	userID := ""
	username := ""
	if user, exists := c.Get("user"); exists {
		if authUser, ok := user.(*services.AuthenticatedUser); ok {
			userID = authUser.ID
			username = authUser.Username
		}
	}

	if err := h.authService.Logout(c.Request.Context(), token); err != nil {
		// Log failed logout attempt
		logging.LogAuthentication(
			logging.AuditEventLogout,
			userID,
			username,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
			"FAILURE",
			map[string]interface{}{
				"error": err.Error(),
				"endpoint": "/api/auth/logout",
			},
		)
		RespondWithInternalError(c, "Failed to logout", err.Error())
		return
	}

	// Log successful logout
	logging.LogAuthentication(
		logging.AuditEventLogout,
		userID,
		username,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		"SUCCESS",
		map[string]interface{}{
			"endpoint": "/api/auth/logout",
		},
	)

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

	// Validate old password (basic validation)
	if strings.TrimSpace(req.OldPassword) == "" {
		RespondWithValidationError(c, "Old password is required")
		return
	}

	if len(req.OldPassword) > 128 {
		RespondWithValidationError(c, "Old password too long")
		return
	}

	// Validate new password strength using our validator
	if validationErr := h.validator.ValidatePassword("new_password", req.NewPassword); validationErr != nil {
		RespondWithValidationError(c, "Invalid new password", validationErr.Error())
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
		// Log failed password change attempt
		logging.LogAuthentication(
			logging.AuditEventPasswordChange,
			user.ID,
			user.Username,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
			"FAILURE",
			map[string]interface{}{
				"error": err.Error(),
				"endpoint": "/api/change-password",
			},
		)
		MapServiceErrorToHTTP(c, err)
		return
	}

	// Log successful password change
	logging.LogAuthentication(
		logging.AuditEventPasswordChange,
		user.ID,
		user.Username,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		"SUCCESS",
		map[string]interface{}{
			"endpoint": "/api/change-password",
		},
	)

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