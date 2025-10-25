// Helper functions for status bar and UI-related utilities.
// This file contains GTK-dependent helper functions used by the main package.
// Pure config parsing functions have been moved to internal/config/helpers.go.
package main

import (
	"fmt"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/lawrab/warren/internal/ui"
)

// updateStatusBar updates the status bar label based on current selection and yank state.
func updateStatusBar(label *gtk.Label, fileView *ui.FileView) {
	selected := fileView.GetSelected()
	yanked := fileView.GetYanked()

	var status string
	if selected != nil {
		status = selected.Path
	} else {
		status = "Ready"
	}

	// Add yank indicator if files are yanked
	if len(yanked) > 0 {
		if len(yanked) == 1 {
			yankName := filepath.Base(yanked[0])
			status = fmt.Sprintf("%s  [Yanked: %s]", status, yankName)
		} else {
			status = fmt.Sprintf("%s  [Yanked: %d files]", status, len(yanked))
		}
	}

	label.SetText(status)
}

// formatSortMode returns a formatted string showing the current sort mode and order.
func formatSortMode(fileView *ui.FileView) string {
	mode := fileView.GetSortMode()
	order := fileView.GetSortOrder()

	arrow := "↑"
	if order == 1 { // SortDescending
		arrow = "↓"
	}

	return fmt.Sprintf("Sort: %s %s", mode.String(), arrow)
}
