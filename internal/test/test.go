package test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Test represents a test case for a script
type Test struct {
	Name        string
	Input       string
	WantOutput  string
	WantError   bool
	WantExitErr bool
}

// TestRunner handles executing tests in a controlled environment
type TestRunner struct {
	workDir string
}

// NewTestRunner creates a new test runner with a temporary working directory
func NewTestRunner() (*TestRunner, error) {
	workDir, err := os.MkdirTemp("", "llmscript-test-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	return &TestRunner{workDir: workDir}, nil
}

// RunTest executes a test script in a controlled environment
func (r *TestRunner) RunTest(ctx context.Context, mainScript, testScript string) error {
	// Create test directory
	testDir := filepath.Join(r.workDir, "test")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	// Write scripts to files
	mainScriptPath := filepath.Join(testDir, "script.sh")
	testScriptPath := filepath.Join(testDir, "test.sh")

	if err := os.WriteFile(mainScriptPath, []byte(mainScript), 0750); err != nil {
		return fmt.Errorf("failed to write main script: %w", err)
	}
	if err := os.WriteFile(testScriptPath, []byte(testScript), 0750); err != nil {
		return fmt.Errorf("failed to write test script: %w", err)
	}

	// Run the test script
	cmd := exec.CommandContext(ctx, testScriptPath)
	cmd.Dir = testDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("test failed: %w\nOutput:\n%s", err, output)
	}

	return nil
}

// Cleanup removes the temporary working directory
func (r *TestRunner) Cleanup() error {
	return os.RemoveAll(r.workDir)
}
