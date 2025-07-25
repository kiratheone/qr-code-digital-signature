package handlers_test

import (
	"bytes"
	"digital-signature-system/internal/domain/usecases"
	"digital-signature-system/internal/infrastructure/server/handlers"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthUseCase mocks the authentication use case
type MockAuthUseCase struct {
	mock.Mock
}

func (m *MockAuthUseCase) Register(ctx interface{}, req usecases.RegisterRequest) (*usecases.RegisterResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.RegisterResponse), args.Error(1)
}

func (m *MockAuthUseCase) Login(ctx interface{}, req usecases.LoginRequest) (*usecases.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.LoginResponse), args.Error(1)
}

func (m *MockAuthUseCase) Logout(ctx interface{}, sessionToken string) error {
	args := m.Called(ctx, sessionToken)
	return args.Error(0)
}

func (m *MockAuthUseCase) ValidateSession(ctx interface{}, sessionToken string) (*usecases.SessionInfo, error) {
	args := m.Called(ctx, sessionToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.SessionInfo), args.Error(1)
}

func (m *MockAuthUseCase) RefreshSession(ctx interface{}, refreshToken string) (*usecases.LoginResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.LoginResponse), args.Error(1)
}

func TestAuthHandler_Register(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthUseCase := new(MockAuthUseCase)
	authHandler := handlers.NewAuthHandler(mockAuthUseCase)

	// Test cases
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Valid registration",
			requestBody: map[string]interface{}{
				"username":  "testuser",
				"password":  "password123",
				"full_name": "Test User",
				"email":     "test@example.com",
			},
			setupMock: func() {
				mockAuthUseCase.On("Register", mock.Anything, mock.MatchedBy(func(req usecases.RegisterRequest) bool {
					return req.Username == "testuser" &&
						req.Password == "password123" &&
						req.FullName == "Test User" &&
						req.Email == "test@example.com"
				})).Return(&usecases.RegisterResponse{
					UserID:   "user-123",
					Username: "testuser",
					FullName: "Test User",
					Email:    "test@example.com",
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"user_id":   "user-123",
				"username":  "testuser",
				"full_name": "Test User",
				"email":     "test@example.com",
				"message":   "User registered successfully",
			},
		},
		{
			name: "Invalid registration - missing fields",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "password123",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid registration - username exists",
			requestBody: map[string]interface{}{
				"username":  "existinguser",
				"password":  "password123",
				"full_name": "Test User",
				"email":     "test@example.com",
			},
			setupMock: func() {
				mockAuthUseCase.On("Register", mock.Anything, mock.MatchedBy(func(req usecases.RegisterRequest) bool {
					return req.Username == "existinguser"
				})).Return(nil, errors.New("username already exists"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "username already exists",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock
			tc.setupMock()

			// Setup router
			router := gin.New()
			router.POST("/register", authHandler.Register)

			// Create request
			reqBody, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(resp, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, resp.Code)
			if tc.expectedBody != nil {
				var respBody map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &respBody)
				assert.NoError(t, err)
				
				// Check expected fields
				for k, v := range tc.expectedBody {
					assert.Equal(t, v, respBody[k])
				}
			}
			mockAuthUseCase.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthUseCase := new(MockAuthUseCase)
	authHandler := handlers.NewAuthHandler(mockAuthUseCase)

	// Test cases
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Valid login",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "password123",
			},
			setupMock: func() {
				expiresAt := time.Now().Add(24 * time.Hour)
				mockAuthUseCase.On("Login", mock.Anything, mock.MatchedBy(func(req usecases.LoginRequest) bool {
					return req.Username == "testuser" && req.Password == "password123"
				})).Return(&usecases.LoginResponse{
					UserID:       "user-123",
					Username:     "testuser",
					FullName:     "Test User",
					Email:        "test@example.com",
					Role:         "user",
					SessionToken: "session-token",
					RefreshToken: "refresh-token",
					ExpiresAt:    expiresAt,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"user_id":       "user-123",
				"username":      "testuser",
				"full_name":     "Test User",
				"email":         "test@example.com",
				"role":          "user",
				"token":         "session-token",
				"refresh_token": "refresh-token",
			},
		},
		{
			name: "Invalid login - missing fields",
			requestBody: map[string]interface{}{
				"username": "testuser",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid login - wrong credentials",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "wrongpassword",
			},
			setupMock: func() {
				mockAuthUseCase.On("Login", mock.Anything, mock.MatchedBy(func(req usecases.LoginRequest) bool {
					return req.Username == "testuser" && req.Password == "wrongpassword"
				})).Return(nil, errors.New("invalid username or password"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "invalid username or password",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock
			tc.setupMock()

			// Setup router
			router := gin.New()
			router.POST("/login", authHandler.Login)

			// Create request
			reqBody, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(resp, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, resp.Code)
			if tc.expectedBody != nil {
				var respBody map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &respBody)
				assert.NoError(t, err)
				
				// Check expected fields
				for k, v := range tc.expectedBody {
					assert.Equal(t, v, respBody[k])
				}
			}
			mockAuthUseCase.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthUseCase := new(MockAuthUseCase)
	authHandler := handlers.NewAuthHandler(mockAuthUseCase)

	// Test cases
	tests := []struct {
		name           string
		authHeader     string
		setupMock      func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "Valid logout",
			authHeader: "Bearer session-token",
			setupMock: func() {
				mockAuthUseCase.On("Logout", mock.Anything, "session-token").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Logged out successfully",
			},
		},
		{
			name:           "Invalid logout - missing header",
			authHeader:     "",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Authorization header is required",
			},
		},
		{
			name:       "Invalid logout - error",
			authHeader: "Bearer session-token",
			setupMock: func() {
				mockAuthUseCase.On("Logout", mock.Anything, "session-token").Return(errors.New("logout failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "logout failed",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock
			tc.setupMock()

			// Setup router
			router := gin.New()
			router.POST("/logout", authHandler.Logout)

			// Create request
			req, _ := http.NewRequest(http.MethodPost, "/logout", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			resp := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(resp, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, resp.Code)
			if tc.expectedBody != nil {
				var respBody map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &respBody)
				assert.NoError(t, err)
				
				// Check expected fields
				for k, v := range tc.expectedBody {
					assert.Equal(t, v, respBody[k])
				}
			}
			mockAuthUseCase.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthUseCase := new(MockAuthUseCase)
	authHandler := handlers.NewAuthHandler(mockAuthUseCase)

	// Test cases
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Valid refresh",
			requestBody: map[string]interface{}{
				"refresh_token": "refresh-token",
			},
			setupMock: func() {
				expiresAt := time.Now().Add(24 * time.Hour)
				mockAuthUseCase.On("RefreshSession", mock.Anything, "refresh-token").Return(&usecases.LoginResponse{
					UserID:       "user-123",
					Username:     "testuser",
					FullName:     "Test User",
					Email:        "test@example.com",
					Role:         "user",
					SessionToken: "new-session-token",
					RefreshToken: "new-refresh-token",
					ExpiresAt:    expiresAt,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"user_id":       "user-123",
				"username":      "testuser",
				"full_name":     "Test User",
				"email":         "test@example.com",
				"role":          "user",
				"token":         "new-session-token",
				"refresh_token": "new-refresh-token",
			},
		},
		{
			name: "Invalid refresh - missing token",
			requestBody: map[string]interface{}{},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid refresh - invalid token",
			requestBody: map[string]interface{}{
				"refresh_token": "invalid-token",
			},
			setupMock: func() {
				mockAuthUseCase.On("RefreshSession", mock.Anything, "invalid-token").Return(nil, errors.New("invalid refresh token"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "invalid refresh token",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock
			tc.setupMock()

			// Setup router
			router := gin.New()
			router.POST("/refresh", authHandler.RefreshToken)

			// Create request
			reqBody, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(resp, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, resp.Code)
			if tc.expectedBody != nil {
				var respBody map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &respBody)
				assert.NoError(t, err)
				
				// Check expected fields
				for k, v := range tc.expectedBody {
					assert.Equal(t, v, respBody[k])
				}
			}
			mockAuthUseCase.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Me(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthUseCase := new(MockAuthUseCase)
	authHandler := handlers.NewAuthHandler(mockAuthUseCase)

	// Test case
	t.Run("Get current user", func(t *testing.T) {
		// Setup router
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", "user-123")
			c.Set("username", "testuser")
			c.Set("full_name", "Test User")
			c.Set("email", "test@example.com")
			c.Set("role", "user")
			c.Next()
		})
		router.GET("/me", authHandler.Me)

		// Create request
		req, _ := http.NewRequest(http.MethodGet, "/me", nil)
		resp := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		var respBody map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, "user-123", respBody["user_id"])
		assert.Equal(t, "testuser", respBody["username"])
		assert.Equal(t, "Test User", respBody["full_name"])
		assert.Equal(t, "test@example.com", respBody["email"])
		assert.Equal(t, "user", respBody["role"])
	})
}

func TestAuthHandler_ChangePassword(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthUseCase := new(MockAuthUseCase)
	authHandler := handlers.NewAuthHandler(mockAuthUseCase)

	// Test case
	t.Run("Change password", func(t *testing.T) {
		// Setup router
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", "user-123")
			c.Next()
		})
		router.POST("/change-password", authHandler.ChangePassword)

		// Create request
		reqBody := map[string]interface{}{
			"current_password": "oldpassword",
			"new_password":     "newpassword123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/change-password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(resp, req)

		// Assert - this should return not implemented for now
		assert.Equal(t, http.StatusNotImplemented, resp.Code)
	})

	t.Run("Change password - missing fields", func(t *testing.T) {
		// Setup router
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", "user-123")
			c.Next()
		})
		router.POST("/change-password", authHandler.ChangePassword)

		// Create request with missing fields
		reqBody := map[string]interface{}{
			"current_password": "oldpassword",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/change-password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}

func TestAuthHandler_LogoutAll(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthUseCase := new(MockAuthUseCase)
	authHandler := handlers.NewAuthHandler(mockAuthUseCase)

	// Test case
	t.Run("Logout all sessions", func(t *testing.T) {
		// Setup router
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", "user-123")
			c.Next()
		})
		router.POST("/logout-all", authHandler.LogoutAll)

		// Create request
		req, _ := http.NewRequest(http.MethodPost, "/logout-all", nil)
		resp := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(resp, req)

		// Assert - this should return not implemented for now
		assert.Equal(t, http.StatusNotImplemented, resp.Code)
	})
}

func TestAuthHandler_ValidateToken(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthUseCase := new(MockAuthUseCase)
	authHandler := handlers.NewAuthHandler(mockAuthUseCase)

	// Test cases
	tests := []struct {
		name           string
		token          string
		setupMock      func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:  "Valid token",
			token: "valid-token",
			setupMock: func() {
				mockAuthUseCase.On("ValidateSession", mock.Anything, "valid-token").Return(&usecases.SessionInfo{
					UserID:   "user-123",
					Username: "testuser",
					FullName: "Test User",
					Email:    "test@example.com",
					Role:     "user",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"valid":     true,
				"user_id":   "user-123",
				"username":  "testuser",
				"full_name": "Test User",
				"email":     "test@example.com",
				"role":      "user",
			},
		},
		{
			name:           "Missing token",
			token:          "",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Token parameter is required",
			},
		},
		{
			name:  "Invalid token",
			token: "invalid-token",
			setupMock: func() {
				mockAuthUseCase.On("ValidateSession", mock.Anything, "invalid-token").Return(nil, errors.New("invalid token"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "Invalid or expired token",
				"valid": false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock
			tc.setupMock()

			// Setup router
			router := gin.New()
			router.GET("/validate", authHandler.ValidateToken)

			// Create request
			url := "/validate"
			if tc.token != "" {
				url += "?token=" + tc.token
			}
			req, _ := http.NewRequest(http.MethodGet, url, nil)
			resp := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(resp, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, resp.Code)
			if tc.expectedBody != nil {
				var respBody map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &respBody)
				assert.NoError(t, err)
				
				// Check expected fields
				for k, v := range tc.expectedBody {
					assert.Equal(t, v, respBody[k])
				}
			}
			mockAuthUseCase.AssertExpectations(t)
		})
	}
}