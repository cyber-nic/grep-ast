package grepast

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// TreeContext stores context about source code lines, parsing, scopes, and line-of-interest management.
type TreeContext struct {
	filename                 string             // Name of the file being processed.
	source                   []byte             // Source code content as a byte array.
	color                    bool               // Whether to use color for highlighted output.
	verbose                  bool               // Whether to enable verbose output for debugging.
	showLineNumber           bool               // Whether to include line numbers in the output.
	showLastLine             bool               // Whether to always include the larger context's last line in the output.
	margin                   int                // Number of lines to include as a margin at the top of the output.
	markLOIs                 bool               // Whether to visually mark lines of interest (LOI).
	headerMax                int                // Maximum number of header lines to display.
	loiPad                   int                // Number of lines of padding around lines of interest.
	showTopOfFileParentScope bool               // Whether to include the parent scope starting from the top of the file.
	parentContext            bool               // Whether to include parent context in the output.
	showChildContext         bool               // Whether to include child context in the output.
	lines                    []string           // Source code split into individual lines.
	numLines                 int                // Total number of lines in the source code (including an optional trailing newline adjustment).
	outputLines              map[int]string     // Map of output lines, optionally with highlights.
	scopes                   []map[int]struct{} // Tracks scope relationships by line.
	header                   [][]int            // Each element is a slice representing [startLine, endLine] of headers.
	nodes                    [][]*sitter.Node   // Tracks parse-tree nodes indexed by their start line.
	showLines                map[int]struct{}   // Lines to show in the final output.
	linesOfInterest          map[int]struct{}   // Lines explicitly marked as "lines of interest" (LOI).
	doneParentScopes         map[int]struct{}   // Tracks parent scopes that have already been processed.
}

// TreeContextOptions specifies various options for initializing TreeContext.
type TreeContextOptions struct {
	Color                    bool // Use colored output for matches or highlights.
	HeaderMax                int  // Maximum number of header lines to display.
	LinesOfInterestPadding   int  // Number of lines of padding around each line of interest.
	MarginPadding            int  // Number of lines to add as a margin at the top of the output.
	MarkLinesOfInterest      bool // Visually mark lines of interest (LOI) in the output.
	ShowChildContext         bool // Show the child scope of lines of interest in the output.
	ShowLastLine             bool // Always include the overall context's last line in the output.
	ShowLineNumber           bool // Include line numbers in the output.
	ShowParentContext        bool // Show the parent scope of lines of interest in the output.
	ShowTopOfFileParentScope bool // Always include the top-most parent scope from the file's beginning.
	Verbose                  bool // Enable verbose mode for additional debugging or insights.
}

// NewTreeContext is the Go-equivalent constructor for TreeContext.
// It initializes the context for analyzing and working with source code.
func NewTreeContext(filename string, source []byte, options TreeContextOptions) (*TreeContext, error) {
	// Get the language from the filename.
	// Determines the programming language to use for parsing based on the file extension.
	lang, _, err := GetLanguageFromFileName(filename)
	if err != nil {
		return nil, err // Return an error if the file type cannot be recognized.
	}

	// Return an error if the language is not supported.
	if lang == nil {
		return nil, fmt.Errorf("unrecognized or unsupported file type (%s)", filename)
	}

	// Initialize Tree-sitter parser for parsing source code into an abstract syntax tree (AST).
	parser := sitter.NewParser()
	parser.SetLanguage(lang) // Set the parser's language to match the file type.

	// Parse the source code into a syntax tree.
	tree := parser.Parse(source, nil)

	// Retrieve the root node of the syntax tree for traversal.
	rootNode := tree.RootNode()

	// Split the source code into lines for easier processing.
	lines := strings.Split(string(source), "\n")

	// Determine the total number of lines
	numLines := len(lines)

	// Initialize scopes, headers, and nodes for tracking relationships and parsing metadata.
	scopes := make([]map[int]struct{}, numLines)
	header := make([][]int, numLines)         // Track start and end lines for each header.
	nodes := make([][]*sitter.Node, numLines) // Track AST nodes by their starting line.

	for i := 0; i <= numLines-1; i++ {
		scopes[i] = make(map[int]struct{})
		header[i] = []int{0, 0}
		nodes[i] = []*sitter.Node{}
	}

	// Create and populate the TreeContext object with initialized values.
	tc := &TreeContext{
		filename:                 filename,
		source:                   source,
		color:                    options.Color,
		verbose:                  options.Verbose,
		showLineNumber:           options.ShowLineNumber,
		parentContext:            options.ShowParentContext,
		showChildContext:         options.ShowChildContext,
		showLastLine:             options.ShowLastLine,
		margin:                   options.MarginPadding,
		markLOIs:                 options.MarkLinesOfInterest,
		headerMax:                options.HeaderMax,
		loiPad:                   options.LinesOfInterestPadding,
		showTopOfFileParentScope: options.ShowTopOfFileParentScope,
		lines:                    lines,
		numLines:                 numLines,
		outputLines:              make(map[int]string),
		scopes:                   scopes,
		header:                   header,
		nodes:                    nodes,
		showLines:                make(map[int]struct{}),
		linesOfInterest:          make(map[int]struct{}),
		doneParentScopes:         make(map[int]struct{}),
	}

	// Walk through the parse tree to populate headers, scopes, and nodes.
	tc.walkTree(rootNode, 0)

	// Perform additional processing on scopes and headers after tree traversal.
	tc.postWalkProcessing()

	// Return the initialized TreeContext object.
	return tc, nil
}

