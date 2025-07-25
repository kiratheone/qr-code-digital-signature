package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSRFMiddleware_SafeMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	csrf := NewCSRFMiddleware(CSRFConfig{
		TokenLifetime: time.Hour,
	})
	
	router := gin.New()
	router.Use(csrf.Middleware())
	router.GET("/test", func(c *gin.Context) {
		token := csrf.GetToken(c)
		c.JSON(http.StatusOK, gin.H{"csrf_token": token})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Check that CSRF token cookie is set
	cookies := w.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "_csrf_token" {
			csrfCookie = cookie
			break
		}
	}
	
	require.NotNil(t, csrfCookie)
	assert.NotEmpty(t, csrfCookie.Value)
	assert.True(t, csrfCookie.HttpOnly)
}

func TestCSRFMiddleware_UnsafeMethods_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	csrf := NewCSRFMiddleware(CSRFConfig{
		TokenLifetime: time.Hour,
	})
	
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Mock user session
		c.Set("user_id", "test-user")
		c.Set("session_token", "test-session")
		c.Next()
	})
	router.Use(csrf.Middleware())
	
	router.GET("/get-token", func(c *gin.Context) {
		token := csrf.GetToken(c)
		c.JSON(http.StatusOK, gin.H{"csrf_token": token})
	})
	
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	
	// First, get a CSRF token
	getReq := httptest.NewRequest("GET", "/get-token", nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	
	require.Equal(t, http.StatusOK, getW.Code)
	
	// Extract token from response and cookie
	var tokenResponse map[string]string
	err := json.Unmarshal(getW.Body.Bytes(), &tokenResponse)
	require.NoError(t, err)
	token := tokenResponse["csrf_token"]
	
	// Get cookie
	cookies := getW.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "_csrf_token" {
			csrfCookie = cookie
			break
		}
	}
	require.NotNil(t, csrfCookie)
	
	// Now make POST request with valid token
	postReq := httptest.NewRequest("POST", "/test", nil)
	postReq.Header.Set("X-CSRF-Token", token)
	postReq.AddCookie(csrfCookie)
	
	// Mock same user session
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)
	
	assert.Equal(t, http.StatusOK, postW.Code)
}

func TestCSRFMiddleware_UnsafeMethods_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	csrf := NewCSRFMiddleware(CSRFConfig{
		TokenLifetime: time.Hour,
	})
	
	router := gin.New()
	router.Use(csrf.Middleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "AUTHORIZATION_ERROR", response["type"])
}

func TestCSRFMiddleware_UnsafeMethods_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	csrf := NewCSRFMiddleware(CSRFConfig{
		TokenLifetime: time.Hour,
	})
	
	router := gin.New()
	router.Use(csrf.Middleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-CSRF-Token", "invalid-token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddleware_TokenExpiration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	csrf := NewCSRFMiddleware(CSRFConfig{
		TokenLifetime: time.Millisecond * 100, // Very short lifetime
	})
	
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Set("session_token", "test-session")
		c.Next()
	})
	router.Use(csrf.Middleware())
	
	router.GET("/get-token", func(c *gin.Context) {
		token := csrf.GetToken(c)
		c.JSON(http.StatusOK, gin.H{"csrf_token": token})
	})
	
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	
	// Get token
	getReq := httptest.NewRequest("GET", "/get-token", nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	
	var tokenResponse map[string]string
	err := json.Unmarshal(getW.Body.Bytes(), &tokenResponse)
	require.NoError(t, err)
	token := tokenResponse["csrf_token"]
	
	// Wait for token to expire
	time.Sleep(time.Millisecond * 200)
	
	// Try to use expired token
	postReq := httptest.NewRequest("POST", "/test", nil)
	postReq.Header.Set("X-CSRF-Token", token)
	postW := httptest.NewRecorder()
	
	router.ServeHTTP(postW, postReq)
	
	assert.Equal(t, http.StatusForbidden, postW.Code)
}

func TestCSRFMiddleware_SkipFunction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	csrf := NewCSRFMiddleware(CSRFConfig{
		TokenLifetime: time.Hour,
		SkipFunc: func(c *gin.Context) bool {
			return strings.HasPrefix(c.Request.URL.Path, "/api/")
		},
	})
	
	router := gin.New()
	router.Use(csrf.Middleware())
	router.POST("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	
	req := httptest.NewRequest("POST", "/api/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Should succeed because CSRF is skipped for /api/ paths
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDoubleSubmitCSRF_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	csrf := NewDoubleSubmitCSRF(CSRFConfig{})
	
	router := gin.New()
	router.Use(csrf.Middleware())
	
	router.GET("/get-token", func(c *gin.Context) {
		if token, exists := c.Get("csrf_token"); exists {
			c.JSON(http.StatusOK, gin.H{"csrf_token": token})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no token"})
		}
	})
	
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	
	// Get token
	getReq := httptest.NewRequest("GET", "/get-token", nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	
	require.Equal(t, http.StatusOK, getW.Code)
	
	var tokenResponse map[string]string
	err := json.Unmarshal(getW.Body.Bytes(), &tokenResponse)
	require.NoError(t, err)
	token := tokenResponse["csrf_token"]
	
	// Get cookie
	cookies := getW.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "_csrf_token" {
			csrfCookie = cookie
			break
		}
	}
	require.NotNil(t, csrfCookie)
	
	// Make POST request with matching token in header and cookie
	postReq := httptest.NewRequest("POST", "/test", nil)
	postReq.Header.Set("X-CSRF-Token", token)
	postReq.AddCookie(csrfCookie)
	
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)
	
	assert.Equal(t, http.StatusOK, postW.Code)
}

