package script

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ShellConfig holds shell-specific configuration
type ShellConfig struct {
	Path    string
	Args    []string
	IsLogin bool
}

// DetectShell determines the user's shell and returns its configuration
func DetectShell() (*ShellConfig, error) {
	// Get the current user's shell from the environment or default
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		if runtime.GOOS == "windows" {
			// On Windows, default to PowerShell if available, otherwise cmd.exe
			if path, err := exec.LookPath("powershell.exe"); err == nil {
				return &ShellConfig{
					Path: path,
					Args: []string{"-NoProfile", "-NonInteractive", "-Command"},
				}, nil
			}
			return &ShellConfig{
				Path: "cmd.exe",
				Args: []string{"/C"},
			}, nil
		}
		// On Unix-like systems, default to /bin/sh
		shellPath = "/bin/sh"
	}

	// Get the shell name from the path
	shellName := filepath.Base(shellPath)

	// Configure shell-specific arguments
	config := &ShellConfig{Path: shellPath}
	switch strings.ToLower(shellName) {
	case "bash":
		config.Args = []string{"--noprofile", "--norc", "-e", "-o", "pipefail", "-c"}
	case "zsh":
		config.Args = []string{"--no-rcs", "-e", "-o", "pipefail", "-c"}
	case "fish":
		config.Args = []string{"--no-config", "--command"}
	case "powershell.exe", "pwsh", "pwsh.exe":
		config.Args = []string{"-NoProfile", "-NonInteractive", "-Command"}
	default:
		// For unknown shells, use a simple -c argument
		config.Args = []string{"-c"}
	}

	return config, nil
}

// ValidateScript performs basic security checks on a script
func ValidateScript(script string) error {
	// Check for suspicious commands or patterns
	suspiciousPatterns := []string{
		"rm -rf /*",
		"mkfs",
		"> /dev/sd",
		"dd if=/dev/zero",
		":(){:|:&};:",
		"wget",
		"curl.*| *sh",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(strings.ToLower(script), strings.ToLower(pattern)) {
			return fmt.Errorf("script contains potentially dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// PrepareScriptEnvironment sets up a secure environment for script execution
func PrepareScriptEnvironment(workDir string) (string, error) {
	// Create a new temporary directory for script execution
	tmpDir, err := os.MkdirTemp(workDir, "script-")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Set restrictive permissions
	if err := os.Chmod(tmpDir, 0750); err != nil {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Printf("failed to remove temp dir: %v", err)
		}
		return "", fmt.Errorf("failed to set directory permissions: %w", err)
	}

	return tmpDir, nil
}
