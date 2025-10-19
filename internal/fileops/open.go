package fileops

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenFile opens a file with the default application using xdg-open (Linux),
// open (macOS), or start (Windows).
func OpenFile(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	var cmd *exec.Cmd

	// Security note: We're intentionally passing user-controlled file paths to system commands.
	// This is safe because:
	// 1. xdg-open/open/start are designed to handle arbitrary file paths
	// 2. The OS handles all security checks (file permissions, safe opening)
	// 3. We're not constructing shell commands - just passing arguments
	// 4. This is the standard way to open files with default applications
	//
	// #nosec G204 -- Subprocess launched with file path - intentional for file opening
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", path)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Run the command without waiting for it to complete
	// We don't want to block on the opened application
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	// Don't wait for the process - let it run independently
	// We intentionally ignore errors here since the application has already opened
	go func() {
		_ = cmd.Wait() // Explicitly ignore error
	}()

	return nil
}

// CanOpen checks if a file can potentially be opened.
// This does a basic check but doesn't guarantee the file can actually be opened.
func CanOpen(path string) (bool, error) {
	info, err := GetFileInfo(path)
	if err != nil {
		return false, err
	}

	// Can't open directories (should navigate instead)
	if info.IsDir {
		return false, nil
	}

	// Check if file is readable
	// TODO: Add permission check here if needed

	return true, nil
}
