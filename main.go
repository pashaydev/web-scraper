//package main
//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/gocolly/colly/v2"
//	"github.com/patrickmn/go-cache"
//	"golang.org/x/time/rate"
//	"log"
//	"net/http"
//	"sort"
//	"strings"
//	"sync"
//	"time"
//)
//
//type Config struct {
//	Port          string
//	RateLimit     int
//	CacheDuration time.Duration
//	MaxCacheSize  int
//	MaxCacheBytes int64 // Maximum cache size in bytes
//}
//
//var config Config = Config{
//	Port:          "8080",
//	RateLimit:     5,
//	CacheDuration: 5 * time.Minute,
//	MaxCacheSize:  1000,             // Cache up to 1000 search results
//	MaxCacheBytes: 50 * 1024 * 1024, // 50MB maximum cache size
//
//}
//
//// CacheStats holds statistics about the cache
//type CacheStats struct {
//	ItemCount int
//	BytesUsed int64
//}
//
//type CacheItem struct {
//	Value     SearchResponse
//	CreatedAt time.Time
//	Size      int64
//}
//
//// CacheManager handles cache operations with size limitations
//type CacheManager struct {
//	cache     *cache.Cache
//	mutex     sync.RWMutex
//	bytesUsed int64
//	maxBytes  int64
//	maxItems  int
//}
//
//// NewCacheManager Create a new CacheManager
//func NewCacheManager(maxItems int, maxBytes int64, defaultExpiration, cleanupInterval time.Duration) *CacheManager {
//	return &CacheManager{
//		cache:    cache.New(defaultExpiration, cleanupInterval),
//		maxBytes: maxBytes,
//		maxItems: maxItems,
//	}
//}
//
//// Initialize the cache manager
//var cacheManager = NewCacheManager(
//	config.MaxCacheSize,
//	config.MaxCacheBytes,
//	config.CacheDuration,
//	10*time.Minute,
//)
//
//// Set adds an item to the cache with size checking
//func (cm *CacheManager) Set(key string, value SearchResponse) error {
//	cm.mutex.Lock()
//	defer cm.mutex.Unlock()
//
//	// Calculate size of new item
//	valueBytes, err := json.Marshal(value)
//	if err != nil {
//		return fmt.Errorf("error calculating item size: %v", err)
//	}
//
//	itemSize := int64(len(valueBytes))
//
//	// Create cache item with metadata
//	item := CacheItem{
//		Value:     value,
//		CreatedAt: time.Now(),
//		Size:      itemSize,
//	}
//
//	// Check if adding this item would exceed the maximum cache size
//	if cm.bytesUsed+itemSize > cm.maxBytes {
//		cm.evictOldest()
//	}
//
//	// Check if we're at the maximum number of items
//	if cm.cache.ItemCount() >= cm.maxItems {
//		cm.evictOldest()
//	}
//
//	// Add the item to the cache
//	cm.cache.Set(key, item, cache.DefaultExpiration)
//	cm.bytesUsed += itemSize
//
//	return nil
//}
//
//// Get retrieves an item from the cache
//func (cm *CacheManager) Get(key string) (SearchResponse, bool) {
//	cm.mutex.RLock()
//	defer cm.mutex.RUnlock()
//
//	if value, found := cm.cache.Get(key); found {
//		item := value.(CacheItem)
//		return item.Value, true
//	}
//	return SearchResponse{}, false
//}
//
//func (cm *CacheManager) evictOldest() {
//	items := cm.cache.Items()
//
//	type itemInfo struct {
//		key       string
//		createdAt time.Time
//		size      int64
//	}
//
//	var itemsList []itemInfo
//	for k, v := range items {
//		item := v.Object.(CacheItem)
//		itemsList = append(itemsList, itemInfo{
//			key:       k,
//			createdAt: item.CreatedAt,
//			size:      item.Size,
//		})
//	}
//
//	// Sort by creation time
//	sort.Slice(itemsList, func(i, j int) bool {
//		return itemsList[i].createdAt.Before(itemsList[j].createdAt)
//	})
//
//	// Remove oldest items until we're under the limits
//	for _, item := range itemsList {
//		if cm.cache.ItemCount() < cm.maxItems && cm.bytesUsed < cm.maxBytes {
//			break
//		}
//		cm.cache.Delete(item.key)
//		cm.bytesUsed -= item.size
//	}
//}
//
//// GetStats returns current cache statistics
//func (cm *CacheManager) GetStats() CacheStats {
//	cm.mutex.RLock()
//	defer cm.mutex.RUnlock()
//
//	return CacheStats{
//		ItemCount: cm.cache.ItemCount(),
//		BytesUsed: cm.bytesUsed,
//	}
//}
//
//// Clear removes all items from the cache
//func (cm *CacheManager) Clear() {
//	cm.mutex.Lock()
//	defer cm.mutex.Unlock()
//
//	cm.cache.Flush()
//	cm.bytesUsed = 0
//}
//
//// GetMetrics Add a method to get cache metrics
//func (cm *CacheManager) GetMetrics() map[string]interface{} {
//	cm.mutex.RLock()
//	defer cm.mutex.RUnlock()
//
//	return map[string]interface{}{
//		"item_count":         cm.cache.ItemCount(),
//		"bytes_used":         cm.bytesUsed,
//		"max_bytes":          cm.maxBytes,
//		"max_items":          cm.maxItems,
//		"bytes_used_percent": float64(cm.bytesUsed) / float64(cm.maxBytes) * 100,
//		"items_used_percent": float64(cm.cache.ItemCount()) / float64(cm.maxItems) * 100,
//	}
//}
//
//// Add a cache stats endpoint
//func cacheStatsHandler(w http.ResponseWriter, r *http.Request) {
//	stats := cacheManager.GetStats()
//	err := json.NewEncoder(w).Encode(stats)
//	if err != nil {
//		log.Println("Error encoding metrics", err.Error())
//		return
//	}
//}
//
//// Add a metrics endpoint handler
//func cacheMetricsHandler(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Content-Type", "application/json")
//	metrics := cacheManager.GetMetrics()
//	err := json.NewEncoder(w).Encode(metrics)
//	if err != nil {
//		log.Println("Error encoding metrics", err.Error())
//		return
//	}
//}
//
//var limiter = rate.NewLimiter(rate.Every(1*time.Second), config.RateLimit)
//
//var searchCache = cache.New(config.CacheDuration, 10*time.Minute)
//
//// SearchResult represents a single search result
//type SearchResult struct {
//	Title   string `json:"title"`
//	Snippet string `json:"snippet"`
//	Link    string `json:"link"`
//	Source  string `json:"source"` // Which search engine provided this result
//}
//
//// SearchResponse represents the complete response
//type SearchResponse struct {
//	Query    string         `json:"query"`
//	Results  []SearchResult `json:"results"`
//	Duration string         `json:"duration"`
//}
//
//// SearchEngine interface defines the contract for different search engine scrapers
//type SearchEngine interface {
//	Search(query string) ([]SearchResult, error)
//}
//
//// GoogleSearch Implementation
//type GoogleSearch struct{}
//
//func (g *GoogleSearch) Search(query string) ([]SearchResult, error) {
//
//	var results []SearchResult
//	c := colly.NewCollector(
//		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
//	)
//
//	c.OnHTML("div.g", func(e *colly.HTMLElement) {
//		result := SearchResult{
//			Title:   e.ChildText("h3"),
//			Snippet: e.ChildText("div.VwiC3b"),
//			Link:    e.ChildAttr("a", "href"),
//			Source:  "Google",
//		}
//		if result.Title != "" && result.Link != "" {
//			results = append(results, result)
//		}
//	})
//
//	searchURL := "https://www.google.com/search?q=" + strings.ReplaceAll(query, " ", "+")
//	err := c.Visit(searchURL)
//	return results, err
//}
//
//// BingSearch Implementation
//type BingSearch struct{}
//
//func (b *BingSearch) Search(query string) ([]SearchResult, error) {
//	var results []SearchResult
//	c := colly.NewCollector(
//		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
//	)
//
//	c.OnHTML("li.b_algo", func(e *colly.HTMLElement) {
//		result := SearchResult{
//			Title:   e.ChildText("h2"),
//			Snippet: e.ChildText("div.b_caption p"),
//			Link:    e.ChildAttr("a", "href"),
//			Source:  "Bing",
//		}
//		if result.Title != "" && result.Link != "" {
//			results = append(results, result)
//		}
//	})
//
//	searchURL := "https://www.bing.com/search?q=" + strings.ReplaceAll(query, " ", "+")
//	err := c.Visit(searchURL)
//	return results, err
//}
//
//// DuckDuckGoSearch Implementation
//type DuckDuckGoSearch struct{}
//
//func (d *DuckDuckGoSearch) Search(query string) ([]SearchResult, error) {
//	var results []SearchResult
//	c := colly.NewCollector(
//		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
//	)
//
//	c.OnHTML("article.result", func(e *colly.HTMLElement) {
//		result := SearchResult{
//			Title:   e.ChildText("h2"),
//			Snippet: e.ChildText("div.result__snippet"),
//			Link:    e.ChildAttr("a.result__url", "href"),
//			Source:  "DuckDuckGo",
//		}
//		if result.Title != "" && result.Link != "" {
//			results = append(results, result)
//		}
//	})
//
//	searchURL := "https://duckduckgo.com/html/?q=" + strings.ReplaceAll(query, " ", "+")
//	err := c.Visit(searchURL)
//	return results, err
//}
//
//// SearchHandler handles the /scraper endpoint
//func SearchHandler(w http.ResponseWriter, r *http.Request) {
//	// Limitation
//	if !limiter.Allow() {
//		http.Error(w, "Too many requests", http.StatusTooManyRequests)
//		return
//	}
//
//	// Set CORS headers
//	w.Header().Set("Access-Control-Allow-Origin", "*")
//	w.Header().Set("Access-Control-Allow-Methods", "GET")
//	w.Header().Set("Content-Type", "application/json")
//
//	query := r.URL.Query().Get("search")
//	if query == "" {
//		http.Error(w, "Missing search parameter", http.StatusBadRequest)
//		return
//	}
//
//	// Check cache
//	if cached, found := cacheManager.Get(query); found {
//		log.Printf("Cache hit for query: %s", query)
//		err := json.NewEncoder(w).Encode(cached)
//		if err != nil {
//			log.Printf("Error encoding cached response: %v", err)
//			http.Error(w, "Internal server error", http.StatusInternalServerError)
//		}
//		return
//	}
//
//	// Perform search
//	startTime := time.Now()
//
//	// Create search engines
//	searchEngines := []SearchEngine{
//		&GoogleSearch{},
//		&BingSearch{},
//		&DuckDuckGoSearch{},
//	}
//
//	// Create channels for results and errors
//	resultsChan := make(chan []SearchResult, len(searchEngines))
//	errorsChan := make(chan error, len(searchEngines))
//
//	// Create wait group
//	var wg sync.WaitGroup
//
//	// Launch searches in parallel
//	for _, engine := range searchEngines {
//		wg.Add(1)
//		go func(e SearchEngine) {
//			defer wg.Done()
//			results, err := e.Search(query)
//			if err != nil {
//				errorsChan <- err
//				return
//			}
//			resultsChan <- results
//		}(engine)
//	}
//
//	// Wait for all searches to complete in a separate goroutine
//	go func() {
//		wg.Wait()
//		close(resultsChan)
//		close(errorsChan)
//	}()
//
//	// Collect all results
//	var allResults []SearchResult
//	for results := range resultsChan {
//		allResults = append(allResults, results...)
//	}
//
//	// Check for errors
//	for err := range errorsChan {
//		log.Printf("Search error: %v", err)
//	}
//
//	// Create response
//	response := SearchResponse{
//		Query:    query,
//		Results:  allResults,
//		Duration: time.Since(startTime).String(),
//	}
//
//	// Store in cache
//	err := cacheManager.Set(query, response)
//	if err != nil {
//		log.Printf("Error caching response: %v", err)
//	}
//
//	// Send response
//	err = json.NewEncoder(w).Encode(response)
//	if err != nil {
//		log.Printf("Error encoding response: %v", err)
//		http.Error(w, "Internal server error", http.StatusInternalServerError)
//		return
//	}
//}
//
//// Health check
//func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
//	w.WriteHeader(http.StatusOK)
//	_, err := w.Write([]byte("OK"))
//	if err != nil {
//		return
//	}
//}
//
//func main() {
//	mux := http.NewServeMux()
//
//	mux.HandleFunc("/scraper", RecoveryMiddleware(LoggingMiddleware(SearchHandler)))
//	mux.HandleFunc("/health", healthCheckHandler)
//	mux.HandleFunc("/cache/stats", cacheStatsHandler)
//	mux.HandleFunc("/cache/metrics", cacheMetricsHandler)
//
//	// Configure server
//	server := &http.Server{
//		Addr:         ":8080",
//		Handler:      mux,
//		ReadTimeout:  15 * time.Second,
//		WriteTimeout: 15 * time.Second,
//		IdleTimeout:  60 * time.Second,
//	}
//
//	// Start server
//	log.Println("Server starting on :8080...")
//	if err := server.ListenAndServe(); err != nil {
//		log.Fatal(err)
//	}
//}

package main

import (
	"log"
	"net/http"
	"time"
	"web-scraper/internal/config"
	"web-scraper/internal/handlers"
	"web-scraper/internal/middleware"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/scraper", middleware.RecoveryMiddleware(middleware.LoggingMiddleware(handlers.SearchHandler)))
	mux.HandleFunc("/health", handlers.HealthCheckHandler)
	mux.HandleFunc("/cache/stats", handlers.CacheStatsHandler)
	mux.HandleFunc("/cache/metrics", handlers.CacheMetricsHandler)

	server := &http.Server{
		Addr:         ":" + config.Config.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server starting on :%s...", config.Config.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
