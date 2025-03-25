package config

import (
	"os"
	"path/filepath"

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
	MaxFixes         int    `yaml:"max_fixes"`
	MaxAttempts      int    `yaml:"max_attempts"`
	Timeout          int    `yaml:"timeout"`
	AdditionalPrompt string `yaml:"additional_prompt"`
}

func Load(configPath string) (*Config, error) {
	cfg := &Config{
		MaxFixes:    10,
		MaxAttempts: 3,
		Timeout:     30,
	}

	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		configPath = filepath.Join(home, ".config", "llmscript", "config.yaml")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func WriteDefaultConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ".config", "llmscript")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	cfg := &Config{
		MaxFixes:    10,
		MaxAttempts: 3,
		Timeout:     30,
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(configDir, "config.yaml"), data, 0644)
}