func TestDoubleSubmitCSRF_TokenMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	csrf := NewDoubleSubmitCSRF(CSRFConfig{})
	
	router := gin.New()
	router.Use(csrf.Middleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	
	// Create mismatched tokens
	cookieToken := "cookie-token"
	headerToken := "header-token"
	
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-CSRF-Token", headerToken)
	req.AddCookie(&http.Cookie{
		Name:  "_csrf_token",
		Value: cookieToken,
	})
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDoubleSubmitCSRF_FormField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	csrf := NewDoubleSubmitCSRF(CSRFConfig{})
	
	router := gin.New()
	router.Use(csrf.Middleware())
	
	router.GET("/get-token", func(c *gin.Context) {
		if token, exists := c.Get("csrf_token"); exists {
			c.JSON(http.StatusOK, gin.H{"csrf_token": token})
		}
	})
	
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	
	// Get token
	getReq := httptest.NewRequest("GET", "/get-token", nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	
	var tokenResponse map[string]string
	err := json.Unmarshal(getW.Body.Bytes(), &tokenResponse)
	require.NoError(t, err)
	token := tokenResponse["csrf_token"]
	
	// Get cookie
	cookies := getW.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "_csrf_token" {
			csrfCookie = cookie
			break
		}
	}
	require.NotNil(t, csrfCookie)
	
	// Make POST request with token in form field
	formData := url.Values{}
	formData.Set("_csrf_token", token)
	
	postReq := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(csrfCookie)
	
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)
	
	assert.Equal(t, http.StatusOK, postW.Code)
}

func TestCSRFMiddleware_CustomErrorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	customErrorCalled := false
	csrf := NewCSRFMiddleware(CSRFConfig{
		TokenLifetime: time.Hour,
		ErrorHandler: func(c *gin.Context, err error) {
			customErrorCalled = true
			c.JSON(http.StatusTeapot, gin.H{"custom_error": err.Error()})
			c.Abort()
		},
	})
	
	router := gin.New()
	router.Use(csrf.Middleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusTeapot, w.Code)
	assert.True(t, customErrorCalled)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "custom_error")
}

func TestCSRFSkipFunctions(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		skipFunc func(*gin.Context) bool
		expected bool
	}{
		{
			name:     "skip API endpoints",
			path:     "/api/test",
			skipFunc: SkipCSRFForAPI,
			expected: true,
		},
		{
			name:     "don't skip non-API endpoints",
			path:     "/web/test",
			skipFunc: SkipCSRFForAPI,
			expected: false,
		},
		{
			name:     "skip public endpoints - health",
			path:     "/health",
			skipFunc: SkipCSRFForPublicEndpoints,
			expected: true,
		},
		{
			name:     "skip public endpoints - verify",
			path:     "/api/verify/123",
			skipFunc: SkipCSRFForPublicEndpoints,
			expected: true,
		},
		{
			name:     "skip public endpoints - login",
			path:     "/api/auth/login",
			skipFunc: SkipCSRFForPublicEndpoints,
			expected: true,
		},
		{
			name:     "don't skip private endpoints",
			path:     "/api/documents",
			skipFunc: SkipCSRFForPublicEndpoints,
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("GET", tt.path, nil)
			
			result := tt.skipFunc(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCSRFMiddleware_ConcurrentAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	csrf := NewCSRFMiddleware(CSRFConfig{
		TokenLifetime: time.Hour,
	})
	
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Set("session_token", "test-session")
		c.Next()
	})
	router.Use(csrf.Middleware())
	
	router.GET("/get-token", func(c *gin.Context) {
		token := csrf.GetToken(c)
		c.JSON(http.StatusOK, gin.H{"csrf_token": token})
	})
	
	// Test concurrent token generation
	const numGoroutines = 10
	tokens := make(chan string, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/get-token", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code == http.StatusOK {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				tokens <- response["csrf_token"]
			} else {
				tokens <- ""
			}
		}()
	}
	
	// Collect all tokens
	uniqueTokens := make(map[string]bool)
	for i := 0; i < numGoroutines; i++ {
		token := <-tokens
		if token != "" {
			uniqueTokens[token] = true
		}
	}
	
	// All tokens should be unique
	assert.Equal(t, numGoroutines, len(uniqueTokens))
}