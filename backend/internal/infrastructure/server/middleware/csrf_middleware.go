package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CSRFConfig represents CSRF protection configuration
type CSRFConfig struct {
	TokenLength    int           // Length of CSRF token
	TokenLifetime  time.Duration // How long tokens are valid
	CookieName     string        // Name of CSRF cookie
	HeaderName     string        // Name of CSRF header
	FormFieldName  string        // Name of CSRF form field
	SecureCookie   bool          // Whether to use secure cookies
	SameSite       http.SameSite // SameSite cookie attribute
	SkipFunc       func(*gin.Context) bool // Function to skip CSRF protection
	ErrorHandler   func(*gin.Context, error) // Custom error handler
}

// CSRFToken represents a CSRF token with metadata
type CSRFToken struct {
	Value     string
	CreatedAt time.Time
	UserID    string
	SessionID string
}

// CSRFMiddleware provides CSRF protection
type CSRFMiddleware struct {
	config CSRFConfig
	tokens map[string]*CSRFToken
	mutex  sync.RWMutex
}

// NewCSRFMiddleware creates a new CSRF middleware
func NewCSRFMiddleware(config CSRFConfig) *CSRFMiddleware {
	// Set defaults
	if config.TokenLength == 0 {
		config.TokenLength = 32
	}
	if config.TokenLifetime == 0 {
		config.TokenLifetime = time.Hour
	}
	if config.CookieName == "" {
		config.CookieName = "_csrf_token"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-CSRF-Token"
	}
	if config.FormFieldName == "" {
		config.FormFieldName = "_csrf_token"
	}
	if config.SameSite == 0 {
		config.SameSite = http.SameSiteStrictMode
	}
	if config.SkipFunc == nil {
		config.SkipFunc = func(c *gin.Context) bool {
			return false
		}
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultCSRFErrorHandler
	}

	csrf := &CSRFMiddleware{
		config: config,
		tokens: make(map[string]*CSRFToken),
	}

	// Start cleanup goroutine
	go csrf.cleanupExpiredTokens()

	return csrf
}

// Middleware returns the CSRF protection middleware
func (csrf *CSRFMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF protection if configured to skip
		if csrf.config.SkipFunc(c) {
			c.Next()
			return
		}

		// Skip for safe methods (GET, HEAD, OPTIONS)
		if csrf.isSafeMethod(c.Request.Method) {
			// Generate and set token for safe methods
			token, err := csrf.generateToken(c)
			if err != nil {
				csrf.config.ErrorHandler(c, fmt.Errorf("failed to generate CSRF token: %w", err))
				return
			}
			csrf.setTokenCookie(c, token)
			c.Set("csrf_token", token)
			c.Next()
			return
		}

		// Validate token for unsafe methods (POST, PUT, DELETE, PATCH)
		if err := csrf.validateToken(c); err != nil {
			csrf.config.ErrorHandler(c, err)
			return
		}

		c.Next()
	}
}

// GetToken returns the CSRF token for the current request
func (csrf *CSRFMiddleware) GetToken(c *gin.Context) string {
	if token, exists := c.Get("csrf_token"); exists {
		return token.(string)
	}
	return ""
}

