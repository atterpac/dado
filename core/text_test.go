package core_test

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/core/coretest"
)

func TestText_AppendWidthAndString(t *testing.T) {
	txt := core.NewText().
		Append("foo", tcell.StyleDefault).
		Append("", tcell.StyleDefault). // ignored
		Append(" bar", tcell.StyleDefault)
	assert.Equal(t, 7, txt.Width())
	assert.Equal(t, "foo bar", txt.String())
	assert.Len(t, txt.Spans(), 2)
}

func TestText_DrawClipsToMaxWidth(t *testing.T) {
	screen := coretest.NewTestScreen(20, 1)
	txt := core.NewText().Append("hello", tcell.StyleDefault)
	n := txt.Draw(screen.SimulationScreen, 0, 0, 3)
	screen.SimulationScreen.Show()
	assert.Equal(t, 3, n)
	assert.Equal(t, "hel", strings.TrimRight(screen.GetRow(0), " "))
}

func TestText_DrawPreservesPerSpanStyle(t *testing.T) {
	screen := coretest.NewTestScreen(20, 1)
	red := tcell.StyleDefault.Foreground(tcell.ColorRed)
	blue := tcell.StyleDefault.Foreground(tcell.ColorBlue)
	core.NewText().Append("ab", red).Append("cd", blue).Draw(screen.SimulationScreen, 0, 0, 20)
	screen.SimulationScreen.Show()
	assert.Equal(t, tcell.ColorRed, screen.GetForegroundAt(0, 0))
	assert.Equal(t, tcell.ColorBlue, screen.GetForegroundAt(2, 0))
}

func TestText_DrawFuncOverlayPreservesSpanColors(t *testing.T) {
	screen := coretest.NewTestScreen(20, 1)
	red := tcell.StyleDefault.Foreground(tcell.ColorRed)
	txt := core.NewText().Append("ab", red).Append("cd", tcell.StyleDefault)
	// Additive overlay: reverse on top of each span's own style.
	txt.DrawFunc(screen.SimulationScreen, 0, 0, 20, func(s tcell.Style) tcell.Style {
		return s.Reverse(true)
	})
	screen.SimulationScreen.Show()
	// Span color preserved...
	assert.Equal(t, tcell.ColorRed, screen.GetForegroundAt(0, 0))
	// ...and reverse applied to both spans.
	_, _, attr0 := screen.GetStyleAt(0, 0).Decompose()
	_, _, attr2 := screen.GetStyleAt(2, 0).Decompose()
	assert.NotZero(t, attr0&tcell.AttrReverse)
	assert.NotZero(t, attr2&tcell.AttrReverse)
}

func TestText_DrawFuncNilOverlayEqualsDraw(t *testing.T) {
	screen := coretest.NewTestScreen(20, 1)
	core.NewText().Append("hi", tcell.StyleDefault).DrawFunc(screen.SimulationScreen, 0, 0, 20, nil)
	screen.SimulationScreen.Show()
	assert.Equal(t, "hi", strings.TrimRight(screen.GetRow(0), " "))
}

func TestParseTagged_CompilesSpans(t *testing.T) {
	txt := core.ParseTagged("[#ff0000]err[-] ok", tcell.StyleDefault)
	assert.Equal(t, "err ok", txt.String())
	assert.Equal(t, 6, txt.Width())

	screen := coretest.NewTestScreen(20, 1)
	txt.Draw(screen.SimulationScreen, 0, 0, 20)
	screen.SimulationScreen.Show()
	assert.Equal(t, tcell.NewHexColor(0xff0000), screen.GetForegroundAt(0, 0))
	// after [-] reset, the space/ok use default fg
	assert.Equal(t, tcell.ColorDefault, screen.GetForegroundAt(4, 0))
}

func TestText_WrapBreaksAtSpaces(t *testing.T) {
	txt := core.NewText().Append("aaaa bbbb cccc", tcell.StyleDefault)
	lines := txt.Wrap(9)
	assert.Equal(t, []string{"aaaa bbbb", "cccc"}, lineStrings(lines))
	for _, l := range lines {
		assert.LessOrEqual(t, l.Width(), 9)
	}
}

func TestText_WrapHardBreaksLongWord(t *testing.T) {
	txt := core.NewText().Append("abcdefghij", tcell.StyleDefault)
	lines := txt.Wrap(4)
	assert.Equal(t, []string{"abcd", "efgh", "ij"}, lineStrings(lines))
}

func TestText_WrapHonorsNewlines(t *testing.T) {
	txt := core.NewText().Append("ab\ncd", tcell.StyleDefault)
	lines := txt.Wrap(10)
	assert.Equal(t, []string{"ab", "cd"}, lineStrings(lines))
}

func TestText_WrapNoOpWhenFits(t *testing.T) {
	txt := core.NewText().Append("short", tcell.StyleDefault)
	lines := txt.Wrap(80)
	assert.Len(t, lines, 1)
	assert.Same(t, txt, lines[0])
}

func TestText_WrapPreservesStyleAcrossBreak(t *testing.T) {
	red := tcell.StyleDefault.Foreground(tcell.ColorRed)
	txt := core.NewText().Append("aa ", tcell.StyleDefault).Append("bbbb", red)
	lines := txt.Wrap(4)
	// "aa" on line 0, "bbbb" (red) on line 1
	assert.Equal(t, []string{"aa", "bbbb"}, lineStrings(lines))
	screen := coretest.NewTestScreen(10, 2)
	lines[1].Draw(screen.SimulationScreen, 0, 0, 10)
	screen.SimulationScreen.Show()
	assert.Equal(t, tcell.ColorRed, screen.GetForegroundAt(0, 0))
}

func lineStrings(lines []*core.Text) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = l.String()
	}
	return out
}
