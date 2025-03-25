package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// GenerateScript creates a shell script from a natural language description
func (p *OllamaProvider) GenerateScript(ctx context.Context, description string) (string, error) {
	prompt := p.formatPrompt(generateScriptPrompt, description)
	response, err := p.generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate script: %w", err)
	}
	return response, nil
}

// GenerateTests creates test cases for a script based on its description
func (p *OllamaProvider) GenerateTests(ctx context.Context, description string) ([]Test, error) {
	prompt := p.formatPrompt(generateTestsPrompt, "", description)
	response, err := p.generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tests: %w", err)
	}

	var tests []Test
	if err := json.Unmarshal([]byte(response), &tests); err != nil {
		return nil, fmt.Errorf("failed to parse test cases: %w", err)
	}
	return tests, nil
}

// FixScript attempts to fix a script based on test failures
func (p *OllamaProvider) FixScript(ctx context.Context, script string, failures []TestFailure) (string, error) {
	failuresStr := formatFailures(failures)
	prompt := p.formatPrompt(fixScriptPrompt, script, failuresStr)
	response, err := p.generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to fix script: %w", err)
	}
	return response, nil
}

// generate sends a prompt to Ollama and returns the response
func (p *OllamaProvider) generate(ctx context.Context, prompt string) (string, error) {
	host := p.config.Host
	if host == "" {
		host = "http://localhost:11434"
	}
	url := fmt.Sprintf("%s/api/generate", host)

	model := p.config.Model
	if model == "" {
		model = "llama3.2"
	}

	reqBody := map[string]interface{}{
		"model":  model,
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
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Response, nil
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
