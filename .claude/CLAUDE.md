# Warren Project Instructions

This file contains project-specific instructions for working on Warren, a Hyprland-optimized file manager.

## Project Overview

Warren is a keyboard-driven file manager built specifically for Hyprland users. It's written in Go using GTK4 and emphasizes speed, keyboard navigation, and deep Hyprland integration.

**Important Context:**
- This is a learning project to explore Linux development
- Dedicated to Ann Rabbets (AnnieRabbets) - be respectful of this dedication
- Focus on quality over speed - this is about learning and building something meaningful

## Documentation Structure

Before making changes, ALWAYS review these documents:

1. **docs/VISION.md** - Project philosophy and what makes Warren special
2. **docs/ARCHITECTURE.md** - Code structure and package organization
3. **docs/TECHNOLOGY.md** - Technology stack and setup instructions
4. **docs/PHASES.md** - Development roadmap (where we are in development)
5. **research/HYPRLAND_INTEGRATION.md** - Technical details on Hyprland IPC
6. **design/FEATURES.md** - Complete feature list with priorities

## Code Style and Standards

### Go Code Style
- Follow standard Go conventions (use `gofmt`)
- Package names: short, lowercase, no underscores
- Error handling: always handle errors explicitly, never ignore
- Comments: use godoc format for exported functions
- Keep functions small and focused
- Write tests for non-trivial logic

### Example Code Structure
```go
package fileops

// ListDirectory returns a list of files in the specified directory.
// Hidden files are included based on the showHidden parameter.
func ListDirectory(path string, showHidden bool) ([]FileInfo, error) {
    if path == "" {
        return nil, fmt.Errorf("path cannot be empty")
    }

    // Implementation...
}
```

### GTK4 Code Patterns
- Always run GTK operations on the main thread
- Use `glib.IdleAdd()` for UI updates from goroutines
- Properly handle widget lifecycle (connect/disconnect signals)
- Free resources when done (though Go GC helps)

### Concurrency Patterns
- Use goroutines for file operations
- Use channels for communication
- Protect shared state with mutexes if needed
- Always handle context cancellation

## Project Structure Rules

### Package Organization
- `cmd/warren/` - Only entry point, minimal logic
- `internal/` - All private application code
- `pkg/` - Public APIs (only if needed for external use)
- `testdata/` - Test fixtures

### File Naming
- `something.go` - Implementation
- `something_test.go` - Unit tests
- `doc.go` - Package documentation

## Development Workflow

### When Adding New Features
1. Check `docs/PHASES.md` to ensure feature is in current phase
2. Review `docs/ARCHITECTURE.md` to understand where code should go
3. Update relevant documentation if architecture changes
4. Write tests before or alongside implementation
5. Run `go fmt` and `golangci-lint` before committing

### When Fixing Bugs
1. Write a failing test that reproduces the bug
2. Fix the bug
3. Verify test passes
4. Consider if bug reveals architectural issues

### Code Review Checklist
- [ ] Code follows Go conventions
- [ ] Error handling is explicit and appropriate
- [ ] Tests cover new/changed code
- [ ] Documentation is updated
- [ ] No hardcoded paths (use config)
- [ ] GTK operations are thread-safe
- [ ] Memory leaks considered (especially with GTK)

## Special Considerations

### Hyprland Integration
- Always check if running in Hyprland before using IPC
- Gracefully degrade when Hyprland is not available
- Handle socket connection failures
- Implement retry logic for transient failures
- See `research/HYPRLAND_INTEGRATION.md` for technical details

### Performance Targets
- Startup: < 100ms (stretch goal)
- Directory listing: < 50ms for 1000 files
- No UI blocking for operations > 100ms
- Memory: < 50MB baseline

### Error Handling Philosophy
1. **User-Facing Errors**: Clear, actionable messages
2. **Log Details**: Write technical info to logs
3. **Never Panic**: Always return errors (except truly unrecoverable cases)
4. **Context**: Include relevant context in error messages

Example:
```go
if err != nil {
    log.Printf("Failed to list directory %s: %v", path, err)
    ui.ShowError("Cannot Open Directory",
        fmt.Sprintf("Failed to open %s. Check permissions.", filepath.Base(path)))
    return
}
```

## Testing Strategy

### What to Test
- File operations (list, copy, move, delete)
- Hyprland IPC (with and without Hyprland)
- Configuration parsing
- Edge cases (empty directories, permissions, symlinks)

### What Not to Test (Yet)
- GTK UI components (test manually for now)
- Complex integration scenarios (manual testing)

### Running Tests
```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/fileops/

# With race detector
go test -race ./...
```

## Common Patterns

### Safe File Operations
Always validate paths and handle errors:
```go
func SafeFileOperation(path string) error {
    // Validate path
    if path == "" {
        return fmt.Errorf("path cannot be empty")
    }

    // Check if exists
    info, err := os.Stat(path)
    if err != nil {
        if os.IsNotExist(err) {
            return fmt.Errorf("path does not exist: %s", path)
        }
        return fmt.Errorf("cannot access path: %w", err)
    }

    // Perform operation
    // ...
}
```

### Async Operations with Progress
```go
func CopyFiles(files []string, dest string, progress chan<- float64) error {
    total := len(files)
    for i, file := range files {
        if err := copyFile(file, dest); err != nil {
            return err
        }
        progress <- float64(i+1) / float64(total)
    }
    close(progress)
    return nil
}
```

## Debugging Tips

### GTK Inspector
Enable GTK Inspector for UI debugging:
```bash
GTK_DEBUG=interactive ./warren
# Then press Ctrl+Shift+I
```

### Go Profiling
```bash
# CPU profile
go build -o warren && ./warren -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profile
go build -o warren && ./warren -memprofile=mem.prof
go tool pprof mem.prof
```

### Hyprland IPC Debugging
Test IPC commands manually:
```bash
echo "j/activeworkspace" | socat - UNIX-CONNECT:/tmp/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket.sock
```

## Don't Do This

❌ **Don't** hardcode paths
❌ **Don't** ignore errors (even in examples)
❌ **Don't** block the UI thread
❌ **Don't** use `panic()` for expected errors
❌ **Don't** commit without running tests
❌ **Don't** add features outside current phase without discussion
❌ **Don't** remove or change the dedication to Ann

## Do This

✅ **Do** check documentation before making changes
✅ **Do** write clear error messages
✅ **Do** add tests for new functionality
✅ **Do** use goroutines for long operations
✅ **Do** follow Go conventions
✅ **Do** ask for clarification when uncertain
✅ **Do** update docs when architecture changes
✅ **Do** respect the phased development approach

## Development Environment

### Using Nix Flake
```bash
# Enter dev environment
nix develop

# Or direnv (if you use it)
echo "use flake" > .envrc
direnv allow
```

### Without Nix
See `docs/TECHNOLOGY.md` for manual setup instructions.

## Current Status

Check `docs/PHASES.md` for current development phase and priorities.

Current Status (as of v0.1.1):
- **Phase:** Phase 1 (MVP) - COMPLETE ✅
- **Next Phase:** Phase 2 (Hyprland Integration)
- **Focus:** Ready to begin Hyprland IPC integration and workspace awareness

## Questions or Issues?

When unsure:
1. Check the documentation first
2. Look for similar patterns in existing code
3. Ask for clarification
4. Propose solutions rather than just identifying problems

---

**Remember:** This is a learning project. Take time to understand concepts. Write code that future-you will understand. Build something Ann would be proud of.
