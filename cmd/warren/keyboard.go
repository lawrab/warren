// Keyboard event handling and shortcuts setup.
// This file contains keyboard controller setup, key binding matching logic,
// and application-level shortcuts configuration.
//
//nolint:staticcheck // gtk.Dialog is deprecated in GTK4 but still functional in our version
package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/lawrab/warren/internal/config"
	"github.com/lawrab/warren/internal/fileops"
	"github.com/lawrab/warren/internal/ui"
	"github.com/lawrab/warren/pkg/models"
)

// setupKeyboardHandler creates and configures the keyboard event controller.
//
//nolint:gocyclo // Keyboard handler naturally has high complexity due to many shortcuts
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

		if keyMatchesConfig(keyval, cfg.Keybindings.Yank) {
			selected := fileView.GetSelected()
			if selected != nil {
				// Toggle yank: if already yanked, unyank it
				if fileView.IsYanked(selected.Path) {
					fileView.ClearYanked()
					statusLabel.SetText(fmt.Sprintf("Unyanked: %s", selected.Name))
				} else {
					fileView.YankSelected()
					statusLabel.SetText(fmt.Sprintf("Yanked: %s", selected.Name))
				}
			}
			return true
		}

		if keyMatchesConfig(keyval, cfg.Keybindings.Delete) {
			selected := fileView.GetSelected()
			if selected != nil {
				showDeleteDialog(window, fileView, selected, statusLabel, pathLabel, hyprState)
			}
			return true
		}

		if keyMatchesConfig(keyval, cfg.Keybindings.Paste) {
			yanked := fileView.GetYanked()
			if len(yanked) > 0 {
				showPasteDialog(window, fileView, yanked, statusLabel, pathLabel, hyprState)
			} else {
				statusLabel.SetText("No files yanked")
			}
			return true
		}

		if keyMatchesConfig(keyval, cfg.Keybindings.Rename) {
			selected := fileView.GetSelected()
			if selected != nil {
				showRenameDialog(window, fileView, selected, statusLabel, pathLabel, hyprState)
			}
			return true
		}

		if keyMatchesConfig(keyval, cfg.Keybindings.ShowHelp) {
			showShortcutsWindow(window, cfg)
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

// showDeleteDialog shows a confirmation dialog before deleting a file.
func showDeleteDialog(window *gtk.ApplicationWindow, fileView *ui.FileView, file *models.FileInfo, statusLabel, pathLabel *gtk.Label, hyprState *hyprlandState) {
	dialog := gtk.NewDialog()
	dialog.SetTitle("Delete File")
	dialog.SetTransientFor(&window.Window)
	dialog.SetModal(true)

	// Add message label
	label := gtk.NewLabel(fmt.Sprintf("Delete %s?\n\nThis will permanently delete:\n%s\n\nPress 'y' to confirm or 'n' to cancel", file.Name, file.Path))
	label.SetMarginTop(12)
	label.SetMarginBottom(12)
	label.SetMarginStart(12)
	label.SetMarginEnd(12)

	box := dialog.ContentArea()
	box.Append(label)

	// Add buttons
	dialog.AddButton("Cancel (n)", int(gtk.ResponseCancel))
	dialog.AddButton("Delete (y)", int(gtk.ResponseOK))
	dialog.SetDefaultResponse(int(gtk.ResponseCancel))

	// Add keyboard event controller for y/n shortcuts
	keyController := gtk.NewEventControllerKey()
	keyController.ConnectKeyPressed(func(keyval uint, _ uint, _ gdk.ModifierType) bool {
		switch keyval {
		case gdk.KEY_y, gdk.KEY_Y:
			dialog.Response(int(gtk.ResponseOK))
			return true
		case gdk.KEY_n, gdk.KEY_N, gdk.KEY_Escape:
			dialog.Response(int(gtk.ResponseCancel))
			return true
		}
		return false
	})
	dialog.AddController(keyController)

	dialog.ConnectResponse(func(responseID int) {
		dialog.Destroy()

		if responseID == int(gtk.ResponseOK) {
			// Delete the file using our fileops backend
			op := fileops.Delete(file.Path, nil)

			// Wait for operation to complete
			go func() {
				// Simple polling - in production would use channels
				for {
					time.Sleep(50 * time.Millisecond)
					if op.Status != fileops.StatusPending && op.Status != fileops.StatusRunning {
						break
					}
				}

				// Update UI on GTK thread
				glib.IdleAdd(func() {
					if op.Status == fileops.StatusCompleted {
						statusLabel.SetText(fmt.Sprintf("Deleted: %s", file.Name))
						// Reload directory
						_ = fileView.LoadDirectory(fileView.GetCurrentPath())
						pathLabel.SetText(fileView.GetCurrentPath())
						updateStatusBar(statusLabel, fileView)
						saveCurrentDirectoryToWorkspace(hyprState, fileView.GetCurrentPath())
					} else {
						statusLabel.SetText(fmt.Sprintf("Failed to delete: %v", op.Error))
					}
				})
			}()
		}
	})

	dialog.Show()
}

// showPasteDialog executes paste operation with progress feedback.
func showPasteDialog(_ *gtk.ApplicationWindow, fileView *ui.FileView, yanked []string, statusLabel, pathLabel *gtk.Label, hyprState *hyprlandState) {
	currentDir := fileView.GetCurrentPath()

	// Start copy operation
	op := fileops.CopyMultiple(yanked, currentDir, func(operation *fileops.Operation) {
		// Update UI on GTK thread
		glib.IdleAdd(func() {
			if operation.Status == fileops.StatusCompleted {
				statusLabel.SetText(fmt.Sprintf("Pasted %d file(s)", len(yanked)))
				// Reload directory
				_ = fileView.LoadDirectory(fileView.GetCurrentPath())
				pathLabel.SetText(fileView.GetCurrentPath())
				updateStatusBar(statusLabel, fileView)
				fileView.ClearYanked()
				saveCurrentDirectoryToWorkspace(hyprState, fileView.GetCurrentPath())
			} else if operation.Status == fileops.StatusFailed {
				statusLabel.SetText(fmt.Sprintf("Failed to paste: %v", operation.Error))
			}
		})
	})

	// For small files, this completes quickly. For large files, show progress
	_ = op // Operation runs in background
}

// showRenameDialog shows a dialog to rename a file.
func showRenameDialog(window *gtk.ApplicationWindow, fileView *ui.FileView, file *models.FileInfo, statusLabel, pathLabel *gtk.Label, hyprState *hyprlandState) {
	dialog := gtk.NewDialog()
	dialog.SetTitle("Rename File")
	dialog.SetTransientFor(&window.Window)
	dialog.SetModal(true)

	// Add entry for new name
	entry := gtk.NewEntry()
	entry.SetText(file.Name)
	entry.SetActivatesDefault(true)

	box := dialog.ContentArea()
	box.SetMarginTop(12)
	box.SetMarginBottom(12)
	box.SetMarginStart(12)
	box.SetMarginEnd(12)
	box.Append(entry)

	// Add buttons
	dialog.AddButton("Cancel", int(gtk.ResponseCancel))
	dialog.AddButton("Rename", int(gtk.ResponseOK))
	dialog.SetDefaultResponse(int(gtk.ResponseOK))

	dialog.ConnectResponse(func(responseID int) {
		newName := entry.Text()
		dialog.Destroy()

		if responseID == int(gtk.ResponseOK) && newName != "" && newName != file.Name {
			newPath := filepath.Join(fileView.GetCurrentPath(), newName)
			op := fileops.Rename(file.Path, newPath, nil)

			// Wait for operation
			go func() {
				for {
					time.Sleep(50 * time.Millisecond)
					if op.Status != fileops.StatusPending && op.Status != fileops.StatusRunning {
						break
					}
				}

				glib.IdleAdd(func() {
					if op.Status == fileops.StatusCompleted {
						statusLabel.SetText(fmt.Sprintf("Renamed to: %s", newName))
						_ = fileView.LoadDirectory(fileView.GetCurrentPath())
						pathLabel.SetText(fileView.GetCurrentPath())
						updateStatusBar(statusLabel, fileView)
						saveCurrentDirectoryToWorkspace(hyprState, fileView.GetCurrentPath())
					} else {
						statusLabel.SetText(fmt.Sprintf("Failed to rename: %v", op.Error))
					}
				})
			}()
		}
	})

	dialog.Show()
}

// showShortcutsWindow shows a dialog with all keyboard shortcuts.
func showShortcutsWindow(window *gtk.ApplicationWindow, cfg *config.Config) {
	dialog := gtk.NewDialog()
	dialog.SetTitle("Keyboard Shortcuts")
	dialog.SetTransientFor(&window.Window)
	dialog.SetModal(true)
	dialog.SetDefaultSize(500, 600)

	// Create scrolled window for shortcuts
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetVExpand(true)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

	// Create box to hold all shortcuts
	box := gtk.NewBox(gtk.OrientationVertical, 12)
	box.SetMarginTop(12)
	box.SetMarginBottom(12)
	box.SetMarginStart(12)
	box.SetMarginEnd(12)

	// Helper function to add a section
	addSection := func(title string, shortcuts map[string]string) {
		// Section header
		header := gtk.NewLabel(title)
		header.SetXAlign(0)
		header.SetMarkup(fmt.Sprintf("<b>%s</b>", title))
		header.SetMarginTop(6)
		box.Append(header)

		// Add shortcuts
		for key, desc := range shortcuts {
			shortcutBox := gtk.NewBox(gtk.OrientationHorizontal, 12)

			keyLabel := gtk.NewLabel(key)
			keyLabel.SetXAlign(0)
			keyLabel.SetWidthChars(15)
			keyLabel.AddCSSClass("dim-label")
			shortcutBox.Append(keyLabel)

			descLabel := gtk.NewLabel(desc)
			descLabel.SetXAlign(0)
			descLabel.SetHExpand(true)
			shortcutBox.Append(descLabel)

			box.Append(shortcutBox)
		}
	}

	// Navigation shortcuts
	addSection("Navigation", map[string]string{
		cfg.Keybindings.NavigateUp + "/k":   "Move up",
		cfg.Keybindings.NavigateDown + "/j": "Move down",
		cfg.Keybindings.ParentDir + "/h":    "Parent directory",
		cfg.Keybindings.EnterDir + "/l":     "Enter directory / Open file",
	})

	// File operations
	addSection("File Operations", map[string]string{
		cfg.Keybindings.Yank:   "Yank (copy) file / Unyank if already yanked",
		cfg.Keybindings.Paste:  "Paste yanked files",
		cfg.Keybindings.Delete: "Delete file (y/n to confirm)",
		cfg.Keybindings.Rename: "Rename file",
	})

	// View options
	addSection("View", map[string]string{
		cfg.Keybindings.ToggleHidden:    "Toggle hidden files",
		cfg.Keybindings.CycleSortMode:   "Cycle sort mode",
		cfg.Keybindings.ToggleSortOrder: "Toggle sort order",
	})

	// Application
	addSection("Application", map[string]string{
		cfg.Keybindings.ShowHelp: "Show this help",
		cfg.Keybindings.Quit:     "Quit",
		"Ctrl+Q":                 "Quit (alternative)",
	})

	scrolled.SetChild(box)
	dialog.ContentArea().Append(scrolled)

	// Close button
	dialog.AddButton("Close", int(gtk.ResponseClose))
	dialog.SetDefaultResponse(int(gtk.ResponseClose))

	dialog.ConnectResponse(func(_ int) {
		dialog.Destroy()
	})

	dialog.Show()
}
