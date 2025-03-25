package script

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/statico/llmscript/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLLMProvider implements the LLM provider interface for testing
type mockLLMProvider struct {
	generateScriptFunc func(ctx context.Context, description string) (string, error)
	generateTestsFunc  func(ctx context.Context, description string) ([]llm.Test, error)
	fixScriptFunc      func(ctx context.Context, script string, failures []llm.TestFailure) (string, error)
}

func (m *mockLLMProvider) GenerateScript(ctx context.Context, description string) (string, error) {
	return m.generateScriptFunc(ctx, description)
}

func (m *mockLLMProvider) GenerateTests(ctx context.Context, description string) ([]llm.Test, error) {
	return m.generateTestsFunc(ctx, description)
}

func (m *mockLLMProvider) FixScript(ctx context.Context, script string, failures []llm.TestFailure) (string, error) {
	return m.fixScriptFunc(ctx, script, failures)
}

func TestPipeline_GenerateAndTest(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "llmscript-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock LLM provider
	mockLLM := &mockLLMProvider{
		generateScriptFunc: func(ctx context.Context, description string) (string, error) {
			return `#!/bin/bash
echo "Hello, World!"`, nil
		},
		generateTestsFunc: func(ctx context.Context, description string) ([]llm.Test, error) {
			return []llm.Test{
				{
					Name:     "basic_test",
					Input:    "",
					Expected: "Hello, World!\n",
				},
			}, nil
		},
		fixScriptFunc: func(ctx context.Context, script string, failures []llm.TestFailure) (string, error) {
			return script, nil
		},
	}

	// Create a new pipeline
	pipeline, err := NewPipeline(mockLLM, 1, 1, 5*time.Second, tmpDir)
	require.NoError(t, err)

	// Test script generation
	script, err := pipeline.GenerateAndTest(context.Background(), "Print 'Hello, World!'")
	require.NoError(t, err)

	// Verify the script was generated
	assert.Contains(t, script, "#!/bin/bash")
	assert.Contains(t, script, "echo \"Hello, World!\"")

	// Write the script to a file
	scriptPath := filepath.Join(tmpDir, "script.sh")
	err = os.WriteFile(scriptPath, []byte(script), 0755)
	require.NoError(t, err)

	// Verify the script file was created
	scriptContent, err := os.ReadFile(scriptPath)
	require.NoError(t, err)
	assert.Equal(t, script, string(scriptContent))

	// Test script execution
	output, err := pipeline.executor.ExecuteTest(context.Background(), script, llm.Test{
		Name:     "basic_test",
		Input:    "",
		Expected: "Hello, World!\n",
	})
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!\n", output)
}
