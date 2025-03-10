package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from .env files
func LoadEnv() error {
	// Order of precedence:
	// 1. .env.{environment}.local - Local overrides of environment-specific settings
	// 2. .env.local - Local overrides
	// 3. .env.{environment} - Environment-specific settings
	// 4. .env - Default settings

	// Get environment
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development" // Default to development if not specified
	}

	// Files to load in order
	files := []string{
		".env",                   // Default settings
		".env." + env,            // Environment-specific settings
		".env.local",             // Local overrides
		".env." + env + ".local", // Local overrides of environment-specific settings
	}

	// Get the current working directory
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Try to load each .env file in order - silently continue if any file doesn't exist
	for _, file := range files {
		filePath := filepath.Join(workDir, file)

		// Check if file exists before loading
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue
		}

		if err := godotenv.Load(filePath); err != nil {
			return fmt.Errorf("error loading %s: %w", file, err)
		}
	}

	return nil
}
