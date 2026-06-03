package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// claudeGenerator generates text using Anthropic's Claude via the official SDK.
type claudeGenerator struct {
	client anthropic.Client
	model  string
}

// newClaudeGenerator creates a Claude-backed generator.
func newClaudeGenerator(config ClaudeConfig) *claudeGenerator {
	client := anthropic.NewClient(option.WithAPIKey(config.APIKey))
	return &claudeGenerator{client: client, model: config.Model}
}

func (g *claudeGenerator) name() string { return "Claude" }

func (g *claudeGenerator) generate(ctx context.Context, prompt string) (string, error) {
	msg, err := g.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(g.model),
		MaxTokens: 8192,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("request to Claude failed: %w", err)
	}

	var sb strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	if sb.Len() == 0 {
		return "", fmt.Errorf("no text content in Claude response")
	}
	return sb.String(), nil
}
