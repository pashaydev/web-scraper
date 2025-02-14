package handlersArgs

import (
	"golang.org/x/time/rate"
	"time"
	"web-scraper/internal/ai"
	"web-scraper/internal/config"
)

type HandlersArgs struct{}

var (
	limiter      = rate.NewLimiter(rate.Every(1*time.Second), config.Config.RateLimit)
	openAIClient *ai.OpenAIClient
)

func init() {
	openAIClient = ai.NewOpenAIClient(config.Config.OpenAIKey)

}

func GetOpenAiClient() *ai.OpenAIClient {
	if openAIClient == nil {
		openAIClient = ai.NewOpenAIClient(config.Config.OpenAIKey)
	}

	return openAIClient
}

func GetLimiter() *rate.Limiter {
	return limiter
}
