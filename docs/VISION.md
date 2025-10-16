# Vision: Warren for Hyprland

## The Problem

Traditional file managers (Nautilus, Dolphin, Thunar) are designed for desktop environments with:
- Mouse-centric workflows
- Floating window paradigms
- Generic desktop integration
- Feature bloat for casual users

**Hyprland users are different.** They chose a tiling window manager for:
- Keyboard efficiency
- Minimal resource usage
- Precise control over their environment
- Speed and responsiveness

Yet they're stuck using file managers that fight against these principles.

## The Solution: Warren

Warren is a file manager that **embraces the Hyprland philosophy** from the ground up.

### Core Principles

#### 1. Keyboard-First, Always
- Every action accessible via keyboard
- Vim-inspired navigation (j/k, gg, G, etc.)
- Custom keybinding support
- Mouse support is optional, never required

#### 2. Hyprland-Native Integration
- **IPC Communication**: Direct connection to Hyprland socket
- **Workspace Awareness**: Know which workspace you're on, act accordingly
- **Window Rules**: Suggest optimal window rules for Warren
- **Event-Driven**: React to Hyprland events (workspace changes, window focus)

#### 3. Performance First
- Instant startup (sub-100ms target)
- Lazy loading of directory contents
- Efficient memory usage
- No background daemons or services

#### 4. Minimal and Purposeful
- No cruft, no legacy features
- Clean, distraction-free interface
- Configurable but opinionated defaults
- Each feature must justify its existence

## What Makes Warren Different

### vs. Traditional GUI File Managers (Nautilus, Dolphin)
- ❌ Mouse-required workflows → ✅ Pure keyboard control
- ❌ Designed for GNOME/KDE → ✅ Built for Hyprland
- ❌ Heavy resource usage → ✅ Lightweight and fast
- ❌ Feature bloat → ✅ Minimal and focused

### vs. TUI File Managers (ranger, lf, nnn)
- ❌ Terminal-only → ✅ Native GUI with previews
- ❌ Limited preview support → ✅ Rich previews (images, videos, PDFs)
- ❌ No Hyprland integration → ✅ Deep Hyprland IPC integration
- ❌ ASCII limitations → ✅ Modern GTK4 interface

### vs. Other GUI Keyboard File Managers
- ❌ Generic tiling WM support → ✅ Hyprland-specific optimizations
- ❌ Desktop-environment assumptions → ✅ Standalone, WM-aware
- ❌ Inconsistent keybindings → ✅ Vim-style consistency

## Target User

Warren is built for:

**The Hyprland Power User**
- Spends most time in terminal/editors
- Values keyboard efficiency
- Wants deep system integration
- Appreciates minimal, purposeful tools
- Comfortable with configuration files

**Not For:**
- Users who prefer mouse-driven workflows
- Desktop environment users (GNOME/KDE/XFCE)
- Those wanting WYSIWYG file manager customization
- Users needing Windows-Explorer-like experience

## Key Differentiators

### 1. Hyprland IPC Integration
```go
// Example: Open files in specific workspaces
warren.OpenWith("document.pdf", OpenOptions{
    Workspace: hyprland.CurrentWorkspace(),
    PreferredApp: "zathura",
})
```

### 2. Workspace-Aware Behavior
- Remember last directory per workspace
- Open files in appropriate workspaces
- Quick jump between workspace-specific directories

### 3. Smart Window Management
```hyprlang
# Suggested window rules
windowrule = float, ^(warren)$
windowrule = size 1200 800, ^(warren)$
windowrule = center, ^(warren)$
```

### 4. Keyboard-Driven Operations
```
Navigation:  j/k, h/l, gg, G, Ctrl+u/d
Selection:   Space, v (visual), V (visual line)
Actions:     yy (copy), dd (cut), p (paste), / (search)
Bookmarks:   m<letter> (set), '<letter> (goto)
Workspaces:  Ctrl+1-9 (open in workspace)
```

## Success Criteria

Warren will be considered successful when:

1. **Speed**: Startup < 100ms, directory listing < 50ms
2. **Adoption**: Hyprland users recommend it over alternatives
3. **Integration**: Seamlessly feels like part of Hyprland
4. **Reliability**: Handles edge cases gracefully (network mounts, permissions, etc.)
5. **Maintainability**: Code is clean, documented, extensible

## Anti-Goals

Warren explicitly **will not**:
- Support other window managers/desktop environments (Hyprland only)
- Implement features that require mouse interaction
- Include a GUI settings panel (config file only)
- Become a "Swiss Army knife" tool
- Support Windows or macOS

## Future Vision

Once MVP is solid:
- Plugin system for custom actions
- Network filesystem optimization
- Advanced search (content, metadata, fuzzy)
- Integration with other CLI tools (fzf, ripgrep)
- Custom preview scripts
- Archive handling (browse zip/tar without extracting)

But always: **Fast, focused, Hyprland-native.**

---

*"The best tool is the one that disappears into your workflow."*
