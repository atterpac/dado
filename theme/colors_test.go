package theme

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

var (
	black = tcell.NewRGBColor(0, 0, 0)
	white = tcell.NewRGBColor(255, 255, 255)
)

// Lerp returns the endpoints exactly and the midpoint near the average.
func TestLerp(t *testing.T) {
	if Lerp(black, white, 0) != black {
		t.Error("t=0 should return a")
	}
	if Lerp(black, white, 1) != white {
		t.Error("t=1 should return b")
	}
	// Out-of-range clamps.
	if Lerp(black, white, -1) != black || Lerp(black, white, 2) != white {
		t.Error("t out of [0,1] should clamp")
	}
	r, g, b := Lerp(black, white, 0.5).RGB()
	if r < 120 || r > 135 || g < 120 || g > 135 || b < 120 || b > 135 {
		t.Errorf("midpoint = (%d,%d,%d), want ~127 per channel", r, g, b)
	}
}

// Gradient walks the stops and degrades gracefully on edge cases.
func TestGradient(t *testing.T) {
	if Gradient(0.5) != Fg() {
		t.Error("no stops should fall back to Fg")
	}
	if Gradient(0.5, white) != white {
		t.Error("single stop should return that stop")
	}
	red := tcell.NewRGBColor(255, 0, 0)
	if Gradient(0, black, red, white) != black {
		t.Error("t=0 should return first stop")
	}
	if Gradient(1, black, red, white) != white {
		t.Error("t=1 should return last stop")
	}
	// Midpoint of a 3-stop gradient lands on the middle stop.
	if Gradient(0.5, black, red, white) != red {
		t.Error("t=0.5 of 3 stops should return middle stop")
	}
}
