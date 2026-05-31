package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Draw renders the ERD graph.
func (g *ERDGraph) Draw(screen tcell.Screen) {
	g.Box.DrawForSubclass(screen, g)
	x, y, width, height := g.GetInnerRect()

	if width <= 0 || height <= 0 || g.data == nil || len(g.data.tables) == 0 {
		return
	}

	// Auto-center on first draw after data is set.
	if g.needsCenter {
		g.needsCenter = false
		if g.fitAll {
			g.centerOnAll()
		} else {
			g.centerOnFocused()
		}
	}

	// Precompute the set of tables connected to the focused table.
	activeNeighbors := make(map[string]bool)
	focusID := g.data.focusID
	for _, rel := range g.data.relations {
		if rel.FromTable == focusID {
			activeNeighbors[rel.ToTable] = true
		} else if rel.ToTable == focusID {
			activeNeighbors[rel.FromTable] = true
		}
	}

	th := g.th()
	bgColor := th.Bg()

	// Clear the area.
	bgStyle := tcell.StyleDefault.Background(bgColor)
	fillRect(screen, x, y, width, height, bgStyle)

	// Draw edges behind nodes. Inactive (dotted) first, then active (solid)
	// so solid lines win at overlapping cells.
	for _, rel := range g.data.relations {
		if rel.FromTable != focusID && rel.ToTable != focusID {
			g.drawEdge(screen, x, y, width, height, rel, false)
		}
	}
	for _, rel := range g.data.relations {
		if rel.FromTable == focusID || rel.ToTable == focusID {
			g.drawEdge(screen, x, y, width, height, rel, true)
		}
	}

	// Draw nodes.
	for _, id := range g.data.tableOrder {
		t := g.data.tables[id]
		if t == nil {
			continue
		}
		g.drawNode(screen, x, y, width, height, t, activeNeighbors[id])
	}

	// Draw FK info panel for the focused table (rendered last, on top).
	g.drawFKInfoPanel(screen, x, y, width, height)
}

// setIfVisible sets a cell only if it falls within the viewport.
func setIfVisible(screen tcell.Screen, sx, sy, vx, vy, vw, vh int, ch rune, style tcell.Style) {
	if sx >= vx && sx < vx+vw && sy >= vy && sy < vy+vh {
		screen.SetContent(sx, sy, ch, nil, style)
	}
}

