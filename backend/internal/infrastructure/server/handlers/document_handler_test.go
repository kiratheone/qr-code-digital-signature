package handlers

import (
	"bytes"
	"digital-signature-system/internal/domain/usecases"
	"digital-signature-system/internal/infrastructure/server/middleware"
	"encoding/json"
	"errors"
	"mime/multipart"
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

// Setup test router with authentication middleware
func setupTestRouter(mockUseCase *MockDocumentUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create document handler
	handler := NewDocumentHandler(mockUseCase)
	
	// Add test middleware to set user ID in context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test_user_id")
		c.Set("username", "testuser")
		c.Set("full_name", "Test User")
		c.Set("email", "test@example.com")
		c.Set("role", "user")
		c.Next()
	})
	
	// Setup routes
	router.POST("/api/documents/sign", handler.SignDocument)
	router.GET("/api/documents", handler.GetDocuments)
	router.GET("/api/documents/:id", handler.GetDocumentByID)
	router.DELETE("/api/documents/:id", handler.DeleteDocument)
	
	return router
}

func TestGetDocuments(t *testing.T) {
	// Create mock use case
	mockUseCase := new(MockDocumentUseCase)
	
	// Setup test router
	router := setupTestRouter(mockUseCase)
	
	// Mock response
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
			{
				ID:           "doc2",
				Filename:     "test2.pdf",
				Issuer:       "Issuer 2",
				DocumentHash: "hash2",
				CreatedAt:    now,
				UpdatedAt:    now,
				FileSize:     2048,
				Status:       "active",
			},
		},
		Total:    2,
		Page:     1,
		PageSize: 10,
	}
	
	// Set expectations
	mockUseCase.On("GetDocuments", mock.Anything, "test_user_id", mock.AnythingOfType("usecases.GetDocumentsRequest")).Return(mockResponse, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/api/documents?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	documents := response["documents"].([]interface{})
	assert.Equal(t, 2, len(documents))
	assert.Equal(t, float64(2), response["total"])
	assert.Equal(t, float64(1), response["page"])
	assert.Equal(t, float64(10), response["page_size"])
	
	// Verify mock
	mockUseCase.AssertExpectations(t)
}

func TestGetDocumentByID(t *testing.T) {
	// Create mock use case
	mockUseCase := new(MockDocumentUseCase)
	
	// Setup test router
	router := setupTestRouter(mockUseCase)
	
	// Mock response
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
	mockUseCase.On("GetDocumentByID", mock.Anything, "test_user_id", "doc1").Return(mockResponse, nil)
	
	// Create request
	req, _ := http.NewRequest("GET", "/api/documents/doc1", nil)
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
	assert.Equal(t, "Issuer 1", response.Issuer)
	
	// Verify mock
	mockUseCase.AssertExpectations(t)
}

func TestGetDocumentByID_NotFound(t *testing.T) {
	// Create mock use case
	mockUseCase := new(MockDocumentUseCase)
	
	// Setup test router
	router := setupTestRouter(mockUseCase)
	
	// Set expectations
	mockUseCase.On("GetDocumentByID", mock.Anything, "test_user_id", "doc1").Return(nil, errors.New("document not found"))
	
	// Create request
	req, _ := http.NewRequest("GET", "/api/documents/doc1", nil)
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response["error"], "Document not found")
	
	// Verify mock
	mockUseCase.AssertExpectations(t)
}

func TestDeleteDocument(t *testing.T) {
	// Create mock use case
	mockUseCase := new(MockDocumentUseCase)
	
	// Setup test router
	router := setupTestRouter(mockUseCase)
	
	// Set expectations
	mockUseCase.On("DeleteDocument", mock.Anything, "test_user_id", "doc1").Return(nil)
	
	// Create request
	req, _ := http.NewRequest("DELETE", "/api/documents/doc1", nil)
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "Document deleted successfully", response["message"])
	
	// Verify mock
	mockUseCase.AssertExpectations(t)
}

