package inline

import (
	"errors"

	"github.com/gdamore/tcell/v2"
)

// ErrNilDraw is returned when CaptureFrame is called without a draw function.
var ErrNilDraw = errors.New("inline: nil draw function")

// CaptureFrame draws through a tcell.Screen into an offscreen Frame. It is a
// compatibility bridge for existing Dado widgets whose Draw method accepts a
// tcell.Screen. No real terminal screen is initialized or modified.
//
// The draw callback should treat the supplied screen as width by height and
// should not retain it after returning.
func CaptureFrame(width, height int, draw func(tcell.Screen)) (*Frame, error) {
	if draw == nil {
		return nil, ErrNilDraw
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		return nil, err
	}
	defer screen.Fini()
	screen.SetSize(width, height)
	screen.Clear()
	draw(screen)
	screen.Show()

	contents, actualWidth, actualHeight := screen.GetContents()
	frame := NewFrame(actualWidth, actualHeight)
	for y := range actualHeight {
		for x := range actualWidth {
			cell := contents[y*actualWidth+x]
			if len(cell.Runes) == 0 {
				continue
			}
			frame.Put(x, y, string(cell.Runes), cell.Style)
		}
	}
	if x, y, visible := screen.GetCursor(); visible {
		frame.ShowCursor(x, y)
	}
	return frame, nil
}
