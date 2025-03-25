package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/statico/llmscript/internal/config"
	"github.com/statico/llmscript/internal/llm"
	"github.com/statico/llmscript/internal/log"
	"github.com/statico/llmscript/internal/script"
)

var (
	writeConfig  = flag.Bool("write-config", false, "Write default config to ~/.config/llmscript/config.yaml")
	verbose      = flag.Bool("verbose", false, "Enable verbose output")
	debug        = flag.Bool("debug", false, "Enable debug output")
	showProgress = flag.Bool("progress", true, "Show progress indicators")
	timeout      = flag.Duration("timeout", 30*time.Second, "Timeout for script execution")
	maxFixes     = flag.Int("max-fixes", 10, "Maximum number of attempts to fix the script")
	maxAttempts  = flag.Int("max-attempts", 3, "Maximum number of attempts to generate a working script")
	llmProvider  = flag.String("llm.provider", "", "LLM provider to use (overrides config)")
	llmModel     = flag.String("llm.model", "", "LLM model to use (overrides config)")
	extraPrompt  = flag.String("prompt", "", "Additional prompt to provide to the LLM")
	noCache      = flag.Bool("no-cache", false, "Skip using the cache for script generation")
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

	// Override config with command line flags
	if *llmProvider != "" {
		cfg.LLM.Provider = *llmProvider
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

	// Set up logging
	if *debug {
		log.SetLevel(log.DebugLevel)
		*verbose = true // Enable verbose mode when debug is enabled
	} else if *verbose {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
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
	provider, err := llm.NewProvider(cfg.LLM.Provider, cfg.LLM)
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}

	log.Info("Creating work directory")
	workDir, err := os.MkdirTemp("", "llmscript-*")
	if err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}
	defer os.RemoveAll(workDir)
	if *verbose {
		log.Info("Work directory: %s", workDir)
	}

	log.Info("Creating pipeline")
	pipeline, err := script.NewPipeline(provider, cfg.MaxFixes, cfg.MaxAttempts, cfg.Timeout, workDir, *showProgress, *noCache)
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

	log.Success("Script generated successfully!")
	return nil
}
