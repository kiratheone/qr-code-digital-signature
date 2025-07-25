package routes_test

import (
	"bytes"
	"digital-signature-system/internal/domain/usecases"
	"digital-signature-system/internal/infrastructure/di"
	"digital-signature-system/internal/infrastructure/server/handlers"
	"digital-signature-system/internal/infrastructure/server/routes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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

// Mock container
type MockContainer struct {
	mock.Mock
}

func (m *MockContainer) VerificationUseCase() usecases.VerificationUseCase {
	args := m.Called()
	return args.Get(0).(usecases.VerificationUseCase)
}

func TestVerificationRoutes(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	
	// Create mock use case and container
	mockUC := new(MockVerificationUseCase)
	mockContainer := new(MockContainer)
	mockContainer.On("VerificationUseCase").Return(mockUC)
	
	// Create router
	router := gin.New()
	
	// Setup routes
	routes.SetupVerificationRoutes(router, mockContainer)
	
	// Test case: GET /api/verify/:docId
	t.Run("GET /api/verify/:docId", func(t *testing.T) {
		// Setup
		docID := "test-doc-id"
		info := &usecases.VerificationInfo{
			DocumentID: docID,
			Filename:   "test.pdf",
			Issuer:     "Test Issuer",
			CreatedAt:  time.Now(),
		}
		
		mockUC.On("GetVerificationInfo", mock.Anything, docID).Return(info, nil)
		
		// Create request
		req, _ := http.NewRequest("GET", "/api/verify/"+docID, nil)
		w := httptest.NewRecorder()
		
		// Execute
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})
	
	// Test case: POST /api/verify/:docId/upload
	t.Run("POST /api/verify/:docId/upload", func(t *testing.T) {
		// Setup
		docID := "test-doc-id"
		
		// Create test PDF data
		pdfData := []byte("%PDF-1.5\nTest PDF content")
		
		// Create form with file
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		fw, _ := w.CreateFormFile("document", "test.pdf")
		fw.Write(pdfData)
		w.Close()
		
		// Create result
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
		
		// Create request
		req, _ := http.NewRequest("POST", "/api/verify/"+docID+"/upload", &b)
		req.Header.Set("Content-Type", w.FormDataContentType())
		resp := httptest.NewRecorder()
		
		// Execute
		router.ServeHTTP(resp, req)
		
		// Assert
		assert.Equal(t, http.StatusOK, resp.Code)
		mockUC.AssertExpectations(t)
	})
}