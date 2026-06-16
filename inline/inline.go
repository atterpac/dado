package inline

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/term"
)

// ansiRE matches ANSI escape sequences (SGR color codes, OSC hyperlinks)
// so they can be ignored when measuring display width.
var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m|\x1b\]8;;[^\x1b]*\x1b\\`)

// stripANSI removes ANSI escape sequences from a string.
func stripANSI(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}

// displayWidth returns the number of terminal cells a string occupies.
// It ignores ANSI escape codes, counts runes (not bytes), treats combining
// marks as zero-width, and counts wide runes (CJK, emoji) as two cells. All
// inline padding/alignment code should use this instead of len().
func displayWidth(s string) int {
	w := 0
	for _, r := range stripANSI(s) {
		w += runeWidth(r)
	}
	return w
}

// termWidth returns the terminal width in columns, defaulting to 80. It
// sizes off the current output writer when that writer is a terminal.
func termWidth() int {
	if f, ok := out.(*os.File); ok {
		if w, _, err := term.GetSize(int(f.Fd())); err == nil && w > 0 {
			return w
		}
	}
	return 80
}

// truncate shortens s to at most max display cells, appending "…" if cut.
// ANSI codes are preserved as-is but not counted toward the width.
func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if displayWidth(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := runeWidth(r)
		if w+rw > max-1 {
			break
		}
		b.WriteRune(r)
		w += rw
	}
	b.WriteRune('…')
	return b.String()
}

func runeWidth(r rune) int {
	if r == 0 {
		return 0
	}
	if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Cf, r) {
		return 0 // combining / enclosing marks, format chars
	}
	if isWide(r) {
		return 2
	}
	return 1
}

func isWide(r rune) bool {
	switch {
	case r >= 0x1100 && r <= 0x115F, // Hangul Jamo
		r >= 0x2E80 && r <= 0x303E, // CJK radicals, Kangxi
		r >= 0x3041 && r <= 0x33FF, // Hiragana..CJK symbols
		r >= 0x3400 && r <= 0x4DBF, // CJK Ext A
		r >= 0x4E00 && r <= 0x9FFF, // CJK Unified
		r >= 0xA000 && r <= 0xA4CF, // Yi
		r >= 0xAC00 && r <= 0xD7A3, // Hangul syllables
		r >= 0xF900 && r <= 0xFAFF, // CJK compatibility
		r >= 0xFE30 && r <= 0xFE4F, // CJK compatibility forms
		r >= 0xFF00 && r <= 0xFF60, // Fullwidth forms
		r >= 0xFFE0 && r <= 0xFFE6,
		r >= 0x1F300 && r <= 0x1FAFF, // emoji & pictographs
		r >= 0x20000 && r <= 0x3FFFD: // CJK Ext B+
		return true
	}
	return false
}

// DisplayWidth returns the number of terminal cells a string occupies,
// ignoring ANSI escape codes and accounting for wide/zero-width runes.
// Use it instead of len() when aligning columns by hand.
func DisplayWidth(s string) int { return displayWidth(s) }

// Pad right-pads s with spaces to the given display width (cell-accurate).
func Pad(s string, width int) string { return padRight(s, width) }

// padRight pads s with spaces to the given display width.
func padRight(s string, width int) string {
	if pad := width - displayWidth(s); pad > 0 {
		return s + strings.Repeat(" ", pad)
	}
	return s
}

// PrintWarning prints a warning message.
func PrintWarning(msg string) {
	fmt.Fprintf(out, "  %s%s⚠%s %s%s%s\n", Bold, Yellow, Reset, Yellow, msg, Reset)
}

// PrintKV prints aligned key/value pairs.
func PrintKV(pairs ...[2]string) {
	keyWidth := 0
	for _, p := range pairs {
		if w := displayWidth(p[0]); w > keyWidth {
			keyWidth = w
		}
	}
	for _, p := range pairs {
		fmt.Fprintf(out, "    %s%s%s %s:%s %s\n",
			Cyan, padRight(p[0], keyWidth), Reset,
			Dim, Reset, p[1])
	}
}

// PrintList prints a bulleted list.
func PrintList(items ...string) {
	for _, item := range items {
		fmt.Fprintf(out, "    %s•%s %s\n", Dim, Reset, item)
	}
}

// Hyperlink returns an OSC 8 terminal hyperlink. Terminals that don't
// support it fall back to showing the text.
func Hyperlink(text, url string) string {
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}

// PrintTable prints an auto-sized table with a header row. Columns are
// shrunk and cells truncated with "…" so the table never exceeds the
// terminal width.
func PrintTable(headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = displayWidth(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				if w := displayWidth(cell); w > widths[i] {
					widths[i] = w
				}
			}
		}
	}

	// Fit to terminal: indent(4) + sum(widths) + 2 per gap must fit.
	const indent, gap, minCol = 4, 2, 3
	budget := termWidth() - indent - gap*len(widths)
	for sum(widths) > budget {
		// Shrink the widest column that is still above the minimum.
		wi := -1
		for i, w := range widths {
			if w > minCol && (wi == -1 || w > widths[wi]) {
				wi = i
			}
		}
		if wi == -1 {
			break // can't shrink further
		}
		widths[wi]--
	}

	// Header
	fmt.Fprint(out, "    ")
	for i, h := range headers {
		fmt.Fprintf(out, "%s%s%s%s  ", Bold, BrightWhite, padRight(truncate(h, widths[i]), widths[i]), Reset)
	}
	fmt.Fprintln(out)

	// Separator
	fmt.Fprint(out, "    ")
	for _, w := range widths {
		fmt.Fprintf(out, "%s%s%s  ", Dim, strings.Repeat("─", w), Reset)
	}
	fmt.Fprintln(out)

	// Rows
	for _, row := range rows {
		fmt.Fprint(out, "    ")
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			fmt.Fprintf(out, "%s  ", padRight(truncate(cell, widths[i]), widths[i]))
		}
		fmt.Fprintln(out)
	}
}

func sum(xs []int) int {
	t := 0
	for _, x := range xs {
		t += x
	}
	return t
}

// TreeNode is a node in a printable tree.
type TreeNode struct {
	Label    string
	Children []TreeNode
}

// PrintTree prints a file-tree style listing with ├─ └─ connectors.
func PrintTree(nodes []TreeNode) {
	printTreeNodes(nodes, "")
}

func printTreeNodes(nodes []TreeNode, prefix string) {
	for i, n := range nodes {
		last := i == len(nodes)-1
		branch := "├─"
		childPrefix := prefix + "│  "
		if last {
			branch = "└─"
			childPrefix = prefix + "   "
		}
		fmt.Fprintf(out, "    %s%s%s%s %s\n", Dim, prefix, branch, Reset, n.Label)
		if len(n.Children) > 0 {
			printTreeNodes(n.Children, childPrefix)
		}
	}
}

// ProgressBar renders a single-line progress bar. Call repeatedly with
// increasing pct (0.0–1.0); it redraws in place. Pass done=true on the
// final call to move to the next line.
func ProgressBar(label string, pct float64, done bool) {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	// Without a TTY, in-place redraws are meaningless; emit only the
	// final line so logs stay clean.
	if !stdoutIsTTY {
		if done {
			fmt.Fprintf(out, "  %s %3.0f%%\n", label, pct*100)
		}
		return
	}
	const width = 24
	filled := int(pct * width)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Fprintf(out, "\r  %s%s%s [%s%s%s] %3.0f%%", Bold, label, Reset, Cyan, bar, Reset, pct*100)
	if done {
		fmt.Fprintln(out)
	}
}

// Spinner is an animated single-line loading indicator.
type Spinner struct {
	frames []string
	mu     sync.Mutex
	done   chan struct{}
	label  string
}

// NewSpinner creates a spinner with the given label.
func NewSpinner(label string) *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		label:  label,
	}
}

// Start begins animating the spinner in a background goroutine. On a
// non-TTY it prints a single static "step" line instead of animating.
func (s *Spinner) Start() {
	if !stdoutIsTTY {
		PrintStep(s.label)
		return
	}
	s.done = make(chan struct{})
	go func() {
		i := 0
		t := time.NewTicker(80 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-s.done:
				return
			case <-t.C:
				s.mu.Lock()
				fmt.Fprintf(out, "\r  %s%s%s %s", Cyan, s.frames[i%len(s.frames)], Reset, s.label)
				s.mu.Unlock()
				i++
			}
		}
	}()
}

// Stop halts the spinner and clears its line, then prints a success line.
func (s *Spinner) Stop(msg string) {
	if s.done != nil {
		close(s.done)
		s.mu.Lock()
		fmt.Fprint(out, "\r\033[K") // carriage return + clear to EOL
		s.mu.Unlock()
	}
	PrintSuccess(msg)
}

// Confirm asks a yes/no question and returns the answer. Defaults to def
// on an empty response. Returns def directly when stdin is not a terminal.
func Confirm(prompt string, def bool) bool {
	hint := "[y/N]"
	if def {
		hint = "[Y/n]"
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return def
	}
	fmt.Fprintf(out, "  %s%s?%s %s %s%s%s ", Bold, Yellow, Reset, prompt, Dim, hint, Reset)
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return def
	}
}

// Input prompts for a line of text, returning def if the response is empty
// or stdin is not a terminal.
func Input(prompt, def string) string {
	hint := ""
	if def != "" {
		hint = fmt.Sprintf(" %s(%s)%s", Dim, def, Reset)
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return def
	}
	fmt.Fprintf(out, "  %s%s?%s %s%s %s›%s ", Bold, Yellow, Reset, prompt, hint, Cyan, Reset)
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	if s := strings.TrimSpace(line); s != "" {
		return s
	}
	return def
}

// PrintDiff prints a unified-style line diff: removed lines in red with
// "-", added lines in green with "+", context unchanged.
func PrintDiff(old, new []string) {
	for _, line := range diffLines(old, new) {
		switch line.kind {
		case diffDel:
			fmt.Fprintf(out, "    %s-%s %s%s%s\n", Red, Reset, Red, line.text, Reset)
		case diffAdd:
			fmt.Fprintf(out, "    %s+%s %s%s%s\n", Green, Reset, Green, line.text, Reset)
		default:
			fmt.Fprintf(out, "    %s %s%s\n", Dim, Reset, line.text)
		}
	}
}

type diffKind int

const (
	diffCtx diffKind = iota
	diffDel
	diffAdd
)

type diffLine struct {
	kind diffKind
	text string
}

// diffLines computes a line-level diff via a longest-common-subsequence
// table — small and dependency-free, fine for short CLI output.
func diffLines(a, b []string) []diffLine {
	n, m := len(a), len(b)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}
	var out []diffLine
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			out = append(out, diffLine{diffCtx, a[i]})
			i, j = i+1, j+1
		case lcs[i+1][j] >= lcs[i][j+1]:
			out = append(out, diffLine{diffDel, a[i]})
			i++
		default:
			out = append(out, diffLine{diffAdd, b[j]})
			j++
		}
	}
	for ; i < n; i++ {
		out = append(out, diffLine{diffDel, a[i]})
	}
	for ; j < m; j++ {
		out = append(out, diffLine{diffAdd, b[j]})
	}
	return out
}

// StatusList tracks a sequence of steps, redrawing each with a pending /
// running / done / failed marker. Use for multi-phase work where one
// spinner isn't enough.
type StatusList struct {
	steps []statusStep
}

type stepState int

const (
	StepPending stepState = iota
	StepRunning
	StepDone
	StepFailed
)

type statusStep struct {
	label string
	state stepState
}

// NewStatusList creates a status list from step labels (all pending).
func NewStatusList(labels ...string) *StatusList {
	s := &StatusList{}
	for _, l := range labels {
		s.steps = append(s.steps, statusStep{label: l, state: StepPending})
	}
	if stdoutIsTTY {
		s.render(false)
	}
	return s
}

// Set updates the i-th step's state and redraws.
func (s *StatusList) Set(i int, state stepState) {
	if i < 0 || i >= len(s.steps) {
		return
	}
	s.steps[i].state = state
	if stdoutIsTTY {
		s.render(true)
		return
	}
	// Non-TTY: emit one stable line per terminal transition; skip the
	// pending/running churn so logs stay clean.
	if state == StepDone || state == StepFailed {
		marker, _ := stepMarker(state)
		fmt.Fprintf(out, "  %s %s\n", marker, s.steps[i].label)
	}
}

func stepMarker(state stepState) (string, string) {
	switch state {
	case StepRunning:
		return "▸", Cyan
	case StepDone:
		return "✓", Green
	case StepFailed:
		return "✗", Red
	default:
		return "○", Dim
	}
}

func (s *StatusList) render(redraw bool) {
	// Move the cursor up to overwrite the previous render (TTY only).
	if redraw {
		fmt.Fprintf(out, "\033[%dA", len(s.steps))
	}
	for _, st := range s.steps {
		marker, color := stepMarker(st.state)
		fmt.Fprintf(out, "  %s%s%s %s\033[K\n", color, marker, Reset, st.label)
	}
}
