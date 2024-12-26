package handlers

import (
	"encoding/json"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"sync"
	"time"
	"web-scraper/internal/cache"
	"web-scraper/internal/config"
	"web-scraper/internal/models"
	"web-scraper/internal/search"
)

var limiter = rate.NewLimiter(rate.Every(1*time.Second), config.Config.RateLimit)

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// Limitation
	if !limiter.Allow() {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query().Get("search")
	if query == "" {
		http.Error(w, "Missing search parameter", http.StatusBadRequest)
		return
	}

	// Check cache
	if cached, found := cache.GetInstance().Get(query); found {
		log.Printf("Cache hit for query: %s", query)
		err := json.NewEncoder(w).Encode(cached)
		if err != nil {
			log.Printf("Error encoding cached response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Perform search
	startTime := time.Now()

	// Create search engines
	searchEngines := []search.SearchEngine{
		&search.GoogleSearch{},
		&search.BingSearch{},
		&search.DuckDuckGoSearch{},
	}

	// Create channels for results and errors
	resultsChan := make(chan []models.SearchResult, len(searchEngines))
	errorsChan := make(chan error, len(searchEngines))

	// Create wait group
	var wg sync.WaitGroup

	// Launch searches in parallel
	for _, engine := range searchEngines {
		wg.Add(1)
		go func(e search.SearchEngine) {
			defer wg.Done()
			results, err := e.Search(query)
			if err != nil {
				errorsChan <- err
				return
			}
			resultsChan <- results
		}(engine)
	}

	// Wait for all searches to complete in a separate goroutine
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	// Collect all results
	var allResults []models.SearchResult
	for results := range resultsChan {
		allResults = append(allResults, results...)
	}

	// Check for errors
	for err := range errorsChan {
		log.Printf("Search error: %v", err)
	}

	// Create response
	response := models.SearchResponse{
		Query:    query,
		Results:  allResults,
		Duration: time.Since(startTime).String(),
	}

	// Store in cache
	err := cache.GetInstance().Set(query, response)
	if err != nil {
		log.Printf("Error caching response: %v", err)
	}

	// Send response
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
