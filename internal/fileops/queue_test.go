package fileops

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue(3)

	if q.MaxConcurrent != 3 {
		t.Errorf("MaxConcurrent = %d, want 3", q.MaxConcurrent)
	}
	if q.Count() != 0 {
		t.Errorf("Initial count = %d, want 0", q.Count())
	}
	if q.RunningCount() != 0 {
		t.Errorf("Initial running count = %d, want 0", q.RunningCount())
	}
}

func TestNewQueue_MinConcurrency(t *testing.T) {
	q := NewQueue(0)

	if q.MaxConcurrent != 1 {
		t.Errorf("MaxConcurrent = %d, want 1 (minimum)", q.MaxConcurrent)
	}
}

func TestQueue_AddAndGet(t *testing.T) {
	q := NewQueue(2)

	op := NewOperation(OpCopy, []string{"/src"}, "/dst")
	q.Add(op)

	// Give it a moment to be added
	time.Sleep(10 * time.Millisecond)

	retrieved := q.Get(op.ID)
	if retrieved == nil {
		t.Fatal("Get() returned nil")
	}
	if retrieved.ID != op.ID {
		t.Errorf("Retrieved operation ID = %s, want %s", retrieved.ID, op.ID)
	}
}

func TestQueue_GetAll(t *testing.T) {
	q := NewQueue(2)

	op1 := NewOperation(OpCopy, []string{"/src1"}, "/dst1")
	op2 := NewOperation(OpMove, []string{"/src2"}, "/dst2")

	q.Add(op1)
	q.Add(op2)

	// Give them a moment to be added
	time.Sleep(10 * time.Millisecond)

	all := q.GetAll()
	if len(all) != 2 {
		t.Errorf("GetAll() returned %d operations, want 2", len(all))
	}
}

//nolint:gosec // Test file permissions are intentionally relaxed
func TestQueue_GetRunning(t *testing.T) {
	tmpDir := t.TempDir()
	q := NewQueue(2)

	// Create a file to copy
	srcFile := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	dstFile := filepath.Join(tmpDir, "dest.txt")
	op := Copy(srcFile, dstFile, nil)
	q.Add(op)

	// Wait a moment for operation to start
	time.Sleep(50 * time.Millisecond)

	running := q.GetRunning()
	// Might be 0 or 1 depending on timing
	t.Logf("Running operations: %d", len(running))
}

func TestQueue_GetPending(t *testing.T) {
	q := NewQueue(1) // Only 1 concurrent

	op1 := NewOperation(OpCopy, []string{"/src1"}, "/dst1")
	op2 := NewOperation(OpMove, []string{"/src2"}, "/dst2")

	q.Add(op1)
	q.Add(op2)

	// One should be running, one might be pending
	time.Sleep(10 * time.Millisecond)

	pending := q.GetPending()
	t.Logf("Pending operations: %d", len(pending))
}

func TestQueue_Cancel(t *testing.T) {
	q := NewQueue(2)

	op := NewOperation(OpCopy, []string{"/src"}, "/dst")
	q.Add(op)

	time.Sleep(10 * time.Millisecond)

	success := q.Cancel(op.ID)
	if !success {
		t.Error("Cancel() should return true for existing operation")
	}

	// Verify operation was cancelled
	if !op.IsCancelled() {
		t.Error("Operation should be cancelled")
	}
}

func TestQueue_CancelNonexistent(t *testing.T) {
	q := NewQueue(2)

	success := q.Cancel("nonexistent-id")
	if success {
		t.Error("Cancel() should return false for non-existent operation")
	}
}

func TestQueue_CancelAll(t *testing.T) {
	q := NewQueue(2)

	op1 := NewOperation(OpCopy, []string{"/src1"}, "/dst1")
	op2 := NewOperation(OpMove, []string{"/src2"}, "/dst2")

	q.Add(op1)
	q.Add(op2)

	time.Sleep(10 * time.Millisecond)

	q.CancelAll()

	// Both operations should be cancelled
	if !op1.IsCancelled() {
		t.Error("Operation 1 should be cancelled")
	}
	if !op2.IsCancelled() {
		t.Error("Operation 2 should be cancelled")
	}
}

//nolint:gosec // Test file permissions are intentionally relaxed
func TestQueue_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	q := NewQueue(2)

	// Create completed operation
	srcFile := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	dstFile := filepath.Join(tmpDir, "dest.txt")
	op := Copy(srcFile, dstFile, nil)
	q.Add(op)

	// Wait for completion
	waitForOperation(t, op, 5*time.Second)

	// Clear completed operations
	q.Clear()

	// Completed operation should be removed
	all := q.GetAll()
	if len(all) != 0 {
		t.Errorf("After Clear(), GetAll() returned %d operations, want 0", len(all))
	}
}

func TestQueue_Count(t *testing.T) {
	q := NewQueue(2)

	if q.Count() != 0 {
		t.Errorf("Initial count = %d, want 0", q.Count())
	}

	op1 := NewOperation(OpCopy, []string{"/src1"}, "/dst1")
	q.Add(op1)

	time.Sleep(10 * time.Millisecond)

	if q.Count() != 1 {
		t.Errorf("Count after 1 add = %d, want 1", q.Count())
	}

	op2 := NewOperation(OpMove, []string{"/src2"}, "/dst2")
	q.Add(op2)

	time.Sleep(10 * time.Millisecond)

	if q.Count() != 2 {
		t.Errorf("Count after 2 adds = %d, want 2", q.Count())
	}
}

//nolint:gosec // Test file permissions are intentionally relaxed
func TestQueue_Concurrency(t *testing.T) {
	tmpDir := t.TempDir()
	q := NewQueue(2) // Max 2 concurrent

	// Create multiple files to copy
	var ops []*Operation
	for i := 0; i < 5; i++ {
		srcFile := filepath.Join(tmpDir, "source"+string(rune('A'+i))+".txt")
		if err := os.WriteFile(srcFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		dstFile := filepath.Join(tmpDir, "dest"+string(rune('A'+i))+".txt")
		op := Copy(srcFile, dstFile, nil)
		q.Add(op)
		ops = append(ops, op)
	}

	// Check that no more than 2 are running at once
	time.Sleep(50 * time.Millisecond)
	runningCount := q.RunningCount()
	if runningCount > 2 {
		t.Errorf("Running count = %d, should not exceed max concurrent (2)", runningCount)
	}

	// Wait for all to complete
	for _, op := range ops {
		waitForOperation(t, op, 5*time.Second)
	}

	// All should be completed now
	completed := q.GetCompleted()
	if len(completed) != 5 {
		t.Errorf("Completed count = %d, want 5", len(completed))
	}
}
