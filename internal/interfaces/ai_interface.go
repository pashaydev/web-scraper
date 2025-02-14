package ai

type AIProvider interface {
	ProcessSearchResults(results []SearchResult) (string, error)
}

type SearchResult struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	Link    string `json:"link"`
	Source  string `json:"source"`
}
