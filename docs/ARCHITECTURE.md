# Architecture

## Overview

Warren follows standard Go project layout with clear separation of concerns. The architecture prioritizes simplicity, testability, and maintainability.

## Project Structure

```
warren/
├── cmd/
│   └── warren/
│       └── main.go                  # Application entry point
├── internal/                        # Private application code
│   ├── app/
│   │   ├── app.go                   # Application lifecycle
│   │   └── state.go                 # Global state management
│   ├── ui/
│   │   ├── window.go                # Main window
│   │   ├── fileview.go              # File list widget
│   │   ├── statusbar.go             # Status bar
│   │   ├── preview.go               # Preview pane
│   │   └── keybindings.go           # Keyboard shortcuts
│   ├── fileops/
│   │   ├── list.go                  # Directory listing
│   │   ├── operations.go            # Copy/move/delete
│   │   ├── watch.go                 # Filesystem watching
│   │   └── permissions.go           # Permission handling
│   ├── hyprland/
│   │   ├── ipc.go                   # IPC client
│   │   ├── events.go                # Event handling
│   │   └── workspace.go             # Workspace queries
│   └── config/
│       ├── config.go                # Configuration loading
│       ├── keymaps.go               # Keymap definitions
│       └── defaults.go              # Default settings
├── pkg/                             # Public libraries (if needed)
│   └── models/
│       ├── file.go                  # File/directory models
│       └── operation.go             # Operation types
├── testdata/                        # Test fixtures
│   ├── sample/
│   │   ├── file1.txt
│   │   └── subdir/
│   └── configs/
│       └── warren.toml
├── docs/                            # Documentation
├── go.mod
└── go.sum
```

## Package Responsibilities

### `cmd/warren`
**Purpose:** Application entry point

```go
// main.go
package main

import (
    "github.com/lawrab/warren/internal/app"
    "github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func main() {
    application := gtk.NewApplication("com.rabbets.warren", 0)
    application.ConnectActivate(func() {
        app.Run(application)
    })
    application.Run(os.Args)
}
```

**Responsibilities:**
- Initialize GTK application
- Handle command-line arguments
- Set up logging
- Delegate to internal/app

---

### `internal/app`
**Purpose:** Application lifecycle and state

```go
// app.go
package app

type Application struct {
    config      *config.Config
    hyprland    *hyprland.Client
    currentDir  string
    mainWindow  *ui.MainWindow
}

func Run(gtkApp *gtk.Application) {
    app := &Application{}
    app.loadConfig()
    app.connectHyprland()
    app.buildUI(gtkApp)
    app.setupEventHandlers()
}
```

**Responsibilities:**
- Coordinate between packages
- Manage application state
- Handle high-level application logic
- Connect UI, file operations, and Hyprland

---

### `internal/ui`
**Purpose:** GTK4 user interface components

```go
// window.go
package ui

type MainWindow struct {
    window      *gtk.ApplicationWindow
    fileView    *FileView
    previewPane *PreviewPane
    statusBar   *StatusBar
}

func NewMainWindow(app *gtk.Application) *MainWindow {
    // Create window and widgets
}
```

**Key Components:**

**FileView:** Directory listing widget
- Tree/list view of files
- Selection handling
- Sorting/filtering
- Icons and metadata display

**PreviewPane:** File preview panel
- Image previews
- Text file contents
- Video thumbnails
- PDF rendering (future)

**StatusBar:** Information display
- Current path
- Selection info
- Operation progress
- Disk space

**KeyBindings:** Keyboard shortcut handling
- Vim-style navigation
- Custom keybindings
- Context-aware shortcuts

**Responsibilities:**
- All GTK4 widgets and UI logic
- Event handling (clicks, keys)
- UI state management
- Theming and styling

---

### `internal/fileops`
**Purpose:** Filesystem operations

```go
// list.go
package fileops

type FileInfo struct {
    Name         string
    Path         string
    Size         int64
    IsDir        bool
    Permissions  os.FileMode
    ModTime      time.Time
    Symlink      string  // Target if symlink
}

func ListDirectory(path string) ([]FileInfo, error) {
    // List directory contents
}
```

**Key Functions:**

**Listing:**
- `ListDirectory()` - Get directory contents
- `GetFileInfo()` - Detailed file information
- `Search()` - Search for files

**Operations:**
- `Copy()` - Copy files/directories
- `Move()` - Move/rename
- `Delete()` - Delete with confirmation
- `CreateDir()` - Create directory
- `CreateFile()` - Create empty file

**Watching:**
- `WatchDirectory()` - Filesystem events
- Handle create/delete/modify events

**Utilities:**
- `FormatSize()` - Human-readable sizes
- `GetMimeType()` - Detect file types
- `IsHidden()` - Hidden file detection

**Responsibilities:**
- All filesystem interactions
- Safe file operations (atomic where possible)
- Error handling and validation
- Async operations via goroutines

---

### `internal/hyprland`
**Purpose:** Hyprland IPC integration

```go
// ipc.go
package hyprland

type Client struct {
    socketPath string
    conn       net.Conn
}

func New() (*Client, error) {
    signature := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
    // Connect to socket
}

func (c *Client) GetActiveWorkspace() (*Workspace, error) {
    // Query active workspace
}

func (c *Client) DispatchCommand(cmd string) error {
    // Send hyprctl command
}
```

**Key Components:**

**IPC Client:**
- Socket connection management
- Command/response handling
- Error recovery

**Events:**
- Subscribe to Hyprland events
- Workspace changes
- Window focus changes
- Monitor events

