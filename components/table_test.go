package components

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTable_NewTable tests Table creation.
func TestTable_NewTable(t *testing.T) {
	table := NewTable()

	assert.NotNil(t, table)
	assert.Equal(t, 0, table.GetDataRowCount())
}

// TestTable_SetHeaders tests setting column headers.
func TestTable_SetHeaders(t *testing.T) {
	table := NewTable()

	table.SetHeaders("Name", "Age", "Status")

	// Header row is row 0
	assert.Equal(t, 0, table.GetDataRowCount()) // No data rows yet
}

// TestTable_AddRow tests adding basic rows.
func TestTable_AddRow(t *testing.T) {
	table := NewTable()

	table.AddRow("Alice", "30", "Active")
	table.AddRow("Bob", "25", "Inactive")

	assert.Equal(t, 2, table.GetDataRowCount())
}

// TestTable_AddRowWithHeader tests adding rows with headers.
func TestTable_AddRowWithHeader(t *testing.T) {
	table := NewTable()

	table.SetHeaders("Name", "Age")
	table.AddRow("Alice", "30")
	table.AddRow("Bob", "25")

	assert.Equal(t, 2, table.GetDataRowCount())
}

// TestTable_AddColoredRow tests adding rows with colors.
func TestTable_AddColoredRow(t *testing.T) {
	table := NewTable()

	colors := []tcell.Color{tcell.ColorRed, tcell.ColorGreen, tcell.ColorBlue}
	table.AddColoredRow([]string{"Red", "Green", "Blue"}, colors)

	assert.Equal(t, 1, table.GetDataRowCount())
}

// TestTable_AddStyledRow tests adding rows with full styling.
func TestTable_AddStyledRow(t *testing.T) {
	table := NewTable()

	table.AddStyledRow([]TableCell{
		{Text: "Cell 1", Color: tcell.ColorRed, Align: 0, Expansion: 1, Selectable: true},
		{Text: "Cell 2", Color: tcell.ColorGreen, Align: 1, Expansion: 2, Selectable: true},
		{Text: "Cell 3", Color: 0, Align: 2, Expansion: 1, MaxWidth: 20, Selectable: false},
	})

	assert.Equal(t, 1, table.GetDataRowCount())
}

// TestTable_ClearRows tests clearing data rows.
func TestTable_ClearRows(t *testing.T) {
	table := NewTable()

	table.SetHeaders("Name", "Age")
	table.AddRow("Alice", "30")
	table.AddRow("Bob", "25")
	require.Equal(t, 2, table.GetDataRowCount())

	table.ClearRows()

	assert.Equal(t, 0, table.GetDataRowCount())
}

// TestTable_ClearRowsWithoutHeader tests clearing rows without headers.
func TestTable_ClearRowsWithoutHeader(t *testing.T) {
	table := NewTable()

	table.AddRow("Row 1")
	table.AddRow("Row 2")
	require.Equal(t, 2, table.GetDataRowCount())

	table.ClearRows()

	assert.Equal(t, 0, table.GetDataRowCount())
}

// TestTable_MultiSelect tests multi-selection functionality.
func TestTable_MultiSelect(t *testing.T) {
	table := NewTable()

	table.SetMultiSelect(true)
	table.AddRow("Row 1")
	table.AddRow("Row 2")
	table.AddRow("Row 3")

	// Initially no selection
	assert.Empty(t, table.GetSelectedRows())

	// Toggle selection
	table.Table.Select(0, 0) // Select first row in tview
	table.ToggleSelection()

	assert.True(t, table.IsRowSelected(0))
	assert.Len(t, table.GetSelectedRows(), 1)

	// Toggle another row
	table.Table.Select(2, 0)
	table.ToggleSelection()

	assert.True(t, table.IsRowSelected(2))
	assert.Len(t, table.GetSelectedRows(), 2)

	// Toggle off
	table.Table.Select(0, 0)
	table.ToggleSelection()

	assert.False(t, table.IsRowSelected(0))
	assert.Len(t, table.GetSelectedRows(), 1)
}