// generateToken generates a new CSRF token
func (csrf *CSRFMiddleware) generateToken(c *gin.Context) (string, error) {
	// Generate random bytes
	bytes := make([]byte, csrf.config.TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode to base64
	token := base64.URLEncoding.EncodeToString(bytes)

	// Get user and session info
	userID := GetUserID(c)
	sessionID := GetSessionToken(c)

	// Store token
	csrf.mutex.Lock()
	csrf.tokens[token] = &CSRFToken{
		Value:     token,
		CreatedAt: time.Now(),
		UserID:    userID,
		SessionID: sessionID,
	}
	csrf.mutex.Unlock()

	return token, nil
}

// validateToken validates the CSRF token from request
func (csrf *CSRFMiddleware) validateToken(c *gin.Context) error {
	// Get token from various sources
	token := csrf.getTokenFromRequest(c)
	if token == "" {
		return fmt.Errorf("CSRF token not found")
	}

	// Check if token exists and is valid
	csrf.mutex.RLock()
	storedToken, exists := csrf.tokens[token]
	csrf.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("invalid CSRF token")
	}

	// Check if token is expired
	if time.Since(storedToken.CreatedAt) > csrf.config.TokenLifetime {
		csrf.mutex.Lock()
		delete(csrf.tokens, token)
		csrf.mutex.Unlock()
		return fmt.Errorf("CSRF token expired")
	}

	// Validate token against current user/session
	currentUserID := GetUserID(c)
	currentSessionID := GetSessionToken(c)

	if storedToken.UserID != currentUserID || storedToken.SessionID != currentSessionID {
		return fmt.Errorf("CSRF token mismatch")
	}

	// Token is valid, regenerate for next request
	newToken, err := csrf.generateToken(c)
	if err != nil {
		return fmt.Errorf("failed to regenerate CSRF token: %w", err)
	}

	// Remove old token
	csrf.mutex.Lock()
	delete(csrf.tokens, token)
	csrf.mutex.Unlock()

	// Set new token
	csrf.setTokenCookie(c, newToken)
	c.Set("csrf_token", newToken)

	return nil
}

// getTokenFromRequest extracts CSRF token from request
func (csrf *CSRFMiddleware) getTokenFromRequest(c *gin.Context) string {
	// Try header first
	if token := c.GetHeader(csrf.config.HeaderName); token != "" {
		return token
	}

	// Try form field
	if token := c.PostForm(csrf.config.FormFieldName); token != "" {
		return token
	}

	// Try cookie as fallback
	if cookie, err := c.Cookie(csrf.config.CookieName); err == nil {
		return cookie
	}

	return ""
}

// setTokenCookie sets the CSRF token cookie
func (csrf *CSRFMiddleware) setTokenCookie(c *gin.Context, token string) {
	c.SetSameSite(csrf.config.SameSite)
	c.SetCookie(
		csrf.config.CookieName,
		token,
		int(csrf.config.TokenLifetime.Seconds()),
		"/",
		"",
		csrf.config.SecureCookie,
		true, // HttpOnly
	)
}

// isSafeMethod checks if HTTP method is safe (doesn't modify state)
func (csrf *CSRFMiddleware) isSafeMethod(method string) bool {
	safeMethods := []string{"GET", "HEAD", "OPTIONS", "TRACE"}
	for _, safe := range safeMethods {
		if method == safe {
			return true
		}
	}
	return false
}

// cleanupExpiredTokens removes expired tokens from memory
func (csrf *CSRFMiddleware) cleanupExpiredTokens() {
	ticker := time.NewTicker(time.Minute * 10) // Cleanup every 10 minutes
	defer ticker.Stop()

	for range ticker.C {
		csrf.mutex.Lock()
		now := time.Now()
		for token, tokenData := range csrf.tokens {
			if now.Sub(tokenData.CreatedAt) > csrf.config.TokenLifetime {
				delete(csrf.tokens, token)
			}
		}
		csrf.mutex.Unlock()
	}
}

// defaultCSRFErrorHandler is the default error handler for CSRF failures
func defaultCSRFErrorHandler(c *gin.Context, err error) {
	apiError := NewAuthorizationError("CSRF token validation failed")
	apiError.Details = err.Error()
	
	if requestID, exists := c.Get("request_id"); exists {
		apiError.RequestID = requestID.(string)
	}
	apiError.Path = c.Request.URL.Path
	apiError.Method = c.Request.Method

	c.JSON(http.StatusForbidden, apiError)
	c.Abort()
}

// DoubleSubmitCSRF implements double submit cookie pattern
type DoubleSubmitCSRF struct {
	config CSRFConfig
}

