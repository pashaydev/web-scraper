package ai

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"strings"
	"web-scraper/internal/models"
)

type OpenAIClient struct {
	client *openai.Client
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
	}
}

func (o *OpenAIClient) ProcessSearchResults(results []models.SearchResult) (string, error) {
	// Convert search results to a formatted string
	var input strings.Builder
	input.WriteString("Please format the following search results in markdown format:\n\n")

	for _, result := range results {
		input.WriteString(fmt.Sprintf("Title: %s\nSnippet: %s\nLink: %s\nSource: %s\n\n",
			result.Title, result.Snippet, result.Link, result.Source))
	}

	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: input.String(),
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %v", err)
	}

	return resp.Choices[0].Message.Content, nil
}

func (o *OpenAIClient) FormatResults(input string) (string, error) {
	resp, err := o.client.CreateChatCompletion(
		context.Background(),

		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: input,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %v", err)
	}

	return resp.Choices[0].Message.Content, nil
}
