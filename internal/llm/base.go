package llm

import (
	"context"
	"fmt"
)

// BaseProvider provides common functionality for LLM providers
type BaseProvider struct {
	Config interface{}
}

// GenerateScript is a default implementation that returns an error
func (p *BaseProvider) GenerateScript(ctx context.Context, description string) (string, error) {
	return "", fmt.Errorf("GenerateScript not implemented")
}

// GenerateTests is a default implementation that returns an error
func (p *BaseProvider) GenerateTests(ctx context.Context, description string) ([]Test, error) {
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
