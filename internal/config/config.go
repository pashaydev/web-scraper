package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

type Configuration struct {
	Port           string
	RateLimit      int
	CacheDuration  time.Duration
	MaxCacheSize   int
	MaxCacheBytes  int64
	OpenAIKey      string
	AnthropicAIKey string
}

var Config Configuration

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	Config = Configuration{
		Port:           "8080",
		RateLimit:      5,
		CacheDuration:  5 * time.Minute,
		MaxCacheSize:   1000,
		MaxCacheBytes:  50 * 1024 * 1024,
		OpenAIKey:      os.Getenv("OPENAI_KEY"),
		AnthropicAIKey: os.Getenv("ANTHROPIC_AI_KEY"),
	}
}