// postWalkProcessing sets header ranges and optionally prints scopes.
func (tc *TreeContext) postWalkProcessing() {
	// print and set header ranges
	var scopeWidth int

	if tc.verbose {
		// find the maximum width for printing scopes
		for i := 0; i < tc.numLines; i++ {
			scopeStr := fmt.Sprintf("%v", mapKeysSorted(tc.scopes[i]))
			if len(scopeStr) > scopeWidth {
				scopeWidth = len(scopeStr)
			}
		}
	}

	for i := 0; i < tc.numLines; i++ {
		headerSlice := tc.header[i]
		if len(headerSlice) < 2 {
			// default
			tc.header[i] = []int{i, i + 1}
		} else {
			size := headerSlice[0]
			headStart := headerSlice[1]
			headEnd := headerSlice[1] + 1
			if len(headerSlice) > 2 {
				headEnd = headerSlice[2]
			}
			if size > tc.headerMax {
				headEnd = headStart + tc.headerMax
			}
			tc.header[i] = []int{headStart, headEnd}
		}

		if tc.verbose && i < tc.numLines {
			scopeStr := fmt.Sprintf("%v", mapKeysSorted(tc.scopes[i]))
			if i < len(tc.lines) {
				lineStr := tc.lines[i]
				fmt.Printf("%-*s %3d %s\n", scopeWidth, scopeStr, i, lineStr)
			}
		}
	}
}

// Grep finds lines matching a pattern and highlights them.
func (tc *TreeContext) Grep(pat string, ignoreCase bool) map[int]struct{} {
	found := make(map[int]struct{})
	if ignoreCase {
		// Go's regex doesn't have "IGNORECASE" as a flag (like Python),
		// you compile different patterns or use (?i).
		pat = "(?i)" + pat
	}
	re := regexp.MustCompile(pat)

	for i, line := range tc.lines {
		if re.FindStringIndex(line) != nil {
			// highlight
			if tc.color {
				highlighted := re.ReplaceAllStringFunc(line, func(m string) string {
					return fmt.Sprintf("\033[1;31m%s\033[0m", m)
				})
				tc.outputLines[i] = highlighted
			}
			found[i] = struct{}{}
		}
	}
	return found
}

// AddLinesOfInterest adds lines of interest.
func (tc *TreeContext) AddLinesOfInterest(lineNums map[int]struct{}) {
	for ln := range lineNums {
		tc.linesOfInterest[ln] = struct{}{}
	}
}

// AddContext expands lines to show (showLines) based on linesOfInterest.
func (tc *TreeContext) AddContext() {
	if len(tc.linesOfInterest) == 0 {
		return
	}

	// Ensure all linesOfInterest are in showLines
	for line := range tc.linesOfInterest {
		tc.showLines[line] = struct{}{}
	}

	// Add padding lines around each LOI
	if tc.loiPad > 0 {
		var toAdd []int
		for line := range tc.showLines {
			start := line - tc.loiPad
			end := line + tc.loiPad
			for nl := start; nl <= end; nl++ {
				if nl < 0 || nl >= tc.numLines {
					continue
				}
				toAdd = append(toAdd, nl)
			}
		}
		for _, x := range toAdd {
			tc.showLines[x] = struct{}{}
		}
	}

	// Optionally add bottom line (plus parent context)
	if tc.showLastLine {
		bottomLine := tc.numLines - 1
		tc.showLines[bottomLine] = struct{}{}
		tc.addParentScopes(bottomLine)
	}

	// Add parent contexts
	if tc.parentContext {
		for i := range tc.linesOfInterest {
			tc.addParentScopes(i)
		}
	}

	// Add child contexts
	// NOTE: This is where we fix partial expansions. If you want the entire function body,
	// you can remove or adjust the logic in addChildContext.
	if tc.showChildContext {
		for i := range tc.linesOfInterest {
			tc.addChildContext(i)
		}
	}

	// Add top margin lines
	if tc.margin > 0 {
		for i := 0; i < tc.margin && i < tc.numLines; i++ {
			tc.showLines[i] = struct{}{}
		}
	}

	// Close small gaps between lines to produce a smoother snippet
	tc.closeSmallGaps()
}

