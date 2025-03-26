package llm

// OllamaConfig represents configuration for the Ollama provider
type OllamaConfig struct {
	Model string `yaml:"model"`
	Host  string `yaml:"host"`
}

// ClaudeConfig represents configuration for the Claude provider
type ClaudeConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

// OpenAIConfig represents configuration for the OpenAI provider
type OpenAIConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

// ProviderConfig represents the configuration for any LLM provider
type ProviderConfig struct {
	Provider string       `yaml:"provider"`
	Ollama   OllamaConfig `yaml:"ollama,omitempty"`
	Claude   ClaudeConfig `yaml:"claude,omitempty"`
	OpenAI   OpenAIConfig `yaml:"openai,omitempty"`
}