**Workspace Management:**
- Query workspace info
- Switch workspaces
- Get window list

**Responsibilities:**
- All Hyprland communication
- Event subscription and handling
- State synchronization
- Graceful degradation (if not in Hyprland)

---

### `internal/config`
**Purpose:** Configuration management

```go
// config.go
package config

type Config struct {
    Keybindings map[string]string
    Theme       ThemeConfig
    Preview     PreviewConfig
    Behavior    BehaviorConfig
}

func Load() (*Config, error) {
    // Load from ~/.config/warren/config.toml
}
```

**Configuration File (TOML):**
```toml
[theme]
show_hidden = false
icon_size = 24

[keybindings]
quit = "q"
navigate_up = "k"
navigate_down = "j"
parent_dir = "h"
enter_dir = "l"

[preview]
enabled = true
max_file_size = 1048576  # 1MB

[hyprland]
workspace_memory = true
suggested_window_rule = "float, ^(warren)$"
```

**Responsibilities:**
- Parse config files
- Provide defaults
- Validate settings
- Hot-reload support (future)

---

### `pkg/models`
**Purpose:** Shared data structures

```go
// file.go
package models

type File struct {
    Path        string
    Name        string
    Size        int64
    IsDirectory bool
    // ... more fields
}

// operation.go
type Operation struct {
    Type     OperationType  // Copy, Move, Delete
    Source   []string
    Dest     string
    Progress chan float64
}
```

**Responsibilities:**
- Define core data types
- Shared across packages
- No business logic (pure data)

## Data Flow

### Startup Flow
```
main.go
  → app.Run()
    → config.Load()
    → hyprland.New()
    → ui.NewMainWindow()
    → fileops.ListDirectory(homeDir)
    → ui.RenderFileList()
```

### Navigation Flow
```
User presses 'j' (down)
  → ui.HandleKeyPress()
    → fileView.SelectNext()
    → ui.UpdateSelection()
    → fileops.GetFileInfo(selected)
    → ui.UpdatePreview()
```

### File Operation Flow
```
User presses 'dd' (cut)
  → ui.HandleKeyPress()
    → app.MarkForCut(files)
    → statusBar.ShowMessage("Marked for cut")

User presses 'p' (paste)
  → ui.HandleKeyPress()
    → app.ExecutePaste()
      → fileops.Move(marked, currentDir)
        [goroutine] → Perform move
        [goroutine] → Update progress
      → ui.RefreshFileList()
      → statusBar.ShowMessage("Moved X files")
```

### Hyprland Event Flow
```
Hyprland workspace changes
  → hyprland.EventListener()
    → app.OnWorkspaceChange(newWorkspace)
      → config.GetWorkspaceDirectory(newWorkspace)
      → fileops.ListDirectory(workspaceDir)
      → ui.RenderFileList()
```

## Concurrency Model

### Goroutines Usage

**File Operations:** Non-blocking file operations
```go
go func() {
    result, err := fileops.Copy(src, dst)
    progressChan <- result
}()
```

**Filesystem Watching:** Background watcher
```go
watcher := fileops.NewWatcher(currentDir)
go watcher.Start()
for event := range watcher.Events {
    ui.HandleFSEvent(event)
}
```

**Hyprland Events:** Event listener
```go
go hyprland.ListenEvents(func(event Event) {
    app.HandleHyprlandEvent(event)
})
```

### Thread Safety
- GTK operations must run on main thread
- Use `glib.IdleAdd()` for UI updates from goroutines
- Channels for inter-goroutine communication
- Mutexes for shared state (if needed)

## Error Handling Strategy

### Principles
1. **Never panic in production** - Always return errors
2. **User-friendly messages** - Show helpful error dialogs
3. **Log details** - Write technical info to logs
4. **Graceful degradation** - Continue when possible

### Example
```go
// fileops/list.go
func ListDirectory(path string) ([]FileInfo, error) {
    entries, err := os.ReadDir(path)
    if err != nil {
        if os.IsPermission(err) {
            return nil, fmt.Errorf("permission denied: %w", err)
        }
        return nil, fmt.Errorf("cannot read directory: %w", err)
    }
    // ...
}

// In app
files, err := fileops.ListDirectory(path)
if err != nil {
    log.Printf("Failed to list %s: %v", path, err)
    ui.ShowErrorDialog("Cannot open directory", err.Error())
    return
}
```

## Testing Strategy

### Unit Tests
- Test individual functions
- Mock filesystem operations
- Test edge cases

```go
// fileops/list_test.go
func TestListDirectory(t *testing.T) {
    files, err := ListDirectory("../../testdata/sample")
    require.NoError(t, err)
    assert.Len(t, files, 2)
}
```

### Integration Tests
- Test package interactions
- Require test environment
- Test Hyprland IPC (when available)

### UI Testing
- Manual testing primarily
- GTK Inspector for UI debugging
- Accessibility testing

## Performance Targets

- **Startup:** < 100ms
- **Directory listing:** < 50ms (1000 files)
- **File operation:** No UI blocking
- **Memory:** < 50MB baseline
- **Search:** < 200ms (10,000 files)

## Future Architectural Considerations

### Plugin System
```go
// pkg/plugin/plugin.go
type Plugin interface {
    Name() string
    Execute(context Context) error
}
```

### Event Bus
```go
// internal/events/bus.go
type EventBus struct {
    subscribers map[EventType][]Handler
}
```

### Caching Layer
```go
// internal/cache/cache.go
type FileCache struct {
    entries map[string]CachedEntry
}
```

---

**Next:** See [PHASES.md](PHASES.md) for implementation roadmap.
