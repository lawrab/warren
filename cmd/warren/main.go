package main

import (
	"os"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const appID = "com.lawrab.warren"

func main() {
	app := gtk.NewApplication(appID, 0)
	app.ConnectActivate(func() { activate(app) })

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate(app *gtk.Application) {
	// Create main window
	window := gtk.NewApplicationWindow(app)
	window.SetTitle("Warren")
	window.SetDefaultSize(800, 600)

	// Create a header bar
	headerBar := gtk.NewHeaderBar()
	headerBar.SetShowTitleButtons(true)
	window.SetTitlebar(headerBar)

	// Create main box
	box := gtk.NewBox(gtk.OrientationVertical, 12)
	box.SetMarginTop(24)
	box.SetMarginBottom(24)
	box.SetMarginStart(24)
	box.SetMarginEnd(24)

	// Welcome label
	label := gtk.NewLabel("")
	label.SetMarkup("<span size='x-large' weight='bold'>🐰 Warren File Manager</span>")
	box.Append(label)

	// Subtitle
	subtitle := gtk.NewLabel("For Ann - Development Environment Ready")
	subtitle.AddCSSClass("dim-label")
	box.Append(subtitle)

	// Info label
	info := gtk.NewLabel("\nPress Ctrl+Q to quit")
	info.SetMarginTop(24)
	box.Append(info)

	// Add box to window
	window.SetChild(box)

	// Keyboard shortcuts
	setupShortcuts(app, window)

	// Show window
	window.Present()
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
