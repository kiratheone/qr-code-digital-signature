package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"digital-signature-system/internal/config"
	"digital-signature-system/internal/domain/services"
	"digital-signature-system/internal/infrastructure/database"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create simplified tables for testing
	err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			full_name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			role TEXT DEFAULT 'user',
			created_at DATETIME,
			updated_at DATETIME,
			is_active BOOLEAN DEFAULT true
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			session_token TEXT UNIQUE NOT NULL,
			refresh_token TEXT UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME,
			last_accessed DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`).Error
	require.NoError(t, err)

	return db
}

func setupTestServer(t *testing.T) (*Server, *gorm.DB) {
	db := setupTestDB(t)
	cfg := &config.Config{
		JWTSecret:   "test-secret-key",
		Environment: "test",
	}

	server := NewServer(cfg, db)
	return server, db
}

func TestAuthHandler_Register(t *testing.T) {
	server, _ := setupTestServer(t)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful registration",
			requestBody: RegisterRequest{
				Username: "testuser",
				Password: "Password123!",
				FullName: "Test User",
				Email:    "test@example.com",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request - missing username",
			requestBody: RegisterRequest{
				Password: "Password123!",
				FullName: "Test User",
				Email:    "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  ErrCodeValidationFailed,
		},
		{
			name: "invalid request - short password",
			requestBody: RegisterRequest{
				Username: "testuser",
				Password: "short",
				FullName: "Test User",
				Email:    "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  ErrCodeValidationFailed,
		},
		{
			name: "invalid request - invalid email",
			requestBody: RegisterRequest{
				Username: "testuser",
				Password: "Password123!",
				FullName: "Test User",
				Email:    "invalid-email",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  ErrCodeValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response StandardError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Code)
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	server, db := setupTestServer(t)

	// Create a test user
	userRepo := database.NewUserRepository(db)
	sessionRepo := database.NewSessionRepository(db)
	authService := services.NewAuthService(userRepo, sessionRepo, "test-secret-key")

	registerReq := services.RegisterRequest{
		Username: "testuser",
		Password: "Password123!",
		FullName: "Test User",
		Email:    "test@example.com",
	}
	_, err := authService.Register(context.Background(), registerReq)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    LoginRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful login",
			requestBody: LoginRequest{
				Username: "testuser",
				Password: "Password123!",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid username",
			requestBody: LoginRequest{
				Username: "nonexistent",
				Password: "Password123!",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrCodeUnauthorized,
		},
		{
			name: "invalid password",
			requestBody: LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrCodeUnauthorized,
		},
		{
			name: "missing username",
			requestBody: LoginRequest{
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  ErrCodeValidationFailed,
		},
		{
			name: "missing password",
			requestBody: LoginRequest{
				Username: "testuser",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  ErrCodeValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response StandardError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Code)
			} else if tt.expectedStatus == http.StatusOK {
				var response services.LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.NotNil(t, response.User)
			}
		})
	}
}

func TestAuthHandler_GetProfile(t *testing.T) {
	server, db := setupTestServer(t)

	// Create a test user and get a token
	userRepo := database.NewUserRepository(db)
	sessionRepo := database.NewSessionRepository(db)
	authService := services.NewAuthService(userRepo, sessionRepo, "test-secret-key")

	registerReq := services.RegisterRequest{
		Username: "testuser",
		Password: "Password123!",
		FullName: "Test User",
		Email:    "test@example.com",
	}
	user, err := authService.Register(context.Background(), registerReq)
	require.NoError(t, err)

	loginReq := services.LoginRequest{
		Username: "testuser",
		Password: "Password123!",
	}
	loginResp, err := authService.Login(context.Background(), loginReq)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful profile retrieval",
			token:          loginResp.Token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrCodeUnauthorized,
		},
		{
			name:           "invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrCodeUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/profile", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response StandardError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Code)
			} else if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "user")

				userMap := response["user"].(map[string]interface{})
				assert.Equal(t, user.Username, userMap["username"])
				assert.Equal(t, user.Email, userMap["email"])
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	server, db := setupTestServer(t)

	// Create a test user and get a token
	userRepo := database.NewUserRepository(db)
	sessionRepo := database.NewSessionRepository(db)
	authService := services.NewAuthService(userRepo, sessionRepo, "test-secret-key")

	registerReq := services.RegisterRequest{
		Username: "testuser",
		Password: "Password123!",
		FullName: "Test User",
		Email:    "test@example.com",
	}
	_, err := authService.Register(context.Background(), registerReq)
	require.NoError(t, err)

	loginReq := services.LoginRequest{
		Username: "testuser",
		Password: "Password123!",
	}
	loginResp, err := authService.Login(context.Background(), loginReq)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "successful logout",
			token:          loginResp.Token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			token:          "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  ErrCodeValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/auth/logout", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response StandardError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Code)
			}
		})
	}
}

func TestAuthHandler_ChangePassword(t *testing.T) {
	server, db := setupTestServer(t)

	// Create a test user and get a token
	userRepo := database.NewUserRepository(db)
	sessionRepo := database.NewSessionRepository(db)
	authService := services.NewAuthService(userRepo, sessionRepo, "test-secret-key")

	registerReq := services.RegisterRequest{
		Username: "testuser",
		Password: "Password123!",
		FullName: "Test User",
		Email:    "test@example.com",
	}
	_, err := authService.Register(context.Background(), registerReq)
	require.NoError(t, err)

	loginReq := services.LoginRequest{
		Username: "testuser",
		Password: "Password123!",
	}
	loginResp, err := authService.Login(context.Background(), loginReq)
	require.NoError(t, err)

	tests := []struct {
		name           string
		token          string
		requestBody    ChangePasswordRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name:  "successful password change",
			token: loginResp.Token,
			requestBody: ChangePasswordRequest{
				OldPassword: "Password123!",
				NewPassword: "NewPassword123!",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "invalid old password",
			token: loginResp.Token,
			requestBody: ChangePasswordRequest{
				OldPassword: "wrongpassword",
				NewPassword: "NewPassword123!",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrCodeUnauthorized,
		},
		{
			name:  "weak new password",
			token: loginResp.Token,
			requestBody: ChangePasswordRequest{
				OldPassword: "Password123!",
				NewPassword: "weak",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  ErrCodeValidationFailed,
		},
		{
			name:  "missing token",
			token: "",
			requestBody: ChangePasswordRequest{
				OldPassword: "password123",
				NewPassword: "newpassword123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrCodeUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/change-password", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response StandardError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Code)
			}
		})
	}
}

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	server, db := setupTestServer(t)

	// Create a test user and get a token
	userRepo := database.NewUserRepository(db)
	sessionRepo := database.NewSessionRepository(db)
	authService := services.NewAuthService(userRepo, sessionRepo, "test-secret-key")

	registerReq := services.RegisterRequest{
		Username: "testuser",
		Password: "Password123!",
		FullName: "Test User",
		Email:    "test@example.com",
	}
	_, err := authService.Register(context.Background(), registerReq)
	require.NoError(t, err)

	loginReq := services.LoginRequest{
		Username: "testuser",
		Password: "Password123!",
	}
	loginResp, err := authService.Login(context.Background(), loginReq)
	require.NoError(t, err)

	// Add a test protected route
	server.router.GET("/test-protected", server.authMiddleware.RequireAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "protected"})
	})

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "valid token",
			token:          loginResp.Token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrCodeUnauthorized,
		},
		{
			name:           "invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  ErrCodeUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test-protected", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response StandardError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Code)
			}
		})
	}
}
