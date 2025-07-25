package usecases

import (
	"context"
	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
	"digital-signature-system/internal/domain/services"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AuthUseCase defines the interface for authentication operations
type AuthUseCase interface {
	// Register registers a new user
	Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error)
	
	// Login authenticates a user and returns a session
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	
	// Logout invalidates a user's session
	Logout(ctx context.Context, sessionToken string) error
	
	// ValidateSession validates a session token
	ValidateSession(ctx context.Context, sessionToken string) (*SessionInfo, error)
	
	// RefreshSession refreshes a session using a refresh token
	RefreshSession(ctx context.Context, refreshToken string) (*LoginResponse, error)
}

// RegisterRequest represents a request to register a new user
type RegisterRequest struct {
	Username string
	Password string
	FullName string
	Email    string
}

// RegisterResponse represents the response to a register request
type RegisterResponse struct {
	UserID   string
	Username string
	FullName string
	Email    string
}

// LoginRequest represents a request to authenticate a user
type LoginRequest struct {
	Username string
	Password string
}

// LoginResponse represents the response to a login request
type LoginResponse struct {
	UserID       string
	Username     string
	FullName     string
	Email        string
	Role         string
	SessionToken string
	RefreshToken string
	ExpiresAt    time.Time
}

// SessionInfo represents information about a session
type SessionInfo struct {
	UserID   string
	Username string
	FullName string
	Email    string
	Role     string
}

// AuthAuditService interface for audit logging
type AuthAuditService interface {
	LogAuthEvent(ctx context.Context, eventType string, userID, ip string, success bool, details map[string]interface{})
}

// AuthMonitoringService interface for monitoring
type AuthMonitoringService interface {
	TrackAuthFailure(ctx context.Context, userID, ip string, reason string)
}

// AuthUseCaseImpl implements AuthUseCase
type AuthUseCaseImpl struct {
	userRepo          repositories.UserRepository
	sessionRepo       repositories.SessionRepository
	passwordService   services.PasswordService
	tokenService      services.TokenService
	sessionDuration   time.Duration
	auditService      AuthAuditService
	monitoringService AuthMonitoringService
}

// NewAuthUseCase creates a new authentication use case
func NewAuthUseCase(
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	passwordService services.PasswordService,
	tokenService services.TokenService,
	sessionDuration time.Duration,
	auditService AuthAuditService,
	monitoringService AuthMonitoringService,
) *AuthUseCaseImpl {
	return &AuthUseCaseImpl{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		passwordService:   passwordService,
		tokenService:      tokenService,
		sessionDuration:   sessionDuration,
		auditService:      auditService,
		monitoringService: monitoringService,
	}
}

// Register registers a new user
func (uc *AuthUseCaseImpl) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	// Validate request
	if req.Username == "" || req.Password == "" || req.FullName == "" || req.Email == "" {
		return nil, errors.New("all fields are required")
	}
	
	// Check if username already exists
	existingUser, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}
	
	// Check if email already exists
	existingUser, err = uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}
	
	// Hash password
	passwordHash, err := uc.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Create user
	user := &entities.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		PasswordHash: passwordHash,
		FullName:     req.FullName,
		Email:        req.Email,
		Role:         "user",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	// Save user
	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	// Return response
	return &RegisterResponse{
		UserID:   user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
	}, nil
}

