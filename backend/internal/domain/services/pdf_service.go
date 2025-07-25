package services

import "context"

// PDFService defines the interface for PDF processing operations
type PDFService interface {
	// ValidatePDF validates if the provided data is a valid PDF file
	// Returns nil if valid, error otherwise
	ValidatePDF(ctx context.Context, pdfData []byte) error

	// CalculateHash calculates SHA-256 hash for the provided PDF document
	// Returns the hash as a byte slice or error if processing fails
	CalculateHash(ctx context.Context, pdfData []byte) ([]byte, error)

	// GetPDFInfo extracts basic information from the PDF document
	// Returns information about the PDF or error if processing fails
	GetPDFInfo(ctx context.Context, pdfData []byte) (*PDFInfo, error)
}

// PDFInfo contains basic information about a PDF document
type PDFInfo struct {
	PageCount  int    `json:"page_count"`
	Title      string `json:"title,omitempty"`
	Author     string `json:"author,omitempty"`
	Creator    string `json:"creator,omitempty"`
	Producer   string `json:"producer,omitempty"`
	FileSize   int64  `json:"file_size"`
	IsEncrypted bool  `json:"is_encrypted"`
}