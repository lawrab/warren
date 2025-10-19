package fileops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lawrab/warren/pkg/models"
)

func TestSortFiles(t *testing.T) {
	// Create test files with different properties
	now := time.Now()
	files := []models.FileInfo{
		{Name: "zebra.txt", Size: 100, IsDir: false, ModTime: now.Add(-1 * time.Hour)},
		{Name: "apple.txt", Size: 500, IsDir: false, ModTime: now.Add(-2 * time.Hour)},
		{Name: "Documents", Size: 4096, IsDir: true, ModTime: now.Add(-3 * time.Hour)},
		{Name: "banana.doc", Size: 200, IsDir: false, ModTime: now},
		{Name: "Config", Size: 4096, IsDir: true, ModTime: now.Add(-30 * time.Minute)},
		{Name: "readme.md", Size: 50, IsDir: false, ModTime: now.Add(-5 * time.Hour)},
	}

	t.Run("sort by name ascending", func(t *testing.T) {
		testFiles := make([]models.FileInfo, len(files))
		copy(testFiles, files)

		SortFiles(testFiles, models.SortByName, models.SortAscending)

		// Directories should come first
		if !testFiles[0].IsDir || !testFiles[1].IsDir {
			t.Error("Directories should be listed before files")
		}

		// Check directory sort order
		if testFiles[0].Name != "Config" {
			t.Errorf("First dir should be Config, got %s", testFiles[0].Name)
		}
		if testFiles[1].Name != "Documents" {
			t.Errorf("Second dir should be Documents, got %s", testFiles[1].Name)
		}

		// Check file sort order (alphabetical)
		fileNames := []string{}
		for i := 2; i < len(testFiles); i++ {
			fileNames = append(fileNames, testFiles[i].Name)
		}
		expected := []string{"apple.txt", "banana.doc", "readme.md", "zebra.txt"}
		for i, name := range expected {
			if fileNames[i] != name {
				t.Errorf("File[%d] should be %s, got %s", i, name, fileNames[i])
			}
		}
	})

	t.Run("sort by name descending", func(t *testing.T) {
		testFiles := make([]models.FileInfo, len(files))
		copy(testFiles, files)

		SortFiles(testFiles, models.SortByName, models.SortDescending)

		// Directories still come first, but in reverse order
		if !testFiles[0].IsDir || !testFiles[1].IsDir {
			t.Error("Directories should be listed before files")
		}

		// Check directory sort order (reverse alphabetical)
		if testFiles[0].Name != "Documents" {
			t.Errorf("First dir should be Documents, got %s", testFiles[0].Name)
		}

		// Files should be in reverse alphabetical order
		if testFiles[2].Name != "zebra.txt" {
			t.Errorf("First file should be zebra.txt, got %s", testFiles[2].Name)
		}
	})

	t.Run("sort by size ascending", func(t *testing.T) {
		testFiles := make([]models.FileInfo, len(files))
		copy(testFiles, files)

		SortFiles(testFiles, models.SortBySize, models.SortAscending)

		// Directories first
		if !testFiles[0].IsDir || !testFiles[1].IsDir {
			t.Error("Directories should be listed before files")
		}

		// Files sorted by size (smallest first)
		if testFiles[2].Size != 50 { // readme.md
			t.Errorf("Smallest file should be first, got size %d", testFiles[2].Size)
		}
		if testFiles[len(testFiles)-1].Size != 500 { // apple.txt
			t.Errorf("Largest file should be last, got size %d", testFiles[len(testFiles)-1].Size)
		}
	})

	t.Run("sort by size descending", func(t *testing.T) {
		testFiles := make([]models.FileInfo, len(files))
		copy(testFiles, files)

		SortFiles(testFiles, models.SortBySize, models.SortDescending)

		// Directories first, then files by size (largest first)
		if !testFiles[0].IsDir {
			t.Error("Directories should be listed before files")
		}
		if testFiles[2].Size != 500 { // apple.txt
			t.Errorf("Largest file should be first, got size %d", testFiles[2].Size)
		}
	})

	t.Run("sort by modified time ascending", func(t *testing.T) {
		testFiles := make([]models.FileInfo, len(files))
		copy(testFiles, files)

		SortFiles(testFiles, models.SortByModTime, models.SortAscending)

		// Directories first
		if !testFiles[0].IsDir {
			t.Error("Directories should be listed before files")
		}

		// Files sorted by modification time (oldest first)
		// readme.md is oldest (-5 hours)
		if testFiles[2].Name != "readme.md" {
			t.Errorf("Oldest file should be first, got %s", testFiles[2].Name)
		}
	})

	t.Run("sort by extension", func(t *testing.T) {
		testFiles := make([]models.FileInfo, len(files))
		copy(testFiles, files)

		SortFiles(testFiles, models.SortByExtension, models.SortAscending)

		// Directories first
		if !testFiles[0].IsDir {
			t.Error("Directories should be listed before files")
		}

		// Files grouped by extension
		extensions := []string{}
		for i := 2; i < len(testFiles); i++ {
			ext := filepath.Ext(testFiles[i].Name)
			extensions = append(extensions, ext)
		}

		// Should have: .doc, .md, .txt, .txt
		if extensions[0] != ".doc" {
			t.Errorf("First extension should be .doc, got %s", extensions[0])
		}
		if extensions[1] != ".md" {
			t.Errorf("Second extension should be .md, got %s", extensions[1])
		}
	})
}

