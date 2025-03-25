package llm

import (
	"context"
	"time"
)

// Test represents a test case for a script
type Test struct {
	Name        string
	Setup       []string
	Input       string
	Expected    string
	Timeout     time.Duration
	Environment map[string]string
}

// TestFailure represents a failed test case
type TestFailure struct {
	Test     Test
	Output   string
	Error    error
	ExitCode int
}

// Provider defines the interface for LLM providers
type Provider interface {
	// GenerateScript creates a shell script from a natural language description
	GenerateScript(ctx context.Context, description string) (string, error)

	// GenerateTests creates test cases for a script based on its description
	GenerateTests(ctx context.Context, description string) ([]Test, error)

	// FixScript attempts to fix a script based on test failures
	FixScript(ctx context.Context, script string, failures []TestFailure) (string, error)
}

// Prompt templates for different operations
const (
	generateScriptPrompt = `You are an expert shell script developer. Create a shell script that accomplishes the following task:

Task:
%s

Requirements:
1. Use standard shell commands (sh/bash)
2. Handle errors appropriately
3. Use clear variable names
4. Add helpful comments
5. Follow shell scripting best practices

Output only the shell script, nothing else.`

	generateTestsPrompt = `You are an expert in testing shell scripts. Create test cases for the following script:

Script:
%s

Description:
%s

Requirements:
1. Test both success and failure cases
2. Include setup steps if needed
3. Test edge cases
4. Verify output matches expectations
5. Include timeout values
6. Specify any required environment variables

Output the test cases in JSON format.`

	fixScriptPrompt = `You are an expert shell script developer. Fix the following script based on the test failures:

Script:
%s

Test Failures:
%s

Requirements:
1. Fix all test failures
2. Maintain existing functionality
3. Keep the code clean and readable
4. Add error handling if missing
5. Follow shell scripting best practices

Output only the fixed shell script, nothing else.`
)

// NewProvider creates a new LLM provider based on the provider name
func NewProvider(name string, config interface{}) (Provider, error) {
	// TODO: Implement provider factory
	return nil, nil
}
