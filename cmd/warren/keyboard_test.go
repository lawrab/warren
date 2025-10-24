package main

import (
	"testing"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
)

func TestKeyMatchesConfig(t *testing.T) {
	tests := []struct {
		name     string
		keyval   uint
		config   string
		expected bool
	}{
		// Single character matches
		{"lowercase j", uint('j'), "j", true},
		{"lowercase k", uint('k'), "k", true},
		{"lowercase h", uint('h'), "h", true},
		{"lowercase l", uint('l'), "l", true},
		{"period", uint('.'), ".", true},

		// Single character non-matches
		{"j doesn't match k", uint('j'), "k", false},
		{"k doesn't match j", uint('k'), "j", false},

		// Special key names
		{"Return key", gdk.KEY_Return, "Return", true},
		{"Escape key", gdk.KEY_Escape, "Escape", true},
		{"BackSpace key", gdk.KEY_BackSpace, "BackSpace", true},
		{"Tab key", gdk.KEY_Tab, "Tab", true},

		// Empty config should never match
		{"empty config returns false", uint('j'), "", false},
		{"empty config with special key", gdk.KEY_Return, "", false},

		// Case sensitivity
		{"uppercase J config matches lowercase keyval", uint('j'), "J", false},

		// Arrow keys
		{"Down arrow", gdk.KEY_Down, "Down", true},
		{"Up arrow", gdk.KEY_Up, "Up", true},
		{"Left arrow", gdk.KEY_Left, "Left", true},
		{"Right arrow", gdk.KEY_Right, "Right", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := keyMatchesConfig(tt.keyval, tt.config)
			if result != tt.expected {
				t.Errorf("keyMatchesConfig(%d, %q) = %v, want %v",
					tt.keyval, tt.config, result, tt.expected)
			}
		})
	}
}

func TestKeyMatchesConfigEdgeCases(t *testing.T) {
	// Test that multi-character strings only match if they're key names
	if keyMatchesConfig(uint('a'), "abc") {
		t.Error("Multi-character non-keyname should not match single character")
	}

	// Test that number keys work
	if !keyMatchesConfig(uint('0'), "0") {
		t.Error("Number key '0' should match config '0'")
	}

	// Test space key
	if !keyMatchesConfig(uint(' '), " ") {
		t.Error("Space key should match single space config")
	}
}
