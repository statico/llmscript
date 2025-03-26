package llm

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Test represents a test case for a script
type Test struct {
	Name        string
	Setup       []string
	Input       string
	Expected    string
	Timeout     time.Duration
	Environment map[string]string
}

// TestFailure represents a failed test case
type TestFailure struct {
	Test     Test
	Output   string
	Error    error
	ExitCode int
}

// Provider defines the interface for LLM providers
type Provider interface {
	// GenerateScripts creates a main script and test script from a natural language description
	GenerateScripts(ctx context.Context, description string) (ScriptPair, error)
	// FixScripts attempts to fix both scripts based on test failures
	FixScripts(ctx context.Context, scripts ScriptPair, error string) (ScriptPair, error)
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

// NewProvider creates a new LLM provider based on the provider name
func NewProvider(name string, config interface{}) (Provider, error) {
	if name == "" {
		name = "ollama" // Default to Ollama if no provider specified
	}

	// Default Ollama config
	ollamaConfig := OllamaConfig{
		Model: "llama3.2",
		Host:  "http://localhost:11434",
	}

	// If config is provided, try to extract values
	if config != nil {
		// Try to convert the config to a map[string]interface{} first
		if cfgMap, ok := config.(map[string]interface{}); ok {
			// Handle Ollama config
			if ollamaCfg, ok := cfgMap["ollama"].(map[string]interface{}); ok {
				if model, ok := ollamaCfg["model"].(string); ok && model != "" {
					ollamaConfig.Model = model
				}
				if host, ok := ollamaCfg["host"].(string); ok && host != "" {
					ollamaConfig.Host = host
				}
			}
		}
	}

	switch name {
	case "ollama":
		return NewOllamaProvider(ollamaConfig)
	case "claude":
		if config == nil {
			return nil, fmt.Errorf("a Claude API key is required")
		}
		// Try to convert the config to a map[string]interface{}
		cfgMap, ok := config.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid config type for Claude provider")
		}
		// Extract Claude config
		claudeCfg, ok := cfgMap["claude"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("missing Claude configuration")
		}
		apiKey, ok := claudeCfg["api_key"].(string)
		if !ok || apiKey == "" {
			return nil, fmt.Errorf("a Claude API key is required")
		}
		model, _ := claudeCfg["model"].(string)
		if model == "" {
			model = "claude-3-opus-20240229" // Default model
		}
		return NewClaudeProvider(ClaudeConfig{
			APIKey: apiKey,
			Model:  model,
		})
	case "openai":
		if config == nil {
			return nil, fmt.Errorf("an OpenAI API key is required")
		}
		// Try to convert the config to a map[string]interface{}
		cfgMap, ok := config.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid config type for OpenAI provider")
		}
		// Extract OpenAI config
		openaiCfg, ok := cfgMap["openai"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("missing OpenAI configuration")
		}
		apiKey, ok := openaiCfg["api_key"].(string)
		if !ok || apiKey == "" {
			return nil, fmt.Errorf("an OpenAI API key is required")
		}
		model, _ := openaiCfg["model"].(string)
		if model == "" {
			model = "gpt-4-turbo-preview" // Default model
		}
		return NewOpenAIProvider(OpenAIConfig{
			APIKey: apiKey,
			Model:  model,
		})
	default:
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
}
