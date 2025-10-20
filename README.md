# Warren

[![Version](https://img.shields.io/badge/version-0.1.1-blue.svg)](https://github.com/lawrab/warren/releases)
[![Go](https://img.shields.io/badge/go-1.25+-00ADD8.svg)](https://go.dev/)
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

**Version 0.1.1** - Phase 1 MVP Complete! 🎉

✅ **Completed Features:**
- Directory browsing with file metadata (Name, Size, Modified)
- Vim-style keyboard navigation (j/k/h/l + arrow keys)
- Open files with default applications (xdg-open)
- Toggle hidden files (. key)
- Configurable keybindings (TOML configuration)
- Multiple sort modes (name, size, modified, extension)
- Sort order toggle (ascending/descending)
- Performance optimized for large directories
- CI/CD pipeline with automated testing

🚧 **Next Phase:**
- Phase 2: Hyprland Integration (IPC, workspace awareness)

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
go build -o warren cmd/warren/main.go

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
go build -o warren cmd/warren/main.go && ./warren
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

## Philosophy

> "A warren is never just a collection of holes. It's a community, a system, a home."

Warren isn't just another file manager. It's built for people who:
- Use Hyprland and want deep integration
- Prefer keyboard over mouse
- Value speed and efficiency
- Appreciate minimal, purposeful design

## Contributing

This is a personal learning project, but suggestions and ideas are always welcome!

## License

*(To be determined)*

---

*Built with 💜 in memory of late nights, learning curves, and the quiet support that makes everything possible.*
