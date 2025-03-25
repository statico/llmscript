package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

func Run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: llmscript <script-file>")
	}

	scriptPath := os.Args[1]
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script file not found: %s", scriptPath)
	}

	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read script: %w", err)
	}

	// TODO: Implement script parsing and execution
	fmt.Printf("Reading script from: %s (%d bytes)\n", filepath.Base(scriptPath), len(content))
	return nil
}
