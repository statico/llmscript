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
	// First generate the main script
	log.Info("Generating main script with Ollama...")
	mainPrompt := p.formatPrompt(FeatureScriptPrompt, description)
	mainScript, err := p.generate(ctx, mainPrompt)
	if err != nil {
		return ScriptPair{}, fmt.Errorf("failed to generate main script: %w", err)
	}
	mainScript = ExtractScriptContent(mainScript)
	log.Debug("Main script generated:\n%s", mainScript)

	// Then generate the test script
	log.Info("Generating test script with Ollama...")
	testPrompt := p.formatPrompt(TestScriptPrompt, mainScript, description)
	testScript, err := p.generate(ctx, testPrompt)
	if err != nil {
		return ScriptPair{}, fmt.Errorf("failed to generate test script: %w", err)
	}
	testScript = ExtractScriptContent(testScript)
	log.Debug("Test script generated:\n%s", testScript)

	return ScriptPair{
		MainScript: strings.TrimSpace(mainScript),
		TestScript: strings.TrimSpace(testScript),
	}, nil
}

// FixScripts attempts to fix both scripts based on test failures
func (p *OllamaProvider) FixScripts(ctx context.Context, scripts ScriptPair, error string) (ScriptPair, error) {
	// Only fix the main script
	mainPrompt := p.formatPrompt(FixScriptPrompt, scripts.MainScript, error)
	fixedMainScript, err := p.generate(ctx, mainPrompt)
	if err != nil {
		return ScriptPair{}, fmt.Errorf("failed to fix main script: %w", err)
	}
	fixedMainScript = ExtractScriptContent(fixedMainScript)

	return ScriptPair{
		MainScript: strings.TrimSpace(fixedMainScript),
		TestScript: scripts.TestScript,
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read the entire response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the response
	var result struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Response, nil
}

// Name returns a human-readable name for the provider
func (p *OllamaProvider) Name() string {
	return "Ollama"
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
