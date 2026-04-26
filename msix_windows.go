//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func checkMSIXAndPrompt(instanceName string) {
	if !isMSIXInstalled() {
		return
	}

	choice := loadMSIXChoice()

	if strings.HasPrefix(choice, "keep:") {
		savedVersion := strings.TrimPrefix(choice, "keep:")
		if savedVersion == Version {
			return
		}
		fmt.Println("Launcher version changed since you last chose to keep the official Claude app.")
	}

	if choice == "uninstall" {
		fmt.Println("Official Claude MSIX was reinstalled since you last removed it.")
	}

	switch promptMSIXChoice() {
	case "uninstall":
		if err := uninstallMSIX(); err != nil {
			fmt.Printf("Failed to uninstall MSIX: %v\n", err)
			fmt.Println("Continuing without removing it. You can try again next launch.")
			return
		}
		saveMSIXChoice("uninstall")
		fmt.Println("Official Claude app removed successfully.")
	case "keep":
		saveMSIXChoice("keep:" + Version)
		fmt.Println("Keeping official Claude app. Magic link login will not work with the patched app.")
	default:
		fmt.Println("Skipping for now. You'll be asked again next launch.")
	}
}

func isMSIXInstalled() bool {
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		"Get-AppxPackage -Name 'Claude' -ErrorAction SilentlyContinue").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

func promptMSIXChoice() string {
	fmt.Println()
	fmt.Println("============================================================")
	fmt.Println("Official Claude Desktop (MSIX) detected.")
	fmt.Println()
	fmt.Println("The official MSIX installation overrides the claude:// protocol")
	fmt.Println("handler, which prevents magic link login from working with the")
	fmt.Println("patched app.")
	fmt.Println()
	fmt.Println("[1] Uninstall the official app (recommended)")
	fmt.Println("[2] Keep it installed (login via magic link won't work)")
	fmt.Println("[3] Ask me later")
	fmt.Println("============================================================")
	fmt.Print("Choose [1/2/3]: ")

	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)

	switch input {
	case "1":
		return "uninstall"
	case "2":
		return "keep"
	default:
		return "ask-later"
	}
}

func uninstallMSIX() error {
	fmt.Println("Stopping official Claude process...")
	exec.Command("powershell", "-NoProfile", "-Command",
		"Get-Process -Name 'Claude' -ErrorAction SilentlyContinue | "+
			"Where-Object { $_.Path -and $_.Path -notlike '*ClaudeWebExtLauncher*' } | "+
			"Stop-Process -Force").Run()

	fmt.Println("Removing MSIX package...")
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		"Get-AppxPackage -Name 'Claude' | Remove-AppxPackage").CombinedOutput()
	if err != nil {
		return fmt.Errorf("Remove-AppxPackage failed: %v\n%s", err, out)
	}

	if isMSIXInstalled() {
		return fmt.Errorf("package still present after removal")
	}

	return nil
}

func msixChoicePath() string {
	return filepath.Join(os.Getenv("APPDATA"), "ClaudeWebExtLauncher", "msix-choice.txt")
}

func loadMSIXChoice() string {
	data, err := os.ReadFile(msixChoicePath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func saveMSIXChoice(choice string) {
	p := msixChoicePath()
	os.MkdirAll(filepath.Dir(p), 0755)
	os.WriteFile(p, []byte(choice), 0644)
}
