package fileops

import (
	"sync"
	"testing"
	"time"
)

func TestDebouncer_BasicDebouncing(t *testing.T) {
	debouncer := NewDebouncer(50 * time.Millisecond)
	defer debouncer.Stop()

	callCount := 0
	var mu sync.Mutex

	increment := func() {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	// Rapidly call debounce multiple times
	for i := 0; i < 10; i++ {
		debouncer.Debounce(increment)
		time.Sleep(5 * time.Millisecond) // Small delay between calls
	}

	// Wait for debounce timeout plus some buffer
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Should only have been called once
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestDebouncer_ExecutesAfterTimeout(t *testing.T) {
	timeout := 30 * time.Millisecond
	debouncer := NewDebouncer(timeout)
	defer debouncer.Stop()

	executed := false
	var mu sync.Mutex

	debouncer.Debounce(func() {
		mu.Lock()
		executed = true
		mu.Unlock()
	})

	// Check it hasn't executed immediately
	mu.Lock()
	if executed {
		t.Error("Function executed too early")
	}
	mu.Unlock()

	// Wait for timeout
	time.Sleep(timeout + 10*time.Millisecond)

	// Should have executed now
	mu.Lock()
	defer mu.Unlock()
	if !executed {
		t.Error("Function didn't execute after timeout")
	}
}

func TestDebouncer_StopCancelsPending(t *testing.T) {
	debouncer := NewDebouncer(50 * time.Millisecond)

	executed := false
	var mu sync.Mutex

	debouncer.Debounce(func() {
		mu.Lock()
		executed = true
		mu.Unlock()
	})

	// Stop before timeout
	debouncer.Stop()

	// Wait longer than timeout
	time.Sleep(100 * time.Millisecond)

	// Should not have executed
	mu.Lock()
	defer mu.Unlock()
	if executed {
		t.Error("Function executed after Stop() was called")
	}
}

func TestDebouncer_MultipleDebounceCallsResetTimer(t *testing.T) {
	timeout := 40 * time.Millisecond
	debouncer := NewDebouncer(timeout)
	defer debouncer.Stop()

	callCount := 0
	var mu sync.Mutex

	increment := func() {
		mu.Lock()
		callCount++
		mu.Unlock()
	}

	// Call debounce, then call it again before timeout expires
	debouncer.Debounce(increment)
	time.Sleep(20 * time.Millisecond) // Half the timeout

	debouncer.Debounce(increment)     // This should reset the timer
	time.Sleep(20 * time.Millisecond) // Another half timeout

	// At this point, 40ms total has passed, but timer was reset at 20ms
	// So function should not have executed yet
	mu.Lock()
	if callCount != 0 {
		t.Errorf("Function executed too early, got %d calls", callCount)
	}
	mu.Unlock()

	// Wait for the reset timer to expire
	time.Sleep(30 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if callCount != 1 {
		t.Errorf("Expected 1 call after timer reset, got %d", callCount)
	}
}

func TestDebouncer_StopOnNoPendingCallsIsSafe(t *testing.T) {
	debouncer := NewDebouncer(50 * time.Millisecond)

	// Should not panic
	debouncer.Stop()
	debouncer.Stop() // Call multiple times

	// Use it after stopping
	executed := false
	var mu sync.Mutex

	debouncer.Debounce(func() {
		mu.Lock()
		executed = true
		mu.Unlock()
	})

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if !executed {
		t.Error("Function didn't execute after Stop() and subsequent Debounce()")
	}
}

func TestDebouncer_DifferentFunctions(t *testing.T) {
	debouncer := NewDebouncer(30 * time.Millisecond)
	defer debouncer.Stop()

	firstCalled := false
	secondCalled := false
	var mu sync.Mutex

	// Schedule first function
	debouncer.Debounce(func() {
		mu.Lock()
		firstCalled = true
		mu.Unlock()
	})

	// Schedule second function (should cancel first)
	time.Sleep(10 * time.Millisecond)
	debouncer.Debounce(func() {
		mu.Lock()
		secondCalled = true
		mu.Unlock()
	})

	// Wait for timeout
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Only second function should have executed
	if firstCalled {
		t.Error("First function executed when it should have been canceled")
	}
	if !secondCalled {
		t.Error("Second function didn't execute")
	}
}