func TestDeleteDocument_NotFound(t *testing.T) {
	// Create mock use case
	mockUseCase := new(MockDocumentUseCase)
	
	// Setup test router
	router := setupTestRouter(mockUseCase)
	
	// Set expectations
	mockUseCase.On("DeleteDocument", mock.Anything, "test_user_id", "doc1").Return(errors.New("document not found"))
	
	// Create request
	req, _ := http.NewRequest("DELETE", "/api/documents/doc1", nil)
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response["error"], "Failed to delete document")
	
	// Verify mock
	mockUseCase.AssertExpectations(t)
}

func TestSignDocument(t *testing.T) {
	// Create mock use case
	mockUseCase := new(MockDocumentUseCase)
	
	// Setup test router
	router := setupTestRouter(mockUseCase)
	
	// Mock response
	now := time.Now()
	mockResponse := &usecases.SignDocumentResponse{
		DocumentID:   "doc1",
		Filename:     "test.pdf",
		Issuer:       "Test Issuer",
		DocumentHash: "hash1",
		SignedPDF:    []byte("signed pdf data"),
		QRCodeData:   "qr code data",
		CreatedAt:    now,
	}
	
	// Set expectations
	mockUseCase.On(
		"SignDocument",
		mock.Anything,
		mock.MatchedBy(func(req usecases.SignDocumentRequest) bool {
			return req.UserID == "test_user_id" &&
				req.Filename == "test.pdf" &&
				req.Issuer == "Test Issuer" &&
				len(req.PDFData) > 0
		}),
	).Return(mockResponse, nil)
	
	// Create multipart form
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	
	// Add form fields
	_ = writer.WriteField("filename", "test.pdf")
	_ = writer.WriteField("issuer", "Test Issuer")
	
	// Add PDF file
	part, _ := writer.CreateFormFile("pdf", "test.pdf")
	part.Write([]byte("fake pdf data"))
	
	writer.Close()
	
	// Create request
	req, _ := http.NewRequest("POST", "/api/documents/sign", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Equal(t, `attachment; filename="signed_test.pdf"`, w.Header().Get("Content-Disposition"))
	assert.Equal(t, mockResponse.SignedPDF, w.Body.Bytes())
	
	// Verify mock
	mockUseCase.AssertExpectations(t)
}

func TestSignDocument_ValidationError(t *testing.T) {
	// Create mock use case
	mockUseCase := new(MockDocumentUseCase)
	
	// Setup test router
	router := setupTestRouter(mockUseCase)
	
	// Create multipart form with missing fields
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	
	// Add PDF file but missing required fields
	part, _ := writer.CreateFormFile("pdf", "test.pdf")
	part.Write([]byte("fake pdf data"))
	
	writer.Close()
	
	// Create request
	req, _ := http.NewRequest("POST", "/api/documents/sign", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response["error"], "Invalid form data")
}

func TestSignDocument_ProcessingError(t *testing.T) {
	// Create mock use case
	mockUseCase := new(MockDocumentUseCase)
	
	// Setup test router
	router := setupTestRouter(mockUseCase)
	
	// Set expectations
	mockUseCase.On(
		"SignDocument",
		mock.Anything,
		mock.MatchedBy(func(req usecases.SignDocumentRequest) bool {
			return req.UserID == "test_user_id" &&
				req.Filename == "test.pdf" &&
				req.Issuer == "Test Issuer"
		}),
	).Return(nil, errors.New("failed to sign document"))
	
	// Create multipart form
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	
	// Add form fields
	_ = writer.WriteField("filename", "test.pdf")
	_ = writer.WriteField("issuer", "Test Issuer")
	
	// Add PDF file
	part, _ := writer.CreateFormFile("pdf", "test.pdf")
	part.Write([]byte("fake pdf data"))
	
	writer.Close()
	
	// Create request
	req, _ := http.NewRequest("POST", "/api/documents/sign", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response["error"], "Failed to sign document")
	
	// Verify mock
	mockUseCase.AssertExpectations(t)
}

func TestNoAuthentication(t *testing.T) {
	// Create mock use case
	mockUseCase := new(MockDocumentUseCase)
	
	// Create router without authentication middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create document handler
	handler := NewDocumentHandler(mockUseCase)
	
	// Setup routes
	router.GET("/api/documents", handler.GetDocuments)
	
	// Create request
	req, _ := http.NewRequest("GET", "/api/documents", nil)
	w := httptest.NewRecorder()
	
	// Perform request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Contains(t, response["error"], "Authentication required")
}