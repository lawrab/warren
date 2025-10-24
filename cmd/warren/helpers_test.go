package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lawrab/warren/pkg/models"
)

func TestParseSortMode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected models.SortBy
	}{
		{"name lowercase", "name", models.SortByName},
		{"name capitalized", "Name", models.SortByName},
		{"size lowercase", "size", models.SortBySize},
		{"size capitalized", "Size", models.SortBySize},
		{"modified", "modified", models.SortByModTime},
		{"Modified capitalized", "Modified", models.SortByModTime},
		{"modtime", "modtime", models.SortByModTime},
		{"extension", "extension", models.SortByExtension},
		{"Extension capitalized", "Extension", models.SortByExtension},
		{"ext", "ext", models.SortByExtension},
		{"invalid defaults to name", "invalid", models.SortByName},
		{"empty defaults to name", "", models.SortByName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSortMode(tt.input)
			if result != tt.expected {
				t.Errorf("parseSortMode(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseSortOrder(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected models.SortOrder
	}{
		{"ascending", "ascending", models.SortAscending},
		{"Ascending capitalized", "Ascending", models.SortAscending},
		{"asc", "asc", models.SortAscending},
		{"descending", "descending", models.SortDescending},
		{"Descending capitalized", "Descending", models.SortDescending},
		{"desc", "desc", models.SortDescending},
		{"invalid defaults to ascending", "invalid", models.SortAscending},
		{"empty defaults to ascending", "", models.SortAscending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSortOrder(tt.input)
			if result != tt.expected {
				t.Errorf("parseSortOrder(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetStartDirectory(t *testing.T) {
	// Get actual home directory for tests
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"tilde returns home", "~", homeDir},
		{"empty returns home", "", homeDir},
		{"absolute existing path", tmpDir, tmpDir},
		{"non-existent absolute falls back to home", "/nonexistent/path/12345", homeDir},
		{"relative path falls back to home", "relative/path", homeDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStartDirectory(tt.input)
			if result != tt.expected {
				t.Errorf("getStartDirectory(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetStartDirectoryFallback(t *testing.T) {
	// Create a file (not a directory) to test the IsDir check
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	// Passing a file path should fall back to home
	result := getStartDirectory(tmpFile)
	if result != homeDir {
		t.Errorf("getStartDirectory(file path) = %q, want home dir %q", result, homeDir)
	}
}
