package path

import (
	"fmt"
	"os"
	"path/filepath"
)

const EnvVar = "CHESS_ENGINE"

// DefaultRelative is the engine path relative to the tui/ module directory.
const DefaultRelative = "../target/release/chess"

// Resolve returns the engine binary path from CHESS_ENGINE or the default
// relative to baseDir (typically the tui/ module root).
func Resolve(baseDir string) string {
	if v := os.Getenv(EnvVar); v != "" {
		return v
	}
	if baseDir == "" {
		return DefaultRelative
	}
	return filepath.Clean(filepath.Join(baseDir, DefaultRelative))
}

// Validate checks that path exists and is a regular file.
func Validate(enginePath string) error {
	info, err := os.Stat(enginePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("engine binary not found: %s", enginePath)
		}
		return fmt.Errorf("engine binary not accessible: %s: %w", enginePath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("engine path is a directory: %s", enginePath)
	}
	return nil
}
