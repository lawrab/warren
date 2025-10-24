// Helper functions for status bar, sorting, and directory handling.
// This file contains pure utility functions used by the main package
// that don't require GTK context and are easily testable.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/lawrab/warren/internal/ui"
	"github.com/lawrab/warren/pkg/models"
)

// updateStatusBar updates the status bar label based on current selection.
func updateStatusBar(label *gtk.Label, fileView *ui.FileView) {
	selected := fileView.GetSelected()
	if selected != nil {
		label.SetText(selected.Path)
	} else {
		label.SetText("Ready")
	}
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

// getStartDirectory determines the starting directory from config.
// Handles ~ for home directory, absolute paths, and falls back to home.
func getStartDirectory(configDir string) string {
	// Handle home directory
	if configDir == "~" || configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Failed to get home directory: %v", err)
			return "/"
		}
		return homeDir
	}

	// Handle absolute paths
	if filepath.IsAbs(configDir) {
		// Verify directory exists
		if info, err := os.Stat(configDir); err == nil && info.IsDir() {
			return configDir
		}
		log.Printf("Configured start directory %s does not exist, using home", configDir)
	}

	// Fall back to home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return homeDir
}

// parseSortMode converts a string to a SortBy value.
func parseSortMode(mode string) models.SortBy {
	switch mode {
	case "name", "Name":
		return models.SortByName
	case "size", "Size":
		return models.SortBySize
	case "modified", "Modified", "modtime":
		return models.SortByModTime
	case "extension", "Extension", "ext":
		return models.SortByExtension
	default:
		return models.SortByName
	}
}

// parseSortOrder converts a string to a SortOrder value.
func parseSortOrder(order string) models.SortOrder {
	switch order {
	case "ascending", "Ascending", "asc":
		return models.SortAscending
	case "descending", "Descending", "desc":
		return models.SortDescending
	default:
		return models.SortAscending
	}
}