// Login authenticates a user and returns a session
func (uc *AuthUseCaseImpl) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Extract IP address from context
	var clientIP string
	if ip := ctx.Value("client_ip"); ip != nil {
		clientIP = fmt.Sprintf("%v", ip)
	}
	
	// Validate request
	if req.Username == "" || req.Password == "" {
		// Log audit event for invalid request
		if uc.auditService != nil {
			uc.auditService.LogAuthEvent(ctx, "login_attempt", req.Username, clientIP, false, map[string]interface{}{
				"reason": "missing_credentials",
			})
		}
		
		// Track monitoring failure
		if uc.monitoringService != nil {
			uc.monitoringService.TrackAuthFailure(ctx, req.Username, clientIP, "missing_credentials")
		}
		
		return nil, errors.New("username and password are required")
	}
	
	// Get user by username
	user, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// Log audit event for database error
		if uc.auditService != nil {
			uc.auditService.LogAuthEvent(ctx, "login_attempt", req.Username, clientIP, false, map[string]interface{}{
				"reason": "database_error",
				"error":  err.Error(),
			})
		}
		
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		// Log audit event for invalid username
		if uc.auditService != nil {
			uc.auditService.LogAuthEvent(ctx, "login_attempt", req.Username, clientIP, false, map[string]interface{}{
				"reason": "invalid_username",
			})
		}
		
		// Track monitoring failure
		if uc.monitoringService != nil {
			uc.monitoringService.TrackAuthFailure(ctx, req.Username, clientIP, "invalid_username")
		}
		
		return nil, errors.New("invalid username or password")
	}
	
	// Check if user is active
	if !user.IsActive {
		// Log audit event for inactive user
		if uc.auditService != nil {
			uc.auditService.LogAuthEvent(ctx, "login_attempt", user.ID, clientIP, false, map[string]interface{}{
				"reason":   "user_inactive",
				"username": req.Username,
			})
		}
		
		// Track monitoring failure
		if uc.monitoringService != nil {
			uc.monitoringService.TrackAuthFailure(ctx, user.ID, clientIP, "user_inactive")
		}
		
		return nil, errors.New("user account is inactive")
	}
	
	// Verify password
	valid, err := uc.passwordService.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		// Log audit event for password verification error
		if uc.auditService != nil {
			uc.auditService.LogAuthEvent(ctx, "login_attempt", user.ID, clientIP, false, map[string]interface{}{
				"reason":   "password_verification_error",
				"username": req.Username,
				"error":    err.Error(),
			})
		}
		
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}
	if !valid {
		// Log audit event for invalid password
		if uc.auditService != nil {
			uc.auditService.LogAuthEvent(ctx, "login_attempt", user.ID, clientIP, false, map[string]interface{}{
				"reason":   "invalid_password",
				"username": req.Username,
			})
		}
		
		// Track monitoring failure
		if uc.monitoringService != nil {
			uc.monitoringService.TrackAuthFailure(ctx, user.ID, clientIP, "invalid_password")
		}
		
		return nil, errors.New("invalid username or password")
	}
	
	// Generate session token
	sessionToken, err := uc.tokenService.GenerateToken(user.ID, user.Username, user.Role, uc.sessionDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}
	
	// Generate refresh token
	refreshToken, expiresAt, err := uc.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	
	// Create session
	session := &entities.Session{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
	}
	
	// Save session
	err = uc.sessionRepo.Create(ctx, session)
	if err != nil {
		// Log audit event for session creation failure
		if uc.auditService != nil {
			uc.auditService.LogAuthEvent(ctx, "login_attempt", user.ID, clientIP, false, map[string]interface{}{
				"reason":   "session_creation_failed",
				"username": req.Username,
				"error":    err.Error(),
			})
		}
		
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	
	// Log successful login
	if uc.auditService != nil {
		uc.auditService.LogAuthEvent(ctx, "login_success", user.ID, clientIP, true, map[string]interface{}{
			"username":    req.Username,
			"user_role":   user.Role,
			"session_id":  session.ID,
			"expires_at":  expiresAt,
		})
	}
	
	// Return response
	return &LoginResponse{
		UserID:       user.ID,
		Username:     user.Username,
		FullName:     user.FullName,
		Email:        user.Email,
		Role:         user.Role,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// Logout invalidates a user's session
func (uc *AuthUseCaseImpl) Logout(ctx context.Context, sessionToken string) error {
	// Extract IP address from context
	var clientIP string
	if ip := ctx.Value("client_ip"); ip != nil {
		clientIP = fmt.Sprintf("%v", ip)
	}
	
	// Validate request
	if sessionToken == "" {
		return errors.New("session token is required")
	}
	
	// Get session info before deletion for audit logging
	var userID string
	if session, err := uc.sessionRepo.GetByToken(ctx, sessionToken); err == nil && session != nil {
		userID = session.UserID
	}
	
	// Delete session
	err := uc.sessionRepo.Delete(ctx, sessionToken)
	if err != nil {
		// Log audit event for logout failure
		if uc.auditService != nil {
			uc.auditService.LogAuthEvent(ctx, "logout_attempt", userID, clientIP, false, map[string]interface{}{
				"reason": "session_deletion_failed",
				"error":  err.Error(),
			})
		}
		
		return fmt.Errorf("failed to delete session: %w", err)
	}
	
	// Log successful logout
	if uc.auditService != nil {
		uc.auditService.LogAuthEvent(ctx, "logout_success", userID, clientIP, true, map[string]interface{}{
			"session_token": sessionToken[:10] + "...", // Log partial token for reference
		})
	}
	
	return nil
}

// ValidateSession validates a session token
func (uc *AuthUseCaseImpl) ValidateSession(ctx context.Context, sessionToken string) (*SessionInfo, error) {
	// Extract IP address from context
	var clientIP string
	if ip := ctx.Value("client_ip"); ip != nil {
		clientIP = fmt.Sprintf("%v", ip)
	}
	
	// Validate request
	if sessionToken == "" {
		return nil, errors.New("session token is required")
	}
	
	// Validate token
	claims, err := uc.tokenService.ValidateToken(sessionToken)
	if err != nil {
		// Log audit event for invalid token
		if uc.auditService != nil {
			uc.auditService.LogAuthEvent(ctx, "session_validation", "", clientIP, false, map[string]interface{}{
				"reason": "invalid_token",
				"error":  err.Error(),
			})
		}
		
		return nil, fmt.Errorf("invalid session token: %w", err)
	}
	
	// Get session
	session, err := uc.sessionRepo.GetByToken(ctx, sessionToken)
	if err != nil {
		// Log audit event for session lookup failure
		if uc.auditService != nil {
			uc.auditService.LogAuthEvent(ctx, "session_validation", claims.UserID, clientIP, false, map[string]interface{}{
				"reason": "session_lookup_failed",
				"error":  err.Error(),
			})
		}
		
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		// Log audit event for session not found
		if uc.auditService != nil {
			uc.auditService.LogAuthEvent(ctx, "session_validation", claims.UserID, clientIP, false, map[string]interface{}{
				"reason": "session_not_found",
			})
		}
		
		return nil, errors.New("session not found")
	}
	
	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		_ = uc.sessionRepo.Delete(ctx, sessionToken)
		
		// Log audit event for expired session
		if uc.auditService != nil {
			uc.auditService.LogAuthEvent(ctx, "session_validation", session.UserID, clientIP, false, map[string]interface{}{
				"reason":     "session_expired",
				"expired_at": session.ExpiresAt,
			})
		}
		
		return nil, errors.New("session expired")
	}
	
	// Get user
	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	
	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}
	
	// Update last accessed time
	session.LastAccessed = time.Now()
	err = uc.sessionRepo.Update(ctx, session)
	if err != nil {
		// Non-critical error, just log it
		fmt.Printf("Failed to update session last accessed time: %v\n", err)
	}
	
	// Return session info
	return &SessionInfo{
		UserID:   user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
		Role:     user.Role,
	}, nil
}

