// Helper functions for parsing configuration values.
package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/lawrab/warren/pkg/models"
)

// GetStartDirectory determines the starting directory from config.
// Handles ~ for home directory, absolute paths, and falls back to home.
func GetStartDirectory(configDir string) string {
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

// ParseSortMode converts a string to a SortBy value.
func ParseSortMode(mode string) models.SortBy {
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

// ParseSortOrder converts a string to a SortOrder value.
func ParseSortOrder(order string) models.SortOrder {
	switch order {
	case "ascending", "Ascending", "asc":
		return models.SortAscending
	case "descending", "Descending", "desc":
		return models.SortDescending
	default:
		return models.SortAscending
	}
}
