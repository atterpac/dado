package components

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// DiffLineType indicates the type of diff line
type DiffLineType int

const (
	DiffLineContext DiffLineType = iota // Unchanged line
	DiffLineAdded                       // Added line (+)
	DiffLineRemoved                     // Removed line (-)
	DiffLineHeader                      // Hunk header (@@ ... @@)
	DiffLineFile                        // File header (diff --git, ---, +++)
)

// DiffLine represents a single line in the diff
type DiffLine struct {
	Type      DiffLineType
	OldLineNo int    // Line number in old content (0 if added)
	NewLineNo int    // Line number in new content (0 if removed)
	Content   string // Line content without +/- prefix
	Selected  bool   // Whether line is selected (for staging)
	HunkIndex int    // Index of hunk this line belongs to
}

// DiffHunk represents a contiguous block of changes
type DiffHunk struct {
	Header   string     // @@ -start,count +start,count @@ context
	OldStart int        // Starting line in old file
	OldCount int        // Number of lines from old file
	NewStart int        // Starting line in new file
	NewCount int        // Number of lines in new file
	Lines    []DiffLine // Lines in this hunk
}

// DiffResult is the parsed output of a unified diff. Pass it to DiffViewer.SetResult
// or let DiffViewer.SetDiff parse the raw diff string for you.
type DiffResult struct {
	OldName string     // Original file name
	NewName string     // New file name
	Hunks   []DiffHunk // All hunks
	Binary  bool       // True if binary file
}

// DiffStats is a summary count of additions, deletions, and affected files
// across an entire diff. Returned by DiffViewer.GetStats.
type DiffStats struct {
	Additions int
	Deletions int
	Files     int
}

// DiffViewer displays diff content with syntax highlighting
type DiffViewer struct {
	widgetBase

	result        *DiffResult
	lines         []DiffLine // Flattened lines for display
	selectedIndex int
	offset        int

	// Display options
	sideBySide       bool
	showLineNumbers  bool
	wordDiff         bool
	contextLines     int
	title            string
	selectionEnabled bool // Enable line selection mode (for staging)
	showSelection    bool // Show selection indicators

	// Callbacks
	onLineSelect  func(line DiffLine)
	onHunkAction  func(hunkIndex int, lines []DiffLine) // Called for hunk-level operations
	onLinesAction func(lines []DiffLine)                // Called for selected lines operations
}

// NewDiffViewer creates a new diff viewer component
func NewDiffViewer() *DiffViewer {
	d := &DiffViewer{
		showLineNumbers: true,
		contextLines:    3,
	}
	d.initWidget()
	return d
}

// SetDiff computes and displays diff between old and new content
func (d *DiffViewer) SetDiff(old, new string) *DiffViewer {
	result := computeDiff(old, new)
	return d.SetDiffResult(result)
}

// SetDiffResult sets a pre-computed diff result
func (d *DiffViewer) SetDiffResult(result *DiffResult) *DiffViewer {
	d.result = result
	d.flattenLines()
	d.selectedIndex = 0
	if len(d.lines) > 0 {
		d.selectedIndex = d.firstChangeInHunk(d.lines[0].HunkIndex, 0)
	}
	d.offset = 0
	return d
}

// SetUnifiedDiff parses and displays a unified diff string (e.g., from git)
func (d *DiffViewer) SetUnifiedDiff(diff string) *DiffViewer {
	result := parseUnifiedDiff(diff)
	return d.SetDiffResult(result)
}

// SetSideBySide enables/disables side-by-side mode
func (d *DiffViewer) SetSideBySide(enabled bool) *DiffViewer {
	d.sideBySide = enabled
	return d
}

// SetShowLineNumbers toggles line number display
func (d *DiffViewer) SetShowLineNumbers(show bool) *DiffViewer {
	d.showLineNumbers = show
	return d
}

// SetWordDiff enables word-level diff highlighting
func (d *DiffViewer) SetWordDiff(enabled bool) *DiffViewer {
	d.wordDiff = enabled
	return d
}

// SetContextLines sets context lines around changes
func (d *DiffViewer) SetContextLines(count int) *DiffViewer {
	d.contextLines = count
	return d
}

// SetTitle sets the header title
func (d *DiffViewer) SetTitle(title string) *DiffViewer {
	d.title = title
	return d
}

