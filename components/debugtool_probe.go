package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// cellProbeTool is an inline tool that reports the screen cell under the mouse
// cursor — coordinate, rune, fg/bg colors — plus the widget path at that point.
// It is passthrough: the app keeps receiving input, so you can drive the real
// UI and watch the probe update. It installs a mouse observer (via the App) to
// repaint on cursor motion.
type cellProbeTool struct {
	app *core.App
}

// NewCellProbeTool builds the cell-probe inline tool.
func NewCellProbeTool(app *core.App) DebugTool {
	return &cellProbeTool{app: app}
}

func (t *cellProbeTool) Name() string        { return "Probe" }
func (t *cellProbeTool) Kind() DebugToolKind { return DebugToolInline }

func (t *cellProbeTool) Hints() []KeyHint {
	return []KeyHint{{Key: "move mouse", Description: "Probe cell"}}
}

func (t *cellProbeTool) Activate() {
	// Request a redraw on every mouse move so the HUD tracks the cursor.
	t.app.SetMouseObserver(func(core.MouseAction, *tcell.EventMouse) bool { return true })
}

func (t *cellProbeTool) Deactivate() {
	t.app.SetMouseObserver(nil)
}

// HandleKey is a no-op; the probe is passthrough (return value ignored).
func (t *cellProbeTool) HandleKey(*tcell.EventKey) bool { return false }

func (t *cellProbeTool) Draw(screen tcell.Screen, x, y, w, h int) {
	mx, my := t.app.MousePosition()
	if mx < 0 || my < 0 || mx >= x+w || my >= y+h {
		return
	}

	str, style, _ := screen.Get(mx, my)
	fg, bg, _ := style.Decompose()
	glyph := str
	if str == " " || str == "" {
		glyph = "·"
	}

	lines := []string{
		fmt.Sprintf("%d,%d  '%s'", mx, my, glyph),
		fmt.Sprintf("fg %s  bg %s", colorHex(fg), colorHex(bg)),
		widgetPathAt(t.app.Root(), mx, my),
	}

	t.drawHUD(screen, mx, my, x, y, w, h, lines)
}

// drawHUD renders a small bordered box of lines, offset from the cursor and
// clamped to the available region so it never clips off-screen.
func (t *cellProbeTool) drawHUD(screen tcell.Screen, mx, my, ax, ay, aw, ah int, lines []string) {
	boxW := 0
	for _, l := range lines {
		if n := len([]rune(l)); n > boxW {
			boxW = n
		}
	}
	boxW += 2 // borders
	boxH := len(lines) + 2

	// Prefer below-right of the cursor; flip when it would overflow.
	bx := mx + 2
	by := my + 1
	if bx+boxW > ax+aw {
		bx = mx - boxW - 1
	}
	if by+boxH > ay+ah {
		by = my - boxH
	}
	if bx < ax {
		bx = ax
	}
	if by < ay {
		by = ay
	}

	bg := tcell.StyleDefault.Background(theme.BgDark()).Foreground(theme.Fg())
	core.FillRect(screen, bx, by, boxW, boxH, ' ', bg)
	core.DrawBorder(screen, bx, by, boxW, boxH,
		tcell.StyleDefault.Background(theme.BgDark()).Foreground(theme.Accent()))
	core.DrawTitle(screen, bx, by, boxW, " probe ", core.AlignLeft,
		tcell.StyleDefault.Background(theme.BgDark()).Foreground(theme.Accent()))
	for i, l := range lines {
		core.PrintClipped(screen, l, bx+1, by+1+i, boxW-2, bg)
	}
}

// widgetPathAt returns a root→leaf type path of widgets containing (x, y).
func widgetPathAt(root core.Widget, x, y int) string {
	if root == nil {
		return ""
	}
	c, ok := root.(core.Container)
	if !ok {
		return widgetTypeName(root)
	}
	desc := c.DescendantsAt(x, y) // deepest first
	if len(desc) == 0 {
		return widgetTypeName(root)
	}
	// Show the two deepest for context, leaf last.
	leaf := widgetTypeName(desc[0])
	if len(desc) > 1 {
		return widgetTypeName(desc[1]) + "/" + leaf
	}
	return leaf
}

// colorHex formats a tcell color as #rrggbb, or "default" if unset/non-RGB.
func colorHex(c tcell.Color) string {
	h := c.Hex()
	if h < 0 {
		return "default"
	}
	return fmt.Sprintf("#%06x", h)
}
