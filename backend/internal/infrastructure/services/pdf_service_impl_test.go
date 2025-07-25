package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestPDFServiceImpl_ValidatePDF(t *testing.T) {
	// Create a new PDF service with 5MB max file size for testing
	pdfService := NewPDFService(5 * 1024 * 1024)
	ctx := context.Background()

	tests := []struct {
		name     string
		pdfData  []byte
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Empty data",
			pdfData:  []byte{},
			wantErr:  true,
			errMsg:   "invalid PDF format",
		},
		{
			name:     "Invalid PDF data",
			pdfData:  []byte("This is not a PDF file"),
			wantErr:  true,
			errMsg:   "invalid PDF format",
		},
		{
			name:     "File too large",
			pdfData:  make([]byte, 6*1024*1024), // 6MB (exceeds 5MB limit)
			wantErr:  true,
			errMsg:   "exceeds maximum allowed size",
		},
		{
			name:     "Valid PDF",
			pdfData:  MinimalValidPDF,
			wantErr:  false,
			errMsg:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pdfService.ValidatePDF(ctx, tt.pdfData)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePDF() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePDF() error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestPDFServiceImpl_CalculateHash(t *testing.T) {
	// Create a new PDF service
	pdfService := NewPDFService(50 * 1024 * 1024)
	ctx := context.Background()

	// Test with invalid PDF (should fail validation)
	t.Run("Invalid PDF", func(t *testing.T) {
		invalidData := []byte("This is not a PDF file")
		_, err := pdfService.CalculateHash(ctx, invalidData)
		if err == nil {
			t.Errorf("CalculateHash() should fail with invalid PDF")
		}
	})

	// Test with valid PDF
	t.Run("Valid PDF", func(t *testing.T) {
		// Calculate expected hash
		expectedHash := sha256.Sum256(MinimalValidPDF)
		
		// Calculate hash using service
		hash, err := pdfService.CalculateHash(ctx, MinimalValidPDF)
		if err != nil {
			t.Errorf("CalculateHash() error = %v", err)
			return
		}
		
		// Compare hashes
		if hex.EncodeToString(hash) != hex.EncodeToString(expectedHash[:]) {
			t.Errorf("CalculateHash() = %v, want %v", 
				hex.EncodeToString(hash), 
				hex.EncodeToString(expectedHash[:]))
		}
	})
}

func TestPDFServiceImpl_GetPDFInfo(t *testing.T) {
	// Create a new PDF service
	pdfService := NewPDFService(50 * 1024 * 1024)
	ctx := context.Background()

	// Test with invalid PDF
	t.Run("Invalid PDF", func(t *testing.T) {
		invalidData := []byte("This is not a PDF")
		_, err := pdfService.GetPDFInfo(ctx, invalidData)
		if err == nil {
			t.Errorf("GetPDFInfo() should fail with invalid PDF")
		}
	})

	// Test with valid PDF
	t.Run("Valid PDF", func(t *testing.T) {
		info, err := pdfService.GetPDFInfo(ctx, MinimalValidPDF)
		if err != nil {
			t.Errorf("GetPDFInfo() error = %v", err)
			return
		}
		
		// Check basic info
		if info.PageCount <= 0 {
			t.Errorf("GetPDFInfo() PageCount = %v, want > 0", info.PageCount)
		}
		
		if info.FileSize != int64(len(MinimalValidPDF)) {
			t.Errorf("GetPDFInfo() FileSize = %v, want %v", info.FileSize, len(MinimalValidPDF))
		}
	})
}