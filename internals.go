package grepast

import (
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

// PrintStruct prints a struct as JSON.
func PrintStruct(w io.Writer, t interface{}) {
	j, _ := json.MarshalIndent(t, "", "  ")
	fmt.Fprintln(w, string(j))
}

func PrintStructOut(t interface{}) {
	PrintStruct(os.Stdout, t)
}
