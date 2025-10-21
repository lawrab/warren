package fileops

import "time"

// Debouncer delays function execution until after a quiet period.
// It's useful for coalescing rapid events (like file system changes)
// into a single action after things settle down.
type Debouncer struct {
	timer   *time.Timer
	timeout time.Duration
}

// NewDebouncer creates a debouncer with the given timeout.
// The timeout specifies how long to wait after the last call
// before executing the debounced function.
func NewDebouncer(timeout time.Duration) *Debouncer {
	return &Debouncer{timeout: timeout}
}

// Debounce schedules fn to run after the timeout, canceling any pending call.
// If called multiple times rapidly, only the last call's function will execute,
// and only after the timeout period of inactivity.
func (d *Debouncer) Debounce(fn func()) {
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.timeout, fn)
}

// Stop cancels any pending debounced call.
// It's safe to call Stop multiple times or on a debouncer with no pending calls.
func (d *Debouncer) Stop() {
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
}
