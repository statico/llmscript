package llm

import "context"

// Test represents a test case for a script
type Test struct {
	Name        string
	Description string
	Setup       []string
	Input       string
	Expected    string
}

// TestFailure represents a failed test case
type TestFailure struct {
	Test     Test
	Actual   string
	Error    error
	ExitCode int
	Stdout   string
	Stderr   string
}

// Provider defines the interface for LLM providers
type Provider interface {
	// GenerateScript creates a shell script from a natural language description
	GenerateScript(ctx context.Context, description string) (string, error)

	// GenerateTests creates test cases for the script
	GenerateTests(ctx context.Context, description string) ([]Test, error)

	// FixScript attempts to fix a script based on test failures
	FixScript(ctx context.Context, script string, failures []TestFailure) (string, error)
}

// NewProvider creates a new LLM provider based on the provider name
func NewProvider(name string, config interface{}) (Provider, error) {
	// TODO: Implement provider factory
	return nil, nil
}
