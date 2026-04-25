package components

import (
	"github.com/atterpac/jig/theme"
	"github.com/gdamore/tcell/v2"
)

// drawSnapshot holds the geometry and theme colors computed during the
// state-mutation phase of Draw. The paint phase reads only this snapshot
// (and mutation-free fields like selectedRows / cursor / source) so the
// boundary between "compute layout / advance state" and "write to screen"
// is explicit.
type drawSnapshot struct {
	headerHeight int
	dataHeight   int
	gutterWidth  int
	contentWidth int
	hasVScroll   bool
	startRow     int
	endRow       int
	cols         []GridColumn
	bg, fg       tcell.Color
	fgDim        tcell.Color
	accent       tcell.Color
	warning      tcell.Color
	header       tcell.Color
}

// Draw renders the DataGrid to the screen. State mutation lives in
// prepareDrawLocked; screen writes live in paintLocked. Both run while
// dg.mu is held — concurrency semantics are unchanged from the prior
// monolithic implementation.
func (dg *DataGrid) Draw(screen tcell.Screen) {
	dg.Box.DrawForSubclass(screen, dg)
	x, y, width, height := dg.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	dg.mu.Lock()
	defer dg.mu.Unlock()

	if dg.source == nil || dg.source.RowCount() == 0 {
		return
	}

	snap, ok := dg.prepareDrawLocked(width, height)
	if !ok {
		return
	}
	dg.paintLocked(screen, x, y, width, height, snap)
}

// prepareDrawLocked performs all dg.* mutations Draw needs (viewport
// fit, column-width recompute, source prefetch, cursor clamp) and
// returns the geometry the paint phase will consume. Returns ok=false
// when the available area is too small to render anything.
//
// Caller must hold dg.mu.
func (dg *DataGrid) prepareDrawLocked(width, height int) (drawSnapshot, bool) {
	snap := drawSnapshot{
		bg:      theme.Bg(),
		fg:      theme.Fg(),
		fgDim:   theme.FgDim(),
		accent:  theme.Accent(),
		warning: theme.Warning(),
		header:  theme.TableHeader(),
	}

	if dg.showHeader {
		snap.headerHeight = 1
	}
	snap.dataHeight = height - snap.headerHeight
	if snap.dataHeight <= 0 {
		return snap, false
	}

	if dg.showRowNumbers {
		snap.gutterWidth = len(itoa(dg.source.RowCount())) + 2
		if snap.gutterWidth < 4 {
			snap.gutterWidth = 4
		}
	}

	snap.hasVScroll = dg.source.RowCount() > snap.dataHeight
	scrollWidth := 0
	if snap.hasVScroll {
		scrollWidth = 1
	}
	snap.contentWidth = width - snap.gutterWidth - scrollWidth

	dg.viewport.VisRows = snap.dataHeight
	dg.colWidths = computeColumnWidths(dg.source, &dg.viewport, snap.contentWidth, dg.showRowNumbers)
	snap.cols = dg.source.Columns()

	dg.ensureCursorColVisible(snap.contentWidth)

	startRow := dg.viewport.RowOffset
	endRow := startRow + dg.viewport.VisRows
	if endRow > dg.source.RowCount() {
		endRow = dg.source.RowCount()
	}
	fetchStart := startRow - dg.overscan
	if fetchStart < 0 {
		fetchStart = 0
	}
	fetchEnd := endRow + dg.overscan
	if fetchEnd > dg.source.RowCount() {
		fetchEnd = dg.source.RowCount()
	}
	dg.source.FetchRange(fetchStart, fetchEnd)

	dg.clampCursor()
	if dg.cursor.Row < dg.viewport.RowOffset {
		dg.viewport.RowOffset = dg.cursor.Row
	}
	if dg.cursor.Row >= dg.viewport.RowOffset+dg.viewport.VisRows {
		dg.viewport.RowOffset = dg.cursor.Row - dg.viewport.VisRows + 1
	}
	snap.startRow = dg.viewport.RowOffset
	snap.endRow = snap.startRow + snap.dataHeight
	if snap.endRow > dg.source.RowCount() {
		snap.endRow = dg.source.RowCount()
	}

	return snap, true
}

