package llm

import (
	"context"
	"testing"
)

func TestBaseProvider(t *testing.T) {
	provider := &BaseProvider{
		Config: struct{}{},
	}

	ctx := context.Background()

	t.Run("GenerateScript returns error", func(t *testing.T) {
		script, err := provider.GenerateScript(ctx, "test description")
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if script != "" {
			t.Errorf("Expected empty string, got %q", script)
		}
	})

	t.Run("GenerateTests returns error", func(t *testing.T) {
		tests, err := provider.GenerateTests(ctx, "test description")
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if tests != nil {
			t.Error("Expected nil tests, got non-nil")
		}
	})

	t.Run("FixScript returns error", func(t *testing.T) {
		script, err := provider.FixScript(ctx, "test script", nil)
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if script != "" {
			t.Errorf("Expected empty string, got %q", script)
		}
	})

	t.Run("ValidateConfig with nil config", func(t *testing.T) {
		provider.Config = nil
		err := provider.ValidateConfig()
		if err == nil {
			t.Error("Expected error for nil config, got nil")
		}
	})

	t.Run("ValidateConfig with valid config", func(t *testing.T) {
		provider.Config = struct{}{}
		err := provider.ValidateConfig()
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
		}
	})
}

func TestTestStruct(t *testing.T) {
	test := Test{
		Name:        "test1",
		Description: "test description",
		Setup:       []string{"setup1", "setup2"},
		Input:       "test input",
		Expected:    "test output",
	}

	if test.Name != "test1" {
		t.Errorf("Expected Name to be 'test1', got %q", test.Name)
	}
	if test.Description != "test description" {
		t.Errorf("Expected Description to be 'test description', got %q", test.Description)
	}
	if len(test.Setup) != 2 {
		t.Errorf("Expected Setup to have 2 items, got %d", len(test.Setup))
	}
	if test.Input != "test input" {
		t.Errorf("Expected Input to be 'test input', got %q", test.Input)
	}
	if test.Expected != "test output" {
		t.Errorf("Expected Expected to be 'test output', got %q", test.Expected)
	}
}

func TestTestFailureStruct(t *testing.T) {
	test := Test{
		Name:        "test1",
		Description: "test description",
		Setup:       []string{"setup1"},
		Input:       "test input",
		Expected:    "test output",
	}

	failure := TestFailure{
		Test:     test,
		Actual:   "actual output",
		Error:    nil,
		ExitCode: 1,
		Stdout:   "stdout",
		Stderr:   "stderr",
	}

	if failure.Test.Name != "test1" {
		t.Errorf("Expected Test.Name to be 'test1', got %q", failure.Test.Name)
	}
	if failure.Actual != "actual output" {
		t.Errorf("Expected Actual to be 'actual output', got %q", failure.Actual)
	}
	if failure.ExitCode != 1 {
		t.Errorf("Expected ExitCode to be 1, got %d", failure.ExitCode)
	}
	if failure.Stdout != "stdout" {
		t.Errorf("Expected Stdout to be 'stdout', got %q", failure.Stdout)
	}
	if failure.Stderr != "stderr" {
		t.Errorf("Expected Stderr to be 'stderr', got %q", failure.Stderr)
	}
}
