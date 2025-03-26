package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	customlog "github.com/statico/llmscript/internal/log"
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

	// Try XDG_CONFIG_HOME first
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		// If XDG_CONFIG_HOME is not set, try ~/.config
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, ".config")
	}

	configPath := filepath.Join(configDir, "llmscript", "config.yaml")
	customlog.Debug("Looking for config file at: %s", configPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If not found in ~/.config, try UserConfigDir() as fallback
			configDir, err = os.UserConfigDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get config dir: %w", err)
			}
			configPath = filepath.Join(configDir, "llmscript", "config.yaml")
			customlog.Debug("Looking for config file at fallback location: %s", configPath)
			data, err = os.ReadFile(configPath)
			if err != nil {
				if os.IsNotExist(err) {
					customlog.Debug("Config file not found, using defaults")
					return config, nil
				}
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	customlog.Debug("Found config file: %s", configPath)

	// Interpolate environment variables before unmarshaling
	data = interpolateEnvVars(data)

	// Create a temporary config to unmarshal into
	var loadedConfig Config
	if err := yaml.Unmarshal(data, &loadedConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	customlog.Debug("Loaded config from file: provider=%s", loadedConfig.LLM.Provider)

	// Merge the loaded config with the default config
	config.LLM.Provider = loadedConfig.LLM.Provider
	config.LLM.Ollama.Model = loadedConfig.LLM.Ollama.Model
	config.LLM.Ollama.Host = loadedConfig.LLM.Ollama.Host
	config.LLM.Claude.APIKey = loadedConfig.LLM.Claude.APIKey
	config.LLM.Claude.Model = loadedConfig.LLM.Claude.Model
	config.LLM.OpenAI.APIKey = loadedConfig.LLM.OpenAI.APIKey
	config.LLM.OpenAI.Model = loadedConfig.LLM.OpenAI.Model

	customlog.Debug("Merged config: provider=%s", config.LLM.Provider)

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