// drawNode renders a single table node with header, separator, and columns.
// isNeighbor is true when this table is directly connected to the focused table via FK.
func (g *ERDGraph) drawNode(screen tcell.Screen, vx, vy, vw, vh int, t *ERDTable, isNeighbor bool) {
	th := g.th()
	bgColor := th.Bg()
	fgColor := th.Fg()

	isFocused := g.data.focusID == t.ID
	borderColor := th.Border()
	if isFocused {
		borderColor = th.BorderFocus()
	} else if isNeighbor {
		borderColor = th.Info()
	}

	sx := vx + t.x - g.offsetX
	sy := vy + t.y - g.offsetY

	borderStyle := tcell.StyleDefault.Background(bgColor).Foreground(borderColor)

	// Top border: ╭───────────────────╮
	setIfVisible(screen, sx, sy, vx, vy, vw, vh, '╭', borderStyle)
	for c := 1; c < t.width-1; c++ {
		setIfVisible(screen, sx+c, sy, vx, vy, vw, vh, '─', borderStyle)
	}
	setIfVisible(screen, sx+t.width-1, sy, vx, vy, vw, vh, '╮', borderStyle)

	// Header row (row 1): │   tablename   │
	headerY := sy + 1
	accentColor := th.Accent()
	headerBg := accentColor
	headerFg := bgColor
	if !isFocused {
		if isNeighbor {
			headerBg = th.Info()
			headerFg = bgColor
		} else {
			headerBg = th.BgLight()
			headerFg = fgColor
		}
	}
	headerStyle := tcell.StyleDefault.Background(headerBg).Foreground(headerFg).Bold(true)

	setIfVisible(screen, sx, headerY, vx, vy, vw, vh, '│', borderStyle)
	nameRunes := []rune(t.Name)
	innerW := t.width - 2
	padding := (innerW - len(nameRunes)) / 2
	if padding < 0 {
		padding = 0
	}
	for c := 0; c < innerW; c++ {
		ch := ' '
		nameIdx := c - padding
		if nameIdx >= 0 && nameIdx < len(nameRunes) {
			ch = nameRunes[nameIdx]
		}
		setIfVisible(screen, sx+1+c, headerY, vx, vy, vw, vh, ch, headerStyle)
	}
	setIfVisible(screen, sx+t.width-1, headerY, vx, vy, vw, vh, '│', borderStyle)

	// Separator row: ├───────────────────┤
	sepY := sy + 2
	setIfVisible(screen, sx, sepY, vx, vy, vw, vh, '├', borderStyle)
	for c := 1; c < t.width-1; c++ {
		setIfVisible(screen, sx+c, sepY, vx, vy, vw, vh, '─', borderStyle)
	}
	setIfVisible(screen, sx+t.width-1, sepY, vx, vy, vw, vh, '┤', borderStyle)

	// Column rows
	pkColor := th.Warning()
	fkColor := th.Info()
	dimColor := th.FgDim()

	for i, col := range t.Columns {
		rowY := sy + 3 + i

		setIfVisible(screen, sx, rowY, vx, vy, vw, vh, '│', borderStyle)

		// Determine icon
		icon := ' '
		iconColor := fgColor
		if col.IsPK {
			icon = 'K'
			iconColor = pkColor
		} else if col.IsFK {
			icon = '→'
			iconColor = fkColor
		}

		// "│ K  name         type │"
		colStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		iconStyle := tcell.StyleDefault.Background(bgColor).Foreground(iconColor)
		typeStyle := tcell.StyleDefault.Background(bgColor).Foreground(dimColor)

		// Icon (column 1)
		setIfVisible(screen, sx+1, rowY, vx, vy, vw, vh, icon, iconStyle)
		// Space after icon
		setIfVisible(screen, sx+2, rowY, vx, vy, vw, vh, ' ', colStyle)

		// Column name
		nameR := []rune(col.Name)
		typeR := []rune(col.Type)
		// Available space: innerW - 3 (icon + space + trailing space)
		nameEnd := sx + 3
		for _, r := range nameR {
			setIfVisible(screen, nameEnd, rowY, vx, vy, vw, vh, r, colStyle)
			nameEnd++
		}

		// Type (right-aligned within remaining space)
		typeStart := sx + t.width - 1 - len(typeR) - 1
		if typeStart <= nameEnd {
			typeStart = nameEnd + 1
		}
		// Fill gap with spaces
		for c := nameEnd; c < typeStart; c++ {
			setIfVisible(screen, c, rowY, vx, vy, vw, vh, ' ', colStyle)
		}
		for _, r := range typeR {
			setIfVisible(screen, typeStart, rowY, vx, vy, vw, vh, r, typeStyle)
			typeStart++
		}
		// Trailing space before border
		for c := typeStart; c < sx+t.width-1; c++ {
			setIfVisible(screen, c, rowY, vx, vy, vw, vh, ' ', colStyle)
		}

		setIfVisible(screen, sx+t.width-1, rowY, vx, vy, vw, vh, '│', borderStyle)
	}

	// Bottom border: ╰───────────────────╯
	botY := sy + t.height - 1
	setIfVisible(screen, sx, botY, vx, vy, vw, vh, '╰', borderStyle)
	for c := 1; c < t.width-1; c++ {
		setIfVisible(screen, sx+c, botY, vx, vy, vw, vh, '─', borderStyle)
	}
	setIfVisible(screen, sx+t.width-1, botY, vx, vy, vw, vh, '╯', borderStyle)
}