// TestTable_SelectAll tests selecting all rows.
func TestTable_SelectAll(t *testing.T) {
	table := NewTable()

	table.SetMultiSelect(true)
	table.AddRow("Row 1")
	table.AddRow("Row 2")
	table.AddRow("Row 3")

	table.SelectAll()

	assert.Len(t, table.GetSelectedRows(), 3)
	assert.True(t, table.IsRowSelected(0))
	assert.True(t, table.IsRowSelected(1))
	assert.True(t, table.IsRowSelected(2))
}

// TestTable_SelectAllWithHeader tests selecting all with header present.
func TestTable_SelectAllWithHeader(t *testing.T) {
	table := NewTable()

	table.SetMultiSelect(true)
	table.SetHeaders("Name")
	table.AddRow("Row 1")
	table.AddRow("Row 2")

	table.SelectAll()

	// Should select data rows but not header
	selected := table.GetSelectedRows()
	assert.Len(t, selected, 2)
	assert.False(t, table.IsRowSelected(0)) // Header row
}

// TestTable_ClearSelection tests clearing selection.
func TestTable_ClearSelection(t *testing.T) {
	table := NewTable()

	table.SetMultiSelect(true)
	table.AddRow("Row 1")
	table.AddRow("Row 2")

	table.SelectAll()
	require.Len(t, table.GetSelectedRows(), 2)

	table.ClearSelection()

	assert.Empty(t, table.GetSelectedRows())
}

// TestTable_OnSelect tests selection callback.
func TestTable_OnSelect(t *testing.T) {
	table := NewTable()

	var selectedRow int = -1
	table.SetOnSelect(func(row int) {
		selectedRow = row
	})

	table.AddRow("Row 1")
	table.AddRow("Row 2")

	// Simulate Enter key on row 1
	table.Table.Select(1, 0)
	handler := table.InputHandler()
	handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)

	assert.Equal(t, 1, selectedRow)
}

// TestTable_OnSelectWithHeader tests selection callback with header.
func TestTable_OnSelectWithHeader(t *testing.T) {
	table := NewTable()

	var selectedRow int = -1
	table.SetOnSelect(func(row int) {
		selectedRow = row
	})

	table.SetHeaders("Name")
	table.AddRow("Row 1")
	table.AddRow("Row 2")

	// Select table row 2 (data row 1)
	table.Table.Select(2, 0)
	handler := table.InputHandler()
	handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)

	assert.Equal(t, 1, selectedRow) // Data index, not table row
}

// TestTable_OnSelectionChange tests multi-selection change callback.
func TestTable_OnSelectionChange(t *testing.T) {
	table := NewTable()

	var changes [][]int
	table.SetOnSelectionChange(func(rows []int) {
		changes = append(changes, rows)
	})

	table.SetMultiSelect(true)
	table.AddRow("Row 1")
	table.AddRow("Row 2")

	table.Table.Select(0, 0)
	table.ToggleSelection()

	require.Len(t, changes, 1)
	assert.Contains(t, changes[0], 0)

	table.Table.Select(1, 0)
	table.ToggleSelection()

	require.Len(t, changes, 2)
}

// TestTable_InputHandler_MultiSelect tests multi-select key handling.
func TestTable_InputHandler_MultiSelect(t *testing.T) {
	table := NewTable()

	table.SetMultiSelect(true)
	table.AddRow("Row 1")
	table.AddRow("Row 2")
	table.AddRow("Row 3")

	handler := table.InputHandler()

	t.Run("Space toggles selection", func(t *testing.T) {
		table.ClearSelection()
		table.Table.Select(0, 0)
		handler(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone), nil)
		assert.True(t, table.IsRowSelected(0))
	})

	t.Run("Ctrl+A selects all", func(t *testing.T) {
		table.ClearSelection()
		handler(tcell.NewEventKey(tcell.KeyCtrlA, 0, tcell.ModNone), nil)
		assert.Len(t, table.GetSelectedRows(), 3)
	})
}

