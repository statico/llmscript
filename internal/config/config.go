package config

import (
	"fmt"
	"os"
	"path/filepath"
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

func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "llmscript", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

func WriteConfig(config *Config) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config dir: %w", err)
	}

	llmscriptDir := filepath.Join(configDir, "llmscript")
	if err := os.MkdirAll(llmscriptDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(llmscriptDir, "config.yaml")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
