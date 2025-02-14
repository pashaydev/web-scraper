package search

import "web-scraper/internal/models"

type SearchEngine interface {
	Search(query string) ([]models.SearchResult, error)
	DeepSearch(query string) ([]models.SearchResult, error)
	GetName() string
}
