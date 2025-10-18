// Package fileops provides filesystem operations for Warren.
//
// This package handles all filesystem interactions including directory listing,
// file operations (copy, move, delete), and file information retrieval.
// It has no dependencies on GTK or UI code, making it easily testable.
//
// All operations that may take significant time should be designed to run
// in goroutines without blocking the UI thread.
package fileops
