package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"digital-signature-system/internal/config"
	"digital-signature-system/internal/domain/entities"
	"digital-signature-system/internal/infrastructure/database"
	"fmt"
	"testing"
	"time"
)

// BenchmarkDatabaseConnectionPool tests database connection pool performance
func BenchmarkDatabaseConnectionPool(b *testing.B) {
	cfg := config.DatabaseConfig{
		Host:            "localhost",
		Port:            "5432",
		User:            "postgres",
		Password:        "password",
		DBName:          "digital_signature_test",
		SSLMode:         "disable",
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: 60,
		ConnMaxIdleTime: 30,
	}

	db, err := database.NewConnection(cfg)
	if err != nil {
		b.Skipf("Database connection failed: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var result int
			err := db.Raw("SELECT 1").Scan(&result).Error
			if err != nil {
				b.Errorf("Query failed: %v", err)
			}
		}
	})
}

// BenchmarkPDFHashCalculation tests PDF hash calculation performance
func BenchmarkPDFHashCalculation(b *testing.B) {
	pdfService := NewPDFService(50 * 1024 * 1024) // 50MB limit
	ctx := context.Background()

	// Test with different file sizes
	testSizes := []int{
		1024,        // 1KB
		10 * 1024,   // 10KB
		100 * 1024,  // 100KB
		1024 * 1024, // 1MB
	}

	for _, size := range testSizes {
		b.Run(fmt.Sprintf("Size_%dKB", size/1024), func(b *testing.B) {
			// Generate test PDF data
			testData := make([]byte, size)
			rand.Read(testData)
			
			// Add PDF header to make it look like a PDF
			pdfHeader := []byte("%PDF-1.4\n")
			testData = append(pdfHeader, testData...)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := pdfService.(*PDFServiceImpl).calculateHashStreaming(ctx, bytes.NewReader(testData))
				if err != nil {
					b.Errorf("Hash calculation failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkCacheService tests cache service performance
func BenchmarkCacheService(b *testing.B) {
	cache := NewCacheService(5 * time.Minute)
	
	// Test data
	testDoc := &entities.Document{
		ID:       "test-doc-id",
		Filename: "test.pdf",
		Issuer:   "Test Issuer",
	}

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cache.SetDocument(fmt.Sprintf("doc-%d", i), testDoc)
		}
	})

	// Pre-populate cache for get tests
	for i := 0; i < 1000; i++ {
		cache.SetDocument(fmt.Sprintf("doc-%d", i), testDoc)
	}

	b.Run("Get", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				_, _ = cache.GetDocument(fmt.Sprintf("doc-%d", i%1000))
				i++
			}
		})
	})
}

// BenchmarkSignatureService tests signature service performance
func BenchmarkSignatureService(b *testing.B) {
	// Skip if no keys available
	privateKey := "test-private-key"
	publicKey := "test-public-key"
	
	if privateKey == "" || publicKey == "" {
		b.Skip("No test keys available")
	}

	sigService, err := NewSignatureService(privateKey, publicKey)
	if err != nil {
		b.Skipf("Failed to create signature service: %v", err)
	}

	// Test data
	testHash := make([]byte, 32) // SHA-256 hash size
	rand.Read(testHash)

	b.Run("Sign", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := sigService.SignDocument(testHash)
			if err != nil {
				b.Errorf("Signing failed: %v", err)
			}
		}
	})

	// Create a signature for verification tests
	signature, err := sigService.SignDocument(testHash)
	if err != nil {
		b.Fatalf("Failed to create test signature: %v", err)
	}

	b.Run("Verify", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := sigService.VerifySignature(testHash, signature)
			if err != nil {
				b.Errorf("Verification failed: %v", err)
			}
		}
	})
}

// TestDatabaseConnectionPoolStress tests database connection pool under stress
func TestDatabaseConnectionPoolStress(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:            "localhost",
		Port:            "5432",
		User:            "postgres",
		Password:        "password",
		DBName:          "digital_signature_test",
		SSLMode:         "disable",
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: 60,
		ConnMaxIdleTime: 30,
	}

	db, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("Database connection failed: %v", err)
	}

	// Test concurrent connections
	concurrency := 50
	iterations := 100
	
	errChan := make(chan error, concurrency)
	
	for i := 0; i < concurrency; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				var result int
				err := db.Raw("SELECT 1").Scan(&result).Error
				if err != nil {
					errChan <- err
					return
				}
			}
			errChan <- nil
		}()
	}

	// Collect results
	for i := 0; i < concurrency; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent query failed: %v", err)
		}
	}
}

// TestCacheServiceConcurrency tests cache service under concurrent access
func TestCacheServiceConcurrency(t *testing.T) {
	cache := NewCacheService(5 * time.Minute)
	
	testDoc := &entities.Document{
		ID:       "test-doc-id",
		Filename: "test.pdf",
		Issuer:   "Test Issuer",
	}

	concurrency := 50
	iterations := 100
	
	errChan := make(chan error, concurrency*2)
	
	// Writers
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				cache.SetDocument(fmt.Sprintf("doc-%d-%d", id, j), testDoc)
			}
			errChan <- nil
		}(i)
	}

	// Readers
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				_, _ = cache.GetDocument(fmt.Sprintf("doc-%d-%d", id, j))
			}
			errChan <- nil
		}(i)
	}

	// Collect results
	for i := 0; i < concurrency*2; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent cache operation failed: %v", err)
		}
	}
}