func TestListDirectory(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create test structure:
	// tmpDir/
	//   file1.txt
	//   file2.log
	//   .hidden
	//   subdir/

	// Create regular files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.log")
	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create hidden file
	hiddenFile := filepath.Join(tmpDir, ".hidden")
	if err := os.WriteFile(hiddenFile, []byte("hidden"), 0644); err != nil {
		t.Fatalf("Failed to create hidden file: %v", err)
	}

	// Create subdirectory
	subdir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	t.Run("list without hidden files", func(t *testing.T) {
		files, err := ListDirectory(tmpDir, false)
		if err != nil {
			t.Fatalf("ListDirectory failed: %v", err)
		}

		// Should have 3 entries (file1.txt, file2.log, subdir)
		if len(files) != 3 {
			t.Errorf("Expected 3 files, got %d", len(files))
		}

		// Check that hidden file is not included
		for _, f := range files {
			if strings.HasPrefix(f.Name, ".") {
				t.Errorf("Hidden file %s should not be included", f.Name)
			}
		}

		// Verify we have the expected files
		foundFiles := make(map[string]bool)
		for _, f := range files {
			foundFiles[f.Name] = true
		}

		if !foundFiles["file1.txt"] {
			t.Error("file1.txt not found")
		}
		if !foundFiles["file2.log"] {
			t.Error("file2.log not found")
		}
		if !foundFiles["subdir"] {
			t.Error("subdir not found")
		}
	})

	t.Run("list with hidden files", func(t *testing.T) {
		files, err := ListDirectory(tmpDir, true)
		if err != nil {
			t.Fatalf("ListDirectory failed: %v", err)
		}

		// Should have 4 entries including .hidden
		if len(files) != 4 {
			t.Errorf("Expected 4 files, got %d", len(files))
		}

		// Check that hidden file is included
		foundHidden := false
		for _, f := range files {
			if f.Name == ".hidden" {
				foundHidden = true
				if !f.IsHidden {
					t.Error("Hidden file should be marked as hidden")
				}
			}
		}

		if !foundHidden {
			t.Error("Hidden file .hidden not found")
		}
	})

	t.Run("verify directory flag", func(t *testing.T) {
		files, err := ListDirectory(tmpDir, false)
		if err != nil {
			t.Fatalf("ListDirectory failed: %v", err)
		}

		for _, f := range files {
			if f.Name == "subdir" {
				if !f.IsDir {
					t.Error("subdir should be marked as directory")
				}
			} else {
				if f.IsDir {
					t.Errorf("%s should not be marked as directory", f.Name)
				}
			}
		}
	})

	t.Run("verify file properties", func(t *testing.T) {
		files, err := ListDirectory(tmpDir, false)
		if err != nil {
			t.Fatalf("ListDirectory failed: %v", err)
		}

		for _, f := range files {
			if f.Name == "file1.txt" {
				if f.Size != 8 { // "content1" is 8 bytes
					t.Errorf("file1.txt should be 8 bytes, got %d", f.Size)
				}
				if f.Path != file1 {
					t.Errorf("file1.txt path should be %s, got %s", file1, f.Path)
				}
			}
		}
	})

	t.Run("error on empty path", func(t *testing.T) {
		_, err := ListDirectory("", false)
		if err == nil {
			t.Error("Expected error for empty path")
		}
	})

	t.Run("error on nonexistent directory", func(t *testing.T) {
		_, err := ListDirectory("/this/does/not/exist", false)
		if err == nil {
			t.Error("Expected error for nonexistent directory")
		}
	})
}

func TestGetFileInfo(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "testfile.txt")
	testContent := []byte("test content")

	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("get info for regular file", func(t *testing.T) {
		info, err := GetFileInfo(testFile)
		if err != nil {
			t.Fatalf("GetFileInfo failed: %v", err)
		}

		if info.Name != "testfile.txt" {
			t.Errorf("Expected name testfile.txt, got %s", info.Name)
		}

		if info.Size != int64(len(testContent)) {
			t.Errorf("Expected size %d, got %d", len(testContent), info.Size)
		}

		if info.IsDir {
			t.Error("File should not be marked as directory")
		}

		if info.IsHidden {
			t.Error("File should not be marked as hidden")
		}
	})

	t.Run("get info for directory", func(t *testing.T) {
		info, err := GetFileInfo(tmpDir)
		if err != nil {
			t.Fatalf("GetFileInfo failed: %v", err)
		}

		if !info.IsDir {
			t.Error("Directory should be marked as IsDir")
		}
	})

	t.Run("get info for hidden file", func(t *testing.T) {
		hiddenFile := filepath.Join(tmpDir, ".hidden")
		if err := os.WriteFile(hiddenFile, []byte("hidden"), 0644); err != nil {
			t.Fatalf("Failed to create hidden file: %v", err)
		}

		info, err := GetFileInfo(hiddenFile)
		if err != nil {
			t.Fatalf("GetFileInfo failed: %v", err)
		}

		if !info.IsHidden {
			t.Error("Hidden file should be marked as IsHidden")
		}

		if info.Name != ".hidden" {
			t.Errorf("Expected name .hidden, got %s", info.Name)
		}
	})

	t.Run("error on empty path", func(t *testing.T) {
		_, err := GetFileInfo("")
		if err == nil {
			t.Error("Expected error for empty path")
		}
	})

	t.Run("error on nonexistent file", func(t *testing.T) {
		_, err := GetFileInfo(filepath.Join(tmpDir, "does-not-exist.txt"))
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})

	t.Run("symlink handling", func(t *testing.T) {
		// Create a symlink
		linkPath := filepath.Join(tmpDir, "link.txt")
		if err := os.Symlink(testFile, linkPath); err != nil {
			t.Skip("Skipping symlink test (requires symlink support)")
		}

		info, err := GetFileInfo(linkPath)
		if err != nil {
			t.Fatalf("GetFileInfo failed: %v", err)
		}

		if !info.IsSymlink {
			t.Error("Symlink should be marked as IsSymlink")
		}

		if info.SymlinkTarget == "" {
			t.Error("Symlink target should be set")
		}
	})
}
