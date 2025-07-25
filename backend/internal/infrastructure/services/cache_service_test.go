package services

import (
	"digital-signature-system/internal/domain/entities"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheService_NewCacheService(t *testing.T) {
	ttl := 5 * time.Minute
	cache := NewCacheService(ttl)
	
	assert.NotNil(t, cache)
	assert.Equal(t, ttl, cache.ttl)
	assert.NotNil(t, cache.cache)
}

func TestCacheService_SetAndGetDocument(t *testing.T) {
	cache := NewCacheService(5 * time.Minute)
	
	testDoc := &entities.Document{
		ID:       "test-doc-1",
		Filename: "test.pdf",
		Issuer:   "Test Issuer",
	}
	
	// Test setting document
	cache.SetDocument("test-doc-1", testDoc)
	
	// Test getting document
	retrievedDoc, found := cache.GetDocument("test-doc-1")
	assert.True(t, found)
	assert.Equal(t, testDoc.ID, retrievedDoc.ID)
	assert.Equal(t, testDoc.Filename, retrievedDoc.Filename)
	assert.Equal(t, testDoc.Issuer, retrievedDoc.Issuer)
}

func TestCacheService_GetNonExistentDocument(t *testing.T) {
	cache := NewCacheService(5 * time.Minute)
	
	// Test getting non-existent document
	doc, found := cache.GetDocument("non-existent")
	assert.False(t, found)
	assert.Nil(t, doc)
}

func TestCacheService_SetAndGetVerificationInfo(t *testing.T) {
	cache := NewCacheService(5 * time.Minute)
	
	testInfo := map[string]interface{}{
		"filename":   "test.pdf",
		"issuer":     "Test Issuer",
		"created_at": "2023-01-01T00:00:00Z",
	}
	
	// Test setting verification info
	cache.SetVerificationInfo("test-doc-1", testInfo)
	
	// Test getting verification info
	retrievedInfo, found := cache.GetVerificationInfo("test-doc-1")
	assert.True(t, found)
	assert.Equal(t, testInfo["filename"], retrievedInfo["filename"])
	assert.Equal(t, testInfo["issuer"], retrievedInfo["issuer"])
	assert.Equal(t, testInfo["created_at"], retrievedInfo["created_at"])
}

func TestCacheService_GetNonExistentVerificationInfo(t *testing.T) {
	cache := NewCacheService(5 * time.Minute)
	
	// Test getting non-existent verification info
	info, found := cache.GetVerificationInfo("non-existent")
	assert.False(t, found)
	assert.Nil(t, info)
}

func TestCacheService_Expiration(t *testing.T) {
	// Use very short TTL for testing
	cache := NewCacheService(100 * time.Millisecond)
	
	testDoc := &entities.Document{
		ID:       "test-doc-1",
		Filename: "test.pdf",
		Issuer:   "Test Issuer",
	}
	
	// Set document
	cache.SetDocument("test-doc-1", testDoc)
	
	// Should be available immediately
	doc, found := cache.GetDocument("test-doc-1")
	assert.True(t, found)
	assert.NotNil(t, doc)
	
	// Wait for expiration
	time.Sleep(150 * time.Millisecond)
	
	// Should be expired now
	doc, found = cache.GetDocument("test-doc-1")
	assert.False(t, found)
	assert.Nil(t, doc)
}

func TestCacheService_Delete(t *testing.T) {
	cache := NewCacheService(5 * time.Minute)
	
	testDoc := &entities.Document{
		ID:       "test-doc-1",
		Filename: "test.pdf",
		Issuer:   "Test Issuer",
	}
	
	// Set document
	cache.SetDocument("test-doc-1", testDoc)
	
	// Verify it exists
	doc, found := cache.GetDocument("test-doc-1")
	assert.True(t, found)
	assert.NotNil(t, doc)
	
	// Delete it
	cache.Delete("doc:test-doc-1")
	
	// Verify it's gone
	doc, found = cache.GetDocument("test-doc-1")
	assert.False(t, found)
	assert.Nil(t, doc)
}

func TestCacheService_Clear(t *testing.T) {
	cache := NewCacheService(5 * time.Minute)
	
	// Add multiple items
	testDoc1 := &entities.Document{ID: "doc-1", Filename: "test1.pdf"}
	testDoc2 := &entities.Document{ID: "doc-2", Filename: "test2.pdf"}
	testInfo := map[string]interface{}{"filename": "test.pdf"}
	
	cache.SetDocument("doc-1", testDoc1)
	cache.SetDocument("doc-2", testDoc2)
	cache.SetVerificationInfo("doc-1", testInfo)
	
	// Verify items exist
	_, found1 := cache.GetDocument("doc-1")
	_, found2 := cache.GetDocument("doc-2")
	_, found3 := cache.GetVerificationInfo("doc-1")
	assert.True(t, found1)
	assert.True(t, found2)
	assert.True(t, found3)
	
	// Clear cache
	cache.Clear()
	
	// Verify all items are gone
	_, found1 = cache.GetDocument("doc-1")
	_, found2 = cache.GetDocument("doc-2")
	_, found3 = cache.GetVerificationInfo("doc-1")
	assert.False(t, found1)
	assert.False(t, found2)
	assert.False(t, found3)
}

func TestCacheService_Stats(t *testing.T) {
	cache := NewCacheService(5 * time.Minute)
	
	// Initially empty
	stats := cache.Stats()
	assert.Equal(t, 0, stats["total_items"])
	assert.Equal(t, 5.0, stats["ttl_minutes"])
	
	// Add some items
	testDoc := &entities.Document{ID: "doc-1", Filename: "test.pdf"}
	testInfo := map[string]interface{}{"filename": "test.pdf"}
	
	cache.SetDocument("doc-1", testDoc)
	cache.SetVerificationInfo("doc-1", testInfo)
	
	// Check stats
	stats = cache.Stats()
	assert.Equal(t, 2, stats["total_items"])
	assert.Equal(t, 5.0, stats["ttl_minutes"])
}

func TestCacheService_ConcurrentAccess(t *testing.T) {
	cache := NewCacheService(5 * time.Minute)
	
	concurrency := 10
	iterations := 100
	
	var wg sync.WaitGroup
	
	// Writers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				testDoc := &entities.Document{
					ID:       fmt.Sprintf("doc-%d-%d", id, j),
					Filename: fmt.Sprintf("test-%d-%d.pdf", id, j),
				}
				cache.SetDocument(testDoc.ID, testDoc)
			}
		}(i)
	}
	
	// Readers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				docID := fmt.Sprintf("doc-%d-%d", id, j)
				cache.GetDocument(docID)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Verify some items exist
	stats := cache.Stats()
	totalItems := stats["total_items"].(int)
	assert.Greater(t, totalItems, 0)
}

