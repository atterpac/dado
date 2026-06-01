package testutil

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/atterpac/dado/core"
)

// CastRecorder records a sequence of rendered frames into an asciinema v2 cast.
// The same script drives both demo generation and golden-file regression testing.
//
// Basic usage:
//
//	rec := testutil.NewCastRecorder(80, 24)
//	rec.Capture(myWidget, 1.0)            // initial frame, hold 1 second
//	rec.Step(myWidget, func() {           // simulate action then capture
//	    testutil.SimulateKey(myWidget.InputHandler(), tcell.KeyDown)
//	}, 0.5)
//	rec.AssertGolden(t, "testdata/widget.cast", "Widget Demo")
//	// Set UPDATE_CAST=1 to regenerate golden files.
type CastRecorder struct {
	screen  *TestScreen
	events  []castEvent
	elapsed float64
}

type castEvent struct {
	time float64
	data string
}

// NewCastRecorder creates a recorder with a simulation screen of the given size.
func NewCastRecorder(width, height int) *CastRecorder {
	return &CastRecorder{
		screen: NewTestScreen(width, height),
	}
}

// Screen returns the underlying TestScreen for direct inspection between steps.
func (r *CastRecorder) Screen() *TestScreen { return r.screen }

// Capture renders p at its current state and records a frame held for holdSecs seconds.
func (r *CastRecorder) Capture(p core.Widget, holdSecs float64) {
	r.screen.DrawPrimitive(p)
	frame := screenToANSI(r.screen)
	r.events = append(r.events, castEvent{time: r.elapsed, data: frame})
	r.elapsed += holdSecs
}

// Step runs action (use nil to skip), then captures a frame held for holdSecs seconds.
// Combine with SimulateKey/SimulateRune/TypeString from helpers.go to drive input.
func (r *CastRecorder) Step(p core.Widget, action func(), holdSecs float64) {
	if action != nil {
		action()
	}
	r.Capture(p, holdSecs)
}

