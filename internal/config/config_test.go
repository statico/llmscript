package config

import (
	"os"
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
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		defer os.Unsetenv("XDG_CONFIG_HOME")

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
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		defer os.Unsetenv("XDG_CONFIG_HOME")

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if cfg.MaxFixes != 10 {
			t.Errorf("expected MaxFixes=10, got %d", cfg.MaxFixes)
		}
	})
}
