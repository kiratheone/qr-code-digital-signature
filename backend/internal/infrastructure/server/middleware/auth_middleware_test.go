package middleware_test

import (
	"digital-signature-system/internal/domain/usecases"
	"digital-signature-system/internal/infrastructure/server/middleware"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestAuthMiddleware_Authenticate(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthUseCase := new(MockAuthUseCase)
	authMiddleware := middleware.NewAuthMiddleware(mockAuthUseCase)

	// Test cases
	tests := []struct {
		name           string
		authHeader     string
		setupMock      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "Valid token",
			authHeader: "Bearer valid-token",
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
			expectedBody:   `{"message":"success"}`,
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Authorization header is required"}`,
		},
		{
			name:           "Invalid authorization header format",
			authHeader:     "InvalidFormat",
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid authorization header format"}`,
		},
		{
			name:       "Invalid token",
			authHeader: "Bearer invalid-token",
			setupMock: func() {
				mockAuthUseCase.On("ValidateSession", mock.Anything, "invalid-token").Return(nil, errors.New("invalid token"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid or expired token"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock
			tc.setupMock()

			// Setup router
			router := gin.New()
			router.Use(authMiddleware.Authenticate())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create request
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			resp := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(resp, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, resp.Code)
			mockAuthUseCase.AssertExpectations(t)
		})
	}
}

func TestAuthMiddleware_RequireRole(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthUseCase := new(MockAuthUseCase)
	authMiddleware := middleware.NewAuthMiddleware(mockAuthUseCase)

	// Test cases
	tests := []struct {
		name           string
		role           string
		setupContext   func(c *gin.Context)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "User has required role",
			role: "admin",
			setupContext: func(c *gin.Context) {
				c.Set("role", "admin")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"success"}`,
		},
		{
			name: "User does not have required role",
			role: "admin",
			setupContext: func(c *gin.Context) {
				c.Set("role", "user")
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"Insufficient permissions"}`,
		},
		{
			name:           "No role in context",
			role:           "admin",
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Authentication required"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup router
			router := gin.New()
			router.Use(func(c *gin.Context) {
				tc.setupContext(c)
				c.Next()
			})
			router.Use(authMiddleware.RequireRole(tc.role))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create request
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			resp := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(resp, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, resp.Code)
		})
	}
}

func TestAuthMiddleware_RequireAnyRole(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthUseCase := new(MockAuthUseCase)
	authMiddleware := middleware.NewAuthMiddleware(mockAuthUseCase)

	// Test cases
	tests := []struct {
		name           string
		roles          []string
		setupContext   func(c *gin.Context)
		expectedStatus int
	}{
		{
			name:  "User has one of the required roles",
			roles: []string{"admin", "manager"},
			setupContext: func(c *gin.Context) {
				c.Set("role", "admin")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "User has another of the required roles",
			roles: []string{"admin", "manager"},
			setupContext: func(c *gin.Context) {
				c.Set("role", "manager")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "User does not have any of the required roles",
			roles: []string{"admin", "manager"},
			setupContext: func(c *gin.Context) {
				c.Set("role", "user")
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "No role in context",
			roles:          []string{"admin", "manager"},
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup router
			router := gin.New()
			router.Use(func(c *gin.Context) {
				tc.setupContext(c)
				c.Next()
			})
			router.Use(authMiddleware.RequireAnyRole(tc.roles...))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create request
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			resp := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(resp, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, resp.Code)
		})
	}
}

func TestAuthMiddleware_RateLimiter(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthUseCase := new(MockAuthUseCase)
	authMiddleware := middleware.NewAuthMiddleware(mockAuthUseCase)

	// Test rate limiter
	t.Run("Rate limiter allows requests under limit", func(t *testing.T) {
		// Setup router with rate limiter (3 requests per second)
		router := gin.New()
		router.Use(authMiddleware.RateLimiter(3, time.Second))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// Make 3 requests (should all succeed)
		for i := 0; i < 3; i++ {
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			assert.Equal(t, http.StatusOK, resp.Code)
		}

		// Make 4th request (should be rate limited)
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusTooManyRequests, resp.Code)
	})
}

func TestAuthMiddleware_CORS(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockAuthUseCase := new(MockAuthUseCase)
	authMiddleware := middleware.NewAuthMiddleware(mockAuthUseCase)

	// Test CORS middleware
	t.Run("CORS middleware sets headers", func(t *testing.T) {
		// Setup router with CORS
		router := gin.New()
		router.Use(authMiddleware.CORS())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// Make request
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "*", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", resp.Header().Get("Access-Control-Allow-Credentials"))
		assert.Contains(t, resp.Header().Get("Access-Control-Allow-Headers"), "Authorization")
		assert.Contains(t, resp.Header().Get("Access-Control-Allow-Methods"), "GET")
	})

	t.Run("CORS middleware handles OPTIONS request", func(t *testing.T) {
		// Setup router with CORS
		router := gin.New()
		router.Use(authMiddleware.CORS())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// Make OPTIONS request
		req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		// Assert
		assert.Equal(t, http.StatusNoContent, resp.Code)
	})
}