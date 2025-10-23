// Hyprland integration setup and event handling.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/lawrab/warren/internal/config"
	"github.com/lawrab/warren/internal/hyprland"
	"github.com/lawrab/warren/internal/ui"
)

// hyprlandState holds Hyprland integration state.
type hyprlandState struct {
	client *hyprland.Client
	memory *hyprland.WorkspaceMemory
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
			// Handle both workspace switches and Warren being moved between workspaces
			if (event.Type == "workspace" || event.Type == "movewindow") && cfg.Hyprland.AutoSwitch && hs.memory != nil {
				var workspaceID int
				var err error

				switch event.Type {
				case "workspace":
					// workspace>>2 - user switched to workspace 2
					workspaceID, err = strconv.Atoi(strings.TrimSpace(event.Data))
				case "movewindow":
					// movewindow>>windowaddr,3 - window moved to workspace 3
					// Parse: "122e5f40,3" -> workspace ID is 3
					parts := strings.Split(event.Data, ",")
					if len(parts) < 2 {
						return
					}
					workspaceID, err = strconv.Atoi(strings.TrimSpace(parts[1]))
				}

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
