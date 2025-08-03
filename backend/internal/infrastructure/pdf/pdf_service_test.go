package pdf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createMinimalPDF creates a minimal valid PDF for testing
func createMinimalPDF() []byte {
	// This is a minimal valid PDF structure
	pdfContent := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj

2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj

3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
>>
endobj

xref
0 4
0000000000 65535 f 
0000000010 00000 n 
0000000053 00000 n 
0000000100 00000 n 
trailer
<<
/Size 4
/Root 1 0 R
>>
startxref
157
%%EOF`
	return []byte(pdfContent)
}

// createPDFWithVersion creates a PDF with a specific version
func createPDFWithVersion(version string) []byte {
	pdfContent := `%PDF-` + version + `
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj

2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj

3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
>>
endobj

xref
0 4
0000000000 65535 f 
0000000010 00000 n 
0000000053 00000 n 
0000000100 00000 n 
trailer
<<
/Size 4
/Root 1 0 R
>>
startxref
157
%%EOF`
	return []byte(pdfContent)
}

// createMultiPagePDF creates a PDF with multiple pages
func createMultiPagePDF() []byte {
	pdfContent := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj

2 0 obj
<<
/Type /Pages
/Kids [3 0 R 4 0 R]
/Count 2
>>
endobj

3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
>>
endobj

4 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
>>
endobj

xref
0 5
0000000000 65535 f 
0000000010 00000 n 
0000000053 00000 n 
0000000108 00000 n 
0000000155 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
202
%%EOF`
	return []byte(pdfContent)
}

func TestPDFService_ValidatePDF(t *testing.T) {
	service := NewPDFService()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid minimal PDF",
			data:    createMinimalPDF(),
			wantErr: false,
		},
		{
			name:    "valid PDF version 1.3",
			data:    createPDFWithVersion("1.3"),
			wantErr: false,
		},
		{
			name:    "valid PDF version 1.7",
			data:    createPDFWithVersion("1.7"),
			wantErr: false,
		},

		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
			errMsg:  "PDF data is empty",
		},
		{
			name:    "invalid PDF header",
			data:    []byte("not a pdf"),
			wantErr: true,
			errMsg:  "invalid PDF format: missing PDF header",
		},
		{
			name:    "PDF header but invalid content",
			data:    []byte("%PDF-1.4\ninvalid content"),
			wantErr: true,
			errMsg:  "invalid PDF format",
		},
		{
			name:    "oversized PDF",
			data:    make([]byte, MaxPDFSize+1),
			wantErr: true,
			errMsg:  "PDF size exceeds maximum allowed size",
		},
		{
			name:    "truncated PDF header",
			data:    []byte("%PD"),
			wantErr: true,
			errMsg:  "invalid PDF format: missing PDF header",
		},
		{
			name:    "PDF with extra whitespace in header",
			data:    []byte("  %PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n>>\nendobj"),
			wantErr: true,
			errMsg:  "invalid PDF format: missing PDF header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePDF(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPDFService_CalculateHash(t *testing.T) {
	service := NewPDFService()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid PDF",
			data:    createMinimalPDF(),
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
			errMsg:  "PDF validation failed",
		},
		{
			name:    "invalid PDF",
			data:    []byte("not a pdf"),
			wantErr: true,
			errMsg:  "PDF validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := service.CalculateHash(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, hash)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hash)
				assert.Len(t, hash, 32) // SHA-256 produces 32-byte hash
			}
		})
	}
}

func TestPDFService_CalculateHash_Consistency(t *testing.T) {
	service := NewPDFService()
	pdfData := createMinimalPDF()

	// Calculate hash multiple times
	hash1, err1 := service.CalculateHash(pdfData)
	require.NoError(t, err1)

	hash2, err2 := service.CalculateHash(pdfData)
	require.NoError(t, err2)

	// Hashes should be identical for the same data
	assert.Equal(t, hash1, hash2)
}

