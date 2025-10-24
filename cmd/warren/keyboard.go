// Keyboard event handling and shortcuts setup.
// This file contains keyboard controller setup, key binding matching logic,
// and application-level shortcuts configuration.
package main

import (
	"fmt"
	"log"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/lawrab/warren/internal/config"
	"github.com/lawrab/warren/internal/fileops"
	"github.com/lawrab/warren/internal/ui"
)

// setupKeyboardHandler creates and configures the keyboard event controller.
func setupKeyboardHandler(cfg *config.Config, fileView *ui.FileView, pathLabel, statusLabel, sortLabel *gtk.Label, window *gtk.ApplicationWindow, hyprState *hyprlandState) *gtk.EventControllerKey {
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
	return keyController
}

// setupShortcuts configures application-level keyboard shortcuts.
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