// Write serializes the recording as an asciinema v2 cast to w.
func (r *CastRecorder) Write(w io.Writer, title string) error {
	width, height := r.screen.Size()
	header := map[string]any{
		"version":   2,
		"width":     width,
		"height":    height,
		"timestamp": 0,
		"title":     title,
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(header); err != nil {
		return err
	}
	for _, ev := range r.events {
		row := []any{ev.time, "o", ev.data}
		if err := enc.Encode(row); err != nil {
			return err
		}
	}
	return nil
}

// WriteTo writes the cast to a file at path, creating or overwriting it.
func (r *CastRecorder) WriteTo(path, title string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return r.Write(f, title)
}

// AssertGolden compares the cast output against a golden file at path.
// On mismatch it reports which frames changed and shows a text diff of the
// screen content (ANSI stripped) for each differing frame.
// Set UPDATE_CAST=1 to regenerate the golden file.
func (r *CastRecorder) AssertGolden(t testing.TB, path, title string) {
	t.Helper()
	var sb strings.Builder
	if err := r.Write(&sb, title); err != nil {
		t.Fatalf("cast write: %v", err)
	}
	got := sb.String()

	if os.Getenv("UPDATE_CAST") != "" {
		if err := os.MkdirAll(pathDir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("updated golden: %s", path)
		return
	}
	wantBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing golden %s (run UPDATE_CAST=1 to create): %v", path, err)
	}
	want := string(wantBytes)
	if want == got {
		return
	}

	wantFrames := parseCastFrames(want)
	gotFrames := parseCastFrames(got)

	var buf strings.Builder
	fmt.Fprintf(&buf, "cast changed: %s  (UPDATE_CAST=1 to regenerate)\n", path)

	max := len(wantFrames)
	if len(gotFrames) > max {
		max = len(gotFrames)
	}
	for i := range max {
		var wf, gf castFrame
		if i < len(wantFrames) {
			wf = wantFrames[i]
		}
		if i < len(gotFrames) {
			gf = gotFrames[i]
		}
		if wf.data == gf.data {
			continue
		}
		fmt.Fprintf(&buf, "\nframe %d (t=%.2fs):\n", i, gf.time)
		diffCells(&buf, decodeANSI(wf.data), decodeANSI(gf.data))
	}
	t.Error(buf.String())
}

type castFrame struct {
	time float64
	data string
}

// parseCastFrames extracts the output frames from an asciinema v2 cast string.
// Line 0 is the header; subsequent lines are [time, "o", data] events.
func parseCastFrames(cast string) []castFrame {
	var frames []castFrame
	for i, line := range strings.Split(strings.TrimSpace(cast), "\n") {
		if i == 0 || line == "" {
			continue
		}
		var row []json.RawMessage
		if err := json.Unmarshal([]byte(line), &row); err != nil || len(row) < 3 {
			continue
		}
		var t float64
		var kind, data string
		if err := json.Unmarshal(row[0], &t); err != nil {
			continue
		}
		if err := json.Unmarshal(row[1], &kind); err != nil || kind != "o" {
			continue
		}
		if err := json.Unmarshal(row[2], &data); err != nil {
			continue
		}
		frames = append(frames, castFrame{time: t, data: data})
	}
	return frames
}

// decodedCell is a single screen cell with its rune and resolved color strings.
type decodedCell struct {
	ch rune
	fg string // "#rrggbb" or "" for default
	bg string
}

// decodedRow is one row of decoded cells.
type decodedRow []decodedCell

// decodeANSI parses an asciinema frame's ANSI data into rows of cells.
// It handles the SGR sequences emitted by styleToSGR (reset, 24-bit, 256, named).
func decodeANSI(s string) []decodedRow {
	var rows []decodedRow
	var row decodedRow
	var curFg, curBg string

	parseSGR := func(params string) {
		parts := strings.Split(params, ";")
		for i := 0; i < len(parts); i++ {
			switch parts[i] {
			case "", "0":
				curFg, curBg = "", ""
			case "1", "2", "3", "4", "7": // attrs — ignore for color diff
			case "38":
				if i+4 < len(parts) && parts[i+1] == "2" {
					curFg = fmt.Sprintf("#%02x%02x%02x", atoi(parts[i+2]), atoi(parts[i+3]), atoi(parts[i+4]))
					i += 4
				} else if i+2 < len(parts) && parts[i+1] == "5" {
					curFg = fmt.Sprintf("c%s", parts[i+2])
					i += 2
				}
			case "48":
				if i+4 < len(parts) && parts[i+1] == "2" {
					curBg = fmt.Sprintf("#%02x%02x%02x", atoi(parts[i+2]), atoi(parts[i+3]), atoi(parts[i+4]))
					i += 4
				} else if i+2 < len(parts) && parts[i+1] == "5" {
					curBg = fmt.Sprintf("c%s", parts[i+2])
					i += 2
				}
			default:
				n := atoi(parts[i])
				switch {
				case n >= 30 && n <= 37:
					curFg = fmt.Sprintf("ansi%d", n-30)
				case n >= 90 && n <= 97:
					curFg = fmt.Sprintf("ansi%d", n-90+8)
				case n >= 40 && n <= 47:
					curBg = fmt.Sprintf("ansi%d", n-40)
				case n >= 100 && n <= 107:
					curBg = fmt.Sprintf("ansi%d", n-100+8)
				}
			}
		}
	}

	i := 0
	for i < len(s) {
		// Skip the home+clear preamble (\x1b[H\x1b[2J).
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			start := i
			for i < len(s) && (s[i] < 0x40 || s[i] > 0x7e) {
				i++
			}
			if i < len(s) {
				final := s[i]
				i++
				if final == 'm' {
					parseSGR(s[start : i-1])
				}
				// H, J — skip (cursor/clear)
			}
			continue
		}
		if i+1 < len(s) && s[i] == '\r' && s[i+1] == '\n' {
			rows = append(rows, row)
			row = nil
			i += 2
			continue
		}
		r, size := rune(s[i]), 1
		if s[i] >= 0x80 {
			// multi-byte UTF-8
			var tmp [4]byte
			n := copy(tmp[:], s[i:])
			r2, sz := rune(0), 0
			for j := range n {
				tmp2 := string(tmp[:j+1])
				runes := []rune(tmp2)
				if len(runes) == 1 && runes[0] != 0xfffd {
					r2, sz = runes[0], j+1
				}
			}
			if sz > 0 {
				r, size = r2, sz
			}
		}
		row = append(row, decodedCell{ch: r, fg: curFg, bg: curBg})
		i += size
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	return rows
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// diffCells writes a per-row diff between two decoded frames.
// Rows with no cell-level differences are printed as plain text.
// Changed rows show - (want) and + (got) with inline color annotations.
func diffCells(buf *strings.Builder, want, got []decodedRow) {
	max := len(want)
	if len(got) > max {
		max = len(got)
	}
	for i := range max {
		var wr, gr decodedRow
		if i < len(want) {
			wr = want[i]
		}
		if i < len(got) {
			gr = got[i]
		}
		if rowsEqual(wr, gr) {
			fmt.Fprintf(buf, "  %s\n", rowRunes(gr))
			continue
		}
		fmt.Fprintf(buf, "- %s\n", rowStyled(wr))
		fmt.Fprintf(buf, "+ %s\n", rowStyled(gr))
	}
}

func rowsEqual(a, b decodedRow) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func rowRunes(r decodedRow) string {
	var sb strings.Builder
	for _, c := range r {
		sb.WriteRune(c.ch)
	}
	return strings.TrimRight(sb.String(), " ")
}

// rowStyled annotates runs of cells that share the same fg/bg.
// Output: "text[fg=X bg=Y]text[fg=A bg=B]"
func rowStyled(r decodedRow) string {
	if len(r) == 0 {
		return ""
	}
	var sb strings.Builder
	start := 0
	for i := 1; i <= len(r); i++ {
		if i == len(r) || r[i].fg != r[start].fg || r[i].bg != r[start].bg {
			text := strings.TrimRight(rowRunes(r[start:i]), "")
			fg, bg := r[start].fg, r[start].bg
			if fg == "" && bg == "" {
				sb.WriteString(text)
			} else {
				fmt.Fprintf(&sb, "%s[fg=%s bg=%s]", text, fg, bg)
			}
			start = i
		}
	}
	return strings.TrimRight(sb.String(), " []")
}

// screenToANSI converts the current screen state to a VT100 string suitable
// for an asciinema "o" event. Emits home+clear then the full frame.
func screenToANSI(ts *TestScreen) string {
	cells, w, h := ts.SimulationScreen.GetContents()
	var sb strings.Builder

	// Home cursor + erase display so the player redraws cleanly.
	sb.WriteString("\x1b[H\x1b[2J")

	var lastStyle tcell.Style
	styleSet := false

	for y := range h {
		if y > 0 {
			sb.WriteString("\r\n")
		}
		for x := range w {
			cell := cells[y*w+x]
			if !styleSet || cell.Style != lastStyle {
				sb.WriteString(styleToSGR(cell.Style))
				lastStyle = cell.Style
				styleSet = true
			}
			if len(cell.Runes) == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(cell.Runes[0])
			}
		}
	}
	sb.WriteString("\x1b[0m")
	return sb.String()
}

// styleToSGR converts a tcell.Style to an ANSI SGR escape sequence.
func styleToSGR(s tcell.Style) string {
	fg, bg, attrs := s.Decompose()
	var parts []string
	parts = append(parts, "0") // always reset first

	if fg != tcell.ColorDefault {
		parts = append(parts, colorSGR(fg, true))
	}
	if bg != tcell.ColorDefault {
		parts = append(parts, colorSGR(bg, false))
	}
	if attrs&tcell.AttrBold != 0 {
		parts = append(parts, "1")
	}
	if attrs&tcell.AttrDim != 0 {
		parts = append(parts, "2")
	}
	if attrs&tcell.AttrItalic != 0 {
		parts = append(parts, "3")
	}
	if attrs&tcell.AttrUnderline != 0 {
		parts = append(parts, "4")
	}
	if attrs&tcell.AttrReverse != 0 {
		parts = append(parts, "7")
	}

	return "\x1b[" + strings.Join(parts, ";") + "m"
}

// colorSGR returns the SGR parameter string for a tcell.Color.
// RGB colors use 24-bit sequences; others fall back to 256-color.
func colorSGR(c tcell.Color, fg bool) string {
	if c.IsRGB() {
		r, g, b := c.RGB()
		if fg {
			return fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
		}
		return fmt.Sprintf("48;2;%d;%d;%d", r, g, b)
	}
	// Named / 256-color palette: tcell encodes these as ColorValid | index.
	idx := int(c) & 0xff
	if fg {
		if idx < 8 {
			return fmt.Sprintf("%d", 30+idx)
		}
		if idx < 16 {
			return fmt.Sprintf("%d", 90+idx-8)
		}
		return fmt.Sprintf("38;5;%d", idx)
	}
	if idx < 8 {
		return fmt.Sprintf("%d", 40+idx)
	}
	if idx < 16 {
		return fmt.Sprintf("%d", 100+idx-8)
	}
	return fmt.Sprintf("48;5;%d", idx)
}

func pathDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}
