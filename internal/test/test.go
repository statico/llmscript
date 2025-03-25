package test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const defaultTimeout = 30 * time.Second

// Test represents a test case that will be run against a script
type Test struct {
	Name        string            // Name of the test case
	Command     string            // The command to execute
	Setup       []string          // Commands to run before the test
	Teardown    []string          // Commands to run after the test
	Input       map[string]string // Input files/directories to create
	Expected    map[string]string // Expected file contents after script runs
	Environment map[string]string // Environment variables to set
	Timeout     time.Duration     // Maximum time to wait for test completion
}

// TestResult represents the result of running a test
type TestResult struct {
	Test     Test
	Passed   bool
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	Error    error
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

// RunTest executes a single test case in a controlled environment
func (r *TestRunner) RunTest(ctx context.Context, test Test) (*TestResult, error) {
	result := &TestResult{
		Test: test,
	}

	// Create test directory
	testDir := filepath.Join(r.workDir, test.Name)
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create test directory: %w", err)
	}

	// Setup test environment
	if err := r.setupTest(testDir, test); err != nil {
		return nil, fmt.Errorf("test setup failed: %w", err)
	}

	// Create a context with timeout if not already set
	if test.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, test.Timeout)
		defer cancel()
	} else {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}

	// Run setup commands
	if err := r.runCommands(ctx, testDir, test.Setup); err != nil {
		return nil, fmt.Errorf("setup commands failed: %w", err)
	}

	// Run the test
	start := time.Now()
	cmd := exec.CommandContext(ctx, "bash", "-c", test.Command)
	cmd.Dir = testDir
	cmd.Env = r.buildEnv(test.Environment)

	var stdout, stderr []byte
	var err error
	stdout, err = cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = exitErr.Stderr
			result.ExitCode = exitErr.ExitCode()
		}
		result.Error = err
		result.Passed = false
	}

	result.Duration = time.Since(start)
	result.Stdout = string(stdout)
	result.Stderr = string(stderr)

	// Run teardown commands
	if err := r.runCommands(ctx, testDir, test.Teardown); err != nil {
		return nil, fmt.Errorf("teardown commands failed: %w", err)
	}

	// Only verify results if the command succeeded
	if result.Error == nil {
		if err := r.verifyResults(testDir, test.Expected); err != nil {
			result.Error = err
			result.Passed = false
			return result, nil
		}
		result.Passed = true
	}

	return result, nil
}

// Cleanup removes the temporary working directory
func (r *TestRunner) Cleanup() error {
	return os.RemoveAll(r.workDir)
}

func (r *TestRunner) setupTest(testDir string, test Test) error {
	// Create input files
	for path, content := range test.Input {
		fullPath := filepath.Join(testDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create input file %s: %w", path, err)
		}
	}
	return nil
}

func (r *TestRunner) runCommands(ctx context.Context, testDir string, commands []string) error {
	for _, cmd := range commands {
		command := exec.CommandContext(ctx, "bash", "-c", cmd)
		command.Dir = testDir
		if output, err := command.CombinedOutput(); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("command timed out")
			}
			return fmt.Errorf("command failed: %s\nOutput: %s", err, output)
		}
	}
	return nil
}

func (r *TestRunner) verifyResults(testDir string, expected map[string]string) error {
	for path, expectedContent := range expected {
		fullPath := filepath.Join(testDir, path)
		actualContent, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read result file %s: %w", path, err)
		}
		if string(actualContent) != expectedContent {
			return fmt.Errorf("content mismatch in %s: expected %q, got %q", path, expectedContent, string(actualContent))
		}
	}
	return nil
}

func (r *TestRunner) buildEnv(env map[string]string) []string {
	// Start with current environment
	result := os.Environ()

	// Add test-specific environment variables
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}

	return result
}
