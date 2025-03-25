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

// ClaudeProvider implements the Provider interface using Anthropic's Claude
type ClaudeProvider struct {
	BaseProvider
	config ClaudeConfig
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(config ClaudeConfig) (*ClaudeProvider, error) {
	return &ClaudeProvider{
		BaseProvider: BaseProvider{Config: config},
		config:       config,
	}, nil
}

// GenerateScripts creates a shell script and its test script from a natural language description
func (p *ClaudeProvider) GenerateScripts(ctx context.Context, description string) (ScriptPair, error) {
	prompt := p.formatPrompt(generateScriptsPrompt, description)
	response, err := p.generate(ctx, prompt)
	if err != nil {
		return ScriptPair{}, fmt.Errorf("failed to generate scripts: %w", err)
	}

	// Split response into main script and test script
	parts := strings.Split(response, "\n---\n")
	if len(parts) != 2 {
		return ScriptPair{}, fmt.Errorf("expected two scripts separated by '---', got %d parts", len(parts))
	}

	return ScriptPair{
		MainScript: strings.TrimSpace(parts[0]),
		TestScript: strings.TrimSpace(parts[1]),
	}, nil
}

// FixScripts attempts to fix both scripts based on test failures
func (p *ClaudeProvider) FixScripts(ctx context.Context, scripts ScriptPair, error string) (ScriptPair, error) {
	prompt := p.formatPrompt(fixScriptsPrompt, scripts.MainScript, scripts.TestScript, error)
	response, err := p.generate(ctx, prompt)
	if err != nil {
		return ScriptPair{}, fmt.Errorf("failed to fix scripts: %w", err)
	}

	// Split response into main script and test script
	parts := strings.Split(response, "\n---\n")
	if len(parts) != 2 {
		return ScriptPair{}, fmt.Errorf("expected two scripts separated by '---', got %d parts", len(parts))
	}

	return ScriptPair{
		MainScript: strings.TrimSpace(parts[0]),
		TestScript: strings.TrimSpace(parts[1]),
	}, nil
}

// GenerateScript creates a shell script from a natural language description
func (p *ClaudeProvider) GenerateScript(ctx context.Context, description string) (string, error) {
	prompt := p.formatPrompt(generateScriptPrompt, description)
	response, err := p.generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate script: %w", err)
	}
	return response, nil
}

// GenerateTests creates test cases for a script based on its description
func (p *ClaudeProvider) GenerateTests(ctx context.Context, description string) ([]Test, error) {
	prompt := p.formatPrompt(generateTestsPrompt, "", description)
	response, err := p.generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tests: %w", err)
	}

	log.Debug("Raw LLM response:\n%s", response)

	// Try to extract JSON from the response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return nil, fmt.Errorf("failed to find valid JSON in response: %s", response)
	}
	jsonStr := response[jsonStart : jsonEnd+1]

	var tests []Test
	if err := json.Unmarshal([]byte(jsonStr), &tests); err != nil {
		return nil, fmt.Errorf("failed to parse test cases: %w\nRaw response:\n%s", err, response)
	}
	return tests, nil
}

// FixScript attempts to fix a script based on test failures
func (p *ClaudeProvider) FixScript(ctx context.Context, script string, failures []TestFailure) (string, error) {
	failuresStr := formatFailures(failures)
	prompt := p.formatPrompt(fixScriptPrompt, script, failuresStr)
	response, err := p.generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to fix script: %w", err)
	}
	return response, nil
}

// generate sends a prompt to Claude and returns the response
func (p *ClaudeProvider) generate(ctx context.Context, prompt string) (string, error) {
	url := "https://api.anthropic.com/v1/messages"

	reqBody := map[string]interface{}{
		"model":    p.config.Model,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
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
	req.Header.Set("x-api-key", p.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

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
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return result.Content[0].Text, nil
}
