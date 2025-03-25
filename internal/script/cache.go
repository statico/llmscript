package script

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Cache handles caching of successful scripts and their test plans
type Cache struct {
	dir string
}

// NewCache creates a new cache instance
func NewCache() (*Cache, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config dir: %w", err)
	}

	cacheDir := filepath.Join(configDir, "llmscript", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{dir: cacheDir}, nil
}

// Get retrieves a cached script and its test plan
func (c *Cache) Get(description string) (string, []Test, error) {
	hash := c.hashDescription(description)
	scriptPath := filepath.Join(c.dir, hash+".sh")
	testsPath := filepath.Join(c.dir, hash+".tests.json")

	// Check if both files exist
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return "", nil, nil
	}
	if _, err := os.Stat(testsPath); os.IsNotExist(err) {
		return "", nil, nil
	}

	// Read script
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read cached script: %w", err)
	}

	// Read tests
	testsData, err := os.ReadFile(testsPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read cached tests: %w", err)
	}

	var tests []Test
	if err := json.Unmarshal(testsData, &tests); err != nil {
		return "", nil, fmt.Errorf("failed to parse cached tests: %w", err)
	}

	return string(script), tests, nil
}

// Set stores a successful script and its test plan
func (c *Cache) Set(description string, script string, tests []Test) error {
	hash := c.hashDescription(description)
	scriptPath := filepath.Join(c.dir, hash+".sh")
	testsPath := filepath.Join(c.dir, hash+".tests.json")

	// Write script
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return fmt.Errorf("failed to write cached script: %w", err)
	}

	// Write tests
	testsData, err := json.Marshal(tests)
	if err != nil {
		return fmt.Errorf("failed to marshal tests: %w", err)
	}
	if err := os.WriteFile(testsPath, testsData, 0644); err != nil {
		return fmt.Errorf("failed to write cached tests: %w", err)
	}

	return nil
}

// hashDescription generates a SHA-256 hash of the script description
func (c *Cache) hashDescription(description string) string {
	hash := sha256.Sum256([]byte(description))
	return hex.EncodeToString(hash[:])
}
