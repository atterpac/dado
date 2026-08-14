package inline

import (
	"strconv"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
)

// TableAlign controls how cell content is positioned within a column.
type TableAlign uint8

const (
	AlignLeft TableAlign = iota
	AlignCenter
	AlignRight
)

// TableColumn describes one renderer-native table column. Width supplies its
// preferred content width when positive; otherwise it is inferred. Columns
// may still shrink to fit the frame. MinWidth defaults to one cell.
type TableColumn struct {
	Header   string
	Width    int
	MinWidth int
	Align    TableAlign
}

// Table renders tabular data into a Frame and safely snapshots concurrent
// updates.
type Table struct {
	mu      sync.RWMutex
	title   string
	columns []TableColumn
	rows    [][]string
	maxRows int
	theme   InlineTheme
}

// NewTable creates a table with an optional title and column definitions.
func NewTable(title string, columns ...TableColumn) *Table {
	return &Table{title: title, columns: append([]TableColumn(nil), columns...), theme: RoundedInlineTheme()}
}

// SetTheme changes the table's visual preset.
func (t *Table) SetTheme(theme InlineTheme) *Table {
	t.mu.Lock()
	t.theme = normalizedInlineTheme(theme)
	t.mu.Unlock()
	return t
}

// SetMaxRows limits visible data rows. Zero shows every row.
func (t *Table) SetMaxRows(maxRows int) *Table {
	t.mu.Lock()
	t.maxRows = max(maxRows, 0)
	t.mu.Unlock()
	return t
}

// SetRows replaces the table data.
func (t *Table) SetRows(rows ...[]string) *Table {
	t.mu.Lock()
	t.rows = cloneTableRows(rows)
	t.mu.Unlock()
	return t
}

// AddRow appends one data row.
func (t *Table) AddRow(cells ...string) *Table {
	t.mu.Lock()
	t.rows = append(t.rows, append([]string(nil), cells...))
	t.mu.Unlock()
	return t
}

// Frame returns the table's current renderer frame.
func (t *Table) Frame(width int) *Frame {
	width = max(width, 0)
	t.mu.RLock()
	title := t.title
	columns := append([]TableColumn(nil), t.columns...)
	rows := cloneTableRows(t.rows)
	maxRows := t.maxRows
	theme := t.theme
	t.mu.RUnlock()

	visible := len(rows)
	if maxRows > 0 {
		visible = min(visible, maxRows)
	}
	overflow := len(rows) - visible
	height := 0
	if title != "" {
		height++
	}
	if len(columns) > 0 {
		height += 4 + visible
		if overflow > 0 {
			height++
		}
	}
	frame := NewFrame(width, height)
	y := 0
	if title != "" {
		drawClipped(frame, 0, y, title, width, theme.Accent.Bold(true))
		y++
	}
	if len(columns) == 0 || width == 0 {
		return frame
	}
	columnWidths := fitTableColumns(width, columns, rows)
	tableWidth := min(width, tableWidthForColumns(columnWidths))
	drawTableBorder(frame, y, tableWidth, columnWidths, theme.Borders.TopLeft, theme.Borders.TopRight, theme.Table.Top, theme)
	y++
	drawTableRow(frame, y, tableWidth, columnWidths, columns, tableHeaders(columns), theme.Label.Bold(true), theme)
	y++
	drawTableBorder(frame, y, tableWidth, columnWidths, theme.Table.Left, theme.Table.Right, theme.Table.Middle, theme)
	y++
	for _, row := range rows[:visible] {
		drawTableRow(frame, y, tableWidth, columnWidths, columns, row, theme.Text, theme)
		y++
	}
	if overflow > 0 {
		drawTableSpanningRow(frame, y, tableWidth, "… "+moreRowsLabel(overflow), theme.Muted, theme)
		y++
	}
	drawTableBorder(frame, y, tableWidth, columnWidths, theme.Borders.BottomLeft, theme.Borders.BottomRight, theme.Table.Bottom, theme)
	return frame
}

