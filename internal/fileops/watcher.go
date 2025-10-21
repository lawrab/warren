package fileops

import (
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher watches a directory for changes and triggers a callback.
// It wraps fsnotify.Watcher and provides a simple interface for directory watching.
type FileWatcher struct {
	watcher     *fsnotify.Watcher
	onChange    func()        // Callback when files change
	stopChan    chan struct{} // Signal to stop watching
	mu          sync.Mutex    // Protects currentPath
	currentPath string        // Currently watched directory
	running     bool          // Whether watcher is running
}

// NewFileWatcher creates a new file watcher with the given onChange callback.
// The onChange callback will be called from a goroutine whenever files in the
// watched directory change. The callback should use appropriate thread-safety
// mechanisms (like glib.IdleAdd for GTK).
func NewFileWatcher(onChange func()) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := &FileWatcher{
		watcher:  watcher,
		onChange: onChange,
		stopChan: make(chan struct{}),
	}

	return fw, nil
}

// Start begins watching the specified directory.
// If already watching a directory, it stops watching the old one first.
func (fw *FileWatcher) Start(path string) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	// If already watching a different path, remove it
	if fw.currentPath != "" && fw.currentPath != path {
		if err := fw.watcher.Remove(fw.currentPath); err != nil {
			log.Printf("Warning: failed to remove old watch path %s: %v", fw.currentPath, err)
		}
	}

	// Add the new path
	if err := fw.watcher.Add(path); err != nil {
		return err
	}

	fw.currentPath = path

	// Start event loop if not already running
	if !fw.running {
		fw.running = true
		go fw.eventLoop()
	}

	return nil
}

// Stop stops watching and cleans up resources.
func (fw *FileWatcher) Stop() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if !fw.running {
		return nil
	}

	fw.running = false
	close(fw.stopChan)

	return fw.watcher.Close()
}

// eventLoop runs in a goroutine and processes file system events.
func (fw *FileWatcher) eventLoop() {
	// Create debouncer to coalesce rapid file changes
	debouncer := NewDebouncer(100 * time.Millisecond)
	defer debouncer.Stop()

	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Call onChange callback for relevant events
			// We care about: Create, Write, Remove, Rename
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				// Log only events we're acting on
				log.Printf("File watcher event: %s %s", event.Op, event.Name)

				if fw.onChange != nil {
					// Debounce the callback to avoid excessive reloads
					debouncer.Debounce(fw.onChange)
				}
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)

		case <-fw.stopChan:
			return
		}
	}
}

// CurrentPath returns the currently watched directory path.
func (fw *FileWatcher) CurrentPath() string {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.currentPath
}
