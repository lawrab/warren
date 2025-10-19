# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Warren is a keyboard-driven GTK4 file manager built specifically for Hyprland, written in Go. It's currently in early development (Phase 1 MVP). This is a learning project dedicated to Ann Rabbets.

**Tech Stack:** Go 1.25+, GTK4 (via gotk4 bindings), Hyprland IPC

## Development Environment

### Using Nix (Recommended)
```bash
# Enter development environment with all dependencies
nix develop

# Or use direnv
echo "use flake" > .envrc && direnv allow
```

The Nix flake provides: Go toolchain, GTK4, pkg-config, gopls, golangci-lint, and debugging tools.

### Environment Variables (in Nix shell)
- `CGO_ENABLED=1` - Required for GTK4 bindings
- `GTK_DEBUG=interactive` - Enables GTK Inspector (Ctrl+Shift+I)
- `WARREN_DEV=1` - Development mode flag

## Build and Run

```bash
# Build
go build -o warren cmd/warren/main.go

# Run directly
go run cmd/warren/main.go

# Run with GTK Inspector enabled
GTK_DEBUG=interactive go run cmd/warren/main.go

# Build for production (from Nix)
nix build
```

## Testing

```bash
# Run all tests
go test ./...

# Test specific package
go test ./internal/fileops/

# With coverage
go test -cover ./...

# With race detector
go test -race ./...

# Single test
go test ./internal/fileops/ -run TestListDirectory
```

**Note:** Currently no tests exist - this is early development phase.

## Code Standards

```bash
# Format code (always run before commits)
go fmt ./...

# Lint code
golangci-lint run

# Check for common issues
go vet ./...
```

## Architecture Overview

### Package Structure
```
cmd/warren/main.go           # Entry point - minimal logic
internal/
  ├── app/                   # Application lifecycle and state coordination
  ├── ui/                    # GTK4 widgets and UI logic
  ├── fileops/              # Filesystem operations (list, copy, move, delete)
  ├── hyprland/             # Hyprland IPC client and event handling
  └── config/               # Configuration management (TOML)
pkg/models/                  # Shared data structures (no business logic)
```

### Key Architectural Principles

**Separation of Concerns:**
- `cmd/warren`: Only initializes GTK app and delegates to `internal/app`
- `internal/app`: Coordinates between packages, manages global state
- `internal/ui`: All GTK4 code lives here
- `internal/fileops`: Pure filesystem operations, no UI dependencies
- `internal/hyprland`: Hyprland-specific code with graceful degradation

**Concurrency:**
- File operations run in goroutines (never block UI)
- GTK updates must use `glib.IdleAdd()` from goroutines
- Channels for inter-goroutine communication
- Mutexes for shared state protection

**Error Handling:**
- Never `panic()` for expected errors (return errors instead)
- User-facing: Clear, actionable error dialogs
- Logs: Include technical details for debugging
- Gracefully degrade when features unavailable (e.g., non-Hyprland environment)

## GTK4-Specific Patterns

### Thread Safety
```go
// WRONG: Updating UI from goroutine
go func() {
    files := loadFiles()
    label.SetText("Done")  // ❌ Will crash
}()

// CORRECT: Use glib.IdleAdd
go func() {
    files := loadFiles()
    glib.IdleAdd(func() {
        label.SetText("Done")  // ✅ Safe
    })
}()
```

### Widget Lifecycle
- Connect signals when creating widgets
- Disconnect when cleaning up (if needed)
- Let Go GC handle most cleanup

## Hyprland Integration

**IPC Socket Location:** `/tmp/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket.sock`

**Always check availability:**
```go
if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
    // Not in Hyprland - gracefully degrade
    return nil
}
```

**Manual IPC testing:**
```bash
echo "j/activeworkspace" | socat - UNIX-CONNECT:/tmp/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket.sock
```

## Current Development Phase

**Phase 1: MVP (Weeks 1-4)** - Current focus
- Basic GTK4 window ✅
- Directory listing (in progress)
- Keyboard navigation
- File opening with xdg-open
- Configuration file support

See `docs/PHASES.md` for complete roadmap.

## Important Project Context

- **Learning project** - Focus on understanding concepts, not just shipping features
- **Dedicated to Ann Rabbets** - Be respectful of this dedication
- **Quality over speed** - Take time to write maintainable code
- **Phased development** - Don't implement features from later phases without discussion

## Key Documentation

Before making significant changes, review:
- `docs/ARCHITECTURE.md` - Detailed architecture and data flow
- `docs/PHASES.md` - Development roadmap and current priorities
- `docs/TECHNOLOGY.md` - Setup instructions and technology decisions
- `.claude/CLAUDE.md` - Comprehensive project instructions and coding patterns

## Common Tasks

### Adding a New Internal Package
1. Create under `internal/` with lowercase name (e.g., `internal/cache`)
2. Add `doc.go` with package documentation
3. Update `docs/ARCHITECTURE.md` with package responsibilities
4. Keep packages focused - single responsibility principle

### Debugging GTK UI Issues
```bash
# Launch with GTK Inspector
GTK_DEBUG=interactive ./warren
# Press Ctrl+Shift+I to open inspector
# Navigate widget tree, check CSS classes, view signals
```

### Profiling Performance
```bash
# CPU profile
go build -o warren && ./warren -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profile
go build -o warren && ./warren -memprofile=mem.prof
go tool pprof mem.prof
```

## Performance Targets

- Startup: < 100ms (stretch goal, current target < 200ms)
- Directory listing: < 50ms for 1000 files
- Memory baseline: < 50MB
- No UI blocking for operations > 100ms

## Dependencies

**Core:**
- `github.com/diamondburned/gotk4/pkg` - GTK4 bindings

**Future (not yet added):**
- TOML parser (for configuration)
- Filesystem watcher (for live updates)
- Image libraries (for preview pane)

## Development Workflow

### Branch Strategy
**IMPORTANT:** The `main` branch is protected. All changes must go through pull requests.

```bash
# Start from up-to-date main
git checkout main
git pull

# Create feature branch
git checkout -b feature/description

# Make changes and commit
git add .
git commit -m "feat: description"

# Push and create PR
git push -u origin feature/description
gh pr create
```

### Branch Naming
- `feature/description` - New features
- `bugfix/description` - Bug fixes
- `docs/description` - Documentation
- `refactor/description` - Code refactoring
- `test/description` - Tests

### Git Hooks (Local Checks)
```bash
# One-time setup - runs checks before each commit
bash .githooks/setup.sh

# Hooks run: format check, lint, tests
# Skip if needed: git commit --no-verify
```

### CI/CD Pipeline
Every push and PR triggers automated checks:
- ✅ **Tests** - Go 1.23 & 1.25 with race detector
- ✅ **Lint** - golangci-lint with project config
- ✅ **Format** - gofmt check
- ✅ **Build** - Binary compilation and execution test

All checks must pass before PR can be merged.

### Pull Request Process
1. Create PR from feature branch to `main`
2. Fill out PR template
3. Wait for CI checks to pass
4. Request review if needed
5. Merge when approved and CI passes

See `CONTRIBUTING.md` for detailed guidelines.

## Commit Guidelines

- Keep commits focused and atomic
- Reference issue/feature in commit message if applicable
- Run `go fmt` before committing
- Ensure tests pass
- Follow conventional commits format when possible
