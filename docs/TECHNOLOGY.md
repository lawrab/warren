# Technology Stack

## Overview

Warren is built with Go and GTK4, combining the performance and simplicity of Go with the maturity and capabilities of GTK4.

## Why Go?

### The Right Tool for the Job

**Reasons We Chose Go:**

1. **Fast Compilation & Development Cycle**
   - Compile times measured in seconds, not minutes
   - Quick iteration when building features
   - Single binary output - easy to distribute

2. **Excellent Standard Library**
   - `os` package: file operations, permissions, symlinks
   - `path/filepath`: cross-platform path handling
   - `io/fs`: modern filesystem interfaces
   - `encoding/json`: config file parsing

3. **Built-in Concurrency**
   - Goroutines for async file operations
   - Channels for event communication
   - Perfect for watching filesystem changes
   - Non-blocking UI updates

4. **Memory Safety Without Garbage Collection Pain**
   - No manual memory management (unlike C)
   - GC is fast enough for desktop apps
   - No segfaults or use-after-free bugs

5. **Great Tooling**
   - `go fmt` - automatic formatting
   - `go test` - built-in testing
   - `go mod` - dependency management
   - Rich ecosystem of linters

6. **Gentle Learning Curve**
   - Simpler than Rust (no lifetimes/borrowing)
   - Safer than C (no manual memory)
   - Faster than Python
   - More productive than C++

## Core Dependencies

### GTK4 + gotk4

**GTK4**: Mature, actively developed GUI toolkit
- Native Wayland support
- Modern, hardware-accelerated rendering
- Excellent accessibility support
- Rich widget set

**gotk4**: Go bindings for GTK4
- Repository: https://github.com/diamondburned/gotk4
- Code generation from GIR files
- Type-safe API
- Active maintenance

### Why GTK4 (not Qt/FLTK/etc)?

- **Native Linux Feel**: GTK is the de facto Linux GUI toolkit
- **Wayland-First**: Built for modern display servers
- **Rich Documentation**: Excellent official docs + community
- **Ecosystem**: Tons of examples and existing apps
- **Hyprland Friendly**: Many Hyprland users run GTK apps

## Development Environment Setup

### Prerequisites

```bash
# System requirements
- Go 1.23+ (latest version recommended)
- GTK4 development libraries
- pkg-config
- gcc (for CGo)

# Arch Linux (Hyprland users likely on Arch)
sudo pacman -S go gtk4 pkg-config base-devel

# Ubuntu/Debian
sudo apt install golang gtk4 libgtk-4-dev pkg-config build-essential

# Fedora
sudo dnf install golang gtk4-devel pkg-config gcc
```

### Project Initialization

```bash
# Create Go module
cd warren
go mod init github.com/lawrab/warren

# Install gotk4
go get -u github.com/diamondburned/gotk4/pkg/gtk/v4
go get -u github.com/diamondburned/gotk4/pkg/glib/v2
go get -u github.com/diamondburned/gotk4/pkg/gdk/v4

# Verify installation
go mod tidy
```

### IDE Setup

**Recommended: VSCode with Go extension**
```json
// .vscode/settings.json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "editor.formatOnSave": true,
  "go.formatTool": "gofmt"
}
```

**Or: Neovim with gopls**
```lua
-- Built-in LSP with gopls
require('lspconfig').gopls.setup{}
```

## Additional Libraries

### Planned Dependencies

**Core Functionality:**
- `github.com/fsnotify/fsnotify` - Filesystem watching
- `github.com/spf13/viper` - Configuration management
- Standard library only for most features!

**Hyprland Integration:**
- Custom IPC implementation (no external deps needed)
- Unix socket communication via `net` package

**Future/Optional:**
- `github.com/h2non/filetype` - MIME type detection
- `github.com/disintegration/imaging` - Thumbnail generation
- `github.com/gabriel-vasile/mimetype` - Advanced MIME detection

## Build System

### Basic Build

```bash
# Development build
go build -o warren cmd/warren/main.go

# Run directly
go run cmd/warren/main.go

# With race detector (debugging)
go run -race cmd/warren/main.go
```

### Optimized Release Build

```bash
# Production build (smaller binary, no debug info)
go build -ldflags="-s -w" -o warren cmd/warren/main.go

# With version info
VERSION=$(git describe --tags --always)
go build -ldflags="-s -w -X main.Version=$VERSION" -o warren cmd/warren/main.go

# Further compression with upx (optional)
upx --best --lzma warren
```

### Development Workflow

```bash
# Run with live reload (install with: go install github.com/cosmtrek/air@latest)
air

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Lint
golangci-lint run

# Format all code
go fmt ./...
```

## Project Structure (Preview)

```
warren/
├── cmd/
│   └── warren/
│       └── main.go          # Entry point
├── internal/
│   ├── ui/                  # GTK4 UI components
│   ├── fileops/             # File operations
│   ├── hyprland/            # Hyprland IPC
│   └── config/              # Configuration
├── pkg/                     # Public APIs (if any)
├── testdata/                # Test fixtures
├── go.mod
└── go.sum
```

## Testing Strategy

### Unit Tests
```go
// internal/fileops/operations_test.go
func TestListDirectory(t *testing.T) {
    // Use testdata directory
    files, err := ListDirectory("../../testdata/sample")
    if err != nil {
        t.Fatal(err)
    }
    // Assertions...
}
```

### Integration Tests
```go
// internal/hyprland/ipc_test.go
func TestHyprlandConnection(t *testing.T) {
    if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
        t.Skip("Not running under Hyprland")
    }
    // Test IPC...
}
```

### Manual Testing
- Run in actual Hyprland environment
- Test various filesystem scenarios
- Performance profiling with `pprof`

## Performance Considerations

### Go Performance Tips
1. **Avoid premature optimization** - Profile first
2. **Use sync.Pool** for frequently allocated objects
3. **Minimize allocations** in hot paths
4. **Goroutines are cheap** - use them liberally
5. **Buffered channels** for better throughput

### GTK4 Performance
1. **Lazy load directory contents** - don't block UI
2. **Virtual scrolling** for large directories
3. **Async thumbnail generation** - goroutines!
4. **Cache stat() results** - filesystem calls are slow

## Debugging Tools

```bash
# CPU profiling
go build -o warren && ./warren -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profiling
go build -o warren && ./warren -memprofile=mem.prof
go tool pprof mem.prof

# Race detection
go run -race cmd/warren/main.go

# GTK Inspector (debug UI)
GTK_DEBUG=interactive ./warren
```

## Resources

### Go Learning
- Official Tour: https://go.dev/tour/
- Effective Go: https://go.dev/doc/effective_go
- Go by Example: https://gobyexample.com/

### GTK4 Documentation
- Official GTK4 Docs: https://docs.gtk.org/gtk4/
- gotk4 Examples: https://github.com/diamondburned/gotk4-examples
- GTK4 Tutorial: https://www.gtk.org/docs/getting-started/

### Hyprland Development
- Hyprland Wiki: https://wiki.hyprland.org/
- IPC Documentation: https://wiki.hyprland.org/IPC/

---

**Next Steps:** See [ARCHITECTURE.md](ARCHITECTURE.md) for code organization details.
