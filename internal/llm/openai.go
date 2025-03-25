package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAIProvider implements the Provider interface using OpenAI's API
type OpenAIProvider struct {
	BaseProvider
	config OpenAIConfig
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config OpenAIConfig) (*OpenAIProvider, error) {
	return &OpenAIProvider{
		BaseProvider: BaseProvider{Config: config},
		config:       config,
	}, nil
}

// GenerateScript creates a shell script from a natural language description
func (p *OpenAIProvider) GenerateScript(ctx context.Context, description string) (string, error) {
	prompt := p.formatPrompt(generateScriptPrompt, description)
	response, err := p.generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate script: %w", err)
	}
	return response, nil
}

// GenerateTests creates test cases for a script based on its description
func (p *OpenAIProvider) GenerateTests(ctx context.Context, description string) ([]Test, error) {
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
func (p *OpenAIProvider) FixScript(ctx context.Context, script string, failures []TestFailure) (string, error) {
	failuresStr := formatFailures(failures)
	prompt := p.formatPrompt(fixScriptPrompt, script, failuresStr)
	response, err := p.generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to fix script: %w", err)
	}
	return response, nil
}

// generate sends a prompt to OpenAI and returns the response
func (p *OpenAIProvider) generate(ctx context.Context, prompt string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	reqBody := map[string]interface{}{
		"model": p.config.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
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
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

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
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return result.Choices[0].Message.Content, nil
}