// SetOnLineSelect sets callback for line selection
func (d *DiffViewer) SetOnLineSelect(fn func(line DiffLine)) *DiffViewer {
	d.onLineSelect = fn
	return d
}

// SetSelectionEnabled enables/disables line selection mode (for staging)
func (d *DiffViewer) SetSelectionEnabled(enabled bool) *DiffViewer {
	d.selectionEnabled = enabled
	d.showSelection = enabled
	return d
}

// SetOnHunkAction sets callback for hunk-level operations
func (d *DiffViewer) SetOnHunkAction(fn func(hunkIndex int, lines []DiffLine)) *DiffViewer {
	d.onHunkAction = fn
	return d
}

// SetOnLinesAction sets callback for selected lines operations
func (d *DiffViewer) SetOnLinesAction(fn func(lines []DiffLine)) *DiffViewer {
	d.onLinesAction = fn
	return d
}

// ToggleLineSelection toggles selection of the current line
func (d *DiffViewer) ToggleLineSelection() {
	if !d.selectionEnabled || d.selectedIndex >= len(d.lines) {
		return
	}
	line := &d.lines[d.selectedIndex]
	if line.Type == DiffLineAdded || line.Type == DiffLineRemoved {
		line.Selected = !line.Selected
	}
}

// SelectAllInHunk selects all changed lines in the current hunk
func (d *DiffViewer) SelectAllInHunk() {
	if !d.selectionEnabled || d.selectedIndex >= len(d.lines) {
		return
	}
	currentHunk := d.lines[d.selectedIndex].HunkIndex
	for i := range d.lines {
		if d.lines[i].HunkIndex == currentHunk {
			if d.lines[i].Type == DiffLineAdded || d.lines[i].Type == DiffLineRemoved {
				d.lines[i].Selected = true
			}
		}
	}
}

// ClearSelection clears all line selections
func (d *DiffViewer) ClearSelection() {
	for i := range d.lines {
		d.lines[i].Selected = false
	}
}

// GetSelectedLines returns all selected lines
func (d *DiffViewer) GetSelectedLines() []DiffLine {
	var selected []DiffLine
	for _, line := range d.lines {
		if line.Selected {
			selected = append(selected, line)
		}
	}
	return selected
}

// GetCurrentHunkIndex returns the hunk index of the current line
func (d *DiffViewer) GetCurrentHunkIndex() int {
	if d.selectedIndex >= 0 && d.selectedIndex < len(d.lines) {
		return d.lines[d.selectedIndex].HunkIndex
	}
	return -1
}

// GetHunkLines returns all lines in the specified hunk
func (d *DiffViewer) GetHunkLines(hunkIndex int) []DiffLine {
	var lines []DiffLine
	for _, line := range d.lines {
		if line.HunkIndex == hunkIndex {
			lines = append(lines, line)
		}
	}
	return lines
}

// firstChangeInHunk returns the index of the first added/removed line in the
// given hunk. Falls back to the supplied index (typically the hunk header) when
// the hunk has no changed lines.
func (d *DiffViewer) firstChangeInHunk(hunkIndex, fallback int) int {
	for i := range d.lines {
		if d.lines[i].HunkIndex != hunkIndex {
			continue
		}
		if d.lines[i].Type == DiffLineAdded || d.lines[i].Type == DiffLineRemoved {
			return i
		}
	}
	return fallback
}

// NextHunk moves to the first changed line of the next hunk
func (d *DiffViewer) NextHunk() {
	currentHunk := d.GetCurrentHunkIndex()
	for i := d.selectedIndex + 1; i < len(d.lines); i++ {
		if d.lines[i].HunkIndex > currentHunk {
			d.selectedIndex = d.firstChangeInHunk(d.lines[i].HunkIndex, i)
			d.ensureVisible()
			return
		}
	}
}

// PrevHunk moves to the previous hunk
func (d *DiffViewer) PrevHunk() {
	currentHunk := d.GetCurrentHunkIndex()
	// Find start of current hunk first
	hunkStart := d.selectedIndex
	for hunkStart > 0 && d.lines[hunkStart-1].HunkIndex == currentHunk {
		hunkStart--
	}
	// Now find previous hunk
	for i := hunkStart - 1; i >= 0; i-- {
		if d.lines[i].HunkIndex < currentHunk {
			// Find start of this hunk
			for i > 0 && d.lines[i-1].HunkIndex == d.lines[i].HunkIndex {
				i--
			}
			d.selectedIndex = d.firstChangeInHunk(d.lines[i].HunkIndex, i)
			d.ensureVisible()
			return
		}
	}
}

