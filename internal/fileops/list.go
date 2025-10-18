package fileops

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lawrab/warren/pkg/models"
)

// ListDirectory reads the contents of a directory and returns a list of FileInfo.
// Hidden files are included based on the showHidden parameter.
func ListDirectory(path string, showHidden bool) ([]models.FileInfo, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	// Clean and make absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Read directory entries
	entries, err := os.ReadDir(absPath)
	if err != nil {
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied: %w", err)
		}
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory does not exist: %w", err)
		}
		return nil, fmt.Errorf("cannot read directory: %w", err)
	}

	// Convert to FileInfo
	files := make([]models.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			// Skip files we can't stat (rare, but possible)
			continue
		}

		name := entry.Name()
		isHidden := IsHidden(name)

		// Skip hidden files if requested
		if isHidden && !showHidden {
			continue
		}

		fullPath := filepath.Join(absPath, name)
		fileInfo := models.FileInfo{
			Name:        name,
			Path:        fullPath,
			Size:        info.Size(),
			IsDir:       info.IsDir(),
			Permissions: info.Mode(),
			ModTime:     info.ModTime(),
			IsHidden:    isHidden,
		}

		// Check for symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			fileInfo.IsSymlink = true
			target, err := os.Readlink(fullPath)
			if err == nil {
				fileInfo.SymlinkTarget = target
			}
		}

		files = append(files, fileInfo)
	}

	return files, nil
}

// GetFileInfo returns detailed information about a single file or directory.
func GetFileInfo(path string) (models.FileInfo, error) {
	if path == "" {
		return models.FileInfo{}, fmt.Errorf("path cannot be empty")
	}

	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return models.FileInfo{}, fmt.Errorf("file does not exist: %w", err)
		}
		if os.IsPermission(err) {
			return models.FileInfo{}, fmt.Errorf("permission denied: %w", err)
		}
		return models.FileInfo{}, fmt.Errorf("cannot access file: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	fileInfo := models.FileInfo{
		Name:        filepath.Base(path),
		Path:        absPath,
		Size:        info.Size(),
		IsDir:       info.IsDir(),
		Permissions: info.Mode(),
		ModTime:     info.ModTime(),
		IsHidden:    IsHidden(filepath.Base(path)),
	}

	// Check for symlinks
	if info.Mode()&os.ModeSymlink != 0 {
		fileInfo.IsSymlink = true
		target, err := os.Readlink(path)
		if err == nil {
			fileInfo.SymlinkTarget = target
		}
	}

	return fileInfo, nil
}

// SortFiles sorts a list of files according to the specified criteria.
// Directories are always listed before files.
func SortFiles(files []models.FileInfo, sortBy models.SortBy, order models.SortOrder) {
	sort.Slice(files, func(i, j int) bool {
		// Always sort directories before files
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}

		var less bool
		switch sortBy {
		case models.SortByName:
			less = strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
		case models.SortBySize:
			less = files[i].Size < files[j].Size
		case models.SortByModTime:
			less = files[i].ModTime.Before(files[j].ModTime)
		case models.SortByExtension:
			extI := filepath.Ext(files[i].Name)
			extJ := filepath.Ext(files[j].Name)
			if extI == extJ {
				less = strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
			} else {
				less = extI < extJ
			}
		default:
			less = strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
		}

		if order == models.SortDescending {
			return !less
		}
		return less
	})
}

// IsHidden returns true if a filename should be considered hidden.
// On Unix systems, this means the name starts with a dot.
func IsHidden(name string) bool {
	return len(name) > 0 && name[0] == '.'
}
