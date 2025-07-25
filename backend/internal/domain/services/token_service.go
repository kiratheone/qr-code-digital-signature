package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenService defines the interface for token operations
type TokenService interface {
	// GenerateToken generates a new token
	GenerateToken(userID, username string, role string, duration time.Duration) (string, error)
	
	// ValidateToken validates a token
	ValidateToken(tokenString string) (*TokenClaims, error)
	
	// GenerateRefreshToken generates a new refresh token
	GenerateRefreshToken() (string, time.Time, error)
}

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTTokenService implements TokenService using JWT
type JWTTokenService struct {
	secretKey []byte
	issuer    string
}

// NewJWTTokenService creates a new JWT token service
func NewJWTTokenService(secretKey string, issuer string) *JWTTokenService {
	return &JWTTokenService{
		secretKey: []byte(secretKey),
		issuer:    issuer,
	}
}

// GenerateToken generates a new JWT token
func (s *JWTTokenService) GenerateToken(userID, username string, role string, duration time.Duration) (string, error) {
	if userID == "" || username == "" {
		return "", errors.New("userID and username cannot be empty")
	}
	
	// Create claims
	now := time.Now()
	claims := TokenClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}
	
	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Sign token
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	
	return tokenString, nil
}

// ValidateToken validates a JWT token
func (s *JWTTokenService) ValidateToken(tokenString string) (*TokenClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		
		return s.secretKey, nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	
	// Validate token
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	
	// Get claims
	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	
	return claims, nil
}

// GenerateRefreshToken generates a new refresh token
func (s *JWTTokenService) GenerateRefreshToken() (string, time.Time, error) {
	// Generate random token
	refreshToken := uuid.New().String()
	
	// Set expiration time (30 days)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	
	return refreshToken, expiresAt, nil
}