package pdf

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/skip2/go-qrcode"
	"github.com/unidoc/unipdf/v3/creator"
	"github.com/unidoc/unipdf/v3/model"
)

const (
	MaxPDFSize = 50 * 1024 * 1024 // 50MB
)

// PDFService handles PDF processing operations
// Note: PDF modification operations (InjectQRCode) require a UniPDF license for production use.
// For development and testing, these operations will return license errors.
// Get a free trial license at https://unidoc.io
type PDFService struct{}

// NewPDFService creates a new PDF service instance
func NewPDFService() *PDFService {
	return &PDFService{}
}

// ValidatePDF validates if the provided data is a valid PDF
func (s *PDFService) ValidatePDF(pdfData []byte) error {
	if len(pdfData) == 0 {
		return fmt.Errorf("PDF data is empty")
	}

	if len(pdfData) > MaxPDFSize {
		return fmt.Errorf("PDF size exceeds maximum allowed size of %d bytes", MaxPDFSize)
	}

	// Check PDF header
	if len(pdfData) < 4 || string(pdfData[:4]) != "%PDF" {
		return fmt.Errorf("invalid PDF format: missing PDF header")
	}

	// Try to parse the PDF to ensure it's valid
	reader := bytes.NewReader(pdfData)
	_, err := model.NewPdfReader(reader)
	if err != nil {
		return fmt.Errorf("invalid PDF format: %w", err)
	}

	return nil
}

// CalculateHash calculates SHA-256 hash of the PDF document
func (s *PDFService) CalculateHash(pdfData []byte) ([]byte, error) {
	if err := s.ValidatePDF(pdfData); err != nil {
		return nil, fmt.Errorf("PDF validation failed: %w", err)
	}

	hash := sha256.Sum256(pdfData)
	return hash[:], nil
}

// GetPDFInfo extracts basic information from the PDF
func (s *PDFService) GetPDFInfo(pdfData []byte) (*PDFInfo, error) {
	if err := s.ValidatePDF(pdfData); err != nil {
		return nil, fmt.Errorf("PDF validation failed: %w", err)
	}

	reader := bytes.NewReader(pdfData)
	pdfReader, err := model.NewPdfReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF reader: %w", err)
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, fmt.Errorf("failed to get number of pages: %w", err)
	}

	info := &PDFInfo{
		NumPages: numPages,
		FileSize: int64(len(pdfData)),
	}

	// Note: Document metadata extraction would require additional unipdf API calls
	// For now, we focus on basic info (pages and file size) which are sufficient
	// for the core functionality

	return info, nil
}

// ReadPDFFromReader reads PDF data from an io.Reader
func (s *PDFService) ReadPDFFromReader(reader io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	
	// Limit the read to MaxPDFSize to prevent memory issues
	limitedReader := io.LimitReader(reader, MaxPDFSize+1)
	
	_, err := buf.ReadFrom(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF data: %w", err)
	}

	pdfData := buf.Bytes()
	
	// Check if the file exceeds the maximum size
	if len(pdfData) > MaxPDFSize {
		return nil, fmt.Errorf("PDF size exceeds maximum allowed size of %d bytes", MaxPDFSize)
	}

	return pdfData, nil
}

// GenerateQRCode generates a QR code image from the provided data
func (s *PDFService) GenerateQRCode(data QRCodeData) ([]byte, error) {
	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal QR code data: %w", err)
	}

	// Generate QR code with medium error correction level
	qrCode, err := qrcode.Encode(string(jsonData), qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	return qrCode, nil
}

// InjectQRCode injects a QR code into the PDF at the specified position
// Note: This requires a UniPDF license for PDF modification operations
func (s *PDFService) InjectQRCode(pdfData []byte, qrCodeData QRCodeData, position *QRPosition) ([]byte, error) {
	if err := s.ValidatePDF(pdfData); err != nil {
		return nil, fmt.Errorf("PDF validation failed: %w", err)
	}

	// Use default position if none provided
	if position == nil {
		defaultPos := DefaultQRPosition()
		position = &defaultPos
	}

	// Generate QR code image
	qrCodeImage, err := s.GenerateQRCode(qrCodeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Read the original PDF
	reader := bytes.NewReader(pdfData)
	pdfReader, err := model.NewPdfReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF reader: %w", err)
	}

	// Create a new PDF creator
	c := creator.New()

	// Get the number of pages
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, fmt.Errorf("failed to get number of pages: %w", err)
	}

	// Copy all pages to the new PDF
	for i := 1; i <= numPages; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get page %d: %w", i, err)
		}

		// Import the page
		err = c.AddPage(page)
		if err != nil {
			return nil, fmt.Errorf("failed to add page %d: %w", i, err)
		}

		// Add QR code to the last page
		if i == numPages {
			err = s.addQRCodeToPage(c, qrCodeImage, *position)
			if err != nil {
				return nil, fmt.Errorf("failed to add QR code to page: %w", err)
			}
		}
	}

	// Write the modified PDF to buffer
	var buf bytes.Buffer
	err = c.Write(&buf)
	if err != nil {
		// Handle license error gracefully for development
		if strings.Contains(err.Error(), "license") {
			return nil, fmt.Errorf("PDF modification requires UniPDF license: %w", err)
		}
		return nil, fmt.Errorf("failed to write modified PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// addQRCodeToPage adds a QR code image to the current page in the creator
func (s *PDFService) addQRCodeToPage(c *creator.Creator, qrCodeImage []byte, position QRPosition) error {
	// Create image from QR code bytes
	img, err := c.NewImageFromData(qrCodeImage)
	if err != nil {
		return fmt.Errorf("failed to create image from QR code data: %w", err)
	}

	// Set image position and size
	img.SetPos(position.X, position.Y)
	img.ScaleToWidth(position.Width)

	// Add image to the current page
	err = c.Draw(img)
	if err != nil {
		return fmt.Errorf("failed to draw QR code image: %w", err)
	}

	return nil
}

// PDFInfo contains basic information about a PDF document
type PDFInfo struct {
	NumPages int    `json:"num_pages"`
	FileSize int64  `json:"file_size"`
	Title    string `json:"title,omitempty"`
	Author   string `json:"author,omitempty"`
	Subject  string `json:"subject,omitempty"`
}

// QRCodeData contains the data to be encoded in the QR code
type QRCodeData struct {
	DocID     string `json:"doc_id"`
	Hash      string `json:"hash"`
	Signature string `json:"signature"`
	Timestamp int64  `json:"timestamp"`
}

// QRPosition defines where to place the QR code on the page
type QRPosition struct {
	X      float64 `json:"x"`       // X coordinate (points from left)
	Y      float64 `json:"y"`       // Y coordinate (points from bottom)
	Width  float64 `json:"width"`   // QR code width in points
	Height float64 `json:"height"`  // QR code height in points
}

// DefaultQRPosition returns the default position for QR code (bottom right of last page)
func DefaultQRPosition() QRPosition {
	return QRPosition{
		X:      450, // 450 points from left (about 6.25 inches on 8.5" page)
		Y:      50,  // 50 points from bottom
		Width:  100, // 100 points wide (about 1.4 inches)
		Height: 100, // 100 points tall (about 1.4 inches)
	}
}