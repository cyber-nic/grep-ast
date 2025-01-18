package grepast

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

var (
	languageMap map[string]*sitter.Language
)

var extensionMap = map[string]string{
	".py":     "python",
	".js":     "javascript",
	".mjs":    "javascript",
	".go":     "go",
	".tsx":    "typescript",
	".ts":     "typescript",
	".jsx":    "javascript",
	".bash":   "bash",
	".c":      "c",
	".cpp":    "cpp",
	".cc":     "cpp",
	".cs":     "c_sharp",
	".cl":     "commonlisp",
	".css":    "css",
	".el":     "elisp",
	".ex":     "elixir",
	".elm":    "elm",
	".et":     "embedded_template",
	".erl":    "erlang",
	".gomod":  "gomod",
	".hack":   "hack",
	".hs":     "haskell",
	".hcl":    "hcl",
	".html":   "html",
	".java":   "java",
	".json":   "json",
	".jl":     "julia",
	".kt":     "kotlin",
	".lua":    "lua",
	".mk":     "make",
	".m":      "objc",
	".ml":     "ocaml",
	".pl":     "perl",
	".php":    "php",
	".ql":     "ql",
	".r":      "r",
	".regex":  "regex",
	".rst":    "rst",
	".rb":     "ruby",
	".rs":     "rust",
	".scala":  "scala",
	".sql":    "sql",
	".sqlite": "sqlite",
	".toml":   "toml",
	".yaml":   "yaml",
}

func init() {
	languageMap = map[string]*sitter.Language{
		"go":         sitter.NewLanguage(tree_sitter_go.Language()),
		"javascript": sitter.NewLanguage(tree_sitter_javascript.Language()),
		"python":     sitter.NewLanguage(tree_sitter_python.Language()),
		"typescript": sitter.NewLanguage(tree_sitter_typescript.LanguageTypescript()),
	}
}

// GetLanguageFromFileName maps file name to tree-sitter Language instances
func GetLanguageFromFileName(path string) (*sitter.Language, error) {
	if strings.EqualFold(filepath.Base(path), "Dockerfile") {
		return nil, nil
	}

	ext := strings.ToLower(filepath.Ext(path))

	if lang, ok := extensionMap[ext]; ok {
		return languageMap[lang], nil
	}

	return nil, fmt.Errorf("unrecognized or unsupported file type (%s): %s", path, ext)
}

// loadIgnoreList reads the ignore file and returns the list of patterns to ignore
func loadIgnoreList(ignoreFilePath string) ([]string, error) {
	ignoreList := make(map[string]struct{})

	file, err := os.Open(ignoreFilePath)
	if err != nil {
		return []string{}, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			ignoreList[line] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read ignore file: %v", err)
	}

	keys := make([]string, 0, len(ignoreList))

	// convert hash to array
	for key := range ignoreList {
		keys = append(keys, key)
	}

	return keys, nil
}

// Default ignore patterns
var DefaultIgnorePatterns = map[string]bool{
	".astignore":    true,
	".git/":         true,
	".gitignore":    true,
	".venv/":        true,
	"venv/":         true,
	"testdata/":     true,
	"go.sum":        true,
	"node_modules/": true,
	"dist/":         true,
	".*":            true,
}

func MatchIgnorePattern(value string, ignorePatterns map[string]bool) bool {
	for pattern, _ := range ignorePatterns {
		if strings.HasSuffix(pattern, "/") {
			// Directory pattern: check if name matches or is within the directory
			if strings.HasPrefix(value, strings.TrimSuffix(pattern, "/")) {
				return true
			}
		} else if strings.HasSuffix(pattern, "/*") {
			// Directory contents pattern: match files and subdirectories within
			base := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(value, base) {
				return true
			}
		} else {
			// Exact match or prefix match
			if strings.HasPrefix(value, pattern) {
				return true
			}
		}
	}
	return false
}
