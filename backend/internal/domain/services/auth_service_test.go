package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"digital-signature-system/internal/domain/entities"
)

// Mock repositories
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

func (m *MockSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.Session, error) {
	args := m.Called(ctx, userID)
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

func (m *MockSessionRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name           string
		request        LoginRequest
		setupMocks     func(*MockUserRepository, *MockSessionRepository)
		expectedError  error
		expectedResult bool
	}{
		{
			name: "successful login",
			request: LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &entities.User{
					ID:           "user-1",
					Username:     "testuser",
					PasswordHash: string(hashedPassword),
					IsActive:     true,
				}
				userRepo.On("GetByUsername", mock.Anything, "testuser").Return(user, nil)
				sessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Session")).Return(nil)
			},
			expectedError:  nil,
			expectedResult: true,
		},
		{
			name: "invalid username",
			request: LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				userRepo.On("GetByUsername", mock.Anything, "nonexistent").Return(nil, errors.New("user not found"))
			},
			expectedError:  ErrInvalidCredentials,
			expectedResult: false,
		},
		{
			name: "invalid password",
			request: LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &entities.User{
					ID:           "user-1",
					Username:     "testuser",
					PasswordHash: string(hashedPassword),
					IsActive:     true,
				}
				userRepo.On("GetByUsername", mock.Anything, "testuser").Return(user, nil)
			},
			expectedError:  ErrInvalidCredentials,
			expectedResult: false,
		},
		{
			name: "inactive user",
			request: LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &entities.User{
					ID:           "user-1",
					Username:     "testuser",
					PasswordHash: string(hashedPassword),
					IsActive:     false,
				}
				userRepo.On("GetByUsername", mock.Anything, "testuser").Return(user, nil)
			},
			expectedError:  ErrUserInactive,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			sessionRepo := new(MockSessionRepository)
			authService := NewAuthService(userRepo, sessionRepo, "test-secret")

			tt.setupMocks(userRepo, sessionRepo)

			result, err := authService.Login(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Token)
				assert.NotNil(t, result.User)
			}

			userRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name          string
		request       RegisterRequest
		setupMocks    func(*MockUserRepository)
		expectedError error
	}{
		{
			name: "successful registration",
			request: RegisterRequest{
				Username: "newuser",
				Password: "password123",
				FullName: "New User",
				Email:    "newuser@example.com",
			},
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.On("GetByUsername", mock.Anything, "newuser").Return(nil, errors.New("user not found"))
				userRepo.On("GetByEmail", mock.Anything, "newuser@example.com").Return(nil, errors.New("user not found"))
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "username already exists",
			request: RegisterRequest{
				Username: "existinguser",
				Password: "password123",
				FullName: "New User",
				Email:    "newuser@example.com",
			},
			setupMocks: func(userRepo *MockUserRepository) {
				existingUser := &entities.User{ID: "user-1", Username: "existinguser"}
				userRepo.On("GetByUsername", mock.Anything, "existinguser").Return(existingUser, nil)
			},
			expectedError: ErrUserAlreadyExists,
		},
		{
			name: "email already exists",
			request: RegisterRequest{
				Username: "newuser",
				Password: "password123",
				FullName: "New User",
				Email:    "existing@example.com",
			},
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.On("GetByUsername", mock.Anything, "newuser").Return(nil, errors.New("user not found"))
				existingUser := &entities.User{ID: "user-1", Email: "existing@example.com"}
				userRepo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			expectedError: ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			sessionRepo := new(MockSessionRepository)
			authService := NewAuthService(userRepo, sessionRepo, "test-secret")

			tt.setupMocks(userRepo)

			result, err := authService.Register(context.Background(), tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Username, result.Username)
				assert.Equal(t, tt.request.FullName, result.FullName)
				assert.Equal(t, tt.request.Email, result.Email)
				assert.Equal(t, "user", result.Role)
				assert.True(t, result.IsActive)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	authService := NewAuthService(userRepo, sessionRepo, "test-secret")

	// Create a valid token
	user := &entities.User{
		ID:       "user-1",
		Username: "testuser",
		Role:     "user",
		IsActive: true,
	}

	token, _, err := authService.generateJWT(user)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		token         string
		setupMocks    func(*MockUserRepository, *MockSessionRepository)
		expectedError error
	}{
		{
			name:  "valid token",
			token: token,
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				userRepo.On("GetByID", mock.Anything, "user-1").Return(user, nil)
				session := &entities.Session{
					ID:           "session-1",
					UserID:       "user-1",
					SessionToken: token,
					ExpiresAt:    time.Now().Add(time.Hour),
				}
				sessionRepo.On("GetByToken", mock.Anything, token).Return(session, nil)
				sessionRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Session")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "invalid token format",
			token: "invalid-token",
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				// No mocks needed for invalid token
			},
			expectedError: ErrInvalidToken,
		},
		{
			name:  "user not found",
			token: token,
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				userRepo.On("GetByID", mock.Anything, "user-1").Return(nil, errors.New("user not found"))
			},
			expectedError: ErrUserNotFound,
		},
		{
			name:  "inactive user",
			token: token,
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				inactiveUser := &entities.User{
					ID:       "user-1",
					Username: "testuser",
					Role:     "user",
					IsActive: false,
				}
				userRepo.On("GetByID", mock.Anything, "user-1").Return(inactiveUser, nil)
			},
			expectedError: ErrUserInactive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo.ExpectedCalls = nil
			sessionRepo.ExpectedCalls = nil

			tt.setupMocks(userRepo, sessionRepo)

			result, err := authService.ValidateToken(context.Background(), tt.token)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, user.ID, result.ID)
				assert.Equal(t, user.Username, result.Username)
			}

			userRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	authService := NewAuthService(userRepo, sessionRepo, "test-secret")

	token := "test-token"

	sessionRepo.On("Delete", mock.Anything, token).Return(nil)

	err := authService.Logout(context.Background(), token)

	assert.NoError(t, err)
	sessionRepo.AssertExpectations(t)
}

