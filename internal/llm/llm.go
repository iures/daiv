package llm

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

// Client represents an LLM client
type Client struct {
	llm llms.Model
}

// NewClient creates a new LLM client with Anthropic as the default provider
func NewClient() (*Client, error) {
	apiKey := viper.GetString("llm.anthropic.apikey")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is not set")
	}

	llm, err := anthropic.New(
		anthropic.WithToken(apiKey),
		anthropic.WithModel("claude-3-5-sonnet-20241022"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Anthropic client: %w", err)
	}

	return &Client{
		llm: llm,
	}, nil
}

func (c *Client) GenerateFromSinglePrompt(prompt string) (string, error) {
	ctx := context.Background()

	completion, err := llms.GenerateFromSinglePrompt(ctx, c.llm, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	return completion, nil
}
