package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTestRunner(t *testing.T) {
	runner, err := NewTestRunner()
	if err != nil {
		t.Fatalf("Failed to create test runner: %v", err)
	}
	defer runner.Cleanup()

	ctx := context.Background()

	tests := []struct {
		name     string
		test     Test
		wantPass bool
	}{
		{
			name: "simple file creation",
			test: Test{
				Name:    "simple_file_creation",
				Command: "echo 'hello' > output.txt",
				Timeout: 5 * time.Second,
				Expected: map[string]string{
					"output.txt": "hello\n",
				},
			},
			wantPass: true,
		},
		{
			name: "setup and teardown",
			test: Test{
				Name:    "setup_and_teardown",
				Command: "echo 'test' > test.txt",
				Timeout: 5 * time.Second,
				Setup: []string{
					"mkdir -p testdir",
				},
				Teardown: []string{
					"rm -rf testdir",
				},
				Expected: map[string]string{
					"test.txt": "test\n",
				},
			},
			wantPass: true,
		},
		{
			name: "environment variables",
			test: Test{
				Name:    "environment_variables",
				Command: "echo $TEST_VAR > env.txt",
				Timeout: 5 * time.Second,
				Environment: map[string]string{
					"TEST_VAR": "test_value",
				},
				Expected: map[string]string{
					"env.txt": "test_value\n",
				},
			},
			wantPass: true,
		},
		{
			name: "timeout",
			test: Test{
				Name:    "timeout_test",
				Command: "sleep 2",
				Timeout: 100 * time.Millisecond,
			},
			wantPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := runner.RunTest(ctx, tt.test)
			if err != nil {
				t.Fatalf("RunTest failed: %v", err)
			}

			if result.Passed != tt.wantPass {
				t.Errorf("RunTest() passed = %v, want %v\nError: %v\nStdout: %s\nStderr: %s",
					result.Passed, tt.wantPass, result.Error, result.Stdout, result.Stderr)
			}
		})
	}
}

func TestTestRunnerCleanup(t *testing.T) {
	runner, err := NewTestRunner()
	if err != nil {
		t.Fatalf("Failed to create test runner: %v", err)
	}

	// Create a file in the work directory
	testFile := filepath.Join(runner.workDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Cleanup should remove the work directory
	if err := runner.Cleanup(); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify the directory is gone
	if _, err := os.Stat(runner.workDir); !os.IsNotExist(err) {
		t.Error("Work directory still exists after cleanup")
	}
}