// paintLocked writes the DataGrid to screen using the snapshot's geometry.
// It does not mutate any dg.* fields. Caller must hold dg.mu.
func (dg *DataGrid) paintLocked(screen tcell.Screen, x, y, width, height int, snap drawSnapshot) {
	baseStyle := tcell.StyleDefault.Background(snap.bg).Foreground(snap.fg)

	if dg.showHeader {
		headerStyle := tcell.StyleDefault.Background(snap.bg).Foreground(snap.header).Bold(true)
		headerY := y
		for col := x; col < x+width; col++ {
			screen.SetContent(col, headerY, ' ', nil, headerStyle)
		}
		drawX := x + snap.gutterWidth
		dg.drawHeaderCells(screen, drawX, headerY, snap.contentWidth, snap.cols, headerStyle)
	}

	for rowIdx := snap.startRow; rowIdx < snap.endRow; rowIdx++ {
		screenY := y + snap.headerHeight + (rowIdx - snap.startRow)
		isCursorRow := rowIdx == dg.cursor.Row
		isSelectedRow := dg.selectedRows[rowIdx]

		rowBg := snap.bg
		if isSelectedRow {
			rowBg = snap.accent
		}
		clearStyle := tcell.StyleDefault.Background(rowBg).Foreground(snap.fg)
		if isSelectedRow {
			clearStyle = clearStyle.Foreground(snap.bg)
		}
		for col := x; col < x+width; col++ {
			screen.SetContent(col, screenY, ' ', nil, clearStyle)
		}

		if dg.showRowNumbers {
			gutterStyle := tcell.StyleDefault.Background(snap.bg).Foreground(snap.fgDim)
			if isSelectedRow {
				gutterStyle = tcell.StyleDefault.Background(snap.accent).Foreground(snap.bg)
			}
			numStr := padLeft(itoa(rowIdx+1), snap.gutterWidth-1) + " "
			drawCol := x
			for _, ch := range numStr {
				if drawCol < x+snap.gutterWidth {
					screen.SetContent(drawCol, screenY, ch, nil, gutterStyle)
					drawCol++
				}
			}
		}

		drawX := x + snap.gutterWidth
		dg.drawDataCells(screen, drawX, screenY, snap.contentWidth, rowIdx, snap.cols,
			isCursorRow, isSelectedRow, baseStyle,
			snap.bg, snap.fg, snap.accent, snap.warning, snap.fgDim)
	}

	for emptyY := y + snap.headerHeight + (snap.endRow - snap.startRow); emptyY < y+height; emptyY++ {
		for col := x; col < x+width; col++ {
			screen.SetContent(col, emptyY, ' ', nil, baseStyle)
		}
	}

	if snap.hasVScroll {
		dg.drawVScrollbar(screen, x+width-1, y+snap.headerHeight, snap.dataHeight)
	}
}

// drawHeaderCells renders the column headers.
func (dg *DataGrid) drawHeaderCells(screen tcell.Screen, startX, y, maxWidth int, cols []GridColumn, style tcell.Style) {
	drawX := startX

	// Draw frozen columns first
	for i, col := range cols {
		if !col.Frozen || i >= len(dg.colWidths) {
			continue
		}
		if drawX-startX >= maxWidth {
			break
		}
		colW := dg.colWidths[i]
		dg.drawCellText(screen, drawX, y, colW, maxWidth-(drawX-startX), col.Name, col.Align, style, false)
		// Separator
		drawX += colW
		if drawX-startX < maxWidth && dg.separator != 0 {
			sepStyle := style.Dim(true)
			screen.SetContent(drawX-1, y, dg.separator, nil, sepStyle)
		}
	}

	// Draw non-frozen columns starting from ColOffset
	nonFrozenIdx := 0
	for i, col := range cols {
		if col.Frozen || i >= len(dg.colWidths) {
			continue
		}
		if nonFrozenIdx < dg.viewport.ColOffset {
			nonFrozenIdx++
			continue
		}
		if drawX-startX >= maxWidth {
			break
		}
		colW := dg.colWidths[i]
		avail := maxWidth - (drawX - startX)
		dg.drawCellText(screen, drawX, y, colW, avail, col.Name, col.Align, style, false)
		drawX += colW
		if drawX-startX < maxWidth && dg.separator != 0 {
			sepStyle := style.Dim(true)
			screen.SetContent(drawX-1, y, dg.separator, nil, sepStyle)
		}
		nonFrozenIdx++
	}
}