// GetHunks returns all hunks from the result
func (d *DiffViewer) GetHunks() []DiffHunk {
	if d.result == nil {
		return nil
	}
	return d.result.Hunks
}

// GetSelectedLine returns the currently selected line
func (d *DiffViewer) GetSelectedLine() *DiffLine {
	if d.selectedIndex >= 0 && d.selectedIndex < len(d.lines) {
		return &d.lines[d.selectedIndex]
	}
	return nil
}

// GetStats returns diff statistics
func (d *DiffViewer) GetStats() DiffStats {
	stats := DiffStats{}
	if d.result != nil {
		stats.Files = 1
	}
	for _, line := range d.lines {
		switch line.Type {
		case DiffLineAdded:
			stats.Additions++
		case DiffLineRemoved:
			stats.Deletions++
		}
	}
	return stats
}

// NextChange jumps to the next changed line
func (d *DiffViewer) NextChange() {
	for i := d.selectedIndex + 1; i < len(d.lines); i++ {
		if d.lines[i].Type == DiffLineAdded || d.lines[i].Type == DiffLineRemoved {
			d.selectedIndex = i
			d.ensureVisible()
			return
		}
	}
}

// PrevChange jumps to the previous changed line
func (d *DiffViewer) PrevChange() {
	for i := d.selectedIndex - 1; i >= 0; i-- {
		if d.lines[i].Type == DiffLineAdded || d.lines[i].Type == DiffLineRemoved {
			d.selectedIndex = i
			d.ensureVisible()
			return
		}
	}
}

func (d *DiffViewer) flattenLines() {
	d.lines = nil
	if d.result == nil {
		return
	}

	for hunkIdx, hunk := range d.result.Hunks {
		// Add hunk header
		d.lines = append(d.lines, DiffLine{
			Type:      DiffLineHeader,
			Content:   hunk.Header,
			HunkIndex: hunkIdx,
		})
		// Add hunk lines with hunk index
		for _, line := range hunk.Lines {
			line.HunkIndex = hunkIdx
			d.lines = append(d.lines, line)
		}
	}
}

func (d *DiffViewer) ensureVisible() {
	_, _, _, height := d.GetInnerRect()
	if height <= 0 {
		return
	}

	if d.selectedIndex < d.offset {
		d.offset = d.selectedIndex
	}
	if d.selectedIndex >= d.offset+height {
		d.offset = d.selectedIndex - height + 1
	}
}

