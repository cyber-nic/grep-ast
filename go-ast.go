package goast

import (
	"fmt"
	"regexp"
	"strings"
)

// TreeContext mimics the Python class that stores context
// about source code lines, parsing, scopes, and line-of-interest management.
type TreeContext struct {
	filename                 string
	color                    bool
	verbose                  bool
	lineNumber               bool
	lastLine                 bool
	margin                   int
	markLOIs                 bool
	headerMax                int
	loiPad                   int
	showTopOfFileParentScope bool
	parentContext            bool
	childContext             bool
	lines                    []string
	numLines                 int
	outputLines              map[int]string
	scopes                   []map[int]struct{}
	header                   [][]int   // each element: [startLine, endLine]
	nodes                    [][]*Node // tracks parse-tree nodes by start line
	showLines                map[int]struct{}
	linesOfInterest          map[int]struct{}
	doneParentScopes         map[int]struct{}
}

// Node is a simplified stand-in for tree-sitter nodes in Go.
// In a real environment, you would replace this with the actual tree-sitter struct or wrapper.
type Node struct {
	Type       string
	Text       string
	StartPoint [2]int // [line, column]
	EndPoint   [2]int // [line, column]
	Children   []*Node
	IsNamed    bool
}

// filenameToLang is a placeholder for your logic of mapping file names to a language
func filenameToLang(filename string) (string, error) {
	// For example, return "go" for .go files, "python" for .py, etc.
	// Return an error if you can’t determine the language.
	if strings.HasSuffix(filename, ".py") {
		return "python", nil
	} else if strings.HasSuffix(filename, ".go") {
		return "go", nil
	}
	return "", fmt.Errorf("Unknown language for %s", filename)
}

// getParser is a placeholder for returning a parser based on language.
// Replace with real usage of a Go tree-sitter or relevant parser.
func getParser(lang string) (*Node, error) {
	// Here, we just create a dummy root node.
	// In reality, you’d parse the code with tree-sitter or a suitable parser.
	root := &Node{
		Type:       "root",
		Text:       "root node text",
		StartPoint: [2]int{0, 0},
		EndPoint:   [2]int{0, 0},
		IsNamed:    true,
	}
	return root, nil
}

// NewTreeContext is the Go-equivalent constructor for TreeContext.
func NewTreeContext(
	filename string,
	code string,
	color bool,
	verbose bool,
	lineNumber bool,
	parentContext bool,
	childContext bool,
	lastLine bool,
	margin int,
	markLOIs bool,
	headerMax int,
	showTopOfFileParentScope bool,
	loiPad int,
) (*TreeContext, error) {

	// Map filename -> language
	lang, err := filenameToLang(filename)
	if err != nil {
		return nil, err
	}

	// In Python, `get_parser(lang)` returns a specialized parser. We replicate that here:
	rootNode, err := getParser(lang)
	if err != nil {
		return nil, err
	}

	// Split lines
	lines := strings.Split(code, "\n")
	numLines := len(lines)
	if len(code) > 0 && code[len(code)-1] == '\n' {
		// Adjust if there's a trailing newline, so the python version's "len + 1" logic is matched
		numLines += 0
	}

	// Initialize scopes and headers
	scopes := make([]map[int]struct{}, numLines+1) // +1 to mimic python’s len+1 usage
	header := make([][]int, numLines+1)
	nodes := make([][]*Node, numLines+1)
	for i := 0; i <= numLines; i++ {
		scopes[i] = make(map[int]struct{})
		header[i] = []int{0, 0}
		nodes[i] = []*Node{}
	}

	tc := &TreeContext{
		filename:                 filename,
		color:                    color,
		verbose:                  verbose,
		lineNumber:               lineNumber,
		parentContext:            parentContext,
		childContext:             childContext,
		lastLine:                 lastLine,
		margin:                   margin,
		markLOIs:                 markLOIs,
		headerMax:                headerMax,
		loiPad:                   loiPad,
		showTopOfFileParentScope: showTopOfFileParentScope,
		lines:                    lines,
		numLines:                 numLines + 1, // python was len(lines) + 1
		outputLines:              make(map[int]string),
		scopes:                   scopes,
		header:                   header,
		nodes:                    nodes,
		showLines:                make(map[int]struct{}),
		linesOfInterest:          make(map[int]struct{}),
		doneParentScopes:         make(map[int]struct{}),
	}

	// Walk parse tree
	tc.walkTree(rootNode, 0)

	// Post-process scopes / headers
	tc.postWalkProcessing()

	return tc, nil
}

