package fileops

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOperationType_String(t *testing.T) {
	tests := []struct {
		op       OperationType
		expected string
	}{
		{OpCopy, "Copy"},
		{OpMove, "Move"},
		{OpDelete, "Delete"},
		{OpRename, "Rename"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.op.String(); got != tt.expected {
				t.Errorf("OperationType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOperationStatus_String(t *testing.T) {
	tests := []struct {
		status   OperationStatus
		expected string
	}{
		{StatusPending, "Pending"},
		{StatusRunning, "Running"},
		{StatusCompleted, "Completed"},
		{StatusFailed, "Failed"},
		{StatusCancelled, "Cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("OperationStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewOperation(t *testing.T) {
	op := NewOperation(OpCopy, []string{"/src"}, "/dst")

	if op.ID == "" {
		t.Error("Operation ID should not be empty")
	}
	if op.Type != OpCopy {
		t.Errorf("Operation type = %v, want %v", op.Type, OpCopy)
	}
	if len(op.Source) != 1 || op.Source[0] != "/src" {
		t.Errorf("Operation source = %v, want [/src]", op.Source)
	}
	if op.Destination != "/dst" {
		t.Errorf("Operation destination = %v, want /dst", op.Destination)
	}
	if op.Status != StatusPending {
		t.Errorf("Operation status = %v, want %v", op.Status, StatusPending)
	}
	if op.Progress != 0.0 {
		t.Errorf("Operation progress = %v, want 0.0", op.Progress)
	}
}

func TestOperation_UpdateProgress(t *testing.T) {
	op := NewOperation(OpCopy, []string{"/src"}, "/dst")
	op.UpdateProgress(50, 100, "/src/file.txt")

	progress, bytesProcessed, bytesTotal, currentFile := op.GetProgress()

	if progress != 0.5 {
		t.Errorf("Progress = %v, want 0.5", progress)
	}
	if bytesProcessed != 50 {
		t.Errorf("BytesProcessed = %v, want 50", bytesProcessed)
	}
	if bytesTotal != 100 {
		t.Errorf("BytesTotal = %v, want 100", bytesTotal)
	}
	if currentFile != "/src/file.txt" {
		t.Errorf("CurrentFile = %v, want /src/file.txt", currentFile)
	}
}

func TestOperation_Cancel(t *testing.T) {
	op := NewOperation(OpCopy, []string{"/src"}, "/dst")
	op.SetStatus(StatusRunning)
	op.Cancel()

	if op.Status != StatusCancelled {
		t.Errorf("Status after cancel = %v, want %v", op.Status, StatusCancelled)
	}
	if !op.IsCancelled() {
		t.Error("IsCancelled() should return true after Cancel()")
	}
}

//nolint:gosec // Test file permissions are intentionally relaxed
func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.txt")
	content := "Hello, Warren!"
	if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	dstFile := filepath.Join(tmpDir, "destination.txt")
	op := Copy(srcFile, dstFile, nil)

	// Wait for operation to complete
	waitForOperation(t, op, 5*time.Second)

	// Verify operation completed
	if op.Status != StatusCompleted {
		t.Errorf("Operation status = %v, want %v", op.Status, StatusCompleted)
		if op.Error != nil {
			t.Errorf("Operation error: %v", op.Error)
		}
	}

	// Verify file was copied
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}
	if string(dstContent) != content {
		t.Errorf("Destination content = %q, want %q", string(dstContent), content)
	}

	// Verify permissions were preserved
	srcInfo, _ := os.Stat(srcFile)
	dstInfo, _ := os.Stat(dstFile)
	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("Destination mode = %v, want %v", dstInfo.Mode(), srcInfo.Mode())
	}
}

//nolint:gosec // Test file/directory permissions are intentionally relaxed
func TestCopyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source directory structure
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create files in source directory
	file1 := filepath.Join(srcDir, "file1.txt")
	file2 := filepath.Join(srcDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Create subdirectory
	subDir := filepath.Join(srcDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	file3 := filepath.Join(subDir, "file3.txt")
	if err := os.WriteFile(file3, []byte("content3"), 0644); err != nil {
		t.Fatalf("Failed to create file3: %v", err)
	}

	// Copy directory
	dstDir := filepath.Join(tmpDir, "dst")
	op := Copy(srcDir, dstDir, nil)

	// Wait for operation to complete
	waitForOperation(t, op, 5*time.Second)

	// Verify operation completed
	if op.Status != StatusCompleted {
		t.Errorf("Operation status = %v, want %v", op.Status, StatusCompleted)
		if op.Error != nil {
			t.Errorf("Operation error: %v", op.Error)
		}
	}

	// Verify all files were copied
	verifyFileExists(t, filepath.Join(dstDir, "file1.txt"), "content1")
	verifyFileExists(t, filepath.Join(dstDir, "file2.txt"), "content2")
	verifyFileExists(t, filepath.Join(dstDir, "subdir", "file3.txt"), "content3")
}

//nolint:gosec // Test file permissions are intentionally relaxed
func TestCopyMultiple(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Create destination directory
	dstDir := filepath.Join(tmpDir, "dst")
	if err := os.Mkdir(dstDir, 0755); err != nil {
		t.Fatalf("Failed to create destination: %v", err)
	}

	// Copy multiple files
	op := CopyMultiple([]string{file1, file2}, dstDir, nil)

	// Wait for operation to complete
	waitForOperation(t, op, 5*time.Second)

	// Verify operation completed
	if op.Status != StatusCompleted {
		t.Errorf("Operation status = %v, want %v", op.Status, StatusCompleted)
		if op.Error != nil {
			t.Errorf("Operation error: %v", op.Error)
		}
	}

	// Verify files were copied
	verifyFileExists(t, filepath.Join(dstDir, "file1.txt"), "content1")
	verifyFileExists(t, filepath.Join(dstDir, "file2.txt"), "content2")
}

//nolint:dupl,gosec // Test similarity acceptable, test files use relaxed permissions
func TestMove(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.txt")
	content := "Move test"
	if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Move file
	dstFile := filepath.Join(tmpDir, "destination.txt")
	op := Move(srcFile, dstFile, nil)

	// Wait for operation to complete
	waitForOperation(t, op, 5*time.Second)

	// Verify operation completed
	if op.Status != StatusCompleted {
		t.Errorf("Operation status = %v, want %v", op.Status, StatusCompleted)
		if op.Error != nil {
			t.Errorf("Operation error: %v", op.Error)
		}
	}

	// Verify file was moved
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("Source file should not exist after move")
	}

	dstContent, err := os.ReadFile(dstFile) //nolint:gosec // Test file
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}
	if string(dstContent) != content {
		t.Errorf("Destination content = %q, want %q", string(dstContent), content)
	}
}

//nolint:gosec // Test file permissions are intentionally relaxed
func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file to delete
	file := filepath.Join(tmpDir, "delete_me.txt")
	if err := os.WriteFile(file, []byte("delete"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Delete file
	op := Delete(file, nil)

	// Wait for operation to complete
	waitForOperation(t, op, 5*time.Second)

	// Verify operation completed
	if op.Status != StatusCompleted {
		t.Errorf("Operation status = %v, want %v", op.Status, StatusCompleted)
		if op.Error != nil {
			t.Errorf("Operation error: %v", op.Error)
		}
	}

	// Verify file was deleted
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Error("File should not exist after delete")
	}
}

//nolint:gosec // Test file permissions are intentionally relaxed
func TestDeleteMultiple(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files to delete
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("delete1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("delete2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Delete files
	op := DeleteMultiple([]string{file1, file2}, nil)

	// Wait for operation to complete
	waitForOperation(t, op, 5*time.Second)

	// Verify operation completed
	if op.Status != StatusCompleted {
		t.Errorf("Operation status = %v, want %v", op.Status, StatusCompleted)
		if op.Error != nil {
			t.Errorf("Operation error: %v", op.Error)
		}
	}

	// Verify files were deleted
	if _, err := os.Stat(file1); !os.IsNotExist(err) {
		t.Error("File1 should not exist after delete")
	}
	if _, err := os.Stat(file2); !os.IsNotExist(err) {
		t.Error("File2 should not exist after delete")
	}
}

//nolint:dupl // Test similarity is acceptable
func TestRename(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file to rename
	oldPath := filepath.Join(tmpDir, "old_name.txt")
	content := "rename test"
	if err := os.WriteFile(oldPath, []byte(content), 0644); err != nil { //nolint:gosec // Test file
		t.Fatalf("Failed to create file: %v", err)
	}

	// Rename file
	newPath := filepath.Join(tmpDir, "new_name.txt")
	op := Rename(oldPath, newPath, nil)

	// Wait for operation to complete
	waitForOperation(t, op, 5*time.Second)

	// Verify operation completed
	if op.Status != StatusCompleted {
		t.Errorf("Operation status = %v, want %v", op.Status, StatusCompleted)
		if op.Error != nil {
			t.Errorf("Operation error: %v", op.Error)
		}
	}

	// Verify file was renamed
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("Old path should not exist after rename")
	}

	newContent, err := os.ReadFile(newPath) //nolint:gosec // Test file
	if err != nil {
		t.Fatalf("Failed to read renamed file: %v", err)
	}
	if string(newContent) != content {
		t.Errorf("Renamed file content = %q, want %q", string(newContent), content)
	}
}

func TestCopyWithProgress(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a larger file to test progress tracking
	srcFile := filepath.Join(tmpDir, "large_source.txt")
	content := make([]byte, 100*1024) // 100KB
	for i := range content {
		content[i] = byte(i % 256)
	}
	if err := os.WriteFile(srcFile, content, 0644); err != nil { //nolint:gosec // Test file
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Track progress updates
	progressUpdates := 0
	callback := func(op *Operation) {
		progressUpdates++
		t.Logf("Progress: %.2f%% (%d/%d bytes)", op.Progress*100, op.BytesProcessed, op.BytesTotal)
	}

	// Copy file with progress tracking
	dstFile := filepath.Join(tmpDir, "large_dest.txt")
	op := Copy(srcFile, dstFile, callback)

	// Wait for operation to complete
	waitForOperation(t, op, 5*time.Second)

	// Verify operation completed
	if op.Status != StatusCompleted {
		t.Errorf("Operation status = %v, want %v", op.Status, StatusCompleted)
		if op.Error != nil {
			t.Errorf("Operation error: %v", op.Error)
		}
	}

	// Verify progress callback was called
	if progressUpdates == 0 {
		t.Error("Progress callback should have been called at least once")
	}

	// Verify final progress
	if op.Progress != 1.0 && op.Status == StatusCompleted {
		t.Errorf("Final progress = %v, want 1.0", op.Progress)
	}
}

//nolint:gosec // Test file permissions are intentionally relaxed
func TestOperationCancel(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a very large file to give us time to cancel
	srcFile := filepath.Join(tmpDir, "large_source.txt")
	content := make([]byte, 100*1024*1024) // 100MB
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	dstFile := filepath.Join(tmpDir, "large_dest.txt")
	op := Copy(srcFile, dstFile, nil)

	// Cancel very quickly after starting
	time.Sleep(1 * time.Millisecond)
	op.Cancel()

	// Wait for operation to stop
	waitForOperation(t, op, 5*time.Second)

	// Verify operation was cancelled
	// Note: It's possible the operation completes before cancellation on very fast systems
	// so we check if it's either cancelled or completed very quickly
	switch op.Status {
	case StatusCancelled:
		if !op.IsCancelled() {
			t.Error("IsCancelled() should return true when status is Cancelled")
		}
	case StatusCompleted:
		t.Logf("Operation completed before cancellation could take effect (very fast system)")
	default:
		t.Errorf("Operation status = %v, want Cancelled or Completed", op.Status)
	}
}

//nolint:gosec // Test file permissions are intentionally relaxed
func TestCalculateSize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	content1 := "Hello"  // 5 bytes
	content2 := "World!" // 6 bytes
	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Calculate size of directory
	size, err := calculateSize(tmpDir)
	if err != nil {
		t.Fatalf("calculateSize failed: %v", err)
	}

	expectedSize := int64(len(content1) + len(content2))
	if size != expectedSize {
		t.Errorf("Size = %d, want %d", size, expectedSize)
	}
}

// Helper functions

//nolint:unparam // timeout parameter useful for test flexibility
func waitForOperation(t *testing.T, op *Operation, timeout time.Duration) {
	t.Helper()
	start := time.Now()
	for {
		if time.Since(start) > timeout {
			t.Fatalf("Operation timed out after %v", timeout)
		}

		status := op.Status
		if status == StatusCompleted || status == StatusFailed || status == StatusCancelled {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func verifyFileExists(t *testing.T, path string, expectedContent string) {
	t.Helper()

	content, err := os.ReadFile(path) //nolint:gosec // Test helper
	if err != nil {
		t.Errorf("Failed to read file %s: %v", path, err)
		return
	}

	if string(content) != expectedContent {
		t.Errorf("File %s content = %q, want %q", path, string(content), expectedContent)
	}
}
