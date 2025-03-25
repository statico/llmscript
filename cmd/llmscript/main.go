package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/statico/llmscript/internal/config"
	"github.com/statico/llmscript/internal/llm"
	"github.com/statico/llmscript/internal/log"
	"github.com/statico/llmscript/internal/script"
)

func main() {
	writeConfig := flag.Bool("write-config", false, "Write default config to ~/.config/llmscript/config.yaml")
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

	if len(flag.Args()) == 0 {
		log.Fatal("No script file provided")
	}

	scriptFile := flag.Args()[0]
	if err := runScript(cfg, scriptFile); err != nil {
		log.Fatal("Failed to run script:", err)
	}
}

func runScript(cfg *config.Config, scriptFile string) error {
	// Read script file
	content, err := os.ReadFile(scriptFile)
	if err != nil {
		return fmt.Errorf("failed to read script file: %w", err)
	}

	// Create LLM provider
	provider, err := llm.NewProvider(cfg.LLM.Provider, cfg.LLM)
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}

	// Create work directory
	workDir, err := os.MkdirTemp("", "llmscript-*")
	if err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}
	defer os.RemoveAll(workDir)

	// Create pipeline
	pipeline, err := script.NewPipeline(provider, cfg.MaxFixes, cfg.MaxAttempts, cfg.Timeout, workDir)
	if err != nil {
		return fmt.Errorf("failed to create pipeline: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Generate and test script
	script, err := pipeline.GenerateAndTest(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to generate working script: %w", err)
	}

	// Write script to output file
	outputFile := scriptFile + ".sh"
	if err := os.WriteFile(outputFile, []byte(script), 0755); err != nil {
		return fmt.Errorf("failed to write output script: %w", err)
	}

	// Execute the script
	cmd := exec.Command("bash", outputFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	return nil
}
