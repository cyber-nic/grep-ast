package grepast

import (
	"testing"
)

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
