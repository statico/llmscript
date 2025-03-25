package llm

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
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

// GetPlatformInfo returns information about the current platform
func GetPlatformInfo() string {
	info := []string{
		"Operating System: " + runtime.GOOS,
		"Architecture: " + runtime.GOARCH,
	}

	// Get additional system information using uname
	cmd := exec.Command("uname", "-a")
	if output, err := cmd.Output(); err == nil {
		info = append(info, "System Info: "+string(output))
	}

	// Get shell information
	cmd = exec.Command("bash", "--version")
	if output, err := cmd.Output(); err == nil {
		info = append(info, "Shell Info: "+string(output))
	}

	return strings.Join(info, "\n")
}

// Prompt templates for different operations
const (
	generateScriptPrompt = `You are an expert shell script developer with deep knowledge of Unix/Linux systems, shell scripting best practices, and error handling.
Your task is to create robust, maintainable shell scripts that work reliably across different environments.

Create a shell script that accomplishes the following task:

<description>
%s
</description>

<platform>
Platform Information:
%s
</platform>

<requirements>
1. Use standard shell commands (sh/bash) with POSIX compliance where possible
2. Implement comprehensive error handling with descriptive messages
3. Use clear, descriptive variable names following shell naming conventions
4. Add concise, meaningful comments for complex logic
5. Follow shell scripting best practices and security guidelines
6. Ensure cross-platform compatibility
7. Include input validation where appropriate
8. Use proper exit codes for different scenarios
</requirements>

<output_format>
Output only the shell script, nothing else. The script should be ready to execute.
</output_format>`

	generateTestsPrompt = `You are an expert in testing shell scripts with extensive experience in test automation and quality assurance.
Your goal is to create comprehensive test cases that verify script functionality, edge cases, and error conditions.

Create test cases for the following script:

<script>
%s
</script>

<description>
%s
</description>

<platform>
Platform Information:
%s
</platform>

<requirements>
1. Test both success and failure scenarios
2. Include necessary setup and teardown steps
3. Cover edge cases and boundary conditions
4. Verify exact output matching
5. Set appropriate timeout values
6. Define required environment variables
7. Ensure platform compatibility
8. Include cleanup steps to restore system state
</requirements>

<output_format>
Output the test cases in JSON format with the following structure:
{
  "tests": [
    {
      "name": "string",
      "setup": ["string"],
      "input": "string",
      "expected": "string",
      "timeout": "duration",
      "environment": {"key": "value"}
    }
  ]
}
</output_format>`

	fixScriptPrompt = `You are an expert shell script developer specializing in debugging and fixing shell scripts.
Your expertise includes error handling, cross-platform compatibility, and shell scripting best practices.

Fix the following script based on the test failures:

<script>
%s
</script>

<test_failures>
%s
</test_failures>

<platform>
Platform Information:
%s
</platform>

<requirements>
1. Fix all test failures while maintaining existing functionality
2. Improve error handling and validation
3. Enhance code readability and maintainability
4. Follow shell scripting best practices
5. Ensure cross-platform compatibility
6. Add appropriate logging for debugging
7. Implement proper cleanup in error cases
8. Optimize performance where possible
</requirements>

<output_format>
Output only the fixed shell script, nothing else. The script should pass all tests and maintain its original functionality.
</output_format>`
)

// NewProvider creates a new LLM provider based on the provider name
func NewProvider(name string, config interface{}) (Provider, error) {
	switch name {
	case "ollama":
		cfg, ok := config.(struct {
			Provider string `yaml:"provider"`
			Ollama   struct {
				Model string `yaml:"model"`
				Host  string `yaml:"host"`
			} `yaml:"ollama"`
		})
		if !ok {
			return nil, fmt.Errorf("invalid config type for Ollama provider")
		}
		return NewOllamaProvider(OllamaConfig{
			Model: cfg.Ollama.Model,
			Host:  cfg.Ollama.Host,
		})
	case "claude":
		cfg, ok := config.(struct {
			Provider string `yaml:"provider"`
			Claude   struct {
				APIKey string `yaml:"api_key"`
				Model  string `yaml:"model"`
			} `yaml:"claude"`
		})
		if !ok {
			return nil, fmt.Errorf("invalid config type for Claude provider")
		}
		return NewClaudeProvider(ClaudeConfig{
			APIKey: cfg.Claude.APIKey,
			Model:  cfg.Claude.Model,
		})
	case "openai":
		cfg, ok := config.(struct {
			Provider string `yaml:"provider"`
			OpenAI   struct {
				APIKey string `yaml:"api_key"`
				Model  string `yaml:"model"`
			} `yaml:"openai"`
		})
		if !ok {
			return nil, fmt.Errorf("invalid config type for OpenAI provider")
		}
		return NewOpenAIProvider(OpenAIConfig{
			APIKey: cfg.OpenAI.APIKey,
			Model:  cfg.OpenAI.Model,
		})
	default:
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
}