// TestTable_ScrollToRow tests scrolling to a specific row.
func TestTable_ScrollToRow(t *testing.T) {
	table := NewTable()

	for i := 0; i < 100; i++ {
		table.AddRow("Row")
	}

	// Should not panic
	table.ScrollToRow(50)
	table.ScrollToRow(0)
	table.ScrollToRow(99)
}

// TestTable_GetDataRowCount tests data row counting.
func TestTable_GetDataRowCount(t *testing.T) {
	t.Run("without header", func(t *testing.T) {
		table := NewTable()
		assert.Equal(t, 0, table.GetDataRowCount())

		table.AddRow("Row 1")
		assert.Equal(t, 1, table.GetDataRowCount())

		table.AddRow("Row 2")
		assert.Equal(t, 2, table.GetDataRowCount())
	})

	t.Run("with header", func(t *testing.T) {
		table := NewTable()
		table.SetHeaders("Col1")
		assert.Equal(t, 0, table.GetDataRowCount())

		table.AddRow("Row 1")
		assert.Equal(t, 1, table.GetDataRowCount())
	})
}

// TestTable_FluentAPI tests method chaining.
func TestTable_FluentAPI(t *testing.T) {
	var selected int = -1

	table := NewTable().
		SetHeaders("Name", "Age", "Status").
		AddRow("Alice", "30", "Active").
		AddRow("Bob", "25", "Inactive").
		SetMultiSelect(true).
		SetOnSelect(func(row int) {
			selected = row
		}).
		SetOnSelectionChange(func(rows []int) {})

	assert.Equal(t, 2, table.GetDataRowCount())
	_ = selected
}

// TestTable_IsRowSelected tests row selection check.
func TestTable_IsRowSelected(t *testing.T) {
	table := NewTable()
	table.SetMultiSelect(true)
	table.AddRow("Row 1")

	assert.False(t, table.IsRowSelected(0))

	table.Table.Select(0, 0)
	table.ToggleSelection()

	assert.True(t, table.IsRowSelected(0))
}

// TestTable_ToggleSelectionOnHeader tests that header can't be selected.
func TestTable_ToggleSelectionOnHeader(t *testing.T) {
	table := NewTable()
	table.SetMultiSelect(true)
	table.SetHeaders("Name")
	table.AddRow("Row 1")

	// Try to toggle header row
	table.Table.Select(0, 0)
	table.ToggleSelection()

	// Should not be selected
	assert.False(t, table.IsRowSelected(0))
}

// TestTable_IndexConversion tests data index to table row conversion.
func TestTable_IndexConversion(t *testing.T) {
	t.Run("without header", func(t *testing.T) {
		table := NewTable()
		table.AddRow("Row 0")
		table.AddRow("Row 1")

		// Data index 0 should be table row 0
		assert.Equal(t, 0, table.dataIndexToTableRow(0))
		assert.Equal(t, 1, table.dataIndexToTableRow(1))

		// Table row 0 should be data index 0
		assert.Equal(t, 0, table.tableRowToDataIndex(0))
		assert.Equal(t, 1, table.tableRowToDataIndex(1))
	})

	t.Run("with header", func(t *testing.T) {
		table := NewTable()
		table.SetHeaders("Name")
		table.AddRow("Row 0")
		table.AddRow("Row 1")

		// Data index 0 should be table row 1 (after header)
		assert.Equal(t, 1, table.dataIndexToTableRow(0))
		assert.Equal(t, 2, table.dataIndexToTableRow(1))

		// Table row 1 should be data index 0
		assert.Equal(t, 0, table.tableRowToDataIndex(1))
		assert.Equal(t, 1, table.tableRowToDataIndex(2))
	})
}

// TestTable_EmptyTable tests operations on empty table.
func TestTable_EmptyTable(t *testing.T) {
	table := NewTable()

	// Should not panic
	table.ClearRows()
	table.ClearSelection()
	table.SelectAll()
	table.GetSelectedRows()
	table.GetDataRowCount()

	assert.Equal(t, 0, table.GetDataRowCount())
	assert.Empty(t, table.GetSelectedRows())
}
