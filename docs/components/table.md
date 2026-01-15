# Table Patterns

Advanced table usage, selection, and data manipulation.

## Basic Usage

```go
table := components.NewTable().
    SetHeaders("Name", "Status", "Created").
    SetSelectable(true)

// Add data rows
table.AddRow("Alice", "Active", "2024-01-15")
table.AddRow("Bob", "Inactive", "2024-01-10")
table.AddRow("Charlie", "Pending", "2024-01-20")

// Handle selection
table.SetOnSelect(func(row int) {
    data := table.GetRowData(row)
    log.Printf("Selected: %v", data)
})
```

---

## Row Styling

### Colored Rows

```go
// Single color for all cells
index := table.AddRowWithColor(theme.Success(), "Success Item", "Active", "Now")

// Different colors per cell
table.AddColoredRow(
    []string{"Mixed", "Colors", "Row"},
    []tcell.Color{theme.Fg(), theme.Warning(), theme.Info()},
)
```

### Status-Based Styling

Use typed statuses for automatic theme updates:

```go
// Status in specific column (with icon)
table.AddRowWithStatus(
    theme.StatusSuccess(),  // Status type
    1,                      // Column index for status
    "Alice", "Active", "2024-01-15",
)

table.AddRowWithStatus(
    theme.StatusError(),
    1,
    "Bob", "Failed", "2024-01-10",
)

// Available statuses:
// theme.StatusSuccess()
// theme.StatusWarning()
// theme.StatusError()
// theme.StatusInfo()
// theme.StatusPending()
```

### Styled Cells

Full control over cell styling:

```go
table.AddStyledRow([]components.TableCell{
    {Text: "Alice", Color: theme.Fg(), Align: tview.AlignLeft, Expansion: 2},
    {Text: "Active", Status: theme.StatusSuccess(), Align: tview.AlignCenter},
    {Text: "2024-01-15", Color: theme.FgDim(), Align: tview.AlignRight, MaxWidth: 12},
})
```

---

## Multi-Selection

```go
table := components.NewTable().
    SetHeaders("Name", "Email", "Role").
    SetMultiSelect(true)

// Selection controls
// Space - Toggle current row
// Ctrl+A - Select all

// Track selection changes
table.SetOnSelectionChange(func(rows []int) {
    log.Printf("Selected %d rows", len(rows))
})

// Programmatic selection
table.ToggleSelection()   // Toggle current row
table.SelectAll()
table.ClearSelection()

// Check selection
selected := table.GetSelectedRows()  // []int
isSelected := table.IsRowSelected(0)
```

---

## Key-Based Selection

For stable selection across data refreshes:

```go
// Associate keys with rows
table.AddRow("Alice", "alice@example.com", "Admin")
table.SetRowKey(0, "user-123")  // Associate key with data index

table.AddRow("Bob", "bob@example.com", "User")
table.SetRowKey(1, "user-456")

// Select by key (survives data refresh)
table.SelectByKey("user-123")
table.DeselectByKey("user-123")
table.ToggleSelectionByKey("user-123")

// Get selected keys
keys := table.GetSelectedKeys()  // []string

// Find row by key
index := table.GetRowByKey("user-123")  // Returns -1 if not found

// Get key for row
key := table.GetRowKey(0)  // Returns "" if not set
```

---

## Row Manipulation

### Update Rows

```go
// Update by index (0-based, after header)
table.UpdateRow(0, "Alice Updated", "Inactive", "2024-01-16")

// Update with colors
table.UpdateColoredRow(0,
    []string{"Alice", "Inactive", "2024-01-16"},
    []tcell.Color{0, theme.Warning(), 0},
)

// Update with full styling
table.UpdateStyledRow(0, []components.TableCell{
    {Text: "Alice", Color: theme.Fg()},
    {Text: "Inactive", Status: theme.StatusWarning()},
    {Text: "2024-01-16"},
})
```

### Insert and Remove

