package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a simple file-based logger
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	level       LogLevel
}

var defaultLogger *Logger

// Initialize sets up the default logger with file-based logging
func Initialize(logDir string, level LogLevel) error {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log files
	infoLogFile := filepath.Join(logDir, "app.log")
	errorLogFile := filepath.Join(logDir, "error.log")

	// Open log files
	infoFile, err := os.OpenFile(infoLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open info log file: %w", err)
	}

	errorFile, err := os.OpenFile(errorLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open error log file: %w", err)
	}

	// Create multi-writers to write to both file and stdout
	infoWriter := io.MultiWriter(os.Stdout, infoFile)
	errorWriter := io.MultiWriter(os.Stderr, errorFile)

	// Create loggers
	infoLogger := log.New(infoWriter, "", log.LstdFlags)
	errorLogger := log.New(errorWriter, "", log.LstdFlags)

	defaultLogger = &Logger{
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
		level:       level,
	}

	return nil
}

// GetLogger returns the default logger instance
func GetLogger() *Logger {
	if defaultLogger == nil {
		// Fallback to stdout/stderr if not initialized
		defaultLogger = &Logger{
			infoLogger:  log.New(os.Stdout, "", log.LstdFlags),
			errorLogger: log.New(os.Stderr, "", log.LstdFlags),
			level:       INFO,
		}
	}
	return defaultLogger
}

// logf formats and logs a message at the specified level
func (l *Logger) logf(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logEntry := fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), message)

	if level >= ERROR {
		l.errorLogger.Println(logEntry)
	} else {
		l.infoLogger.Println(logEntry)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.logf(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.logf(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.logf(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.logf(ERROR, format, args...)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.logf(FATAL, format, args...)
	os.Exit(1)
}

// Package-level convenience functions
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	GetLogger().Fatal(format, args...)
}