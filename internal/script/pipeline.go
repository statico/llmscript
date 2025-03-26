package script

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	workDir      string
	cache        *Cache
	showProgress bool
	noCache      bool
}

// NewPipeline creates a new script generation pipeline
func NewPipeline(llm llm.Provider, maxFixes, maxAttempts int, timeout time.Duration, workDir string, showProgress bool, noCache bool) (*Pipeline, error) {
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	var cache *Cache
	if !noCache {
		var err error
		cache, err = NewCache()
		if err != nil {
			return nil, fmt.Errorf("failed to create cache: %w", err)
		}
	}

	return &Pipeline{
		llm:          llm,
		maxFixes:     maxFixes,
		maxAttempts:  maxAttempts,
		timeout:      timeout,
		workDir:      workDir,
		cache:        cache,
		showProgress: showProgress,
		noCache:      noCache,
	}, nil
}

// GenerateAndTest generates a script from a natural language description and tests it
func (p *Pipeline) GenerateAndTest(ctx context.Context, description string) (string, error) {
	// Check cache first if enabled
	if !p.noCache && p.cache != nil {
		if scripts, err := p.cache.Get(description); err == nil && scripts.MainScript != "" {
			if p.showProgress {
				log.Info("Found cached scripts, verifying...")
			}
			// Run test script to verify
			if err := p.runTestScript(ctx, scripts); err == nil {
				if p.showProgress {
					log.Success("Cached scripts verified successfully")
				}
				return scripts.MainScript, nil
			}
			if p.showProgress {
				log.Warn("Cached scripts failed verification, generating new scripts")
			}
		}
	}

	// Generate initial script and tests
	if p.showProgress {
		log.Info("Generating initial script...")
	}
	script, err := p.llm.GenerateScript(ctx, description)
	if err != nil {
		return "", fmt.Errorf("failed to generate initial script: %w", err)
	}
	if p.showProgress {
		log.Debug("Initial script generated")
	}

	// Generate tests for the script
	if p.showProgress {
		log.Info("Generating tests...")
	}
	tests, err := p.llm.GenerateTests(ctx, script, description)
	if err != nil {
		return "", fmt.Errorf("failed to generate tests: %w", err)
	}
	if p.showProgress {
		log.Debug("Tests generated")
	}

	// Run test script and fix failures
	for attempt := 0; attempt < p.maxAttempts; attempt++ {
		if attempt > 0 {
			if p.showProgress {
				log.Info("Attempt %d/%d: Generating new script...", attempt+1, p.maxAttempts)
			}
			script, err = p.llm.GenerateScript(ctx, description)
			if err != nil {
				return "", fmt.Errorf("failed to generate new script: %w", err)
			}
			if p.showProgress {
				log.Debug("New script generated")
			}

			// Generate new tests for the new script
			if p.showProgress {
				log.Info("Generating new tests...")
			}
			tests, err = p.llm.GenerateTests(ctx, script, description)
			if err != nil {
				return "", fmt.Errorf("failed to generate tests: %w", err)
			}
			if p.showProgress {
				log.Debug("New tests generated")
			}
		}

		// Try to fix any failures
		for fix := 0; fix < p.maxFixes; fix++ {
			// Run tests first to get initial failures
			failures, err := p.runTests(ctx, script, tests)
			if err != nil {
				return "", fmt.Errorf("failed to run tests: %w", err)
			}
			if len(failures) == 0 {
				// Cache successful script if caching is enabled
				if !p.noCache && p.cache != nil {
					if err := p.cache.Set(description, llm.ScriptPair{
						MainScript: script,
						TestScript: generateTestScript(tests),
					}); err != nil {
						log.Warn("Failed to cache successful script: %v", err)
					}
				}
				return script, nil
			}

			if fix < p.maxFixes-1 { // Don't try to fix on the last iteration
				if p.showProgress {
					log.Info("Fix attempt %d/%d...", fix+1, p.maxFixes)
				}
				script, err = p.llm.FixScript(ctx, script, failures)
				if err != nil {
					return "", fmt.Errorf("failed to fix script: %w", err)
				}
				if p.showProgress {
					log.Debug("Script fixed")
				}
			}
		}
	}

	return "", fmt.Errorf("failed to generate working script after %d attempts", p.maxAttempts)
}