```go
// Insert at position (shifts existing rows)
table.InsertRowAt(1, "New User", "Pending", "2024-01-25")
table.InsertColoredRowAt(1,
    []string{"New User", "Pending", "2024-01-25"},
    []tcell.Color{0, theme.Warning(), 0},
)

// Remove row
table.RemoveRowAt(2)

// Clear all data rows (keeps headers)
table.ClearRows()
```

### Read Row Data

```go
// Get text content
data := table.GetRowData(0)  // []string{"Alice", "Active", "2024-01-15"}

// Get cell structs
cells := table.GetRowCells(0)  // []TableCell with full styling info
```

---

## Navigation and Selection

```go
// Get current selection
row := table.SelectedRow()      // 0-based data index (-1 if none)
tableRow, col := table.GetSelection()  // tview row (includes header)

// Set selection
table.SelectRow(0)      // Select by data index
table.ScrollToRow(10)   // Scroll to row

// Row count
count := table.RowCount()       // Data rows only
count := table.GetDataRowCount()  // Same as RowCount
```

---

## Selection Callbacks

```go
table := components.NewTable().
    SetHeaders("Name", "Email").
    SetSelectable(true)

// Called on Enter key
table.SetOnSelect(func(row int) {
    user := users[row]
    showUserDetails(user)
})

// Called when cursor moves
table.SetSelectionChangedFunc(func(row, col int) {
    if row > 0 {  // Skip header
        updatePreview(row - 1)
    }
})

// Called when multi-selection changes
table.SetOnSelectionChange(func(rows []int) {
    updateBulkActions(len(rows) > 0)
})
```

---

## Data Binding Pattern

```go
type UserTable struct {
    table *components.Table
    users []User
}

func NewUserTable() *UserTable {
    ut := &UserTable{
        table: components.NewTable().
            SetHeaders("Name", "Email", "Role").
            SetSelectable(true),
    }

    ut.table.SetOnSelect(func(row int) {
        user := ut.users[row]
        showUser(user)
    })

    return ut
}

func (ut *UserTable) SetData(users []User) {
    ut.users = users
    ut.table.ClearRows()

    for i, u := range users {
        ut.table.AddRow(u.Name, u.Email, u.Role)
        ut.table.SetRowKey(i, u.ID)  // For stable selection
    }
}

func (ut *UserTable) GetSelected() *User {
    row := ut.table.SelectedRow()
    if row >= 0 && row < len(ut.users) {
        return &ut.users[row]
    }
    return nil
}
```

---

## Refresh with Preserved Selection

```go
func (ut *UserTable) Refresh(users []User) {
    // Save selected keys
    selectedKeys := ut.table.GetSelectedKeys()
    currentKey := ut.table.GetRowKey(ut.table.SelectedRow())

    // Update data
    ut.SetData(users)

    // Restore selection by key
    for _, key := range selectedKeys {
        ut.table.SelectByKey(key)
    }

    // Restore cursor position
    if idx := ut.table.GetRowByKey(currentKey); idx >= 0 {
        ut.table.SelectRow(idx)
    }
}
```

---

## Theme Integration

Tables automatically update with theme changes:

```go
// Status colors update automatically
table.AddRowWithStatus(theme.StatusSuccess(), 1, "Item", "Active", "Now")

// When theme changes, status colors are refreshed in Draw()
// No manual refresh needed
```

---

## Best Practices

1. **Use key-based selection** for data that refreshes frequently
2. **Set headers** to enable fixed header row
3. **Use status types** for semantic coloring (auto-updates with theme)
4. **Handle empty state** when table has no data
5. **Consider virtual list** for 1000+ rows

```go
// Empty state handling
if len(users) == 0 {
    table.AddRow("No users found", "", "")
    table.SetSelectable(false)
} else {
    table.SetSelectable(true)
    for _, u := range users {
        table.AddRow(u.Name, u.Email, u.Role)
    }
}
```
