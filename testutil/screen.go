package testutil

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TestScreen wraps tcell.SimulationScreen with helper methods for testing.
type TestScreen struct {
	tcell.SimulationScreen
}

// NewTestScreen creates a new test screen with the given dimensions.
func NewTestScreen(width, height int) *TestScreen {
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	screen.SetSize(width, height)
	return &TestScreen{screen}
}

// GetContent extracts a string of characters from the screen starting at (x, y)
// for the given length.
func (ts *TestScreen) GetContent(x, y, length int) string {
	var runes []rune
	for i := 0; i < length; i++ {
		str, _, _ := ts.SimulationScreen.Get(x+i, y)
		if len(str) > 0 {
			runes = append(runes, []rune(str)[0])
		} else {
			runes = append(runes, ' ')
		}
	}
	return string(runes)
}

// GetRow extracts an entire row from the screen.
func (ts *TestScreen) GetRow(y int) string {
	w, _ := ts.Size()
	return ts.GetContent(0, y, w)
}

// GetRect extracts a rectangular region from the screen.
func (ts *TestScreen) GetRect(x, y, width, height int) []string {
	rows := make([]string, height)
	for i := 0; i < height; i++ {
		rows[i] = ts.GetContent(x, y+i, width)
	}
	return rows
}

// ContainsText checks if the screen contains the given text anywhere.
func (ts *TestScreen) ContainsText(text string) bool {
	_, h := ts.Size()
	for y := 0; y < h; y++ {
		row := ts.GetRow(y)
		if strings.Contains(row, text) {
			return true
		}
	}
	return false
}

// FindText returns the (x, y) position of the first occurrence of text,
// or (-1, -1) if not found.
func (ts *TestScreen) FindText(text string) (int, int) {
	_, h := ts.Size()
	for y := 0; y < h; y++ {
		row := ts.GetRow(y)
		if x := strings.Index(row, text); x >= 0 {
			return x, y
		}
	}
	return -1, -1
}

// GetStyleAt returns the style at the given position.
func (ts *TestScreen) GetStyleAt(x, y int) tcell.Style {
	_, style, _ := ts.SimulationScreen.Get(x, y)
	return style
}

// GetForegroundAt returns the foreground color at the given position.
func (ts *TestScreen) GetForegroundAt(x, y int) tcell.Color {
	fg, _, _ := ts.GetStyleAt(x, y).Decompose()
	return fg
}

// GetBackgroundAt returns the background color at the given position.
func (ts *TestScreen) GetBackgroundAt(x, y int) tcell.Color {
	_, bg, _ := ts.GetStyleAt(x, y).Decompose()
	return bg
}

// Clear resets the screen and clears all content.
func (ts *TestScreen) Clear() {
	ts.SimulationScreen.Clear()
	ts.SimulationScreen.Show()
}

// DrawPrimitive renders a tview.Primitive to the test screen.
func (ts *TestScreen) DrawPrimitive(p tview.Primitive) {
	w, h := ts.Size()
	p.SetRect(0, 0, w, h)
	p.Draw(ts.SimulationScreen)
	ts.SimulationScreen.Show()
}

// DrawPrimitiveAt renders a tview.Primitive at a specific position.
func (ts *TestScreen) DrawPrimitiveAt(p tview.Primitive, x, y, width, height int) {
	p.SetRect(x, y, width, height)
	p.Draw(ts.SimulationScreen)
	ts.SimulationScreen.Show()
}

// Dump returns a string representation of the entire screen for debugging.
func (ts *TestScreen) Dump() string {
	_, h := ts.Size()
	var sb strings.Builder
	for y := 0; y < h; y++ {
		sb.WriteString(ts.GetRow(y))
		sb.WriteRune('\n')
	}
	return sb.String()
}

// DumpTrimmed returns a trimmed dump with trailing whitespace removed.
func (ts *TestScreen) DumpTrimmed() string {
	w, h := ts.Size()
	var lines []string
	for y := 0; y < h; y++ {
		line := strings.TrimRight(ts.GetContent(0, y, w), " ")
		lines = append(lines, line)
	}
	// Remove trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// AssertContains fails the test if the screen does not contain the text.
func (ts *TestScreen) AssertContains(t interface{ Helper(); Errorf(string, ...any) }, text string) {
	t.Helper()
	if !ts.ContainsText(text) {
		t.Errorf("screen does not contain %q\n%s", text, ts.Dump())
	}
}

// AssertNotContains fails the test if the screen contains the text.
func (ts *TestScreen) AssertNotContains(t interface{ Helper(); Errorf(string, ...any) }, text string) {
	t.Helper()
	if ts.ContainsText(text) {
		t.Errorf("screen should not contain %q\n%s", text, ts.Dump())
	}
}

// AssertTextAt fails the test if the text at (x, y) does not match.
func (ts *TestScreen) AssertTextAt(t interface{ Helper(); Errorf(string, ...any) }, x, y int, expected string) {
	t.Helper()
	actual := ts.GetContent(x, y, len(expected))
	if actual != expected {
		t.Errorf("at (%d, %d): expected %q, got %q", x, y, expected, actual)
	}
}
