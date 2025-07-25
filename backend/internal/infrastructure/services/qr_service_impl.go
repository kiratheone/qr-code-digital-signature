package services

import (
	"context"
	"digital-signature-system/internal/domain/services"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"bytes"

	"github.com/skip2/go-qrcode"
	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/creator"
	"github.com/unidoc/unipdf/v3/model"
)

// QRServiceImpl implements the QRService interface
type QRServiceImpl struct {
	// Default QR code size in pixels
	defaultQRSize int
	// Default QR code position margin in points (1/72 inch)
	defaultMargin float64
	// Base URL for verification
	verificationBaseURL string
}

// NewQRService creates a new instance of QRServiceImpl
func NewQRService(verificationBaseURL string) services.QRService {
	return &QRServiceImpl{
		defaultQRSize:       256,
		defaultMargin:       20.0,
		verificationBaseURL: verificationBaseURL,
	}
}

// GenerateQRCode generates a QR code image from the provided data
func (s *QRServiceImpl) GenerateQRCode(ctx context.Context, data *services.QRCodeData, size int) ([]byte, error) {
	if data == nil {
		return nil, errors.New("QR code data cannot be nil")
	}

	// Use default size if not specified or invalid
	if size <= 0 {
		size = s.defaultQRSize
	}

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling QR code data: %w", err)
	}

	// Generate QR code
	qr, err := qrcode.New(string(jsonData), qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("error generating QR code: %w", err)
	}

	// Set size
	qr.DisableBorder = false

	// Get PNG image
	pngData, err := qr.PNG(size)
	if err != nil {
		return nil, fmt.Errorf("error generating QR code PNG: %w", err)
	}

	return pngData, nil
}

// InjectQRCode injects a QR code into a PDF document at the specified position
func (s *QRServiceImpl) InjectQRCode(ctx context.Context, pdfData []byte, qrCodeData []byte, position *services.QRCodePosition) ([]byte, error) {
	if pdfData == nil || len(pdfData) == 0 {
		return nil, errors.New("PDF data cannot be empty")
	}

	if qrCodeData == nil || len(qrCodeData) == 0 {
		return nil, errors.New("QR code data cannot be empty")
	}

	// If position is nil, use default position (last page, bottom right)
	if position == nil {
		var err error
		position, err = s.GetDefaultPosition(ctx, pdfData)
		if err != nil {
			return nil, fmt.Errorf("error getting default position: %w", err)
		}
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

	// Determine target page
	targetPage := position.Page
	if targetPage < 0 || targetPage >= numPages {
		targetPage = numPages - 1 // Last page
	}

	// Create PDF creator
	c := creator.New()

	// Load all pages from the original PDF
	for i := 0; i < numPages; i++ {
		page, err := reader.GetPage(i + 1) // Pages are 1-indexed in UniDoc
		if err != nil {
			return nil, fmt.Errorf("error getting page %d: %w", i+1, err)
		}

		// Import page
		importedPage, err := c.NewPdfPageFromPage(page)
		if err != nil {
			return nil, fmt.Errorf("error importing page %d: %w", i+1, err)
		}

		// Add page to creator
		c.AddPage(importedPage)

		// If this is the target page, add QR code
		if i == targetPage {
			// Decode QR code image
			qrImg, _, err := image.Decode(bytes.NewReader(qrCodeData))
			if err != nil {
				return nil, fmt.Errorf("error decoding QR code image: %w", err)
			}

			// Create image component
			imgComp, err := c.NewImageFromGoImage(qrImg)
			if err != nil {
				return nil, fmt.Errorf("error creating image component: %w", err)
			}

			// Set position and size
			imgComp.SetPos(position.X, position.Y)
			if position.Width > 0 && position.Height > 0 {
				imgComp.ScaleToWidth(position.Width)
				imgComp.ScaleToHeight(position.Height)
			} else {
				// Default size if not specified
				imgComp.ScaleToWidth(100)
			}

			// Draw on the current page
			c.Draw(imgComp)
		}
	}

	// Generate output PDF
	var buf bytes.Buffer
	err = c.Write(&buf)
	if err != nil {
		return nil, fmt.Errorf("error writing PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// ParseQRCode parses a QR code image and extracts the encoded data
func (s *QRServiceImpl) ParseQRCode(ctx context.Context, qrCodeImage []byte) (*services.QRCodeData, error) {
	// This is a placeholder for actual QR code parsing
	// In a real implementation, you would use a QR code scanning library
	// For now, we'll just return an error
	return nil, errors.New("QR code parsing not implemented")
}

// GetDefaultPosition returns the default position for QR code injection (last page, bottom right)
func (s *QRServiceImpl) GetDefaultPosition(ctx context.Context, pdfData []byte) (*services.QRCodePosition, error) {
	if pdfData == nil || len(pdfData) == 0 {
		return nil, errors.New("PDF data cannot be empty")
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

	// Get last page
	page, err := reader.GetPage(numPages)
	if err != nil {
		return nil, fmt.Errorf("error getting last page: %w", err)
	}

	// Get page dimensions
	mediaBox, err := page.GetMediaBox()
	if err != nil {
		return nil, fmt.Errorf("error getting page dimensions: %w", err)
	}

	// Calculate position (bottom right with margin)
	width := 100.0  // QR code width in points
	height := 100.0 // QR code height in points
	x := mediaBox.Width() - width - s.defaultMargin
	y := s.defaultMargin // From bottom

	// Create position
	position := &services.QRCodePosition{
		Page:   numPages - 1, // 0-based index
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}

	return position, nil
}