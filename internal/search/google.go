package search

import (
	"github.com/gocolly/colly/v2"
	"strings"
	"web-scraper/internal/models"
)

type GoogleSearch struct{}

func (g *GoogleSearch) Search(query string) ([]models.SearchResult, error) {

	var results []models.SearchResult
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	c.OnHTML("div.g", func(e *colly.HTMLElement) {
		result := models.SearchResult{
			Title:   e.ChildText("h3"),
			Snippet: e.ChildText("div.VwiC3b"),
			Link:    e.ChildAttr("a", "href"),
			Source:  "Google",
		}
		if result.Title != "" && result.Link != "" {
			results = append(results, result)
		}
	})

	searchURL := "https://www.google.com/search?q=" + strings.ReplaceAll(query, " ", "+")
	err := c.Visit(searchURL)
	return results, err
}
