package main

import (
	"fmt"
	"os"
	"path/filepath"

	goignore "github.com/cyber-nic/go-gitignore"
	grepast "github.com/cyber-nic/grep-ast"
)

func main() {
	// Check for the correct number of arguments
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintf(os.Stderr, "Usage: grep-ast search_pattern <file/directory path>\n")
		return
	}

	var err error
	rootPath := "."

	// Get the search query
	searchQuery := os.Args[1]

	// Get the root path
	if len(os.Args) == 2 {
		// If only one argument, set rootPath to current working directory
		rootPath, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting current working directory: %v\n", err)
			os.Exit(1)
		}
	} else if len(os.Args) == 3 {
		if os.Args[2] == "." {
			// Convert "." to current working directory
			rootPath, err = os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error getting current working directory: %v\n", err)
				os.Exit(1)
			}
		} else {
			rootPath = os.Args[2]
		}
	}

	// Get the custom ignore file path
	customIgnoreFilePath := filepath.Join(rootPath, ".astignore")

	// Load the ignore file
	gi, err := goignore.CompileIgnoreFile(customIgnoreFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading ignore file: %v\n", err)
	}

	// Walk the directory
	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		// Skip errors
		if err != nil {
			return err
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip files that match the ignore patterns
		if gi.MatchesPath(path) {
			return nil
		}

		rel, err := filepath.Rel(rootPath, path)
		if err != nil {
			return nil
		}

		parseAndGrep(rel, searchQuery)
		return nil
	})

	if err != nil {
		panic(fmt.Errorf("Error walking the path: %v", err))
	}

}

func parseAndGrep(filePath, search string) error {
	source, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	// Attempt to create a TreeContext. Non-Go files may fail.
	tc, err := grepast.NewTreeContext(filePath, source, grepast.TreeContextOptions{
		Color:                    true,
		Verbose:                  false,
		ShowLineNumber:           true,
		ShowParentContext:        true,
		ShowChildContext:         true,
		ShowLastLine:             false,
		MarginPadding:            3,
		MarkLinesOfInterest:      true,
		HeaderMax:                10,
		ShowTopOfFileParentScope: true,
		LinesOfInterestPadding:   1,
	})
	if err != nil {
		return fmt.Errorf("error parsing file %s: %v", filePath, err)
	}

	found := tc.Grep(search, false)
	tc.AddLinesOfInterest(found)
	tc.AddContext()

	// Format and print the output
	out := tc.Format()

	fmt.Printf("\n%s:%s\n", filePath, out)

	return nil
}
