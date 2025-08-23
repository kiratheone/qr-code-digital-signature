package pdf

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"strings"

	"github.com/skip2/go-qrcode"
	"github.com/unidoc/unipdf/v3/creator"
	"github.com/unidoc/unipdf/v3/model"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
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

// GenerateQRCodeWithCenterLabel generates a QR code with a text label in the center
func (s *PDFService) GenerateQRCodeWithCenterLabel(url string, label string, size int) ([]byte, error) {
	if url == "" {
		return nil, fmt.Errorf("URL is required")
	}

	if size <= 0 {
		size = 256 // Default size
	}

	// Generate QR code with highest error correction to allow center modifications
	qrCode, err := qrcode.New(url, qrcode.Highest) // Highest allows ~30% damage
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	qrCode.DisableBorder = false

	// Get QR code as image
	qrImage := qrCode.Image(size)

	// If no label provided, return simple QR code
	if label == "" {
		var buf bytes.Buffer
		if err := png.Encode(&buf, qrImage); err != nil {
			return nil, fmt.Errorf("failed to encode QR code: %w", err)
		}
		return buf.Bytes(), nil
	}

	// Convert to RGBA for modifications
	bounds := qrImage.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, qrImage, bounds.Min, draw.Src)

	// Add center label
	if err := s.addCenterLabelToQR(rgba, label, size); err != nil {
		return nil, fmt.Errorf("failed to add center label: %w", err)
	}

	// Encode final image
	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		return nil, fmt.Errorf("failed to encode QR code with label: %w", err)
	}

	return buf.Bytes(), nil
}

// addCenterLabelToQR adds a text label in the center of the QR code
func (s *PDFService) addCenterLabelToQR(img *image.RGBA, text string, qrSize int) error {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate center area size (about 20% of QR code)
	centerSize := qrSize / 5
	centerX := width / 2
	centerY := height / 2

	// Create white background circle/rectangle for the label
	labelBounds := image.Rect(
		centerX-centerSize/2,
		centerY-centerSize/2,
		centerX+centerSize/2,
		centerY+centerSize/2,
	)

	// Draw white background with border
	s.drawRoundedRect(img, labelBounds, color.RGBA{255, 255, 255, 255}, color.RGBA{0, 0, 0, 255})

	// Add text to center
	if err := s.drawCenterText(img, text, labelBounds); err != nil {
		return fmt.Errorf("failed to draw center text: %w", err)
	}

	return nil
}

// drawRoundedRect draws a rounded rectangle with background and border
func (s *PDFService) drawRoundedRect(img *image.RGBA, bounds image.Rectangle, bgColor, borderColor color.RGBA) {
	// For simplicity, draw a regular rectangle with rounded corners effect
	// Draw background
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Simple rounded corners by skipping corner pixels
			if (x == bounds.Min.X || x == bounds.Max.X-1) && (y == bounds.Min.Y || y == bounds.Max.Y-1) {
				continue // Skip corner pixels for rounded effect
			}
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, bgColor)
			}
		}
	}

	// Draw border
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		if x >= 0 && x < img.Bounds().Dx() {
			if bounds.Min.Y >= 0 && bounds.Min.Y < img.Bounds().Dy() {
				img.Set(x, bounds.Min.Y, borderColor) // Top border
			}
			if bounds.Max.Y-1 >= 0 && bounds.Max.Y-1 < img.Bounds().Dy() {
				img.Set(x, bounds.Max.Y-1, borderColor) // Bottom border
			}
		}
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		if y >= 0 && y < img.Bounds().Dy() {
			if bounds.Min.X >= 0 && bounds.Min.X < img.Bounds().Dx() {
				img.Set(bounds.Min.X, y, borderColor) // Left border
			}
			if bounds.Max.X-1 >= 0 && bounds.Max.X-1 < img.Bounds().Dx() {
				img.Set(bounds.Max.X-1, y, borderColor) // Right border
			}
		}
	}
}

// drawCenterText draws text in the center of the given bounds
func (s *PDFService) drawCenterText(img *image.RGBA, text string, bounds image.Rectangle) error {
	// Use basic font for simplicity
	face := basicfont.Face7x13

	// Calculate text dimensions
	textWidth := font.MeasureString(face, text)
	textHeight := face.Metrics().Height

	// Calculate center position
	centerX := bounds.Min.X + bounds.Dx()/2
	centerY := bounds.Min.Y + bounds.Dy()/2

	// Starting position for text (center it)
	x := centerX - textWidth.Ceil()/2
	y := centerY + textHeight.Ceil()/4 // Adjust for baseline

	// Draw text
	point := fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	// Create a font drawer
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{0, 0, 0, 255}), // Black text
		Face: face,
		Dot:  point,
	}

	// Truncate text if too long
	maxChars := bounds.Dx() / 8 // Rough estimation
	if len(text) > maxChars {
		text = text[:maxChars-3] + "..."
	}

	d.DrawString(text)
	return nil
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
	X      float64 `json:"x"`      // X coordinate (points from left)
	Y      float64 `json:"y"`      // Y coordinate (points from bottom)
	Width  float64 `json:"width"`  // QR code width in points
	Height float64 `json:"height"` // QR code height in points
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
