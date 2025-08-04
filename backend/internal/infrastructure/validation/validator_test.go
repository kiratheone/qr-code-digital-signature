package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateAndSanitizeString(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		field    string
		value    string
		minLen   int
		maxLen   int
		required bool
		want     string
		wantErr  bool
		errCode  string
	}{
		{
			name:     "valid string",
			field:    "name",
			value:    "John Doe",
			minLen:   1,
			maxLen:   50,
			required: true,
			want:     "John Doe",
			wantErr:  false,
		},
		{
			name:     "empty required field",
			field:    "name",
			value:    "",
			minLen:   1,
			maxLen:   50,
			required: true,
			want:     "",
			wantErr:  true,
			errCode:  "REQUIRED",
		},
		{
			name:     "empty optional field",
			field:    "name",
			value:    "",
			minLen:   1,
			maxLen:   50,
			required: false,
			want:     "",
			wantErr:  false,
		},
		{
			name:     "string too short",
			field:    "name",
			value:    "Jo",
			minLen:   5,
			maxLen:   50,
			required: true,
			want:     "",
			wantErr:  true,
			errCode:  "MIN_LENGTH",
		},
		{
			name:     "string too long",
			field:    "name",
			value:    "This is a very long string that exceeds the maximum length",
			minLen:   1,
			maxLen:   10,
			required: true,
			want:     "",
			wantErr:  true,
			errCode:  "MAX_LENGTH",
		},
		{
			name:     "string with HTML",
			field:    "name",
			value:    "John <script>alert('xss')</script> Doe",
			minLen:   1,
			maxLen:   100,
			required: true,
			want:     "",
			wantErr:  true,
			errCode:  "SUSPICIOUS_CONTENT",
		},
		{
			name:     "string with SQL injection",
			field:    "name",
			value:    "John'; DROP TABLE users; --",
			minLen:   1,
			maxLen:   100,
			required: true,
			want:     "",
			wantErr:  true,
			errCode:  "SUSPICIOUS_CONTENT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validator.ValidateAndSanitizeString(tt.field, tt.value, tt.minLen, tt.maxLen, tt.required)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errCode != "" && err.Code != tt.errCode {
					t.Errorf("Expected error code %s but got %s", tt.errCode, err.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Expected %s but got %s", tt.want, got)
				}
			}
		})
	}
}

func TestValidator_ValidateEmail(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		field    string
		email    string
		required bool
		want     string
		wantErr  bool
		errCode  string
	}{
		{
			name:     "valid email",
			field:    "email",
			email:    "john.doe@example.com",
			required: true,
			want:     "john.doe@example.com",
			wantErr:  false,
		},
		{
			name:     "email with uppercase",
			field:    "email",
			email:    "John.Doe@EXAMPLE.COM",
			required: true,
			want:     "john.doe@example.com",
			wantErr:  false,
		},
		{
			name:     "invalid email format",
			field:    "email",
			email:    "invalid-email",
			required: true,
			want:     "",
			wantErr:  true,
			errCode:  "INVALID_EMAIL",
		},
		{
			name:     "empty required email",
			field:    "email",
			email:    "",
			required: true,
			want:     "",
			wantErr:  true,
			errCode:  "REQUIRED",
		},
		{
			name:     "empty optional email",
			field:    "email",
			email:    "",
			required: false,
			want:     "",
			wantErr:  false,
		},
		{
			name:     "email too long",
			field:    "email",
			email:    strings.Repeat("a", 250) + "@example.com",
			required: true,
			want:     "",
			wantErr:  true,
			errCode:  "MAX_LENGTH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validator.ValidateEmail(tt.field, tt.email, tt.required)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errCode != "" && err.Code != tt.errCode {
					t.Errorf("Expected error code %s but got %s", tt.errCode, err.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Expected %s but got %s", tt.want, got)
				}
			}
		})
	}
}

