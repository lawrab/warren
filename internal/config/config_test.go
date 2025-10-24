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

	// Check hyprland defaults
	if cfg.Hyprland.Enabled != true {
		t.Errorf("Expected Hyprland.Enabled to be true, got %v", cfg.Hyprland.Enabled)
	}
	if cfg.Hyprland.WorkspaceMemory != true {
		t.Errorf("Expected Hyprland.WorkspaceMemory to be true, got %v", cfg.Hyprland.WorkspaceMemory)
	}
	if cfg.Hyprland.AutoSwitch != true {
		t.Errorf("Expected Hyprland.AutoSwitch to be true, got %v", cfg.Hyprland.AutoSwitch)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Override the config path for testing
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

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
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

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
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

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

func TestLoadWithPartialConfig(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Override the config path for testing
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create a partial config file (missing Hyprland section)
	partialConfig := `[appearance]
show_hidden = true
window_width = 1200

[keybindings]
quit = "Q"
`

	// Write partial config
	configDir := filepath.Join(tmpDir, "warren")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(partialConfig), 0600); err != nil {
		t.Fatalf("Failed to write partial config: %v", err)
	}

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load partial config: %v", err)
	}

	// Verify user-specified values were loaded
	if cfg.Appearance.ShowHidden != true {
		t.Errorf("Expected ShowHidden to be true, got %v", cfg.Appearance.ShowHidden)
	}
	if cfg.Appearance.WindowWidth != 1200 {
		t.Errorf("Expected WindowWidth to be 1200, got %d", cfg.Appearance.WindowWidth)
	}
	if cfg.Keybindings.Quit != "Q" {
		t.Errorf("Expected Quit to be 'Q', got %s", cfg.Keybindings.Quit)
	}

	// Verify missing fields got defaults (especially Hyprland section)
	if cfg.Hyprland.Enabled != true {
		t.Errorf("Expected Hyprland.Enabled default (true), got %v", cfg.Hyprland.Enabled)
	}
	if cfg.Hyprland.WorkspaceMemory != true {
		t.Errorf("Expected Hyprland.WorkspaceMemory default (true), got %v", cfg.Hyprland.WorkspaceMemory)
	}
	if cfg.Hyprland.AutoSwitch != true {
		t.Errorf("Expected Hyprland.AutoSwitch default (true), got %v", cfg.Hyprland.AutoSwitch)
	}

	// Verify other defaults were applied
	if cfg.Keybindings.NavigateUp != "k" {
		t.Errorf("Expected NavigateUp default ('k'), got %s", cfg.Keybindings.NavigateUp)
	}
	if cfg.General.StartDirectory != "~" {
		t.Errorf("Expected StartDirectory default ('~'), got %s", cfg.General.StartDirectory)
	}
}

func TestLoadWithPartialHyprlandSection(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Override the config path for testing
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create a config file with partial Hyprland section (some fields missing)
	partialConfig := `[appearance]
show_hidden = false

[hyprland]
enabled = false
# workspace_memory and auto_switch are missing
`

	// Write partial config
	configDir := filepath.Join(tmpDir, "warren")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(partialConfig), 0600); err != nil {
		t.Fatalf("Failed to write partial config: %v", err)
	}

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load partial config: %v", err)
	}

	// Verify user-specified value was loaded
	if cfg.Hyprland.Enabled != false {
		t.Errorf("Expected Hyprland.Enabled to be false (from config), got %v", cfg.Hyprland.Enabled)
	}

	// Verify missing fields got defaults
	if cfg.Hyprland.WorkspaceMemory != true {
		t.Errorf("Expected Hyprland.WorkspaceMemory default (true), got %v", cfg.Hyprland.WorkspaceMemory)
	}
	if cfg.Hyprland.AutoSwitch != true {
		t.Errorf("Expected Hyprland.AutoSwitch default (true), got %v", cfg.Hyprland.AutoSwitch)
	}
}

func TestDir(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	// Override the config path for testing
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Dir should return the config directory path
	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() failed: %v", err)
	}

	expectedDir := filepath.Join(tmpDir, "warren")
	if dir != expectedDir {
		t.Errorf("Dir() = %s, want %s", dir, expectedDir)
	}
}

func TestDirWithNoXDG(t *testing.T) {
	// Unset XDG_CONFIG_HOME to test fallback
	t.Setenv("XDG_CONFIG_HOME", "")

	// Dir should fall back to ~/.config/warren
	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() failed: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expectedDir := filepath.Join(homeDir, ".config", "warren")
	if dir != expectedDir {
		t.Errorf("Dir() = %s, want %s", dir, expectedDir)
	}
}