// runTestScript executes the test script in a controlled environment
func (p *Pipeline) runTestScript(ctx context.Context, scripts llm.ScriptPair) error {
	// Create a secure temporary directory for this test
	testDir, err := os.MkdirTemp(p.workDir, "test-*")
	if err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}
	defer os.RemoveAll(testDir)

	// Write both scripts to files
	mainScriptPath := filepath.Join(testDir, "script.sh")
	testScriptPath := filepath.Join(testDir, "test.sh")

	if err := os.WriteFile(mainScriptPath, []byte(scripts.MainScript), 0750); err != nil {
		return fmt.Errorf("failed to write main script: %w", err)
	}
	if err := os.WriteFile(testScriptPath, []byte(scripts.TestScript), 0750); err != nil {
		return fmt.Errorf("failed to write test script: %w", err)
	}

	// Run the test script
	cmd := exec.CommandContext(ctx, "sh", testScriptPath)
	cmd.Dir = testDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("test script failed: %w\nOutput:\n%s", err, output)
	}

	return nil
}

// runTests executes the tests in a controlled environment
func (p *Pipeline) runTests(ctx context.Context, script string, tests []Test) ([]TestFailure, error) {
	// Create a secure temporary directory for this test
	testDir, err := os.MkdirTemp(p.workDir, "test-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create test directory: %w", err)
	}
	defer os.RemoveAll(testDir)

	// Write the script to a file
	scriptPath := filepath.Join(testDir, "script.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0750); err != nil {
		return nil, fmt.Errorf("failed to write script: %w", err)
	}

	var failures []TestFailure
	for _, test := range tests {
		// Create a clean test directory for each test
		testCaseDir := filepath.Join(testDir, test.Name)
		if err := os.MkdirAll(testCaseDir, 0750); err != nil {
			return nil, fmt.Errorf("failed to create test case directory: %w", err)
		}

		// Run setup commands
		for _, cmd := range test.Setup {
			setupCmd := exec.CommandContext(ctx, "sh", "-c", cmd)
			setupCmd.Dir = testCaseDir
			setupCmd.Env = os.Environ()
			for k, v := range test.Environment {
				setupCmd.Env = append(setupCmd.Env, fmt.Sprintf("%s=%s", k, v))
			}
			if output, err := setupCmd.CombinedOutput(); err != nil {
				failures = append(failures, TestFailure{
					Test:     test,
					Output:   string(output),
					Error:    fmt.Errorf("setup failed: %w", err),
					ExitCode: -1,
				})
				continue
			}
		}

		// Run the test
		cmd := exec.CommandContext(ctx, scriptPath)
		cmd.Dir = testCaseDir
		cmd.Env = os.Environ()
		for k, v := range test.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}

		var stdout, stderr strings.Builder
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		// Set up input if provided
		if test.Input != "" {
			stdin, err := cmd.StdinPipe()
			if err != nil {
				return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
			}
			go func() {
				defer stdin.Close()
				fmt.Fprint(stdin, test.Input)
			}()
		}

		// Run with timeout
		err := cmd.Start()
		if err != nil {
			failures = append(failures, TestFailure{
				Test:     test,
				Output:   stdout.String() + stderr.String(),
				Error:    fmt.Errorf("failed to start test: %w", err),
				ExitCode: -1,
			})
			continue
		}

		done := make(chan error)
		go func() {
			done <- cmd.Wait()
		}()

		timeout := test.Timeout
		if timeout == 0 {
			timeout = p.timeout
		}

		select {
		case err := <-done:
			if err != nil {
				var exitErr *exec.ExitError
				exitCode := -1
				if errors.As(err, &exitErr) {
					exitCode = exitErr.ExitCode()
				}
				failures = append(failures, TestFailure{
					Test:     test,
					Output:   stdout.String() + stderr.String(),
					Error:    err,
					ExitCode: exitCode,
				})
				continue
			}
		case <-time.After(timeout):
			if err := cmd.Process.Kill(); err != nil {
				log.Warn("Failed to kill process: %v", err)
			}
			failures = append(failures, TestFailure{
				Test:     test,
				Output:   stdout.String() + stderr.String(),
				Error:    fmt.Errorf("test timed out after %v", timeout),
				ExitCode: -1,
			})
			continue
		}

		// Compare output
		output := stdout.String()
		if output != test.Expected {
			failures = append(failures, TestFailure{
				Test:     test,
				Output:   output,
				Error:    fmt.Errorf("output does not match expected:\nExpected: %s\nGot: %s", test.Expected, output),
				ExitCode: 0,
			})
		}
	}

	return failures, nil
}

// generateTestScript converts a list of tests into a shell script
func generateTestScript(tests []Test) string {
	var b strings.Builder
	b.WriteString("#!/bin/sh\nset -e\n\n")
	for _, test := range tests {
		b.WriteString(fmt.Sprintf("# Test: %s\n", test.Name))
		for _, cmd := range test.Setup {
			b.WriteString(cmd + "\n")
		}
		b.WriteString("echo '" + test.Input + "' | ./script.sh\n")
		b.WriteString("[ \"$(echo '" + test.Input + "' | ./script.sh)\" = \"" + test.Expected + "\" ] || exit 1\n\n")
	}
	return b.String()
}
