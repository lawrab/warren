package hyprland

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIsHyprland(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{
			name:     "with signature",
			envValue: "test_signature_123",
			want:     true,
		},
		{
			name:     "without signature",
			envValue: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env value
			orig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
			defer os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", orig)

			// Set test env value
			if tt.envValue != "" {
				os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", tt.envValue)
			} else {
				os.Unsetenv("HYPRLAND_INSTANCE_SIGNATURE")
			}

			if got := IsHyprland(); got != tt.want {
				t.Errorf("IsHyprland() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		wantErr   bool
		errString string
	}{
		{
			name:      "no signature env var",
			envValue:  "",
			wantErr:   true,
			errString: "not running under Hyprland",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env value
			orig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
			defer os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", orig)

			// Set test env value
			if tt.envValue != "" {
				os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", tt.envValue)
			} else {
				os.Unsetenv("HYPRLAND_INSTANCE_SIGNATURE")
			}

			client, err := New()

			if tt.wantErr {
				if err == nil {
					t.Errorf("New() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errString != "" && !strings.Contains(err.Error(), tt.errString) {
					t.Errorf("New() error = %v, want error containing %v", err, tt.errString)
				}
				return
			}

			if err != nil {
				t.Errorf("New() unexpected error = %v", err)
				return
			}

			if client == nil {
				t.Errorf("New() returned nil client")
			}
		})
	}
}

func TestNewWithMockSocket(t *testing.T) {
	// Create temporary directory for mock socket
	tmpDir := t.TempDir()
	signature := "test_sig_123"

	// Create the hypr directory structure
	hyprDir := filepath.Join(tmpDir, "hypr", signature)
	if err := os.MkdirAll(hyprDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a mock socket file
	socketPath := filepath.Join(hyprDir, ".socket.sock")
	if err := os.WriteFile(socketPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create mock socket: %v", err)
	}

	// Save and restore original env
	origSig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	origXDG := os.Getenv("XDG_RUNTIME_DIR")
	defer func() {
		os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", origSig)
		if origXDG != "" {
			os.Setenv("XDG_RUNTIME_DIR", origXDG)
		} else {
			os.Unsetenv("XDG_RUNTIME_DIR")
		}
	}()

	// Set test environment to use tmpDir
	os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", signature)
	os.Setenv("XDG_RUNTIME_DIR", tmpDir)

	// Create client - should succeed with mocked XDG_RUNTIME_DIR
	client, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Verify socket paths use XDG_RUNTIME_DIR
	expectedCommandSocket := filepath.Join(tmpDir, "hypr", signature, ".socket.sock")
	expectedEventSocket := filepath.Join(tmpDir, "hypr", signature, ".socket2.sock")

	if client.commandSocket != expectedCommandSocket {
		t.Errorf("commandSocket = %v, want %v", client.commandSocket, expectedCommandSocket)
	}
	if client.eventSocket != expectedEventSocket {
		t.Errorf("eventSocket = %v, want %v", client.eventSocket, expectedEventSocket)
	}
}

// Mock server for testing IPC commands
func setupMockCommandServer(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	// Create Unix socket server
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create mock server: %v", err)
	}

	// Handle connections in background
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // Server closed
			}

			go handleMockConnection(conn)
		}
	}()

	cleanup := func() {
		listener.Close()
		os.Remove(socketPath)
	}

	return socketPath, cleanup
}

func handleMockConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	cmd := string(buf[:n])

	// Mock responses based on command
	var response []byte
	switch cmd {
	case "j/activeworkspace":
		ws := Workspace{
			ID:            1,
			Name:          "1",
			Monitor:       "DP-1",
			Windows:       3,
			HasFullscreen: false,
			LastWindow:    "0x123456",
		}
		response, _ = json.Marshal(ws)

	case "j/workspaces":
		workspaces := []Workspace{
			{ID: 1, Name: "1", Monitor: "DP-1", Windows: 3},
			{ID: 2, Name: "2", Monitor: "DP-1", Windows: 1},
		}
		response, _ = json.Marshal(workspaces)

	case "j/activewindow":
		win := Window{
			Address: "0x123456",
			At:      [2]int{100, 100},
			Size:    [2]int{800, 600},
			Class:   "kitty",
			Title:   "Terminal",
			PID:     12345,
		}
		win.Workspace.ID = 1
		win.Workspace.Name = "1"
		response, _ = json.Marshal(win)

	default:
		response = []byte("ok")
	}

	conn.Write(response)
}

