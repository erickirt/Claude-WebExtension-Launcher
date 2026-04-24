//go:build windows

package utils

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

var (
	shell32             = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteExW = shell32.NewProc("ShellExecuteExW")

	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")
	procGetExitCodeProcess  = kernel32.NewProc("GetExitCodeProcess")
	procCloseHandle         = kernel32.NewProc("CloseHandle")
)

type shellExecuteInfo struct {
	cbSize       uint32
	fMask        uint32
	hwnd         uintptr
	lpVerb       *uint16
	lpFile       *uint16
	lpParameters *uint16
	lpDirectory  *uint16
	nShow        int32
	hInstApp     uintptr
	lpIDList     uintptr
	lpClass      *uint16
	hkeyClass    uintptr
	dwHotKey     uint32
	hIcon        uintptr
	hProcess     uintptr
}

// IsAdmin checks if the current process is running with administrator privileges.
func IsAdmin() bool {
	f, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// RelaunchAsAdmin re-launches the current executable with elevated privileges
// via ShellExecuteEx "runas". Returns immediately after launching; the caller
// should os.Exit(0).
func RelaunchAsAdmin() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	args := strings.Join(os.Args[1:], " ")

	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exe)
	params, _ := syscall.UTF16PtrFromString(args)

	sei := shellExecuteInfo{
		lpVerb:       verb,
		lpFile:       file,
		lpParameters: params,
		nShow:        1, // SW_SHOWNORMAL
	}
	sei.cbSize = uint32(unsafe.Sizeof(sei))

	ret, _, err := procShellExecuteExW.Call(uintptr(unsafe.Pointer(&sei)))
	if ret == 0 {
		return fmt.Errorf("ShellExecuteEx failed: %v", err)
	}

	return nil
}

// RunElevatedAndWait launches the given executable elevated via UAC and waits
// for it to exit. Returns the process exit code.
func RunElevatedAndWait(exe string, args string) (int, error) {
	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exe)
	params, _ := syscall.UTF16PtrFromString(args)

	sei := shellExecuteInfo{
		fMask:        0x00000040, // SEE_MASK_NOCLOSEPROCESS
		lpVerb:       verb,
		lpFile:       file,
		lpParameters: params,
		nShow:        1, // SW_SHOWNORMAL
	}
	sei.cbSize = uint32(unsafe.Sizeof(sei))

	ret, _, err := procShellExecuteExW.Call(uintptr(unsafe.Pointer(&sei)))
	if ret == 0 {
		return -1, fmt.Errorf("ShellExecuteEx failed: %v", err)
	}

	// Wait for the elevated process to finish
	procWaitForSingleObject.Call(sei.hProcess, 0xFFFFFFFF) // INFINITE

	var exitCode uint32
	procGetExitCodeProcess.Call(sei.hProcess, uintptr(unsafe.Pointer(&exitCode)))
	procCloseHandle.Call(sei.hProcess)

	return int(exitCode), nil
}
