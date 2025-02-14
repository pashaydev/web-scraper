package search

import (
	"github.com/gocolly/colly/v2"
	"log"
	"strings"
	"web-scraper/internal/models"
)

type BingSearch struct{}

func (g *BingSearch) GetName() string {
	return "Bing"
}

func (b *BingSearch) Search(query string) ([]models.SearchResult, error) {
	var results []models.SearchResult
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4473.124 Safari/537.36"),
	)

	c.OnHTML("li.b_algo", func(e *colly.HTMLElement) {
		result := models.SearchResult{
			Title:   e.ChildText("h2"),
			Snippet: e.ChildText("div.b_caption p"),
			Link:    e.ChildAttr("a", "href"),
			Source:  "Bing",
		}
		if result.Title != "" && result.Link != "" {
			results = append(results, result)
		}
	})

	searchURL := "https://www.bing.com/search?q=" + strings.ReplaceAll(query, " ", "+")
	err := c.Visit(searchURL)
	log.Println("Search results: ", results)

	return results, err
}
