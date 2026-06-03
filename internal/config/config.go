package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/statico/llmscript/internal/llm"
	customlog "github.com/statico/llmscript/internal/log"
	"gopkg.in/yaml.v3"
)

// LLMConfig holds the provider selection and per-provider settings. The
// per-provider structs are reused from the llm package so there is a single
// source of truth for their shape.
type LLMConfig struct {
	Provider   string               `yaml:"provider"`
	Ollama     llm.OllamaConfig     `yaml:"ollama"`
	Claude     llm.ClaudeConfig     `yaml:"claude"`
	OpenAI     llm.OpenAIConfig     `yaml:"openai"`
	Gemini     llm.GeminiConfig     `yaml:"gemini"`
	OpenRouter llm.OpenRouterConfig `yaml:"openrouter"`
}

type Config struct {
	LLM         LLMConfig     `yaml:"llm"`
	MaxFixes    int           `yaml:"max_fixes"`
	MaxAttempts int           `yaml:"max_attempts"`
	Timeout     time.Duration `yaml:"timeout"`
	ExtraPrompt string        `yaml:"additional_prompt"`
}

func DefaultConfig() *Config {
	return &Config{
		LLM: LLMConfig{
			Provider:   "ollama",
			Ollama:     llm.OllamaConfig{Model: llm.DefaultOllamaModel, Host: llm.DefaultOllamaHost},
			Claude:     llm.ClaudeConfig{APIKey: "${ANTHROPIC_API_KEY}", Model: llm.DefaultClaudeModel},
			OpenAI:     llm.OpenAIConfig{APIKey: "${OPENAI_API_KEY}", Model: llm.DefaultOpenAIModel},
			Gemini:     llm.GeminiConfig{APIKey: "${GEMINI_API_KEY}", Model: llm.DefaultGeminiModel},
			OpenRouter: llm.OpenRouterConfig{APIKey: "${OPENROUTER_API_KEY}", Model: llm.DefaultOpenRouterModel},
		},
		MaxFixes:    10,
		MaxAttempts: 3,
		Timeout:     30 * time.Second,
		ExtraPrompt: "Use ANSI color codes to make the output more readable.",
	}
}

func interpolateEnvVars(data []byte) []byte {
	return []byte(os.ExpandEnv(string(data)))
}

func LoadConfig() (*Config, error) {
	// Start from defaults and let the config file override only the keys it
	// actually specifies (yaml.v3 leaves absent fields untouched).
	config := DefaultConfig()

	configPath, ok, err := findConfigFile()
	if err != nil {
		return nil, err
	}
	if !ok {
		customlog.Debug("Config file not found, using defaults")
		return config, nil
	}

	customlog.Debug("Found config file: %s", configPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Interpolate environment variables before unmarshaling
	data = interpolateEnvVars(data)

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	customlog.Debug("Loaded config: provider=%s", config.LLM.Provider)
	return config, nil
}

// findConfigFile locates the config file, preferring XDG_CONFIG_HOME / ~/.config
// and falling back to the OS-specific user config dir. The second return value
// reports whether a file was found.
func findConfigFile() (string, bool, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", false, fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	configPath := filepath.Join(configDir, "llmscript", "config.yaml")
	customlog.Debug("Looking for config file at: %s", configPath)
	if _, err := os.Stat(configPath); err == nil {
		return configPath, true, nil
	} else if !os.IsNotExist(err) {
		return "", false, fmt.Errorf("failed to stat config file: %w", err)
	}

	// Fall back to the OS-specific user config dir.
	osConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", false, fmt.Errorf("failed to get config dir: %w", err)
	}
	configPath = filepath.Join(osConfigDir, "llmscript", "config.yaml")
	customlog.Debug("Looking for config file at fallback location: %s", configPath)
	if _, err := os.Stat(configPath); err == nil {
		return configPath, true, nil
	} else if !os.IsNotExist(err) {
		return "", false, fmt.Errorf("failed to stat config file: %w", err)
	}

	return "", false, nil
}

func WriteConfig(config *Config) error {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	customlog.Debug("Config directory: %s", configDir)

	llmscriptDir := filepath.Join(configDir, "llmscript")
	customlog.Debug("Creating directory: %s", llmscriptDir)
	if err := os.MkdirAll(llmscriptDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(llmscriptDir, "config.yaml")
	customlog.Debug("Writing config to: %s", configPath)
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("failed to close config file: %v", err)
		}
	}()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("failed to close encoder: %w", err)
	}

	return nil
}
