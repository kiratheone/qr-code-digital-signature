package handlers_test

import (
	"bytes"
	"digital-signature-system/internal/domain/usecases"
	"digital-signature-system/internal/infrastructure/server/handlers"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock verification use case
type MockVerificationUseCase struct {
	mock.Mock
}

func (m *MockVerificationUseCase) GetVerificationInfo(ctx interface{}, docID string) (*usecases.VerificationInfo, error) {
	args := m.Called(ctx, docID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.VerificationInfo), args.Error(1)
}

func (m *MockVerificationUseCase) VerifyDocument(ctx interface{}, docID string, documentData []byte, verifierIP string) (*usecases.VerificationResult, error) {
	args := m.Called(ctx, docID, documentData, verifierIP)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecases.VerificationResult), args.Error(1)
}

func TestGetVerificationInfo(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockUC := new(MockVerificationUseCase)
	handler := handlers.NewVerificationHandler(mockUC)

	// Test case: Valid document ID
	t.Run("Valid document ID", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		docID := "valid-doc-id"
		c.Params = []gin.Param{{Key: "docId", Value: docID}}
		
		info := &usecases.VerificationInfo{
			DocumentID: docID,
			Filename:   "test.pdf",
			Issuer:     "Test Issuer",
			CreatedAt:  time.Now(),
		}
		
		mockUC.On("GetVerificationInfo", mock.Anything, docID).Return(info, nil)
		
		// Execute
		handler.GetVerificationInfo(c)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response usecases.VerificationInfo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, docID, response.DocumentID)
		assert.Equal(t, "test.pdf", response.Filename)
		assert.Equal(t, "Test Issuer", response.Issuer)
		
		mockUC.AssertExpectations(t)
	})

	// Test case: Document not found
	t.Run("Document not found", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		docID := "non-existent-id"
		c.Params = []gin.Param{{Key: "docId", Value: docID}}
		
		mockUC.On("GetVerificationInfo", mock.Anything, docID).Return(nil, errors.New("document not found"))
		
		// Execute
		handler.GetVerificationInfo(c)
		
		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "not found")
		
		mockUC.AssertExpectations(t)
	})

	// Test case: Missing document ID
	t.Run("Missing document ID", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		c.Params = []gin.Param{{Key: "docId", Value: ""}}
		
		// Execute
		handler.GetVerificationInfo(c)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "required")
	})

	// Test case: Internal server error
	t.Run("Internal server error", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		docID := "error-doc-id"
		c.Params = []gin.Param{{Key: "docId", Value: docID}}
		
		mockUC.On("GetVerificationInfo", mock.Anything, docID).Return(nil, errors.New("database error"))
		
		// Execute
		handler.GetVerificationInfo(c)
		
		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		mockUC.AssertExpectations(t)
	})
}

func TestVerifyDocument(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockUC := new(MockVerificationUseCase)
	handler := handlers.NewVerificationHandler(mockUC)

	// Helper function to create multipart form with file
	createFormFile := func(fieldName, filename string, content []byte) (*bytes.Buffer, string) {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		fw, _ := w.CreateFormFile(fieldName, filename)
		fw.Write(content)
		w.Close()
		return &b, w.FormDataContentType()
	}

	// Test case: Valid verification
	t.Run("Valid verification", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		docID := "valid-doc-id"
		c.Params = []gin.Param{{Key: "docId", Value: docID}}
		
		// Create test PDF data
		pdfData := []byte("%PDF-1.5\nTest PDF content")
		
		// Create form with file
		body, contentType := createFormFile("document", "test.pdf", pdfData)
		
		// Setup request
		req, _ := http.NewRequest("POST", "/api/verify/"+docID+"/upload", body)
		req.Header.Set("Content-Type", contentType)
		c.Request = req
		
		// Mock client IP
		c.Request.RemoteAddr = "192.168.1.1:1234"
		
		// Setup mock response
		result := &usecases.VerificationResult{
			Status:           "valid",
			Message:          "✅ Document is valid",
			DocumentID:       docID,
			Issuer:           "Test Issuer",
			CreatedAt:        time.Now(),
			VerificationTime: time.Now(),
			Details:          "The document is authentic and has not been modified",
		}
		
		mockUC.On("VerifyDocument", mock.Anything, docID, pdfData, mock.Anything).Return(result, nil)
		
		// Execute
		handler.VerifyDocument(c)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response usecases.VerificationResult
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "valid", response.Status)
		assert.Equal(t, "✅ Document is valid", response.Message)
		assert.Equal(t, docID, response.DocumentID)
		
		mockUC.AssertExpectations(t)
	})

	// Test case: Document not found
	t.Run("Document not found", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		docID := "non-existent-id"
		c.Params = []gin.Param{{Key: "docId", Value: docID}}
		
		// Create test PDF data
		pdfData := []byte("%PDF-1.5\nTest PDF content")
		
		// Create form with file
		body, contentType := createFormFile("document", "test.pdf", pdfData)
		
		// Setup request
		req, _ := http.NewRequest("POST", "/api/verify/"+docID+"/upload", body)
		req.Header.Set("Content-Type", contentType)
		c.Request = req
		
		// Mock client IP
		c.Request.RemoteAddr = "192.168.1.1:1234"
		
		mockUC.On("VerifyDocument", mock.Anything, docID, pdfData, mock.Anything).Return(nil, errors.New("document not found"))
		
		// Execute
		handler.VerifyDocument(c)
		
		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		mockUC.AssertExpectations(t)
	})

	// Test case: Invalid file type
	t.Run("Invalid file type", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		docID := "valid-doc-id"
		c.Params = []gin.Param{{Key: "docId", Value: docID}}
		
		// Create test non-PDF data
		data := []byte("This is not a PDF file")
		
		// Create form with file
		body, contentType := createFormFile("document", "test.txt", data)
		
		// Setup request
		req, _ := http.NewRequest("POST", "/api/verify/"+docID+"/upload", body)
		req.Header.Set("Content-Type", contentType)
		c.Request = req
		
		// Execute
		handler.VerifyDocument(c)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Only PDF files are supported")
	})

	// Test case: Missing file
	t.Run("Missing file", func(t *testing.T) {
		// Setup
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		docID := "valid-doc-id"
		c.Params = []gin.Param{{Key: "docId", Value: docID}}
		
		// Setup request without file
		req, _ := http.NewRequest("POST", "/api/verify/"+docID+"/upload", strings.NewReader(""))
		req.Header.Set("Content-Type", "multipart/form-data")
		c.Request = req
		
		// Execute
		handler.VerifyDocument(c)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "required")
	})
}