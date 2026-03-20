package internal

import (
	"fmt"

	"github.com/joho/godotenv"
)

// Default .env files loaded automatically in order (later overrides earlier).
var defaultEnvFiles = []string{".env", ".env.local"}

// LoadDotenvFiles loads environment variables from .env files.
// It first loads the default files (.env, .env.local) if they exist,
// then loads any explicitly specified files from the config.
// Explicit files must exist or an error is returned.
// Later files override earlier ones.
func LoadDotenvFiles(explicit []string, fs FileSystem) error {
	for _, f := range defaultEnvFiles {
		if fs.FileExists(f) {
			if err := godotenv.Overload(f); err != nil {
				return fmt.Errorf("failed to load %s: %w", f, err)
			}
		}
	}

	for _, f := range explicit {
		if !fs.FileExists(f) {
			return fmt.Errorf("env_file %q not found", f)
		}
		if err := godotenv.Overload(f); err != nil {
			return fmt.Errorf("failed to load %s: %w", f, err)
		}
	}

	return nil
}
