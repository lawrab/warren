package hyprland

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// WorkspaceMemory tracks the last directory accessed per workspace.
// This allows Warren to remember and restore the directory when switching workspaces.
type WorkspaceMemory struct {
	workspaceDirs map[int]string
	mu            sync.RWMutex
	configPath    string // Path to save/load memory
}

// memoryData is the structure saved to disk.
type memoryData struct {
	WorkspaceDirs map[int]string `json:"workspace_dirs"`
}

// NewWorkspaceMemory creates a new workspace memory tracker.
// If configDir is empty, uses ~/.config/warren/hyprland-memory.json
func NewWorkspaceMemory(configDir string) (*WorkspaceMemory, error) {
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		configDir = filepath.Join(home, ".config", "warren")
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "hyprland-memory.json")

	wm := &WorkspaceMemory{
		workspaceDirs: make(map[int]string),
		configPath:    configPath,
	}

	// Load existing memory if file exists
	if err := wm.Load(); err != nil && !os.IsNotExist(err) {
		// Log error but don't fail - we'll start fresh
		// In a real implementation, you'd use a proper logger here
	}

	return wm, nil
}

// Set saves the directory for a workspace.
func (wm *WorkspaceMemory) Set(workspaceID int, directory string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.workspaceDirs[workspaceID] = directory
}

// Get retrieves the last directory for a workspace.
// Returns empty string if no directory is remembered for this workspace.
func (wm *WorkspaceMemory) Get(workspaceID int) string {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.workspaceDirs[workspaceID]
}

// Clear removes the directory mapping for a workspace.
func (wm *WorkspaceMemory) Clear(workspaceID int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	delete(wm.workspaceDirs, workspaceID)
}

// ClearAll removes all directory mappings.
func (wm *WorkspaceMemory) ClearAll() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.workspaceDirs = make(map[int]string)
}

// Save persists the workspace memory to disk.
func (wm *WorkspaceMemory) Save() error {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	data := memoryData{
		WorkspaceDirs: wm.workspaceDirs,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(wm.configPath, jsonData, 0644)
}

// Load reads the workspace memory from disk.
func (wm *WorkspaceMemory) Load() error {
	data, err := os.ReadFile(wm.configPath)
	if err != nil {
		return err
	}

	var loaded memoryData
	if err := json.Unmarshal(data, &loaded); err != nil {
		return err
	}

	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.workspaceDirs = loaded.WorkspaceDirs
	if wm.workspaceDirs == nil {
		wm.workspaceDirs = make(map[int]string)
	}

	return nil
}

// GetAll returns a copy of all workspace directories.
// This is useful for debugging or displaying current state.
func (wm *WorkspaceMemory) GetAll() map[int]string {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	copy := make(map[int]string, len(wm.workspaceDirs))
	for k, v := range wm.workspaceDirs {
		copy[k] = v
	}
	return copy
}
