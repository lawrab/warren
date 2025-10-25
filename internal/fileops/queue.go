package fileops

import (
	"sync"
)

// OperationQueue manages a queue of file operations.
// It limits concurrent operations and provides status tracking.
type OperationQueue struct {
	// MaxConcurrent is the maximum number of concurrent operations
	MaxConcurrent int

	// operations stores all operations (pending, running, completed)
	operations []*Operation

	// running tracks currently running operations
	running map[string]*Operation

	// mutex protects concurrent access
	mu sync.RWMutex

	// sem limits concurrent operations
	sem chan struct{}
}

// NewQueue creates a new operation queue.
func NewQueue(maxConcurrent int) *OperationQueue {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}

	return &OperationQueue{
		MaxConcurrent: maxConcurrent,
		operations:    make([]*Operation, 0),
		running:       make(map[string]*Operation),
		sem:           make(chan struct{}, maxConcurrent),
	}
}

// Add adds an operation to the queue and starts it if possible.
func (q *OperationQueue) Add(op *Operation) {
	q.mu.Lock()
	q.operations = append(q.operations, op)
	q.mu.Unlock()

	go q.runOperation(op)
}

// runOperation runs an operation when a slot is available.
func (q *OperationQueue) runOperation(op *Operation) {
	// Wait for available slot
	q.sem <- struct{}{}
	defer func() { <-q.sem }()

	q.mu.Lock()
	q.running[op.ID] = op
	q.mu.Unlock()

	// Wait for operation to complete
	// The operation goroutine will update its status
waitLoop:
	for {
		status := op.Status
		if status == StatusCompleted || status == StatusFailed || status == StatusCancelled {
			break waitLoop
		}
		// Small sleep to avoid busy waiting
		// In production, consider using channels or condition variables
		select {
		case <-op.ctx.Done():
			break waitLoop
		default:
		}
	}

	q.mu.Lock()
	delete(q.running, op.ID)
	q.mu.Unlock()
}

// Get returns an operation by ID.
func (q *OperationQueue) Get(id string) *Operation {
	q.mu.RLock()
	defer q.mu.RUnlock()

	for _, op := range q.operations {
		if op.ID == id {
			return op
		}
	}
	return nil
}

// GetAll returns all operations.
func (q *OperationQueue) GetAll() []*Operation {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]*Operation, len(q.operations))
	copy(result, q.operations)
	return result
}

// GetRunning returns all currently running operations.
func (q *OperationQueue) GetRunning() []*Operation {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]*Operation, 0, len(q.running))
	for _, op := range q.running {
		result = append(result, op)
	}
	return result
}

// GetPending returns all pending operations.
func (q *OperationQueue) GetPending() []*Operation {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]*Operation, 0)
	for _, op := range q.operations {
		if op.Status == StatusPending {
			result = append(result, op)
		}
	}
	return result
}

// GetCompleted returns all completed operations.
func (q *OperationQueue) GetCompleted() []*Operation {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]*Operation, 0)
	for _, op := range q.operations {
		if op.Status == StatusCompleted {
			result = append(result, op)
		}
	}
	return result
}

// Cancel cancels an operation by ID.
func (q *OperationQueue) Cancel(id string) bool {
	op := q.Get(id)
	if op != nil {
		op.Cancel()
		return true
	}
	return false
}

// CancelAll cancels all pending and running operations.
func (q *OperationQueue) CancelAll() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, op := range q.operations {
		if op.Status == StatusPending || op.Status == StatusRunning {
			op.Cancel()
		}
	}
}

// Clear removes completed, failed, and cancelled operations from history.
func (q *OperationQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	active := make([]*Operation, 0)
	for _, op := range q.operations {
		if op.Status == StatusPending || op.Status == StatusRunning {
			active = append(active, op)
		}
	}
	q.operations = active
}

// Count returns the total number of operations.
func (q *OperationQueue) Count() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.operations)
}

// RunningCount returns the number of currently running operations.
func (q *OperationQueue) RunningCount() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.running)
}
