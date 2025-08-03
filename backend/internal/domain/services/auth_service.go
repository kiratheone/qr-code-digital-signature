package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/repositories"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrSessionExpired     = errors.New("session expired")
	ErrUserInactive       = errors.New("user account is inactive")
)

type AuthService struct {
	userRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
	jwtSecret   string
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

type LoginResponse struct {
	Token     string         `json:"token"`
	ExpiresAt time.Time      `json:"expires_at"`
	User      *entities.User `json:"user"`
}

type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type AuthenticatedUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

func NewAuthService(userRepo repositories.UserRepository, sessionRepo repositories.SessionRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtSecret:   jwtSecret,
	}
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, expiresAt, err := s.generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create session record
	session := &entities.Session{
		UserID:       user.ID,
		SessionToken: token,
		RefreshToken: s.generateRefreshToken(),
		ExpiresAt:    expiresAt,
		LastAccessed: time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*entities.User, error) {
	// Check if username already exists
	if existingUser, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Check if email already exists
	if existingUser, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &entities.User{
		Username:     req.Username,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Email:        req.Email,
		Role:         "user",
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Logout invalidates a user's session
func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessionRepo.Delete(ctx, token)
}

// ValidateToken validates a JWT token and returns the user
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*entities.User, error) {
	// Parse and validate JWT token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check if token is expired
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrSessionExpired
	}

	// Get user from database
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Update session last accessed time
	session, err := s.sessionRepo.GetByToken(ctx, tokenString)
	if err == nil && session != nil {
		session.LastAccessed = time.Now()
		s.sessionRepo.Update(ctx, session)
	}

	return user, nil
}

// ValidateSession validates a session token
func (s *AuthService) ValidateSession(ctx context.Context, sessionToken string) (*entities.User, error) {
	session, err := s.sessionRepo.GetByToken(ctx, sessionToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Check if session is expired
	if session.ExpiresAt.Before(time.Now()) {
		s.sessionRepo.Delete(ctx, sessionToken)
		return nil, ErrSessionExpired
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Update last accessed time
	session.LastAccessed = time.Now()
	s.sessionRepo.Update(ctx, session)

	return user, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	user.PasswordHash = hashedPassword
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all existing sessions for this user
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

// CleanupExpiredSessions removes expired sessions from the database
func (s *AuthService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.DeleteExpired(ctx)
}

// generateJWT creates a new JWT token for the user
func (s *AuthService) generateJWT(user *entities.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hours

	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "digital-signature-system",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// generateRefreshToken creates a random refresh token
func (s *AuthService) generateRefreshToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// hashPassword hashes a password using bcrypt
func (s *AuthService) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// ValidatePassword validates password strength
func (s *AuthService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}