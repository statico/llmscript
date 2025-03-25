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
	generateScriptsFunc func(ctx context.Context, description string) (llm.ScriptPair, error)
	fixScriptsFunc      func(ctx context.Context, scripts llm.ScriptPair, error string) (llm.ScriptPair, error)
}

func (m *mockLLMProvider) GenerateScripts(ctx context.Context, description string) (llm.ScriptPair, error) {
	return m.generateScriptsFunc(ctx, description)
}

func (m *mockLLMProvider) FixScripts(ctx context.Context, scripts llm.ScriptPair, error string) (llm.ScriptPair, error) {
	return m.fixScriptsFunc(ctx, scripts, error)
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
		generateScriptsFunc: func(ctx context.Context, description string) (llm.ScriptPair, error) {
			return llm.ScriptPair{
				MainScript: `#!/bin/bash
echo "Hello, World!"`,
				TestScript: `#!/bin/bash
echo "Hello, World!"`,
			}, nil
		},
		fixScriptsFunc: func(ctx context.Context, scripts llm.ScriptPair, error string) (llm.ScriptPair, error) {
			return scripts, nil
		},
	}

	// Create a new pipeline
	pipeline, err := NewPipeline(mockLLM, 1, 1, 5*time.Second, tmpDir, false, false)
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
	scripts := llm.ScriptPair{
		MainScript: script,
		TestScript: `#!/bin/bash
echo "Hello, World!"`,
	}
	err = pipeline.runTestScript(context.Background(), scripts)
	require.NoError(t, err)
}
