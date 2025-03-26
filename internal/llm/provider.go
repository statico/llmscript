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
		cfg, ok := config.(struct {
			Provider string `yaml:"provider"`
			Ollama   struct {
				Model string `yaml:"model"`
				Host  string `yaml:"host"`
			} `yaml:"ollama"`
			Claude struct {
				APIKey string `yaml:"api_key"`
				Model  string `yaml:"model"`
			} `yaml:"claude"`
			OpenAI struct {
				APIKey string `yaml:"api_key"`
				Model  string `yaml:"model"`
			} `yaml:"openai"`
		})
		if ok {
			if cfg.Ollama.Model != "" {
				ollamaConfig.Model = cfg.Ollama.Model
			}
			if cfg.Ollama.Host != "" {
				ollamaConfig.Host = cfg.Ollama.Host
			}
		}
	}

	switch name {
	case "ollama":
		return NewOllamaProvider(ollamaConfig)
	case "claude":
		cfg, ok := config.(struct {
			Provider string `yaml:"provider"`
			Ollama   struct {
				Model string `yaml:"model"`
				Host  string `yaml:"host"`
			} `yaml:"ollama"`
			Claude struct {
				APIKey string `yaml:"api_key"`
				Model  string `yaml:"model"`
			} `yaml:"claude"`
			OpenAI struct {
				APIKey string `yaml:"api_key"`
				Model  string `yaml:"model"`
			} `yaml:"openai"`
		})
		if !ok || cfg.Claude.APIKey == "" {
			return nil, fmt.Errorf("a Claude API key is required")
		}
		return NewClaudeProvider(ClaudeConfig{
			APIKey: cfg.Claude.APIKey,
			Model:  cfg.Claude.Model,
		})
	case "openai":
		cfg, ok := config.(struct {
			OpenAI struct {
				APIKey string `yaml:"api_key"`
				Model  string `yaml:"model"`
			} `yaml:"openai"`
		})
		if !ok || cfg.OpenAI.APIKey == "" {
			return nil, fmt.Errorf("OpenAI API key is required")
		}
		return NewOpenAIProvider(OpenAIConfig{
			APIKey: cfg.OpenAI.APIKey,
			Model:  cfg.OpenAI.Model,
		})
	default:
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
}
