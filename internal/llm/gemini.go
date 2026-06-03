package llm

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// geminiGenerator generates text using Google's Gemini via the official
// google.golang.org/genai SDK (the unified replacement for the deprecated
// generative-ai-go package).
type geminiGenerator struct {
	client *genai.Client
	model  string
}

// newGeminiGenerator creates a Gemini-backed generator. The API key is passed
// explicitly rather than relying on env vars, since the SDK silently prefers
// GOOGLE_API_KEY over GEMINI_API_KEY when both are set.
func newGeminiGenerator(ctx context.Context, config GeminiConfig) (*geminiGenerator, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  config.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}
	return &geminiGenerator{client: client, model: config.Model}, nil
}

func (g *geminiGenerator) name() string { return "Gemini" }

func (g *geminiGenerator) generate(ctx context.Context, prompt string) (string, error) {
	result, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("request to Gemini failed: %w", err)
	}
	text := result.Text()
	if text == "" {
		return "", fmt.Errorf("empty response from Gemini (check finish reason / safety filters)")
	}
	return text, nil
}
