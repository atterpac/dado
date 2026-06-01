package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// widgetTreeTool is a panel tool that walks the live widget tree from
// App.Root(), lets you navigate nodes (j/k or arrows), and highlights the
// selected node's bounds on the app area so you can see what each node covers.
type widgetTreeTool struct {
	app    *core.App
	sel    int
	offset int
	nodes  []treeNode // rebuilt each Draw from the live tree
}

type treeNode struct {
	w     core.Widget
	depth int
}

// NewWidgetTreeTool builds the widget-tree inspector tool.
func NewWidgetTreeTool(app *core.App) DebugTool {
	return &widgetTreeTool{app: app}
}

func (t *widgetTreeTool) Name() string        { return "Tree" }
func (t *widgetTreeTool) Kind() DebugToolKind { return DebugToolPanel }

func (t *widgetTreeTool) Hints() []KeyHint {
	return []KeyHint{
		{Key: "j/k ↑/↓", Description: "Select node"},
		{Key: "g/G", Description: "Top/bottom"},
	}
}

func (t *widgetTreeTool) Activate()   {}
func (t *widgetTreeTool) Deactivate() {}

func (t *widgetTreeTool) HandleKey(ev *tcell.EventKey) bool {
	n := len(t.nodes)
	if n == 0 {
		return true
	}
	switch {
	case ev.Key() == tcell.KeyDown, ev.Rune() == 'j':
		if t.sel < n-1 {
			t.sel++
		}
	case ev.Key() == tcell.KeyUp, ev.Rune() == 'k':
		if t.sel > 0 {
			t.sel--
		}
	case ev.Rune() == 'g':
		t.sel = 0
	case ev.Rune() == 'G':
		t.sel = n - 1
	}
	return true
}

func (t *widgetTreeTool) Draw(screen tcell.Screen, x, y, w, h int) {
	// Rebuild the flattened tree from the live root each frame.
	t.nodes = t.nodes[:0]
	if root := t.app.Root(); root != nil {
		t.flatten(root, 0)
	}
	if t.sel >= len(t.nodes) {
		t.sel = len(t.nodes) - 1
	}
	if t.sel < 0 {
		t.sel = 0
	}

	// Highlight the selected node's bounds on the app area (outside the panel).
	if t.sel < len(t.nodes) {
		t.highlight(screen, t.nodes[t.sel].w, x)
	}

	t.drawPanel(screen, x, y, w, h)
}

func (t *widgetTreeTool) flatten(w core.Widget, depth int) {
	t.nodes = append(t.nodes, treeNode{w: w, depth: depth})
	if c, ok := w.(core.Container); ok {
		for _, ch := range c.Children() {
			t.flatten(ch, depth+1)
		}
	}
}

// highlight outlines the widget's rect on screen, but only the portion left of
// the panel (panelX) so the outline marks the live app area, not the panel.
func (t *widgetTreeTool) highlight(screen tcell.Screen, w core.Widget, panelX int) {
	wx, wy, ww, wh := w.Rect()
	if ww <= 0 || wh <= 0 || wx >= panelX {
		return
	}
	style := tcell.StyleDefault.Foreground(theme.Accent()).Bold(true)
	right := wx + ww - 1
	if right >= panelX {
		right = panelX - 1
	}
	bottom := wy + wh - 1
	// Top & bottom edges.
	for cx := wx; cx <= right; cx++ {
		putRune(screen, cx, wy, '─', style)
		putRune(screen, cx, bottom, '─', style)
	}
	// Left & right edges.
	for cy := wy; cy <= bottom; cy++ {
		putRune(screen, wx, cy, '│', style)
		putRune(screen, right, cy, '│', style)
	}
	putRune(screen, wx, wy, '┌', style)
	putRune(screen, right, wy, '┐', style)
	putRune(screen, wx, bottom, '└', style)
	putRune(screen, right, bottom, '┘', style)
}

func putRune(screen tcell.Screen, x, y int, r rune, style tcell.Style) {
	if x < 0 || y < 0 {
		return
	}
	w, h := screen.Size()
	if x >= w || y >= h {
		return
	}
	screen.SetContent(x, y, r, nil, style)
}

func (t *widgetTreeTool) drawPanel(screen tcell.Screen, x, y, w, h int) {
	bg := tcell.StyleDefault.Background(theme.Bg()).Foreground(theme.Fg())
	core.FillRect(screen, x, y, w, h, ' ', bg)
	border := tcell.StyleDefault.Background(theme.Bg()).Foreground(theme.PanelBorder())
	core.DrawBorder(screen, x, y, w, h, border)
	core.DrawTitle(screen, x, y, w, " Widget Tree ", core.AlignLeft,
		tcell.StyleDefault.Background(theme.Bg()).Foreground(theme.PanelTitle()))

	innerX, innerY := x+1, y+1
	innerW, innerH := w-2, h-2
	if innerW <= 0 || innerH <= 0 {
		return
	}

	// Keep selection in view.
	if t.sel < t.offset {
		t.offset = t.sel
	}
	if t.sel >= t.offset+innerH {
		t.offset = t.sel - innerH + 1
	}

	sel := tcell.StyleDefault.Background(theme.SelectionBg()).Foreground(theme.SelectionFg())
	for row := 0; row < innerH; row++ {
		idx := t.offset + row
		if idx >= len(t.nodes) {
			break
		}
		node := t.nodes[idx]
		line := nodeLabel(node)
		style := bg
		if idx == t.sel {
			style = sel
			core.FillRect(screen, innerX, innerY+row, innerW, 1, ' ', sel)
		}
		core.PrintClipped(screen, line, innerX, innerY+row, innerW, style)
	}
}

func nodeLabel(n treeNode) string {
	indent := strings.Repeat("  ", n.depth)
	x, y, w, h := n.w.Rect()
	focus := " "
	if n.w.HasFocus() {
		focus = "●"
	}
	return fmt.Sprintf("%s%s %s %d,%d %dx%d", indent, focus, widgetTypeName(n.w), x, y, w, h)
}

// widgetTypeName returns the short type name, e.g. "*core.List" -> "List".
func widgetTypeName(w core.Widget) string {
	full := fmt.Sprintf("%T", w)
	full = strings.TrimPrefix(full, "*")
	if i := strings.LastIndex(full, "."); i >= 0 {
		return full[i+1:]
	}
	return full
}
