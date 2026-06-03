package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// openaiCompatGenerator generates text against any OpenAI Chat Completions
// compatible endpoint using the official openai-go SDK. It backs both the
// OpenAI and OpenRouter providers, which differ only by base URL, key, model
// ID format, and optional headers.
type openaiCompatGenerator struct {
	client openai.Client
	model  string
	label  string
}

// newOpenAIGenerator creates a generator backed by OpenAI's own API.
func newOpenAIGenerator(apiKey, model string) *openaiCompatGenerator {
	return &openaiCompatGenerator{
		client: openai.NewClient(option.WithAPIKey(apiKey)),
		model:  model,
		label:  "OpenAI",
	}
}

// newOpenRouterGenerator creates a generator backed by OpenRouter, which is
// OpenAI Chat Completions compatible. The X-Title header is optional and only
// used for attribution on openrouter.ai.
func newOpenRouterGenerator(apiKey, model string) *openaiCompatGenerator {
	return &openaiCompatGenerator{
		client: openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL(openRouterBaseURL),
			// Optional attribution headers used for ranking on openrouter.ai.
			option.WithHeader("HTTP-Referer", "https://github.com/statico/llmscript"),
			option.WithHeader("X-Title", "llmscript"),
		),
		model: model,
		label: "OpenRouter",
	}
}

func (g *openaiCompatGenerator) name() string { return g.label }

func (g *openaiCompatGenerator) generate(ctx context.Context, prompt string) (string, error) {
	resp, err := g.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModel(g.model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	})
	if err != nil {
		return "", fmt.Errorf("%s request failed: %w", g.label, err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in %s response", g.label)
	}
	return resp.Choices[0].Message.Content, nil
}
