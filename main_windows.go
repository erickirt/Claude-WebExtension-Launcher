//go:build windows

package main

import (
	"claude-webext-patcher/patcher"
	"claude-webext-patcher/utils"
	"fmt"
	"os"
	"path/filepath"
)

// prepareAdminContext self-elevates if needed, takes WindowsApps ownership,
// and cleans up old installation files from the launcher directory.
func prepareAdminContext() error {
	// On Windows, ensure we're running as admin (needed for WindowsApps folder setup).
	// We use programmatic elevation instead of a manifest so the self-update flow works.
	if !utils.IsAdmin() {
		fmt.Println("Requesting administrator privileges...")
		if err := utils.RelaunchAsAdmin(); err != nil {
			return fmt.Errorf("failed to elevate: %v", err)
		}
		os.Exit(0)
	}

	// Take ownership of WindowsApps early so all subsequent operations can access it
	if err := patcher.TakeWindowsAppsOwnership(); err != nil {
		fmt.Printf("Warning: failed to take WindowsApps ownership: %v\n", err)
	}

	// Clean up old installation files next to the executable (from before the move to WindowsApps)
	execDir := utils.GetExecutableDir()
	for _, oldDir := range []string{"app-latest", "web-extensions"} {
		oldPath := filepath.Join(execDir, oldDir)
		if _, err := os.Stat(oldPath); err == nil {
			fmt.Printf("Removing old %s from launcher directory...\n", oldDir)
			if err := os.RemoveAll(oldPath); err != nil {
				fmt.Printf("Warning: could not remove %s: %v\n", oldPath, err)
			}
		}
	}

	return nil
}

// releaseAdminContext restores TrustedInstaller ownership of WindowsApps
// before the unelevated launch of Claude.
func releaseAdminContext() {
	patcher.ReleaseWindowsAppsOwnership()
}

func claudeUserDataDir(instance string) string {
	return filepath.Join(os.Getenv("APPDATA"), "Claude-"+instance)
}

func claudeExecutablePath() string {
	return filepath.Join(patcher.AppFolder, "claude.exe")
}
