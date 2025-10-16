# Features

This document lists all planned features organized by priority and implementation phase.

## Legend

- ✅ = Implemented
- 🚧 = In Progress
- 📋 = Planned
- 💡 = Future/Maybe

## Phase 1: MVP Features (Must-Have)

### Core Navigation
- 📋 **Directory Listing** - Display files and directories
  - File name, size, modified date
  - Icons based on file type
  - Sorting (name, size, date, type)
- 📋 **Keyboard Navigation** - Vim-style movement
  - `j/k` - Move down/up
  - `h/l` - Parent dir / Enter dir
  - `gg` - Go to top
  - `G` - Go to bottom
  - `Ctrl+u/d` - Page up/down
- 📋 **File Opening** - Open with default application
  - Enter key to open selected file
  - Use `xdg-open` or equivalent
- 📋 **Basic UI** - Clean, minimal interface
  - Main file list
  - Status bar (current path, file count)
  - Menu bar (minimal)

### Configuration
- 📋 **Config File Support** - TOML configuration
  - `~/.config/warren/config.toml`
  - Default settings
  - Custom keybindings
- 📋 **Show/Hide Hidden Files** - Toggle dot files
  - `gh` keybinding
  - Persist preference

### Quality of Life
- 📋 **Error Handling** - User-friendly error dialogs
- 📋 **Performance** - Fast startup and navigation
- 📋 **Window State** - Remember size/position

---

## Phase 2: Hyprland Integration (High Priority)

### IPC Communication
- 📋 **Hyprland Client** - Connect to Hyprland socket
  - Command socket for queries
  - Event socket for real-time updates
  - Graceful fallback if not in Hyprland
- 📋 **Workspace Awareness** - Know current workspace
  - Query active workspace
  - React to workspace changes

### Workspace Features
- 📋 **Per-Workspace Directory Memory** - Remember last directory per workspace
  - Auto-save on workspace switch
  - Auto-restore when returning
  - Configurable in settings
- 📋 **Workspace Quick Jump** - Jump to workspace-specific directory
  - Map workspaces to directories
  - Quick switch with keybinding
- 📋 **Open in Workspace** - Open files in specific workspaces
  - Context menu option
  - Keybinding with workspace number

### Window Management
- 📋 **Suggested Window Rules** - Optimal Hyprland configuration
  - Documentation for recommended rules
  - Auto-generate config snippet
- 📋 **Smart Positioning** - Position based on Hyprland state
  - Center on active monitor
  - Respect monitor boundaries

---

## Phase 3: Power User Features (Medium Priority)

### Visual Mode & Selection
- 📋 **Visual Selection** - Vim-style visual mode
  - `v` - Enter visual mode
  - `V` - Visual line mode
  - `Escape` - Exit visual mode
  - Arrow keys or `j/k` to extend selection
- 📋 **Multi-Select** - Select multiple files
  - `Space` - Toggle selection on current file
  - Select ranges
  - Pattern-based selection (`*.txt`)

### File Operations
- 📋 **Copy/Cut/Paste** - Clipboard-style operations
  - `yy` - Yank (copy) file(s)
  - `dd` - Cut file(s)
  - `p` - Paste
  - Show operation status
- 📋 **Delete** - Safe file deletion
  - Confirmation dialog
  - Support for trash (if available)
  - Permanent delete option
- 📋 **Rename** - Rename files
  - `r` or `cw` keybinding
  - Inline editing
  - Validate new name
- 📋 **Create** - Create new files/directories
  - `o` - New file
  - `O` - New directory
  - Name input dialog
- 📋 **Progress Tracking** - For long operations
  - Progress bar
  - Cancel button
  - Time remaining estimate

### Dual-Pane Mode
- 📋 **Split View** - Two directories side-by-side
  - `Ctrl+w v` - Vertical split
  - `Tab` - Switch active pane
  - `Ctrl+w q` - Close pane
- 📋 **Cross-Pane Operations** - Copy/move between panes
  - Move from left to right
  - Quick transfer keybindings

### Preview Pane
- 📋 **Text Preview** - Show text file contents
  - First 100-500 lines
  - Syntax highlighting (basic)
  - Configurable max size
- 📋 **Image Preview** - Show image thumbnails
  - PNG, JPG, GIF support
  - Scaled to fit
  - Show dimensions
- 📋 **Media Info** - Metadata for media files
  - Video: duration, resolution, codec
  - Audio: duration, bitrate
  - Use external tools (mediainfo, ffprobe)

### Search & Filter
- 📋 **File Search** - Search by name
  - `/` - Forward search
  - `?` - Backward search
  - `n/N` - Next/previous result
  - Real-time filtering
- 📋 **Fuzzy Matching** - Fuzzy file name search
  - Smart matching algorithm
  - Highlight matches
- 📋 **Filter by Type** - Show only certain file types
  - Documents, images, videos, etc.
  - Custom filters

### Bookmarks
- 📋 **Named Bookmarks** - Vim-style marks
  - `m<letter>` - Set bookmark
  - `'<letter>` - Jump to bookmark
  - Persist across sessions
- 📋 **Quick Access** - Predefined bookmarks
  - Home, Downloads, Documents, etc.
  - Configurable in settings

---

## Phase 4: Polish Features (Nice-to-Have)

### Theming
- 📋 **GTK Theme Integration** - Respect system theme
  - Dark mode support
  - Light mode support
  - Auto-switch with system
- 📋 **Custom Colors** - Configure colors
  - Hyprland color integration
  - Custom CSS support
  - Color scheme presets