// addChildContext tries to show a child scope for the line i (e.g. function body)
// If the scope is small (<5 lines) everything is revealed.  Otherwise, partial expansions
// is added by calling addParentScopes(childStart) for each child, up to a max limit.
func (tc *TreeContext) addChildContext(i int) {
	if i < 0 || i >= len(tc.nodes) {
		return
	}
	if len(tc.nodes[i]) == 0 {
		return
	}

	lastLine := tc.getLastLineOfScope(i)
	size := lastLine - i
	if size < 0 {
		return
	}

	// If the scope is small enough, reveal everything.
	if size < 5 {
		for line := i; line <= lastLine && line < tc.numLines; line++ {
			tc.showLines[line] = struct{}{}
		}
		return
	}

	// Gather all children for node(s) on line i, then sort by size descending.
	children := []*sitter.Node{}
	for _, node := range tc.nodes[i] {
		children = append(children, tc.findAllChildren(node)...)
	}
	sortNodesBySize(children)

	currentlyShowing := len(tc.showLines)

	// We only reveal ~10% of the larger scope, at least 5 lines, at most 25 lines,
	// matching the Python logic.
	maxToShow := 25
	minToShow := 5
	percentToShow := 0.10
	computedMax := int(float64(size)*percentToShow + 0.5)
	if computedMax < minToShow {
		computedMax = minToShow
	} else if computedMax > maxToShow {
		computedMax = maxToShow
	}

	// For each child, we only expand up to computedMax times by revealing
	// its parent scopes.  (Mirrors Python's "self.add_parent_scopes(child_start_line)")
	for _, child := range children {
		if len(tc.showLines) > currentlyShowing+computedMax {
			break
		}
		childStart := int(child.StartPosition().Row)
		childEnd := int(child.EndPosition().Row)
		for line := childStart; line <= childEnd && line < tc.numLines; line++ {
			tc.showLines[line] = struct{}{}
		}
		tc.addParentScopes(childStart)
	}
}

// findAllChildren gathers all descendants (recursive)
func (tc *TreeContext) findAllChildren(node *sitter.Node) []*sitter.Node {
	out := []*sitter.Node{node}
	for i := uint(0); i < node.ChildCount(); i++ {
		if child := node.NamedChild(i); child != nil {
			out = append(out, tc.findAllChildren(child)...)
		}
	}
	return out
}

type scopeBoundry struct {
	start int
	end   int
}

// getScopeBoundry returns all the lines belonging to the scope that starts at line i.
func (tc *TreeContext) getScopeBoundry(i int) scopeBoundry {
	// Check if index is out of range or if no nodes exist at the given line.
	if i < 0 || i >= len(tc.nodes) || len(tc.nodes[i]) == 0 {
		return scopeBoundry{start: i, end: i} // No valid nodes, return the original line.
	}

	lastLine := i // Initialize last line to the starting line instead of 0.

	// Iterate over all nodes that start on line i. Work backwards to find the last line.
	for startLine := i; startLine >= 0; startLine-- {

		// Look at every node that started on that line
		for _, node := range tc.nodes[startLine] {
			s := int(node.StartPosition().Row)
			e := int(node.EndPosition().Row)

			// fmt.Printf("%d : %d-%d - %s\n", i, s, e, node.Utf8Text(tc.source))

			// Confirm i is actually inside this node's range
			if i >= s && i <= e {
				if e > lastLine {
					return scopeBoundry{start: s, end: e}
				}
			}
		}
	}

	return scopeBoundry{start: i, end: i}
}