// RefreshSession refreshes a session using a refresh token
func (uc *AuthUseCaseImpl) RefreshSession(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// Validate request
	if refreshToken == "" {
		return nil, errors.New("refresh token is required")
	}
	
	// Find session by refresh token
	var session *entities.Session
	sessions, err := uc.sessionRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if len(sessions) == 0 {
		return nil, errors.New("invalid refresh token")
	}
	session = sessions[0]
	
	// Check if refresh token is expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		_ = uc.sessionRepo.Delete(ctx, session.SessionToken)
		return nil, errors.New("refresh token expired")
	}
	
	// Get user
	user, err := uc.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	
	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}
	
	// Generate new session token
	sessionToken, err := uc.tokenService.GenerateToken(user.ID, user.Username, user.Role, uc.sessionDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}
	
	// Generate new refresh token
	newRefreshToken, expiresAt, err := uc.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	
	// Update session
	session.SessionToken = sessionToken
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = expiresAt
	session.LastAccessed = time.Now()
	
	// Save session
	err = uc.sessionRepo.Update(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}
	
	// Return response
	return &LoginResponse{
		UserID:       user.ID,
		Username:     user.Username,
		FullName:     user.FullName,
		Email:        user.Email,
		Role:         user.Role,
		SessionToken: sessionToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}