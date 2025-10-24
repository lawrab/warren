// Package fileops provides file operation functions (copy, move, delete, etc.)
package fileops

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// OperationType represents the type of file operation.
type OperationType int

const (
	// OpCopy represents a copy operation
	OpCopy OperationType = iota
	// OpMove represents a move operation
	OpMove
	// OpDelete represents a delete operation
	OpDelete
	// OpRename represents a rename operation
	OpRename
)

// String returns a human-readable name for the operation type.
func (o OperationType) String() string {
	switch o {
	case OpCopy:
		return "Copy"
	case OpMove:
		return "Move"
	case OpDelete:
		return "Delete"
	case OpRename:
		return "Rename"
	default:
		return "Unknown"
	}
}

// OperationStatus represents the current status of an operation.
type OperationStatus int

const (
	// StatusPending means the operation is queued but not started
	StatusPending OperationStatus = iota
	// StatusRunning means the operation is currently executing
	StatusRunning
	// StatusCompleted means the operation finished successfully
	StatusCompleted
	// StatusFailed means the operation failed with an error
	StatusFailed
	// StatusCancelled means the operation was cancelled by user
	StatusCancelled
)

// String returns a human-readable name for the status.
func (s OperationStatus) String() string {
	switch s {
	case StatusPending:
		return "Pending"
	case StatusRunning:
		return "Running"
	case StatusCompleted:
		return "Completed"
	case StatusFailed:
		return "Failed"
	case StatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// Operation represents a file operation with progress tracking.
type Operation struct {
	// ID is a unique identifier for this operation
	ID string

	// Type is the operation type (copy, move, delete, rename)
	Type OperationType

	// Source is the source path(s) for the operation
	Source []string

	// Destination is the destination path (not used for delete)
	Destination string

	// Status is the current operation status
	Status OperationStatus

	// Progress is the current progress (0.0 to 1.0)
	Progress float64

	// BytesProcessed is the number of bytes processed so far
	BytesProcessed int64

	// BytesTotal is the total number of bytes to process
	BytesTotal int64

	// CurrentFile is the file currently being processed
	CurrentFile string

	// Error holds any error that occurred
	Error error

	// StartTime is when the operation started
	StartTime time.Time

	// EndTime is when the operation completed
	EndTime time.Time

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Mutex for thread-safe updates
	mu sync.RWMutex
}

// ProgressCallback is called when operation progress updates.
type ProgressCallback func(op *Operation)

// NewOperation creates a new operation with a unique ID.
func NewOperation(opType OperationType, source []string, destination string) *Operation {
	ctx, cancel := context.WithCancel(context.Background())
	return &Operation{
		ID:          generateOperationID(),
		Type:        opType,
		Source:      source,
		Destination: destination,
		Status:      StatusPending,
		Progress:    0.0,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Cancel cancels the operation.
func (op *Operation) Cancel() {
	op.mu.Lock()
	defer op.mu.Unlock()
	if op.Status == StatusRunning || op.Status == StatusPending {
		op.Status = StatusCancelled
		op.cancel()
	}
}

// IsCancelled returns true if the operation was cancelled.
func (op *Operation) IsCancelled() bool {
	select {
	case <-op.ctx.Done():
		return true
	default:
		return false
	}
}

// UpdateProgress updates the operation's progress.
func (op *Operation) UpdateProgress(bytesProcessed, bytesTotal int64, currentFile string) {
	op.mu.Lock()
	defer op.mu.Unlock()

	op.BytesProcessed = bytesProcessed
	op.BytesTotal = bytesTotal
	op.CurrentFile = currentFile

	if bytesTotal > 0 {
		op.Progress = float64(bytesProcessed) / float64(bytesTotal)
	}
}

// SetStatus sets the operation status.
func (op *Operation) SetStatus(status OperationStatus) {
	op.mu.Lock()
	defer op.mu.Unlock()
	op.Status = status

	if status == StatusRunning && op.StartTime.IsZero() {
		op.StartTime = time.Now()
	}
	if status == StatusCompleted || status == StatusFailed || status == StatusCancelled {
		op.EndTime = time.Now()
	}
}

// SetError sets an error and marks the operation as failed.
func (op *Operation) SetError(err error) {
	op.mu.Lock()
	defer op.mu.Unlock()
	op.Error = err
	op.Status = StatusFailed
	op.EndTime = time.Now()
}

// GetProgress returns the current progress information (thread-safe).
func (op *Operation) GetProgress() (float64, int64, int64, string) {
	op.mu.RLock()
	defer op.mu.RUnlock()
	return op.Progress, op.BytesProcessed, op.BytesTotal, op.CurrentFile
}

// Copy performs a copy operation from source to destination.
// It supports copying files and directories recursively.
func Copy(source string, destination string, callback ProgressCallback) *Operation {
	op := NewOperation(OpCopy, []string{source}, destination)
	go performCopy(op, source, destination, callback)
	return op
}

// CopyMultiple copies multiple files/directories to a destination directory.
func CopyMultiple(sources []string, destination string, callback ProgressCallback) *Operation {
	op := NewOperation(OpCopy, sources, destination)
	go performCopyMultiple(op, sources, destination, callback)
	return op
}

// Move performs a move operation from source to destination.
func Move(source string, destination string, callback ProgressCallback) *Operation {
	op := NewOperation(OpMove, []string{source}, destination)
	go performMove(op, source, destination, callback)
	return op
}

// MoveMultiple moves multiple files/directories to a destination directory.
func MoveMultiple(sources []string, destination string, callback ProgressCallback) *Operation {
	op := NewOperation(OpMove, sources, destination)
	go performMoveMultiple(op, sources, destination, callback)
	return op
}

// Delete performs a delete operation on the given path.
func Delete(path string, callback ProgressCallback) *Operation {
	op := NewOperation(OpDelete, []string{path}, "")
	go performDelete(op, path, callback)
	return op
}

// DeleteMultiple deletes multiple files/directories.
func DeleteMultiple(paths []string, callback ProgressCallback) *Operation {
	op := NewOperation(OpDelete, paths, "")
	go performDeleteMultiple(op, paths, callback)
	return op
}

// Rename performs a rename operation.
func Rename(oldPath string, newPath string, callback ProgressCallback) *Operation {
	op := NewOperation(OpRename, []string{oldPath}, newPath)
	go performRename(op, oldPath, newPath, callback)
	return op
}

// performCopy executes the copy operation.
func performCopy(op *Operation, source, destination string, callback ProgressCallback) {
	op.SetStatus(StatusRunning)

	// Calculate total size
	totalSize, err := calculateSize(source)
	if err != nil {
		op.SetError(fmt.Errorf("failed to calculate size: %w", err))
		if callback != nil {
			callback(op)
		}
		return
	}

	op.UpdateProgress(0, totalSize, source)
	if callback != nil {
		callback(op)
	}

	// Perform the copy
	var bytesProcessed int64
	err = copyRecursive(op, source, destination, &bytesProcessed, totalSize, callback)
	if err != nil {
		if !op.IsCancelled() {
			op.SetError(err)
		}
	} else {
		op.SetStatus(StatusCompleted)
	}

	if callback != nil {
		callback(op)
	}
}

// performCopyMultiple executes copy operation for multiple sources.
func performCopyMultiple(op *Operation, sources []string, destination string, callback ProgressCallback) {
	op.SetStatus(StatusRunning)

	// Calculate total size
	var totalSize int64
	for _, src := range sources {
		size, err := calculateSize(src)
		if err != nil {
			op.SetError(fmt.Errorf("failed to calculate size for %s: %w", src, err))
			if callback != nil {
				callback(op)
			}
			return
		}
		totalSize += size
	}

	var bytesProcessed int64
	for _, src := range sources {
		if op.IsCancelled() {
			break
		}

		// Determine destination path
		destPath := filepath.Join(destination, filepath.Base(src))

		err := copyRecursive(op, src, destPath, &bytesProcessed, totalSize, callback)
		if err != nil {
			if !op.IsCancelled() {
				op.SetError(fmt.Errorf("failed to copy %s: %w", src, err))
			}
			if callback != nil {
				callback(op)
			}
			return
		}
	}

	if !op.IsCancelled() {
		op.SetStatus(StatusCompleted)
	}

	if callback != nil {
		callback(op)
	}
}

// copyRecursive recursively copies files and directories.
func copyRecursive(op *Operation, src, dst string, bytesProcessed *int64, totalSize int64, callback ProgressCallback) error {
	if op.IsCancelled() {
		return fmt.Errorf("operation cancelled")
	}

	srcInfo, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	// Handle symlinks
	if srcInfo.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return fmt.Errorf("failed to read symlink: %w", err)
		}
		return os.Symlink(target, dst)
	}

	// Handle directories
	if srcInfo.IsDir() {
		return copyDir(op, src, dst, bytesProcessed, totalSize, callback)
	}

	// Handle regular files
	return copyFile(op, src, dst, bytesProcessed, totalSize, callback)
}

// copyDir copies a directory recursively.
func copyDir(op *Operation, src, dst string, bytesProcessed *int64, totalSize int64, callback ProgressCallback) error {
	// Create destination directory
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Read directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		if op.IsCancelled() {
			return fmt.Errorf("operation cancelled")
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if err := copyRecursive(op, srcPath, dstPath, bytesProcessed, totalSize, callback); err != nil {
			return err
		}
	}

	return nil
}

// copyFile copies a single file with progress tracking.
func copyFile(op *Operation, src, dst string, bytesProcessed *int64, totalSize int64, callback ProgressCallback) error {
	srcFile, err := os.Open(src) // #nosec G304 - file path from user operation
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if err := srcFile.Close(); err != nil {
			// Log error but don't fail the operation
			_ = err
		}
	}()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	dstFile, err := os.Create(dst) // #nosec G304 - file path from user operation
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		if err := dstFile.Close(); err != nil {
			// Log error but don't fail the operation
			_ = err
		}
	}()

	// Set permissions
	if err := dstFile.Chmod(srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Copy with progress tracking
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		if op.IsCancelled() {
			return fmt.Errorf("operation cancelled")
		}

		n, err := srcFile.Read(buf)
		if n > 0 {
			if _, writeErr := dstFile.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("failed to write to destination: %w", writeErr)
			}

			*bytesProcessed += int64(n)
			op.UpdateProgress(*bytesProcessed, totalSize, src)

			if callback != nil {
				callback(op)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read from source: %w", err)
		}
	}

	return nil
}

