package grepast

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
		// Basic .git directory cases
		{
			name:           "Exact match for .git",
			value:          ".git",
			ignorePatterns: DefaultIgnorePatterns,
			expected:       true,
		},
		{
			name:           "Directory match for .git/",
			value:          ".git/",
			ignorePatterns: DefaultIgnorePatterns,
			expected:       true,
		},

		// Gitignore glob patterns
		{
			name:           "Wildcard match for all .log files",
			value:          "test.log",
			ignorePatterns: map[string]bool{"*.log": true},
			expected:       true,
		},
		{
			name:           "Wildcard match for nested .log file",
			value:          "logs/debug.log",
			ignorePatterns: map[string]bool{"*.log": true},
			expected:       true,
		},
		{
			name:           "Double asterisk match for any .pdf in any directory",
			value:          "path/to/deep/file.pdf",
			ignorePatterns: map[string]bool{"**/*.pdf": true},
			expected:       true,
		},
		{
			name:           "Directory contents with double asterisk",
			value:          "node_modules/some/deep/path/file.js",
			ignorePatterns: map[string]bool{"node_modules/**": true},
			expected:       true,
		},
		{
			name:           "Directory wildcard",
			value:          "logs/test.log",
			ignorePatterns: map[string]bool{"logs/*": true},
			expected:       true,
		},
		{
			name:           "Single character wildcard",
			value:          "file-a.txt",
			ignorePatterns: map[string]bool{"file-?.txt": true},
			expected:       true,
		},
		{
			name:           "Character class",
			value:          "file-1.txt",
			ignorePatterns: map[string]bool{"file-[0-9].txt": true},
			expected:       true,
		},
		// {
		// 	name:           "Negated pattern",
		// 	value:          "!important.log",
		// 	ignorePatterns: map[string]bool{"*.log": true, "!important.log": true},
		// 	expected:       false,
		// },
		// {
		// 	name:           "Negated character class",
		// 	value:          "file-a.txt",
		// 	ignorePatterns: map[string]bool{"file-[!0-9].txt": true},
		// 	expected:       true,
		// },

		// Common .gitignore patterns
		{
			name:           "Build directory",
			value:          "build/output.exe",
			ignorePatterns: map[string]bool{"build/": true},
			expected:       true,
		},
		{
			name:           "All files with extension",
			value:          "src/main.pyc",
			ignorePatterns: map[string]bool{"*.pyc": true},
			expected:       true,
		},
		{
			name:           "Hidden files",
			value:          ".env",
			ignorePatterns: map[string]bool{".*": true},
			expected:       true,
		},
		{
			name:           "OS specific files",
			value:          ".DS_Store",
			ignorePatterns: map[string]bool{".DS_Store": true},
			expected:       true,
		},
		{
			name:           "Log files in any directory",
			value:          "logs/debug/test.log",
			ignorePatterns: map[string]bool{"**/*.log": true},
			expected:       true,
		},
		{
			name:           "Coverage directory anywhere",
			value:          "src/tests/coverage/lcov.info",
			ignorePatterns: map[string]bool{"**/coverage/": true},
			expected:       true,
		},

		// Edge cases
		// {
		// 	name:           "Complex nested pattern",
		// 	value:          "tests/cache/jest/coverage/report.xml",
		// 	ignorePatterns: map[string]bool{"**/cache/**/coverage/**": true},
		// 	expected:       true,
		// },
		{
			name:           "Multiple asterisks in pattern",
			value:          "test/a/b/c/file.txt",
			ignorePatterns: map[string]bool{"test/**/file.txt": true},
			expected:       true,
		},
		{
			name:           "Directory vs file distinction",
			value:          "cache.log/important.txt",
			ignorePatterns: map[string]bool{"*.log": true},
			expected:       true,
		},
		{
			name:           "Escaped special characters",
			value:          "file[abc].txt",
			ignorePatterns: map[string]bool{"file\\[abc\\].txt": true},
			expected:       true,
		},
		// {
		// 	name:           "Path with spaces",
		// 	value:          "path/with space/file.txt",
		// 	ignorePatterns: map[string]bool{"path/**/*.txt": true},
		// 	expected:       true,
		// },
		// {
		// 	name:           "Path with spaces and special chars",
		// 	value:          "path/with space/and#hash/file.txt",
		// 	ignorePatterns: map[string]bool{"path/**/*.txt": true},
		// 	expected:       true,
		// },
		{
			name:           "Root level only",
			value:          "subdir/package.json",
			ignorePatterns: map[string]bool{"/package.json": true},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchIgnorePattern(tt.value, tt.ignorePatterns)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for value %q", tt.expected, result, tt.value)
			}
		})
	}
}

// TestGetLanguageFromFileName tests the GetLanguageFromFileName function
func TestGetLanguageFromFileName(t *testing.T) {

	// Test cases
	tests := []struct {
		name          string
		filePath      string
		expectedLang  string
		expectedError error
	}{
		{
			name:          "Valid Python File",
			filePath:      "script.py",
			expectedLang:  "python",
			expectedError: nil,
		},
		{
			name:          "Valid JavaScript File",
			filePath:      "app.js",
			expectedLang:  "javascript",
			expectedError: nil,
		},
		{
			name:          "Dockerfile",
			filePath:      "Dockerfile",
			expectedLang:  "Dockerfile",
			expectedError: nil,
		},
		{
			name:          "Unsupported File",
			filePath:      "unknown.xyz",
			expectedLang:  "",
			expectedError: ErrorUnrecognizedFiletype,
		},
		{
			name:          "File Without Extension",
			filePath:      "Makefile",
			expectedLang:  "",
			expectedError: ErrorUnsupportedLanguage,
		},
		{
			name:          "Valid Go File",
			filePath:      "main.go",
			expectedLang:  "go",
			expectedError: nil,
		},
		{
			name:          "Case-Insensitive Extension",
			filePath:      "Style.CSS",
			expectedLang:  "css",
			expectedError: nil,
		},
		{
			name:          "Valid TypeScript File",
			filePath:      "component.tsx",
			expectedLang:  "typescript",
			expectedError: nil,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang, detectedLang, err := GetLanguageFromFileName(tt.filePath)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected an error but got none")
				}
			} else {
				if err != nil && err != ErrorUnrecognizedFiletype {
					t.Errorf("unexpected error: %v", err)
				}
				if detectedLang != tt.expectedLang {
					t.Errorf("expected language %q, got %q", tt.expectedLang, detectedLang)
				}
				if detectedLang != "Dockerfile" && lang == nil {
					t.Errorf("expected a valid *sitter.Language instance, got nil")
				}
			}
		})
	}
}
