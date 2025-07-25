package server

import (
	"bytes"
	"context"
	"digital-signature-system/internal/config"
	"digital-signature-system/internal/infrastructure/database"
	"digital-signature-system/internal/infrastructure/di"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntegrationTestSuite holds the test suite setup
type IntegrationTestSuite struct {
	server    *gin.Engine
	container *di.Container
	testDB    string
}

// SetupIntegrationTest sets up the integration test environment
func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create test database configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            "5432",
			User:            "postgres",
			Password:        "password",
			DBName:          "digital_signature_test",
			SSLMode:         "disable",
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 60,
			ConnMaxIdleTime: 30,
		},
		Security: config.SecurityConfig{
			PrivateKey:      "test-private-key",
			PublicKey:       "test-public-key",
			JWTSecret:       "test-jwt-secret",
			KeyRotationDays: 90,
		},
		Server: config.ServerConfig{
			Port: "8080",
			Host: "localhost",
		},
	}

	// Connect to test database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		t.Skipf("Database connection failed: %v", err)
	}

	// Run migrations
	err = database.Migrate(db)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Create DI container
	container := di.NewContainer(cfg, db)

	// Create server
	server := NewServer(container)

	return &IntegrationTestSuite{
		server:    server,
		container: container,
		testDB:    cfg.Database.DBName,
	}
}

// TeardownIntegrationTest cleans up the test environment
func (suite *IntegrationTestSuite) TeardownIntegrationTest(t *testing.T) {
	// Clean up database
	if suite.container != nil && suite.container.DB() != nil {
		// Drop test tables
		db := suite.container.DB()
		db.Exec("DROP TABLE IF EXISTS verification_logs CASCADE")
		db.Exec("DROP TABLE IF EXISTS documents CASCADE")
		db.Exec("DROP TABLE IF EXISTS sessions CASCADE")
		db.Exec("DROP TABLE IF EXISTS users CASCADE")
		
		// Close container
		suite.container.Close()
	}
}

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.TeardownIntegrationTest(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	suite.server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "healthy", response["status"])
	assert.Contains(t, response, "timestamp")
}

// TestDocumentSigningFlow tests the complete document signing flow
func TestDocumentSigningFlow(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.TeardownIntegrationTest(t)

	// First, create a test user and get auth token
	authToken := suite.createTestUserAndLogin(t)

	// Create a test PDF file
	pdfContent := createTestPDF(t)

	// Test document signing
	t.Run("SignDocument", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add file
		part, err := writer.CreateFormFile("document", "test.pdf")
		require.NoError(t, err)
		_, err = part.Write(pdfContent)
		require.NoError(t, err)

		// Add issuer
		err = writer.WriteField("issuer", "Test Issuer")
		require.NoError(t, err)

		writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/documents/sign", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+authToken)
		suite.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "document_id")
		assert.Contains(t, response, "qr_code")
		assert.Contains(t, response, "download_url")
	})
}

// TestDocumentManagementFlow tests document management operations
func TestDocumentManagementFlow(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.TeardownIntegrationTest(t)

	authToken := suite.createTestUserAndLogin(t)
	docID := suite.createTestDocument(t, authToken)

	// Test getting documents list
	t.Run("GetDocuments", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/documents?page=1&limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		suite.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "documents")
		assert.Contains(t, response, "total")
		assert.Contains(t, response, "page")
	})

	// Test getting single document
	t.Run("GetDocumentByID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/documents/"+docID, nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		suite.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, docID, response["id"])
	})

	// Test deleting document
	t.Run("DeleteDocument", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/documents/"+docID, nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		suite.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestVerificationFlow tests document verification flow
func TestVerificationFlow(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.TeardownIntegrationTest(t)

	authToken := suite.createTestUserAndLogin(t)
	docID := suite.createTestDocument(t, authToken)

	// Test getting verification info
	t.Run("GetVerificationInfo", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/verify/"+docID, nil)
		suite.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "filename")
		assert.Contains(t, response, "issuer")
		assert.Contains(t, response, "created_at")
	})

	// Test document verification
	t.Run("VerifyDocument", func(t *testing.T) {
		pdfContent := createTestPDF(t)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("document", "test.pdf")
		require.NoError(t, err)
		_, err = part.Write(pdfContent)
		require.NoError(t, err)

		writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/verify/"+docID+"/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		suite.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "status")
		assert.Contains(t, response, "message")
	})
}