func TestClient_GetActiveWorkspace(t *testing.T) {
	socketPath, cleanup := setupMockCommandServer(t)
	defer cleanup()

	client := &Client{
		commandSocket: socketPath,
	}

	ws, err := client.GetActiveWorkspace()
	if err != nil {
		t.Fatalf("GetActiveWorkspace() error = %v", err)
	}

	if ws.ID != 1 {
		t.Errorf("Workspace ID = %d, want 1", ws.ID)
	}
	if ws.Monitor != "DP-1" {
		t.Errorf("Workspace Monitor = %s, want DP-1", ws.Monitor)
	}
}

func TestClient_GetWorkspaces(t *testing.T) {
	socketPath, cleanup := setupMockCommandServer(t)
	defer cleanup()

	client := &Client{
		commandSocket: socketPath,
	}

	workspaces, err := client.GetWorkspaces()
	if err != nil {
		t.Fatalf("GetWorkspaces() error = %v", err)
	}

	if len(workspaces) != 2 {
		t.Errorf("Got %d workspaces, want 2", len(workspaces))
	}

	if workspaces[0].ID != 1 {
		t.Errorf("First workspace ID = %d, want 1", workspaces[0].ID)
	}
	if workspaces[1].ID != 2 {
		t.Errorf("Second workspace ID = %d, want 2", workspaces[1].ID)
	}
}

func TestClient_GetActiveWindow(t *testing.T) {
	socketPath, cleanup := setupMockCommandServer(t)
	defer cleanup()

	client := &Client{
		commandSocket: socketPath,
	}

	win, err := client.GetActiveWindow()
	if err != nil {
		t.Fatalf("GetActiveWindow() error = %v", err)
	}

	if win.Class != "kitty" {
		t.Errorf("Window Class = %s, want kitty", win.Class)
	}
	if win.Title != "Terminal" {
		t.Errorf("Window Title = %s, want Terminal", win.Title)
	}
	if win.Workspace.ID != 1 {
		t.Errorf("Window Workspace ID = %d, want 1", win.Workspace.ID)
	}
}

// Mock event server
func setupMockEventServer(t *testing.T, events []string) (string, func()) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "events.sock")

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("Failed to create mock event server: %v", err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Send mock events
		for _, event := range events {
			conn.Write([]byte(event + "\n"))
			time.Sleep(10 * time.Millisecond)
		}
	}()

	cleanup := func() {
		listener.Close()
		os.Remove(socketPath)
	}

	return socketPath, cleanup
}

func TestClient_ListenEvents(t *testing.T) {
	mockEvents := []string{
		"workspace>>2",
		"activewindow>>kitty,Terminal",
		"fullscreen>>1",
	}

	socketPath, cleanup := setupMockEventServer(t, mockEvents)
	defer cleanup()

	client := &Client{
		eventSocket: socketPath,
	}

	receivedEvents := make([]Event, 0)
	done := make(chan bool)

	go func() {
		err := client.ListenEvents(func(event Event) {
			receivedEvents = append(receivedEvents, event)
			if len(receivedEvents) == 3 {
				done <- true
			}
		})
		if err != nil && err.Error() != "event listener error: EOF" {
			t.Errorf("ListenEvents() error = %v", err)
		}
	}()

	// Wait for events or timeout
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for events")
	}

	if len(receivedEvents) != 3 {
		t.Fatalf("Received %d events, want 3", len(receivedEvents))
	}

	// Verify events
	if receivedEvents[0].Type != "workspace" || receivedEvents[0].Data != "2" {
		t.Errorf("Event 0 = %+v, want workspace>>2", receivedEvents[0])
	}
	if receivedEvents[1].Type != "activewindow" || receivedEvents[1].Data != "kitty,Terminal" {
		t.Errorf("Event 1 = %+v, want activewindow>>kitty,Terminal", receivedEvents[1])
	}
	if receivedEvents[2].Type != "fullscreen" || receivedEvents[2].Data != "1" {
		t.Errorf("Event 2 = %+v, want fullscreen>>1", receivedEvents[2])
	}
}

func TestClient_Dispatch(t *testing.T) {
	socketPath, cleanup := setupMockCommandServer(t)
	defer cleanup()

	client := &Client{
		commandSocket: socketPath,
	}

	err := client.Dispatch("workspace 3")
	if err != nil {
		t.Errorf("Dispatch() error = %v", err)
	}
}

func TestClient_SwitchWorkspace(t *testing.T) {
	socketPath, cleanup := setupMockCommandServer(t)
	defer cleanup()

	client := &Client{
		commandSocket: socketPath,
	}

	err := client.SwitchWorkspace(3)
	if err != nil {
		t.Errorf("SwitchWorkspace() error = %v", err)
	}
}
