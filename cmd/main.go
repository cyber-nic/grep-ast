package main

import (
	"fmt"
	"os"

	goast "github.com/cyber-nic/grep-ast-go"
)

func main() {
	// if len(os.Args) != 2 {
	// 	fmt.Fprintf(os.Stderr, "Usage: %s <file/directory path>\n", os.Args[0])
	// 	os.Exit(1)
	// }

	// path := os.Args[1]

	path := "cmd/main.go"
	info, err := os.Stat(path)
	if err != nil {
		panic(fmt.Errorf("Error accessing path: %v", err))
	}

	if info.IsDir() {
		fmt.Println("not accepting dirs at this time")
		return
	}

	// read file content
	code, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("Error reading file: %v", err))
	}

	tc, err := goast.NewTreeContext(
		path,
		string(code),
		true, // color
		true, // verbose
		true, // lineNumber
		true, // parentContext
		true, // childContext
		true, // lastLine
		3,    // margin
		true, // markLOIs
		10,   // headerMax
		true, // showTopOfFileParentScope
		1,    // loiPad
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Example: grep for "main"
	found := tc.Grep("main", false)
	tc.AddLinesOfInterest(found)

	// Add context
	// tc.AddContext()

	// Format output
	out := tc.Format()
	fmt.Println(out)
}