func TestAuthService_ChangePassword(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		oldPassword   string
		newPassword   string
		setupMocks    func(*MockUserRepository, *MockSessionRepository)
		expectedError error
	}{
		{
			name:        "successful password change",
			userID:      "user-1",
			oldPassword: "oldpassword",
			newPassword: "newpassword123",
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
				user := &entities.User{
					ID:           "user-1",
					Username:     "testuser",
					PasswordHash: string(hashedPassword),
				}
				userRepo.On("GetByID", mock.Anything, "user-1").Return(user, nil)
				userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil)
				sessionRepo.On("DeleteByUserID", mock.Anything, "user-1").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "user not found",
			userID:      "nonexistent",
			oldPassword: "oldpassword",
			newPassword: "newpassword123",
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				userRepo.On("GetByID", mock.Anything, "nonexistent").Return(nil, errors.New("user not found"))
			},
			expectedError: ErrUserNotFound,
		},
		{
			name:        "invalid old password",
			userID:      "user-1",
			oldPassword: "wrongpassword",
			newPassword: "newpassword123",
			setupMocks: func(userRepo *MockUserRepository, sessionRepo *MockSessionRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
				user := &entities.User{
					ID:           "user-1",
					Username:     "testuser",
					PasswordHash: string(hashedPassword),
				}
				userRepo.On("GetByID", mock.Anything, "user-1").Return(user, nil)
			},
			expectedError: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(MockUserRepository)
			sessionRepo := new(MockSessionRepository)
			authService := NewAuthService(userRepo, sessionRepo, "test-secret")

			tt.setupMocks(userRepo, sessionRepo)

			err := authService.ChangePassword(context.Background(), tt.userID, tt.oldPassword, tt.newPassword)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_ValidatePassword(t *testing.T) {
	authService := NewAuthService(nil, nil, "test-secret")

	tests := []struct {
		name          string
		password      string
		expectedError bool
	}{
		{
			name:          "valid password",
			password:      "password123",
			expectedError: false,
		},
		{
			name:          "password too short",
			password:      "short",
			expectedError: true,
		},
		{
			name:          "empty password",
			password:      "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authService.ValidatePassword(tt.password)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_generateJWT(t *testing.T) {
	authService := NewAuthService(nil, nil, "test-secret")

	user := &entities.User{
		ID:       "user-1",
		Username: "testuser",
		Role:     "user",
	}

	tokenString, expiresAt, err := authService.generateJWT(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)
	assert.True(t, expiresAt.After(time.Now()))

	// Parse the token to verify its contents
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})

	assert.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(*JWTClaims)
	assert.True(t, ok)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Role, claims.Role)
}

func TestAuthService_CleanupExpiredSessions(t *testing.T) {
	userRepo := new(MockUserRepository)
	sessionRepo := new(MockSessionRepository)
	authService := NewAuthService(userRepo, sessionRepo, "test-secret")

	sessionRepo.On("DeleteExpired", mock.Anything).Return(nil)

	err := authService.CleanupExpiredSessions(context.Background())

	assert.NoError(t, err)
	sessionRepo.AssertExpectations(t)
}