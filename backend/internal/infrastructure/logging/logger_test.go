package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

func TestInitialize(t *testing.T) {
	// Create temporary directory for test logs
	tempDir := t.TempDir()

	err := Initialize(tempDir, INFO)
	require.NoError(t, err)

	// Check that log files are created
	infoLogPath := filepath.Join(tempDir, "app.log")
	errorLogPath := filepath.Join(tempDir, "error.log")

	_, err = os.Stat(infoLogPath)
	assert.NoError(t, err, "app.log should be created")

	_, err = os.Stat(errorLogPath)
	assert.NoError(t, err, "error.log should be created")

	// Test that logger is initialized
	logger := GetLogger()
	assert.NotNil(t, logger)
	assert.Equal(t, INFO, logger.level)
}

func TestLogger_LogLevels(t *testing.T) {
	// Create temporary directory for test logs
	tempDir := t.TempDir()

	err := Initialize(tempDir, DEBUG)
	require.NoError(t, err)

	logger := GetLogger()

	// Test different log levels
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")

	// Read log files to verify content
	infoLogPath := filepath.Join(tempDir, "app.log")
	errorLogPath := filepath.Join(tempDir, "error.log")

	infoContent, err := os.ReadFile(infoLogPath)
	require.NoError(t, err)

	errorContent, err := os.ReadFile(errorLogPath)
	require.NoError(t, err)

	infoStr := string(infoContent)
	errorStr := string(errorContent)

	// Check that messages appear in correct files
	assert.Contains(t, infoStr, "DEBUG: Debug message")
	assert.Contains(t, infoStr, "INFO: Info message")
	assert.Contains(t, infoStr, "WARN: Warning message")
	assert.Contains(t, errorStr, "ERROR: Error message")
}

func TestLogger_LogLevelFiltering(t *testing.T) {
	// Create temporary directory for test logs
	tempDir := t.TempDir()

	err := Initialize(tempDir, WARN)
	require.NoError(t, err)

	logger := GetLogger()

	// Log messages at different levels
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")

	// Read log files
	infoLogPath := filepath.Join(tempDir, "app.log")
	errorLogPath := filepath.Join(tempDir, "error.log")

	infoContent, err := os.ReadFile(infoLogPath)
	require.NoError(t, err)

	errorContent, err := os.ReadFile(errorLogPath)
	require.NoError(t, err)

	infoStr := string(infoContent)
	errorStr := string(errorContent)

	// Only WARN and ERROR should be logged (level is WARN)
	assert.NotContains(t, infoStr, "DEBUG: Debug message")
	assert.NotContains(t, infoStr, "INFO: Info message")
	assert.Contains(t, infoStr, "WARN: Warning message")
	assert.Contains(t, errorStr, "ERROR: Error message")
}

func TestPackageLevelFunctions(t *testing.T) {
	// Create temporary directory for test logs
	tempDir := t.TempDir()

	err := Initialize(tempDir, DEBUG)
	require.NoError(t, err)

	// Test package-level functions
	Debug("Package debug message")
	Info("Package info message")
	Warn("Package warning message")
	Error("Package error message")

	// Read log files
	infoLogPath := filepath.Join(tempDir, "app.log")
	errorLogPath := filepath.Join(tempDir, "error.log")

	infoContent, err := os.ReadFile(infoLogPath)
	require.NoError(t, err)

	errorContent, err := os.ReadFile(errorLogPath)
	require.NoError(t, err)

	infoStr := string(infoContent)
	errorStr := string(errorContent)

	assert.Contains(t, infoStr, "DEBUG: Package debug message")
	assert.Contains(t, infoStr, "INFO: Package info message")
	assert.Contains(t, infoStr, "WARN: Package warning message")
	assert.Contains(t, errorStr, "ERROR: Package error message")
}

func TestGetLogger_WithoutInitialization(t *testing.T) {
	// Reset the default logger
	defaultLogger = nil

	logger := GetLogger()
	assert.NotNil(t, logger)
	assert.Equal(t, INFO, logger.level)

	// Test that it works without initialization
	logger.Info("Test message without initialization")
}

func TestLogger_FormattedMessages(t *testing.T) {
	// Create temporary directory for test logs
	tempDir := t.TempDir()

	err := Initialize(tempDir, INFO)
	require.NoError(t, err)

	logger := GetLogger()

	// Test formatted messages
	logger.Info("User %s logged in with ID %d", "john_doe", 123)
	logger.Error("Database error: %v", "connection timeout")

	// Read log files
	infoLogPath := filepath.Join(tempDir, "app.log")
	errorLogPath := filepath.Join(tempDir, "error.log")

	infoContent, err := os.ReadFile(infoLogPath)
	require.NoError(t, err)

	errorContent, err := os.ReadFile(errorLogPath)
	require.NoError(t, err)

	infoStr := string(infoContent)
	errorStr := string(errorContent)

	assert.Contains(t, infoStr, "User john_doe logged in with ID 123")
	assert.Contains(t, errorStr, "Database error: connection timeout")
}

func TestInitialize_InvalidDirectory(t *testing.T) {
	// Try to initialize with an invalid directory path
	err := Initialize("/invalid/path/that/does/not/exist", INFO)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create log directory")
}

func TestLogger_TimestampFormat(t *testing.T) {
	// Create temporary directory for test logs
	tempDir := t.TempDir()

	err := Initialize(tempDir, INFO)
	require.NoError(t, err)

	logger := GetLogger()
	logger.Info("Timestamp test message")

	// Read log file
	infoLogPath := filepath.Join(tempDir, "app.log")
	infoContent, err := os.ReadFile(infoLogPath)
	require.NoError(t, err)

	infoStr := string(infoContent)

	// Check that timestamp format is present (YYYY-MM-DD HH:MM:SS)
	timestampPattern := `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`
	assert.Regexp(t, timestampPattern, infoStr)
}

func TestLogger_ConcurrentAccess(t *testing.T) {
	// Create temporary directory for test logs
	tempDir := t.TempDir()

	err := Initialize(tempDir, INFO)
	require.NoError(t, err)

	logger := GetLogger()

	// Test concurrent logging
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Info("Concurrent message from goroutine %d", id)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Read log file
	infoLogPath := filepath.Join(tempDir, "app.log")
	infoContent, err := os.ReadFile(infoLogPath)
	require.NoError(t, err)

	infoStr := string(infoContent)

	// Check that all messages are present
	for i := 0; i < 10; i++ {
		expectedMsg := "Concurrent message from goroutine"
		assert.Contains(t, infoStr, expectedMsg)
	}

	// Count the number of log entries
	lines := strings.Split(strings.TrimSpace(infoStr), "\n")
	assert.GreaterOrEqual(t, len(lines), 10)
}