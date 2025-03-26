package script

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/statico/llmscript/internal/llm"
	"github.com/statico/llmscript/internal/log"
)

// Cache handles caching of successful scripts and their test plans
type Cache struct {
	dir string
}

// NewCache creates a new cache instance
func NewCache() (*Cache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".config", "llmscript", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{dir: cacheDir}, nil
}

// Get retrieves a cached script pair
func (c *Cache) Get(description string) (llm.ScriptPair, error) {
	hash := c.hashDescription(description)
	scriptPath := filepath.Join(c.dir, hash+".json")

	log.Debug("Checking cache for script with hash: %s", hash)
	log.Debug("Script path: %s", scriptPath)

	// Check if file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		log.Debug("Script not found in cache")
		return llm.ScriptPair{}, nil
	}

	// Read script pair
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		return llm.ScriptPair{}, fmt.Errorf("failed to read cached script: %w", err)
	}

	var scripts llm.ScriptPair
	if err := json.Unmarshal(data, &scripts); err != nil {
		return llm.ScriptPair{}, fmt.Errorf("failed to parse cached scripts: %w", err)
	}
	log.Debug("Found cached scripts")

	return scripts, nil
}

// Set stores a successful script pair
func (c *Cache) Set(description string, scripts llm.ScriptPair) error {
	hash := c.hashDescription(description)
	scriptPath := filepath.Join(c.dir, hash+".json")

	log.Debug("Caching script with hash: %s", hash)
	log.Debug("Script path: %s", scriptPath)

	// Write script pair
	data, err := json.Marshal(scripts)
	if err != nil {
		return fmt.Errorf("failed to marshal scripts: %w", err)
	}
	if err := os.WriteFile(scriptPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cached script: %w", err)
	}
	log.Debug("Scripts cached successfully")

	return nil
}

// hashDescription generates a SHA-256 hash of the script description
func (c *Cache) hashDescription(description string) string {
	hash := sha256.Sum256([]byte(strings.TrimSpace(description)))
	return hex.EncodeToString(hash[:])
}
