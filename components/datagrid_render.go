package components

import (
	"github.com/atterpac/jig/theme"
	"github.com/gdamore/tcell/v2"
)

// Draw renders the DataGrid to the screen.
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

	// Read theme colors at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()
	warningColor := theme.Warning()
	headerColor := theme.TableHeader()

	// Calculate layout
	headerHeight := 0
	if dg.showHeader {
		headerHeight = 1
	}
	dataHeight := height - headerHeight
	if dataHeight <= 0 {
		return
	}

	// Row number gutter width
	gutterWidth := 0
	if dg.showRowNumbers {
		rowCount := dg.source.RowCount()
		gutterWidth = len(itoa(rowCount)) + 2 // digits + padding
		if gutterWidth < 4 {
			gutterWidth = 4
		}
	}

	// Reserve scrollbar space
	hasVScroll := dg.source.RowCount() > dataHeight
	scrollWidth := 0
	if hasVScroll {
		scrollWidth = 1
	}

	contentWidth := width - gutterWidth - scrollWidth

	// Update viewport dimensions
	dg.viewport.VisRows = dataHeight

	// Compute column widths from source
	dg.colWidths = computeColumnWidths(dg.source, &dg.viewport, contentWidth, dg.showRowNumbers)
	cols := dg.source.Columns()

	// Adjust horizontal scroll so cursor column is visible
	dg.ensureCursorColVisible(contentWidth)

	// Compute horizontal scroll bounds
	totalColWidth := 0
	for _, w := range dg.colWidths {
		totalColWidth += w
	}
	hasHScroll := totalColWidth > contentWidth

	// Pre-fetch visible range
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

	// Ensure cursor visible
	dg.clampCursor()
	if dg.cursor.Row < dg.viewport.RowOffset {
		dg.viewport.RowOffset = dg.cursor.Row
	}
	if dg.cursor.Row >= dg.viewport.RowOffset+dg.viewport.VisRows {
		dg.viewport.RowOffset = dg.cursor.Row - dg.viewport.VisRows + 1
	}
	// Recalculate after potential adjustment
	startRow = dg.viewport.RowOffset
	endRow = startRow + dg.viewport.VisRows
	if endRow > dg.source.RowCount() {
		endRow = dg.source.RowCount()
	}

	// Base style
	baseStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)

	// Draw header row
	if dg.showHeader {
		headerStyle := tcell.StyleDefault.Background(bgColor).Foreground(headerColor).Bold(true)
		headerY := y

		// Clear header line
		for col := x; col < x+width; col++ {
			screen.SetContent(col, headerY, ' ', nil, headerStyle)
		}

		// Draw gutter header space
		drawX := x + gutterWidth

		// Draw header cells
		dg.drawHeaderCells(screen, drawX, headerY, contentWidth, cols, headerStyle)
	}

	// Draw data rows
	for rowIdx := startRow; rowIdx < endRow; rowIdx++ {
		screenY := y + headerHeight + (rowIdx - startRow)
		isCursorRow := rowIdx == dg.cursor.Row
		isSelectedRow := dg.selectedRows[rowIdx]

		// Clear entire row
		rowBg := bgColor
		if isSelectedRow {
			rowBg = accentColor
		}
		clearStyle := tcell.StyleDefault.Background(rowBg).Foreground(fgColor)
		if isSelectedRow {
			clearStyle = clearStyle.Foreground(bgColor)
		}
		for col := x; col < x+width; col++ {
			screen.SetContent(col, screenY, ' ', nil, clearStyle)
		}

		// Draw row number gutter
		if dg.showRowNumbers {
			gutterStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
			if isSelectedRow {
				gutterStyle = tcell.StyleDefault.Background(accentColor).Foreground(bgColor)
			}
			numStr := padLeft(itoa(rowIdx+1), gutterWidth-1) + " "
			drawCol := x
			for _, ch := range numStr {
				if drawCol < x+gutterWidth {
					screen.SetContent(drawCol, screenY, ch, nil, gutterStyle)
					drawCol++
				}
			}
		}

		// Draw cells
		drawX := x + gutterWidth
		dg.drawDataCells(screen, drawX, screenY, contentWidth, rowIdx, cols,
			isCursorRow, isSelectedRow, baseStyle, bgColor, fgColor, accentColor, warningColor, fgDimColor)
	}

	// Clear empty rows below data
	for emptyY := y + headerHeight + (endRow - startRow); emptyY < y+height; emptyY++ {
		for col := x; col < x+width; col++ {
			screen.SetContent(col, emptyY, ' ', nil, baseStyle)
		}
	}

	// Draw vertical scrollbar
	if hasVScroll {
		dg.drawVScrollbar(screen, x+width-1, y+headerHeight, dataHeight)
	}

	// Draw horizontal scrollbar indicator (in bottom-right if needed)
	_ = hasHScroll
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
