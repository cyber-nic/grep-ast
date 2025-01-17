package goast

import (
	"testing"
)

func TestMatchIgnorePattern(t *testing.T) {
	tests := []struct {
		name           string
		value          string
		ignorePatterns map[string]bool
		expected       bool
	}{
		{
			name:           "Exact match for .git",
			value:          ".git",
			ignorePatterns: defaultIgnorePatterns,
			expected:       true,
		},
		{
			name:           "Directory match for .git/",
			value:          ".git/",
			ignorePatterns: defaultIgnorePatterns,
			expected:       true,
		},
		{
			name:           "Subdirectory match for .git/objects",
			value:          ".git/objects",
			ignorePatterns: defaultIgnorePatterns,
			expected:       true,
		},
		{
			name:           "Subdirectory match for .git/objects/34/",
			value:          ".git/objects/34/",
			ignorePatterns: defaultIgnorePatterns,
			expected:       true,
		},
		{
			name:           "File match for .git/objects/34/c51325a29a62565b0cdea41174dc3d13c860a1",
			value:          ".git/objects/34/c51325a29a62565b0cdea41174dc3d13c860a1",
			ignorePatterns: defaultIgnorePatterns,
			expected:       true,
		},
		{
			name:           "No match for unrelated file",
			value:          "not_git_file",
			ignorePatterns: defaultIgnorePatterns,
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchIgnorePattern(tt.value, tt.ignorePatterns)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for value %q", tt.expected, result, tt.value)
			}
		})
	}
}
