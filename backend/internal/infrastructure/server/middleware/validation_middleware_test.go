package middleware

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationMiddleware_ValidateJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	type TestRequest struct {
		Username string `json:"username" binding:"required,min=3,max=50" validate:"username"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8" validate:"password"`
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid request",
			requestBody: TestRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "Password123",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "missing required field",
			requestBody: TestRequest{
				Email:    "test@example.com",
				Password: "Password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "invalid email format",
			requestBody: TestRequest{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "Password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "password too short",
			requestBody: TestRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "short",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "invalid JSON",
			requestBody:    `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			vm := NewValidationMiddleware()
			
			router.POST("/test", vm.ValidateJSON(&TestRequest{}), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				var err error
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "type")
				assert.Equal(t, "VALIDATION_ERROR", response["type"])
			}
		})
	}
}

func TestValidationMiddleware_ValidateQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "valid query params",
			queryParams:    "page=1&page_size=10&search=test",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid page parameter",
			queryParams:    "page=invalid&page_size=10",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "page size too large",
			queryParams:    "page=1&page_size=1000",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "missing required parameter",
			queryParams:    "page_size=10",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "search too long",
			queryParams:    "page=1&search=" + strings.Repeat("a", 101),
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			vm := NewValidationMiddleware()
			
			defaultPage := 1
			defaultPageSize := 10
			maxPageSize := 100
			
			queryConfig := map[string]QueryParamConfig{
				"page": {
					Type:         "int",
					Required:     true,
					DefaultValue: &defaultPage,
					MinValue:     &defaultPage,
				},
				"page_size": {
					Type:         "int",
					Required:     false,
					DefaultValue: &defaultPageSize,
					MinValue:     &defaultPage,
					MaxValue:     &maxPageSize,
				},
				"search": {
					Type:      "string",
					Required:  false,
					MaxLength: 100,
				},
			}
			
			router.GET("/test", vm.ValidateQueryParams(queryConfig), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "type")
				assert.Equal(t, "VALIDATION_ERROR", response["type"])
			}
		})
	}
}

func TestValidationMiddleware_ValidatePathParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "valid UUID",
			path:           "/test/123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid UUID format",
			path:           "/test/invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "empty parameter",
			path:           "/test/",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			vm := NewValidationMiddleware()
			
			pathConfig := map[string]PathParamConfig{
				"id": {
					Type:      "uuid",
					MinLength: 36,
					MaxLength: 36,
				},
			}
			
			router.GET("/test/:id", vm.ValidatePathParams(pathConfig), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "type")
				assert.Equal(t, "VALIDATION_ERROR", response["type"])
			}
		})
	}
}

func TestValidationMiddleware_ValidatePDFUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		filename       string
		content        string
		contentType    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "valid PDF file",
			filename:       "test.pdf",
			content:        "%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj",
			contentType:    "application/pdf",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid file extension",
			filename:       "test.txt",
			content:        "This is not a PDF",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "file too large",
			filename:       "large.pdf",
			content:        strings.Repeat("a", 1024*1024+1), // 1MB + 1 byte
			contentType:    "application/pdf",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			vm := NewValidationMiddleware()
			
			router.POST("/test", vm.ValidatePDFUpload("pdf", 1024*1024, true), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			// Create multipart form
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			
			part, err := writer.CreateFormFile("pdf", tt.filename)
			require.NoError(t, err)
			
			_, err = part.Write([]byte(tt.content))
			require.NoError(t, err)
			
			err = writer.Close()
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/test", &buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "type")
				assert.Equal(t, "VALIDATION_ERROR", response["type"])
			}
		})
	}
}

func TestValidationMiddleware_SanitizeAllInputs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParam     string
		headerValue    string
		expectedStatus int
	}{
		{
			name:           "clean inputs",
			queryParam:     "normal_value",
			headerValue:    "normal_header",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "malicious script in query",
			queryParam:     "<script>alert('xss')</script>",
			headerValue:    "normal_header",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "malicious script in header",
			queryParam:     "normal_value",
			headerValue:    "javascript:alert('xss')",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			vm := NewValidationMiddleware()
			
			router.GET("/test", vm.SanitizeAllInputs(), func(c *gin.Context) {
				// Check that dangerous content has been sanitized
				queryValue := c.Query("test")
				headerValue := c.GetHeader("X-Test-Header")
				
				// Verify sanitization occurred
				assert.NotContains(t, queryValue, "<script")
				assert.NotContains(t, queryValue, "javascript:")
				assert.NotContains(t, headerValue, "<script")
				assert.NotContains(t, headerValue, "javascript:")
				
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test?test="+tt.queryParam, nil)
			req.Header.Set("X-Test-Header", tt.headerValue)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestValidationMiddleware_ValidateHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		headers        map[string]string
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid headers",
			headers: map[string]string{
				"X-API-Key":     "valid-api-key-123",
				"X-Request-ID":  "req-123",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "missing required header",
			headers: map[string]string{
				"X-Request-ID": "req-123",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "header too long",
			headers: map[string]string{
				"X-API-Key":    strings.Repeat("a", 101),
				"X-Request-ID": "req-123",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			vm := NewValidationMiddleware()
			
			headerConfig := map[string]HeaderConfig{
				"X-API-Key": {
					Required:  true,
					MaxLength: 100,
				},
				"X-Request-ID": {
					Required:  false,
					MaxLength: 50,
				},
			}
			
			router.GET("/test", vm.ValidateHeaders(headerConfig), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectError {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "type")
				assert.Equal(t, "VALIDATION_ERROR", response["type"])
			}
		})
	}
}

func TestConvenienceFunctions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ValidateDocumentID", func(t *testing.T) {
		router := gin.New()
		router.GET("/documents/:id", ValidateDocumentID(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Valid UUID
		req := httptest.NewRequest("GET", "/documents/123e4567-e89b-12d3-a456-426614174000", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Invalid UUID
		req = httptest.NewRequest("GET", "/documents/invalid-id", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ValidateDocumentQuery", func(t *testing.T) {
		router := gin.New()
		router.GET("/documents", ValidateDocumentQuery(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Valid query
		req := httptest.NewRequest("GET", "/documents?page=1&page_size=10&search=test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Invalid page size
		req = httptest.NewRequest("GET", "/documents?page=1&page_size=1000", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestSecurityValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("XSS Prevention", func(t *testing.T) {
		router := gin.New()
		vm := NewValidationMiddleware()
		
		router.GET("/test", vm.SanitizeAllInputs(), func(c *gin.Context) {
			value := c.Query("input")
			// Should not contain script tags
			assert.NotContains(t, value, "<script")
			assert.NotContains(t, value, "javascript:")
			c.JSON(http.StatusOK, gin.H{"sanitized": value})
		})

		maliciousInputs := []string{
			"<script>alert('xss')</script>",
			"javascript:alert('xss')",
			"<img src=x onerror=alert('xss')>",
			"data:text/html,<script>alert('xss')</script>",
		}

		for _, input := range maliciousInputs {
			req := httptest.NewRequest("GET", "/test?input="+input, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("Path Traversal Prevention", func(t *testing.T) {
		router := gin.New()
		vm := NewValidationMiddleware()
		
		type FileRequest struct {
			Filename string `json:"filename" validate:"filename"`
		}
		
		router.POST("/test", vm.ValidateJSON(&FileRequest{}), func(c *gin.Context) {
			data, _ := c.Get("validated_data")
			req := data.(*FileRequest)
			
			// Should not contain path traversal sequences
			assert.NotContains(t, req.Filename, "..")
			assert.NotContains(t, req.Filename, "/")
			assert.NotContains(t, req.Filename, "\\")
			
			c.JSON(http.StatusOK, gin.H{"filename": req.Filename})
		})

		maliciousFilenames := []string{
			"../../../etc/passwd",
			"..\\..\\windows\\system32\\config\\sam",
			"file/with/slashes",
			"file\\with\\backslashes",
		}

		for _, filename := range maliciousFilenames {
			body, _ := json.Marshal(FileRequest{Filename: filename})
			req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			// Should either be sanitized (200) or rejected (400)
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
		}
	})
}