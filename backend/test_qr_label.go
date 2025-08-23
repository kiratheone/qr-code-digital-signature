package main

import (
	"fmt"
	"os"

	"digital-signature-system/internal/infrastructure/pdf"
)

func main() {
	// Create PDF service
	pdfService := pdf.NewPDFService()

	// Test QR code generation with center label
	url := "https://yourdomain.com/verify/document123"
	label := "John Doe"
	size := 256

	fmt.Printf("Generating QR code with center label...\n")
	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Label: %s\n", label)
	fmt.Printf("Size: %d\n", size)

	qrData, err := pdfService.GenerateQRCodeWithCenterLabel(url, label, size)
	if err != nil {
		fmt.Printf("Error generating QR code: %v\n", err)
		os.Exit(1)
	}

	// Save to file for inspection
	filename := "qr_with_label_test.png"
	err = os.WriteFile(filename, qrData, 0644)
	if err != nil {
		fmt.Printf("Error saving QR code to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Success! QR code with center label saved to %s\n", filename)
	fmt.Printf("QR code size: %d bytes\n", len(qrData))

	// Test without label
	fmt.Printf("\nGenerating QR code without label...\n")
	qrDataNoLabel, err := pdfService.GenerateQRCodeWithCenterLabel(url, "", size)
	if err != nil {
		fmt.Printf("Error generating QR code without label: %v\n", err)
		os.Exit(1)
	}

	filenameNoLabel := "qr_no_label_test.png"
	err = os.WriteFile(filenameNoLabel, qrDataNoLabel, 0644)
	if err != nil {
		fmt.Printf("Error saving QR code to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Success! QR code without label saved to %s\n", filenameNoLabel)
	fmt.Printf("QR code size: %d bytes\n", len(qrDataNoLabel))

	// Test with different label lengths
	labels := []string{
		"A",
		"AB",
		"ABC",
		"ABCD",
		"ABCDEFGH",
		"VeryLongLabelName",
		"This is a very long label that should be truncated",
	}

	for i, testLabel := range labels {
		fmt.Printf("\nTesting label %d: '%s'\n", i+1, testLabel)
		qrData, err := pdfService.GenerateQRCodeWithCenterLabel(url, testLabel, size)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		filename := fmt.Sprintf("qr_label_test_%d.png", i+1)
		err = os.WriteFile(filename, qrData, 0644)
		if err != nil {
			fmt.Printf("Error saving: %v\n", err)
			continue
		}

		fmt.Printf("Saved to %s (%d bytes)\n", filename, len(qrData))
	}

	fmt.Println("\nAll tests completed successfully!")
}
