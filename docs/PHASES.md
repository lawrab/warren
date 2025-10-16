# Development Phases

Warren will be developed in four phases, each building on the previous. This phased approach ensures we deliver a working MVP quickly while maintaining quality.

## Phase 1: MVP - Basic File Manager (2-4 weeks)

**Goal:** A working file manager with core functionality.

### Features
- ✅ GTK4 application window
- ✅ Single-pane directory listing
- ✅ Navigate directories (up/down, enter/back)
- ✅ Open files with default applications
- ✅ Basic file information display (name, size, date)
- ✅ Keyboard shortcuts (vim-style: j/k/h/l)
- ✅ Configuration file support (.toml)
- ✅ Show/hide hidden files

### Technical Implementation

**Week 1-2: Basic UI**
```
Tasks:
1. Set up Go module and dependencies (gotk4)
2. Create main window with menu bar
3. Implement basic file listing widget (GtkColumnView)
4. Add keyboard event handling
5. Implement directory navigation
6. Basic styling and layout
```

**Week 3-4: File Operations & Config**
```
Tasks:
1. File opening with xdg-open
2. Configuration file loading (TOML)
3. Keybinding customization
4. Show/hide hidden files toggle
5. Sorting (by name, size, date)
6. Basic error handling and dialogs
7. Testing and bug fixes
```

### Success Criteria
- Can navigate filesystem with keyboard only
- Can open files in default applications
- Configuration persists between sessions
- No crashes on common operations
- Startup time < 200ms

### Deliverable
A usable file browser that replaces basic file managers for simple tasks.

---

## Phase 2: Hyprland Integration (2-3 weeks)

**Goal:** Deep Hyprland integration that makes Warren feel native.

### Features
- ✅ Hyprland IPC communication
- ✅ Workspace awareness
- ✅ Per-workspace directory memory
- ✅ Custom window rules (suggested)
- ✅ Event-driven updates
- ✅ Quick workspace file opening

### Technical Implementation

**Week 1: IPC Foundation**
```
Tasks:
1. Implement Hyprland IPC client
2. Socket connection and command handling
3. Query active workspace
4. Get window and monitor information
5. Error handling (non-Hyprland environment)
```

**Week 2-3: Integration Features**
```
Tasks:
1. Per-workspace directory memory
   - Save last directory per workspace
   - Auto-restore on workspace switch
2. Event listener for workspace changes
3. Open files in specific workspaces
4. Window rule suggestions (float, size, position)
5. Integration with Hyprland's window history
```

### Integration Example
```go
// When user changes workspace in Hyprland
hyprland.OnWorkspaceChange(func(ws int) {
    lastDir := config.GetWorkspaceDir(ws)
    if lastDir != "" {
        app.NavigateToDirectory(lastDir)
    }
})

// When user opens file
app.OpenFile("document.pdf", OpenOptions{
    Workspace:   hyprland.CurrentWorkspace(),
    Application: "zathura",
})
```

### Success Criteria
- Seamlessly switches directories with workspaces
- Responds to Hyprland events in real-time
- Gracefully handles non-Hyprland environments
- Window positioning feels natural in Hyprland

### Deliverable
A file manager that feels like a native part of Hyprland.

---

## Phase 3: Power Features (3-4 weeks)

**Goal:** Advanced functionality for power users.

### Features
- ✅ Dual-pane mode (side-by-side directories)
- ✅ File operations (copy, move, delete, rename)
- ✅ Visual selection mode (like vim visual mode)
- ✅ Bulk operations
- ✅ Search functionality
- ✅ Image/text preview pane
- ✅ File size and permission display
- ✅ Symlink handling
- ✅ Bookmarks

### Technical Implementation

**Week 1: File Operations**
```
Tasks:
1. Copy files/directories (recursive)
2. Move and rename operations
3. Delete with confirmation
4. Progress tracking for long operations
5. Async operations (goroutines)
6. Operation queue management
```

**Week 2: Visual Selection & Bulk Ops**
```
Tasks:
1. Visual mode (select multiple files)
2. Range selection (select from A to B)
3. Pattern matching selection (*.txt)
4. Bulk operations on selection
5. Clipboard integration (yank/paste)
```

**Week 3: Search & Preview**
```
Tasks:
1. File name search (fuzzy matching)
2. Preview pane widget
3. Image preview (thumbnails)
4. Text file preview (first 100 lines)
5. Syntax highlighting (basic)
6. Video thumbnails (if ffmpeg available)
```

**Week 4: Polish & Extras**
```
Tasks:
1. Bookmark system (named locations)
2. Quick bookmark jumping (like vim marks)
3. Symlink detection and display
4. Permission display and basic editing
5. Disk space indicator in status bar
```

