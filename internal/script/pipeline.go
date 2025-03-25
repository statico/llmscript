package script

import (
	"context"
	"fmt"
	"time"

	"github.com/statico/llmscript/internal/llm"
	"github.com/statico/llmscript/internal/log"
)

// Test represents a test case for a script
type Test = llm.Test

// TestFailure represents a failed test case
type TestFailure = llm.TestFailure

// Pipeline handles the script generation and testing process
type Pipeline struct {
	llm          llm.Provider
	maxFixes     int
	maxAttempts  int
	timeout      time.Duration
	executor     *Executor
	cache        *Cache
	showProgress bool
}

// NewPipeline creates a new script generation pipeline
func NewPipeline(llm llm.Provider, maxFixes, maxAttempts int, timeout time.Duration, workDir string, showProgress bool) (*Pipeline, error) {
	executor, err := NewExecutor(workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	cache, err := NewCache()
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	return &Pipeline{
		llm:          llm,
		maxFixes:     maxFixes,
		maxAttempts:  maxAttempts,
		timeout:      timeout,
		executor:     executor,
		cache:        cache,
		showProgress: showProgress,
	}, nil
}

// GenerateAndTest generates a script from a natural language description and tests it
func (p *Pipeline) GenerateAndTest(ctx context.Context, description string) (string, error) {
	// Check cache first
	if script, tests, err := p.cache.Get(description); err == nil && script != "" {
		if p.showProgress {
			log.Info("Found cached script, verifying...")
		}
		// Run cached tests to verify
		failures := p.runTests(ctx, script, tests)
		if len(failures) == 0 {
			if p.showProgress {
				log.Success("Cached script verified successfully")
			}
			return script, nil
		}
		if p.showProgress {
			log.Warn("Cached script failed verification, generating new script")
		}
	}

	// Generate initial script
	if p.showProgress {
		log.Info("Generating initial script...")
	}
	script, err := p.llm.GenerateScript(ctx, description)
	if err != nil {
		return "", fmt.Errorf("failed to generate initial script: %w", err)
	}
	if p.showProgress {
		log.Debug("Initial script generated:\n%s", script)
	}

	// Generate test cases
	if p.showProgress {
		log.Info("Generating test cases...")
	}
	tests, err := p.llm.GenerateTests(ctx, description)
	if err != nil {
		return "", fmt.Errorf("failed to generate test cases: %w", err)
	}
	if p.showProgress {
		log.Debug("Generated %d test cases", len(tests))
		for i, test := range tests {
			log.Debug("Test %d: %s", i+1, test.Name)
		}
	}

	// Run test cases and fix failures
	for attempt := 0; attempt < p.maxAttempts; attempt++ {
		if p.showProgress {
			log.Info("Attempt %d/%d: Running tests...", attempt+1, p.maxAttempts)
		}
		failures := p.runTests(ctx, script, tests)
		if len(failures) == 0 {
			if p.showProgress {
				log.Success("All tests passed!")
			}
			// Cache successful script and tests
			if err := p.cache.Set(description, script, tests); err != nil {
				log.Warn("Failed to cache script: %v", err)
			} else if p.showProgress {
				log.Debug("Script and tests cached successfully")
			}
			return script, nil
		}

		if p.showProgress {
			log.Warn("Found %d failing tests", len(failures))
			for i, failure := range failures {
				log.Debug("Test failure %d: %s", i+1, failure.Test.Name)
				if failure.Error != nil {
					log.Debug("Error: %v", failure.Error)
				}
				if failure.Output != "" {
					log.Debug("Output: %s", failure.Output)
				}
			}
		}

		// Fix script based on failures
		for fix := 0; fix < p.maxFixes; fix++ {
			if p.showProgress {
				log.Info("Fix %d/%d: Attempting to fix script...", fix+1, p.maxFixes)
			}
			script, err = p.llm.FixScript(ctx, script, failures)
			if err != nil {
				return "", fmt.Errorf("failed to fix script: %w", err)
			}
			if p.showProgress {
				log.Debug("Fixed script:\n%s", script)
			}

			failures = p.runTests(ctx, script, tests)
			if len(failures) == 0 {
				if p.showProgress {
					log.Success("All tests passed after fix!")
				}
				// Cache successful script and tests
				if err := p.cache.Set(description, script, tests); err != nil {
					log.Warn("Failed to cache script: %v", err)
				} else if p.showProgress {
					log.Debug("Script and tests cached successfully")
				}
				return script, nil
			}
			if p.showProgress {
				log.Warn("Fix %d/%d: Still have %d failing tests", fix+1, p.maxFixes, len(failures))
			}
		}

		// If we've exhausted fixes, try generating a new script
		if p.showProgress {
			log.Info("Attempt %d/%d: Generating new script...", attempt+1, p.maxAttempts)
		}
		script, err = p.llm.GenerateScript(ctx, description)
		if err != nil {
			return "", fmt.Errorf("failed to generate new script: %w", err)
		}
		if p.showProgress {
			log.Debug("New script generated:\n%s", script)
		}
	}

	return "", fmt.Errorf("failed to generate working script after %d attempts", p.maxAttempts)
}

// runTests executes all test cases and returns any failures
func (p *Pipeline) runTests(ctx context.Context, script string, tests []Test) []TestFailure {
	var failures []TestFailure
	for i, test := range tests {
		if p.showProgress {
			log.Debug("Running test %d/%d...", i+1, len(tests))
		}
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
