package llm

import (
	"context"
	"fmt"
)

// ScriptPair represents a feature script and its test script
type ScriptPair struct {
	MainScript string // The feature script that implements the functionality
	TestScript string // The test script that verifies the feature script
}

// BaseProvider provides common functionality for LLM providers
type BaseProvider struct {
	Config interface{}
}

// GenerateScripts generates a main script and test script pair
func (p *BaseProvider) GenerateScripts(ctx context.Context, description string) (ScriptPair, error) {
	return ScriptPair{}, fmt.Errorf("GenerateScripts not implemented")
}

// FixScripts attempts to fix a script pair based on test failures
func (p *BaseProvider) FixScripts(ctx context.Context, scripts ScriptPair, error string) (ScriptPair, error) {
	return ScriptPair{}, fmt.Errorf("FixScripts not implemented")
}

// GenerateScript is a default implementation that returns an error
func (p *BaseProvider) GenerateScript(ctx context.Context, description string) (string, error) {
	return "", fmt.Errorf("GenerateScript not implemented")
}

// GenerateTests is a default implementation that returns an error
func (p *BaseProvider) GenerateTests(ctx context.Context, script string, description string) ([]Test, error) {
	return nil, fmt.Errorf("GenerateTests not implemented")
}

// FixScript is a default implementation that returns an error
func (p *BaseProvider) FixScript(ctx context.Context, script string, failures []TestFailure) (string, error) {
	return "", fmt.Errorf("FixScript not implemented")
}

// ValidateConfig validates the provider configuration
func (p *BaseProvider) ValidateConfig() error {
	if p.Config == nil {
		return fmt.Errorf("config is required")
	}
	return nil
}

// formatPrompt formats a prompt template with the given arguments and platform information
func (p *BaseProvider) formatPrompt(template string, args ...interface{}) string {
	// Add platform information as the last argument
	args = append(args, GetPlatformInfo())
	return fmt.Sprintf(template, args...)
}