// postWalkProcessing tries to replicate the logic in the Python constructor after tree walk.
func (tc *TreeContext) postWalkProcessing() {
	// We replicate the Python verbose printing and setting header ranges
	var scopeWidth int
	if tc.verbose {
		// find the maximum width for printing scopes
		for i := 0; i < tc.numLines-1; i++ {
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

		if tc.verbose && i < tc.numLines-1 {
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

	// ensure we have linesOfInterest in showLines
	for line := range tc.linesOfInterest {
		tc.showLines[line] = struct{}{}
	}

	// add padding lines around each LOI
	if tc.loiPad > 0 {
		toAdd := []int{}
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

	// optionally add bottom line (plus parent context)
	if tc.lastLine {
		bottomLine := tc.numLines - 2
		tc.showLines[bottomLine] = struct{}{}
		tc.addParentScopes(bottomLine)
	}

	// add parent contexts
	if tc.parentContext {
		for i := range tc.linesOfInterest {
			tc.addParentScopes(i)
		}
	}

	// add child contexts
	if tc.childContext {
		for i := range tc.linesOfInterest {
			tc.addChildContext(i)
		}
	}

	// add top margin lines
	if tc.margin > 0 {
		for i := 0; i < tc.margin && i < tc.numLines; i++ {
			tc.showLines[i] = struct{}{}
		}
	}

	// close small gaps
	tc.closeSmallGaps()
}

// addChildContext tries to show a child scope if it’s small or show partial scopes if large
func (tc *TreeContext) addChildContext(i int) {
	if i < 0 || i >= len(tc.nodes) {
		return
	}
	if len(tc.nodes[i]) == 0 {
		return
	}

	lastLine := tc.getLastLineOfScope(i)
	size := lastLine - i
	if size < 5 {
		for line := i; line <= lastLine; line++ {
			tc.showLines[line] = struct{}{}
		}
		return
	}

	// gather all children
	children := []*Node{}
	for _, node := range tc.nodes[i] {
		children = append(children, tc.findAllChildren(node)...)
	}

	// sort them by (end_line-start_line) descending
	sortNodesBySize(children)

	currentlyShowing := len(tc.showLines)
	maxToShow := 25
	minToShow := 5
	percentToShow := 0.10
	computedMax := int(float64(size)*percentToShow + 0.5)
	if computedMax < minToShow {
		computedMax = minToShow
	} else if computedMax > maxToShow {
		computedMax = maxToShow
	}

	for _, child := range children {
		if len(tc.showLines) > currentlyShowing+computedMax {
			break
		}
		childStart := child.StartPoint[0]
		tc.addParentScopes(childStart)
	}
}

// findAllChildren gathers all descendants (recursive)
func (tc *TreeContext) findAllChildren(node *Node) []*Node {
	out := []*Node{node}
	for _, child := range node.Children {
		out = append(out, tc.findAllChildren(child)...)
	}
	return out
}

// getLastLineOfScope finds the maximum end_line for nodes that start on line i
func (tc *TreeContext) getLastLineOfScope(i int) int {
	if i < 0 || i >= len(tc.nodes) || len(tc.nodes[i]) == 0 {
		return i
	}
	lastLine := 0
	for _, node := range tc.nodes[i] {
		if node.EndPoint[0] > lastLine {
			lastLine = node.EndPoint[0]
		}
	}
	return lastLine
}

// closeSmallGaps closes single-line gaps.
func (tc *TreeContext) closeSmallGaps() {
	closedShow := make(map[int]struct{}, len(tc.showLines))
	for k := range tc.showLines {
		closedShow[k] = struct{}{}
	}

	sortedShow := mapKeysSorted(tc.showLines)

	// fill i+1 if i and i+2 are present
	for i := 0; i < len(sortedShow)-1; i++ {
		curr := sortedShow[i]
		next := sortedShow[i+1]
		if next-curr == 2 {
			closedShow[curr+1] = struct{}{}
		}
	}

	// pick up adjacent blank lines
	for i, line := range tc.lines {
		if _, ok := closedShow[i]; ok {
			if strings.TrimSpace(line) != "" && i < tc.numLines-2 {
				// check if next line is blank
				if len(tc.lines) > i+1 && strings.TrimSpace(tc.lines[i+1]) == "" {
					closedShow[i+1] = struct{}{}
				}
			}
		}
	}

	tc.showLines = closedShow
}

// Format outputs the final lines, replicating the logic in the Python method.
func (tc *TreeContext) Format() string {
	if len(tc.showLines) == 0 {
		return ""
	}

	var sb strings.Builder

	// Optional reset color
	if tc.color {
		sb.WriteString("\033[0m\n")
	}

	dots := false
	if _, present := tc.showLines[0]; present {
		dots = true
	}

	for i, line := range tc.lines {
		_, shouldShow := tc.showLines[i]
		if !shouldShow {
			if dots {
				// We print the "⋮..." line once as an omitted block, like Python does
				if tc.lineNumber {
					sb.WriteString("...⋮...\n")
				} else {
					sb.WriteString("⋮...\n")
				}
				dots = false
			}
			continue
		}

		// line-of-interest marker
		spacer := "│"
		if _, isLOI := tc.linesOfInterest[i]; isLOI && tc.markLOIs {
			spacer = "█"
			if tc.color {
				spacer = "\033[31m█\033[0m"
			}
		}

		// use outputLines if we have highlighted version
		oline, ok := tc.outputLines[i]
		if !ok {
			oline = line
		}

		if tc.lineNumber {
			sb.WriteString(fmt.Sprintf("%3d%s%s\n", i+1, spacer, oline))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s\n", spacer, oline))
		}
		dots = true
	}

	return sb.String()
}

// addParentScopes recursively shows lines for parent scopes
func (tc *TreeContext) addParentScopes(i int) {
	if i < 0 || i >= len(tc.scopes) {
		return
	}
	if _, done := tc.doneParentScopes[i]; done {
		return
	}
	tc.doneParentScopes[i] = struct{}{}

	// for each scope that starts at line_num
	for lineNum := range tc.scopes[i] {
		headerSlice := tc.header[lineNum]
		if len(headerSlice) >= 2 {
			headStart := headerSlice[0]
			headEnd := headerSlice[1]
			if headStart > 0 || tc.showTopOfFileParentScope {
				for ln := headStart; ln < headEnd && ln < tc.numLines; ln++ {
					tc.showLines[ln] = struct{}{}
				}
			}
			// optionally add last line
			if tc.lastLine {
				lastLine := tc.getLastLineOfScope(lineNum)
				tc.addParentScopes(lastLine)
			}
		}
	}
}

// walkTree populates scopes, headers, etc.
func (tc *TreeContext) walkTree(node *Node, depth int) (int, int) {
	startLine := node.StartPoint[0]
	endLine := node.EndPoint[0]
	size := endLine - startLine

	if startLine < 0 || startLine >= len(tc.nodes) {
		return startLine, endLine
	}
	tc.nodes[startLine] = append(tc.nodes[startLine], node)

	if tc.verbose && node.IsNamed {
		// mimic the Python debugging prints
		textLine := strings.Split(node.Text, "\n")[0]
		var codeLine string
		if startLine < len(tc.lines) {
			codeLine = tc.lines[startLine]
		}
		fmt.Printf("%s %s %d-%d=%d %s %s\n",
			strings.Repeat("   ", depth),
			node.Type,
			startLine,
			endLine,
			size+1,
			textLine,
			codeLine,
		)
	}

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

	for _, child := range node.Children {
		tc.walkTree(child, depth+1)
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

// sortNodesBySize sorts nodes by (EndLine-StartLine) descending.
func sortNodesBySize(nodes []*Node) {
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			sizeI := nodes[i].EndPoint[0] - nodes[i].StartPoint[0]
			sizeJ := nodes[j].EndPoint[0] - nodes[j].StartPoint[0]
			if sizeJ > sizeI {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}
