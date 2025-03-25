package llm

import (
	"testing"
)

func TestOllamaConfig(t *testing.T) {
	config := OllamaConfig{
		Model: "llama2",
		Host:  "http://localhost:11434",
	}

	if config.Model != "llama2" {
		t.Errorf("Expected Model to be 'llama2', got %q", config.Model)
	}
	if config.Host != "http://localhost:11434" {
		t.Errorf("Expected Host to be 'http://localhost:11434', got %q", config.Host)
	}
}

func TestClaudeConfig(t *testing.T) {
	config := ClaudeConfig{
		APIKey: "test-key",
		Model:  "claude-3-opus-20240229",
	}

	if config.APIKey != "test-key" {
		t.Errorf("Expected APIKey to be 'test-key', got %q", config.APIKey)
	}
	if config.Model != "claude-3-opus-20240229" {
		t.Errorf("Expected Model to be 'claude-3-opus-20240229', got %q", config.Model)
	}
}

func TestOpenAIConfig(t *testing.T) {
	config := OpenAIConfig{
		APIKey: "test-key",
		Model:  "gpt-4-turbo-preview",
	}

	if config.APIKey != "test-key" {
		t.Errorf("Expected APIKey to be 'test-key', got %q", config.APIKey)
	}
	if config.Model != "gpt-4-turbo-preview" {
		t.Errorf("Expected Model to be 'gpt-4-turbo-preview', got %q", config.Model)
	}
}

func TestProviderConfig(t *testing.T) {
	config := ProviderConfig{
		Provider: "ollama",
		Ollama: OllamaConfig{
			Model: "llama2",
			Host:  "http://localhost:11434",
		},
	}

	if config.Provider != "ollama" {
		t.Errorf("Expected Provider to be 'ollama', got %q", config.Provider)
	}
	if config.Ollama.Model != "llama2" {
		t.Errorf("Expected Ollama.Model to be 'llama2', got %q", config.Ollama.Model)
	}
	if config.Ollama.Host != "http://localhost:11434" {
		t.Errorf("Expected Ollama.Host to be 'http://localhost:11434', got %q", config.Ollama.Host)
	}
}
