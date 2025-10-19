package fileops

import (
	"testing"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"single byte", 1, "1 B"},
		{"bytes under 1KB", 500, "500 B"},
		{"exactly 1KB", 1024, "1.0 KB"},
		{"2KB", 2048, "2.0 KB"},
		{"half MB", 524288, "512.0 KB"},
		{"exactly 1MB", 1048576, "1.0 MB"},
		{"5.2MB", 5452595, "5.2 MB"},
		{"exactly 1GB", 1073741824, "1.0 GB"},
		{"2.5GB", 2684354560, "2.5 GB"},
		{"exactly 1TB", 1099511627776, "1.0 TB"},
		{"large TB", 5497558138880, "5.0 TB"},
		{"petabyte", 1125899906842624, "1.0 PB"},
		{"very large", 99999999999999999, "88.8 PB"}, // Test upper bound
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestGetParentDir(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"root directory", "/", "/"},
		{"root child", "/home", "/"},
		{"nested path", "/home/user/documents", "/home/user"},
		{"deep nested", "/a/b/c/d/e", "/a/b/c/d"},
		{"single level", "/tmp", "/"},
		{"trailing slash", "/home/user/", "/home"},
		{"multiple trailing slashes", "/home/user///", "/home"},
		{"relative path", "documents", "."},
		{"current dir", ".", "."},
		{"just filename", "file.txt", "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetParentDir(tt.path)
			if result != tt.expected {
				t.Errorf("GetParentDir(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsHidden(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"regular file", "document.txt", false},
		{"regular dir", "Documents", false},
		{"hidden file", ".bashrc", true},
		{"hidden dir", ".config", true},
		{"dot only", ".", true},
		{"double dot", "..", true},
		{"hidden with extension", ".gitignore", true},
		{"starts with dot in middle", "my.file", false},
		{"empty string", "", false},
		{"uppercase", "README.md", false},
		{"hidden uppercase", ".HIDDEN", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHidden(tt.filename)
			if result != tt.expected {
				t.Errorf("IsHidden(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}
