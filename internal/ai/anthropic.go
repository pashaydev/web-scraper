package ai

import (
	"context"
	"errors"
	"fmt"
	"github.com/liushuangls/go-anthropic/v2"
	"strings"
	ai "web-scraper/internal/interfaces"
)

type AnthropicClient struct {
	client *anthropic.Client
}

func NewAnthropicClient(apiKey string) *AnthropicClient {
	return &AnthropicClient{
		client: anthropic.NewClient(apiKey),
	}
}

func (a *AnthropicClient) ProcessSearchResults(results []ai.SearchResult) (string, error) {
	// Convert search results to a formatted string
	var input strings.Builder
	input.WriteString("Please format the following search results in markdown format:\n\n")

	for _, result := range results {
		input.WriteString(fmt.Sprintf("Title: %s\nSnippet: %s\nLink: %s\nSource: %s\n\n",
			result.Title, result.Snippet, result.Link, result.Source))
	}

	resp, err := a.client.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model: anthropic.ModelClaude3Haiku20240307,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage(input.String()),
		},
		MaxTokens: 2000,
	})

	if err != nil {
		var e *anthropic.APIError
		if errors.As(err, &e) {
			fmt.Printf("Messages error, type: %s, message: %s", e.Type, e.Message)
		} else {
			fmt.Printf("Messages error: %v\n", err)
		}
		return "", err
	}
	var resultText string = resp.Content[0].GetText()
	fmt.Println(resultText)
	return resultText, nil
}
