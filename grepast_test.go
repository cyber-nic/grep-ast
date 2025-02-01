package grepast

import (
	"fmt"
	"strings"
	"testing"
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

	go func() {
		smallScope()
	}()
}
`)
}

// ExampleTreeContext_Format_linesOfInterest demonstrates how NewTreeContext behaves
// when we only add Lines of Interest (LOI) but do not call AddContext().
func ExampleTreeContext_Format_linesOfInterest() {

	options := TreeContextOptions{}

	sourceCode := getExampleSourceCode()

	// Line 11 is "func largeScope"
	loi := map[int]struct{}{24: {}}

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

// TestTreeContext_addChildContext tests the addChildContext method of TreeContext.
func TestTreeContext_addChildContext(t *testing.T) {
	sourceCode := getExampleSourceCode()

	options := TreeContextOptions{
		ShowChildContext: true,
		// HeaderMax:        10,
	}

	tc, err := NewTreeContext("example.go", sourceCode, options)
	if err != nil {
		t.Fatalf("NewTreeContext() error = %v", err)
	}

	// Line 22 is "func largeScope"
	loi := map[int]struct{}{22: {}}
	tc.AddLinesOfInterest(loi)
	tc.AddContext()

	got := tc.Format()
	fmt.Println(got)

	expectedLines := map[int]struct{}{
		11: {}, 12: {}, 13: {}, 14: {}, 15: {}, 16: {}, 17: {}, 18: {}, 19: {},
	}

	for line := range expectedLines {
		if _, ok := tc.showLines[line]; !ok {
			t.Errorf("Expected line %d to be shown, but it was not", line)
		}
	}
}

// TestTreeContext_getLastLineOfScope tests the getLastLineOfScope method of TreeContext.
func TestTreeContext_getLastLineOfScope(t *testing.T) {
	sourceCode := getExampleSourceCode()

	options := TreeContextOptions{}

	tc, err := NewTreeContext("example.go", sourceCode, options)
	if err != nil {
		t.Fatalf("NewTreeContext() error = %v", err)
	}

	tests := []struct {
		line     int
		expected int
	}{
		{line: 0, expected: 25},
		{line: 6, expected: 7},   // smallScope function
		{line: 10, expected: 19}, // largeScope function
		{line: 15, expected: 19}, // fmt.Println("bigger scope!")
		{line: 20, expected: 20}, // empty line
		{line: 21, expected: 24}, // main function
		{line: 25, expected: 25}, // EOF

	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("line %d", tt.line), func(t *testing.T) {
			got := tc.getLastLineOfScope(tt.line)
			if got != tt.expected {
				t.Errorf("getLastLineOfScope(%d) = %d; want %d", tt.line, got, tt.expected)
			}
		})
	}
}

// TestMapKeysSorted tests the mapKeysSorted function.
func TestMapKeysSorted(t *testing.T) {
	tests := []struct {
		name     string
		input    map[int]struct{}
		expected []int
	}{
		{
			name:     "Empty map",
			input:    map[int]struct{}{},
			expected: []int{},
		},
		{
			name:     "Single element",
			input:    map[int]struct{}{1: {}},
			expected: []int{1},
		},
		{
			name:     "Multiple elements",
			input:    map[int]struct{}{3: {}, 1: {}, 2: {}},
			expected: []int{1, 2, 3},
		},
		{
			name:     "Already sorted",
			input:    map[int]struct{}{1: {}, 2: {}, 3: {}},
			expected: []int{1, 2, 3},
		},
		{
			name:     "Reverse order",
			input:    map[int]struct{}{3: {}, 2: {}, 1: {}},
			expected: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapKeysSorted(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("mapKeysSorted() = %v, want %v", got, tt.expected)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("mapKeysSorted() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

// TestSortBySize tests the sortBySize function.
func TestSortBySize(t *testing.T) {
	tests := []struct {
		name     string
		input    []struct{ start, end uint }
		expected []struct{ start, end uint }
	}{
		{
			name:     "Empty slice",
			input:    []struct{ start, end uint }{},
			expected: []struct{ start, end uint }{},
		},
		{
			name: "Single element",
			input: []struct{ start, end uint }{
				{start: 1, end: 2},
			},
			expected: []struct{ start, end uint }{
				{start: 1, end: 2},
			},
		},
		{
			name: "Multiple elements",
			input: []struct{ start, end uint }{
				{start: 1, end: 3},
				{start: 2, end: 5},
				{start: 0, end: 1},
			},
			expected: []struct{ start, end uint }{
				{start: 2, end: 5},
				{start: 1, end: 3},
				{start: 0, end: 1},
			},
		},
		{
			name: "Already sorted",
			input: []struct{ start, end uint }{
				{start: 2, end: 5},
				{start: 1, end: 3},
				{start: 0, end: 1},
			},
			expected: []struct{ start, end uint }{
				{start: 2, end: 5},
				{start: 1, end: 3},
				{start: 0, end: 1},
			},
		},
		{
			name: "Reverse order",
			input: []struct{ start, end uint }{
				{start: 0, end: 1},
				{start: 1, end: 3},
				{start: 2, end: 5},
			},
			expected: []struct{ start, end uint }{
				{start: 2, end: 5},
				{start: 1, end: 3},
				{start: 0, end: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortBySize(tt.input, func(item struct{ start, end uint }) int {
				return nodeSize(item.start, item.end)
			})
			for i := range tt.input {
				if tt.input[i] != tt.expected[i] {
					t.Errorf("sortBySize() = %v, want %v", tt.input, tt.expected)
				}
			}
		})
	}
}

// TestTreeContext_closeSmallGaps tests the closeSmallGaps method of TreeContext.
func TestTreeContext_closeSmallGaps(t *testing.T) {
	tests := []struct {
		name      string
		showLines map[int]struct{}
		lines     []string
		numLines  int
		wantLines map[int]struct{}
	}{
		{
			name: "No gaps",
			showLines: map[int]struct{}{
				0: {}, 1: {}, 2: {}, 3: {}, 4: {},
			},
			lines: []string{
				"line 1", "line 2", "line 3", "line 4", "line 5",
			},
			numLines: 5,
			wantLines: map[int]struct{}{
				0: {}, 1: {}, 2: {}, 3: {}, 4: {},
			},
		},
		{
			name: "Single line gap",
			showLines: map[int]struct{}{
				0: {}, 2: {}, 4: {},
			},
			lines: []string{
				"line 1", "line 2", "line 3", "line 4", "line 5",
			},
			numLines: 5,
			wantLines: map[int]struct{}{
				0: {}, 1: {}, 2: {}, 3: {}, 4: {},
			},
		},
		{
			name: "Multiple gaps",
			showLines: map[int]struct{}{
				0: {}, 2: {}, 5: {}, 7: {},
			},
			lines: []string{
				"line 1", "line 2", "line 3", "line 4", "line 5", "line 6", "line 7", "line 8",
			},
			numLines: 8,
			wantLines: map[int]struct{}{
				0: {}, 1: {}, 2: {}, 5: {}, 6: {}, 7: {},
			},
		},
		{
			name: "Adjacent blank lines",
			showLines: map[int]struct{}{
				0: {}, 2: {}, 4: {},
			},
			lines: []string{
				"line 1", "line 2", "line 3", "", "line 5",
			},
			numLines: 5,
			wantLines: map[int]struct{}{
				0: {}, 1: {}, 2: {}, 3: {}, 4: {},
			},
		},
		{
			name:      "No lines to show",
			showLines: map[int]struct{}{},
			lines: []string{
				"line 1", "line 2", "line 3", "line 4", "line 5",
			},
			numLines:  5,
			wantLines: map[int]struct{}{},
		},
		{
			name: "No trailing blank lines",
			showLines: map[int]struct{}{
				1: {}, 2: {}, 3: {},
			},
			lines: []string{
				"", "line 2", "", "line 4", "",
			},
			numLines: 5,
			wantLines: map[int]struct{}{
				1: {}, 2: {}, 3: {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &TreeContext{
				showLines: tt.showLines,
				lines:     tt.lines,
				numLines:  tt.numLines,
			}
			// fmt.Printf("before: %v\n", tc.showLines)
			tc.closeSmallGaps()
			// fmt.Printf("after: %v\n", tc.showLines)

			if len(tc.showLines) != len(tt.wantLines) {
				t.Errorf("closeSmallGaps() = %v, want %v", tc.showLines, tt.wantLines)
			}
			for line := range tt.wantLines {
				if _, ok := tc.showLines[line]; !ok {
					t.Errorf("Expected line %d to be shown, but it was not", line)
				}
			}
		})
	}
}

// TestTreeContext_addParentScopesSmall tests the addParentScopes method of TreeContext.
func TestTreeContext_addParentScopesSmall(t *testing.T) {
	sourceCode := getExampleSourceCode()

	options := TreeContextOptions{
		ShowParentContext: true,
		HeaderMax:         2,
	}

	tc, err := NewTreeContext("example.go", sourceCode, options)
	if err != nil {
		t.Fatalf("NewTreeContext() error = %v", err)
	}

	// Line 23 is "smallScope" of which the parent scope is "main"
	loi := map[int]struct{}{22: {}}
	tc.AddLinesOfInterest(loi)
	tc.AddContext()

	// got := tc.Format()
	// fmt.Println(tc.showLines)
	// fmt.Println(got)

	expectedLines := map[int]struct{}{
		21: {}, 22: {}, 23: {},
	}

	if len(expectedLines) != len(tc.showLines) {
		t.Errorf("addParentScopes() = %v, want %v", tc.showLines, expectedLines)
	}

	for line := range expectedLines {
		if _, ok := tc.showLines[line]; !ok {
			t.Errorf("Expected line %d to be shown, but it was not", line)
		}
	}
}

// TestTreeContext_addParentScopesLarge tests the addParentScopes method of TreeContext.
func TestTreeContext_addParentScopesLarge(t *testing.T) {
	sourceCode := getExampleSourceCode()

	options := TreeContextOptions{
		ShowParentContext: true,
		HeaderMax:         10,
		// ShowLastLine:      true,
	}

	tc, err := NewTreeContext("example.go", sourceCode, options)
	if err != nil {
		t.Fatalf("NewTreeContext() error = %v", err)
	}

	// Line 23 is "smallScope" of which the parent scope is "main"
	loi := map[int]struct{}{22: {}}
	tc.AddLinesOfInterest(loi)
	tc.AddContext()

	got := tc.Format()
	fmt.Println(tc.showLines)
	fmt.Println(got)

	expectedLines := map[int]struct{}{
		21: {}, 22: {}, 23: {}, 24: {}, 25: {},
	}

	if len(expectedLines) != len(tc.showLines) {
		t.Errorf("addParentScopes() = %v, want %v", tc.showLines, expectedLines)
	}

	for line := range expectedLines {
		if _, ok := tc.showLines[line]; !ok {
			t.Errorf("Expected line %d to be shown, but it was not", line)
		}
	}
}

// TestTreeContext_addParentScopes tests the addParentScopes method of TreeContext.
func TestTreeContext_addParentScopesNested(t *testing.T) {
	sourceCode := getExampleSourceCode()

	options := TreeContextOptions{
		ShowParentContext: true,
	}

	tc, err := NewTreeContext("example.go", sourceCode, options)
	if err != nil {
		t.Fatalf("NewTreeContext() error = %v", err)
	}

	// Line 26 is "smallScope" inside main -> go func
	loi := map[int]struct{}{26: {}}
	tc.AddLinesOfInterest(loi)
	tc.AddContext()

	// got := tc.Format()
	// fmt.Println(tc.showLines)
	// fmt.Println(got)

	expectedLines := map[int]struct{}{
		21: {}, 22: {}, 23: {}, 24: {}, 25: {}, 26: {}, 27: {}, 28: {},
	}

	if len(expectedLines) != len(tc.showLines) {
		t.Errorf("addParentScopes() = %v, want %v", tc.showLines, expectedLines)
	}

	for line := range expectedLines {
		if _, ok := tc.showLines[line]; !ok {
			t.Errorf("Expected line %d to be shown, but it was not", line)
		}
	}
}

// TestTreeContext_addParentScopes tests the addParentScopes method of TreeContext.
func TestTreeContext_addParentScopesWithFileTop(t *testing.T) {
	sourceCode := getExampleSourceCode()

	options := TreeContextOptions{
		ShowTopOfFileParentScope: true,
		ShowParentContext:        true,
		// ShowLastLine:             true,
		HeaderMax: 3,
	}

	tc, err := NewTreeContext("example.go", sourceCode, options)
	if err != nil {
		t.Fatalf("NewTreeContext() error = %v", err)
	}

	// Line 23 is "smallScope" of which the parent scope is "main"
	loi := map[int]struct{}{22: {}}
	tc.AddLinesOfInterest(loi)
	tc.AddContext()

	got := tc.Format()
	fmt.Println(tc.showLines)
	fmt.Println(got)

	expectedLines := map[int]struct{}{
		21: {}, 22: {}, 23: {}, 24: {}, 25: {}, 26: {}, 27: {}, 28: {},
	}

	if len(expectedLines) != len(tc.showLines) {
		t.Errorf("addParentScopes() = %v, want %v", tc.showLines, expectedLines)
	}

	for line := range expectedLines {
		if _, ok := tc.showLines[line]; !ok {
			t.Errorf("Expected line %d to be shown, but it was not", line)
		}
	}
}

// // ExampleTreeContext_Format_singleLineOfInterestWithContext demonstrates a TreeContext with a single line of interest with context
// func ExampleTreeContext_Format_singleLineOfInterestWithContext() {

// 	options := TreeContextOptions{
// 		ShowLineNumber:    true,
// 		ShowParentContext: true,
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

// // ExampleTreeContext_Format_singleLineOfInterestWithChildContext demonstrates a TreeContext with a single line of interest with child context
// func ExampleTreeContext_Format_singleLineOfInterestWithChildContext() {

// 	options := TreeContextOptions{
// 		ShowLineNumber:    true,
// 		ShowParentContext: true,
// 		ShowChildContext:  true,
// 		// HeaderMax:         6,
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
// 	//	func smallScope() {
// 	//		// short function that should be fully shown
// 	//		fmt.Println("short scope")
// 	//	}
// 	//
// 	//	22│func main() {
// 	//	23│    smallScope()
// 	//	24│    largeScope()
// 	//	25│}
// 	//	26│
// }