// getLastLineOfScope finds the last line number of a code block starting at line i.
// It iterates over all syntax tree nodes on line i and determines the maximum end line.
func (tc *TreeContext) getLastLineOfScope(i int) int {
	// Check if index is out of range or if no nodes exist at the given line.
	if i < 0 || i >= len(tc.nodes) || len(tc.nodes[i]) == 0 {
		return i // No valid nodes, return the original line.
	}

	lastLine := i // Initialize last line to the starting line instead of 0.

	// Iterate over all nodes that start on line i. Work backwards to find the last line.
	for startLine := i; startLine >= 0; startLine-- {

		// Look at every node that started on that line
		for _, node := range tc.nodes[startLine] {
			s := int(node.StartPosition().Row)
			e := int(node.EndPosition().Row)

			// fmt.Printf("%d-%d - %s\n", s, e, node.Utf8Text(tc.source))

			// Confirm i is actually inside this node's range
			if i >= s && i <= e {
				if e > lastLine {
					return e
				}
			}
		}
	}

	return lastLine
}

// closeSmallGaps attempts to fill in small gaps in the displayed lines.
// It performs two key tasks:
// 1. Closes single-line gaps between visible lines (i.e., fills i+1 if i and i+2 are present).
// 2. Includes adjacent blank lines when the previous line is visible.
func (tc *TreeContext) closeSmallGaps() {
	// Create a new map to store the updated set of visible lines.
	// This starts as a copy of tc.showLines.
	closedShow := make(map[int]struct{}, len(tc.showLines))
	for k := range tc.showLines {
		closedShow[k] = struct{}{}
	}

	// Extract and sort the keys (line numbers) from the showLines map.
	// This allows sequential processing of the visible lines.
	sortedShow := mapKeysSorted(tc.showLines)

	last := len(sortedShow) - 1

	// Pass 1: Fill in single-line gaps.
	// If two consecutive visible lines have exactly one line between them (i.e., `i` and `i+2` exist),
	// mark the middle line (`i+1`) as visible.
	for i := 0; i < last; i++ {
		curr := sortedShow[i]
		next := sortedShow[i+1]

		// skip trailing empty lines
		if next == last && strings.TrimSpace(tc.lines[next]) == "" {
			continue
		}
		if next-curr == 2 {
			closedShow[curr+1] = struct{}{}
		}
	}

	// // Pass 2: Include blank lines that are adjacent to visible lines.
	// // If a visible line contains non-whitespace content and the next line is blank,
	// // mark the blank line as visible.
	// for i, line := range tc.lines {
	// 	if _, ok := closedShow[i]; ok {
	// 		// Ensure the current line is non-empty and within valid bounds.
	// 		if strings.TrimSpace(line) != "" && i < tc.numLines-1 {
	// 			// If the next line is blank, mark it as visible.
	// 			if len(tc.lines) > i+1 && strings.TrimSpace(tc.lines[i+1]) == "" {
	// 				closedShow[i+1] = struct{}{}
	// 			}
	// 		}
	// 	}
	// }

	// Update the showLines map with the modified set of visible lines.
	tc.showLines = closedShow
}

// Format outputs the final lines. This version prints an initial ellipsis
// if the first line is NOT in showLines, replicating the Python code's
// "dots = not (0 in self.show_lines)" behavior.
func (tc *TreeContext) Format() string {
	if len(tc.showLines) == 0 {
		return ""
	}

	var sb strings.Builder

	// Optional color reset at the start
	if tc.color {
		sb.WriteString("\033[0m\n")
	}

	// If the first line is *not* in showLines, we begin in "ellipses" mode,
	// so we will print an ellipsis when we next skip lines.
	_, firstLineShown := tc.showLines[0]
	printEllipsis := !firstLineShown

	for i, line := range tc.lines {
		_, shouldShow := tc.showLines[i]
		if !shouldShow {
			// Print ellipsis once after last shown line
			if printEllipsis {
				sb.WriteString("⋮...\n")
				printEllipsis = false
			}
			continue
		}

		// Show the line
		spacer := tc.lineOfInterestSpacer(i)
		oline := tc.highlightedOrOriginalLine(i, line)
		if tc.showLineNumber {
			fmt.Fprintf(&sb, "%3d%s%s\n", i+1, spacer, oline)
		} else {
			fmt.Fprintf(&sb, "%s%s\n", spacer, oline)
		}

		// If we skip lines after this, we want an ellipsis
		printEllipsis = true
	}

	return sb.String()
}

// lineOfInterestSpacer returns "│" or "█" (with color if needed)
func (tc *TreeContext) lineOfInterestSpacer(i int) string {
	if _, isLOI := tc.linesOfInterest[i]; isLOI && tc.markLOIs {
		if tc.color {
			return "\033[31m█\033[0m"
		}
		return "█"
	}
	return "│"
}

// highlightedOrOriginalLine uses the highlighted version if present
func (tc *TreeContext) highlightedOrOriginalLine(i int, original string) string {
	if hl, ok := tc.outputLines[i]; ok {
		return hl
	}
	return original
}

