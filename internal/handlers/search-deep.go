package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	"web-scraper/internal/cache"
	"web-scraper/internal/handlersArgs"
	"web-scraper/internal/models"
	"web-scraper/internal/search"
)

func SearchDeepHandler(w http.ResponseWriter, r *http.Request) {
	// Limitation
	var limiter = handlersArgs.GetLimiter()
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
			log.Println("Searching engine:", engine.GetName())
			log.Println("Searching query:", query)
			results, err := e.DeepSearch(query)
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

	// Process with OpenAI
	openAIResult, err := getOpenAIResults(query, allResults)
	if err != nil {
		log.Printf("OpenAI error: %v", err)
		openAIResult = "Error processing results with AI"
	}

	// Create response
	response := models.SearchResponse{
		Query:           query,
		Results:         allResults,
		FormattedResult: openAIResult,
		Duration:        time.Since(startTime).String(),
	}

	// Store in cache
	err1 := cache.GetInstance().Set(query, response)
	if err1 != nil {
		log.Printf("Error caching response: %v", err1)
	}

	// Send response
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func formatSearchResultsForAI(query string, results []models.SearchResult) string {
	var formattedText strings.Builder

	// Add the query
	formattedText.WriteString(fmt.Sprintf("Search Query: %s\n\n", query))

	// Add search results
	formattedText.WriteString("Search Results:\n")
	for _, result := range results {
		formattedText.WriteString(fmt.Sprintf("\nTitle: %s\n", result.Title))
		formattedText.WriteString(fmt.Sprintf("URL: %s\n", result.Link))
		formattedText.WriteString(fmt.Sprintf("Description: %s\n", result.Snippet))
		formattedText.WriteString(fmt.Sprintf("PageContent: %s\n", result.InnerContent))
	}

	// Add instructions
	formattedText.WriteString(`
		Instructions:
		You are tasked with generating a response based on the search results from a given query. The goal is to summarize the key information and insights from the search results in a clear and concise manner.
		1. Review the search results and identify the most relevant and important information.
		2. Summarize the key points and insights from the search results.
		3. Provide a brief overview of the main topics and themes covered in the search results.
		3. Use simple, straightforward language that is easy to understand.
		4. Avoid repeating information or including unnecessary details.
		5. Keep the response concise and focused on the main points.
		7. Attach links to the original sources of information under each point.
	`)

	return formattedText.String()
}

func getOpenAIResults(query string, results []models.SearchResult) (string, error) {
	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		var openAIClient = handlersArgs.GetOpenAiClient()
		result, err := openAIClient.FormatResults(formatSearchResultsForAI(query, results))
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	select {
	case err := <-errChan:
		return "", err
	case result := <-resultChan:
		return result, nil
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("AI processing timed out")
	}
}