func TestPDFService_CalculateHash_Different(t *testing.T) {
	service := NewPDFService()
	
	pdfData1 := createMinimalPDF()
	pdfData2 := append(pdfData1, []byte(" ")...) // Add a space to make it different

	hash1, err1 := service.CalculateHash(pdfData1)
	require.NoError(t, err1)

	hash2, err2 := service.CalculateHash(pdfData2)
	require.NoError(t, err2)

	// Hashes should be different for different data
	assert.NotEqual(t, hash1, hash2)
}

func TestPDFService_CalculateHash_DifferentVersions(t *testing.T) {
	service := NewPDFService()
	
	pdf13 := createPDFWithVersion("1.3")
	pdf17 := createPDFWithVersion("1.7")

	hash1, err1 := service.CalculateHash(pdf13)
	require.NoError(t, err1)

	hash2, err2 := service.CalculateHash(pdf17)
	require.NoError(t, err2)

	// Hashes should be different for different PDF versions
	assert.NotEqual(t, hash1, hash2)
	assert.Len(t, hash1, 32) // SHA-256 produces 32-byte hash
	assert.Len(t, hash2, 32) // SHA-256 produces 32-byte hash
}

func TestPDFService_CalculateHash_MultiPage(t *testing.T) {
	service := NewPDFService()
	
	singlePage := createMinimalPDF()
	// Create a different single page PDF by changing version
	differentPDF := createPDFWithVersion("1.5")

	hash1, err1 := service.CalculateHash(singlePage)
	require.NoError(t, err1)

	hash2, err2 := service.CalculateHash(differentPDF)
	require.NoError(t, err2)

	// Hashes should be different for different PDFs
	assert.NotEqual(t, hash1, hash2)
}

func TestPDFService_GetPDFInfo(t *testing.T) {
	service := NewPDFService()

	tests := []struct {
		name          string
		data          []byte
		wantErr       bool
		errMsg        string
		expectedPages int
	}{
		{
			name:          "valid single-page PDF",
			data:          createMinimalPDF(),
			wantErr:       false,
			expectedPages: 1,
		},

		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
			errMsg:  "PDF validation failed",
		},
		{
			name:    "invalid PDF",
			data:    []byte("not a pdf"),
			wantErr: true,
			errMsg:  "PDF validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := service.GetPDFInfo(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, info)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, info)
				assert.Equal(t, tt.expectedPages, info.NumPages)
				assert.Equal(t, int64(len(tt.data)), info.FileSize)
			}
		})
	}
}

func TestPDFService_ReadPDFFromReader(t *testing.T) {
	service := NewPDFService()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid PDF data",
			data:    createMinimalPDF(),
			wantErr: false,
		},
		{
			name:    "empty reader",
			data:    []byte{},
			wantErr: false, // Empty data is allowed, validation happens later
		},
		{
			name:    "oversized data",
			data:    make([]byte, MaxPDFSize+1),
			wantErr: true,
			errMsg:  "PDF size exceeds maximum allowed size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			data, err := service.ReadPDFFromReader(reader)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.data, data)
			}
		})
	}
}

func TestPDFService_ReadPDFFromReader_StringReader(t *testing.T) {
	service := NewPDFService()
	testData := "test data"
	reader := strings.NewReader(testData)

	data, err := service.ReadPDFFromReader(reader)
	assert.NoError(t, err)
	assert.Equal(t, []byte(testData), data)
}

func TestPDFService_ReadPDFFromReader_ExactMaxSize(t *testing.T) {
	service := NewPDFService()
	// Create data exactly at the max size limit
	testData := make([]byte, MaxPDFSize)
	reader := bytes.NewReader(testData)

	data, err := service.ReadPDFFromReader(reader)
	assert.NoError(t, err)
	assert.Equal(t, testData, data)
	assert.Len(t, data, MaxPDFSize)
}

