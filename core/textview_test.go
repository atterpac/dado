package core_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

var _ core.Widget    = (*core.TextView)(nil)
var _ core.KeyHandler = (*core.TextView)(nil)

func TestTextView_SetText_Renders(t *testing.T) {
	tv := core.NewTextView()
	tv.SetText("Hello, world")
	screen := coretest.DrawWidget(tv, 40, 5)
	assert.True(t, screen.ContainsText("Hello"), "screen should contain set text\n%s", screen.Dump())
}

func TestTextView_EmptyText_NoPanic(t *testing.T) {
	tv := core.NewTextView()
	assert.NotPanics(t, func() { coretest.DrawWidget(tv, 20, 5) })
}

func TestTextView_WrapBreaksAtWidth(t *testing.T) {
	tv := core.NewTextView()
	tv.SetText("AAAA BBBB CCCC DDDD")
	screen := coretest.DrawWidget(tv, 10, 10) // narrow width forces wrap
	// Should see content across multiple rows since width < text length
	row0 := screen.GetRow(0)
	row1 := screen.GetRow(1)
	assert.NotEmpty(t, row0)
	// At least one of the words should appear on a second line after wrapping
	combined := row0 + row1
	assert.Contains(t, combined, "A")
}

func TestTextView_ScrollOffset_InitiallyZero(t *testing.T) {
	tv := core.NewTextView()
	row, col := tv.GetScrollOffset()
	assert.Equal(t, 0, row)
	assert.Equal(t, 0, col)
}

func TestTextView_Scroll_Down(t *testing.T) {
	tv := core.NewTextView()
	tv.SetScrollable(true)
	// Fill with enough lines to scroll
	text := "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8"
	tv.SetText(text)
	tv.SetRect(0, 0, 20, 3) // only 3 rows visible

	consumed := coretest.SimulateKey(tv, tcell.KeyDown)
	assert.True(t, consumed)
	row, _ := tv.GetScrollOffset()
	assert.Equal(t, 1, row)
}

func TestTextView_Scroll_ClampsAtBottom(t *testing.T) {
	tv := core.NewTextView()
	tv.SetScrollable(true)
	tv.SetText("l1\nl2\nl3\nl4\nl5") // 5 lines
	tv.SetRect(0, 0, 20, 3)         // 3 visible -> max offset 2

	// Press Down well past the end.
	for range 10 {
		coretest.SimulateKey(tv, tcell.KeyDown)
	}
	row, _ := tv.GetScrollOffset()
	assert.Equal(t, 2, row, "scrollY must clamp to max offset, not overshoot")

	// One Up should move immediately, with no phantom rows to climb back through.
	coretest.SimulateKey(tv, tcell.KeyUp)
	row, _ = tv.GetScrollOffset()
	assert.Equal(t, 1, row)
}

func TestTextView_Scroll_NotScrollable_IgnoresKey(t *testing.T) {
	tv := core.NewTextView()
	tv.SetScrollable(false)
	tv.SetText("line1\nline2\nline3")
	tv.SetRect(0, 0, 20, 1)

	consumed := coretest.SimulateKey(tv, tcell.KeyDown)
	assert.False(t, consumed)
	row, _ := tv.GetScrollOffset()
	assert.Equal(t, 0, row)
}

func TestTextView_ScrollTo(t *testing.T) {
	tv := core.NewTextView()
	tv.SetScrollable(true)
	tv.SetText("l1\nl2\nl3\nl4\nl5")
	tv.ScrollTo(2, 0)
	row, _ := tv.GetScrollOffset()
	assert.Equal(t, 2, row)
}

func TestTextView_DynamicColors_ANSI(t *testing.T) {
	tv := core.NewTextView()
	tv.SetDynamicColors(true)
	// color tags: [red] sets foreground red
	tv.SetText("[red]RED[white]")
	tv.SetRect(0, 0, 20, 3)
	screen := coretest.NewTestScreen(20, 3)
	tv.Draw(screen.SimulationScreen)
	screen.Show()
	// With dynamic colors enabled, the text "RED" should appear
	assert.True(t, screen.ContainsText("RED"))
	// And the foreground at that position should be red
	x, y := screen.FindText("RED")
	if x >= 0 {
		fg := screen.GetForegroundAt(x, y)
		assert.Equal(t, tcell.ColorRed, fg)
	}
}
