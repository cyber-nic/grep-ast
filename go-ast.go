package goast

import (
	"fmt"
	"regexp"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// TreeContext stores context about source code lines, parsing, scopes, and line-of-interest management.
type TreeContext struct {
	filename                 string
	source                   []byte
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
	header                   [][]int          // each element: [startLine, endLine]
	nodes                    [][]*sitter.Node // tracks parse-tree nodes by start line
	showLines                map[int]struct{}
	linesOfInterest          map[int]struct{}
	doneParentScopes         map[int]struct{}
}

type TreeContextOptions struct {
	Color                    bool
	Verbose                  bool
	ShowLineNumber           bool
	ShowParentContext        bool
	ShowChildContext         bool
	ShowLastLine             bool
	MarginPadding            int
	MarkLinesOfInterest      bool
	HeaderMax                int
	ShowTopOfFileParentScope bool
	LinesOfInterestPadding   int
}

// NewTreeContext is the Go-equivalent constructor for TreeContext.
func NewTreeContext(filename string, source []byte, options TreeContextOptions) (*TreeContext, error) {
	// Get the language from the filename
	lang, err := getLanguageFromFileName(filename)
	if err != nil {
		return nil, err
	}

	// Return an error if the language is not supported
	if lang == nil {
		return nil, fmt.Errorf("unrecognized or unsupported file type (%s)", filename)
	}

	// Initialize Tree-sitter parser
	parser := sitter.NewParser()
	parser.SetLanguage(lang) // Change to appropriate language parser

	// Parse the source code
	tree := parser.Parse(source, nil)

	// Get the root node
	rootNode := tree.RootNode()

	// Split lines
	lines := strings.Split(string(source), "\n")
	numLines := len(lines)
	if len(source) > 0 && source[len(source)-1] == '\n' {
		// Adjust if there's a trailing newline, so the python version's "len + 1" logic is matched
		numLines += 0
	}

	// Initialize scopes and headers
	scopes := make([]map[int]struct{}, numLines+1) // +1 to mimic python’s len+1 usage
	header := make([][]int, numLines+1)
	nodes := make([][]*sitter.Node, numLines+1)
	for i := 0; i <= numLines; i++ {
		scopes[i] = make(map[int]struct{})
		header[i] = []int{0, 0}
		nodes[i] = []*sitter.Node{}
	}

	tc := &TreeContext{
		filename:                 filename,
		source:                   source,
		color:                    options.Color,
		verbose:                  options.Verbose,
		lineNumber:               options.ShowLineNumber,
		parentContext:            options.ShowParentContext,
		childContext:             options.ShowChildContext,
		lastLine:                 options.ShowLastLine,
		margin:                   options.MarginPadding,
		markLOIs:                 options.MarkLinesOfInterest,
		headerMax:                options.HeaderMax,
		loiPad:                   options.LinesOfInterestPadding,
		showTopOfFileParentScope: options.ShowTopOfFileParentScope,
		lines:                    lines,
		numLines:                 numLines + 1,
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

// postWalkProcessing sets header ranges and optionally prints scopes.
func (tc *TreeContext) postWalkProcessing() {
	// print and set header ranges
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
	children := []*sitter.Node{}
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
		childStart := child.StartPosition().Row
		tc.addParentScopes(int(childStart))
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

// getLastLineOfScope finds the maximum end_line for nodes that start on line i
func (tc *TreeContext) getLastLineOfScope(i int) int {
	if i < 0 || i >= len(tc.nodes) || len(tc.nodes[i]) == 0 {
		return i
	}
	lastLine := 0
	for _, node := range tc.nodes[i] {
		if int(node.EndPosition().Row) > lastLine {
			lastLine = int(node.EndPosition().Row)
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

// Format outputs the final lines
func (tc *TreeContext) Format() string {
	// If there’s nothing to show, bail early
	if len(tc.showLines) == 0 {
		return ""
	}

	var sb strings.Builder

	// Optional color reset at the start
	if tc.color {
		sb.WriteString("\033[0m\n")
	}

	// `printEllipsis` tracks whether we've shown a line since
	// the last time we printed an ellipsis. Initially false.
	var printEllipsis bool

	for i, line := range tc.lines {
		if _, shouldShow := tc.showLines[i]; !shouldShow {
			// Print ellipsis only once after the last shown line
			if printEllipsis {
				sb.WriteString(tc.ellipsisLine())
				printEllipsis = false
			}
			continue
		}

		// Show the line
		spacer := tc.lineOfInterestSpacer(i)
		oline := tc.highlightedOrOriginalLine(i, line)
		if tc.lineNumber {
			fmt.Fprintf(&sb, "%3d%s%s\n", i+1, spacer, oline)
		} else {
			fmt.Fprintf(&sb, "%s%s\n", spacer, oline)
		}

		// Next time we skip lines, we may print an ellipsis
		printEllipsis = true
	}

	return sb.String()
}

// ellipsisLine returns "...⋮...\n" if lineNumber is on, otherwise "⋮...\n"
func (tc *TreeContext) ellipsisLine() string {
	if tc.lineNumber {
		return "...⋮...\n"
	}
	return "⋮...\n"
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

// sortNodesBySize sorts nodes by (EndLine-StartLine) descending.
func sortNodesBySize(nodes []*sitter.Node) {
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			sizeI := int(nodes[i].EndPosition().Row) - int(nodes[i].StartPosition().Row)
			sizeJ := int(nodes[j].EndPosition().Row) - int(nodes[j].StartPosition().Row)
			if sizeJ > sizeI {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}
