package services

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// LogRotationPolicy defines log rotation and retention policies
type LogRotationPolicy struct {
	MaxFileSize    int64         // Maximum file size before rotation (bytes)
	MaxFiles       int           // Maximum number of files to keep
	MaxAge         time.Duration // Maximum age of log files
	RotationTime   time.Duration // Time-based rotation interval
	CompressOld    bool          // Whether to compress old log files
	DeleteOnRotate bool          // Whether to delete files immediately when they exceed MaxFiles
}

// DefaultLogRotationPolicy returns a default log rotation policy
func DefaultLogRotationPolicy() LogRotationPolicy {
	return LogRotationPolicy{
		MaxFileSize:    100 * 1024 * 1024, // 100MB
		MaxFiles:       30,                 // Keep 30 files
		MaxAge:         30 * 24 * time.Hour, // 30 days
		RotationTime:   24 * time.Hour,     // Daily rotation
		CompressOld:    true,               // Compress old files
		DeleteOnRotate: true,               // Delete old files when limit exceeded
	}
}

// fileInfo represents information about a log file
type fileInfo struct {
	path    string
	modTime time.Time
	size    int64
}

// LogRotationManager manages log file rotation and retention
type LogRotationManager struct {
	policy    LogRotationPolicy
	logDir    string
	filePrefix string
}

// NewLogRotationManager creates a new log rotation manager
func NewLogRotationManager(logDir, filePrefix string, policy LogRotationPolicy) *LogRotationManager {
	return &LogRotationManager{
		policy:     policy,
		logDir:     logDir,
		filePrefix: filePrefix,
	}
}

// ShouldRotate checks if the current log file should be rotated
func (lrm *LogRotationManager) ShouldRotate(currentFile *os.File) (bool, string) {
	if currentFile == nil {
		return true, "no_current_file"
	}

	// Check file size
	if stat, err := currentFile.Stat(); err == nil {
		if stat.Size() >= lrm.policy.MaxFileSize {
			return true, "file_size_exceeded"
		}
	}

	// Check time-based rotation
	filename := filepath.Base(currentFile.Name())
	expectedFilename := lrm.generateFilename(time.Now())
	
	if filename != expectedFilename {
		return true, "time_based_rotation"
	}

	return false, ""
}

// RotateLog performs log rotation
func (lrm *LogRotationManager) RotateLog(currentFile *os.File) (*os.File, error) {
	// Close current file if it exists
	if currentFile != nil {
		currentFile.Close()
	}

	// Clean up old files before creating new one
	if err := lrm.CleanupOldFiles(); err != nil {
		// Log error but don't fail rotation
		fmt.Printf("Warning: Failed to cleanup old log files: %v\n", err)
	}

	// Create new log file
	newFilename := lrm.generateFilename(time.Now())
	newFilePath := filepath.Join(lrm.logDir, newFilename)

	newFile, err := os.OpenFile(newFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create new log file: %w", err)
	}

	return newFile, nil
}

// CleanupOldFiles removes old log files based on retention policy
func (lrm *LogRotationManager) CleanupOldFiles() error {
	// Get all log files
	pattern := filepath.Join(lrm.logDir, lrm.filePrefix+"_*.log*")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to list log files: %w", err)
	}

	if len(files) == 0 {
		return nil
	}

	// Get file info for all files
	var fileInfos []fileInfo
	now := time.Now()

	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			continue // Skip files we can't stat
		}

		fileInfos = append(fileInfos, fileInfo{
			path:    file,
			modTime: stat.ModTime(),
			size:    stat.Size(),
		})
	}

	// Sort files by modification time (oldest first)
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime.Before(fileInfos[j].modTime)
	})

	var filesToDelete []string

	// Delete files based on age policy
	if lrm.policy.MaxAge > 0 {
		cutoffTime := now.Add(-lrm.policy.MaxAge)
		for _, info := range fileInfos {
			if info.modTime.Before(cutoffTime) {
				filesToDelete = append(filesToDelete, info.path)
			}
		}
	}

	// Delete files based on count policy
	if lrm.policy.MaxFiles > 0 && len(fileInfos) > lrm.policy.MaxFiles {
		// Add oldest files to deletion list
		excessCount := len(fileInfos) - lrm.policy.MaxFiles
		for i := 0; i < excessCount && i < len(fileInfos); i++ {
			// Avoid duplicates
			found := false
			for _, deleteFile := range filesToDelete {
				if deleteFile == fileInfos[i].path {
					found = true
					break
				}
			}
			if !found {
				filesToDelete = append(filesToDelete, fileInfos[i].path)
			}
		}
	}

	// Perform deletions
	for _, file := range filesToDelete {
		if err := os.Remove(file); err != nil {
			fmt.Printf("Warning: Failed to delete old log file %s: %v\n", file, err)
		}
	}

	// Compress remaining old files if policy is enabled
	if lrm.policy.CompressOld {
		if err := lrm.compressOldFiles(fileInfos, filesToDelete); err != nil {
			fmt.Printf("Warning: Failed to compress old log files: %v\n", err)
		}
	}

	return nil
}

