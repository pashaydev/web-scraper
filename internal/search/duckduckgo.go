package search

import (
	"github.com/gocolly/colly/v2"
	"strings"
	"web-scraper/internal/models"
)

type DuckDuckGoSearch struct{}

func (d *DuckDuckGoSearch) Search(query string) ([]models.SearchResult, error) {
	var results []models.SearchResult
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	c.OnHTML("article.result", func(e *colly.HTMLElement) {
		result := models.SearchResult{
			Title:   e.ChildText("h2"),
			Snippet: e.ChildText("div.result__snippet"),
			Link:    e.ChildAttr("a.result__url", "href"),
			Source:  "DuckDuckGo",
		}
		if result.Title != "" && result.Link != "" {
			results = append(results, result)
		}
	})

	searchURL := "https://duckduckgo.com/html/?q=" + strings.ReplaceAll(query, " ", "+")
	err := c.Visit(searchURL)
	return results, err
}
