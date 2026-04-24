//go:build !windows

package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

// ResolvePath resolves a path relative to the launcher's directory.
// On macOS, uses the Application Support directory instead of the bundle.
// Other non-Windows platforms fall back to the launcher's executable directory.
func ResolvePath(relativePath string) string {
	if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		dataDir := filepath.Join(home, "Library", "Application Support", "Claude WebExtension Launcher")
		os.MkdirAll(dataDir, 0755)
		return filepath.Join(dataDir, relativePath)
	}
	return filepath.Join(execDir, relativePath)
}

// ResolveInstallPath resolves a path relative to the app install directory.
// On non-Windows, this is the same as ResolvePath.
func ResolveInstallPath(relativePath string) string {
	return ResolvePath(relativePath)
}
