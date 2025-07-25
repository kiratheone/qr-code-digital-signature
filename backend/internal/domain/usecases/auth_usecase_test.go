package usecases_test

import (
	"context"
	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/domain/usecases"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock repositories and services
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *entities.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByToken(ctx context.Context, token string) (*entities.Session, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) ([]*entities.Session, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *entities.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockPasswordService struct {
	mock.Mock
}

func (m *MockPasswordService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordService) VerifyPassword(password, hash string) (bool, error) {
	args := m.Called(password, hash)
	return args.Bool(0), args.Error(1)
}

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateToken(userID, username string, role string, duration time.Duration) (string, error) {
	args := m.Called(userID, username, role, duration)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) ValidateToken(tokenString string) (*usecases.TokenClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.TokenClaims), args.Error(1)
}

func (m *MockTokenService) GenerateRefreshToken() (string, time.Time, error) {
	args := m.Called()
	return args.String(0), args.Get(1).(time.Time), args.Error(2)
}

func TestAuthUseCase_Register(t *testing.T) {
	// Setup mocks
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	passwordService := new(MockPasswordService)
	tokenService := new(MockTokenService)
	
	// Create use case
	useCase := usecases.NewAuthUseCase(
		userRepo,
		sessionRepo,
		passwordService,
		tokenService,
		time.Hour,
	)
	
	// Setup test data
	ctx := context.Background()
	req := usecases.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		FullName: "Test User",
		Email:    "test@example.com",
	}
	
	// Setup expectations
	userRepo.On("GetByUsername", ctx, req.Username).Return(nil, nil)
	userRepo.On("GetByEmail", ctx, req.Email).Return(nil, nil)
	passwordService.On("HashPassword", req.Password).Return("hashed_password", nil)
	userRepo.On("Create", ctx, mock.AnythingOfType("*entities.User")).Return(nil)
	
	// Call use case
	resp, err := useCase.Register(ctx, req)
	
	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, req.Username, resp.Username)
	assert.Equal(t, req.FullName, resp.FullName)
	assert.Equal(t, req.Email, resp.Email)
	
	// Verify expectations
	userRepo.AssertExpectations(t)
	passwordService.AssertExpectations(t)
}

func TestAuthUseCase_Login(t *testing.T) {
	// Setup mocks
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	passwordService := new(MockPasswordService)
	tokenService := new(MockTokenService)
	
	// Create use case
	useCase := usecases.NewAuthUseCase(
		userRepo,
		sessionRepo,
		passwordService,
		tokenService,
		time.Hour,
	)
	
	// Setup test data
	ctx := context.Background()
	req := usecases.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	
	user := &entities.User{
		ID:           "user-123",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		FullName:     "Test User",
		Email:        "test@example.com",
		Role:         "user",
		IsActive:     true,
	}
	
	// Setup expectations
	userRepo.On("GetByUsername", ctx, req.Username).Return(user, nil)
	passwordService.On("VerifyPassword", req.Password, user.PasswordHash).Return(true, nil)
	tokenService.On("GenerateToken", user.ID, user.Username, user.Role, time.Hour).Return("session_token", nil)
	tokenService.On("GenerateRefreshToken").Return("refresh_token", time.Now().Add(30*24*time.Hour), nil)
	sessionRepo.On("Create", ctx, mock.AnythingOfType("*entities.Session")).Return(nil)
	
	// Call use case
	resp, err := useCase.Login(ctx, req)
	
	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, user.ID, resp.UserID)
	assert.Equal(t, user.Username, resp.Username)
	assert.Equal(t, user.FullName, resp.FullName)
	assert.Equal(t, user.Email, resp.Email)
	assert.Equal(t, user.Role, resp.Role)
	assert.Equal(t, "session_token", resp.SessionToken)
	assert.Equal(t, "refresh_token", resp.RefreshToken)
	
	// Verify expectations
	userRepo.AssertExpectations(t)
	passwordService.AssertExpectations(t)
	tokenService.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestAuthUseCase_Logout(t *testing.T) {
	// Setup mocks
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	passwordService := new(MockPasswordService)
	tokenService := new(MockTokenService)
	
	// Create use case
	useCase := usecases.NewAuthUseCase(
		userRepo,
		sessionRepo,
		passwordService,
		tokenService,
		time.Hour,
	)
	
	// Setup test data
	ctx := context.Background()
	sessionToken := "session_token"
	
	// Setup expectations
	sessionRepo.On("Delete", ctx, sessionToken).Return(nil)
	
	// Call use case
	err := useCase.Logout(ctx, sessionToken)
	
	// Assert
	require.NoError(t, err)
	
	// Verify expectations
	sessionRepo.AssertExpectations(t)
}