// drawDataCells renders cells for a single data row.
func (dg *DataGrid) drawDataCells(screen tcell.Screen, startX, y, maxWidth, rowIdx int, cols []GridColumn,
	isCursorRow, isSelectedRow bool, baseStyle tcell.Style,
	bgColor, fgColor, accentColor, warningColor, fgDimColor tcell.Color) {

	drawX := startX

	// Draw frozen columns first
	for colIdx, col := range cols {
		if !col.Frozen || colIdx >= len(dg.colWidths) {
			continue
		}
		if drawX-startX >= maxWidth {
			break
		}
		colW := dg.colWidths[colIdx]
		avail := maxWidth - (drawX - startX)

		cellStyle := dg.getCellStyle(rowIdx, colIdx, isCursorRow, isSelectedRow,
			baseStyle, bgColor, fgColor, accentColor, warningColor)
		value := dg.getCellDisplayValue(rowIdx, colIdx)

		dg.drawCellText(screen, drawX, y, colW, avail, value, col.Align, cellStyle, isCursorRow && colIdx == dg.cursor.Col)
		drawX += colW
		if drawX-startX < maxWidth && dg.separator != 0 {
			sepStyle := baseStyle.Foreground(fgDimColor)
			if isSelectedRow {
				sepStyle = tcell.StyleDefault.Background(accentColor).Foreground(bgColor)
			}
			screen.SetContent(drawX-1, y, dg.separator, nil, sepStyle)
		}
	}

	// Draw non-frozen columns
	nonFrozenIdx := 0
	for colIdx, col := range cols {
		if col.Frozen || colIdx >= len(dg.colWidths) {
			continue
		}
		if nonFrozenIdx < dg.viewport.ColOffset {
			nonFrozenIdx++
			continue
		}
		if drawX-startX >= maxWidth {
			break
		}
		colW := dg.colWidths[colIdx]
		avail := maxWidth - (drawX - startX)

		cellStyle := dg.getCellStyle(rowIdx, colIdx, isCursorRow, isSelectedRow,
			baseStyle, bgColor, fgColor, accentColor, warningColor)
		value := dg.getCellDisplayValue(rowIdx, colIdx)

		dg.drawCellText(screen, drawX, y, colW, avail, value, col.Align, cellStyle, isCursorRow && colIdx == dg.cursor.Col)
		drawX += colW
		if drawX-startX < maxWidth && dg.separator != 0 {
			sepStyle := baseStyle.Foreground(fgDimColor)
			if isSelectedRow {
				sepStyle = tcell.StyleDefault.Background(accentColor).Foreground(bgColor)
			}
			screen.SetContent(drawX-1, y, dg.separator, nil, sepStyle)
		}
		nonFrozenIdx++
	}
}

// getCellStyle determines the style for a specific cell.
func (dg *DataGrid) getCellStyle(rowIdx, colIdx int, isCursorRow, isSelectedRow bool,
	baseStyle tcell.Style, bgColor, fgColor, accentColor, warningColor tcell.Color) tcell.Style {

	pos := CellPosition{Row: rowIdx, Col: colIdx}
	isCursorCell := isCursorRow && colIdx == dg.cursor.Col
	isDirty := dg.changeset.IsDirty(pos)

	// Status color from source
	cell := dg.source.Cell(rowIdx, colIdx)
	if cell.Status != nil {
		statusColor := cell.Status.Color()
		return baseStyle.Foreground(statusColor)
	}

	if isCursorCell {
		if dg.mode == GridModeEdit {
			return tcell.StyleDefault.Background(accentColor).Foreground(bgColor).Underline(true)
		}
		return tcell.StyleDefault.Background(accentColor).Foreground(bgColor)
	}

	if isSelectedRow {
		if isDirty {
			return tcell.StyleDefault.Background(warningColor).Foreground(bgColor)
		}
		return tcell.StyleDefault.Background(accentColor).Foreground(bgColor)
	}

	if isDirty {
		return tcell.StyleDefault.Background(warningColor).Foreground(bgColor)
	}

	return baseStyle
}

