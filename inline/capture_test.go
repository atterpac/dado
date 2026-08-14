package inline

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCaptureFrameBridgesTcellDrawing(t *testing.T) {
	t.Parallel()

	style := tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	frame, err := CaptureFrame(6, 2, func(screen tcell.Screen) {
		screen.PutStrStyled(1, 0, "界!", style)
		screen.ShowCursor(4, 1)
	})
	require.NoError(t, err)

	text, gotStyle, width := frame.Cell(1, 0)
	assert.Equal(t, "界", text)
	assert.Equal(t, style, gotStyle)
	assert.Equal(t, 2, width)
	text, gotStyle, width = frame.Cell(3, 0)
	assert.Equal(t, "!", text)
	assert.Equal(t, style, gotStyle)
	assert.Equal(t, 1, width)
	assert.Equal(t, &Cursor{X: 4, Y: 1}, frame.cursor)
}

func TestCaptureFrameRejectsNilDraw(t *testing.T) {
	t.Parallel()

	_, err := CaptureFrame(1, 1, nil)
	assert.ErrorIs(t, err, ErrNilDraw)
}
