package components

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// DiffLineType indicates the type of diff line
type DiffLineType int

const (
	DiffLineContext  DiffLineType = iota // Unchanged line
	DiffLineAdded                        // Added line (+)
	DiffLineRemoved                      // Removed line (-)
	DiffLineHeader                       // Hunk header (@@ ... @@)
	DiffLineFile                         // File header (diff --git, ---, +++)
)

// DiffLine represents a single line in the diff
type DiffLine struct {
	Type      DiffLineType
	OldLineNo int    // Line number in old content (0 if added)
	NewLineNo int    // Line number in new content (0 if removed)
	Content   string // Line content without +/- prefix
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

// DiffResult contains the complete diff
type DiffResult struct {
	OldName string     // Original file name
	NewName string     // New file name
	Hunks   []DiffHunk // All hunks
	Binary  bool       // True if binary file
}

// DiffStats provides summary statistics
type DiffStats struct {
	Additions int
	Deletions int
	Files     int
}

// DiffViewer displays diff content with syntax highlighting
type DiffViewer struct {
	*tview.Box

	result        *DiffResult
	lines         []DiffLine // Flattened lines for display
	selectedIndex int
	offset        int

	// Display options
	sideBySide      bool
	showLineNumbers bool
	wordDiff        bool
	contextLines    int
	title           string

	// Callbacks
	onLineSelect func(line DiffLine)
}

// NewDiffViewer creates a new diff viewer component
func NewDiffViewer() *DiffViewer {
	return &DiffViewer{
		Box:             tview.NewBox(),
		showLineNumbers: true,
		contextLines:    3,
	}
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

	for _, hunk := range d.result.Hunks {
		// Add hunk header
		d.lines = append(d.lines, DiffLine{
			Type:    DiffLineHeader,
			Content: hunk.Header,
		})
		// Add hunk lines
		d.lines = append(d.lines, hunk.Lines...)
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
	d.Box.DrawForSubclass(screen, d)
	x, y, width, height := d.GetInnerRect()

	if width <= 0 || height <= 0 || len(d.lines) == 0 {
		return
	}

	// Colors (read at draw time per jig rules)
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()
	addedColor := theme.Success()
	removedColor := theme.Error()
	headerColor := theme.Info()

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
		for col := x; col < x+width; col++ {
			screen.SetContent(col, rowY, ' ', nil, lineStyle)
		}

		col := x

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

// InputHandler handles keyboard input
func (d *DiffViewer) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return d.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if len(d.lines) == 0 {
			return
		}

		switch event.Key() {
		case tcell.KeyDown:
			d.moveDown()
		case tcell.KeyUp:
			d.moveUp()
		case tcell.KeyHome:
			d.selectedIndex = 0
			d.ensureVisible()
		case tcell.KeyEnd:
			d.selectedIndex = len(d.lines) - 1
			d.ensureVisible()
		case tcell.KeyPgDn:
			_, _, _, height := d.GetInnerRect()
			d.selectedIndex += height
			if d.selectedIndex >= len(d.lines) {
				d.selectedIndex = len(d.lines) - 1
			}
			d.ensureVisible()
		case tcell.KeyPgUp:
			_, _, _, height := d.GetInnerRect()
			d.selectedIndex -= height
			if d.selectedIndex < 0 {
				d.selectedIndex = 0
			}
			d.ensureVisible()
		case tcell.KeyEnter:
			if d.onLineSelect != nil && d.selectedIndex < len(d.lines) {
				d.onLineSelect(d.lines[d.selectedIndex])
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				d.moveDown()
			case 'k':
				d.moveUp()
			case 'g':
				d.selectedIndex = 0
				d.ensureVisible()
			case 'G':
				d.selectedIndex = len(d.lines) - 1
				d.ensureVisible()
			case 'n':
				d.NextChange()
			case 'N':
				d.PrevChange()
			case 'u':
				d.sideBySide = false
			case 's':
				d.sideBySide = true
			case 'l':
				d.showLineNumbers = !d.showLineNumbers
			case 'w':
				d.wordDiff = !d.wordDiff
			}
		case tcell.KeyCtrlD:
			_, _, _, height := d.GetInnerRect()
			d.selectedIndex += height / 2
			if d.selectedIndex >= len(d.lines) {
				d.selectedIndex = len(d.lines) - 1
			}
			d.ensureVisible()
		case tcell.KeyCtrlU:
			_, _, _, height := d.GetInnerRect()
			d.selectedIndex -= height / 2
			if d.selectedIndex < 0 {
				d.selectedIndex = 0
			}
			d.ensureVisible()
		}
	})
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

// MouseHandler handles mouse input
func (d *DiffViewer) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return d.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		_, y, _, _ := d.GetInnerRect()
		mx, my := event.Position()

		if !d.InRect(mx, my) {
			return false, nil
		}

		switch action {
		case tview.MouseLeftClick:
			setFocus(d)
			clickedIndex := d.offset + (my - y)
			if clickedIndex >= 0 && clickedIndex < len(d.lines) {
				d.selectedIndex = clickedIndex
				return true, d
			}
		case tview.MouseLeftDoubleClick:
			clickedIndex := d.offset + (my - y)
			if clickedIndex >= 0 && clickedIndex < len(d.lines) {
				d.selectedIndex = clickedIndex
				if d.onLineSelect != nil {
					d.onLineSelect(d.lines[clickedIndex])
				}
				return true, d
			}
		case tview.MouseScrollUp:
			if d.offset > 0 {
				d.offset--
			}
			return true, d
		case tview.MouseScrollDown:
			_, _, _, height := d.GetInnerRect()
			if d.offset < len(d.lines)-height {
				d.offset++
			}
			return true, d
		}

		return false, nil
	})
}

// Focus handles focus
func (d *DiffViewer) Focus(delegate func(tview.Primitive)) {
	d.Box.Focus(delegate)
}

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
