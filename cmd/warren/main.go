// Warren is a keyboard-driven GTK4 file manager built for Hyprland.
// This package contains the main entry point and UI setup logic.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/lawrab/warren/internal/config"
	"github.com/lawrab/warren/internal/fileops"
	"github.com/lawrab/warren/internal/hyprland"
	"github.com/lawrab/warren/internal/ui"
	"github.com/lawrab/warren/internal/version"
	"github.com/lawrab/warren/pkg/models"
)

const appID = "com.lawrab.warren"

// hyprlandState holds Hyprland integration state.
type hyprlandState struct {
	client *hyprland.Client
	memory *hyprland.WorkspaceMemory
}

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
	// Initialize Hyprland integration
	hyprState := setupHyprland(cfg)

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

	// Add sort mode indicator
	sortLabel := gtk.NewLabel(formatSortMode(fileView))
	sortLabel.AddCSSClass("dim-label")
	sortLabel.SetMarginEnd(12)
	statusBar.Append(sortLabel)

	helpLabel := gtk.NewLabel("j/k: navigate  h: up  l: enter  s: sort  r: reverse  .: hidden  q: quit")
	helpLabel.AddCSSClass("dim-label")
	statusBar.Append(helpLabel)

	box.Append(statusBar)

	// Add box to window
	window.SetChild(box)

	// Determine starting directory
	// First check if there's a remembered directory for current workspace
	startDir := getStartDirectory(cfg.General.StartDirectory)
	if hyprState != nil && hyprState.client != nil && hyprState.memory != nil && cfg.Hyprland.WorkspaceMemory {
		if ws, err := hyprState.client.GetActiveWorkspace(); err == nil {
			if rememberedDir := hyprState.memory.Get(ws.ID); rememberedDir != "" {
				// Verify directory still exists
				if info, err := os.Stat(rememberedDir); err == nil && info.IsDir() {
					startDir = rememberedDir
					log.Printf("Using remembered directory for workspace %d: %s", ws.ID, rememberedDir)
				}
			}
		}
	}

	// Apply sort mode from config
	sortMode := parseSortMode(cfg.Appearance.DefaultSortMode)
	sortOrder := parseSortOrder(cfg.Appearance.DefaultSortOrder)
	fileView.SetSortMode(sortMode, sortOrder)

	// Load initial directory
	if err := fileView.LoadDirectory(startDir); err != nil {
		log.Printf("Failed to load directory: %v", err)
		statusLabel.SetText(err.Error())
	} else {
		pathLabel.SetText(fileView.GetCurrentPath())
		updateStatusBar(statusLabel, fileView)
		// Save initial directory to workspace memory
		saveCurrentDirectoryToWorkspace(hyprState, fileView.GetCurrentPath())
	}

	// Apply initial show hidden files setting from config
	if cfg.Appearance.ShowHidden {
		if err := fileView.ToggleHidden(); err != nil {
			log.Printf("Failed to apply show_hidden setting: %v", err)
		}
	}

	// Update sort label to reflect initial state
	sortLabel.SetText(formatSortMode(fileView))

	// Start Hyprland event listener
	startHyprlandListener(hyprState, cfg, fileView, pathLabel, statusLabel)

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
				// Save new directory to workspace memory
				saveCurrentDirectoryToWorkspace(hyprState, fileView.GetCurrentPath())
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
					// Save new directory to workspace memory
					saveCurrentDirectoryToWorkspace(hyprState, fileView.GetCurrentPath())
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

		if keyMatchesConfig(keyval, cfg.Keybindings.CycleSortMode) {
			if err := fileView.CycleSortMode(); err != nil {
				statusLabel.SetText(err.Error())
			} else {
				sortLabel.SetText(formatSortMode(fileView))
				updateStatusBar(statusLabel, fileView)
			}
			return true
		}

		if keyMatchesConfig(keyval, cfg.Keybindings.ToggleSortOrder) {
			if err := fileView.ToggleSortOrder(); err != nil {
				statusLabel.SetText(err.Error())
			} else {
				sortLabel.SetText(formatSortMode(fileView))
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

	// Cleanup file watcher and save workspace memory when window closes
	window.ConnectCloseRequest(func() bool {
		if err := fileView.Close(); err != nil {
			log.Printf("Warning: Failed to close file watcher: %v", err)
		}
		// Save workspace memory on exit
		if hyprState != nil && hyprState.memory != nil {
			if err := hyprState.memory.Save(); err != nil {
				log.Printf("Warning: Failed to save workspace memory: %v", err)
			}
		}
		return false // Allow window to close
	})

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

// formatSortMode returns a human-readable string for the current sort mode.
// Examples: "Name ↑", "Size ↓"
func formatSortMode(fileView *ui.FileView) string {
	mode := fileView.GetSortMode()
	order := fileView.GetSortOrder()

	arrow := "↑"
	if order == 1 { // SortDescending
		arrow = "↓"
	}

	return fmt.Sprintf("Sort: %s %s", mode.String(), arrow)
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

// parseSortMode converts a config string to a models.SortBy value.
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

// parseSortOrder converts a config string to a models.SortOrder value.
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

// setupHyprland initializes Hyprland integration if enabled and available.
// Returns nil if Hyprland is not available or disabled in config.
func setupHyprland(cfg *config.Config) *hyprlandState {
	// Check if Hyprland integration is enabled
	if !cfg.Hyprland.Enabled {
		log.Println("Hyprland integration disabled in config")
		return nil
	}

	// Check if running in Hyprland
	if !hyprland.IsHyprland() {
		log.Println("Not running in Hyprland, skipping integration")
		return nil
	}

	// Create Hyprland client
	client, err := hyprland.New()
	if err != nil {
		log.Printf("Failed to create Hyprland client: %v", err)
		return nil
	}

	// Create workspace memory if enabled
	var memory *hyprland.WorkspaceMemory
	if cfg.Hyprland.WorkspaceMemory {
		configDir, err := config.Dir()
		if err != nil {
			log.Printf("Failed to get config dir: %v", err)
			configDir = ""
		}

		memory, err = hyprland.NewWorkspaceMemory(configDir)
		if err != nil {
			log.Printf("Failed to create workspace memory: %v", err)
			memory = nil
		} else {
			log.Println("Hyprland workspace memory enabled")
		}
	}

	log.Println("Hyprland integration initialized")
	return &hyprlandState{
		client: client,
		memory: memory,
	}
}

// startHyprlandListener starts listening for Hyprland events in a goroutine.
// It handles workspace changes and updates the file view accordingly.
func startHyprlandListener(hs *hyprlandState, cfg *config.Config, fileView *ui.FileView, pathLabel *gtk.Label, statusLabel *gtk.Label) {
	if hs == nil || hs.client == nil {
		return
	}

	go func() {
		err := hs.client.ListenEvents(func(event hyprland.Event) {
			if event.Type == "workspace" && cfg.Hyprland.AutoSwitch && hs.memory != nil {
				// Parse workspace ID from event data
				workspaceID, err := strconv.Atoi(strings.TrimSpace(event.Data))
				if err != nil {
					log.Printf("Failed to parse workspace ID from event: %v", err)
					return
				}

				// Get remembered directory for this workspace
				rememberedDir := hs.memory.Get(workspaceID)
				if rememberedDir == "" {
					log.Printf("No remembered directory for workspace %d", workspaceID)
					return
				}

				// Verify directory still exists
				if info, err := os.Stat(rememberedDir); err != nil || !info.IsDir() {
					log.Printf("Remembered directory %s no longer exists", rememberedDir)
					return
				}

				// Switch to remembered directory (must use glib.IdleAdd for GTK operations)
				glib.IdleAdd(func() {
					if err := fileView.LoadDirectory(rememberedDir); err != nil {
						log.Printf("Failed to load remembered directory: %v", err)
						statusLabel.SetText(fmt.Sprintf("Failed to load: %v", err))
					} else {
						pathLabel.SetText(fileView.GetCurrentPath())
						updateStatusBar(statusLabel, fileView)
						log.Printf("Switched to workspace %d directory: %s", workspaceID, rememberedDir)
					}
				})
			}
		})

		if err != nil {
			log.Printf("Hyprland event listener stopped: %v", err)
		}
	}()

	log.Println("Hyprland event listener started")
}

// saveCurrentDirectoryToWorkspace saves the current directory to workspace memory.
func saveCurrentDirectoryToWorkspace(hs *hyprlandState, currentPath string) {
	if hs == nil || hs.client == nil || hs.memory == nil {
		return
	}

	// Get current workspace
	ws, err := hs.client.GetActiveWorkspace()
	if err != nil {
		log.Printf("Failed to get active workspace: %v", err)
		return
	}

	// Save current directory to memory
	hs.memory.Set(ws.ID, currentPath)

	// Persist to disk
	if err := hs.memory.Save(); err != nil {
		log.Printf("Failed to save workspace memory: %v", err)
	}
}
