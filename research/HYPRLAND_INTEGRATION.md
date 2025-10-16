# Hyprland Integration Research

## Overview

Hyprland provides a powerful IPC (Inter-Process Communication) system that allows external applications to query state and send commands. Warren will leverage this for deep integration.

## Hyprland IPC Basics

### Socket Location

Hyprland creates two Unix sockets per instance:

```bash
# Socket paths (where SIGNATURE is unique per instance)
/tmp/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket.sock      # Commands
/tmp/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket2.sock     # Events
```

The signature is available via environment variable:
```go
signature := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
socketPath := fmt.Sprintf("/tmp/hypr/%s/.socket.sock", signature)
```

### Detection

Check if running under Hyprland:
```go
func IsHyprland() bool {
    return os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != ""
}
```

## Command Socket (.socket.sock)

### Protocol

- Simple text-based protocol
- Connect via Unix socket
- Send commands, receive JSON responses
- Each command is a request-response

### Basic Client Implementation

```go
package hyprland

import (
    "encoding/json"
    "fmt"
    "net"
    "os"
)

type Client struct {
    socketPath string
}

func New() (*Client, error) {
    sig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
    if sig == "" {
        return nil, fmt.Errorf("not running under Hyprland")
    }

    return &Client{
        socketPath: fmt.Sprintf("/tmp/hypr/%s/.socket.sock", sig),
    }, nil
}

func (c *Client) sendCommand(cmd string) ([]byte, error) {
    conn, err := net.Dial("unix", c.socketPath)
    if err != nil {
        return nil, err
    }
    defer conn.Close()

    _, err = conn.Write([]byte(cmd))
    if err != nil {
        return nil, err
    }

    buf := make([]byte, 65536) // 64KB buffer
    n, err := conn.Read(buf)
    if err != nil {
        return nil, err
    }

    return buf[:n], nil
}
```

### Key Commands

#### Get Active Workspace
```go
type Workspace struct {
    ID             int    `json:"id"`
    Name           string `json:"name"`
    Monitor        string `json:"monitor"`
    Windows        int    `json:"windows"`
    HasFullscreen  bool   `json:"hasfullscreen"`
    LastWindow     string `json:"lastwindow"`
}

func (c *Client) GetActiveWorkspace() (*Workspace, error) {
    resp, err := c.sendCommand("j/activeworkspace")
    if err != nil {
        return nil, err
    }

    var ws Workspace
    err = json.Unmarshal(resp, &ws)
    return &ws, err
}
```

#### List All Workspaces
```go
func (c *Client) GetWorkspaces() ([]Workspace, error) {
    resp, err := c.sendCommand("j/workspaces")
    if err != nil {
        return nil, err
    }

    var workspaces []Workspace
    err = json.Unmarshal(resp, &workspaces)
    return workspaces, err
}
```

#### Get Active Window
```go
type Window struct {
    Address   string  `json:"address"`
    At        [2]int  `json:"at"`
    Size      [2]int  `json:"size"`
    Workspace struct {
        ID   int    `json:"id"`
        Name string `json:"name"`
    } `json:"workspace"`
    Class     string `json:"class"`
    Title     string `json:"title"`
    PID       int    `json:"pid"`
}

func (c *Client) GetActiveWindow() (*Window, error) {
    resp, err := c.sendCommand("j/activewindow")
    if err != nil {
        return nil, err
    }

    var win Window
    err = json.Unmarshal(resp, &win)
    return &win, err
}
```

#### Dispatch Commands
```go
// Switch to workspace
func (c *Client) SwitchWorkspace(id int) error {
    cmd := fmt.Sprintf("dispatch workspace %d", id)
    _, err := c.sendCommand(cmd)
    return err
}

// Focus window by address
func (c *Client) FocusWindow(address string) error {
    cmd := fmt.Sprintf("dispatch focuswindow address:%s", address)
    _, err := c.sendCommand(cmd)
    return err
}

// Move window to workspace
func (c *Client) MoveToWorkspace(id int) error {
    cmd := fmt.Sprintf("dispatch movetoworkspace %d", id)
    _, err := c.sendCommand(cmd)
    return err
}
```