func fitTableColumns(width int, columns []TableColumn, rows [][]string) []int {
	widths := make([]int, len(columns))
	for index, column := range columns {
		widths[index] = max(displayWidth(column.Header), max(column.MinWidth, 1))
		if column.Width > 0 {
			widths[index] = column.Width
		}
	}
	for _, row := range rows {
		for index := range min(len(row), len(widths)) {
			if columns[index].Width == 0 {
				widths[index] = max(widths[index], displayWidth(row[index]))
			}
		}
	}
	budget := max(width-(3*len(columns)+1), 0)
	for tableWidthSum(widths) > budget {
		largest := -1
		for index, columnWidth := range widths {
			minimum := max(columns[index].MinWidth, 1)
			if columnWidth > minimum && (largest < 0 || columnWidth > widths[largest]) {
				largest = index
			}
		}
		if largest < 0 {
			break
		}
		widths[largest]--
	}
	return widths
}

func drawTableRow(frame *Frame, y, width int, widths []int, columns []TableColumn, cells []string, style tcell.Style, theme InlineTheme) {
	x := 0
	drawClipped(frame, x, y, theme.Borders.Vertical, width, theme.Border)
	x += displayWidth(theme.Borders.Vertical)
	for index, columnWidth := range widths {
		drawClipped(frame, x, y, " ", max(width-x, 0), style)
		x++
		cell := ""
		if index < len(cells) {
			cell = cells[index]
		}
		drawClipped(frame, x, y, alignTableCell(cell, columnWidth, columns[index].Align), max(width-x, 0), style)
		x += columnWidth
		drawClipped(frame, x, y, " "+theme.Borders.Vertical, max(width-x, 0), theme.Border)
		x += 1 + displayWidth(theme.Borders.Vertical)
	}
}

func drawTableSpanningRow(frame *Frame, y, width int, text string, style tcell.Style, theme InlineTheme) {
	drawClipped(frame, 0, y, theme.Borders.Vertical, width, theme.Border)
	drawClipped(frame, 2, y, text, max(width-4, 0), style)
	if width > 1 {
		drawClipped(frame, width-1, y, theme.Borders.Vertical, 1, theme.Border)
	}
}

func drawTableBorder(frame *Frame, y, width int, widths []int, left, right, join string, theme InlineTheme) {
	x := 0
	drawClipped(frame, x, y, left, width, theme.Border)
	x += displayWidth(left)
	for index, columnWidth := range widths {
		lineWidth := columnWidth + 2
		drawClipped(frame, x, y, strings.Repeat(theme.Borders.Horizontal, lineWidth), max(width-x, 0), theme.Border)
		x += lineWidth
		glyph := join
		if index == len(widths)-1 {
			glyph = right
		}
		drawClipped(frame, x, y, glyph, max(width-x, 0), theme.Border)
		x += displayWidth(glyph)
	}
}

func alignTableCell(text string, width int, alignment TableAlign) string {
	text = truncateCells(text, width)
	padding := max(width-displayWidth(text), 0)
	left := 0
	switch alignment {
	case AlignRight:
		left = padding
	case AlignCenter:
		left = padding / 2
	}
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", padding-left)
}

func truncateCells(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if displayWidth(text) <= width {
		return text
	}
	if width == 1 {
		return "…"
	}
	result := ""
	used := 0
	for _, r := range text {
		runeWidth := runeWidth(r)
		if used+runeWidth > width-1 {
			break
		}
		result += string(r)
		used += runeWidth
	}
	return result + "…"
}

func tableHeaders(columns []TableColumn) []string {
	headers := make([]string, len(columns))
	for index, column := range columns {
		headers[index] = column.Header
	}
	return headers
}

func cloneTableRows(rows [][]string) [][]string {
	cloned := make([][]string, len(rows))
	for index, row := range rows {
		cloned[index] = append([]string(nil), row...)
	}
	return cloned
}

func tableWidthSum(widths []int) int {
	total := 0
	for _, width := range widths {
		total += width
	}
	return total
}

func tableWidthForColumns(widths []int) int {
	return tableWidthSum(widths) + 3*len(widths) + 1
}

func moreRowsLabel(count int) string {
	if count == 1 {
		return "1 more row"
	}
	return strconv.Itoa(count) + " more rows"
}
