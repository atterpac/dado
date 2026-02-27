package components

import (
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InputHandler handles keyboard input for the ERD graph.
func (g *ERDGraph) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return g.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if g.data == nil {
			return
		}

		switch event.Key() {
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
			switch event.Rune() {
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
			}
		}
	})
}

// MouseHandler handles mouse input for the ERD graph.
func (g *ERDGraph) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return g.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		mx, my := event.Position()

		if !g.InRect(mx, my) {
			return false, nil
		}

		switch action {
		case tview.MouseLeftClick:
			setFocus(g)
			if g.data != nil {
				x, y, _, _ := g.GetInnerRect()
				for _, t := range g.data.tables {
					sx := x + t.x - g.offsetX
					sy := y + t.y - g.offsetY
					if mx >= sx && mx < sx+t.width && my >= sy && my < sy+t.height {
						g.data.focusID = t.ID
						g.centerOnFocused()
						return true, g
					}
				}
			}
			return true, g

		case tview.MouseLeftDoubleClick:
			if t := g.FocusedTable(); t != nil && g.onSelect != nil {
				g.onSelect(t)
			}
			return true, g

		case tview.MouseScrollUp:
			g.offsetY -= 2
			return true, g

		case tview.MouseScrollDown:
			g.offsetY += 2
			return true, g

		case tview.MouseScrollLeft:
			g.offsetX -= 4
			return true, g

		case tview.MouseScrollRight:
			g.offsetX += 4
			return true, g
		}

		return false, nil
	})
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
