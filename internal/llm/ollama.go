package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/statico/llmscript/internal/log"
)

// OllamaProvider implements the Provider interface using Ollama
type OllamaProvider struct {
	BaseProvider
	config OllamaConfig
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(config OllamaConfig) (*OllamaProvider, error) {
	return &OllamaProvider{
		BaseProvider: BaseProvider{Config: config},
		config:       config,
	}, nil
}

// GenerateScripts creates a main script and test script from a natural language description
func (p *OllamaProvider) GenerateScripts(ctx context.Context, description string) (ScriptPair, error) {
	prompt := p.formatPrompt(generateScriptsPrompt, description)
	response, err := p.generate(ctx, prompt)
	if err != nil {
		return ScriptPair{}, fmt.Errorf("failed to generate scripts: %w", err)
	}

	// Split response into main script and test script
	parts := strings.Split(response, "---")
	if len(parts) != 2 {
		return ScriptPair{}, fmt.Errorf("invalid response format: expected two scripts separated by '---'")
	}

	return ScriptPair{
		MainScript: strings.TrimSpace(parts[0]),
		TestScript: strings.TrimSpace(parts[1]),
	}, nil
}

// FixScripts attempts to fix both scripts based on test failures
func (p *OllamaProvider) FixScripts(ctx context.Context, scripts ScriptPair, error string) (ScriptPair, error) {
	prompt := p.formatPrompt(fixScriptsPrompt, scripts.MainScript, scripts.TestScript, error)
	response, err := p.generate(ctx, prompt)
	if err != nil {
		return ScriptPair{}, fmt.Errorf("failed to fix scripts: %w", err)
	}

	// Split response into main script and test script
	parts := strings.Split(response, "---")
	if len(parts) != 2 {
		return ScriptPair{}, fmt.Errorf("invalid response format: expected two scripts separated by '---'")
	}

	return ScriptPair{
		MainScript: strings.TrimSpace(parts[0]),
		TestScript: strings.TrimSpace(parts[1]),
	}, nil
}

// generate sends a prompt to Ollama and returns the response
func (p *OllamaProvider) generate(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("%s/api/generate", p.config.Host)

	reqBody := map[string]interface{}{
		"model":  p.config.Model,
		"prompt": prompt,
		"stream": false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		if err := decoder.Decode(&result); err != nil {
			return "", fmt.Errorf("failed to decode response: %w", err)
		}
		fullResponse.WriteString(result.Response)
	}

	response := fullResponse.String()
	log.Debug("Raw Ollama response:\n%s", response)
	return response, nil
}

// formatFailures formats test failures into a string
func formatFailures(failures []TestFailure) string {
	var parts []string
	for _, f := range failures {
		parts = append(parts, fmt.Sprintf("Test: %s\nOutput: %s\nError: %v\nExit Code: %d",
			f.Test.Name, f.Output, f.Error, f.ExitCode))
	}
	return strings.Join(parts, "\n\n")
}