## Event Socket (.socket2.sock)

### Protocol

- Long-lived connection
- Receive events as they happen
- Text-based, line-delimited
- Format: `EVENT>>DATA`

### Event Types

```
workspace>>WORKSPACE_NAME
focusedmon>>MONITOR_NAME,WORKSPACE_NAME
activewindow>>WINDOW_CLASS,WINDOW_TITLE
activewindowv2>>WINDOW_ADDRESS
fullscreen>>0/1
monitorremoved>>MONITOR_NAME
monitoradded>>MONITOR_NAME
createworkspace>>WORKSPACE_NAME
destroyworkspace>>WORKSPACE_NAME
moveworkspace>>WORKSPACE_NAME,MONITOR_NAME
activelayout>>KEYBOARD_NAME,LAYOUT_NAME
openwindow>>WINDOW_ADDRESS,WORKSPACE_NAME,CLASS,TITLE
closewindow>>WINDOW_ADDRESS
movewindow>>WINDOW_ADDRESS,WORKSPACE_NAME
openlayer>>NAMESPACE
closelayer>>NAMESPACE
```

### Event Listener Implementation

```go
type Event struct {
    Type string
    Data string
}

type EventHandler func(Event)

func (c *Client) ListenEvents(handler EventHandler) error {
    sig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
    socketPath := fmt.Sprintf("/tmp/hypr/%s/.socket2.sock", sig)

    conn, err := net.Dial("unix", socketPath)
    if err != nil {
        return err
    }
    defer conn.Close()

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

    return scanner.Err()
}
```

### Example Usage
```go
client, _ := hyprland.New()

go client.ListenEvents(func(event hyprland.Event) {
    switch event.Type {
    case "workspace":
        fmt.Printf("Switched to workspace: %s\n", event.Data)
        // Update Warren's current directory based on workspace

    case "activewindow":
        parts := strings.SplitN(event.Data, ",", 2)
        class := parts[0]
        title := parts[1]
        fmt.Printf("Active window: %s - %s\n", class, title)

    case "openwindow":
        fmt.Printf("Window opened: %s\n", event.Data)
    }
})
```

## Warren Integration Use Cases

### 1. Per-Workspace Directory Memory

**Goal:** Remember last directory per workspace

```go
type WorkspaceMemory struct {
    workspaceDirs map[int]string
    mu            sync.RWMutex
}

func (app *Application) setupWorkspaceMemory() {
    app.hyprland.ListenEvents(func(event Event) {
        if event.Type == "workspace" {
            workspaceID := parseWorkspaceID(event.Data)

            // Save current directory for old workspace
            app.memory.Set(app.currentWorkspace, app.currentDir)

            // Load directory for new workspace
            if dir := app.memory.Get(workspaceID); dir != "" {
                app.NavigateTo(dir)
            }

            app.currentWorkspace = workspaceID
        }
    })
}
```

### 2. Open Files in Specific Workspaces

**Goal:** Open file in target workspace

```go
func (app *Application) OpenFileInWorkspace(path string, workspaceID int) error {
    // Get current workspace
    currentWS, _ := app.hyprland.GetActiveWorkspace()

    // Switch to target workspace
    if currentWS.ID != workspaceID {
        app.hyprland.SwitchWorkspace(workspaceID)
        time.Sleep(100 * time.Millisecond) // Wait for switch
    }

    // Open file
    cmd := exec.Command("xdg-open", path)
    return cmd.Start()
}
```

### 3. Window Rules Suggestions

**Goal:** Suggest optimal window rules for Warren

```hyprlang
# Add to ~/.config/hypr/hyprland.conf

# Float Warren by default
windowrule = float, ^(warren)$

# Set size (1200x800)
windowrule = size 1200 800, ^(warren)$

# Center on screen
windowrule = center, ^(warren)$

# Start on specific workspace (optional)
windowrule = workspace 9, ^(warren)$

# Optionally: Make it pseudo-tiled
windowrulev2 = pseudo, class:^(warren)$

# Pin to all workspaces (if desired)
windowrule = pin, ^(warren)$
```