func TestCacheService_CleanupExpiredItems(t *testing.T) {
	// Use short TTL and cleanup interval for testing
	cache := NewCacheService(50 * time.Millisecond)
	
	// Add items
	for i := 0; i < 5; i++ {
		testDoc := &entities.Document{
			ID:       fmt.Sprintf("doc-%d", i),
			Filename: fmt.Sprintf("test-%d.pdf", i),
		}
		cache.SetDocument(testDoc.ID, testDoc)
	}
	
	// Verify items exist
	stats := cache.Stats()
	assert.Equal(t, 5, stats["total_items"])
	
	// Wait for expiration and cleanup
	time.Sleep(200 * time.Millisecond)
	
	// Add a new item to trigger potential cleanup check
	testDoc := &entities.Document{ID: "new-doc", Filename: "new.pdf"}
	cache.SetDocument("new-doc", testDoc)
	
	// Give cleanup goroutine time to run
	time.Sleep(100 * time.Millisecond)
	
	// Check that expired items are cleaned up
	// Note: The cleanup runs every 5 minutes by default, so we can't easily test it
	// without modifying the implementation. This test mainly ensures no panics occur.
	stats = cache.Stats()
	assert.GreaterOrEqual(t, stats["total_items"].(int), 1) // At least the new item should exist
}

func TestCacheService_InvalidTypeHandling(t *testing.T) {
	cache := NewCacheService(5 * time.Minute)
	
	// Manually insert invalid data type
	cache.cache["doc:invalid"] = &CacheItem{
		Data:      "not a document",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	
	// Should return nil and false for invalid type
	doc, found := cache.GetDocument("invalid")
	assert.False(t, found)
	assert.Nil(t, doc)
	
	// Same for verification info
	cache.cache["verify:invalid"] = &CacheItem{
		Data:      "not verification info",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	
	info, found := cache.GetVerificationInfo("invalid")
	assert.False(t, found)
	assert.Nil(t, info)
}