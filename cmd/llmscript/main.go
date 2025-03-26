package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/statico/llmscript/internal/config"
	"github.com/statico/llmscript/internal/llm"
	"github.com/statico/llmscript/internal/log"
	"github.com/statico/llmscript/internal/script"
)

var (
	writeConfig = flag.Bool("write-config", false, "Write default config to ~/.config/llmscript/config.yaml")
	verbose     = flag.Bool("verbose", false, "Enable verbose output (includes debug messages)")
	timeout     = flag.Duration("timeout", 30*time.Second, "Timeout for script execution")
	maxFixes    = flag.Int("max-fixes", 3, "Maximum number of attempts to fix the script")
	maxAttempts = flag.Int("max-attempts", 3, "Maximum number of attempts to generate a working script")
	llmProvider = flag.String("llm.provider", "", "LLM provider to use (overrides config)")
	llmModel    = flag.String("llm.model", "", "LLM model to use (overrides config)")
	extraPrompt = flag.String("prompt", "", "Additional prompt to provide to the LLM")
	noCache     = flag.Bool("no-cache", false, "Skip using the cache for script generation")
	printOnly   = flag.Bool("print", false, "Print the generated script without executing it")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <script-file>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s script.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --llm.provider=claude --timeout=10 script.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --write-config\n", os.Args[0])
	}
	flag.Parse()

	// Set up logging first
	if *verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if *writeConfig {
		cfg := config.DefaultConfig()
		if err := config.WriteConfig(cfg); err != nil {
			log.Fatal("Failed to write default config:", err)
		}
		fmt.Println("Default config written to ~/.config/llmscript/config.yaml")
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	log.Debug("Loaded config: provider=%s, has_claude_api_key=%v, claude_model=%s",
		cfg.LLM.Provider,
		cfg.LLM.Claude.APIKey != "",
		cfg.LLM.Claude.Model)

	// Override config with command line flags
	if *llmProvider != "" {
		cfg.LLM.Provider = *llmProvider
		log.Debug("Provider overridden by command line flag: %s", *llmProvider)
	}
	if *llmModel != "" {
		switch cfg.LLM.Provider {
		case "ollama":
			cfg.LLM.Ollama.Model = *llmModel
		case "claude":
			cfg.LLM.Claude.Model = *llmModel
		case "openai":
			cfg.LLM.OpenAI.Model = *llmModel
		}
	}
	if *timeout != 0 {
		cfg.Timeout = *timeout
	}
	if *maxFixes != 0 {
		cfg.MaxFixes = *maxFixes
	}
	if *maxAttempts != 0 {
		cfg.MaxAttempts = *maxAttempts
	}
	if *extraPrompt != "" {
		cfg.ExtraPrompt = *extraPrompt
	}

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	scriptFile := flag.Args()[0]
	if err := runScript(cfg, scriptFile); err != nil {
		log.Fatal("Failed to run script:", err)
	}
}

func runScript(cfg *config.Config, scriptFile string) error {
	log.Info("Reading script file: %s", scriptFile)
	content, err := os.ReadFile(scriptFile)
	if err != nil {
		return fmt.Errorf("failed to read script file: %w", err)
	}

	log.Info("Creating LLM provider: %s", cfg.LLM.Provider)
	provider, err := llm.NewProvider(cfg.LLM.Provider, map[string]interface{}{
		"ollama": map[string]interface{}{
			"model": cfg.LLM.Ollama.Model,
			"host":  cfg.LLM.Ollama.Host,
		},
		"claude": map[string]interface{}{
			"api_key": cfg.LLM.Claude.APIKey,
			"model":   cfg.LLM.Claude.Model,
		},
		"openai": map[string]interface{}{
			"api_key": cfg.LLM.OpenAI.APIKey,
			"model":   cfg.LLM.OpenAI.Model,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}

	log.Info("Creating work directory")
	workDir, err := os.MkdirTemp("", "llmscript-*")
	if err != nil {
		return fmt.Errorf("failed to create working directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(workDir); err != nil {
			log.Error("failed to remove working directory: %v", err)
		}
	}()
	if *verbose {
		log.Info("Work directory: %s", workDir)
	}

	log.Info("Creating pipeline")
	pipeline, err := script.NewPipeline(provider, cfg.MaxFixes, cfg.MaxAttempts, cfg.Timeout, workDir, *noCache)
	if err != nil {
		return fmt.Errorf("failed to create pipeline: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	log.Info("Generating and testing script")
	script, err := pipeline.GenerateAndTest(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to generate working script: %w", err)
	}

	if *verbose {
		log.Info("Generated script:\n%s", script)
	}

	// Write the script to a file
	scriptPath := filepath.Join(workDir, "script.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}

	// Clear the spinner line before printing success message
	log.GetSpinner().Clear()

	// If --print flag is set, just print the script and exit
	if *printOnly {
		fmt.Println(script)
		return nil
	}

	// Get any additional arguments after the script file
	scriptArgs := flag.Args()[1:]

	// Execute the script with any additional arguments
	cmd := exec.Command(scriptPath, scriptArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Run the command and exit with its status code
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		// If it's not an exit error, something else went wrong
		return fmt.Errorf("script execution failed: %w", err)
	}

	return nil
}
