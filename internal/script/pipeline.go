package script

import (
	"context"
	"fmt"
	"time"

	"github.com/statico/llmscript/internal/llm"
)

// Test represents a test case for a script
type Test = llm.Test

// TestFailure represents a failed test case
type TestFailure = llm.TestFailure

// Pipeline handles the script generation and testing process
type Pipeline struct {
	llm         llm.Provider
	maxFixes    int
	maxAttempts int
	timeout     time.Duration
	executor    *Executor
	cache       *Cache
}

// NewPipeline creates a new script generation pipeline
func NewPipeline(llm llm.Provider, maxFixes, maxAttempts int, timeout time.Duration, workDir string) (*Pipeline, error) {
	executor, err := NewExecutor(workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	cache, err := NewCache()
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	return &Pipeline{
		llm:         llm,
		maxFixes:    maxFixes,
		maxAttempts: maxAttempts,
		timeout:     timeout,
		executor:    executor,
		cache:       cache,
	}, nil
}

// GenerateAndTest generates a script from a natural language description and tests it
func (p *Pipeline) GenerateAndTest(ctx context.Context, description string) (string, error) {
	// Check cache first
	if script, tests, err := p.cache.Get(description); err == nil && script != "" {
		// Run cached tests to verify
		failures := p.runTests(ctx, script, tests)
		if len(failures) == 0 {
			return script, nil
		}
	}

	// Generate initial script
	script, err := p.llm.GenerateScript(ctx, description)
	if err != nil {
		return "", fmt.Errorf("failed to generate initial script: %w", err)
	}

	// Generate test cases
	tests, err := p.llm.GenerateTests(ctx, description)
	if err != nil {
		return "", fmt.Errorf("failed to generate test cases: %w", err)
	}

	// Run test cases and fix failures
	for attempt := 0; attempt < p.maxAttempts; attempt++ {
		failures := p.runTests(ctx, script, tests)
		if len(failures) == 0 {
			// Cache successful script and tests
			if err := p.cache.Set(description, script, tests); err != nil {
				// Log cache error but don't fail
				fmt.Printf("Warning: failed to cache script: %v\n", err)
			}
			return script, nil
		}

		// Fix script based on failures
		for fix := 0; fix < p.maxFixes; fix++ {
			script, err = p.llm.FixScript(ctx, script, failures)
			if err != nil {
				return "", fmt.Errorf("failed to fix script: %w", err)
			}

			failures = p.runTests(ctx, script, tests)
			if len(failures) == 0 {
				// Cache successful script and tests
				if err := p.cache.Set(description, script, tests); err != nil {
					// Log cache error but don't fail
					fmt.Printf("Warning: failed to cache script: %v\n", err)
				}
				return script, nil
			}
		}

		// If we've exhausted fixes, try generating a new script
		script, err = p.llm.GenerateScript(ctx, description)
		if err != nil {
			return "", fmt.Errorf("failed to generate new script: %w", err)
		}
	}

	return "", fmt.Errorf("failed to generate working script after %d attempts", p.maxAttempts)
}

// runTests executes all test cases and returns any failures
func (p *Pipeline) runTests(ctx context.Context, script string, tests []Test) []TestFailure {
	var failures []TestFailure
	for _, test := range tests {
		output, err := p.executor.ExecuteTest(ctx, script, test)
		if err != nil {
			failures = append(failures, TestFailure{
				Test:     test,
				Error:    err,
				ExitCode: -1,
			})
			continue
		}

		if output != test.Expected {
			failures = append(failures, TestFailure{
				Test:     test,
				Output:   output,
				ExitCode: 0,
			})
		}
	}
	return failures
}
