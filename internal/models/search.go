package models

type SearchResult struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	Link    string `json:"link"`
	Source  string `json:"source"`
}

type SearchResponse struct {
	Query    string         `json:"query"`
	Results  []SearchResult `json:"results"`
	Duration string         `json:"duration"`
}
