package services

import "context"

// QRCodePosition defines the position for QR code injection in a PDF
type QRCodePosition struct {
	// Page number (0-based index, -1 means last page)
	Page int `json:"page"`
	// X coordinate (from left)
	X float64 `json:"x"`
	// Y coordinate (from bottom)
	Y float64 `json:"y"`
	// Width of the QR code
	Width float64 `json:"width"`
	// Height of the QR code
	Height float64 `json:"height"`
}

// QRCodeData contains the data to be encoded in a QR code
type QRCodeData struct {
	// DocumentID is the unique identifier for the document
	DocumentID string `json:"document_id"`
	// Hash is the document hash
	Hash string `json:"hash"`
	// Signature is the digital signature
	Signature string `json:"signature"`
	// Timestamp is the time when the QR code was generated
	Timestamp int64 `json:"timestamp"`
	// Issuer is the document issuer
	Issuer string `json:"issuer,omitempty"`
	// VerificationURL is the URL for verification
	VerificationURL string `json:"verification_url,omitempty"`
}

// QRService defines the interface for QR code operations
type QRService interface {
	// GenerateQRCode generates a QR code image from the provided data
	// Returns the QR code as a PNG image byte slice or error if generation fails
	GenerateQRCode(ctx context.Context, data *QRCodeData, size int) ([]byte, error)

	// InjectQRCode injects a QR code into a PDF document at the specified position
	// Returns the modified PDF document or error if injection fails
	InjectQRCode(ctx context.Context, pdfData []byte, qrCodeData []byte, position *QRCodePosition) ([]byte, error)

	// ParseQRCode parses a QR code image and extracts the encoded data
	// Returns the parsed data or error if parsing fails
	ParseQRCode(ctx context.Context, qrCodeImage []byte) (*QRCodeData, error)

	// GetDefaultPosition returns the default position for QR code injection (last page, bottom right)
	GetDefaultPosition(ctx context.Context, pdfData []byte) (*QRCodePosition, error)
}