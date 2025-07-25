package main

import (
	"bytes"
	"context"
	"digital-signature-system/internal/config"
	"digital-signature-system/internal/infrastructure/database"
	"digital-signature-system/internal/infrastructure/di"
	"digital-signature-system/internal/infrastructure/server"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

// E2ETestSuite represents the end-to-end test suite
type E2ETestSuite struct {
	baseURL   string
	client    *http.Client
	authToken string
}

// TestResult represents the result of a test
type TestResult struct {
	Name     string
	Success  bool
	Error    error
	Duration time.Duration
}

func main() {
	fmt.Println("🧪 Starting End-to-End Tests for Digital Signature System")
	fmt.Println("=========================================================")

	// Configuration
	baseURL := getEnv("API_URL", "http://localhost:8000")
	
	suite := &E2ETestSuite{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Wait for server to be ready
	if !suite.waitForServer() {
		log.Fatal("❌ Server is not ready")
	}

	// Run test scenarios
	scenarios := []struct {
		name string
		fn   func() error
	}{
		{"Health Check", suite.testHealthCheck},
		{"User Authentication", suite.testAuthentication},
		{"Document Signing Workflow", suite.testDocumentSigningWorkflow},
		{"Document Management Workflow", suite.testDocumentManagementWorkflow},
		{"Document Verification Workflow", suite.testDocumentVerificationWorkflow},
		{"Error Handling", suite.testErrorHandling},
		{"Performance Under Load", suite.testPerformanceUnderLoad},
	}

	results := make([]TestResult, 0, len(scenarios))

	for _, scenario := range scenarios {
		fmt.Printf("\n🔍 Running: %s\n", scenario.name)
		fmt.Println(strings.Repeat("-", 50))

		start := time.Now()
		err := scenario.fn()
		duration := time.Since(start)

		result := TestResult{
			Name:     scenario.name,
			Success:  err == nil,
			Error:    err,
			Duration: duration,
		}
		results = append(results, result)

		if err != nil {
			fmt.Printf("❌ FAILED: %v (took %v)\n", err, duration)
		} else {
			fmt.Printf("✅ PASSED (took %v)\n", duration)
		}
	}

	// Print summary
	fmt.Println("\n📊 Test Summary")
	fmt.Println("================")

	passed := 0
	failed := 0
	totalDuration := time.Duration(0)

	for _, result := range results {
		totalDuration += result.Duration
		if result.Success {
			passed++
			fmt.Printf("✅ %s - %v\n", result.Name, result.Duration)
		} else {
			failed++
			fmt.Printf("❌ %s - %v - %v\n", result.Name, result.Duration, result.Error)
		}
	}

	fmt.Printf("\nTotal: %d, Passed: %d, Failed: %d\n", len(results), passed, failed)
	fmt.Printf("Total Duration: %v\n", totalDuration)

	if failed > 0 {
		os.Exit(1)
	}

	fmt.Println("\n🎉 All tests passed!")
}

func (suite *E2ETestSuite) waitForServer() bool {
	fmt.Print("⏳ Waiting for server to be ready...")
	
	for i := 0; i < 30; i++ { // Wait up to 30 seconds
		resp, err := suite.client.Get(suite.baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			fmt.Println(" ✅ Ready!")
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
		fmt.Print(".")
	}
	
	fmt.Println(" ❌ Timeout!")
	return false
}

func (suite *E2ETestSuite) testHealthCheck() error {
	resp, err := suite.client.Get(suite.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode health check response: %w", err)
	}

	if result["status"] != "healthy" {
		return fmt.Errorf("server is not healthy: %v", result["status"])
	}

	fmt.Println("  ✓ Server is healthy")
	return nil
}

func (suite *E2ETestSuite) testAuthentication() error {
	// Test user registration
	username := fmt.Sprintf("e2euser_%d", time.Now().UnixNano())
	registerPayload := map[string]string{
		"username": username,
		"password": "testpass123",
		"email":    "e2e@example.com",
		"fullName": "E2E Test User",
	}

	jsonPayload, _ := json.Marshal(registerPayload)
	resp, err := suite.client.Post(
		suite.baseURL+"/api/auth/register",
		"application/json",
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Println("  ✓ User registration successful")

	// Test login
	loginPayload := map[string]string{
		"username": username,
		"password": "testpass123",
	}

	jsonPayload, _ = json.Marshal(loginPayload)
	resp, err = suite.client.Post(
		suite.baseURL+"/api/auth/login",
		"application/json",
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&loginResult); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	token, ok := loginResult["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("no auth token received")
	}

	suite.authToken = token
	fmt.Println("  ✓ User login successful")
	return nil
}

func (suite *E2ETestSuite) testDocumentSigningWorkflow() error {
	if suite.authToken == "" {
		return fmt.Errorf("no auth token available")
	}

	// Create test PDF content
	pdfContent := suite.createTestPDF()

	// Prepare multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("document", "e2e_test.pdf")
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(pdfContent); err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	// Add issuer
	if err := writer.WriteField("issuer", "E2E Test Issuer"); err != nil {
		return fmt.Errorf("failed to add issuer field: %w", err)
	}

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", suite.baseURL+"/api/documents/sign", body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	// Send request
	resp, err := suite.client.Do(req)
	if err != nil {
		return fmt.Errorf("document signing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("document signing failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode signing response: %w", err)
	}

	// Verify response contains required fields
	requiredFields := []string{"document_id", "qr_code", "download_url"}
	for _, field := range requiredFields {
		if _, exists := result[field]; !exists {
			return fmt.Errorf("missing required field in response: %s", field)
		}
	}

	fmt.Println("  ✓ Document signing successful")
	fmt.Printf("  ✓ Document ID: %v\n", result["document_id"])
	return nil
}

func (suite *E2ETestSuite) testDocumentManagementWorkflow() error {
	if suite.authToken == "" {
		return fmt.Errorf("no auth token available")
	}

	// First, sign a document to have something to manage
	docID, err := suite.signTestDocument()
	if err != nil {
		return fmt.Errorf("failed to create test document: %w", err)
	}

	// Test getting documents list
	req, err := http.NewRequest("GET", suite.baseURL+"/api/documents?page=1&limit=10", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	resp, err := suite.client.Do(req)
	if err != nil {
		return fmt.Errorf("get documents request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get documents failed with status %d: %s", resp.StatusCode, string(body))
	}

	var listResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&listResult); err != nil {
		return fmt.Errorf("failed to decode documents list response: %w", err)
	}

	fmt.Println("  ✓ Document list retrieval successful")

	// Test getting single document
	req, err = http.NewRequest("GET", suite.baseURL+"/api/documents/"+docID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	resp, err = suite.client.Do(req)
	if err != nil {
		return fmt.Errorf("get document request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get document failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Println("  ✓ Single document retrieval successful")

	// Test deleting document
	req, err = http.NewRequest("DELETE", suite.baseURL+"/api/documents/"+docID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	resp, err = suite.client.Do(req)
	if err != nil {
		return fmt.Errorf("delete document request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete document failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Println("  ✓ Document deletion successful")
	return nil
}

func (suite *E2ETestSuite) testDocumentVerificationWorkflow() error {
	// Sign a document first
	docID, err := suite.signTestDocument()
	if err != nil {
		return fmt.Errorf("failed to create test document: %w", err)
	}

	// Test getting verification info (no auth required)
	resp, err := suite.client.Get(suite.baseURL + "/api/verify/" + docID)
	if err != nil {
		return fmt.Errorf("get verification info request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get verification info failed with status %d: %s", resp.StatusCode, string(body))
	}

	var verifyInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&verifyInfo); err != nil {
		return fmt.Errorf("failed to decode verification info response: %w", err)
	}

	fmt.Println("  ✓ Verification info retrieval successful")

	// Test document verification
	pdfContent := suite.createTestPDF()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("document", "verify_test.pdf")
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(pdfContent); err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", suite.baseURL+"/api/verify/"+docID+"/upload", body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err = suite.client.Do(req)
	if err != nil {
		return fmt.Errorf("document verification request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("document verification failed with status %d: %s", resp.StatusCode, string(body))
	}

	var verifyResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&verifyResult); err != nil {
		return fmt.Errorf("failed to decode verification response: %w", err)
	}

	fmt.Println("  ✓ Document verification successful")
	fmt.Printf("  ✓ Verification status: %v\n", verifyResult["status"])
	return nil
}

func (suite *E2ETestSuite) testErrorHandling() error {
	// Test unauthorized access
	req, err := http.NewRequest("GET", suite.baseURL+"/api/documents", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := suite.client.Do(req)
	if err != nil {
		return fmt.Errorf("unauthorized request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}

	fmt.Println("  ✓ Unauthorized access properly rejected")

	// Test invalid document ID
	if suite.authToken != "" {
		req, err = http.NewRequest("GET", suite.baseURL+"/api/documents/invalid-id", nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+suite.authToken)

		resp, err = suite.client.Do(req)
		if err != nil {
			return fmt.Errorf("invalid document ID request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("expected 404 Not Found, got %d", resp.StatusCode)
		}

		fmt.Println("  ✓ Invalid document ID properly handled")
	}

	return nil
}

func (suite *E2ETestSuite) testPerformanceUnderLoad() error {
	if suite.authToken == "" {
		return fmt.Errorf("no auth token available")
	}

	// Test concurrent requests
	concurrency := 5
	requests := 10
	
	results := make(chan error, concurrency)
	
	for i := 0; i < concurrency; i++ {
		go func() {
			for j := 0; j < requests; j++ {
				req, err := http.NewRequest("GET", suite.baseURL+"/api/documents?page=1&limit=5", nil)
				if err != nil {
					results <- err
					return
				}
				req.Header.Set("Authorization", "Bearer "+suite.authToken)

				resp, err := suite.client.Do(req)
				if err != nil {
					results <- err
					return
				}
				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					results <- fmt.Errorf("request failed with status %d", resp.StatusCode)
					return
				}
			}
			results <- nil
		}()
	}

	// Collect results
	for i := 0; i < concurrency; i++ {
		if err := <-results; err != nil {
			return fmt.Errorf("concurrent request failed: %w", err)
		}
	}

	fmt.Printf("  ✓ Handled %d concurrent requests successfully\n", concurrency*requests)
	return nil
}

// Helper methods

func (suite *E2ETestSuite) signTestDocument() (string, error) {
	if suite.authToken == "" {
		return "", fmt.Errorf("no auth token available")
	}

	pdfContent := suite.createTestPDF()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("document", "helper_test.pdf")
	if err != nil {
		return "", err
	}
	if _, err := part.Write(pdfContent); err != nil {
		return "", err
	}

	if err := writer.WriteField("issuer", "Helper Test Issuer"); err != nil {
		return "", err
	}

	writer.Close()

	req, err := http.NewRequest("POST", suite.baseURL+"/api/documents/sign", body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	resp, err := suite.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("signing failed with status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	docID, ok := result["document_id"].(string)
	if !ok {
		return "", fmt.Errorf("no document_id in response")
	}

	return docID, nil
}

func (suite *E2ETestSuite) createTestPDF() []byte {
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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}