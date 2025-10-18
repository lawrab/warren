// Package ui provides GTK4 user interface components for Warren.
//
// This package contains all GTK-specific code including widgets, windows,
// and event handlers. It depends on the gotk4 library for GTK4 bindings.
//
// All UI updates must happen on the GTK main thread. Use glib.IdleAdd()
// when updating the UI from goroutines.
package ui