// NewDoubleSubmitCSRF creates a double submit CSRF middleware
func NewDoubleSubmitCSRF(config CSRFConfig) *DoubleSubmitCSRF {
	if config.TokenLength == 0 {
		config.TokenLength = 32
	}
	if config.CookieName == "" {
		config.CookieName = "_csrf_token"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-CSRF-Token"
	}
	if config.FormFieldName == "" {
		config.FormFieldName = "_csrf_token"
	}
	if config.SameSite == 0 {
		config.SameSite = http.SameSiteStrictMode
	}
	if config.SkipFunc == nil {
		config.SkipFunc = func(c *gin.Context) bool {
			return false
		}
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultCSRFErrorHandler
	}

	return &DoubleSubmitCSRF{config: config}
}

// Middleware returns the double submit CSRF middleware
func (ds *DoubleSubmitCSRF) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ds.config.SkipFunc(c) {
			c.Next()
			return
		}

		// For safe methods, generate and set token
		if ds.isSafeMethod(c.Request.Method) {
			token, err := ds.generateToken()
			if err != nil {
				ds.config.ErrorHandler(c, fmt.Errorf("failed to generate CSRF token: %w", err))
				return
			}
			ds.setTokenCookie(c, token)
			c.Set("csrf_token", token)
			c.Next()
			return
		}

		// For unsafe methods, validate token
		if err := ds.validateDoubleSubmitToken(c); err != nil {
			ds.config.ErrorHandler(c, err)
			return
		}

		c.Next()
	}
}

// generateToken generates a random token
func (ds *DoubleSubmitCSRF) generateToken() (string, error) {
	bytes := make([]byte, ds.config.TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// validateDoubleSubmitToken validates double submit token
func (ds *DoubleSubmitCSRF) validateDoubleSubmitToken(c *gin.Context) error {
	// Get token from cookie
	cookieToken, err := c.Cookie(ds.config.CookieName)
	if err != nil {
		return fmt.Errorf("CSRF cookie not found")
	}

	// Get token from header or form
	var requestToken string
	if token := c.GetHeader(ds.config.HeaderName); token != "" {
		requestToken = token
	} else if token := c.PostForm(ds.config.FormFieldName); token != "" {
		requestToken = token
	} else {
		return fmt.Errorf("CSRF token not found in request")
	}

	// Compare tokens using constant time comparison
	if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(requestToken)) != 1 {
		return fmt.Errorf("CSRF token mismatch")
	}

	return nil
}

// setTokenCookie sets the CSRF token cookie
func (ds *DoubleSubmitCSRF) setTokenCookie(c *gin.Context, token string) {
	c.SetSameSite(ds.config.SameSite)
	c.SetCookie(
		ds.config.CookieName,
		token,
		0, // Session cookie
		"/",
		"",
		ds.config.SecureCookie,
		false, // Not HttpOnly for double submit pattern
	)
}

// isSafeMethod checks if HTTP method is safe
func (ds *DoubleSubmitCSRF) isSafeMethod(method string) bool {
	safeMethods := []string{"GET", "HEAD", "OPTIONS", "TRACE"}
	for _, safe := range safeMethods {
		if method == safe {
			return true
		}
	}
	return false
}

// Convenience functions

// NewDefaultCSRF creates CSRF middleware with default settings
func NewDefaultCSRF() *CSRFMiddleware {
	return NewCSRFMiddleware(CSRFConfig{
		SecureCookie: true,
		SameSite:     http.SameSiteStrictMode,
	})
}

// NewDoubleSubmitCSRFDefault creates double submit CSRF with default settings
func NewDoubleSubmitCSRFDefault() *DoubleSubmitCSRF {
	return NewDoubleSubmitCSRF(CSRFConfig{
		SecureCookie: true,
		SameSite:     http.SameSiteStrictMode,
	})
}

// SkipCSRFForAPI skips CSRF protection for API endpoints
func SkipCSRFForAPI(c *gin.Context) bool {
	return strings.HasPrefix(c.Request.URL.Path, "/api/")
}

// SkipCSRFForPublicEndpoints skips CSRF for public endpoints
func SkipCSRFForPublicEndpoints(c *gin.Context) bool {
	publicPaths := []string{
		"/health",
		"/metrics",
		"/api/verify/",
		"/api/auth/login",
		"/api/auth/register",
	}
	
	path := c.Request.URL.Path
	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	
	return false
}