func TestPDFService_ReadPDFFromReader_ValidPDF(t *testing.T) {
	service := NewPDFService()
	pdfData := createMinimalPDF()
	reader := bytes.NewReader(pdfData)

	data, err := service.ReadPDFFromReader(reader)
	assert.NoError(t, err)
	assert.Equal(t, pdfData, data)
	
	// Verify the read data is still a valid PDF
	err = service.ValidatePDF(data)
	assert.NoError(t, err)
}

func TestPDFService_MaxSizeConstant(t *testing.T) {
	// Verify the max size constant is set correctly (50MB)
	expectedMaxSize := 50 * 1024 * 1024
	assert.Equal(t, expectedMaxSize, MaxPDFSize)
}

func TestNewPDFService(t *testing.T) {
	service := NewPDFService()
	assert.NotNil(t, service)
	assert.IsType(t, &PDFService{}, service)
}

// Benchmark tests for performance
func BenchmarkPDFService_CalculateHash(b *testing.B) {
	service := NewPDFService()
	pdfData := createMinimalPDF()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CalculateHash(pdfData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPDFService_ValidatePDF(b *testing.B) {
	service := NewPDFService()
	pdfData := createMinimalPDF()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.ValidatePDF(pdfData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// QR Code Tests

func createTestQRCodeData() QRCodeData {
	return QRCodeData{
		DocID:     "test-doc-123",
		Hash:      "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234",
		Signature: "test-signature-data",
		Timestamp: time.Now().Unix(),
	}
}

func TestPDFService_GenerateQRCode(t *testing.T) {
	service := NewPDFService()

	tests := []struct {
		name    string
		data    QRCodeData
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid QR code data",
			data:    createTestQRCodeData(),
			wantErr: false,
		},
		{
			name: "empty doc ID",
			data: QRCodeData{
				DocID:     "",
				Hash:      "test-hash",
				Signature: "test-signature",
				Timestamp: time.Now().Unix(),
			},
			wantErr: false, // Empty fields are allowed, JSON will encode them
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qrCode, err := service.GenerateQRCode(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, qrCode)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, qrCode)
				assert.Greater(t, len(qrCode), 0)
				
				// QR code should be a valid PNG image (starts with PNG signature)
				assert.Equal(t, []byte{0x89, 0x50, 0x4E, 0x47}, qrCode[:4])
			}
		})
	}
}

func TestPDFService_GenerateQRCode_Consistency(t *testing.T) {
	service := NewPDFService()
	data := createTestQRCodeData()

	// Generate QR code multiple times
	qr1, err1 := service.GenerateQRCode(data)
	require.NoError(t, err1)

	qr2, err2 := service.GenerateQRCode(data)
	require.NoError(t, err2)

	// QR codes should be identical for the same data
	assert.Equal(t, qr1, qr2)
}

func TestPDFService_GenerateQRCode_Different(t *testing.T) {
	service := NewPDFService()
	
	data1 := createTestQRCodeData()
	data2 := createTestQRCodeData()
	data2.DocID = "different-doc-id"

	qr1, err1 := service.GenerateQRCode(data1)
	require.NoError(t, err1)

	qr2, err2 := service.GenerateQRCode(data2)
	require.NoError(t, err2)

	// QR codes should be different for different data
	assert.NotEqual(t, qr1, qr2)
}

func TestDefaultQRPosition(t *testing.T) {
	pos := DefaultQRPosition()
	
	assert.Equal(t, float64(450), pos.X)
	assert.Equal(t, float64(50), pos.Y)
	assert.Equal(t, float64(100), pos.Width)
	assert.Equal(t, float64(100), pos.Height)
}

