package models

type SearchResult struct {
	Title        string `json:"title"`
	Snippet      string `json:"snippet"`
	Link         string `json:"link"`
	InnerContent string `json:"inner_content"`
	Source       string `json:"source"`
}

type SearchResponse struct {
	Query           string         `json:"query"`
	Results         []SearchResult `json:"results"`
	FormattedResult string         `json:"formatted_result"`
	Duration        string         `json:"duration"`
}
