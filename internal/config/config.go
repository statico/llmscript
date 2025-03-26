package config

import (
	"fmt"
	"log"
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
		LLM: struct {
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
		}{
			Provider: "ollama",
			Ollama: struct {
				Model string `yaml:"model"`
				Host  string `yaml:"host"`
			}{
				Model: "llama3.2",
				Host:  "http://localhost:11434",
			},
			Claude: struct {
				APIKey string `yaml:"api_key"`
				Model  string `yaml:"model"`
			}{
				APIKey: "${CLAUDE_API_KEY}",
				Model:  "claude-3-opus-20240229",
			},
			OpenAI: struct {
				APIKey string `yaml:"api_key"`
				Model  string `yaml:"model"`
			}{
				APIKey: "${OPENAI_API_KEY}",
				Model:  "gpt-4-turbo-preview",
			},
		},
		MaxFixes:    10,
		MaxAttempts: 3,
		Timeout:     30 * time.Second,
		ExtraPrompt: "Use ANSI color codes to make the output more readable.",
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
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	fmt.Printf("Config directory: %s\n", configDir)

	llmscriptDir := filepath.Join(configDir, "llmscript")
	fmt.Printf("Creating directory: %s\n", llmscriptDir)
	if err := os.MkdirAll(llmscriptDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(llmscriptDir, "config.yaml")
	fmt.Printf("Writing config to: %s\n", configPath)
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