func TestValidator_ValidatePassword(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errCode  string
	}{
		{
			name:     "valid strong password",
			password: "MyStr0ng!Pass",
			wantErr:  false,
		},
		{
			name:     "password too short",
			password: "Short1!",
			wantErr:  true,
			errCode:  "MIN_LENGTH",
		},
		{
			name:     "password too long",
			password: strings.Repeat("a", 130) + "A1!",
			wantErr:  true,
			errCode:  "MAX_LENGTH",
		},
		{
			name:     "password without uppercase",
			password: "mystr0ng!pass",
			wantErr:  true,
			errCode:  "WEAK_PASSWORD",
		},
		{
			name:     "password without lowercase",
			password: "MYSTR0NG!PASS",
			wantErr:  true,
			errCode:  "WEAK_PASSWORD",
		},
		{
			name:     "password without digit",
			password: "MyStrong!Pass",
			wantErr:  true,
			errCode:  "WEAK_PASSWORD",
		},
		{
			name:     "password without special character",
			password: "MyStr0ngPass",
			wantErr:  true,
			errCode:  "WEAK_PASSWORD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePassword("password", tt.password)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errCode != "" && err.Code != tt.errCode {
					t.Errorf("Expected error code %s but got %s", tt.errCode, err.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidator_ValidateFilename(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		filename string
		required bool
		want     string
		wantErr  bool
		errCode  string
	}{
		{
			name:     "valid filename",
			filename: "document.pdf",
			required: true,
			want:     "document.pdf",
			wantErr:  false,
		},
		{
			name:     "filename with path traversal",
			filename: "../../../etc/passwd",
			required: true,
			want:     "etcpasswd",
			wantErr:  false,
		},
		{
			name:     "filename with invalid characters",
			filename: "document<script>.pdf",
			required: true,
			want:     "",
			wantErr:  true,
			errCode:  "INVALID_CHARACTERS",
		},
		{
			name:     "empty required filename",
			filename: "",
			required: true,
			want:     "",
			wantErr:  true,
			errCode:  "REQUIRED",
		},
		{
			name:     "filename too long",
			filename: strings.Repeat("a", 260) + ".pdf",
			required: true,
			want:     "",
			wantErr:  true,
			errCode:  "MAX_LENGTH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validator.ValidateFilename("filename", tt.filename, tt.required)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errCode != "" && err.Code != tt.errCode {
					t.Errorf("Expected error code %s but got %s", tt.errCode, err.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				if got != tt.want {
					t.Errorf("Expected %s but got %s", tt.want, got)
				}
			}
		})
	}
}

func TestValidator_ValidateFileSize(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		size    int64
		maxSize int64
		wantErr bool
		errCode string
	}{
		{
			name:    "valid file size",
			size:    1024,
			maxSize: 2048,
			wantErr: false,
		},
		{
			name:    "empty file",
			size:    0,
			maxSize: 2048,
			wantErr: true,
			errCode: "EMPTY_FILE",
		},
		{
			name:    "file too large",
			size:    3072,
			maxSize: 2048,
			wantErr: true,
			errCode: "FILE_TOO_LARGE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFileSize("file", tt.size, tt.maxSize)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errCode != "" && err.Code != tt.errCode {
					t.Errorf("Expected error code %s but got %s", tt.errCode, err.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidator_ValidateContentType(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name         string
		contentType  string
		allowedTypes []string
		wantErr      bool
		errCode      string
	}{
		{
			name:         "valid content type",
			contentType:  "application/pdf",
			allowedTypes: []string{"application/pdf", "image/jpeg"},
			wantErr:      false,
		},
		{
			name:         "invalid content type",
			contentType:  "text/html",
			allowedTypes: []string{"application/pdf", "image/jpeg"},
			wantErr:      true,
			errCode:      "INVALID_CONTENT_TYPE",
		},
		{
			name:         "empty content type",
			contentType:  "",
			allowedTypes: []string{"application/pdf"},
			wantErr:      true,
			errCode:      "REQUIRED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateContentType("content_type", tt.contentType, tt.allowedTypes)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errCode != "" && err.Code != tt.errCode {
					t.Errorf("Expected error code %s but got %s", tt.errCode, err.Code)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidator_SanitizeString(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal string",
			input: "Hello World",
			want:  "Hello World",
		},
		{
			name:  "string with HTML",
			input: "<script>alert('xss')</script>",
			want:  "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:  "string with whitespace",
			input: "  Hello World  ",
			want:  "Hello World",
		},
		{
			name:  "string with null bytes",
			input: "Hello\x00World",
			want:  "HelloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.SanitizeString(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidator_SanitizeFilename(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal filename",
			input: "document.pdf",
			want:  "document.pdf",
		},
		{
			name:  "filename with path traversal",
			input: "../../../etc/passwd",
			want:  "etcpasswd",
		},
		{
			name:  "filename with backslashes",
			input: "..\\..\\windows\\system32\\config",
			want:  "windowssystem32config",
		},
		{
			name:  "filename with leading/trailing dots",
			input: "...document.pdf...",
			want:  "document.pdf",
		},
		{
			name:  "empty filename after sanitization",
			input: "../..",
			want:  "unnamed_file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.SanitizeFilename(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidator_containsSuspiciousContent(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "normal content",
			input: "Hello World",
			want:  false,
		},
		{
			name:  "SQL injection",
			input: "'; DROP TABLE users; --",
			want:  true,
		},
		{
			name:  "XSS attempt",
			input: "<script>alert('xss')</script>",
			want:  true,
		},
		{
			name:  "path traversal",
			input: "../../../etc/passwd",
			want:  true,
		},
		{
			name:  "javascript protocol",
			input: "javascript:alert('xss')",
			want:  true,
		},
		{
			name:  "excessive control characters",
			input: "Hello\x01\x02\x03\x04\x05\x06World",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.containsSuspiciousContent(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}