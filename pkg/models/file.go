package models

import (
	"os"
	"time"
)

// FileInfo represents information about a file or directory.
// This is Warren's internal representation used throughout the application.
type FileInfo struct {
	// Name is the base name of the file (e.g., "document.txt")
	Name string

	// Path is the full absolute path to the file
	Path string

	// Size in bytes
	Size int64

	// IsDir indicates whether this is a directory
	IsDir bool

	// IsSymlink indicates whether this is a symbolic link
	IsSymlink bool

	// SymlinkTarget is the target path if this is a symlink
	SymlinkTarget string

	// Permissions is the file mode and permission bits
	Permissions os.FileMode

	// ModTime is the last modification time
	ModTime time.Time

	// IsHidden indicates if the file should be considered hidden
	// (starts with . on Unix systems)
	IsHidden bool

	// MimeType is the detected MIME type (filled in lazily if needed)
	MimeType string
}

// FileList represents a collection of files in a directory.
type FileList struct {
	// Path is the directory path
	Path string

	// Files is the list of files and subdirectories
	Files []FileInfo

	// Error if the directory could not be read
	Error error
}

// SortBy represents different sorting options for files.
type SortBy int

const (
	SortByName SortBy = iota
	SortBySize
	SortByModTime
	SortByExtension
)

// SortOrder represents ascending or descending sort order.
type SortOrder int

const (
	SortAscending SortOrder = iota
	SortDescending
)
