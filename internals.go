package grepast

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	sitter_bash "github.com/tree-sitter/tree-sitter-bash/bindings/go"
	sitter_c_sharp "github.com/tree-sitter/tree-sitter-c-sharp/bindings/go"
	sitter_css "github.com/tree-sitter/tree-sitter-css/bindings/go"
	sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	sitter_html "github.com/tree-sitter/tree-sitter-html/bindings/go"
	sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

var (
	ErrorUnrecognizedFiletype = fmt.Errorf("unrecognized file type")
	ErrorUnsupportedLanguage  = fmt.Errorf("unsupported language")
)

var extensionMap = map[string]string{
	".bash":   "bash",
	".cc":     "cpp",
	".cl":     "commonlisp",
	".c":      "c",
	".cpp":    "cpp",
	".cs":     "c_sharp",
	".csm":    "scheme",
	".css":    "css",
	".el":     "elisp",
	".ex":     "elixir",
	".elm":    "elm",
	".et":     "embedded_template",
	".erl":    "erlang",
	".gomod":  "gomod",
	".go":     "go",
	".hack":   "hack",
	".hcl":    "hcl",
	".hs":     "haskell",
	".html":   "html",
	".java":   "java",
	".jl":     "julia",
	".js":     "javascript",
	".json":   "json",
	".jsx":    "javascript",
	".kt":     "kotlin",
	".lua":    "lua",
	".mjs":    "javascript",
	".mk":     "make",
	".ml":     "ocaml",
	".m":      "objc",
	".php":    "php",
	".pl":     "perl",
	".py":     "python",
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
	".ts":     "typescript",
	".tsx":    "typescript",
	".yaml":   "yaml",
}

// GetLanguageFromFileName maps file name to tree-sitter Language instances
func GetLanguageFromFileName(path string) (*sitter.Language, string, error) {

	if strings.EqualFold(filepath.Base(path), "Dockerfile") {
		return nil, "Dockerfile", nil
	}

	ext := strings.ToLower(filepath.Ext(path))

	if lang, ok := extensionMap[ext]; ok {
		switch lang {
		case "bash":
			return sitter.NewLanguage(sitter_bash.Language()), lang, nil
		case "c_sharp":
			return sitter.NewLanguage(sitter_c_sharp.Language()), lang, nil
		case "css":
			return sitter.NewLanguage(sitter_css.Language()), lang, nil
		case "go":
			return sitter.NewLanguage(sitter_go.Language()), lang, nil
		case "java":
			return sitter.NewLanguage(sitter_java.Language()), lang, nil
		case "javascript":
			return sitter.NewLanguage(sitter_javascript.Language()), lang, nil
		case "html":
			return sitter.NewLanguage(sitter_html.Language()), lang, nil
		case "python":
			return sitter.NewLanguage(sitter_python.Language()), lang, nil
		case "typescript":
			return sitter.NewLanguage(sitter_typescript.LanguageTypescript()), lang, nil
		case "rust":
			return sitter.NewLanguage(sitter_rust.Language()), lang, nil
		default:
			return nil, "", ErrorUnsupportedLanguage
		}
	}

	return nil, "", ErrorUnrecognizedFiletype
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
	// Normalize path separators
	value = filepath.ToSlash(value)

	for pattern, ignore := range ignorePatterns {
		if ignore && matchPattern(value, pattern) {
			return true
		}
	}

	return false
}

func matchPattern(value, pattern string) bool {
	// Normalize path separators so that all are "/"
	value = filepath.ToSlash(value)
	pattern = filepath.ToSlash(pattern)

	// 1) Handle patterns starting with "**/" first (e.g. "**/coverage/")
	if strings.HasPrefix(pattern, "**/") {
		// Trim the leading "**/"
		pattern = strings.TrimPrefix(pattern, "**/")
		// Split the value into path segments
		parts := strings.Split(value, "/")
		// Try matching `pattern` against every possible subpath
		for i := range parts {
			subpath := strings.Join(parts[i:], "/")
			if matchPattern(subpath, pattern) {
				return true
			}
		}
		return false
	}

	// 2) Handle directory-specific patterns (those that end with "/")
	//    e.g. "src/foo/"
	if strings.HasSuffix(pattern, "/") {
		// Remove the trailing "/"
		pattern = strings.TrimSuffix(pattern, "/")
		// Match if value is exactly "pattern" or starts with "pattern/"
		return value == pattern || strings.HasPrefix(value, pattern+"/")
	}

	// 3) Handle patterns with "**" in the middle, e.g. "src/**/foo"
	if strings.Contains(pattern, "**") {
		segments := strings.SplitN(pattern, "**", 2)
		// We only split once: ["src/", "/foo"] for "src/**/foo"
		prefix := segments[0]
		suffix := segments[1]

		// If value doesn't start with prefix, it's an immediate miss
		if !strings.HasPrefix(value, prefix) {
			return false
		}

		// Remove prefix and check if remainder ends with suffix
		remainder := value[len(prefix):]
		return strings.HasSuffix(remainder, suffix)
	}

	// 4) Handle single-segment wildcards (no slashes), e.g. "*.go"
	if !strings.Contains(pattern, "/") && strings.Contains(pattern, "*") {
		return matchBasename(value, pattern)
	}

	// 5) Fall back to a direct pattern match
	matched, err := filepath.Match(pattern, value)
	return err == nil && matched
}

func matchBasename(value, pattern string) bool {
	// Match against any component of the path
	parts := strings.Split(value, "/")
	for _, part := range parts {
		matched, err := filepath.Match(pattern, part)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// PrintStruct prints a struct as JSON.
func PrintStruct(w io.Writer, t interface{}) {
	j, _ := json.MarshalIndent(t, "", "  ")
	fmt.Fprintln(w, string(j))
}

func PrintStructOut(t interface{}) {
	PrintStruct(os.Stdout, t)
}
