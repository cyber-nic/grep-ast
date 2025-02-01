package grepast

import (
	"fmt"
	"strings"
)

// options := TreeContextOptions{
// 	Color:                    false,
// 	ShowLineNumber:           false,
// 	ShowParentContext:        false,
// 	ShowChildContext:         false,
// 	ShowLastLine:             false,
// 	MarginPadding:            0,
// 	MarkLinesOfInterest:      false,
// 	HeaderMax:                10,
// 	ShowTopOfFileParentScope: false,
// 	LinesOfInterestPadding:   0,
// }

func getExampleSourceCode() []byte {
	return []byte(`package main

import "fmt"

func smallScope() {
	// short function that should be fully shown
	fmt.Println("short scope")
}

func largeScope() {
	// line 1
	// line 2
	// line 3
	// line 4
	fmt.Println("bigger scope!")
	// line 5
	// line 6
	// line 7
	// line 8
}

func main() {
	smallScope()
	largeScope()
}
`)
}

// ExampleTreeContext_Format_linesOfInterest demonstrates how NewTreeContext behaves
// when we only add Lines of Interest (LOI) but do not call AddContext().
func ExampleTreeContext_Format_linesOfInterest() {
	loi := map[int]struct{}{11: {}} // Line 11 is "func largeScope"

	options := TreeContextOptions{}

	sourceCode := getExampleSourceCode()

	tc, err := NewTreeContext("example.go", sourceCode, options)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	tc.AddLinesOfInterest(loi)
	fmt.Println(tc.Format()) // Expected output should be empty since AddContext was never called.

	// Output:
}

// ExampleTreeContext_Format_addContext shows how NewTreeContext behaves
// when AddContext() is called without setting lines of interest.
func ExampleTreeContext_Format_addContext() {
	options := TreeContextOptions{
		Color:                    false,
		Verbose:                  false,
		ShowLineNumber:           true,
		ShowParentContext:        true,
		ShowChildContext:         true,
		ShowLastLine:             false,
		MarginPadding:            0,
		MarkLinesOfInterest:      true,
		HeaderMax:                2,
		ShowTopOfFileParentScope: true,
		LinesOfInterestPadding:   0,
	}

	sourceCode := getExampleSourceCode()

	tc, err := NewTreeContext("example.go", sourceCode, options)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	tc.AddContext()
	fmt.Println(tc.Format()) // Expected output: empty since no lines of interest are defined.

	// Output:
}

// ExampleTreeContext_Format_singleLineOfInterest demonstrates a minimal NewTreeContext behavior with a single line of interest
func ExampleTreeContext_Format_singleLineOfInterest() {

	options := TreeContextOptions{}

	sourceCode := getExampleSourceCode()

	tc, err := NewTreeContext("example.go", sourceCode, options)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Line 22 is "func smallScope"
	loi := map[int]struct{}{22: {}}

	tc.AddLinesOfInterest(loi)
	tc.AddContext()
	got := tc.Format()

	// handle tabs
	got = strings.ReplaceAll(got, "\t", "    ")
	fmt.Println(got)

	// Output:
	// ⋮...
	// │    smallScope()
	// ⋮...
}

// ExampleTreeContext_Format_singleLineOfInterestWithContext demonstrates a TreeContext with a single line of interest with context
func ExampleTreeContext_Format_singleLineOfInterestWithContext() {

	options := TreeContextOptions{
		ShowLineNumber:           true,
		ShowParentContext:        true,
		ShowChildContext:         false,
		ShowLastLine:             true,
		MarginPadding:            0,
		MarkLinesOfInterest:      false,
		HeaderMax:                10,
		ShowTopOfFileParentScope: false,
		LinesOfInterestPadding:   0,
	}

	sourceCode := getExampleSourceCode()

	tc, err := NewTreeContext("example.go", sourceCode, options)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Line 11 is "func largeScope"
	loi := map[int]struct{}{22: {}}

	tc.AddLinesOfInterest(loi)
	tc.AddContext()
	got := tc.Format()

	// handle tabs
	got = strings.ReplaceAll(got, "\t", "    ")
	fmt.Println(got)

	// Output:
	// ⋮...
	//  22│func main() {
	//  23│    smallScope()
	//  24│    largeScope()
	//  25│}
	//  26│
}

// // ExampleTreeContext_Format_singleLineOfInterestWithChildContext demonstrates a TreeContext with a single line of interest with child context
// func ExampleTreeContext_Format_singleLineOfInterestWithChildContext() {

