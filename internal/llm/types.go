package llm

// OllamaConfig represents configuration for the Ollama provider
type OllamaConfig struct {
	Model string `yaml:"model"`
	Host  string `yaml:"host"`
}

// ClaudeConfig represents configuration for the Claude (Anthropic) provider
type ClaudeConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

// OpenAIConfig represents configuration for the OpenAI provider
type OpenAIConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

// GeminiConfig represents configuration for the Google Gemini provider
type GeminiConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

// OpenRouterConfig represents configuration for the OpenRouter provider
type OpenRouterConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

// Config is the fully-resolved LLM configuration passed to NewProvider. It is
// built from the user's config file, environment, and command-line flags.
type Config struct {
	Provider    string
	ExtraPrompt string
	Ollama      OllamaConfig
	Claude      ClaudeConfig
	OpenAI      OpenAIConfig
	Gemini      GeminiConfig
	OpenRouter  OpenRouterConfig
}