func TestAuthUseCase_ValidateSession(t *testing.T) {
	// Setup mocks
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	passwordService := new(MockPasswordService)
	tokenService := new(MockTokenService)
	
	// Create use case
	useCase := usecases.NewAuthUseCase(
		userRepo,
		sessionRepo,
		passwordService,
		tokenService,
		time.Hour,
	)
	
	// Setup test data
	ctx := context.Background()
	sessionToken := "session_token"
	
	claims := &usecases.TokenClaims{
		UserID:   "user-123",
		Username: "testuser",
		Role:     "user",
	}
	
	session := &entities.Session{
		ID:           "session-123",
		UserID:       "user-123",
		SessionToken: sessionToken,
		RefreshToken: "refresh_token",
		ExpiresAt:    time.Now().Add(time.Hour),
		LastAccessed: time.Now(),
	}
	
	user := &entities.User{
		ID:           "user-123",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		FullName:     "Test User",
		Email:        "test@example.com",
		Role:         "user",
		IsActive:     true,
	}
	
	// Setup expectations
	tokenService.On("ValidateToken", sessionToken).Return(claims, nil)
	sessionRepo.On("GetByToken", ctx, sessionToken).Return(session, nil)
	userRepo.On("GetByID", ctx, claims.UserID).Return(user, nil)
	sessionRepo.On("Update", ctx, mock.AnythingOfType("*entities.Session")).Return(nil)
	
	// Call use case
	info, err := useCase.ValidateSession(ctx, sessionToken)
	
	// Assert
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, user.ID, info.UserID)
	assert.Equal(t, user.Username, info.Username)
	assert.Equal(t, user.FullName, info.FullName)
	assert.Equal(t, user.Email, info.Email)
	assert.Equal(t, user.Role, info.Role)
	
	// Verify expectations
	tokenService.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_RefreshSession(t *testing.T) {
	// Setup mocks
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	passwordService := new(MockPasswordService)
	tokenService := new(MockTokenService)
	
	// Create use case
	useCase := usecases.NewAuthUseCase(
		userRepo,
		sessionRepo,
		passwordService,
		tokenService,
		time.Hour,
	)
	
	// Setup test data
	ctx := context.Background()
	refreshToken := "refresh_token"
	
	session := &entities.Session{
		ID:           "session-123",
		UserID:       "user-123",
		SessionToken: "old_session_token",
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Hour),
		LastAccessed: time.Now(),
	}
	
	sessions := []*entities.Session{session}
	
	user := &entities.User{
		ID:           "user-123",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		FullName:     "Test User",
		Email:        "test@example.com",
		Role:         "user",
		IsActive:     true,
	}
	
	// Setup expectations
	sessionRepo.On("GetByRefreshToken", ctx, refreshToken).Return(sessions, nil)
	userRepo.On("GetByID", ctx, session.UserID).Return(user, nil)
	tokenService.On("GenerateToken", user.ID, user.Username, user.Role, time.Hour).Return("new_session_token", nil)
	tokenService.On("GenerateRefreshToken").Return("new_refresh_token", time.Now().Add(30*24*time.Hour), nil)
	sessionRepo.On("Update", ctx, mock.AnythingOfType("*entities.Session")).Return(nil)
	
	// Call use case
	resp, err := useCase.RefreshSession(ctx, refreshToken)
	
	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, user.ID, resp.UserID)
	assert.Equal(t, user.Username, resp.Username)
	assert.Equal(t, user.FullName, resp.FullName)
	assert.Equal(t, user.Email, resp.Email)
	assert.Equal(t, user.Role, resp.Role)
	assert.Equal(t, "new_session_token", resp.SessionToken)
	assert.Equal(t, "new_refresh_token", resp.RefreshToken)
	
	// Verify expectations
	sessionRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	tokenService.AssertExpectations(t)
}