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
	"github.com/statico/llmscript/internal/progress"
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
	Spinner     *progress.Spinner
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

	spinner := progress.NewSpinner("Initializing...")

	return &Pipeline{
		llm:         llm,
		maxFixes:    maxFixes,
		maxAttempts: maxAttempts,
		timeout:     timeout,
		workDir:     workDir,
		cache:       cache,
		noCache:     noCache,
		Spinner:     spinner,
	}, nil
}

// GenerateAndTest generates a script from a natural language description and tests it
func (p *Pipeline) GenerateAndTest(ctx context.Context, description string) (string, error) {
	// Check cache first if enabled
	if !p.noCache && p.cache != nil {
		p.Spinner.SetMessage("Checking cache...")
		if scripts, err := p.cache.Get(description); err == nil && scripts.MainScript != "" {
			p.Spinner.SetMessage("Found cached scripts, verifying...")
			// Run test script to verify
			if err := p.runTestScript(ctx, scripts); err == nil {
				p.Spinner.Stop()
				log.Success("Cached scripts verified successfully")
				return scripts.MainScript, nil
			}
			log.Warn("Cached scripts failed verification, generating new scripts")
		}
	}

	// Generate initial scripts
	p.Spinner.SetMessage("Generating initial scripts...")
	scripts, err := p.llm.GenerateScripts(ctx, description)
	if err != nil {
		p.Spinner.Stop()
		return "", fmt.Errorf("failed to generate initial scripts: %w", err)
	}
	log.Debug("Initial scripts generated")

	// Run test script and fix failures
	for attempt := 0; attempt < p.maxAttempts; attempt++ {
		if attempt > 0 {
			p.Spinner.SetMessage(fmt.Sprintf("Attempt %d/%d: Generating new scripts...", attempt+1, p.maxAttempts))
			scripts, err = p.llm.GenerateScripts(ctx, description)
			if err != nil {
				p.Spinner.Stop()
				return "", fmt.Errorf("failed to generate new scripts: %w", err)
			}
			log.Debug("New scripts generated")
		}

		// Try to fix any failures
		for fix := 0; fix < p.maxFixes; fix++ {
			// Run test script
			p.Spinner.SetMessage(fmt.Sprintf("Testing script (attempt %d/%d)...", attempt+1, p.maxAttempts))
			err := p.runTestScript(ctx, scripts)
			if err == nil {
				// Cache successful scripts if caching is enabled
				if !p.noCache && p.cache != nil {
					p.Spinner.SetMessage("Caching successful scripts...")
					if err := p.cache.Set(description, scripts); err != nil {
						log.Warn("Failed to cache successful scripts: %v", err)
					}
				}
				p.Spinner.Stop()
				return scripts.MainScript, nil
			}

			if fix < p.maxFixes-1 { // Don't try to fix on the last iteration
				p.Spinner.SetMessage(fmt.Sprintf("Fix attempt %d/%d...", fix+1, p.maxFixes))
				scripts, err = p.llm.FixScripts(ctx, scripts, err.Error())
				if err != nil {
					p.Spinner.Stop()
					return "", fmt.Errorf("failed to fix scripts: %w", err)
				}
				log.Debug("Scripts fixed")
			}
		}
	}

	p.Spinner.Stop()
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
	if err != nil {
		return fmt.Errorf("test script failed: %w\nOutput:\n%s", err, output)
	}

	return nil
}
