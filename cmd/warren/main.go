// Warren is a keyboard-driven GTK4 file manager built for Hyprland.
// This package contains the main entry point and UI setup logic.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/lawrab/warren/internal/config"
	"github.com/lawrab/warren/internal/fileops"
	"github.com/lawrab/warren/internal/ui"
	"github.com/lawrab/warren/internal/version"
)

const appID = "com.lawrab.warren"

func main() {
	// Parse command line flags
	showVersion := flag.Bool("version", false, "Show version information")
	flag.BoolVar(showVersion, "v", false, "Show version information (shorthand)")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Println(version.FullVersion())
		os.Exit(0)
	}

	// Load configuration
	cfg := config.LoadOrDefault()

	app := gtk.NewApplication(appID, 0)
	app.ConnectActivate(func() { activate(app, cfg) })

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate(app *gtk.Application, cfg *config.Config) {
	// Create main window
	window := gtk.NewApplicationWindow(app)
	window.SetTitle(fmt.Sprintf("Warren %s", version.Short()))
	window.SetDefaultSize(cfg.Appearance.WindowWidth, cfg.Appearance.WindowHeight)

	// Create a header bar
	headerBar := gtk.NewHeaderBar()
	headerBar.SetShowTitleButtons(true)

	// Add path label to header
	pathLabel := gtk.NewLabel("")
	pathLabel.AddCSSClass("title")
	headerBar.SetTitleWidget(pathLabel)

	window.SetTitlebar(headerBar)

	// Create main box layout
	box := gtk.NewBox(gtk.OrientationVertical, 0)

	// Create file view
	fileView := ui.NewFileView()
	box.Append(fileView.Widget())

	// Create status bar
	statusBar := gtk.NewBox(gtk.OrientationHorizontal, 12)
	statusBar.SetMarginTop(6)
	statusBar.SetMarginBottom(6)
	statusBar.SetMarginStart(12)
	statusBar.SetMarginEnd(12)
	statusLabel := gtk.NewLabel("Ready")
	statusLabel.SetXAlign(0)
	statusLabel.SetHExpand(true)
	statusBar.Append(statusLabel)

	helpLabel := gtk.NewLabel("j/k: navigate  h: up  l: enter/open  q: quit  .: toggle hidden")
	helpLabel.AddCSSClass("dim-label")
	statusBar.Append(helpLabel)

	box.Append(statusBar)

	// Add box to window
	window.SetChild(box)

	// Determine starting directory
	startDir := getStartDirectory(cfg.General.StartDirectory)

	// Load initial directory
	if err := fileView.LoadDirectory(startDir); err != nil {
		log.Printf("Failed to load directory: %v", err)
		statusLabel.SetText(err.Error())
	} else {
		pathLabel.SetText(fileView.GetCurrentPath())
		updateStatusBar(statusLabel, fileView)
	}

	// Apply initial show hidden files setting from config
	if cfg.Appearance.ShowHidden {
		if err := fileView.ToggleHidden(); err != nil {
			log.Printf("Failed to apply show_hidden setting: %v", err)
		}
	}

	// Set up keyboard event controller
	keyController := gtk.NewEventControllerKey()
	keyController.ConnectKeyPressed(func(keyval uint, _ uint, _ gdk.ModifierType) bool {
		// Convert pressed key to string for comparison
		keyName := gdk.KeyvalName(keyval)

		// Check custom keybindings from config
		if keyMatchesConfig(keyval, cfg.Keybindings.NavigateDown) || keyval == gdk.KEY_Down {
			fileView.SelectNext()
			updateStatusBar(statusLabel, fileView)
			return true
		}

		if keyMatchesConfig(keyval, cfg.Keybindings.NavigateUp) || keyval == gdk.KEY_Up {
			fileView.SelectPrevious()
			updateStatusBar(statusLabel, fileView)
			return true
		}

		if keyMatchesConfig(keyval, cfg.Keybindings.ParentDir) || keyval == gdk.KEY_Left || keyval == gdk.KEY_BackSpace {
			if err := fileView.NavigateUp(); err != nil {
				statusLabel.SetText(err.Error())
			} else {
				pathLabel.SetText(fileView.GetCurrentPath())
				updateStatusBar(statusLabel, fileView)
			}
			return true
		}

		if keyMatchesConfig(keyval, cfg.Keybindings.EnterDir) || keyval == gdk.KEY_Right || keyval == gdk.KEY_Return {
			selected := fileView.GetSelected()
			if selected == nil {
				return true
			}

			if selected.IsDir {
				// Navigate into directory
				if err := fileView.NavigateInto(); err != nil {
					statusLabel.SetText(err.Error())
				} else {
					pathLabel.SetText(fileView.GetCurrentPath())
					updateStatusBar(statusLabel, fileView)
				}
			} else {
				// Open file with default application
				if err := fileops.OpenFile(selected.Path); err != nil {
					statusLabel.SetText(fmt.Sprintf("Failed to open: %v", err))
					log.Printf("Failed to open file %s: %v", selected.Path, err)
				} else {
					statusLabel.SetText(fmt.Sprintf("Opened: %s", selected.Name))
				}
			}
			return true
		}

		if keyMatchesConfig(keyval, cfg.Keybindings.ToggleHidden) {
			if err := fileView.ToggleHidden(); err != nil {
				statusLabel.SetText(err.Error())
			} else {
				updateStatusBar(statusLabel, fileView)
			}
			return true
		}

		if keyMatchesConfig(keyval, cfg.Keybindings.Quit) {
			window.Close()
			return true
		}

		_ = keyName // Keep for potential debugging
		return false
	})

	window.AddController(keyController)

	// Keyboard shortcuts
	setupShortcuts(app, window)

	// Show window
	window.Present()
}

func updateStatusBar(label *gtk.Label, fileView *ui.FileView) {
	selected := fileView.GetSelected()
	if selected != nil {
		label.SetText(selected.Path)
	} else {
		label.SetText("Ready")
	}
}

func setupShortcuts(app *gtk.Application, window *gtk.ApplicationWindow) {
	// Quit on Ctrl+Q
	quitAction := gio.NewSimpleAction("quit", nil)
	quitAction.ConnectActivate(func(_ *glib.Variant) {
		window.Close()
	})
	app.AddAction(quitAction)
	app.SetAccelsForAction("app.quit", []string{"<Ctrl>Q"})
}

// keyMatchesConfig checks if a pressed keyval matches a configured key binding.
// The config string can be a simple letter ("j", "k") or a special key name
// ("period", "Return", "Escape", etc.).
func keyMatchesConfig(keyval uint, configKey string) bool {
	if configKey == "" {
		return false
	}

	// Get the key name from the keyval
	keyName := gdk.KeyvalName(keyval)

	// Direct name match (e.g., "Return", "Escape", "BackSpace")
	if keyName == configKey {
		return true
	}

	// Single character match (e.g., "j" matches 'j')
	if len(configKey) == 1 {
		if keyval == uint(configKey[0]) {
			return true
		}
	}

	return false
}

// getStartDirectory returns the starting directory based on config.
// Supports "~" for home directory and absolute paths.
// TODO: Support "last" to remember last directory (Phase 1 future enhancement).
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
