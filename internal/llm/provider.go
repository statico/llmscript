package llm

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Default models for each provider. These target the most capable tier as of
// 2026; users can override any of them via config or the --llm.model flag.
const (
	DefaultOllamaModel     = "llama3.3"
	DefaultOllamaHost      = "http://localhost:11434"
	DefaultClaudeModel     = "claude-opus-4-8"
	DefaultOpenAIModel     = "gpt-5.5"
	DefaultGeminiModel     = "gemini-2.5-pro"
	DefaultOpenRouterModel = "anthropic/claude-opus-4.8"

	// openRouterBaseURL is OpenRouter's OpenAI-compatible API endpoint.
	openRouterBaseURL = "https://openrouter.ai/api/v1"
)

// Provider defines the interface for LLM providers
type Provider interface {
	// GenerateScripts creates a main script and test script from a natural language description
	GenerateScripts(ctx context.Context, description string) (ScriptPair, error)
	// FixScripts attempts to fix the main script based on a test failure
	FixScripts(ctx context.Context, scripts ScriptPair, failure string) (ScriptPair, error)
	// Name returns a human-readable name for the provider
	Name() string
}

// GetPlatformInfo returns information about the current platform
func GetPlatformInfo() string {
	info := []string{
		"Operating System: " + runtime.GOOS,
		"Architecture: " + runtime.GOARCH,
	}

	// Get additional system information using uname
	cmd := exec.Command("uname", "-a")
	if output, err := cmd.Output(); err == nil {
		info = append(info, "System Info: "+string(output))
	}

	// Get shell information
	cmd = exec.Command("bash", "--version")
	if output, err := cmd.Output(); err == nil {
		info = append(info, "Shell Info: "+string(output))
	}

	return strings.Join(info, "\n")
}

// NewProvider creates a new LLM provider from a fully-resolved Config.
func NewProvider(cfg Config) (Provider, error) {
	provider := cfg.Provider
	if provider == "" {
		provider = "ollama" // Default to Ollama if no provider specified
	}

	var gen generator
	var err error

	switch provider {
	case "ollama":
		oc := cfg.Ollama
		if oc.Model == "" {
			oc.Model = DefaultOllamaModel
		}
		if oc.Host == "" {
			oc.Host = DefaultOllamaHost
		}
		gen = newOllamaGenerator(oc)

	case "claude", "anthropic":
		if cfg.Claude.APIKey == "" {
			return nil, fmt.Errorf("a Claude API key is required")
		}
		model := cfg.Claude.Model
		if model == "" {
			model = DefaultClaudeModel
		}
		gen = newClaudeGenerator(ClaudeConfig{APIKey: cfg.Claude.APIKey, Model: model})

	case "openai":
		if cfg.OpenAI.APIKey == "" {
			return nil, fmt.Errorf("an OpenAI API key is required")
		}
		model := cfg.OpenAI.Model
		if model == "" {
			model = DefaultOpenAIModel
		}
		gen = newOpenAIGenerator(cfg.OpenAI.APIKey, model)

	case "openrouter":
		if cfg.OpenRouter.APIKey == "" {
			return nil, fmt.Errorf("an OpenRouter API key is required")
		}
		model := cfg.OpenRouter.Model
		if model == "" {
			model = DefaultOpenRouterModel
		}
		gen = newOpenRouterGenerator(cfg.OpenRouter.APIKey, model)

	case "gemini", "google":
		if cfg.Gemini.APIKey == "" {
			return nil, fmt.Errorf("a Gemini API key is required")
		}
		model := cfg.Gemini.Model
		if model == "" {
			model = DefaultGeminiModel
		}
		gen, err = newGeminiGenerator(context.Background(), GeminiConfig{APIKey: cfg.Gemini.APIKey, Model: model})
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini client: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	return &scriptProvider{gen: gen, extraPrompt: cfg.ExtraPrompt}, nil
}
