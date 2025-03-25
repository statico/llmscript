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

	// Generate initial scripts
	if p.showProgress {
		log.Info("Generating initial scripts...")
	}
	scripts, err := p.llm.GenerateScripts(ctx, description)
	if err != nil {
		return "", fmt.Errorf("failed to generate initial scripts: %w", err)
	}
	if p.showProgress {
		log.Debug("Initial scripts generated")
	}

	// Run test script and fix failures
	for attempt := 0; attempt < p.maxAttempts; attempt++ {
		if attempt > 0 {
			if p.showProgress {
				log.Info("Attempt %d/%d: Generating new scripts...", attempt+1, p.maxAttempts)
			}
			scripts, err = p.llm.GenerateScripts(ctx, description)
			if err != nil {
				return "", fmt.Errorf("failed to generate new scripts: %w", err)
			}
			if p.showProgress {
				log.Debug("New scripts generated")
			}
		}

		// Try to fix any failures
		for fix := 0; fix < p.maxFixes; fix++ {
			if fix > 0 {
				if p.showProgress {
					log.Info("Fix attempt %d/%d...", fix+1, p.maxFixes)
				}
				scripts, err = p.llm.FixScripts(ctx, scripts, "Test script failed")
				if err != nil {
					return "", fmt.Errorf("failed to fix scripts: %w", err)
				}
				if p.showProgress {
					log.Debug("Scripts fixed")
				}
			}

			// Run test script
			if err := p.runTestScript(ctx, scripts); err == nil {
				// Cache successful scripts if caching is enabled
				if !p.noCache && p.cache != nil {
					if err := p.cache.Set(description, scripts); err != nil {
						log.Warn("Failed to cache successful scripts: %v", err)
					}
				}
				return scripts.MainScript, nil
			}
		}
	}

	return "", fmt.Errorf("failed to generate working scripts after %d attempts", p.maxAttempts)
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
