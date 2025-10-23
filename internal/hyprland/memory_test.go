package hyprland

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceMemory_SetAndGet(t *testing.T) {
	tempDir := t.TempDir()
	wm, err := NewWorkspaceMemory(tempDir)
	if err != nil {
		t.Fatalf("Failed to create workspace memory: %v", err)
	}

	// Test Set and Get
	wm.Set(1, "/home/user/workspace1")
	wm.Set(2, "/home/user/workspace2")

	if got := wm.Get(1); got != "/home/user/workspace1" {
		t.Errorf("Get(1) = %q, want %q", got, "/home/user/workspace1")
	}

	if got := wm.Get(2); got != "/home/user/workspace2" {
		t.Errorf("Get(2) = %q, want %q", got, "/home/user/workspace2")
	}

	// Test non-existent workspace
	if got := wm.Get(999); got != "" {
		t.Errorf("Get(999) = %q, want empty string", got)
	}
}

func TestWorkspaceMemory_Clear(t *testing.T) {
	tempDir := t.TempDir()
	wm, err := NewWorkspaceMemory(tempDir)
	if err != nil {
		t.Fatalf("Failed to create workspace memory: %v", err)
	}

	wm.Set(1, "/home/user/workspace1")
	wm.Set(2, "/home/user/workspace2")

	// Clear workspace 1
	wm.Clear(1)

	if got := wm.Get(1); got != "" {
		t.Errorf("After Clear(1), Get(1) = %q, want empty string", got)
	}

	// Workspace 2 should still exist
	if got := wm.Get(2); got != "/home/user/workspace2" {
		t.Errorf("Get(2) = %q, want %q", got, "/home/user/workspace2")
	}
}

func TestWorkspaceMemory_ClearAll(t *testing.T) {
	tempDir := t.TempDir()
	wm, err := NewWorkspaceMemory(tempDir)
	if err != nil {
		t.Fatalf("Failed to create workspace memory: %v", err)
	}

	wm.Set(1, "/home/user/workspace1")
	wm.Set(2, "/home/user/workspace2")
	wm.Set(3, "/home/user/workspace3")

	wm.ClearAll()

	// All should be cleared
	for i := 1; i <= 3; i++ {
		if got := wm.Get(i); got != "" {
			t.Errorf("After ClearAll(), Get(%d) = %q, want empty string", i, got)
		}
	}

	// GetAll should return empty map
	all := wm.GetAll()
	if len(all) != 0 {
		t.Errorf("After ClearAll(), GetAll() has %d items, want 0", len(all))
	}
}

func TestWorkspaceMemory_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	wm, err := NewWorkspaceMemory(tempDir)
	if err != nil {
		t.Fatalf("Failed to create workspace memory: %v", err)
	}

	// Set some data
	wm.Set(1, "/home/user/workspace1")
	wm.Set(2, "/home/user/workspace2")
	wm.Set(5, "/home/user/workspace5")

	// Save to disk
	if err := wm.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tempDir, "hyprland-memory.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created")
	}

	// Create new instance and load
	wm2, err := NewWorkspaceMemory(tempDir)
	if err != nil {
		t.Fatalf("Failed to create second workspace memory: %v", err)
	}

	// Verify data was loaded
	tests := []struct {
		workspace int
		want      string
	}{
		{1, "/home/user/workspace1"},
		{2, "/home/user/workspace2"},
		{5, "/home/user/workspace5"},
		{999, ""},
	}

	for _, tt := range tests {
		if got := wm2.Get(tt.workspace); got != tt.want {
			t.Errorf("After load, Get(%d) = %q, want %q", tt.workspace, got, tt.want)
		}
	}
}

func TestWorkspaceMemory_GetAll(t *testing.T) {
	tempDir := t.TempDir()
	wm, err := NewWorkspaceMemory(tempDir)
	if err != nil {
		t.Fatalf("Failed to create workspace memory: %v", err)
	}

	// Set some data
	wm.Set(1, "/home/user/workspace1")
	wm.Set(2, "/home/user/workspace2")
	wm.Set(5, "/home/user/workspace5")

	all := wm.GetAll()

	// Verify all entries
	if len(all) != 3 {
		t.Errorf("GetAll() returned %d items, want 3", len(all))
	}

	expected := map[int]string{
		1: "/home/user/workspace1",
		2: "/home/user/workspace2",
		5: "/home/user/workspace5",
	}

	for workspace, dir := range expected {
		if got := all[workspace]; got != dir {
			t.Errorf("GetAll()[%d] = %q, want %q", workspace, got, dir)
		}
	}

	// Verify it's a copy (modifying returned map shouldn't affect internal state)
	all[1] = "/modified/path"
	if got := wm.Get(1); got != "/home/user/workspace1" {
		t.Errorf("Modifying GetAll() result affected internal state: Get(1) = %q", got)
	}
}

func TestWorkspaceMemory_DefaultConfigPath(t *testing.T) {
	// Test that NewWorkspaceMemory with empty string uses default path
	wm, err := NewWorkspaceMemory("")
	if err != nil {
		t.Fatalf("Failed to create workspace memory with default path: %v", err)
	}

	home, _ := os.UserHomeDir()
	expectedPath := filepath.Join(home, ".config", "warren", "hyprland-memory.json")

	if wm.configPath != expectedPath {
		t.Errorf("configPath = %q, want %q", wm.configPath, expectedPath)
	}
}

func TestWorkspaceMemory_LoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	wm, err := NewWorkspaceMemory(tempDir)
	if err != nil {
		t.Fatalf("Failed to create workspace memory: %v", err)
	}

	// Should start empty when no file exists
	if got := wm.Get(1); got != "" {
		t.Errorf("Get(1) on fresh instance = %q, want empty string", got)
	}

	all := wm.GetAll()
	if len(all) != 0 {
		t.Errorf("GetAll() on fresh instance has %d items, want 0", len(all))
	}
}

func TestWorkspaceMemory_Concurrent(t *testing.T) {
	tempDir := t.TempDir()
	wm, err := NewWorkspaceMemory(tempDir)
	if err != nil {
		t.Fatalf("Failed to create workspace memory: %v", err)
	}

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		workspace := i
		go func() {
			wm.Set(workspace, filepath.Join("/path/to/workspace", string(rune('0'+workspace))))
			_ = wm.Get(workspace)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all data is present
	all := wm.GetAll()
	if len(all) != 10 {
		t.Errorf("After concurrent operations, GetAll() has %d items, want 10", len(all))
	}
}
