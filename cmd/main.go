package main

import (
	"fmt"
	"os"

	goast "github.com/cyber-nic/grep-ast-go"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file/directory path>\n", os.Args[0])
		os.Exit(1)
	}

	path := os.Args[1]
	_, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing path: %v\n", err)
		os.Exit(1)
	}

	tree, err := goast.NewTreeContext(path)
	if err != nil {
		panic(fmt.Errorf("failed to create tree context for %s: %v", path, err))
	}

	fmt.Println(tree.Path)
	for _, node := range tree.Nodes {
		fmt.Println(node.Path())
	}
}