### 4. Workspace-Aware Quick Actions

**Goal:** Quick file operations based on workspace context

```go
// Open terminal in current directory on current workspace
func (app *Application) OpenTerminalHere() error {
    ws, _ := app.hyprland.GetActiveWorkspace()

    cmd := exec.Command("kitty", "--directory", app.currentDir)
    cmd.Env = append(os.Environ(),
        fmt.Sprintf("HYPRLAND_WORKSPACE=%d", ws.ID))

    return cmd.Start()
}

// Quick jump to workspace-specific directory
func (app *Application) JumpToWorkspaceProject(wsID int) error {
    // Map workspaces to project directories
    projectDirs := map[int]string{
        1: "/home/user/projects/frontend",
        2: "/home/user/projects/backend",
        3: "/home/user/documents",
        // ...
    }

    if dir, ok := projectDirs[wsID]; ok {
        app.hyprland.SwitchWorkspace(wsID)
        app.NavigateTo(dir)
    }

    return nil
}
```

## Error Handling

### Graceful Degradation

Warren should work even when not in Hyprland:

```go
func (app *Application) initialize() {
    // Try to connect to Hyprland
    if client, err := hyprland.New(); err == nil {
        app.hyprland = client
        app.setupHyprlandIntegration()
        log.Println("Hyprland integration enabled")
    } else {
        log.Println("Not running in Hyprland, IPC disabled")
        app.hyprland = nil
    }
}

func (app *Application) isHyprlandAvailable() bool {
    return app.hyprland != nil
}

// Check before Hyprland-specific features
func (app *Application) doHyprlandThing() {
    if !app.isHyprlandAvailable() {
        log.Println("Feature requires Hyprland")
        return
    }
    // ... do Hyprland thing
}
```

### Connection Recovery

```go
func (c *Client) sendCommandWithRetry(cmd string) ([]byte, error) {
    maxRetries := 3

    for i := 0; i < maxRetries; i++ {
        resp, err := c.sendCommand(cmd)
        if err == nil {
            return resp, nil
        }

        log.Printf("Command failed (attempt %d/%d): %v", i+1, maxRetries, err)
        time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
    }

    return nil, fmt.Errorf("command failed after %d retries", maxRetries)
}
```

## Configuration

### Enable/Disable Integration

```toml
# config.toml
[hyprland]
enabled = true
workspace_memory = true
auto_window_rules = true

# Per-workspace directory mappings
[hyprland.workspace_dirs]
1 = "~/projects/web"
2 = "~/projects/go"
3 = "~/documents"
9 = "~/downloads"
```

## Performance Considerations

1. **Event Listener**: Runs in goroutine, minimal overhead
2. **Command Latency**: ~1-5ms per IPC call
3. **Connection Pooling**: Consider keeping connection open for commands
4. **Rate Limiting**: Debounce rapid events (e.g., workspace switches)

## Testing Strategy

### Unit Tests (Mock IPC)
```go
type MockHyprlandClient struct {
    workspaces []Workspace
}

func (m *MockHyprlandClient) GetActiveWorkspace() (*Workspace, error) {
    return &m.workspaces[0], nil
}
```

### Integration Tests
```go
func TestHyprlandIPC(t *testing.T) {
    if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
        t.Skip("Not running under Hyprland")
    }

    client, err := hyprland.New()
    require.NoError(t, err)

    ws, err := client.GetActiveWorkspace()
    require.NoError(t, err)
    assert.NotNil(t, ws)
}
```

### Manual Testing
- Test in real Hyprland environment
- Multiple workspaces, monitors
- Edge cases: workspace deletion, monitor disconnect

## Resources

- Hyprland Wiki IPC: https://wiki.hyprland.org/IPC/
- Hyprland GitHub: https://github.com/hyprwm/Hyprland
- Example implementations: Look at other Hyprland tools (waybar, eww configs)

---

**Next:** Start implementing basic IPC client in Phase 2.
