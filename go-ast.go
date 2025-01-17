package goast

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

const (
	ignoreFilename = ".gitignore"
)

// TreeContext holds the entire tree
type TreeContext struct {
	Nodes          map[string]TreeContextNode
	IgnorePatterns map[string]bool
	Path           string
}

type TreeContextNode interface {
	Kind() string
	Path() string
}

// TreeContextDir holds the parsed file content and its structure
type TreeContextDir struct {
	children []TreeContextNode
	path     string
}

func (t TreeContextDir) Kind() string {
	return "directory"
}

func (t TreeContextDir) Path() string {
	return t.path
}

// TreeContextFile holds the parsed file content and its structure
type TreeContextFile struct {
	path     string
	content  string
	lines    []string
	numLines int
}

func (t TreeContextFile) Kind() string {
	return "file"
}

func (t TreeContextFile) Path() string {
	return t.path
}

// NewTreeContext creates a new TreeContext for a given path
func NewTreeContext(path string) (*TreeContext, error) {

	// Default ignore patterns
	ignorePatterns := defaultIgnorePatterns

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %v", err)
	}

	// tr@ck - if no <ignoreFilename> then find the nearest if available

	// load user ast ignore file if available
	if _, err := os.Stat(filepath.Join(cwd, ignoreFilename)); err == nil {
		// Load the ignore list
		astIgnorePath := filepath.Join(cwd, ignoreFilename)
		userIgnorePatterns, err := loadIgnoreASTFile(astIgnorePath)
		if err != nil {
			log.Printf("Failed to read ast ignore file (%s): %v", astIgnorePath, err)
		}

		// add the ignore patterns to the default set
		for _, pattern := range userIgnorePatterns {
			ignorePatterns[pattern] = true
		}
	}

	tree, err := buildTreeContext(path, ignorePatterns)

	tc := TreeContext{
		Nodes:          tree,
		Path:           path,
		IgnorePatterns: ignorePatterns,
	}

	return &tc, nil

	// // Get parser based on file extension
	// parser, err := getParser(filename)
	// if err != nil {
	// 	return nil, err
	// }

	// // Parse the content
	// tree := parser.Parse([]byte(content), nil)
	// if tree == nil {
	// 	return nil, fmt.Errorf("failed to parse %s", filename)
	// }

}

// Format returns the formatted content of the file
func (tc *TreeContext) Format() string {
	// For now, just return the raw content
	return "to do"

}

// getParser returns the appropriate tree-sitter parser for the given filename
func getParser(filename string) (*sitter.Parser, error) {
	lang, err := getLanguageFromFileName(filename)
	if err != nil {
		return nil, err
	}

	// Return an error if the language is not supported
	if lang == nil {
		return nil, fmt.Errorf("unsupported file: %s", filename)
	}

	parser := sitter.NewParser()
	if parser == nil {
		return nil, fmt.Errorf("failed to create parser")
	}
	return parser, nil
}

// // buildTreeContext builds the tree context for a given path
// func buildTreeContext(path string, ignorePatterns map[string]bool) ([]TreeContextNode, error) {

// 	nodes := []TreeContextNode{}

// 	entries, err := os.ReadDir(path)
// 	if err != nil {
// 		return []TreeContextNode{}, err
// 	}

// 	for _, entry := range entries {
// 		if matchIgnorePattern(entry.Name(), ignorePatterns) {
// 			continue
// 		}

// 		if entry.IsDir() {
// 			children, err := buildTreeContext(filepath.Join(path, entry.Name()), ignorePatterns)
// 			if err != nil {
// 				return nil, fmt.Errorf("failed to build tree context for %s: %v", path, err)
// 			}

// 			nodes = append(nodes, TreeContextDir{
// 				path:     entry.Name(),
// 				children: children,
// 			})
// 			continue
// 		}

// 		content, err := os.ReadFile(path)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to build tree context for %s: %v", path, err)
// 		}

// 		lines := []string{}
// 		for _, line := range strings.Split(string(content), "\n") {
// 			lines = append(lines, line)
// 		}

// 		nodes = append(nodes, TreeContextFile{
// 			path:     entry.Name(),
// 			content:  string(content),
// 			lines:    lines,
// 			numLines: len(lines),
// 		})

// 	}

// 	return nodes, nil
// }

// buildTreeContext builds the tree context for a given path
func buildTreeContext(path string, ignorePatterns map[string]bool) (map[string]TreeContextNode, error) {
	nodes := make(map[string]TreeContextNode)

	err := filepath.WalkDir(path, func(currPath string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", currPath, err)
		}

		relPath, err := filepath.Rel(path, currPath)
		if err != nil {
			return fmt.Errorf("error calculating relative path: %w", err)
		}

		if relPath == "." || matchIgnorePattern(currPath, ignorePatterns) {
			return nil
		}

		if d.IsDir() {
			nodes[relPath] = TreeContextDir{
				path:     relPath,
				children: []TreeContextNode{}, // Will be populated later if needed
			}
			return nil
		}

		content, err := os.ReadFile(currPath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", currPath, err)
		}

		lines := strings.Split(string(content), "\n")
		nodes[relPath] = TreeContextFile{
			path:     relPath,
			content:  string(content),
			lines:    lines,
			numLines: len(lines),
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return nodes, nil
}