// drawEdge renders a relationship line between two table nodes.
// isActive is true when this edge connects to the currently focused table.
func (g *ERDGraph) drawEdge(screen tcell.Screen, vx, vy, vw, vh int, rel *ERDRelation, isActive bool) {
	from := g.data.tables[rel.FromTable]
	to := g.data.tables[rel.ToTable]
	if from == nil || to == nil {
		return
	}
	// Skip self-referential edges.
	if rel.FromTable == rel.ToTable {
		return
	}

	th := g.th()
	bgColor := th.Bg()
	var edgeColor tcell.Color
	var lineH, lineV rune

	if isActive {
		// Active edges: solid lines, accent color.
		lineH = '─'
		lineV = '│'
		edgeColor = th.Accent()
	} else {
		// Inactive edges: dotted lines, dim color.
		lineH = '·'
		lineV = '·'
		edgeColor = th.FgDim()
	}

	edgeStyle := tcell.StyleDefault.Background(bgColor).Foreground(edgeColor)

	// Find the source and target column rows for precise connection points.
	fromRow := g.columnRow(from, rel.FromColumn)
	toRow := g.columnRow(to, rel.ToColumn)

	// Connection points: exit from side of source node, enter side of target node.
	// Determine whether to connect left or right side based on relative positions.
	var fromX, fromY, toX, toY int

	if from.x+from.width <= to.x {
		// Source is to the left of target.
		fromX = vx + from.x + from.width - g.offsetX
		toX = vx + to.x - 1 - g.offsetX
	} else if to.x+to.width <= from.x {
		// Source is to the right of target.
		fromX = vx + from.x - 1 - g.offsetX
		toX = vx + to.x + to.width - g.offsetX
	} else {
		// Overlapping horizontally — connect via bottom/top.
		fromX = vx + from.x + from.width/2 - g.offsetX
		toX = vx + to.x + to.width/2 - g.offsetX
	}

	fromY = vy + from.y + fromRow - g.offsetY
	toY = vy + to.y + toRow - g.offsetY

	// Cardinality labels placed 1 cell outward from the node edges.
	fromCardLabel := "1"
	toCardLabel := "1"
	if rel.Cardinality == OneToMany {
		toCardLabel = "*"
	} else if rel.Cardinality == ManyToMany {
		fromCardLabel = "*"
		toCardLabel = "*"
	}

	if fromY == toY {
		// Same row — straight horizontal line.
		startX, endX := fromX, toX
		if startX > endX {
			startX, endX = endX, startX
		}
		for x := startX; x <= endX; x++ {
			setIfVisible(screen, x, fromY, vx, vy, vw, vh, lineH, edgeStyle)
		}
		// Cardinality at endpoints.
		g.drawCardinalityLabel(screen, vx, vy, vw, vh, fromX, fromY, fromCardLabel, edgeStyle)
		g.drawCardinalityLabel(screen, vx, vy, vw, vh, toX, toY, toCardLabel, edgeStyle)
		return
	}

	// Different rows — route with two bends:
	// horizontal from source → corner → vertical → corner → horizontal to target.
	midX := (fromX + toX) / 2

	// Segment 1: horizontal from fromX to midX at fromY.
	hStart, hEnd := fromX, midX
	if hStart > hEnd {
		hStart, hEnd = hEnd, hStart
	}
	for x := hStart; x <= hEnd; x++ {
		setIfVisible(screen, x, fromY, vx, vy, vw, vh, lineH, edgeStyle)
	}

	// Segment 2: vertical from fromY to toY at midX (exclusive of corners).
	vStart, vEnd := fromY, toY
	if vStart > vEnd {
		vStart, vEnd = vEnd, vStart
	}
	for y := vStart + 1; y < vEnd; y++ {
		setIfVisible(screen, midX, y, vx, vy, vw, vh, lineV, edgeStyle)
	}

	// Segment 3: horizontal from midX to toX at toY.
	hStart, hEnd = midX, toX
	if hStart > hEnd {
		hStart, hEnd = hEnd, hStart
	}
	for x := hStart; x <= hEnd; x++ {
		setIfVisible(screen, x, toY, vx, vy, vw, vh, lineH, edgeStyle)
	}

	// Corner glyphs at the two bend points (rounded).
	// First corner at (midX, fromY): coming from horizontal, turning vertical.
	// Second corner at (midX, toY): coming from vertical, turning horizontal.
	goingRight := midX > fromX
	goingDown := toY > fromY
	exitRight := toX > midX

	var corner1, corner2 rune
	if isActive {
		// Rounded corners for solid active lines.
		if goingRight && goingDown {
			corner1 = '╮'
		} else if goingRight && !goingDown {
			corner1 = '╯'
		} else if !goingRight && goingDown {
			corner1 = '╭'
		} else {
			corner1 = '╰'
		}

		if exitRight && goingDown {
			corner2 = '╰'
		} else if exitRight && !goingDown {
			corner2 = '╭'
		} else if !exitRight && goingDown {
			corner2 = '╯'
		} else {
			corner2 = '╮'
		}
	} else {
		// Dotted lines use the dot character for corners too.
		corner1 = lineH
		corner2 = lineH
	}

	setIfVisible(screen, midX, fromY, vx, vy, vw, vh, corner1, edgeStyle)
	setIfVisible(screen, midX, toY, vx, vy, vw, vh, corner2, edgeStyle)

	// Cardinality at endpoints.
	g.drawCardinalityLabel(screen, vx, vy, vw, vh, fromX, fromY, fromCardLabel, edgeStyle)
	g.drawCardinalityLabel(screen, vx, vy, vw, vh, toX, toY, toCardLabel, edgeStyle)
}

// columnRow returns the screen row offset within the table for the given column name.
// Returns the header row offset if the column is not found.
func (g *ERDGraph) columnRow(t *ERDTable, colName string) int {
	for i, col := range t.Columns {
		if col.Name == colName {
			return 3 + i // top border + header + separator + index
		}
	}
	return 1 // default to header
}

// drawCardinalityLabel draws a cardinality label ("1" or "*") 1 cell away from the endpoint.
func (g *ERDGraph) drawCardinalityLabel(screen tcell.Screen, vx, vy, vw, vh, ex, ey int, label string, style tcell.Style) {
	// Place label 1 cell above the endpoint.
	ly := ey - 1
	if ly < vy {
		ly = ey + 1
	}
	for i, r := range label {
		setIfVisible(screen, ex+i, ly, vx, vy, vw, vh, r, style)
	}
}

