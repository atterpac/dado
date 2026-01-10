---
label: Table
icon: table
order: 80
---

# Table

Enhanced table with headers, selection, and styling.

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

table := components.NewTable()
table.SetHeaders("Name", "Status", "Created")

// Add rows
table.AddRow("Alice", "Active", "2024-01-15")
table.AddRow("Bob", "Inactive", "2024-01-10")
```

---

## Selection

### Single Selection

```go
table := components.NewTable()

table.SetOnSelect(func(row int) {
    data := table.GetRowData(row)
    log.Printf("Selected: %v", data)
})
```

### Multi-Select

```go
table.SetMultiSelect(true)

// Toggle current row selection
table.ToggleSelection()

// Bulk operations
table.SelectAll()
table.ClearSelection()

// Get selected rows
rows := table.GetSelectedRows()  // []int
```

---

## Styling

### Colored Rows

```go
table.AddColoredRow(
    []string{"Charlie", "Pending", "2024-01-20"},
    []tcell.Color{0, theme.Warning(), 0},  // 0 = default color
)
```

### Status Colors

```go
table.AddRowWithStatus(
    theme.StatusSuccess(),  // Status indicator color
    1,                      // Status column index
    "Dave", "Active", "2024-01-22",
)
```

---

## Methods

### Row Management

| Method | Description |
|--------|-------------|
| `SetHeaders(...string)` | Set column headers |
| `AddRow(...string)` | Add a row |
| `AddColoredRow([]string, []tcell.Color)` | Add row with colors |
| `AddRowWithStatus(tcell.Color, int, ...string)` | Add row with status |
| `UpdateRow(int, ...string)` | Update existing row |
| `InsertRowAt(int, ...string)` | Insert row at index |
| `RemoveRowAt(int)` | Remove row at index |
| `ClearRows()` | Clear all rows (keep headers) |
| `Clear()` | Clear everything |

### Selection

| Method | Description |
|--------|-------------|
| `SetMultiSelect(bool)` | Enable multi-selection |
| `ToggleSelection()` | Toggle current row |
| `SelectAll()` | Select all rows |
| `ClearSelection()` | Deselect all rows |
| `GetSelectedRows()` | Get selected row indices |
| `GetRowData(int)` | Get row data by index |

### Events

| Method | Description |
|--------|-------------|
| `SetOnSelect(func(int))` | Selection callback |
| `SetOnMultiSelect(func([]int))` | Multi-select callback |

---

## Example

```go
type UsersView struct {
    *components.ComponentBase
    table *components.Table
    app   *layout.App
}

func NewUsersView(app *layout.App, users []User) *UsersView {
    table := components.NewTable()
    table.SetHeaders("Name", "Email", "Role", "Status")
    table.SetMultiSelect(true)

    for _, u := range users {
        statusColor := theme.Success()
        if u.Status == "Inactive" {
            statusColor = theme.FgMuted()
        }
        table.AddRowWithStatus(statusColor, 3, u.Name, u.Email, u.Role, u.Status)
    }

    table.SetOnSelect(func(row int) {
        data := table.GetRowData(row)
        app.Pages().Push(NewUserDetail(app, data))
    })

    v := &UsersView{table: table, app: app}
    v.ComponentBase = components.NewComponentBase(table).
        SetName("users").
        AddHint("Enter", "View").
        AddHint("Space", "Select").
        AddHint("a", "Select All").
        AddHint("d", "Delete Selected").
        SetInputHandler(v.handleInput)

    return v
}

func (v *UsersView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    switch event.Rune() {
    case ' ':
        v.table.ToggleSelection()
        return nil
    case 'a':
        v.table.SelectAll()
        return nil
    case 'd':
        selected := v.table.GetSelectedRows()
        if len(selected) > 0 {
            v.deleteUsers(selected)
        }
        return nil
    }
    return event
}
```

---

## Navigation

Tables support vim-style navigation by default:

| Key | Action |
|-----|--------|
| `j` / `Down` | Move down |
| `k` / `Up` | Move up |
| `g` | Go to top |
| `G` | Go to bottom |
| `Enter` | Select row |
| `Space` | Toggle selection (multi-select) |
