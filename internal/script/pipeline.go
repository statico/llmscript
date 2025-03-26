package script

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	llm         llm.Provider
	maxFixes    int
	maxAttempts int
	timeout     time.Duration
	workDir     string
	cache       *Cache
	noCache     bool
}

// NewPipeline creates a new script generation pipeline
func NewPipeline(llm llm.Provider, maxFixes, maxAttempts int, timeout time.Duration, workDir string, noCache bool) (*Pipeline, error) {
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
		llm:         llm,
		maxFixes:    maxFixes,
		maxAttempts: maxAttempts,
		timeout:     timeout,
		workDir:     workDir,
		cache:       cache,
		noCache:     noCache,
	}, nil
}

// GenerateAndTest generates a script from a natural language description and tests it
func (p *Pipeline) GenerateAndTest(ctx context.Context, description string) (string, error) {
	// Check cache first if enabled
	if !p.noCache && p.cache != nil {
		log.Info("Checking cache...")
		if scripts, err := p.cache.Get(description); err == nil && scripts.MainScript != "" {
			// Run test script to verify
			if err := p.runTestScript(ctx, scripts); err == nil {
				log.Success("Cached script found")
				return scripts.MainScript, nil
			}
			log.Warn("Cached scripts failed verification, generating new scripts")
		}
	}

	// Generate initial scripts
	log.Info("Generating initial scripts with %s...", p.llm.Name())
	scripts, err := p.llm.GenerateScripts(ctx, description)
	if err != nil {
		return "", fmt.Errorf("failed to generate initial scripts: %w", err)
	}
	log.Debug("Initial scripts generated")

	// Run test script and fix failures
	for attempt := 0; attempt < p.maxAttempts; attempt++ {
		if attempt > 0 {
			log.Info("Attempt %d/%d: Generating new scripts...", attempt+1, p.maxAttempts)
			scripts, err = p.llm.GenerateScripts(ctx, description)
			if err != nil {
				return "", fmt.Errorf("failed to generate new scripts: %w", err)
			}
			log.Debug("New scripts generated")
		}

		// Try to fix any failures
		for fix := 0; fix < p.maxFixes; fix++ {
			// Run test script
			log.Info("Testing script (attempt %d/%d)...", attempt+1, p.maxAttempts)
			err := p.runTestScript(ctx, scripts)
			if err == nil {
				// Cache successful scripts if caching is enabled
				if !p.noCache && p.cache != nil {
					log.Info("Caching successful scripts...")
					if err := p.cache.Set(description, scripts); err != nil {
						log.Warn("Failed to cache successful scripts: %v", err)
					}
				}
				return scripts.MainScript, nil
			}

			if fix < p.maxFixes-1 { // Don't try to fix on the last iteration
				log.Info("Fix attempt %d/%d...", fix+1, p.maxFixes)
				scripts, err = p.llm.FixScripts(ctx, scripts, err.Error())
				if err != nil {
					return "", fmt.Errorf("failed to fix scripts: %w", err)
				}
				log.Debug("Scripts fixed")
				log.Info("New script:\n%s", scripts.MainScript)
			}
		}
	}

	return "", fmt.Errorf("failed to generate working scripts after %d attempts", p.maxAttempts)
}

// runTestScript executes the test script in a controlled environment
func (p *Pipeline) runTestScript(ctx context.Context, scripts llm.ScriptPair) error {
	// Create a secure temporary directory for this test
	testDir, err := os.MkdirTemp("", "llmscript-test-*")
	if err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			log.Error("failed to remove test directory: %v", err)
		}
	}()

	// Write both scripts to files
	featureScriptPath := filepath.Join(testDir, "script.sh")
	testScriptPath := filepath.Join(testDir, "test.sh")

	if err := os.WriteFile(featureScriptPath, []byte(scripts.MainScript), 0750); err != nil {
		return fmt.Errorf("failed to write feature script: %w", err)
	}
	if err := os.WriteFile(testScriptPath, []byte(scripts.TestScript), 0750); err != nil {
		return fmt.Errorf("failed to write test script: %w", err)
	}

	// Run the test script with timeout
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, testScriptPath)
	cmd.Dir = testDir

	output, err := cmd.CombinedOutput()
	log.Debug("Test script output:\n%s", output)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Debug("Test script exited with code: %d", exitErr.ExitCode())
		}
		return fmt.Errorf("test script failed: %w\nOutput:\n%s", err, output)
	}
	log.Debug("Test script exited with code: 0")

	return nil
}
