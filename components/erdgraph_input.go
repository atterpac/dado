package components

import (
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
)

// HandleKey processes keyboard input for the ERD graph. Returns true when the
// event is consumed. Arrow keys pan the viewport; hjkl jump to the nearest
// node; Tab/Backtab cycle nodes; Enter opens the focused table; c re-centers.
func (g *ERDGraph) HandleKey(ev *tcell.EventKey) bool {
	if g.data == nil {
		return false
	}

	switch ev.Key() {
	case tcell.KeyUp:
		g.offsetY -= 2
	case tcell.KeyDown:
		g.offsetY += 2
	case tcell.KeyLeft:
		g.offsetX -= 4
	case tcell.KeyRight:
		g.offsetX += 4
	case tcell.KeyEnter:
		if t := g.FocusedTable(); t != nil && g.onSelect != nil {
			g.onSelect(t)
		}
	case tcell.KeyTab:
		g.cycleFocus(1)
	case tcell.KeyBacktab:
		g.cycleFocus(-1)
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'h':
			g.jumpNearest(-1, 0)
		case 'j':
			g.jumpNearest(0, 1)
		case 'k':
			g.jumpNearest(0, -1)
		case 'l':
			g.jumpNearest(1, 0)
		case 'c':
			g.centerOnFocused()
		default:
			return false
		}
	default:
		return false
	}
	return true
}

// HandleMouse processes mouse input for the ERD graph: click to focus a node,
// double-click to open it, and scroll to pan the viewport.
func (g *ERDGraph) HandleMouse(action core.MouseAction, ev *tcell.EventMouse) (bool, core.Widget) {
	mx, my := ev.Position()
	if !g.InRect(mx, my) {
		return false, nil
	}

	switch action {
	case core.MouseLeftClick:
		if g.data != nil {
			x, y, _, _ := g.GetInnerRect()
			for _, t := range g.data.tables {
				sx := x + t.x - g.offsetX
				sy := y + t.y - g.offsetY
				if mx >= sx && mx < sx+t.width && my >= sy && my < sy+t.height {
					g.data.focusID = t.ID
					g.centerOnFocused()
					return true, nil
				}
			}
		}
		return true, nil

	case core.MouseLeftDoubleClick:
		if t := g.FocusedTable(); t != nil && g.onSelect != nil {
			g.onSelect(t)
		}
		return true, nil

	case core.MouseScrollUp:
		g.offsetY -= 2
		return true, nil

	case core.MouseScrollDown:
		g.offsetY += 2
		return true, nil

	case core.MouseScrollLeft:
		g.offsetX -= 4
		return true, nil

	case core.MouseScrollRight:
		g.offsetX += 4
		return true, nil
	}

	return false, nil
}

// cycleFocus moves focus to the next or previous table in tableOrder.
func (g *ERDGraph) cycleFocus(dir int) {
	if g.data == nil || len(g.data.tableOrder) == 0 {
		return
	}

	idx := 0
	for i, id := range g.data.tableOrder {
		if id == g.data.focusID {
			idx = i
			break
		}
	}

	idx += dir
	n := len(g.data.tableOrder)
	idx = ((idx % n) + n) % n

	g.data.focusID = g.data.tableOrder[idx]
	g.centerOnFocused()
}

// jumpNearest jumps focus to the nearest node in the given direction.
// dx, dy indicate the direction: (-1,0)=left, (1,0)=right, (0,-1)=up, (0,1)=down.
func (g *ERDGraph) jumpNearest(dx, dy int) {
	if g.data == nil || g.data.focusID == "" {
		return
	}

	current := g.data.tables[g.data.focusID]
	if current == nil {
		return
	}

	cx := float64(current.x + current.width/2)
	cy := float64(current.y + current.height/2)

	bestID := ""
	bestDist := math.MaxFloat64

	for _, id := range g.data.tableOrder {
		if id == g.data.focusID {
			continue
		}
		t := g.data.tables[id]
		if t == nil {
			continue
		}

		tx := float64(t.x + t.width/2)
		ty := float64(t.y + t.height/2)

		// Check direction: the candidate must be in the requested direction.
		diffX := tx - cx
		diffY := ty - cy

		if dx != 0 {
			if dx > 0 && diffX <= 0 {
				continue
			}
			if dx < 0 && diffX >= 0 {
				continue
			}
		}
		if dy != 0 {
			if dy > 0 && diffY <= 0 {
				continue
			}
			if dy < 0 && diffY >= 0 {
				continue
			}
		}

		// Primary axis distance + perpendicular as tiebreaker.
		var dist float64
		if dx != 0 {
			dist = math.Abs(diffX) + math.Abs(diffY)*0.5
		} else {
			dist = math.Abs(diffY) + math.Abs(diffX)*0.5
		}

		if dist < bestDist {
			bestDist = dist
			bestID = id
		}
	}

	if bestID != "" {
		g.data.focusID = bestID
		g.centerOnFocused()
	}
}