- 📋 **Icon Themes** - Support icon packs
  - System icon theme
  - Fallback icons
  - Custom icon paths

### Advanced Configuration
- 📋 **Keybinding Editor** - Visual keybinding config
  - List all current bindings
  - Detect conflicts
  - Reset to defaults
- 📋 **Column Configuration** - Choose displayed columns
  - Show/hide columns
  - Reorder columns
  - Custom column widths
- 📋 **Behavior Settings** - Tweak behavior
  - Confirmation dialogs
  - Double-click vs single-click
  - Follow symlinks behavior

### Performance Optimizations
- 📋 **Lazy Loading** - Load directories on demand
- 📋 **Virtual Scrolling** - Handle huge directories (10k+ files)
- 📋 **Caching** - Cache file info and thumbnails
- 📋 **Background Loading** - Async directory reading

### Documentation
- 📋 **User Guide** - Comprehensive documentation
  - Keybindings reference
  - Configuration guide
  - Troubleshooting
- 📋 **Man Page** - Traditional man page
- 📋 **In-App Help** - Built-in help dialog
  - `?` or `F1` to open
  - Quick reference card

---

## Future/Maybe Features (Post-1.0)

### Advanced File Operations
- 💡 **Batch Rename** - Rename multiple files
  - Regex patterns
  - Sequential numbering
  - Preview changes
- 💡 **File Comparison** - Diff mode for files
  - Side-by-side comparison
  - Syntax-highlighted diff
- 💡 **Advanced Permissions** - Detailed permission editing
  - chmod/chown GUI
  - ACL support
  - Recursive operations

### Archive Support
- 💡 **Browse Archives** - Navigate zip/tar without extracting
  - List contents
  - Extract individual files
  - Create archives
- 💡 **Archive Preview** - See contents in preview pane

### Search Enhancements
- 💡 **Content Search** - Search file contents
  - Integration with ripgrep
  - Show matching lines
  - Jump to matches
- 💡 **Advanced Filters** - Complex filtering
  - Size ranges
  - Date ranges
  - File type combinations
  - Regex patterns
- 💡 **Saved Searches** - Save and reuse searches

### Network Features
- 💡 **Remote Filesystems** - Optimizations for network mounts
  - Detect slow filesystems
  - Adjust behavior
  - Cache aggressively
- 💡 **SFTP/FTP Support** - Built-in remote file access
  - Connect to remote servers
  - Browse like local files

### Integration Features
- 💡 **Git Integration** - Show git status
  - Modified/untracked indicators
  - Quick git operations
  - Commit from file manager
- 💡 **Terminal Integration** - Open terminal here
  - Context menu option
  - Keybinding
  - Send commands to terminal
- 💡 **External Tools** - Integration with CLI tools
  - fzf for fuzzy finding
  - ripgrep for content search
  - Custom scripts/actions

### Plugin System
- 💡 **Plugin API** - Extensibility
  - Lua scripting interface
  - Custom actions
  - Hook into events
- 💡 **Plugin Manager** - Discover and install plugins
  - Community plugin repository
  - Automatic updates

### Advanced UI
- 💡 **Multiple Tabs** - Tab-based navigation
  - Like browser tabs
  - Drag and drop between tabs
- 💡 **Custom Layouts** - Flexible pane layouts
  - More than 2 panes
  - Different arrangements
  - Save layouts
- 💡 **Breadcrumb Navigation** - Visual path navigation
  - Click path segments
  - Right-click for siblings
- 💡 **Tree View** - Hierarchical directory tree
  - Side panel
  - Collapsible folders

### File Tagging
- 💡 **Custom Tags** - Tag files with labels
  - Color-coded tags
  - Filter by tags
  - Tag-based organization
- 💡 **Smart Collections** - Virtual folders based on criteria
  - All images
  - Recent files
  - Large files

### Accessibility
- 💡 **Screen Reader Support** - Full accessibility
- 💡 **High Contrast Mode** - Better visibility
- 💡 **Keyboard-Only Mode** - Fully keyboard accessible (already planned)

### Performance Features
- 💡 **File Indexing** - Fast search across entire filesystem
  - Background indexing
  - Real-time updates
  - Database-backed
- 💡 **Thumbnail Cache** - Persistent thumbnail cache
  - XDG thumbnail spec
  - Share with other apps

---

## Feature Priority Matrix

### Must Have (MVP)
- Basic file navigation
- Keyboard controls
- File opening
- Configuration support

### Should Have (v1.0)
- Hyprland integration
- File operations (copy/move/delete)
- Visual selection
- Search
- Preview pane

### Nice to Have (v1.x)
- Dual-pane mode
- Theming
- Advanced bookmarks
- Performance optimizations

### Future (v2.0+)
- Plugin system
- Git integration
- Archive support
- Network features

---

## User Personas & Key Features

### Persona 1: Hyprland Power User
**Needs:** Fast keyboard-driven workflow, workspace integration
**Key Features:**
- Vim-style navigation
- Workspace directory memory
- Quick workspace switching
- No mouse required

### Persona 2: Developer
**Needs:** File operations, text preview, git integration
**Key Features:**
- Syntax-highlighted preview
- Batch operations
- Search functionality
- Terminal integration (future)

### Persona 3: Media Manager
**Needs:** Preview images/videos, organize large collections
**Key Features:**
- Image previews
- Thumbnail view
- Bulk rename (future)
- Tag system (future)

---

**Note:** Feature priorities may shift based on development experience and user feedback. The goal is a focused, polished tool, not feature bloat.