### Keybinding Examples
```
Visual Mode:
  v          - Enter visual mode
  V          - Visual line mode (select full file entries)
  Esc        - Exit visual mode

Operations:
  yy         - Yank (copy) current file
  dd         - Delete current file
  p          - Paste yanked/deleted files
  r          - Rename file

Search & Navigation:
  /          - Search forward
  ?          - Search backward
  n          - Next search result
  N          - Previous search result

Bookmarks:
  m<letter>  - Set bookmark
  '<letter>  - Jump to bookmark

Dual-pane:
  Tab        - Switch between panes
  Ctrl+w v   - Vertical split
  Ctrl+w q   - Close pane
```

### Success Criteria
- File operations complete successfully and safely
- Large operations don't block UI
- Preview pane useful for common file types
- Search is fast (< 200ms for 10k files)
- Dual-pane mode increases productivity

### Deliverable
A fully-featured file manager competitive with ranger/lf but with GUI benefits.

---

## Phase 4: Polish & Optimization (2-3 weeks)

**Goal:** Production-ready quality and performance.

### Features
- ✅ Theming (respect GTK/Hyprland colors)
- ✅ Custom keybinding editor
- ✅ Performance optimization
- ✅ Comprehensive error handling
- ✅ User documentation
- ✅ Package for distribution

### Technical Implementation

**Week 1: Theming & Config**
```
Tasks:
1. GTK CSS theming support
2. Color scheme configuration
3. Icon theme support
4. Respect system dark mode
5. Hyprland color integration (if possible)
6. Custom theme examples
```

**Week 2: Performance & Reliability**
```
Tasks:
1. Profile with pprof
2. Optimize hot paths
3. Memory usage optimization
4. Caching for repeated operations
5. Lazy loading for large directories
6. Virtual scrolling for huge file lists
7. Comprehensive error handling
8. Crash recovery
```

**Week 3: Documentation & Distribution**
```
Tasks:
1. User guide (keybindings, configuration)
2. Architecture documentation (for contributors)
3. Build and package scripts
4. AUR package (PKGBUILD)
5. Nix flake updates
6. CI/CD setup (GitHub Actions)
7. Release process
```

### Performance Targets
- Startup: < 100ms (was < 200ms)
- Directory listing: < 50ms for 1000 files
- Memory baseline: < 50MB
- Search: < 200ms for 10k files
- No UI stuttering during operations

### Configuration Example
```toml
# ~/.config/warren/config.toml

[appearance]
theme = "dark"
icon_size = 24
show_hidden = false
font = "Monospace 11"

[keybindings]
quit = "q"
navigate_up = "k"
navigate_down = "j"
parent_dir = "h"
enter_dir = "l"
search = "/"
bookmark_set = "m"
bookmark_go = "'"

[hyprland]
workspace_memory = true
window_rule = "float, ^(warren)$"
window_size = "1200 800"

[preview]
enabled = true
max_text_size = 1048576  # 1MB
show_thumbnails = true
```

### Success Criteria
- Feels polished and professional
- No noticeable performance issues
- Easy to install and configure
- Comprehensive documentation
- Package available for Arch Linux
- Users can customize without code changes

### Deliverable
Production-ready v1.0 release.

---

## Post-1.0 Ideas (Future Considerations)

These features are out of scope for v1.0 but could be added later:

### Plugin System
- Lua or Go plugin interface
- Custom actions and shortcuts
- Community plugin repository

### Advanced Features
- Archive browsing (zip/tar without extracting)
- Network filesystem optimization (SMB/NFS)
- Trash/recycle bin support
- File comparison (diff mode)
- Advanced search (content, regex, metadata)
- Git integration (show git status in file list)
- Batch renaming with regex
- Extended attribute support

### Integration
- Integration with terminal (open terminal here)
- FZF integration for fuzzy finding
- Ripgrep integration for content search
- Notification support (operation complete)

### UI Enhancements
- Multiple tabs
- Split view layouts (more than 2 panes)
- Breadcrumb navigation
- Custom columns (show/hide metadata)
- File tagging system

---

## Timeline Summary

```
Phase 1 (MVP):              Weeks 1-4
Phase 2 (Hyprland):         Weeks 5-7
Phase 3 (Power Features):   Weeks 8-11
Phase 4 (Polish):           Weeks 12-14

Total estimated time: 14 weeks (~3.5 months)
```

**Actual timeline will vary based on:**
- Available development time
- Complexity of Hyprland integration
- Testing and bug fixing needs
- Learning curve with gotk4

## Development Approach

### Iterative Development
- Each phase builds on previous
- Maintain working state throughout
- Regular testing on real Hyprland system
- Document as you go

### Quality Gates
Each phase must meet criteria before moving to next:
- All features implemented
- No critical bugs
- Performance targets met
- Basic testing complete
- Documentation updated

### Flexibility
This is a learning project - adapt as needed:
- Skip features if too complex
- Add features if easier than expected
- Reorder based on learning
- Take time to understand new concepts

---

**Remember:** The goal is learning Linux development and building something useful for Ann. Progress > Perfection.

**Next:** See [HYPRLAND_INTEGRATION.md](../research/HYPRLAND_INTEGRATION.md) for technical details on Hyprland integration.
