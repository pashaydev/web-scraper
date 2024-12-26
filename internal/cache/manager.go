package cache

import (
	"encoding/json"
	"fmt"
	"github.com/patrickmn/go-cache"
	"sort"
	"sync"
	"time"
	"web-scraper/internal/config"
	"web-scraper/internal/models"
)

// CacheStats holds statistics about the cache
type CacheStats struct {
	ItemCount int
	BytesUsed int64
}

type CacheItem struct {
	Value     models.SearchResponse
	CreatedAt time.Time
	Size      int64
}

// CacheManager handles cache operations with size limitations
type CacheManager struct {
	cache     *cache.Cache
	mutex     sync.RWMutex
	bytesUsed int64
	maxBytes  int64
	maxItems  int
}

// NewCacheManager Create a new CacheManager
func NewCacheManager(maxItems int, maxBytes int64, defaultExpiration, cleanupInterval time.Duration) *CacheManager {
	return &CacheManager{
		cache:    cache.New(defaultExpiration, cleanupInterval),
		maxBytes: maxBytes,
		maxItems: maxItems,
	}
}

// Set adds an item to the cache with size checking
func (cm *CacheManager) Set(key string, value models.SearchResponse) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Calculate size of new item
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("error calculating item size: %v", err)
	}

	itemSize := int64(len(valueBytes))

	// Create cache item with metadata
	item := CacheItem{
		Value:     value,
		CreatedAt: time.Now(),
		Size:      itemSize,
	}

	// Check if adding this item would exceed the maximum cache size
	if cm.bytesUsed+itemSize > cm.maxBytes {
		cm.evictOldest()
	}

	// Check if we're at the maximum number of items
	if cm.cache.ItemCount() >= cm.maxItems {
		cm.evictOldest()
	}

	// Add the item to the cache
	cm.cache.Set(key, item, cache.DefaultExpiration)
	cm.bytesUsed += itemSize

	return nil
}

// Get retrieves an item from the cache
func (cm *CacheManager) Get(key string) (models.SearchResponse, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if value, found := cm.cache.Get(key); found {
		item := value.(CacheItem)
		return item.Value, true
	}
	return models.SearchResponse{}, false
}

func (cm *CacheManager) evictOldest() {
	items := cm.cache.Items()

	type itemInfo struct {
		key       string
		createdAt time.Time
		size      int64
	}

	var itemsList []itemInfo
	for k, v := range items {
		item := v.Object.(CacheItem)
		itemsList = append(itemsList, itemInfo{
			key:       k,
			createdAt: item.CreatedAt,
			size:      item.Size,
		})
	}

	// Sort by creation time
	sort.Slice(itemsList, func(i, j int) bool {
		return itemsList[i].createdAt.Before(itemsList[j].createdAt)
	})

	// Remove oldest items until we're under the limits
	for _, item := range itemsList {
		if cm.cache.ItemCount() < cm.maxItems && cm.bytesUsed < cm.maxBytes {
			break
		}
		cm.cache.Delete(item.key)
		cm.bytesUsed -= item.size
	}
}

// GetStats returns current cache statistics
func (cm *CacheManager) GetStats() CacheStats {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return CacheStats{
		ItemCount: cm.cache.ItemCount(),
		BytesUsed: cm.bytesUsed,
	}
}

// Clear removes all items from the cache
func (cm *CacheManager) Clear() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.cache.Flush()
	cm.bytesUsed = 0
}

// GetMetrics Add a method to get cache metrics
func (cm *CacheManager) GetMetrics() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return map[string]interface{}{
		"item_count":         cm.cache.ItemCount(),
		"bytes_used":         cm.bytesUsed,
		"max_bytes":          cm.maxBytes,
		"max_items":          cm.maxItems,
		"bytes_used_percent": float64(cm.bytesUsed) / float64(cm.maxBytes) * 100,
		"items_used_percent": float64(cm.cache.ItemCount()) / float64(cm.maxItems) * 100,
	}
}

var (
	instance *CacheManager
	once     sync.Once
)

// GetInstance returns the singleton instance of CacheManager
func GetInstance() *CacheManager {
	once.Do(func() {
		instance = NewCacheManager(
			config.Config.MaxCacheSize,
			config.Config.MaxCacheBytes,
			config.Config.CacheDuration,
			10*time.Minute,
		)
	})
	return instance
}
