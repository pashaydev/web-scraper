package search

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"net/url"
	"strings"
	"time"
	"web-scraper/internal/models"
)

type DuckDuckGoSearch struct{}

func (g *DuckDuckGoSearch) GetName() string {
	return "DuckDuckGo"
}

func (d *DuckDuckGoSearch) Search(query string) ([]models.SearchResult, error) {
	var results []models.SearchResult
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4473.124 Safari/537.36"),
	)

	c.OnHTML(".result", func(e *colly.HTMLElement) {
		result := models.SearchResult{
			Title:   e.ChildText("h2"),
			Snippet: e.ChildText(".result__snippet"),
			Link:    e.ChildAttr("a.result__a", "href"),
			Source:  "DuckDuckGo",
		}
		if result.Title != "" && result.Link != "" {
			results = append(results, result)
		}
	})

	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))
	err := c.Visit(searchURL)
	return results, err
}

func (d *DuckDuckGoSearch) DeepSearch(query string) ([]models.SearchResult, error) {
	var results []models.SearchResult
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4473.124 Safari/537.36"),
	)

	var innerPageContents []string

	// First, collect the search results
	c.OnHTML(".result", func(e *colly.HTMLElement) {
		result := models.SearchResult{
			Title:   e.ChildText("h2"),
			Snippet: e.ChildText(".result__snippet"),
			Link:    e.ChildAttr("a.result__a", "href"),
			Source:  "DuckDuckGo",
		}
		if result.Title != "" && result.Link != "" {
			results = append(results, result)
		}
	})

	// Visit DuckDuckGo search page
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))
	err := c.Visit(searchURL)
	if err != nil {
		return nil, err
	}

	// Create a new collector for scraping individual pages
	innerCollector := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4473.124 Safari/537.36"),
		colly.MaxDepth(1),
	)

	// Set timeout for requests
	innerCollector.SetRequestTimeout(10 * time.Second)

	// Configure inner page scraping with smarter content extraction
	innerCollector.OnHTML("html", func(e *colly.HTMLElement) {
		var contentBuilder strings.Builder

		// Extract meta description
		metaDesc := e.ChildAttr("meta[name='description']", "content")
		if metaDesc != "" {
			contentBuilder.WriteString("Description: " + metaDesc + "\n\n")
		}

		// Extract main content areas
		contentSelectors := []string{
			"article",
			"main",
			".content",
			"#content",
			".post-content",
			".article-content",
			"[role='main']",
		}

		for _, selector := range contentSelectors {
			e.ForEach(selector, func(_ int, el *colly.HTMLElement) {
				// Clean and append the text
				text := cleanText(el.Text)
				if text != "" {
					contentBuilder.WriteString(text + "\n")
				}
			})
		}

		// If no main content areas found, fall back to paragraph text
		if contentBuilder.Len() == 0 {
			e.ForEach("p", func(_ int, el *colly.HTMLElement) {
				text := cleanText(el.Text)
				if text != "" {
					contentBuilder.WriteString(text + "\n")
				}
			})
		}

		// Get the final content
		content := contentBuilder.String()
		content = strings.TrimSpace(content)

		// Limit content length
		if len(content) > 5000 {
			content = content[:5000]
		}

		innerPageContents = append(innerPageContents, content)
	})

	// Error handling for requests
	innerCollector.OnError(func(r *colly.Response, err error) {
		log.Printf("Error scraping %s: %v", r.Request.URL, err)
		innerPageContents = append(innerPageContents, "")
	})

	// Visit each result URL to get inner page content
	for i, result := range results {
		if i >= 10 { // Limit to first 10 results
			break
		}

		if result.Link != "" && (strings.HasPrefix(result.Link, "http://") || strings.HasPrefix(result.Link, "https://")) {
			err := innerCollector.Visit(result.Link)
			if err != nil {
				log.Printf("Error visiting %s: %v", result.Link, err)
				innerPageContents = append(innerPageContents, "")
				continue
			}
		}
	}

	// Combine results with inner page contents
	for i := range results {
		if i < len(innerPageContents) {
			results[i].InnerContent = innerPageContents[i]
		}
	}

	log.Printf("Deep search completed for %s. Found %d results", d.GetName(), len(results))
	return results, nil
}

// Helper function to clean text
func cleanText(text string) string {
	// Remove extra whitespace
	text = strings.Join(strings.Fields(text), " ")
	// Remove any special characters or unnecessary whitespace
	text = strings.TrimSpace(text)
	return text
}
