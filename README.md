# Warren

[![Version](https://img.shields.io/badge/version-0.2.0-blue.svg)](https://github.com/lawrab/warren/releases)
[![Go](https://img.shields.io/badge/go-1.25+-00ADD8.svg)](https://go.dev/)
[![codecov](https://codecov.io/gh/lawrab/warren/branch/main/graph/badge.svg)](https://codecov.io/gh/lawrab/warren)
[![License](https://img.shields.io/badge/license-TBD-lightgrey.svg)](LICENSE)

> *For Ann - who brings light to every burrow*

A blazingly fast, keyboard-driven file manager built specifically for Hyprland.

## About

Warren is a modern file manager designed from the ground up for tiling window manager users who live in Hyprland. Like the intricate network of tunnels in a rabbit warren, this file manager helps you navigate your filesystem with speed and efficiency.

This project is dedicated to Ann Rabbets (AnnieRabbets), whose strength and spirit inspire every line of code.

## Why Warren?

Existing file managers are built for traditional desktop environments. Warren embraces the Hyprland philosophy:
- **Keyboard-first navigation** - Mouse optional, never required
- **Hyprland-native integration** - IPC communication, workspace awareness, custom window rules
- **Performance-focused** - Written in Go for speed and efficiency
- **Minimal and purposeful** - No bloat, just what you need

## Project Status

**Version 0.2.0** - Phase 2 Hyprland Integration Complete! 🎉

✅ **Completed Features:**

**Phase 1 - Core File Manager:**
- Directory browsing with file metadata (Name, Size, Modified)
- Vim-style keyboard navigation (j/k/h/l + arrow keys)
- Open files with default applications (xdg-open)
- Toggle hidden files (. key)
- Configurable keybindings (TOML configuration)
- Multiple sort modes (name, size, modified, extension)
- Sort order toggle (ascending/descending)
- Performance optimized for large directories
- CI/CD pipeline with automated testing

**Phase 2 - Hyprland Integration:**
- Hyprland IPC client with command and event support
- Per-workspace directory memory
- Automatic directory switching on workspace change
- Graceful degradation in non-Hyprland environments
- Persistent workspace memory (saved across sessions)
- Code refactoring: Split main.go into logical modules for better maintainability

🚧 **Next Phase:**
- Phase 3: Power Features (file operations, dual-pane mode, search)

See [docs/PHASES.md](docs/PHASES.md) for the complete development roadmap.

## Documentation

- [VISION.md](docs/VISION.md) - Project philosophy and goals
- [TECHNOLOGY.md](docs/TECHNOLOGY.md) - Technology stack and setup
- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - Code structure and design
- [PHASES.md](docs/PHASES.md) - Development roadmap from MVP to v1.0
- [HYPRLAND_INTEGRATION.md](research/HYPRLAND_INTEGRATION.md) - Hyprland-specific features
- [FEATURES.md](design/FEATURES.md) - Feature list and priorities

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/lawrab/warren.git
cd warren

# Build
go build -o warren ./cmd/warren

# Run
./warren

# Check version
./warren --version
```

### Using Nix (Recommended)

```bash
# Enter development environment
nix develop

# Build and run
go build -o warren ./cmd/warren && ./warren
```

### Keyboard Shortcuts

- **j/k** or **↑/↓** - Navigate up/down
- **h** or **←/Backspace** - Go to parent directory
- **l** or **→/Enter** - Enter directory or open file
- **s** - Cycle sort mode (name → size → modified → extension)
- **r** - Reverse sort order (ascending ↔ descending)
- **.** (period) - Toggle hidden files
- **q** or **Ctrl+Q** - Quit

All keybindings are customizable via `~/.config/warren/config.toml`

### Hyprland Integration

Warren automatically detects and integrates with Hyprland when running in a Hyprland session. Configuration options:

```toml
[hyprland]
enabled = true           # Enable Hyprland integration (auto-detected)
workspace_memory = true  # Remember directory per workspace
auto_switch = true       # Auto-switch to remembered directory on workspace change
```

**Features:**
- Remembers the last directory accessed in each workspace
- Automatically switches to the remembered directory when you switch workspaces
- Persists workspace memory across sessions (`~/.config/warren/hyprland-memory.json`)
- Gracefully degrades when not running in Hyprland

## Philosophy

> "A warren is never just a collection of holes. It's a community, a system, a home."

Warren isn't just another file manager. It's built for people who:
- Use Hyprland and want deep integration
- Prefer keyboard over mouse
- Value speed and efficiency
- Appreciate minimal, purposeful design

## Testing

Warren maintains **~70% test coverage** for testable business logic. The coverage badge shows lower numbers (~34%) because certain components are excluded from coverage metrics:

- **GTK4 UI code** (`internal/ui/`) - Requires display server and complex mocking, verified through manual testing
- **Main entry point** (`cmd/warren/main.go`) - GTK initialization and activate functions, tested implicitly
- **File watcher** (`internal/fileops/watcher.go`) - Goroutine-heavy code tested via integration
- **File operations** (`internal/fileops/open.go`) - System-dependent (xdg-open), verified manually

**Well-tested components:**
- Data models: **100%** coverage (`pkg/models/`)
- Version info: **83.3%** coverage (`internal/version/`)
- Hyprland IPC: **80.9%** coverage (`internal/hyprland/`)
- Configuration: **73.8%** coverage (`internal/config/`)
- File operations: **52.3%** coverage (testable parts of `internal/fileops/`)
- Helper functions: Full test suite (`cmd/warren/helpers_test.go`, `cmd/warren/keyboard_test.go`)

This is standard practice for GTK applications where UI framework code is impractical to unit test. See `.codecov.yml` for exclusion details.

## Contributing

This is a personal learning project, but suggestions and ideas are always welcome!

## License

*(To be determined)*

---

*Built with 💜 in memory of late nights, learning curves, and the quiet support that makes everything possible.*