// performMove executes the move operation.
func performMove(op *Operation, source, destination string, callback ProgressCallback) {
	op.SetStatus(StatusRunning)

	// Try atomic rename first (same filesystem)
	err := os.Rename(source, destination)
	if err == nil {
		op.UpdateProgress(1, 1, source)
		op.SetStatus(StatusCompleted)
		if callback != nil {
			callback(op)
		}
		return
	}

	// Fall back to copy + delete
	totalSize, err := calculateSize(source)
	if err != nil {
		op.SetError(fmt.Errorf("failed to calculate size: %w", err))
		if callback != nil {
			callback(op)
		}
		return
	}

	var bytesProcessed int64
	err = copyRecursive(op, source, destination, &bytesProcessed, totalSize, callback)
	if err != nil {
		op.SetError(fmt.Errorf("failed to copy: %w", err))
		if callback != nil {
			callback(op)
		}
		return
	}

	// Delete source after successful copy
	err = os.RemoveAll(source)
	if err != nil {
		op.SetError(fmt.Errorf("failed to remove source after copy: %w", err))
		if callback != nil {
			callback(op)
		}
		return
	}

	op.SetStatus(StatusCompleted)
	if callback != nil {
		callback(op)
	}
}

// performMoveMultiple executes move operation for multiple sources.
func performMoveMultiple(op *Operation, sources []string, destination string, callback ProgressCallback) {
	op.SetStatus(StatusRunning)

	for _, src := range sources {
		if op.IsCancelled() {
			break
		}

		destPath := filepath.Join(destination, filepath.Base(src))

		// Try atomic rename first
		err := os.Rename(src, destPath)
		if err == nil {
			continue
		}

		// Fall back to copy + delete
		var totalSize int64
		totalSize, err = calculateSize(src)
		if err != nil {
			op.SetError(fmt.Errorf("failed to calculate size for %s: %w", src, err))
			if callback != nil {
				callback(op)
			}
			return
		}

		var bytesProcessed int64
		err = copyRecursive(op, src, destPath, &bytesProcessed, totalSize, callback)
		if err != nil {
			op.SetError(fmt.Errorf("failed to move %s: %w", src, err))
			if callback != nil {
				callback(op)
			}
			return
		}

		err = os.RemoveAll(src)
		if err != nil {
			op.SetError(fmt.Errorf("failed to remove source %s: %w", src, err))
			if callback != nil {
				callback(op)
			}
			return
		}
	}

	if !op.IsCancelled() {
		op.SetStatus(StatusCompleted)
	}

	if callback != nil {
		callback(op)
	}
}