// Draw renders the diff viewer
func (d *DiffViewer) Draw(screen tcell.Screen) {
	d.Box.DrawForSubclass(screen)
	x, y, width, height := d.GetInnerRect()

	if width <= 0 || height <= 0 || len(d.lines) == 0 {
		return
	}

	// Colors (read at draw time per dado rules)
	th := d.th()
	bgColor := th.Bg()
	fgColor := th.Fg()
	fgDimColor := th.FgDim()
	accentColor := th.Accent()
	addedColor := th.Success()
	removedColor := th.Error()
	headerColor := th.Info()
	warningColor := th.Warning()

	// Ensure selection is visible
	d.ensureVisible()

	// Calculate line number width
	lineNumWidth := 0
	if d.showLineNumbers {
		lineNumWidth = 10 // "1234 1234 │"
	}

	// Draw lines
	for i := 0; i < height && d.offset+i < len(d.lines); i++ {
		line := d.lines[d.offset+i]
		rowY := y + i
		isSelected := d.offset+i == d.selectedIndex

		// Determine line style based on type
		var lineStyle tcell.Style
		var prefix rune
		var lineColor tcell.Color

		switch line.Type {
		case DiffLineAdded:
			lineColor = addedColor
			prefix = '+'
		case DiffLineRemoved:
			lineColor = removedColor
			prefix = '-'
		case DiffLineHeader:
			lineColor = headerColor
			prefix = '@'
		case DiffLineFile:
			lineColor = accentColor
			prefix = ' '
		default:
			lineColor = fgColor
			prefix = ' '
		}

		lineStyle = tcell.StyleDefault.Background(bgColor).Foreground(lineColor)
		if isSelected {
			lineStyle = tcell.StyleDefault.Background(accentColor).Foreground(bgColor)
		}

		// Clear the row
		fillLine(screen, x, rowY, width, lineStyle)

		col := x

		// Draw selection indicator (for staging mode)
		if d.showSelection {
			indicator := "  "
			indStyle := lineStyle
			if line.Selected {
				indicator = "> "
				if !isSelected {
					indStyle = tcell.StyleDefault.Background(bgColor).Foreground(warningColor)
				}
			}
			for _, ch := range indicator {
				if col < x+width {
					screen.SetContent(col, rowY, ch, nil, indStyle)
					col++
				}
			}
		}

		// Draw line numbers
		if d.showLineNumbers && line.Type != DiffLineHeader && line.Type != DiffLineFile {
			numStyle := lineStyle
			if !isSelected {
				numStyle = tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
			}

			oldNum := "    "
			newNum := "    "
			if line.OldLineNo > 0 {
				oldNum = fmt.Sprintf("%4d", line.OldLineNo)
			}
			if line.NewLineNo > 0 {
				newNum = fmt.Sprintf("%4d", line.NewLineNo)
			}

			numStr := oldNum + " " + newNum + " │"
			for _, ch := range numStr {
				if col < x+width {
					screen.SetContent(col, rowY, ch, nil, numStyle)
					col++
				}
			}
		} else if d.showLineNumbers {
			// Header/file lines - just show spaces for alignment
			for j := 0; j < lineNumWidth && col < x+width; j++ {
				screen.SetContent(col, rowY, ' ', nil, lineStyle)
				col++
			}
		}

		// Draw prefix character
		if col < x+width {
			screen.SetContent(col, rowY, prefix, nil, lineStyle)
			col++
		}

		// Draw content
		for _, ch := range line.Content {
			if col < x+width {
				screen.SetContent(col, rowY, ch, nil, lineStyle)
				col++
			}
		}
	}
}

func (d *DiffViewer) moveDown() {
	if d.selectedIndex < len(d.lines)-1 {
		d.selectedIndex++
		d.ensureVisible()
	}
}

func (d *DiffViewer) moveUp() {
	if d.selectedIndex > 0 {
		d.selectedIndex--
		d.ensureVisible()
	}
}

// MoveDown moves the selection down by one line.
func (d *DiffViewer) MoveDown() { d.moveDown() }

// MoveUp moves the selection up by one line.
func (d *DiffViewer) MoveUp() { d.moveUp() }

// HasFocus returns whether the component has focus
func (d *DiffViewer) HasFocus() bool {
	return d.Box.HasFocus()
}

