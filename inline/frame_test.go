package inline

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrameDrawStringUsesGraphemeWidths(t *testing.T) {
	t.Parallel()

	frame := NewFrame(8, 2)
	style := tcell.StyleDefault.Foreground(tcell.ColorAqua)
	frame.DrawString(0, 0, "A界B\ne\u0301", style)

	text, gotStyle, width := frame.Cell(0, 0)
	assert.Equal(t, "A", text)
	assert.Equal(t, style, gotStyle)
	assert.Equal(t, 1, width)

	text, gotStyle, width = frame.Cell(1, 0)
	assert.Equal(t, "界", text)
	assert.Equal(t, style, gotStyle)
	assert.Equal(t, 2, width)

	text, gotStyle, width = frame.Cell(3, 0)
	assert.Equal(t, "B", text)
	assert.Equal(t, style, gotStyle)
	assert.Equal(t, 1, width)

	text, gotStyle, width = frame.Cell(0, 1)
	assert.Equal(t, "e\u0301", text)
	assert.Equal(t, style, gotStyle)
	assert.Equal(t, 1, width)
}

func TestFrameResizeAndCursorBounds(t *testing.T) {
	t.Parallel()

	frame := NewFrame(4, 2)
	frame.DrawString(0, 0, "test", tcell.StyleDefault)
	frame.ShowCursor(3, 1)
	frame.Resize(2, 1)

	assert.Equal(t, 2, frame.Width())
	assert.Equal(t, 1, frame.Height())
	assert.Nil(t, frame.cursor)
	text, _, _ := frame.Cell(0, 0)
	assert.Equal(t, "t", text)
	text, _, _ = frame.Cell(1, 0)
	assert.Equal(t, "e", text)
}

func TestFrameNegativeDimensionsAreClamped(t *testing.T) {
	t.Parallel()

	frame := NewFrame(-2, -1)
	w, h := frame.Size()
	require.Zero(t, w)
	require.Zero(t, h)
}
