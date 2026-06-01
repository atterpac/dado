// Package coretest provides test helpers for the dado/core package.
// Import only in _test.go files — production code must not import coretest.
package coretest

import (
	"sync"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
)

// MockWidget is a core.Widget that records Draw calls. Use it to verify that
// layout containers (Flex, Pages) call Draw on children with correct rects.
// All methods are safe for concurrent use.
type MockWidget struct {
	Name      string
	DrawCount int
	x, y, w, h int
	focused   bool
	mu        sync.Mutex
}

// NewMockWidget returns a MockWidget satisfying core.Widget.
func NewMockWidget(name string) *MockWidget {
	return &MockWidget{Name: name}
}

func (m *MockWidget) Draw(_ tcell.Screen) { m.mu.Lock(); m.DrawCount++; m.mu.Unlock() }
func (m *MockWidget) Rect() (int, int, int, int) {
	m.mu.Lock(); defer m.mu.Unlock(); return m.x, m.y, m.w, m.h
}
func (m *MockWidget) SetRect(x, y, w, h int) {
	m.mu.Lock(); defer m.mu.Unlock(); m.x, m.y, m.w, m.h = x, y, w, h
}
func (m *MockWidget) Focus()         { m.mu.Lock(); m.focused = true; m.mu.Unlock() }
func (m *MockWidget) Blur()          { m.mu.Lock(); m.focused = false; m.mu.Unlock() }
func (m *MockWidget) HasFocus() bool { m.mu.Lock(); defer m.mu.Unlock(); return m.focused }

// MockKeyWidget is a MockWidget that also implements core.KeyHandler.
// It records the last key/rune received and returns consumed=true.
// All methods are safe for concurrent use.
type MockKeyWidget struct {
	MockWidget
	lastKey  tcell.Key
	lastRune rune
}

func NewMockKeyWidget() *MockKeyWidget { return &MockKeyWidget{} }

func (m *MockKeyWidget) HandleKey(ev *tcell.EventKey) bool {
	m.mu.Lock()
	m.lastKey = ev.Key()
	m.lastRune = ev.Rune()
	m.mu.Unlock()
	return true
}

// LastKey returns the last key received (thread-safe).
func (m *MockKeyWidget) LastKey() tcell.Key {
	m.mu.Lock(); defer m.mu.Unlock(); return m.lastKey
}

// LastRune returns the last rune received (thread-safe).
func (m *MockKeyWidget) LastRune() rune {
	m.mu.Lock(); defer m.mu.Unlock(); return m.lastRune
}

// TestScreen is a thin alias for tcell.SimulationScreen with Size() exposed.
type TestScreen struct {
	tcell.SimulationScreen
}

// NewTestScreen creates an initialized SimulationScreen at width×height.
func NewTestScreen(width, height int) *TestScreen {
	s := tcell.NewSimulationScreen("UTF-8")
	_ = s.Init()
	s.SetSize(width, height)
	return &TestScreen{s}
}

// DrawWidget renders w at 0,0,width,height on a fresh TestScreen and returns it.
// The screen has been Show()n so content is readable.
func DrawWidget(w core.Widget, width, height int) *TestScreen {
	screen := NewTestScreen(width, height)
	w.SetRect(0, 0, width, height)
	w.Draw(screen.SimulationScreen)
	screen.SimulationScreen.Show()
	return screen
}

// SimulateKey calls w.HandleKey if w implements core.KeyHandler.
// Returns whether the event was consumed.
func SimulateKey(w core.Widget, key tcell.Key) bool {
	kh, ok := w.(core.KeyHandler)
	if !ok {
		return false
	}
	ev := tcell.NewEventKey(key, 0, tcell.ModNone)
	return kh.HandleKey(ev)
}

// SimulateRune calls w.HandleKey with a rune event if w implements core.KeyHandler.
func SimulateRune(w core.Widget, r rune) bool {
	kh, ok := w.(core.KeyHandler)
	if !ok {
		return false
	}
	ev := tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
	return kh.HandleKey(ev)
}

// AssertFocused fails the test if fm.Focused() != w.
func AssertFocused(t interface {
	Helper()
	Errorf(string, ...any)
}, fm *core.FocusManager, w core.Widget) {
	t.Helper()
	if fm.Focused() != w {
		t.Errorf("expected widget to be focused, got %v", fm.Focused())
	}
}

// MockKeyWidget compile-time check that it satisfies core.KeyHandler.
var _ core.KeyHandler = (*MockKeyWidget)(nil)

// GetRow returns the string content of row y.
func (ts *TestScreen) GetRow(y int) string {
	w, _ := ts.Size()
	var runes []rune
	for x := 0; x < w; x++ {
		r, _, _ := ts.SimulationScreen.Get(x, y)
		if len(r) > 0 {
			runes = append(runes, []rune(r)[0])
		} else {
			runes = append(runes, ' ')
		}
	}
	return string(runes)
}

// GetForegroundAt returns the foreground color at (x, y).
func (ts *TestScreen) GetForegroundAt(x, y int) tcell.Color {
	_, style, _ := ts.SimulationScreen.Get(x, y)
	fg, _, _ := style.Decompose()
	return fg
}

// FindText returns the (x, y) of the first occurrence of text, or (-1, -1).
func (ts *TestScreen) FindText(text string) (int, int) {
	_, h := ts.Size()
	w, _ := ts.Size()
	for y := 0; y < h; y++ {
		var runes []rune
		for x := 0; x < w; x++ {
			r, _, _ := ts.SimulationScreen.Get(x, y)
			if len(r) > 0 {
				runes = append(runes, []rune(r)[0])
			} else {
				runes = append(runes, ' ')
			}
		}
		row := string(runes)
		if idx := indexStr(row, text); idx >= 0 {
			return idx, y
		}
	}
	return -1, -1
}

func indexStr(s, sub string) int {
	if len(sub) > len(s) { return -1 }
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub { return i }
	}
	return -1
}

// --- TestScreen style helpers ---

// GetStyleAt returns the tcell.Style at the given position.
func (ts *TestScreen) GetStyleAt(x, y int) tcell.Style {
	_, style, _ := ts.SimulationScreen.Get(x, y)
	return style
}

// GetBackgroundAt returns the background color at the given position.
func (ts *TestScreen) GetBackgroundAt(x, y int) tcell.Color {
	_, bg, _ := ts.GetStyleAt(x, y).Decompose()
	return bg
}

// ContainsText reports whether any row of the screen contains text.
func (ts *TestScreen) ContainsText(text string) bool {
	_, h := ts.Size()
	for y := 0; y < h; y++ {
		if containsAt(ts, y, text) {
			return true
		}
	}
	return false
}

func containsAt(ts *TestScreen, y int, text string) bool {
	w, _ := ts.Size()
	var runes []rune
	for x := 0; x < w; x++ {
		r, _, _ := ts.SimulationScreen.Get(x, y)
		if len(r) > 0 {
			runes = append(runes, []rune(r)[0])
		} else {
			runes = append(runes, ' ')
		}
	}
	row := string(runes)
	return len(text) > 0 && findStr(row, text)
}

func findStr(s, sub string) bool {
	if len(sub) > len(s) { return false }
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub { return true }
	}
	return false
}

// Dump returns the full screen content as a string for debugging.
func (ts *TestScreen) Dump() string {
	w, h := ts.Size()
	var sb []byte
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, _, _ := ts.SimulationScreen.Get(x, y)
			if len(r) > 0 {
				sb = append(sb, []byte(r)...)
			} else {
				sb = append(sb, ' ')
			}
		}
		sb = append(sb, '\n')
	}
	return string(sb)
}
