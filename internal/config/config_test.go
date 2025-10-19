package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	// Check appearance defaults
	if cfg.Appearance.ShowHidden != false {
		t.Errorf("Expected ShowHidden to be false, got %v", cfg.Appearance.ShowHidden)
	}
	if cfg.Appearance.WindowWidth != 1000 {
		t.Errorf("Expected WindowWidth to be 1000, got %d", cfg.Appearance.WindowWidth)
	}
	if cfg.Appearance.WindowHeight != 700 {
		t.Errorf("Expected WindowHeight to be 700, got %d", cfg.Appearance.WindowHeight)
	}

	// Check keybinding defaults
	if cfg.Keybindings.Quit != "q" {
		t.Errorf("Expected Quit to be 'q', got %s", cfg.Keybindings.Quit)
	}
	if cfg.Keybindings.NavigateUp != "k" {
		t.Errorf("Expected NavigateUp to be 'k', got %s", cfg.Keybindings.NavigateUp)
	}
	if cfg.Keybindings.NavigateDown != "j" {
		t.Errorf("Expected NavigateDown to be 'j', got %s", cfg.Keybindings.NavigateDown)
	}

	// Check general defaults
	if cfg.General.StartDirectory != "~" {
		t.Errorf("Expected StartDirectory to be '~', got %s", cfg.General.StartDirectory)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Override the config path for testing
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// Create a custom config
	cfg := Default()
	cfg.Appearance.ShowHidden = true
	cfg.Appearance.WindowWidth = 1200
	cfg.Keybindings.Quit = "Q"
	cfg.General.StartDirectory = "/tmp"

	// Save config
	if err := Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filepath.Join(tmpDir, "warren", "config.toml")); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load config
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values match
	if loaded.Appearance.ShowHidden != true {
		t.Errorf("Expected ShowHidden to be true, got %v", loaded.Appearance.ShowHidden)
	}
	if loaded.Appearance.WindowWidth != 1200 {
		t.Errorf("Expected WindowWidth to be 1200, got %d", loaded.Appearance.WindowWidth)
	}
	if loaded.Keybindings.Quit != "Q" {
		t.Errorf("Expected Quit to be 'Q', got %s", loaded.Keybindings.Quit)
	}
	if loaded.General.StartDirectory != "/tmp" {
		t.Errorf("Expected StartDirectory to be '/tmp', got %s", loaded.General.StartDirectory)
	}
}

func TestLoadNonexistent(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Override the config path for testing
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// Load config when file doesn't exist
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should return defaults when file doesn't exist, got error: %v", err)
	}

	// Should return default config
	defaultCfg := Default()
	if cfg.Appearance.WindowWidth != defaultCfg.Appearance.WindowWidth {
		t.Errorf("Expected default config when file doesn't exist")
	}
}

func TestLoadOrDefault(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Override the config path for testing
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

	// LoadOrDefault should never fail, always return valid config
	cfg := LoadOrDefault()
	if cfg == nil {
		t.Fatal("LoadOrDefault returned nil")
	}

	// Should return defaults
	if cfg.Keybindings.Quit != "q" {
		t.Errorf("Expected default keybindings")
	}
}
