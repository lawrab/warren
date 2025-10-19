package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config represents Warren's configuration structure.
type Config struct {
	Appearance  AppearanceConfig  `toml:"appearance"`
	Keybindings KeybindingsConfig `toml:"keybindings"`
	General     GeneralConfig     `toml:"general"`
}

// AppearanceConfig controls visual appearance settings.
type AppearanceConfig struct {
	ShowHidden       bool   `toml:"show_hidden"`        // Show hidden files by default
	WindowWidth      int    `toml:"window_width"`       // Default window width
	WindowHeight     int    `toml:"window_height"`      // Default window height
	DefaultSortMode  string `toml:"default_sort_mode"`  // Default sort mode: "name", "size", "modified", "extension"
	DefaultSortOrder string `toml:"default_sort_order"` // Default sort order: "ascending", "descending"
}

// KeybindingsConfig defines keyboard shortcuts.
// Each field should contain a single key name (e.g., "j", "period", "space").
// For special keys, use GTK key names (e.g., "Return", "BackSpace", "Escape").
type KeybindingsConfig struct {
	Quit          string `toml:"quit"`            // Quit application
	NavigateUp    string `toml:"navigate_up"`     // Move selection up
	NavigateDown  string `toml:"navigate_down"`   // Move selection down
	ParentDir     string `toml:"parent_dir"`      // Go to parent directory
	EnterDir      string `toml:"enter_dir"`       // Enter directory or open file
	ToggleHidden  string `toml:"toggle_hidden"`   // Toggle hidden files visibility
	CycleSortMode string `toml:"cycle_sort_mode"` // Cycle through sort modes
}

// GeneralConfig contains general application settings.
type GeneralConfig struct {
	StartDirectory string `toml:"start_directory"` // Starting directory ("~", "/", or "last")
}

// Default returns a Config with sensible default values.
func Default() *Config {
	return &Config{
		Appearance: AppearanceConfig{
			ShowHidden:       false,
			WindowWidth:      1000,
			WindowHeight:     700,
			DefaultSortMode:  "name",
			DefaultSortOrder: "ascending",
		},
		Keybindings: KeybindingsConfig{
			Quit:          "q",
			NavigateUp:    "k",
			NavigateDown:  "j",
			ParentDir:     "h",
			EnterDir:      "l",
			ToggleHidden:  "period",
			CycleSortMode: "s",
		},
		General: GeneralConfig{
			StartDirectory: "~",
		},
	}
}

// Dir returns the directory where Warren stores its configuration.
// Follows XDG Base Directory specification: $XDG_CONFIG_HOME/warren
// or defaults to ~/.config/warren if XDG_CONFIG_HOME is not set.
func Dir() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		configHome = filepath.Join(homeDir, ".config")
	}

	configDir := filepath.Join(configHome, "warren")
	return configDir, nil
}

// Path returns the full path to the configuration file.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// Load reads the configuration file and returns a Config.
// If the file doesn't exist, returns the default configuration.
// If the file exists but is invalid, returns an error.
// Missing fields in the config file are filled with default values.
func Load() (*Config, error) {
	configPath, err := Path()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// If config doesn't exist, return defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Default(), nil
	}

	// Read config file
	// #nosec G304 -- configPath is derived from XDG spec, not user input
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Start with defaults, then overlay user config
	config := Default()
	if err := toml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// Save writes the configuration to disk.
// Creates the configuration directory if it doesn't exist.
func Save(cfg *Config) error {
	configPath, err := Path()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to TOML
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadOrDefault loads the configuration file, or returns defaults if it doesn't exist.
// Unlike Load(), this function logs errors but doesn't fail - it always returns a valid config.
func LoadOrDefault() *Config {
	cfg, err := Load()
	if err != nil {
		// Log the error but continue with defaults
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config, using defaults: %v\n", err)
		return Default()
	}
	return cfg
}
