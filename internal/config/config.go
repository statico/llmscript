package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LLM struct {
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
	} `yaml:"llm"`
	MaxFixes    int           `yaml:"max_fixes"`
	MaxAttempts int           `yaml:"max_attempts"`
	Timeout     time.Duration `yaml:"timeout"`
	ExtraPrompt string        `yaml:"additional_prompt"`
}

func DefaultConfig() *Config {
	return &Config{
		MaxFixes:    10,
		MaxAttempts: 3,
		Timeout:     30 * time.Second,
	}
}

func interpolateEnvVars(data []byte) []byte {
	content := string(data)
	content = os.ExpandEnv(content)
	return []byte(content)
}

func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		var err error
		configDir, err = os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get config dir: %w", err)
		}
	}

	configPath := filepath.Join(configDir, "llmscript", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Interpolate environment variables before unmarshaling
	data = interpolateEnvVars(data)

	// Create a temporary config to unmarshal into
	var loadedConfig Config
	if err := yaml.Unmarshal(data, &loadedConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge the loaded config with the default config
	config.LLM.Provider = loadedConfig.LLM.Provider
	config.LLM.Ollama.Model = loadedConfig.LLM.Ollama.Model
	config.LLM.Ollama.Host = loadedConfig.LLM.Ollama.Host
	config.LLM.Claude.APIKey = loadedConfig.LLM.Claude.APIKey
	config.LLM.Claude.Model = loadedConfig.LLM.Claude.Model
	config.LLM.OpenAI.APIKey = loadedConfig.LLM.OpenAI.APIKey
	config.LLM.OpenAI.Model = loadedConfig.LLM.OpenAI.Model

	// Only update numeric values if they are explicitly set in the YAML
	if loadedConfig.MaxFixes != 0 {
		config.MaxFixes = loadedConfig.MaxFixes
	}
	if loadedConfig.MaxAttempts != 0 {
		config.MaxAttempts = loadedConfig.MaxAttempts
	}
	if loadedConfig.Timeout != 0 {
		config.Timeout = loadedConfig.Timeout
	}

	config.ExtraPrompt = loadedConfig.ExtraPrompt

	return config, nil
}

func WriteConfig(config *Config) error {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		var err error
		configDir, err = os.UserConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get config dir: %w", err)
		}
	}

	llmscriptDir := filepath.Join(configDir, "llmscript")
	if err := os.MkdirAll(llmscriptDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	encoder.Close()

	configPath := filepath.Join(llmscriptDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
