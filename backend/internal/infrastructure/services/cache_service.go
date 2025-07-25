package services

import (
	"sync"
	"time"

	"digital-signature-system/internal/domain/entities"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

// CacheService provides in-memory caching functionality
type CacheService struct {
	cache map[string]*CacheItem
	mutex sync.RWMutex
	ttl   time.Duration
}

// NewCacheService creates a new cache service
func NewCacheService(ttl time.Duration) *CacheService {
	cs := &CacheService{
		cache: make(map[string]*CacheItem),
		ttl:   ttl,
	}
	
	// Start cleanup goroutine
	go cs.cleanup()
	
	return cs
}

// GetDocument retrieves a document from cache
func (cs *CacheService) GetDocument(docID string) (*entities.Document, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	
	item, exists := cs.cache["doc:"+docID]
	if !exists || time.Now().After(item.ExpiresAt) {
		return nil, false
	}
	
	if doc, ok := item.Data.(*entities.Document); ok {
		return doc, true
	}
	
	return nil, false
}

// SetDocument stores a document in cache
func (cs *CacheService) SetDocument(docID string, doc *entities.Document) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	cs.cache["doc:"+docID] = &CacheItem{
		Data:      doc,
		ExpiresAt: time.Now().Add(cs.ttl),
	}
}

// GetVerificationInfo retrieves verification info from cache
func (cs *CacheService) GetVerificationInfo(docID string) (map[string]interface{}, bool) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	
	item, exists := cs.cache["verify:"+docID]
	if !exists || time.Now().After(item.ExpiresAt) {
		return nil, false
	}
	
	if info, ok := item.Data.(map[string]interface{}); ok {
		return info, true
	}
	
	return nil, false
}

// SetVerificationInfo stores verification info in cache
func (cs *CacheService) SetVerificationInfo(docID string, info map[string]interface{}) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	cs.cache["verify:"+docID] = &CacheItem{
		Data:      info,
		ExpiresAt: time.Now().Add(cs.ttl),
	}
}

// Delete removes an item from cache
func (cs *CacheService) Delete(key string) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	delete(cs.cache, key)
}

// Clear removes all items from cache
func (cs *CacheService) Clear() {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	cs.cache = make(map[string]*CacheItem)
}

// cleanup removes expired items from cache
func (cs *CacheService) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		cs.mutex.Lock()
		now := time.Now()
		for key, item := range cs.cache {
			if now.After(item.ExpiresAt) {
				delete(cs.cache, key)
			}
		}
		cs.mutex.Unlock()
	}
}

// Stats returns cache statistics
func (cs *CacheService) Stats() map[string]interface{} {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	
	return map[string]interface{}{
		"total_items": len(cs.cache),
		"ttl_minutes": cs.ttl.Minutes(),
	}
}