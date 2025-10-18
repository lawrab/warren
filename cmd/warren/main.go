package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
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

	app := gtk.NewApplication(appID, 0)
	app.ConnectActivate(func() { activate(app) })

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate(app *gtk.Application) {
	// Create main window
	window := gtk.NewApplicationWindow(app)
	window.SetTitle(fmt.Sprintf("Warren %s", version.Short()))
	window.SetDefaultSize(1000, 700)

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

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/"
	}

	// Load initial directory
	if err := fileView.LoadDirectory(homeDir); err != nil {
		log.Printf("Failed to load directory: %v", err)
		statusLabel.SetText(err.Error())
	} else {
		pathLabel.SetText(fileView.GetCurrentPath())
		updateStatusBar(statusLabel, fileView)
	}

	// Set up keyboard event controller
	keyController := gtk.NewEventControllerKey()
	keyController.ConnectKeyPressed(func(keyval uint, keycode uint, state gdk.ModifierType) bool {
		// Handle keyboard navigation
		switch keyval {
		case gdk.KEY_j, gdk.KEY_Down:
			fileView.SelectNext()
			updateStatusBar(statusLabel, fileView)
			return true

		case gdk.KEY_k, gdk.KEY_Up:
			fileView.SelectPrevious()
			updateStatusBar(statusLabel, fileView)
			return true

		case gdk.KEY_h, gdk.KEY_Left, gdk.KEY_BackSpace:
			if err := fileView.NavigateUp(); err != nil {
				statusLabel.SetText(err.Error())
			} else {
				pathLabel.SetText(fileView.GetCurrentPath())
				updateStatusBar(statusLabel, fileView)
			}
			return true

		case gdk.KEY_l, gdk.KEY_Right, gdk.KEY_Return:
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

		case gdk.KEY_period:
			if err := fileView.ToggleHidden(); err != nil {
				statusLabel.SetText(err.Error())
			} else {
				updateStatusBar(statusLabel, fileView)
			}
			return true

		case gdk.KEY_q:
			window.Close()
			return true
		}

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
	quitAction.ConnectActivate(func(parameter *glib.Variant) {
		window.Close()
	})
	app.AddAction(quitAction)
	app.SetAccelsForAction("app.quit", []string{"<Ctrl>Q"})
}
