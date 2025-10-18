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
	go func() {
		cmd.Wait()
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