// performDelete executes the delete operation.
func performDelete(op *Operation, path string, callback ProgressCallback) {
	op.SetStatus(StatusRunning)

	totalSize, err := calculateSize(path)
	if err != nil {
		op.SetError(fmt.Errorf("failed to calculate size: %w", err))
		if callback != nil {
			callback(op)
		}
		return
	}

	op.UpdateProgress(0, totalSize, path)
	if callback != nil {
		callback(op)
	}

	err = os.RemoveAll(path)
	if err != nil {
		op.SetError(fmt.Errorf("failed to delete: %w", err))
	} else {
		op.UpdateProgress(totalSize, totalSize, path)
		op.SetStatus(StatusCompleted)
	}

	if callback != nil {
		callback(op)
	}
}

// performDeleteMultiple executes delete operation for multiple paths.
func performDeleteMultiple(op *Operation, paths []string, callback ProgressCallback) {
	op.SetStatus(StatusRunning)

	for _, path := range paths {
		if op.IsCancelled() {
			break
		}

		err := os.RemoveAll(path)
		if err != nil {
			op.SetError(fmt.Errorf("failed to delete %s: %w", path, err))
			if callback != nil {
				callback(op)
			}
			return
		}
	}

	if !op.IsCancelled() {
		op.SetStatus(StatusCompleted)
	}

	if callback != nil {
		callback(op)
	}
}

// performRename executes the rename operation.
func performRename(op *Operation, oldPath, newPath string, callback ProgressCallback) {
	op.SetStatus(StatusRunning)

	err := os.Rename(oldPath, newPath)
	if err != nil {
		op.SetError(fmt.Errorf("failed to rename: %w", err))
	} else {
		op.UpdateProgress(1, 1, newPath)
		op.SetStatus(StatusCompleted)
	}

	if callback != nil {
		callback(op)
	}
}

// calculateSize calculates the total size of a file or directory.
func calculateSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// generateOperationID generates a unique ID for an operation.
func generateOperationID() string {
	return fmt.Sprintf("op-%d", time.Now().UnixNano())
}
