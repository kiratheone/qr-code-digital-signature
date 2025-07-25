package routes

import (
	"digital-signature-system/internal/domain/usecases"
	"digital-signature-system/internal/infrastructure/server/handlers"
	"digital-signature-system/internal/infrastructure/server/middleware"
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

// Mock document use case
type MockDocumentUseCase struct {
	mock.Mock
}

func (m *MockDocumentUseCase) SignDocument(ctx usecases.Context, req usecases.SignDocumentRequest) (*usecases.SignDocumentResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.SignDocumentResponse), args.Error(1)
}

func (m *MockDocumentUseCase) GetDocuments(ctx usecases.Context, userID string, req usecases.GetDocumentsRequest) (*usecases.GetDocumentsResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.GetDocumentsResponse), args.Error(1)
}

func (m *MockDocumentUseCase) GetDocumentByID(ctx usecases.Context, userID, docID string) (*usecases.DocumentResponse, error) {
	args := m.Called(ctx, userID, docID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.DocumentResponse), args.Error(1)
}

func (m *MockDocumentUseCase) DeleteDocument(ctx usecases.Context, userID, docID string) error {
	args := m.Called(ctx, userID, docID)
	return args.Error(0)
}

// Mock auth use case
type MockAuthUseCase struct {
	mock.Mock
}

func (m *MockAuthUseCase) Register(ctx usecases.Context, req usecases.RegisterRequest) (*usecases.RegisterResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.RegisterResponse), args.Error(1)
}

func (m *MockAuthUseCase) Login(ctx usecases.Context, req usecases.LoginRequest) (*usecases.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.LoginResponse), args.Error(1)
}

func (m *MockAuthUseCase) Logout(ctx usecases.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthUseCase) ValidateSession(ctx usecases.Context, token string) (*usecases.SessionInfo, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.SessionInfo), args.Error(1)
}

func (m *MockAuthUseCase) RefreshToken(ctx usecases.Context, refreshToken string) (*usecases.TokenResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.TokenResponse), args.Error(1)
}

func (m *MockAuthUseCase) ChangePassword(ctx usecases.Context, req usecases.ChangePasswordRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthUseCase) GetUserByID(ctx usecases.Context, userID string) (*usecases.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.UserResponse), args.Error(1)
}

func (m *MockAuthUseCase) LogoutAll(ctx usecases.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Setup test router with authentication middleware
func setupTestRouter() (*gin.Engine, *MockDocumentUseCase, *MockAuthUseCase) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create mocks
	mockDocUseCase := new(MockDocumentUseCase)
	mockAuthUseCase := new(MockAuthUseCase)
	
	// Create handlers and middleware
	documentHandler := handlers.NewDocumentHandler(mockDocUseCase)
	authMiddleware := middleware.NewAuthMiddleware(mockAuthUseCase)
	
	// Setup routes
	api := router.Group("/api")
	SetupDocumentRoutes(api, documentHandler, authMiddleware)
	
	return router, mockDocUseCase, mockAuthUseCase
}

func TestDocumentRoutes_GetDocuments(t *testing.T) {
	// Setup
	router, mockDocUseCase, mockAuthUseCase := setupTestRouter()
	
	// Mock auth validation
	mockAuthUseCase.On("ValidateSession", mock.Anything, "valid-token").Return(&usecases.SessionInfo{
		UserID:   "user123",
		Username: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
		Role:     "user",
	}, nil)
	
	// Mock document response
	now := time.Now()
	mockResponse := &usecases.GetDocumentsResponse{
		Documents: []usecases.DocumentResponse{
			{
				ID:           "doc1",
				Filename:     "test1.pdf",
				Issuer:       "Issuer 1",
				DocumentHash: "hash1",
				CreatedAt:    now,
				UpdatedAt:    now,
				FileSize:     1024,
				Status:       "active",
			},
		},
		Total:    1,
		Page:     1,
		PageSize: 10,
	}
	
	// Set expectations
	mockDocUseCase.On("GetDocuments", mock.Anything, "user123", mock.AnythingOfType("usecases.GetDocumentsRequest")).Return(mockResponse, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/api/documents?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	documents := response["documents"].([]interface{})
	assert.Equal(t, 1, len(documents))
	assert.Equal(t, float64(1), response["total"])
	
	// Verify mocks
	mockDocUseCase.AssertExpectations(t)
	mockAuthUseCase.AssertExpectations(t)
}

func TestDocumentRoutes_GetDocumentByID(t *testing.T) {
	// Setup
	router, mockDocUseCase, mockAuthUseCase := setupTestRouter()
	
	// Mock auth validation
	mockAuthUseCase.On("ValidateSession", mock.Anything, "valid-token").Return(&usecases.SessionInfo{
		UserID:   "user123",
		Username: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
		Role:     "user",
	}, nil)
	
	// Mock document response
	now := time.Now()
	mockResponse := &usecases.DocumentResponse{
		ID:           "doc1",
		Filename:     "test1.pdf",
		Issuer:       "Issuer 1",
		DocumentHash: "hash1",
		CreatedAt:    now,
		UpdatedAt:    now,
		FileSize:     1024,
		Status:       "active",
	}
	
	// Set expectations
	mockDocUseCase.On("GetDocumentByID", mock.Anything, "user123", "doc1").Return(mockResponse, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/api/documents/doc1", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response usecases.DocumentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "doc1", response.ID)
	assert.Equal(t, "test1.pdf", response.Filename)
	
	// Verify mocks
	mockDocUseCase.AssertExpectations(t)
	mockAuthUseCase.AssertExpectations(t)
}

func TestDocumentRoutes_DeleteDocument(t *testing.T) {
	// Setup
	router, mockDocUseCase, mockAuthUseCase := setupTestRouter()
	
	// Mock auth validation
	mockAuthUseCase.On("ValidateSession", mock.Anything, "valid-token").Return(&usecases.SessionInfo{
		UserID:   "user123",
		Username: "testuser",
		FullName: "Test User",
		Email:    "test@example.com",
		Role:     "user",
	}, nil)
	
	// Set expectations
	mockDocUseCase.On("DeleteDocument", mock.Anything, "user123", "doc1").Return(nil)
	
	// Create request
	req, _ := http.NewRequest("DELETE", "/api/documents/doc1", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "Document deleted successfully", response["message"])
	
	// Verify mocks
	mockDocUseCase.AssertExpectations(t)
	mockAuthUseCase.AssertExpectations(t)
}

func TestDocumentRoutes_Unauthorized(t *testing.T) {
	// Setup
	router, _, mockAuthUseCase := setupTestRouter()
	
	// Mock auth validation
	mockAuthUseCase.On("ValidateSession", mock.Anything, "invalid-token").Return(nil, errors.New("invalid token"))
	
	// Create request
	req, _ := http.NewRequest("GET", "/api/documents", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	// Verify mocks
	mockAuthUseCase.AssertExpectations(t)
}

func TestDocumentRoutes_MissingAuth(t *testing.T) {
	// Setup
	router, _, _ := setupTestRouter()
	
	// Create request without auth header
	req, _ := http.NewRequest("GET", "/api/documents", nil)
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}