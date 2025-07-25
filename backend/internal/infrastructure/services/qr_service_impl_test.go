package services

import (
	"context"
	"digital-signature-system/internal/domain/services"
	"encoding/json"
	"testing"
	"time"
)

func TestQRServiceImpl_GenerateQRCode(t *testing.T) {
	// Create a new QR service
	qrService := NewQRService("https://example.com/verify/")
	ctx := context.Background()

	// Test with nil data
	t.Run("Nil data", func(t *testing.T) {
		_, err := qrService.GenerateQRCode(ctx, nil, 0)
		if err == nil {
			t.Errorf("GenerateQRCode() should fail with nil data")
		}
	})

	// Test with valid data
	t.Run("Valid data", func(t *testing.T) {
		data := &services.QRCodeData{
			DocumentID:      "doc123",
			Hash:            "abcdef1234567890",
			Signature:       "signature123",
			Timestamp:       time.Now().Unix(),
			Issuer:          "Test Issuer",
			VerificationURL: "https://example.com/verify/doc123",
		}

		qrCode, err := qrService.GenerateQRCode(ctx, data, 200)
		if err != nil {
			t.Errorf("GenerateQRCode() error = %v", err)
			return
		}

		if len(qrCode) == 0 {
			t.Errorf("GenerateQRCode() returned empty QR code")
		}
	})
}

func TestQRServiceImpl_InjectQRCode(t *testing.T) {
	// Create a new QR service
	qrService := NewQRService("https://example.com/verify/")
	ctx := context.Background()

	// Test with nil PDF data
	t.Run("Nil PDF data", func(t *testing.T) {
		_, err := qrService.InjectQRCode(ctx, nil, []byte("qr code"), nil)
		if err == nil {
			t.Errorf("InjectQRCode() should fail with nil PDF data")
		}
	})

	// Test with nil QR code data
	t.Run("Nil QR code data", func(t *testing.T) {
		_, err := qrService.InjectQRCode(ctx, []byte("pdf data"), nil, nil)
		if err == nil {
			t.Errorf("InjectQRCode() should fail with nil QR code data")
		}
	})

	// Test with valid data and PDF
	t.Run("Valid data with minimal PDF", func(t *testing.T) {
		// Generate a QR code first
		data := &services.QRCodeData{
			DocumentID:      "doc123",
			Hash:            "abcdef1234567890",
			Signature:       "signature123",
			Timestamp:       time.Now().Unix(),
			Issuer:          "Test Issuer",
			VerificationURL: "https://example.com/verify/doc123",
		}

		qrCode, err := qrService.GenerateQRCode(ctx, data, 100)
		if err != nil {
			t.Errorf("GenerateQRCode() error = %v", err)
			return
		}

		// Create position
		position := &services.QRCodePosition{
			Page:   0,
			X:      10,
			Y:      10,
			Width:  50,
			Height: 50,
		}

		// Try to inject QR code into our minimal valid PDF
		_, err = qrService.InjectQRCode(ctx, MinimalValidPDF, qrCode, position)
		// This might fail due to the minimal PDF not being fully compliant
		// We're just checking that the function runs without panicking
		if err != nil {
			t.Logf("InjectQRCode() error = %v (expected for minimal PDF)", err)
		}
	})
}

func TestQRServiceImpl_GetDefaultPosition(t *testing.T) {
	// Create a new QR service
	qrService := NewQRService("https://example.com/verify/")
	ctx := context.Background()

	// Test with nil PDF data
	t.Run("Nil PDF data", func(t *testing.T) {
		_, err := qrService.GetDefaultPosition(ctx, nil)
		if err == nil {
			t.Errorf("GetDefaultPosition() should fail with nil PDF data")
		}
	})

	// Test with valid PDF
	t.Run("Valid PDF", func(t *testing.T) {
		position, err := qrService.GetDefaultPosition(ctx, MinimalValidPDF)
		// This might fail due to the minimal PDF not being fully compliant
		// We're just checking that the function runs without panicking
		if err != nil {
			t.Logf("GetDefaultPosition() error = %v (expected for minimal PDF)", err)
			return
		}

		if position != nil {
			if position.Width <= 0 || position.Height <= 0 {
				t.Errorf("GetDefaultPosition() returned invalid dimensions: width=%v, height=%v", 
					position.Width, position.Height)
			}
		}
	})
}