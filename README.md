# grep-ast-go

Grep soure code files and see matching lines with useful context that show how they fit into the code. See the loops, functions, methods, classes, etc
that contain all the matching lines. Get a sense of what's inside a matched class or function definition. You see relevant code from every layer of the
abstract syntax tree, above and below the matches.

By default, grep-AST recurses the current directory to search all source code files. It respects `.gitignore`, so it will usually "do the right thing" in most repos if you just do `grep-ast <regex>` without specifying any filenames.

Grep-AST is built with [tree-sitter](https://tree-sitter.github.io/tree-sitter/) and was inspired by [grep-ast](https://github.com/ast-grep/ast-grep) and [grep-ast](https://github.com/Aider-AI/grep-ast).

## Install

```bash
go get git@github.com:cyber-nic/grep-ast-go.git
```

## Usage

Basic usage:

```bash
grep-ast [pattern] [filenames...]
```

Full options list:

```
usage: grep_ast.py [-h] [-i] [--color] [--no-color] [--encoding ENCODING] [--languages] [--verbose]
                   [pat] [filenames ...]

positional arguments:
  pat                  the pattern to search for
  filenames            the files to display

options:
  -h, --help           show this help message and exit
  -i, --ignore-case    ignore case distinctions
  --color              force color printing
  --no-color           disable color printing
  --encoding ENCODING  file encoding
  --languages          print the parsers table
  --verbose            enable verbose output
```
