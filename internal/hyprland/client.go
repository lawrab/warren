package hyprland

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	// Buffer size for IPC responses (64KB)
	ipcBufferSize = 65536
)

// Client provides IPC communication with Hyprland.
type Client struct {
	commandSocket string // Path to .socket.sock for commands
	eventSocket   string // Path to .socket2.sock for events
}

// Workspace represents a Hyprland workspace.
type Workspace struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Monitor       string `json:"monitor"`
	Windows       int    `json:"windows"`
	HasFullscreen bool   `json:"hasfullscreen"`
	LastWindow    string `json:"lastwindow"`
}

// Window represents a Hyprland window.
type Window struct {
	Address   string `json:"address"`
	At        [2]int `json:"at"`
	Size      [2]int `json:"size"`
	Workspace struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"workspace"`
	Class string `json:"class"`
	Title string `json:"title"`
	PID   int    `json:"pid"`
}

// Event represents a Hyprland event from the event socket.
type Event struct {
	Type string // Event type (e.g., "workspace", "activewindow")
	Data string // Event data
}

// EventHandler is a function that handles Hyprland events.
type EventHandler func(Event)

// IsHyprland checks if the current environment is running under Hyprland.
func IsHyprland() bool {
	return os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != ""
}

// getRuntimeDir returns the Hyprland runtime directory.
// Follows the same logic as hyprctl: checks XDG_RUNTIME_DIR first,
// then falls back to /run/user/$UID.
func getRuntimeDir() string {
	// Check XDG_RUNTIME_DIR first (standard)
	if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
		return fmt.Sprintf("%s/hypr", xdg)
	}

	// Fallback to /run/user/$UID (same as hyprctl)
	uid := os.Getuid()
	return fmt.Sprintf("/run/user/%d/hypr", uid)
}

// New creates a new Hyprland IPC client.
// Returns an error if not running under Hyprland.
func New() (*Client, error) {
	sig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if sig == "" {
		return nil, fmt.Errorf("not running under Hyprland (HYPRLAND_INSTANCE_SIGNATURE not set)")
	}

	runtimeDir := getRuntimeDir()
	commandSocket := fmt.Sprintf("%s/%s/.socket.sock", runtimeDir, sig)
	eventSocket := fmt.Sprintf("%s/%s/.socket2.sock", runtimeDir, sig)

	// Verify command socket exists
	if _, err := os.Stat(commandSocket); err != nil {
		return nil, fmt.Errorf("hyprland command socket not found: %w", err)
	}

	return &Client{
		commandSocket: commandSocket,
		eventSocket:   eventSocket,
	}, nil
}

// sendCommand sends a command to Hyprland's command socket and returns the response.
func (c *Client) sendCommand(cmd string) ([]byte, error) {
	conn, err := net.Dial("unix", c.commandSocket)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Hyprland socket: %w", err)
	}
	defer func() { _ = conn.Close() }()

	_, err = conn.Write([]byte(cmd))
	if err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	buf := make([]byte, ipcBufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return buf[:n], nil
}

// GetActiveWorkspace returns the currently active workspace.
func (c *Client) GetActiveWorkspace() (*Workspace, error) {
	resp, err := c.sendCommand("j/activeworkspace")
	if err != nil {
		return nil, err
	}

	var ws Workspace
	if err := json.Unmarshal(resp, &ws); err != nil {
		return nil, fmt.Errorf("failed to parse workspace data: %w", err)
	}

	return &ws, nil
}

// GetWorkspaces returns all workspaces.
func (c *Client) GetWorkspaces() ([]Workspace, error) {
	resp, err := c.sendCommand("j/workspaces")
	if err != nil {
		return nil, err
	}

	var workspaces []Workspace
	if err := json.Unmarshal(resp, &workspaces); err != nil {
		return nil, fmt.Errorf("failed to parse workspaces data: %w", err)
	}

	return workspaces, nil
}

// GetActiveWindow returns the currently active window.
func (c *Client) GetActiveWindow() (*Window, error) {
	resp, err := c.sendCommand("j/activewindow")
	if err != nil {
		return nil, err
	}

	var win Window
	if err := json.Unmarshal(resp, &win); err != nil {
		return nil, fmt.Errorf("failed to parse window data: %w", err)
	}

	return &win, nil
}

// ListenEvents listens for Hyprland events and calls the handler for each event.
// This is a blocking call that runs until an error occurs.
// Typically run in a goroutine.
func (c *Client) ListenEvents(handler EventHandler) error {
	conn, err := net.Dial("unix", c.eventSocket)
	if err != nil {
		return fmt.Errorf("failed to connect to Hyprland event socket: %w", err)
	}
	defer func() { _ = conn.Close() }()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ">>", 2)

		if len(parts) == 2 {
			event := Event{
				Type: parts[0],
				Data: parts[1],
			}
			handler(event)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("event listener error: %w", err)
	}

	return nil
}

// Dispatch sends a dispatch command to Hyprland.
func (c *Client) Dispatch(command string) error {
	cmd := fmt.Sprintf("dispatch %s", command)
	_, err := c.sendCommand(cmd)
	return err
}

// SwitchWorkspace switches to the specified workspace.
func (c *Client) SwitchWorkspace(id int) error {
	return c.Dispatch(fmt.Sprintf("workspace %d", id))
}
