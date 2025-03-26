package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		cfg := DefaultConfig()
		if cfg.MaxFixes != 10 {
			t.Errorf("expected MaxFixes=10, got %d", cfg.MaxFixes)
		}
		if cfg.MaxAttempts != 3 {
			t.Errorf("expected MaxAttempts=3, got %d", cfg.MaxAttempts)
		}
		if cfg.Timeout != 30*time.Second {
			t.Errorf("expected Timeout=30s, got %v", cfg.Timeout)
		}
	})

	t.Run("load and write config", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
			t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
				t.Errorf("failed to unset XDG_CONFIG_HOME: %v", err)
			}
		}()

		cfg := DefaultConfig()
		cfg.LLM.Provider = "ollama"
		cfg.LLM.Ollama.Model = "llama2"
		cfg.LLM.Ollama.Host = "http://localhost:11434"

		if err := WriteConfig(cfg); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		loaded, err := LoadConfig()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if loaded.LLM.Provider != cfg.LLM.Provider {
			t.Errorf("expected provider=%s, got %s", cfg.LLM.Provider, loaded.LLM.Provider)
		}
		if loaded.LLM.Ollama.Model != cfg.LLM.Ollama.Model {
			t.Errorf("expected model=%s, got %s", cfg.LLM.Ollama.Model, loaded.LLM.Ollama.Model)
		}
		if loaded.LLM.Ollama.Host != cfg.LLM.Ollama.Host {
			t.Errorf("expected host=%s, got %s", cfg.LLM.Ollama.Host, loaded.LLM.Ollama.Host)
		}
	})

	t.Run("load non-existent config", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
			t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
				t.Errorf("failed to unset XDG_CONFIG_HOME: %v", err)
			}
		}()

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if cfg.MaxFixes != 10 {
			t.Errorf("expected MaxFixes=10, got %d", cfg.MaxFixes)
		}
	})

	t.Run("load yaml file with env vars", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
			t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
				t.Errorf("failed to unset XDG_CONFIG_HOME: %v", err)
			}
		}()

		// Set up test environment variables
		if err := os.Setenv("TEST_API_KEY", "test-key-123"); err != nil {
			t.Fatalf("failed to set TEST_API_KEY: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("TEST_API_KEY"); err != nil {
				t.Errorf("failed to unset TEST_API_KEY: %v", err)
			}
		}()

		// Create test config directory
		configDir := filepath.Join(tmpDir, "llmscript")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("failed to create config directory: %v", err)
		}

		// Create a test YAML file
		yamlContent := []byte(`llm:
  provider: "claude"
  claude:
    api_key: "${TEST_API_KEY}"
    model: "claude-3-opus-20240229"
max_fixes: 5
max_attempts: 2
timeout: 15s
additional_prompt: "Test prompt"
`)
		configPath := filepath.Join(configDir, "config.yaml")
		if err := os.WriteFile(configPath, yamlContent, 0644); err != nil {
			t.Fatalf("failed to write test config: %v", err)
		}

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if cfg.LLM.Provider != "claude" {
			t.Errorf("expected provider=claude, got %s", cfg.LLM.Provider)
		}
		if cfg.LLM.Claude.APIKey != "test-key-123" {
			t.Errorf("expected api_key=test-key-123, got %s", cfg.LLM.Claude.APIKey)
		}
		if cfg.LLM.Claude.Model != "claude-3-opus-20240229" {
			t.Errorf("expected model=claude-3-opus-20240229, got %s", cfg.LLM.Claude.Model)
		}
		if cfg.MaxFixes != 5 {
			t.Errorf("expected MaxFixes=5, got %d", cfg.MaxFixes)
		}
		if cfg.MaxAttempts != 2 {
			t.Errorf("expected MaxAttempts=2, got %d", cfg.MaxAttempts)
		}
		if cfg.Timeout != 15*time.Second {
			t.Errorf("expected Timeout=15s, got %v", cfg.Timeout)
		}
		if cfg.ExtraPrompt != "Test prompt" {
			t.Errorf("expected ExtraPrompt=Test prompt, got %s", cfg.ExtraPrompt)
		}
	})

	t.Run("config snapshot", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.Setenv("XDG_CONFIG_HOME", tmpDir); err != nil {
			t.Fatalf("failed to set XDG_CONFIG_HOME: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("XDG_CONFIG_HOME"); err != nil {
				t.Errorf("failed to unset XDG_CONFIG_HOME: %v", err)
			}
		}()

		// Create test config directory
		configDir := filepath.Join(tmpDir, "llmscript")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("failed to create config directory: %v", err)
		}

		cfg := DefaultConfig()
		cfg.LLM.Provider = "ollama"
		cfg.LLM.Ollama.Model = "llama2"
		cfg.LLM.Ollama.Host = "http://localhost:11434"
		cfg.LLM.Claude.APIKey = ""
		cfg.LLM.Claude.Model = ""
		cfg.LLM.OpenAI.APIKey = ""
		cfg.LLM.OpenAI.Model = ""
		cfg.MaxFixes = 5
		cfg.MaxAttempts = 2
		cfg.Timeout = 15 * time.Second
		cfg.ExtraPrompt = "Test prompt"

		if err := WriteConfig(cfg); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		// Read the written file
		configPath := filepath.Join(configDir, "config.yaml")
		written, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read written config: %v", err)
		}

		// Compare with expected snapshot
		expected := `llm:
  provider: ollama
  ollama:
    model: llama2
    host: http://localhost:11434
  claude:
    api_key: ""
    model: ""
  openai:
    api_key: ""
    model: ""
max_fixes: 5
max_attempts: 2
timeout: 15s
additional_prompt: Test prompt
`
		if string(written) != expected {
			t.Errorf("config snapshot mismatch:\nExpected:\n%s\nGot:\n%s", expected, string(written))
		}
	})
}