// parseUnifiedDiff parses a unified diff string into a DiffResult
func parseUnifiedDiff(diff string) *DiffResult {
	result := &DiffResult{}
	lines := strings.Split(diff, "\n")

	var currentHunk *DiffHunk
	hunkHeaderRegex := regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@(.*)$`)

	oldLineNo := 0
	newLineNo := 0

	for _, line := range lines {
		// Check for file headers
		if strings.HasPrefix(line, "diff --git") {
			continue
		}
		if strings.HasPrefix(line, "---") {
			if len(line) > 4 {
				result.OldName = strings.TrimPrefix(line[4:], "a/")
			}
			continue
		}
		if strings.HasPrefix(line, "+++") {
			if len(line) > 4 {
				result.NewName = strings.TrimPrefix(line[4:], "b/")
			}
			continue
		}

		// Check for binary file
		if strings.HasPrefix(line, "Binary files") {
			result.Binary = true
			continue
		}

		// Check for hunk header
		if matches := hunkHeaderRegex.FindStringSubmatch(line); matches != nil {
			// Save previous hunk
			if currentHunk != nil {
				result.Hunks = append(result.Hunks, *currentHunk)
			}

			oldStart, _ := strconv.Atoi(matches[1])
			oldCount := 1
			if matches[2] != "" {
				oldCount, _ = strconv.Atoi(matches[2])
			}
			newStart, _ := strconv.Atoi(matches[3])
			newCount := 1
			if matches[4] != "" {
				newCount, _ = strconv.Atoi(matches[4])
			}

			currentHunk = &DiffHunk{
				Header:   line,
				OldStart: oldStart,
				OldCount: oldCount,
				NewStart: newStart,
				NewCount: newCount,
			}

			oldLineNo = oldStart
			newLineNo = newStart
			continue
		}

		// Parse diff lines
		if currentHunk != nil && len(line) > 0 {
			diffLine := DiffLine{}

			switch line[0] {
			case '+':
				diffLine.Type = DiffLineAdded
				diffLine.Content = line[1:]
				diffLine.NewLineNo = newLineNo
				newLineNo++
			case '-':
				diffLine.Type = DiffLineRemoved
				diffLine.Content = line[1:]
				diffLine.OldLineNo = oldLineNo
				oldLineNo++
			case ' ':
				diffLine.Type = DiffLineContext
				diffLine.Content = line[1:]
				diffLine.OldLineNo = oldLineNo
				diffLine.NewLineNo = newLineNo
				oldLineNo++
				newLineNo++
			case '\\':
				// "\ No newline at end of file" - skip
				continue
			default:
				// Context line without prefix (some diff formats)
				diffLine.Type = DiffLineContext
				diffLine.Content = line
				diffLine.OldLineNo = oldLineNo
				diffLine.NewLineNo = newLineNo
				oldLineNo++
				newLineNo++
			}

			currentHunk.Lines = append(currentHunk.Lines, diffLine)
		}
	}

	// Save last hunk
	if currentHunk != nil {
		result.Hunks = append(result.Hunks, *currentHunk)
	}

	return result
}

// computeDiff generates a diff between two strings
// This is a simple line-by-line diff using LCS algorithm
func computeDiff(old, new string) *DiffResult {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	// Simple LCS-based diff
	lcs := longestCommonSubsequence(oldLines, newLines)

	result := &DiffResult{
		OldName: "a",
		NewName: "b",
	}

	hunk := DiffHunk{
		Header:   fmt.Sprintf("@@ -1,%d +1,%d @@", len(oldLines), len(newLines)),
		OldStart: 1,
		OldCount: len(oldLines),
		NewStart: 1,
		NewCount: len(newLines),
	}

	oldIdx, newIdx := 0, 0
	lcsIdx := 0
	oldLineNo, newLineNo := 1, 1

	for oldIdx < len(oldLines) || newIdx < len(newLines) {
		if lcsIdx < len(lcs) && oldIdx < len(oldLines) && newIdx < len(newLines) &&
			oldLines[oldIdx] == lcs[lcsIdx] && newLines[newIdx] == lcs[lcsIdx] {
			// Common line
			hunk.Lines = append(hunk.Lines, DiffLine{
				Type:      DiffLineContext,
				Content:   oldLines[oldIdx],
				OldLineNo: oldLineNo,
				NewLineNo: newLineNo,
			})
			oldIdx++
			newIdx++
			lcsIdx++
			oldLineNo++
			newLineNo++
		} else if oldIdx < len(oldLines) && (lcsIdx >= len(lcs) || oldLines[oldIdx] != lcs[lcsIdx]) {
			// Removed line
			hunk.Lines = append(hunk.Lines, DiffLine{
				Type:      DiffLineRemoved,
				Content:   oldLines[oldIdx],
				OldLineNo: oldLineNo,
			})
			oldIdx++
			oldLineNo++
		} else if newIdx < len(newLines) && (lcsIdx >= len(lcs) || newLines[newIdx] != lcs[lcsIdx]) {
			// Added line
			hunk.Lines = append(hunk.Lines, DiffLine{
				Type:      DiffLineAdded,
				Content:   newLines[newIdx],
				NewLineNo: newLineNo,
			})
			newIdx++
			newLineNo++
		}
	}

	if len(hunk.Lines) > 0 {
		result.Hunks = append(result.Hunks, hunk)
	}

	return result
}

// longestCommonSubsequence finds the LCS of two string slices
func longestCommonSubsequence(a, b []string) []string {
	m, n := len(a), len(b)

	// Build LCS table
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	// Backtrack to find LCS
	lcs := make([]string, 0, dp[m][n])
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			lcs = append([]string{a[i-1]}, lcs...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return lcs
}
