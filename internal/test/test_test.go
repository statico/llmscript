package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestTestRunner(t *testing.T) {
	runner, err := NewTestRunner()
	if err != nil {
		t.Fatalf("Failed to create test runner: %v", err)
	}
	defer runner.Cleanup()

	ctx := context.Background()

	tests := []struct {
		name       string
		mainScript string
		testScript string
		wantErr    bool
	}{
		{
			name: "simple echo test",
			mainScript: `#!/bin/bash
echo "hello"`,
			testScript: `#!/bin/bash
set -e
[ "$(./script.sh)" = "hello" ] || exit 1`,
			wantErr: false,
		},
		{
			name: "failing test",
			mainScript: `#!/bin/bash
echo "wrong"`,
			testScript: `#!/bin/bash
set -e
[ "$(./script.sh)" = "right" ] || exit 1`,
			wantErr: true,
		},
		{
			name: "test with setup",
			mainScript: `#!/bin/bash
cat input.txt`,
			testScript: `#!/bin/bash
set -e
echo "test data" > input.txt
[ "$(./script.sh)" = "test data" ] || exit 1`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runner.RunTest(ctx, tt.mainScript, tt.testScript)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunTest() error = %v, wantErr %v", err, tt.wantErr)
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
