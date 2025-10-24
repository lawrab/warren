// Warren is a keyboard-driven GTK4 file manager built for Hyprland.
// This package contains the main entry point and UI setup logic.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/lawrab/warren/internal/config"
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
	// Initialize Hyprland integration
	hyprState := setupHyprland(cfg)

	// Add CSS styling
	cssProvider := gtk.NewCSSProvider()
	cssProvider.LoadFromString(`
		/* Dim label styling */
		.dim-label {
			opacity: 0.65;
		}
	`)
	gtk.StyleContextAddProviderForDisplay(
		gdk.DisplayGetDefault(),
		cssProvider,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)

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

	helpLabel := gtk.NewLabel("?: help  j/k: nav")
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
	keyController := setupKeyboardHandler(cfg, fileView, pathLabel, statusLabel, sortLabel, window, hyprState)
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
