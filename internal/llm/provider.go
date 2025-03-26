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
	// GenerateScripts creates a main script and test script from a natural language description
	GenerateScripts(ctx context.Context, description string) (ScriptPair, error)
	// FixScripts attempts to fix both scripts based on test failures
	FixScripts(ctx context.Context, scripts ScriptPair, error string) (ScriptPair, error)
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
	generateFeatureScriptPrompt = `You are an expert shell script developer with deep knowledge of Unix/Linux systems, shell scripting best practices, and error handling.
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
Output your response in the following format:

<script>
#!/usr/bin/env bash
# Your shell script content here
</script>

You *MUST NOT* include any other text, explanations, or markdown formatting.
</output_format>`

	generateTestScriptPrompt = `You are an expert in testing shell scripts with extensive experience in test automation and quality assurance.
Your goal is to create a comprehensive test script that verifies the functionality of the main script, including edge cases and error conditions.

Create a test script for the following script:

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
1. Create a test script that runs multiple test cases
2. Each test case should:
   - Set up the test environment
   - Run the main script with test inputs
   - Verify the output matches expectations
   - Clean up after the test
3. Include tests for:
   - Success scenarios
   - Error cases
   - Edge cases
   - Boundary conditions
4. Use clear test case names and descriptions
5. Implement proper error handling and reporting
6. Set appropriate timeouts for long-running tests
7. Handle environment variables and cleanup
8. Ensure platform compatibility
</requirements>

<output_format>
Output your response in the following format:

<script>
#!/usr/bin/env bash
# Your test script content here
</script>

You *MUST NOT* include any other text, explanations, or markdown formatting.
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
Output your response in the following format:

<script>
#!/usr/bin/env bash
# Your fixed shell script content here
</script>

Do not include any other text, explanations, or markdown formatting. Only output the script between the markers.
</output_format>`
)

// NewProvider creates a new LLM provider based on the provider name
func NewProvider(name string, config interface{}) (Provider, error) {
	if name == "" {
		name = "ollama" // Default to Ollama if no provider specified
	}

	// Default Ollama config
	ollamaConfig := OllamaConfig{
		Model: "llama3.2",
		Host:  "http://localhost:11434",
	}

	// If config is provided, try to extract values
	if config != nil {
		cfg, ok := config.(struct {
			Provider string `yaml:"provider"`
			Ollama   struct {
				Model string `yaml:"model"`
				Host  string `yaml:"host"`
			} `yaml:"ollama"`
			Claude struct {
				APIKey string `yaml:"api_key"`
				Model  string `yaml:"model"`
			} `yaml:"claude"`
			OpenAI struct {
				APIKey string `yaml:"api_key"`
				Model  string `yaml:"model"`
			} `yaml:"openai"`
		})
		if ok {
			if cfg.Ollama.Model != "" {
				ollamaConfig.Model = cfg.Ollama.Model
			}
			if cfg.Ollama.Host != "" {
				ollamaConfig.Host = cfg.Ollama.Host
			}
		}
	}

	switch name {
	case "ollama":
		return NewOllamaProvider(ollamaConfig)
	case "claude":
		cfg, ok := config.(struct {
			Claude struct {
				APIKey string `yaml:"api_key"`
				Model  string `yaml:"model"`
			} `yaml:"claude"`
		})
		if !ok || cfg.Claude.APIKey == "" {
			return nil, fmt.Errorf("a Claude API key is required")
		}
		return NewClaudeProvider(ClaudeConfig{
			APIKey: cfg.Claude.APIKey,
			Model:  cfg.Claude.Model,
		})
	case "openai":
		cfg, ok := config.(struct {
			OpenAI struct {
				APIKey string `yaml:"api_key"`
				Model  string `yaml:"model"`
			} `yaml:"openai"`
		})
		if !ok || cfg.OpenAI.APIKey == "" {
			return nil, fmt.Errorf("OpenAI API key is required")
		}
		return NewOpenAIProvider(OpenAIConfig{
			APIKey: cfg.OpenAI.APIKey,
			Model:  cfg.OpenAI.Model,
		})
	default:
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
}