// drawFKInfoPanel renders a floating info box anchored to the bottom-right
// corner of the viewport. It lists all FK relationships for the focused table,
// showing column → target.column with a directional hint when the target
// table is off-screen.
func (g *ERDGraph) drawFKInfoPanel(screen tcell.Screen, vx, vy, vw, vh int) {
	if g.data == nil || g.data.focusID == "" {
		return
	}

	focusID := g.data.focusID
	focused := g.data.tables[focusID]
	if focused == nil {
		return
	}

	// Collect FK lines: outbound and inbound relations involving the focused table.
	type fkLine struct {
		text  string
		color tcell.Color
	}
	var lines []fkLine

	th := g.th()
	accentColor := th.Accent()
	infoColor := th.Info()

	for _, rel := range g.data.relations {
		if rel.FromTable == focusID {
			target := g.data.tables[rel.ToTable]
			dirHint := g.offscreenHint(target, vx, vy, vw, vh)
			text := fmt.Sprintf(" → %s.%s", rel.ToTable, rel.ToColumn)
			if dirHint != "" {
				text += " " + dirHint
			}
			lines = append(lines, fkLine{
				text:  fmt.Sprintf(" %s%s", rel.FromColumn, text),
				color: accentColor,
			})
		} else if rel.ToTable == focusID {
			source := g.data.tables[rel.FromTable]
			dirHint := g.offscreenHint(source, vx, vy, vw, vh)
			text := fmt.Sprintf(" ← %s.%s", rel.FromTable, rel.FromColumn)
			if dirHint != "" {
				text += " " + dirHint
			}
			lines = append(lines, fkLine{
				text:  fmt.Sprintf(" %s%s", rel.ToColumn, text),
				color: infoColor,
			})
		}
	}

	if len(lines) == 0 {
		return
	}

	// Add header.
	headerLine := fmt.Sprintf(" %s — Foreign Keys ", focused.Name)
	allLines := make([]fkLine, 0, len(lines)+1)
	allLines = append(allLines, fkLine{text: headerLine, color: th.Fg()})
	allLines = append(allLines, lines...)

	// Compute panel dimensions.
	panelW := 0
	for _, l := range allLines {
		rw := len([]rune(l.text)) + 1 // +1 trailing space
		if rw > panelW {
			panelW = rw
		}
	}
	panelH := len(allLines)

	// Position: bottom-right of the viewport, 1 cell margin.
	px := vx + vw - panelW - 1
	py := vy + vh - panelH - 1
	if px < vx {
		px = vx
	}
	if py < vy {
		py = vy
	}

	bgColor := th.BgLight()
	borderColor := th.Border()
	bgStyle := tcell.StyleDefault.Background(bgColor).Foreground(borderColor)

	// Draw panel background and text.
	for row, l := range allLines {
		sy := py + row
		style := tcell.StyleDefault.Background(bgColor).Foreground(l.color)
		if row == 0 {
			style = style.Bold(true)
		}
		runes := []rune(l.text)
		for col := 0; col < panelW; col++ {
			ch := ' '
			if col < len(runes) {
				ch = runes[col]
			}
			setIfVisible(screen, px+col, sy, vx, vy, vw, vh, ch, style)
		}
		// Thin right border.
		setIfVisible(screen, px+panelW, sy, vx, vy, vw, vh, '│', bgStyle)
	}

	// Bottom border line.
	for col := 0; col <= panelW; col++ {
		setIfVisible(screen, px+col, py+panelH, vx, vy, vw, vh, '─', bgStyle)
	}
}

// offscreenHint returns a directional arrow string if the given table's node
// is outside the current viewport, or "" if it's visible.
func (g *ERDGraph) offscreenHint(t *ERDTable, vx, vy, vw, vh int) string {
	if t == nil {
		return "[?]"
	}

	sx := vx + t.x - g.offsetX
	sy := vy + t.y - g.offsetY

	visible := sx+t.width > vx && sx < vx+vw && sy+t.height > vy && sy < vy+vh
	if visible {
		return ""
	}

	// Build a directional hint based on where the table is relative to viewport center.
	centerX := vx + vw/2
	centerY := vy + vh/2
	nodeCX := sx + t.width/2
	nodeCY := sy + t.height/2

	dx := nodeCX - centerX
	dy := nodeCY - centerY

	arrow := ""
	if dy < 0 {
		arrow += "↑"
	} else if dy > 0 {
		arrow += "↓"
	}
	if dx < 0 {
		arrow += "←"
	} else if dx > 0 {
		arrow += "→"
	}
	return arrow
}
