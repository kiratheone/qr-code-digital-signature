package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityMiddleware_ValidateSecurityHeaders(t *testing.T) {
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
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				"Referer":    "https://example.com/page",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "malicious user agent",
			headers: map[string]string{
				"User-Agent": "<script>alert('xss')</script>",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "malicious referer",
			headers: map[string]string{
				"Referer": "javascript:alert('xss')",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "user agent too long",
			headers: map[string]string{
				"User-Agent": strings.Repeat("a", 600),
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			sm := NewSecurityMiddleware()
			
			router.GET("/test", sm.ValidateSecurityHeaders(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSecurityMiddleware_DetectMaliciousInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "clean query parameters",
			queryParams:    "search=normal&page=1",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "SQL injection in query",
			queryParams:    "search=' OR 1=1--",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "XSS in query",
			queryParams:    "search=<script>alert('xss')</script>",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "path traversal in query",
			queryParams:    "file=../../../etc/passwd",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "javascript injection",
			queryParams:    "callback=javascript:alert('xss')",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			sm := NewSecurityMiddleware()
			
			router.GET("/test", sm.DetectMaliciousInput(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSecurityMiddleware_ValidateRequestSize(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		contentLength  int64
		maxSize        int64
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "valid size",
			contentLength:  1024,
			maxSize:        2048,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "size too large",
			contentLength:  3072,
			maxSize:        2048,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "zero size",
			contentLength:  0,
			maxSize:        2048,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			sm := NewSecurityMiddleware()
			
			router.POST("/test", sm.ValidateRequestSize(tt.maxSize), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			body := strings.Repeat("a", int(tt.contentLength))
			req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
			req.Header.Set("Content-Length", string(rune(tt.contentLength)))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSecurityMiddleware_ValidateHTTPMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "GET method",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "POST method",
			method:         "POST",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "PUT method",
			method:         "PUT",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "DELETE method",
			method:         "DELETE",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "TRACE method (not allowed)",
			method:         "TRACE",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "CONNECT method (not allowed)",
			method:         "CONNECT",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			sm := NewSecurityMiddleware()
			
			// Register handler for all methods
			router.Any("/test", sm.ValidateHTTPMethod(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSecurityMiddleware_ValidateContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		contentType    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "valid JSON content type",
			method:         "POST",
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "valid form content type",
			method:         "POST",
			contentType:    "application/x-www-form-urlencoded",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "valid multipart content type",
			method:         "POST",
			contentType:    "multipart/form-data; boundary=something",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "invalid content type",
			method:         "POST",
			contentType:    "application/xml",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "GET method (no content type validation)",
			method:         "GET",
			contentType:    "application/xml",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "empty content type for POST",
			method:         "POST",
			contentType:    "",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			sm := NewSecurityMiddleware()
			
			router.Any("/test", sm.ValidateContentType(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestValidateAPIRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		headers        map[string]string
		body           string
		expectedStatus int
	}{
		{
			name:   "valid API request",
			method: "POST",
			path:   "/api/test",
			headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			},
			body:           `{"test": "data"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "malicious query parameter",
			method: "GET",
			path:   "/api/test?search=' OR 1=1--",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid HTTP method",
			method: "TRACE",
			path:   "/api/test",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "malicious user agent",
			method: "GET",
			path:   "/api/test",
			headers: map[string]string{
				"User-Agent": "<script>alert('xss')</script>",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			
			router.Any("/api/test", ValidateAPIRequest(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			var body *bytes.Buffer
			if tt.body != "" {
				body = bytes.NewBufferString(tt.body)
			} else {
				body = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(tt.method, tt.path, body)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestValidateFileUploadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "valid multipart content type",
			contentType:    "multipart/form-data; boundary=something",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid content type for file upload",
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty content type",
			contentType:    "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			
			router.POST("/upload", ValidateFileUploadRequest(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("POST", "/upload", nil)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSecurityMiddleware_PreventClickjacking(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	sm := NewSecurityMiddleware()
	
	router.GET("/test", sm.PreventClickjacking(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'")
}

func TestSecurityValidation_SQLInjectionDetection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	maliciousSQLInputs := []string{
		"' OR 1=1--",
		"'; DROP TABLE users;--",
		"1' UNION SELECT * FROM users--",
		"admin'/*",
		"' OR 'a'='a",
		"1; EXEC xp_cmdshell('dir')",
		"' OR 1=1#",
		"1' AND 1=1--",
		"' UNION ALL SELECT NULL,NULL,NULL--",
		"1' OR '1'='1",
	}

	for _, input := range maliciousSQLInputs {
		t.Run("SQL injection: "+input, func(t *testing.T) {
			router := gin.New()
			sm := NewSecurityMiddleware()
			
			router.GET("/test", sm.DetectMaliciousInput(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test?search="+input, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code, "Should detect SQL injection: %s", input)
		})
	}
}

func TestSecurityValidation_XSSDetection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	maliciousXSSInputs := []string{
		"<script>alert('xss')</script>",
		"javascript:alert('xss')",
		"<img src=x onerror=alert('xss')>",
		"<svg onload=alert('xss')>",
		"<iframe src=javascript:alert('xss')>",
		"<body onload=alert('xss')>",
		"<input onfocus=alert('xss') autofocus>",
		"<select onfocus=alert('xss') autofocus>",
		"<textarea onfocus=alert('xss') autofocus>",
		"<keygen onfocus=alert('xss') autofocus>",
	}

	for _, input := range maliciousXSSInputs {
		t.Run("XSS: "+input, func(t *testing.T) {
			router := gin.New()
			sm := NewSecurityMiddleware()
			
			router.GET("/test", sm.DetectMaliciousInput(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test?input="+input, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code, "Should detect XSS: %s", input)
		})
	}
}

func TestSecurityValidation_PathTraversalDetection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	maliciousPathInputs := []string{
		"../../../etc/passwd",
		"..\\..\\windows\\system32\\config\\sam",
		"....//....//etc/passwd",
		"....\\\\....\\\\windows\\system32",
		"/etc/passwd",
		"\\windows\\system32\\",
		"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
		"..%2f..%2f..%2fetc%2fpasswd",
		"..%5c..%5c..%5cwindows%5csystem32",
	}

	for _, input := range maliciousPathInputs {
		t.Run("Path traversal: "+input, func(t *testing.T) {
			router := gin.New()
			sm := NewSecurityMiddleware()
			
			router.GET("/test", sm.DetectMaliciousInput(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test?file="+input, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code, "Should detect path traversal: %s", input)
		})
	}
}