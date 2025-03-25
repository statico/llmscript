package main

import (
	"flag"
	"fmt"

	"github.com/statico/llmscript/internal/config"
	"github.com/statico/llmscript/internal/log"
)

func main() {
	configPath := flag.String("config", "", "Path to config file")
	writeConfig := flag.Bool("write-config", false, "Write default config to ~/.config/llmscript/config.yaml")
	flag.Parse()

	if *writeConfig {
		if err := config.WriteDefaultConfig(); err != nil {
			log.Fatal("Failed to write default config:", err)
		}
		fmt.Println("Default config written to ~/.config/llmscript/config.yaml")
		return
	}

	cfg, err := config.Load(*configPath)
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
	// TODO: Implement script execution
	return nil
}