// TestAuthenticationFlow tests authentication endpoints
func TestAuthenticationFlow(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.TeardownIntegrationTest(t)

	// Test user registration (if implemented)
	t.Run("Register", func(t *testing.T) {
		payload := map[string]string{
			"username": "testuser",
			"password": "testpass123",
			"email":    "test@example.com",
			"fullName": "Test User",
		}

		jsonPayload, _ := json.Marshal(payload)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		suite.server.ServeHTTP(w, req)

		// Accept both 201 (created) or 409 (conflict if user exists)
		assert.True(t, w.Code == http.StatusCreated || w.Code == http.StatusConflict)
	})

	// Test login
	t.Run("Login", func(t *testing.T) {
		payload := map[string]string{
			"username": "testuser",
			"password": "testpass123",
		}

		jsonPayload, _ := json.Marshal(payload)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		suite.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "token")
	})
}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.TeardownIntegrationTest(t)

	// Test unauthorized access
	t.Run("UnauthorizedAccess", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/documents", nil)
		suite.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test invalid document ID
	t.Run("InvalidDocumentID", func(t *testing.T) {
		authToken := suite.createTestUserAndLogin(t)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/documents/invalid-id", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		suite.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test invalid file upload
	t.Run("InvalidFileUpload", func(t *testing.T) {
		authToken := suite.createTestUserAndLogin(t)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add invalid file (not PDF)
		part, err := writer.CreateFormFile("document", "test.txt")
		require.NoError(t, err)
		_, err = part.Write([]byte("This is not a PDF"))
		require.NoError(t, err)

		err = writer.WriteField("issuer", "Test Issuer")
		require.NoError(t, err)

		writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/documents/sign", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+authToken)
		suite.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestConcurrentRequests tests concurrent request handling
func TestConcurrentRequests(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.TeardownIntegrationTest(t)

	authToken := suite.createTestUserAndLogin(t)
	concurrency := 10

	// Test concurrent document list requests
	t.Run("ConcurrentDocumentList", func(t *testing.T) {
		results := make(chan int, concurrency)

		for i := 0; i < concurrency; i++ {
			go func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/api/documents?page=1&limit=10", nil)
				req.Header.Set("Authorization", "Bearer "+authToken)
				suite.server.ServeHTTP(w, req)
				results <- w.Code
			}()
		}

		// Collect results
		for i := 0; i < concurrency; i++ {
			statusCode := <-results
			assert.Equal(t, http.StatusOK, statusCode)
		}
	})
}

// Helper methods

func (suite *IntegrationTestSuite) createTestUserAndLogin(t *testing.T) string {
	// Create test user
	payload := map[string]string{
		"username": "testuser_" + fmt.Sprintf("%d", time.Now().UnixNano()),
		"password": "testpass123",
		"email":    "test@example.com",
		"fullName": "Test User",
	}

	jsonPayload, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	suite.server.ServeHTTP(w, req)

	// Login
	loginPayload := map[string]string{
		"username": payload["username"],
		"password": payload["password"],
	}

	jsonPayload, _ = json.Marshal(loginPayload)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	suite.server.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	token, ok := response["token"].(string)
	if !ok {
		t.Fatal("Failed to get auth token")
	}

	return token
}

func (suite *IntegrationTestSuite) createTestDocument(t *testing.T, authToken string) string {
	pdfContent := createTestPDF(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("document", "test.pdf")
	require.NoError(t, err)
	_, err = part.Write(pdfContent)
	require.NoError(t, err)

	err = writer.WriteField("issuer", "Test Issuer")
	require.NoError(t, err)

	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/documents/sign", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+authToken)
	suite.server.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	docID, ok := response["document_id"].(string)
	if !ok {
		t.Fatal("Failed to get document ID")
	}

	return docID
}

func createTestPDF(t *testing.T) []byte {
	// Create a minimal PDF content for testing
	pdfContent := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj

2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj

3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
>>
endobj

xref
0 4
0000000000 65535 f 
0000000010 00000 n 
0000000053 00000 n 
0000000125 00000 n 
trailer
<<
/Size 4
/Root 1 0 R
>>
startxref
203
%%EOF`

	return []byte(pdfContent)
}