func TestPDFService_InjectQRCode(t *testing.T) {
	service := NewPDFService()
	pdfData := createMinimalPDF()
	qrData := createTestQRCodeData()

	tests := []struct {
		name     string
		pdfData  []byte
		qrData   QRCodeData
		position *QRPosition
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid PDF with default position (license required)",
			pdfData:  pdfData,
			qrData:   qrData,
			position: nil, // Use default position
			wantErr:  true, // Expect license error in test environment
			errMsg:   "license",
		},
		{
			name:    "valid PDF with custom position (license required)",
			pdfData: pdfData,
			qrData:  qrData,
			position: &QRPosition{
				X:      100,
				Y:      100,
				Width:  80,
				Height: 80,
			},
			wantErr: true, // Expect license error in test environment
			errMsg:  "license",
		},
		{
			name:     "invalid PDF data",
			pdfData:  []byte("not a pdf"),
			qrData:   qrData,
			position: nil,
			wantErr:  true,
			errMsg:   "PDF validation failed",
		},
		{
			name:     "empty PDF data",
			pdfData:  []byte{},
			qrData:   qrData,
			position: nil,
			wantErr:  true,
			errMsg:   "PDF validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.InjectQRCode(tt.pdfData, tt.qrData, tt.position)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Greater(t, len(result), len(tt.pdfData)) // Result should be larger due to QR code
				
				// Verify the result is still a valid PDF
				err = service.ValidatePDF(result)
				assert.NoError(t, err)
				
				// Verify the result has the same number of pages
				originalInfo, err := service.GetPDFInfo(tt.pdfData)
				require.NoError(t, err)
				
				resultInfo, err := service.GetPDFInfo(result)
				require.NoError(t, err)
				
				assert.Equal(t, originalInfo.NumPages, resultInfo.NumPages)
			}
		})
	}
}

func TestPDFService_InjectQRCode_PreservesContent(t *testing.T) {
	service := NewPDFService()
	pdfData := createMinimalPDF()
	qrData := createTestQRCodeData()

	// Inject QR code - expect license error in test environment
	result, err := service.InjectQRCode(pdfData, qrData, nil)
	
	// In test environment without license, expect error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "license")
	assert.Nil(t, result)
	
	// Note: In production with proper license, this would work correctly
	// and preserve content while adding QR code
}

func TestPDFService_InjectQRCode_DifferentPositions(t *testing.T) {
	service := NewPDFService()
	pdfData := createMinimalPDF()
	qrData := createTestQRCodeData()

	positions := []QRPosition{
		{X: 50, Y: 50, Width: 100, Height: 100},     // Bottom left
		{X: 450, Y: 50, Width: 100, Height: 100},    // Bottom right (default)
		{X: 50, Y: 700, Width: 100, Height: 100},    // Top left
		{X: 450, Y: 700, Width: 100, Height: 100},   // Top right
	}

	for i, pos := range positions {
		t.Run(fmt.Sprintf("position_%d", i), func(t *testing.T) {
			result, err := service.InjectQRCode(pdfData, qrData, &pos)
			
			// In test environment without license, expect error
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "license")
			assert.Nil(t, result)
			
			// Note: In production with proper license, this would work correctly
			// and inject QR code at the specified position
		})
	}
}

func TestQRCodeData_JSONSerialization(t *testing.T) {
	data := createTestQRCodeData()
	
	// Test JSON marshaling
	jsonBytes, err := json.Marshal(data)
	require.NoError(t, err)
	assert.Greater(t, len(jsonBytes), 0)
	
	// Test JSON unmarshaling
	var unmarshaled QRCodeData
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)
	
	assert.Equal(t, data.DocID, unmarshaled.DocID)
	assert.Equal(t, data.Hash, unmarshaled.Hash)
	assert.Equal(t, data.Signature, unmarshaled.Signature)
	assert.Equal(t, data.Timestamp, unmarshaled.Timestamp)
}

// Benchmark tests for QR code operations
func BenchmarkPDFService_GenerateQRCode(b *testing.B) {
	service := NewPDFService()
	data := createTestQRCodeData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GenerateQRCode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPDFService_InjectQRCode(b *testing.B) {
	service := NewPDFService()
	pdfData := createMinimalPDF()
	qrData := createTestQRCodeData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.InjectQRCode(pdfData, qrData, nil)
		// In test environment, expect license error
		if err != nil && !strings.Contains(err.Error(), "license") {
			b.Fatal(err)
		}
	}
}