package models

import "testing"

func TestSortByString(t *testing.T) {
	tests := []struct {
		name     string
		sortBy   SortBy
		expected string
	}{
		{"name sort", SortByName, "Name"},
		{"size sort", SortBySize, "Size"},
		{"modtime sort", SortByModTime, "Modified"},
		{"extension sort", SortByExtension, "Extension"},
		{"invalid sort defaults to name", SortBy(999), "Name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.sortBy.String()
			if result != tt.expected {
				t.Errorf("SortBy.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}