// compressOldFiles compresses old log files (placeholder implementation)
func (lrm *LogRotationManager) compressOldFiles(fileInfos []fileInfo, deletedFiles []string) error {
	// This is a placeholder for compression logic
	// In a real implementation, you would use gzip or another compression library
	
	for _, info := range fileInfos {
		// Skip files that were deleted
		shouldSkip := false
		for _, deleted := range deletedFiles {
			if deleted == info.path {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			continue
		}

		// Skip current day's file
		if strings.Contains(info.path, time.Now().Format("2006-01-02")) {
			continue
		}

		// Skip already compressed files
		if strings.HasSuffix(info.path, ".gz") {
			continue
		}

		// Here you would implement actual compression
		// For now, we just log what would be compressed
		fmt.Printf("Would compress log file: %s\n", info.path)
	}

	return nil
}

// generateFilename generates a filename based on current time
func (lrm *LogRotationManager) generateFilename(t time.Time) string {
	return fmt.Sprintf("%s_%s.log", lrm.filePrefix, t.Format("2006-01-02"))
}

// GetLogFileStats returns statistics about log files
func (lrm *LogRotationManager) GetLogFileStats() (map[string]interface{}, error) {
	pattern := filepath.Join(lrm.logDir, lrm.filePrefix+"_*.log*")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list log files: %w", err)
	}

	stats := map[string]interface{}{
		"total_files": len(files),
		"log_dir":     lrm.logDir,
		"file_prefix": lrm.filePrefix,
		"policy":      lrm.policy,
	}

	if len(files) == 0 {
		stats["total_size_bytes"] = int64(0)
		stats["oldest_file"] = nil
		stats["newest_file"] = nil
		return stats, nil
	}

	var totalSize int64
	var oldestTime, newestTime time.Time
	var oldestFile, newestFile string

	for i, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			continue
		}

		totalSize += stat.Size()

		if i == 0 {
			oldestTime = stat.ModTime()
			newestTime = stat.ModTime()
			oldestFile = file
			newestFile = file
		} else {
			if stat.ModTime().Before(oldestTime) {
				oldestTime = stat.ModTime()
				oldestFile = file
			}
			if stat.ModTime().After(newestTime) {
				newestTime = stat.ModTime()
				newestFile = file
			}
		}
	}

	stats["total_size_bytes"] = totalSize
	stats["total_size_mb"] = float64(totalSize) / (1024 * 1024)
	stats["oldest_file"] = map[string]interface{}{
		"path":     oldestFile,
		"mod_time": oldestTime,
	}
	stats["newest_file"] = map[string]interface{}{
		"path":     newestFile,
		"mod_time": newestTime,
	}

	return stats, nil
}

// ValidatePolicy validates the log rotation policy
func (lrm *LogRotationManager) ValidatePolicy() error {
	if lrm.policy.MaxFileSize <= 0 {
		return fmt.Errorf("MaxFileSize must be greater than 0")
	}

	if lrm.policy.MaxFiles < 0 {
		return fmt.Errorf("MaxFiles must be non-negative")
	}

	if lrm.policy.MaxAge < 0 {
		return fmt.Errorf("MaxAge must be non-negative")
	}

	if lrm.policy.RotationTime <= 0 {
		return fmt.Errorf("RotationTime must be greater than 0")
	}

	// Warn about potentially problematic configurations
	if lrm.policy.MaxFiles == 0 && lrm.policy.MaxAge == 0 {
		fmt.Println("Warning: Both MaxFiles and MaxAge are 0, logs will never be cleaned up")
	}

	if lrm.policy.MaxFileSize > 1024*1024*1024 { // 1GB
		fmt.Println("Warning: MaxFileSize is very large (>1GB), consider smaller values")
	}

	return nil
}

// LogRetentionReport generates a report about log retention
type LogRetentionReport struct {
	TotalFiles       int           `json:"total_files"`
	TotalSizeBytes   int64         `json:"total_size_bytes"`
	TotalSizeMB      float64       `json:"total_size_mb"`
	OldestFileAge    time.Duration `json:"oldest_file_age"`
	NewestFileAge    time.Duration `json:"newest_file_age"`
	FilesOverAge     int           `json:"files_over_age"`
	FilesOverSize    int           `json:"files_over_size"`
	RecommendCleanup bool          `json:"recommend_cleanup"`
	Policy           LogRotationPolicy `json:"policy"`
}

// GenerateRetentionReport generates a comprehensive retention report
func (lrm *LogRotationManager) GenerateRetentionReport() (*LogRetentionReport, error) {
	pattern := filepath.Join(lrm.logDir, lrm.filePrefix+"_*.log*")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list log files: %w", err)
	}

	report := &LogRetentionReport{
		TotalFiles: len(files),
		Policy:     lrm.policy,
	}

	if len(files) == 0 {
		return report, nil
	}

	now := time.Now()
	var totalSize int64
	var oldestTime, newestTime time.Time
	var filesOverAge, filesOverSize int

	for i, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			continue
		}

		totalSize += stat.Size()

		if i == 0 {
			oldestTime = stat.ModTime()
			newestTime = stat.ModTime()
		} else {
			if stat.ModTime().Before(oldestTime) {
				oldestTime = stat.ModTime()
			}
			if stat.ModTime().After(newestTime) {
				newestTime = stat.ModTime()
			}
		}

		// Check age policy
		if lrm.policy.MaxAge > 0 && now.Sub(stat.ModTime()) > lrm.policy.MaxAge {
			filesOverAge++
		}

		// Check size policy
		if stat.Size() > lrm.policy.MaxFileSize {
			filesOverSize++
		}
	}

	report.TotalSizeBytes = totalSize
	report.TotalSizeMB = float64(totalSize) / (1024 * 1024)
	report.OldestFileAge = now.Sub(oldestTime)
	report.NewestFileAge = now.Sub(newestTime)
	report.FilesOverAge = filesOverAge
	report.FilesOverSize = filesOverSize

	// Recommend cleanup if there are violations
	report.RecommendCleanup = filesOverAge > 0 || 
		(lrm.policy.MaxFiles > 0 && len(files) > lrm.policy.MaxFiles)

	return report, nil
}