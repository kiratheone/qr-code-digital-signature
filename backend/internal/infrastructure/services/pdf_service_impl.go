package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"digital-signature-system/internal/domain/services"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/model"
)

// PDFServiceImpl implements the PDFService interface
type PDFServiceImpl struct {
	maxFileSize int64 // Maximum file size in bytes (default: 50MB)
	bufferSize  int   // Buffer size for streaming operations (default: 64KB)
}

// NewPDFService creates a new instance of PDFServiceImpl
func NewPDFService(maxFileSize int64) services.PDFService {
	// If maxFileSize is not provided or invalid, set default to 50MB
	if maxFileSize <= 0 {
		maxFileSize = 50 * 1024 * 1024 // 50MB in bytes
	}

	return &PDFServiceImpl{
		maxFileSize: maxFileSize,
		bufferSize:  64 * 1024, // 64KB buffer for streaming
	}
}

// ValidatePDF validates if the provided data is a valid PDF file
func (s *PDFServiceImpl) ValidatePDF(ctx context.Context, pdfData []byte) error {
	// Check file size
	if int64(len(pdfData)) > s.maxFileSize {
		return fmt.Errorf("PDF file size exceeds maximum allowed size of %d bytes", s.maxFileSize)
	}

	return s.validatePDFContent(ctx, pdfData)
}

// ValidatePDFFromReader validates PDF from an io.Reader for streaming
func (s *PDFServiceImpl) ValidatePDFFromReader(ctx context.Context, reader io.Reader, size int64) error {
	// Check file size
	if size > s.maxFileSize {
		return fmt.Errorf("PDF file size exceeds maximum allowed size of %d bytes", s.maxFileSize)
	}

	// For streaming validation, we need to read the data
	// Use a limited reader to prevent memory issues
	limitedReader := io.LimitReader(reader, s.maxFileSize)
	pdfData, err := io.ReadAll(limitedReader)
	if err != nil {
		return fmt.Errorf("error reading PDF data: %w", err)
	}

	return s.validatePDFContent(ctx, pdfData)
}

// validatePDFContent validates the actual PDF content
func (s *PDFServiceImpl) validatePDFContent(ctx context.Context, pdfData []byte) error {
	// Check if data is a valid PDF by attempting to parse it
	reader, err := model.NewPdfReader(bytes.NewReader(pdfData))
	if err != nil {
		return fmt.Errorf("invalid PDF format: %w", err)
	}

	// Check if the PDF is encrypted
	isEncrypted, err := reader.IsEncrypted()
	if err != nil {
		return fmt.Errorf("error checking PDF encryption: %w", err)
	}

	if isEncrypted {
		return errors.New("encrypted PDFs are not supported")
	}

	return nil
}

// CalculateHash calculates SHA-256 hash for the provided PDF document
func (s *PDFServiceImpl) CalculateHash(ctx context.Context, pdfData []byte) ([]byte, error) {
	// Validate PDF first
	if err := s.ValidatePDF(ctx, pdfData); err != nil {
		return nil, err
	}

	// Calculate SHA-256 hash using streaming for large files
	return s.calculateHashStreaming(ctx, bytes.NewReader(pdfData))
}

// CalculateHashFromReader calculates SHA-256 hash from an io.Reader for streaming
func (s *PDFServiceImpl) CalculateHashFromReader(ctx context.Context, reader io.Reader) ([]byte, error) {
	return s.calculateHashStreaming(ctx, reader)
}

// calculateHashStreaming performs streaming hash calculation
func (s *PDFServiceImpl) calculateHashStreaming(ctx context.Context, reader io.Reader) ([]byte, error) {
	hash := sha256.New()
	buffer := make([]byte, s.bufferSize)
	
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		
		n, err := reader.Read(buffer)
		if n > 0 {
			if _, writeErr := hash.Write(buffer[:n]); writeErr != nil {
				return nil, fmt.Errorf("error writing to hash: %w", writeErr)
			}
		}
		
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading data: %w", err)
		}
	}

	return hash.Sum(nil), nil
}

// GetPDFInfo extracts basic information from the PDF document
func (s *PDFServiceImpl) GetPDFInfo(ctx context.Context, pdfData []byte) (*services.PDFInfo, error) {
	// Validate PDF first
	if err := s.ValidatePDF(ctx, pdfData); err != nil {
		return nil, err
	}

	// Parse PDF
	reader, err := model.NewPdfReader(bytes.NewReader(pdfData))
	if err != nil {
		return nil, fmt.Errorf("error parsing PDF: %w", err)
	}

	// Get number of pages
	numPages, err := reader.GetNumPages()
	if err != nil {
		return nil, fmt.Errorf("error getting page count: %w", err)
	}

	// Get document metadata
	metadata, err := reader.GetPdfInfo()
	if err != nil {
		// If metadata can't be read, continue with empty metadata
		metadata = &model.PdfInfo{}
	}

	// Check if encrypted
	isEncrypted, _ := reader.IsEncrypted()

	// Create PDFInfo
	info := &services.PDFInfo{
		PageCount:   numPages,
		Title:       metadata.Title.String(),
		Author:      metadata.Author.String(),
		Creator:     metadata.Creator.String(),
		Producer:    metadata.Producer.String(),
		FileSize:    int64(len(pdfData)),
		IsEncrypted: isEncrypted,
	}

	return info, nil
}

// init initializes the UniDoc license if available
func init() {
	// Try to set license from environment variable first
	if licenseKey := os.Getenv("UNIDOC_LICENSE_KEY"); licenseKey != "" {
		err := license.SetMeteredKey(licenseKey)
		if err != nil {
			fmt.Printf("Error setting UniDoc license from env: %v\n", err)
		}
		return
	}
	
	// For development/testing, we can continue without a license
	// This will add watermarks but the functionality will work
	fmt.Println("No UniDoc license key provided - continuing with unlicensed version (will have watermarks)")
}