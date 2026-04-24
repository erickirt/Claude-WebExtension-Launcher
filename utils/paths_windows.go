//go:build windows

package utils

import "path/filepath"

// ResolvePath resolves a path relative to the launcher's directory.
// Used for launcher-local files: node_modules, temp zips, asar-temp, etc.
func ResolvePath(relativePath string) string {
	return filepath.Join(execDir, relativePath)
}

// ResolveInstallPath resolves a path relative to the app install directory.
// On Windows, this is in WindowsApps.
func ResolveInstallPath(relativePath string) string {
	return filepath.Join(WindowsInstallDir, relativePath)
}
