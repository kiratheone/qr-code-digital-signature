package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestStandardError(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		message  string
		details  []string
		expected string
	}{
		{
			name:     "error without details",
			code:     ErrCodeInvalidRequest,
			message:  "Invalid request format",
			details:  nil,
			expected: "INVALID_REQUEST: Invalid request format",
		},
		{
			name:     "error with details",
			code:     ErrCodeValidationFailed,
			message:  "Validation failed",
			details:  []string{"Field 'username' is required"},
			expected: "VALIDATION_FAILED: Validation failed (Field 'username' is required)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewStandardError(tt.code, tt.message, tt.details...)
			assert.Equal(t, tt.expected, err.Error())
			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, tt.message, err.Message)
		})
	}
}

func TestRespondWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		statusCode     int
		err            *StandardError
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "validation error",
			statusCode:     http.StatusBadRequest,
			err:            NewStandardError(ErrCodeValidationFailed, "Invalid input"),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"code":"VALIDATION_FAILED","message":"Invalid input"}`,
		},
		{
			name:           "internal error",
			statusCode:     http.StatusInternalServerError,
			err:            NewStandardError(ErrCodeInternalError, "Something went wrong"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"code":"INTERNAL_ERROR","message":"Something went wrong"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			RespondWithError(c, tt.statusCode, tt.err)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestMapServiceErrorToHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		serviceError   error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "invalid credentials",
			serviceError:   errors.New("invalid credentials"),
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   ErrCodeUnauthorized,
		},
		{
			name:           "user not found",
			serviceError:   errors.New("user not found"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   ErrCodeNotFound,
		},
		{
			name:           "document not found",
			serviceError:   errors.New("document not found"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   ErrCodeNotFound,
		},
		{
			name:           "access denied",
			serviceError:   errors.New("access denied: document belongs to different user"),
			expectedStatus: http.StatusForbidden,
			expectedCode:   ErrCodeForbidden,
		},
		{
			name:           "unknown error",
			serviceError:   errors.New("some unknown error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   ErrCodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			MapServiceErrorToHTTP(c, tt.serviceError)

			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response to check error code
			var response StandardError
			err := c.ShouldBindJSON(&response)
			if err == nil {
				assert.Equal(t, tt.expectedCode, response.Code)
			}
		})
	}
}

func TestRespondWithValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondWithValidationError(c, "Invalid input", "Field is required")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_FAILED")
	assert.Contains(t, w.Body.String(), "Invalid input")
}

func TestRespondWithUnauthorizedError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondWithUnauthorizedError(c, "Token required")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "UNAUTHORIZED")
	assert.Contains(t, w.Body.String(), "Token required")
}

func TestRespondWithInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondWithInternalError(c, "Database error")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "INTERNAL_ERROR")
	assert.Contains(t, w.Body.String(), "Database error")
}