// getCellDisplayValue returns what to display for a cell, including edit buffer overlay.
func (dg *DataGrid) getCellDisplayValue(rowIdx, colIdx int) string {
	pos := CellPosition{Row: rowIdx, Col: colIdx}

	// If currently editing this cell, show the edit buffer
	if dg.mode == GridModeEdit && dg.editState != nil &&
		rowIdx == dg.cursor.Row && colIdx == dg.cursor.Col {
		return string(dg.editState.buffer)
	}

	// Check changeset overlay
	if change := dg.changeset.GetChange(pos); change != nil {
		return change.NewValue
	}

	// Source value
	return dg.source.Cell(rowIdx, colIdx).Value
}

// drawCellText renders text in a cell area with alignment and padding.
// isEditCell should be true only for the cell currently being edited.
func (dg *DataGrid) drawCellText(screen tcell.Screen, x, y, colWidth, availWidth int, text string, align Align, style tcell.Style, isEditCell bool) {
	// Effective width is the minimum of colWidth and available space
	effectiveWidth := colWidth
	if effectiveWidth > availWidth {
		effectiveWidth = availWidth
	}
	if effectiveWidth <= 0 {
		return
	}

	// Padding: 1 char on each side
	innerWidth := effectiveWidth - 2
	if innerWidth <= 0 {
		// Too narrow for padding, just fill with spaces
		for i := 0; i < effectiveWidth; i++ {
			screen.SetContent(x+i, y, ' ', nil, style)
		}
		return
	}

	// Clear cell area
	for i := 0; i < effectiveWidth; i++ {
		screen.SetContent(x+i, y, ' ', nil, style)
	}

	// Truncate text if needed
	runes := []rune(text)
	if len(runes) > innerWidth {
		runes = runes[:innerWidth]
	}

	// Calculate text start position based on alignment
	textStart := x + 1 // Left padding
	switch align {
	case AlignRight:
		textStart = x + effectiveWidth - 1 - len(runes) // Right padding
	case AlignCenter:
		textStart = x + (effectiveWidth-len(runes))/2
	}

	// Draw text
	for i, ch := range runes {
		pos := textStart + i
		if pos >= x && pos < x+effectiveWidth {
			screen.SetContent(pos, y, ch, nil, style)
		}
	}

	// Draw edit cursor if in edit mode on the active cell
	if dg.mode == GridModeEdit && dg.editState != nil && isEditCell {
		cursorX := textStart + dg.editState.cursorPos
		if cursorX >= x && cursorX < x+effectiveWidth {
			// Invert the character at cursor position for visibility
			_, _, curStyle, _ := screen.GetContent(cursorX, y)
			fg, bg, _ := curStyle.Decompose()
			screen.SetContent(cursorX, y, ' ', nil, tcell.StyleDefault.Background(fg).Foreground(bg))
			// Re-draw the character if there is one
			if dg.editState.cursorPos < len(dg.editState.buffer) {
				ch := dg.editState.buffer[dg.editState.cursorPos]
				screen.SetContent(cursorX, y, ch, nil, tcell.StyleDefault.Background(fg).Foreground(bg))
			}
		}
	}
}

// drawVScrollbar renders the vertical scrollbar.
func (dg *DataGrid) drawVScrollbar(screen tcell.Screen, x, y, height int) {
	bgColor := theme.BgLight()
	thumbColor := theme.FgDim()

	totalRows := dg.source.RowCount()

	thumbSize := height * height / totalRows
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > height {
		thumbSize = height
	}

	thumbPos := 0
	if totalRows > height {
		thumbPos = dg.viewport.RowOffset * (height - thumbSize) / (totalRows - height)
	}

	// Track
	trackStyle := tcell.StyleDefault.Background(bgColor)
	for i := 0; i < height; i++ {
		screen.SetContent(x, y+i, ' ', nil, trackStyle)
	}

	// Thumb
	thumbStyle := tcell.StyleDefault.Background(thumbColor)
	for i := 0; i < thumbSize; i++ {
		if thumbPos+i < height {
			screen.SetContent(x, y+thumbPos+i, ' ', nil, thumbStyle)
		}
	}
}
