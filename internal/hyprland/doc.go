// Package hyprland provides IPC communication with the Hyprland window manager.
//
// This package implements a client for Hyprland's IPC protocol, allowing Warren to:
//   - Query active workspace and window information
//   - Listen for Hyprland events (workspace changes, window events, etc.)
//   - Send commands to Hyprland
//   - Maintain per-workspace directory memory
//
// The package gracefully handles non-Hyprland environments by providing detection
// and error handling that allows Warren to function without Hyprland integration.
//
// Basic usage:
//
//	client, err := hyprland.New()
//	if err != nil {
//	    // Not running in Hyprland or IPC unavailable
//	    return
//	}
//
//	workspace, err := client.GetActiveWorkspace()
//	if err != nil {
//	    // Handle error
//	}
//
// Event listening:
//
//	client.ListenEvents(func(event Event) {
//	    if event.Type == "workspace" {
//	        // Handle workspace change
//	    }
//	})
package hyprland
