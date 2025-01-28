package grepast

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddLinesOfInterest(t *testing.T) {
	tc := &TreeContext{
		linesOfInterest: make(map[int]struct{}),
	}

	// Test adding lines of interest
	lines := map[int]struct{}{1: {}, 3: {}, 5: {}}
	tc.AddLinesOfInterest(lines)

	// Verify lines were added correctly
	assert.Equal(t, len(lines), len(tc.linesOfInterest))
	for line := range lines {
		_, exists := tc.linesOfInterest[line]
		assert.True(t, exists, "Line %d should be a line of interest", line)
	}
}

func TestAddContext(t *testing.T) {
	tc := &TreeContext{
		linesOfInterest: map[int]struct{}{
			10: {}, 20: {}, 30: {},
		},
		showLines:        make(map[int]struct{}),
		loiPad:           2,
		margin:           1,
		numLines:         50,
		lastLine:         true,
		parentContext:    true,
		childContext:     true,
		scopes:           make([]map[int]struct{}, 50),
		doneParentScopes: make(map[int]struct{}),
	}

	// Initialize dummy scope data
	for i := range tc.scopes {
		tc.scopes[i] = make(map[int]struct{})
	}

	tc.AddContext()

	// Verify padding was applied correctly
	for line := range tc.linesOfInterest {
		for i := line - tc.loiPad; i <= line+tc.loiPad; i++ {
			_, exists := tc.showLines[i]
			if i >= 0 && i < tc.numLines {
				assert.True(t, exists, "Line %d should be in showLines due to padding", i)
			}
		}
	}

	// Verify margin lines were added
	for i := 0; i < tc.margin; i++ {
		_, exists := tc.showLines[i]
		assert.True(t, exists, "Top margin line %d should be in showLines", i)
	}

	// Verify the last line is added
	_, exists := tc.showLines[tc.numLines-2]
	assert.True(t, exists, "Last line should be in showLines")

	// Verify parent context is added
	// (dependent on `addParentScopes` logic, can mock if needed)
}

// func TestAddChildContext(t *testing.T) {
// 	sourceCode := []byte(`
// 	import "fmt"

// 	func hello() {
// 		fmt.Println("Hello")
// 	}

// 	func main() {
// 		hello()
// 		hello()
// 	}
// 	`)

// 	// Create a new TreeContext
// 	tc, err := NewTreeContext("foo.go", sourceCode, TreeContextOptions{})
// 	if err != nil {
// 		t.Fatalf("Error creating TreeContext: %v", err)
// 	}

// 	tc.addChildContext(10)
// 	tc.linesOfInterest = map[int]struct{}{2: {}} // Mock lines of interest

// 	fmt.Println("format")
// 	fmt.Println(tc.Format())

// 	// // Verify lines corresponding to the child node are added
// 	// for i := 10; i <= 15; i++ {
// 	// 	_, exists := tc.showLines[i]
// 	// 	assert.True(t, exists, "Line %d should be in showLines as part of child context", i)
// 	// }

//		// // Test edge case where the line index is invalid
//		// tc.addChildContext(-1) // Should not panic or add lines
//		// tc.addChildContext(50) // Out of bounds
//		// assert.Equal(t, len(tc.showLines), 6, "No additional lines should be added for invalid indices")
//	}
//
// TestNewTreeContext exercises three usage patterns:
//  1. AddLinesOfInterest only, then Format()
//  2. AddContext only, then Format()
//  3. AddLinesOfInterest + AddContext, then Format()
func TestNewTreeContext(t *testing.T) {
	sourceCode := []byte(`
package main

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

	// We'll treat line 11 (0-based) as a line of interest ("func largeScope")
	// to illustrate expansions and ellipses.
	loi := map[int]struct{}{11: {}}

	t.Run("AddLinesOfInterestOnly", func(t *testing.T) {
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

		tc, err := NewTreeContext("example.go", sourceCode, options)
		assert.NoError(t, err)

		tc.AddLinesOfInterest(loi)

		out := tc.Format()

		// Because AddContext was never called, we never populated tc.showLines,
		// so Format() should return an empty string.
		assert.Equal(t, "", out, "Expected no output if we only add LOI but never call AddContext")
	})

	t.Run("AddContextOnly", func(t *testing.T) {
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
		tc, err := NewTreeContext("example.go", sourceCode, options)
		assert.NoError(t, err)

		// We never set any linesOfInterest, so AddContext() won't show anything.
		tc.AddContext()

		out := tc.Format()

		// Same reasoning: no linesOfInterest means no lines are added to showLines,
		// so the final snippet is empty.
		assert.Equal(t, "", out, "Expected no output if we have no LOI, even after AddContext")
	})

	t.Run("AddLinesOfInterestAndContext", func(t *testing.T) {
		options := TreeContextOptions{
			Color:                    false,
			Verbose:                  false,
			ShowLineNumber:           true,
			ShowParentContext:        true,
			ShowChildContext:         true,
			ShowLastLine:             false,
			MarginPadding:            0,
			MarkLinesOfInterest:      true,
			HeaderMax:                10,
			ShowTopOfFileParentScope: true,
			LinesOfInterestPadding:   0,
		}
		tc, err := NewTreeContext("example.go", sourceCode, options)
		assert.NoError(t, err)

		tc.AddLinesOfInterest(loi)
		tc.AddContext()
		out := tc.Format()

		fmt.Println("Output:\n" + out)

		// Now we expect expansions around line 11, child context from largeScope,
		// plus ellipses for skipped lines, etc.
		assert.True(t, strings.Contains(out, "largeScope()"),
			"Should contain line of interest text (func largeScope)")
		assert.True(t, strings.Contains(out, "fmt.Println(\"bigger scope!\")"),
			"Should contain child context lines from largeScope")

		assert.Contains(t, out, "...", "Should show ellipsis")
	})
}