// 	options := TreeContextOptions{
// 		ShowLineNumber:           true,
// 		ShowParentContext:        true,
// 		ShowChildContext:         true,
// 		ShowLastLine:             true,
// 		MarginPadding:            0,
// 		MarkLinesOfInterest:      false,
// 		HeaderMax:                10,
// 		ShowTopOfFileParentScope: false,
// 		LinesOfInterestPadding:   0,
// 	}

// 	sourceCode := getExampleSourceCode()

// 	tc, err := NewTreeContext("example.go", sourceCode, options)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	// Line 11 is "func largeScope"
// 	loi := map[int]struct{}{22: {}}

// 	tc.AddLinesOfInterest(loi)
// 	tc.AddContext()
// 	got := tc.Format()

// 	// handle tabs
// 	got = strings.ReplaceAll(got, "\t", "    ")
// 	fmt.Println(got)

// 	// Output:
// 	// ⋮...
// 	//  22│func main() {
// 	//  23│    smallScope()
// 	//  24│    largeScope()
// 	//  25│}
// 	//  26│
// }

// ExampleTreeContext_Format_multipleLinesOfInterest demonstrates a minimal NewTreeContext behavior with multiple lines of interest
func ExampleTreeContext_Format_multipleLinesOfInterest() {

	options := TreeContextOptions{
		HeaderMax: 10,
	}

	sourceCode := getExampleSourceCode()

	tc, err := NewTreeContext("example.go", sourceCode, options)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Line 4 is "smallScope" and line 9 "largeScope"
	loi := map[int]struct{}{4: {}, 9: {}}

	tc.AddLinesOfInterest(loi)
	tc.AddContext()
	got := tc.Format()

	// handle tabs
	got = strings.ReplaceAll(got, "\t", "    ")
	fmt.Println(got)

	// Output:
	// ⋮...
	// │func smallScope() {
	// ⋮...
	// │func largeScope() {
	// ⋮...
}

// // ExampleTreeContext_Format_multipleLinesOfInterest demonstrates a minimal NewTreeContext behavior with multiple lines of interest
// func ExampleTreeContext_Format_multipleLinesOfInterest() {

// 	options := TreeContextOptions{
// 		ShowLineNumber:           false,
// 		ShowParentContext:        false,
// 		ShowChildContext:         false,
// 		ShowLastLine:             false,
// 		MarginPadding:            0,
// 		MarkLinesOfInterest:      false,
// 		HeaderMax:                10,
// 		LinesOfInterestPadding:   0,
// 		ShowTopOfFileParentScope: true,
// 	}

// 	sourceCode := getExampleSourceCode()

// 	tc, err := NewTreeContext("example.go", sourceCode, options)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	// Line 11 is "func largeScope"
// 	loi := map[int]struct{}{4: {}, 9: {}}

// 	tc.AddLinesOfInterest(loi)
// 	tc.AddContext()
// 	got := tc.Format()

// 	// handle tabs
// 	got = strings.ReplaceAll(got, "\t", "    ")
// 	fmt.Println(got)

//		// Output:
//		// ⋮...
//		// │func smallScope() {
//		// ⋮...
//		// │func largeScope() {
//		// ⋮...
//	}

// // TestTreeContext_addChildContext tests the addChildContext method of TreeContext.
// func TestTreeContext_addChildContext(t *testing.T) {
// 	sourceCode := getExampleSourceCode()

// 	options := TreeContextOptions{
// 		ShowChildContext: true,
// 	}

// 	tc, err := NewTreeContext("example.go", sourceCode, options)
// 	if err != nil {
// 		t.Fatalf("NewTreeContext() error = %v", err)
// 	}

// 	// Line 11 is "func largeScope"
// 	loi := map[int]struct{}{22: {}}
// 	tc.AddLinesOfInterest(loi)
// 	tc.AddContext()

// 	got := tc.Format()
// 	fmt.Println(got)

// 	expectedLines := map[int]struct{}{
// 		11: {}, 12: {}, 13: {}, 14: {}, 15: {}, 16: {}, 17: {}, 18: {}, 19: {},
// 	}

// 	for line := range expectedLines {
// 		if _, ok := tc.showLines[line]; !ok {
// 			t.Errorf("Expected line %d to be shown, but it was not", line)
// 		}
// 	}
// }
