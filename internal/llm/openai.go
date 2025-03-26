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

// GenerateScripts creates a main script and test script from a natural language description
func (p *OpenAIProvider) GenerateScripts(ctx context.Context, description string) (ScriptPair, error) {
	// First generate the main script
	mainPrompt := p.formatPrompt(FeatureScriptPrompt, description)
	mainScript, err := p.generate(ctx, mainPrompt)
	if err != nil {
		return ScriptPair{}, fmt.Errorf("failed to generate main script: %w", err)
	}
	log.Debug("Main script generated:\n%s", mainScript)

	// Then generate the test script
	testPrompt := p.formatPrompt(TestScriptPrompt, mainScript, description)
	testScript, err := p.generate(ctx, testPrompt)
	if err != nil {
		return ScriptPair{}, fmt.Errorf("failed to generate test script: %w", err)
	}
	log.Debug("Test script generated:\n%s", testScript)

	return ScriptPair{
		MainScript: strings.TrimSpace(mainScript),
		TestScript: strings.TrimSpace(testScript),
	}, nil
}

// FixScripts attempts to fix both scripts based on test failures
func (p *OpenAIProvider) FixScripts(ctx context.Context, scripts ScriptPair, error string) (ScriptPair, error) {
	// Only fix the main script
	mainPrompt := p.formatPrompt(FixScriptPrompt, scripts.MainScript, error)
	fixedMainScript, err := p.generate(ctx, mainPrompt)
	if err != nil {
		return ScriptPair{}, fmt.Errorf("failed to fix main script: %w", err)
	}

	return ScriptPair{
		MainScript: strings.TrimSpace(fixedMainScript),
		TestScript: scripts.TestScript,
	}, nil
}

// GenerateScript creates a shell script from a natural language description
func (p *OpenAIProvider) GenerateScript(ctx context.Context, description string) (string, error) {
	prompt := p.formatPrompt(FeatureScriptPrompt, description)
	response, err := p.generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate script: %w", err)
	}
	return response, nil
}

// GenerateTests creates test cases for a script based on its description
func (p *OpenAIProvider) GenerateTests(ctx context.Context, script string, description string) ([]Test, error) {
	prompt := p.formatPrompt(TestScriptPrompt, script, description)
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

	var result struct {
		Tests []Test `json:"tests"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse test cases: %w\nRaw response:\n%s", err, response)
	}
	return result.Tests, nil
}

// FixScript attempts to fix a script based on test failures
func (p *OpenAIProvider) FixScript(ctx context.Context, script string, failures []TestFailure) (string, error) {
	failuresStr := formatFailures(failures)
	prompt := p.formatPrompt(FixScriptPrompt, script, failuresStr)
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error("failed to close response body: %v", err)
		}
	}()

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