// addParentScopes recursively marks lines for parent scopes as visible.
// This ensures that when a line belongs to a scope, its parent context is also shown.
func (tc *TreeContext) addParentScopes(i int) {
	// Boundary check: Ensure i is within valid range
	if i < 0 || i >= len(tc.scopes) {
		return
	}

	// If this scope has already been processed, avoid redundant work.
	if _, done := tc.doneParentScopes[i]; done {
		return
	}
	tc.doneParentScopes[i] = struct{}{} // Mark this index as processed.

	// Iterate over all scope start line numbers at index i.
	for lineNum := range tc.scopes[i] {
		// Retrieve the scope header (expected to be a slice of at least two elements).
		headerSlice := tc.header[lineNum]

		if len(headerSlice) < 2 {
			// **Potential Issue:** headerSlice does not have at least 2 elements.
			// This may result in an **index out of range** error or unintended behavior.
			// If this happens, consider logging or handling the missing header case.
		}

		headStart := headerSlice[0] // First line of the header
		headEnd := headerSlice[1]   // Last line of the header

		// Skip if the header start is at the beginning of the file.
		if headStart == 0 {
			continue
		}

		// Show lines within the header if either:
		// - headStart is non-zero (ensuring it's part of a meaningful scope)
		// - tc.showTopOfFileParentScope is enabled (forcing top-of-file scopes to be shown)
		if headStart > 0 || tc.showTopOfFileParentScope {
			for ln := headStart; ln < headEnd && ln < tc.numLines; ln++ {
				tc.showLines[ln] = struct{}{} // Mark lines in the header as visible.
			}
		}

		// If the `showLastLine` flag is enabled, determine and include the last line of the scope.
		lines := tc.getScopeBoundry(lineNum) // Get last line of current scope.
		if lines.end == lineNum {
			continue // Skip if the last line is the same as the current line.
		}

		// Mark lines
		for ln := lines.start; ln <= lines.end && ln < tc.numLines; ln++ {
			tc.showLines[ln] = struct{}{}
		}
		// fmt.Printf("%d-%d : %v\n", lines.start, lines.end, tc.showLines)

	}
}

// walkTree populates scopes, headers, etc.
func (tc *TreeContext) walkTree(node *sitter.Node, depth int) (int, int) {
	startLine := int(node.StartPosition().Row)
	endLine := int(node.EndPosition().Row)
	size := endLine - startLine

	if startLine < 0 || startLine >= len(tc.nodes) {
		return startLine, endLine
	}
	tc.nodes[startLine] = append(tc.nodes[startLine], node)

	// if tc.verbose && node.IsNamed() {
	// 	textLine := strings.Split(node.Utf8Text(tc.source), "\n")[0]
	// 	var codeLine string
	// 	if startLine < len(tc.lines) {
	// 		codeLine = tc.lines[startLine]
	// 	}
	// 	fmt.Printf("%s %s %d-%d=%d %s %s\n",
	// 		strings.Repeat("   ", depth),
	// 		node.Kind(),
	// 		startLine,
	// 		endLine,
	// 		size+1,
	// 		textLine,
	// 		codeLine,
	// 	)
	// }

	if size > 0 {
		if startLine < len(tc.header) {
			// store [size, startLine, endLine]
			tc.header[startLine] = []int{size, startLine, endLine}
		}
	}

	// Mark each line in [startLine, endLine] as belonging to scope `startLine`
	for i := startLine; i <= endLine && i < len(tc.scopes); i++ {
		tc.scopes[i][startLine] = struct{}{}
	}

	for i := uint(0); i < node.ChildCount(); i++ {
		if child := node.NamedChild(i); child != nil {
			tc.walkTree(child, depth+1)
		}
	}

	return startLine, endLine
}

// --- Helper functions ---

// mapKeysSorted returns sorted keys of a map[int]struct{} as a slice.
func mapKeysSorted(m map[int]struct{}) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	// a trivial sort
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// sortBySize sorts a slice of items by their computed size in descending order.
func sortBySize[T any](items []T, getSize func(T) uint) {
	sort.Slice(items, func(i, j int) bool {
		return getSize(items[j]) > getSize(items[i])
	})
}

// sortNodesBySize sorts nodes by (EndLine - StartLine) in descending order.
func sortNodesBySize(nodes []*sitter.Node) {
	sortBySize(nodes, func(n *sitter.Node) uint {
		return n.StartPosition().Row - n.EndPosition().Row
	})
}
