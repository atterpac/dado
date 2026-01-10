---
label: List
icon: list-unordered
order: 88
---

# List

Simple list component with selection support and vim-style navigation.

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

list := components.NewList().
    AddItems("Option 1", "Option 2", "Option 3").
    SetOnSelect(func(index int, item components.ListItem) {
        log.Printf("Selected: %s", item.Text)
    })
```

---

## Adding Items

```go
// Add single items
list := components.NewList().
    AddItem("First").
    AddItem("Second").
    AddItem("Third")

// Add multiple at once
list := components.NewList().
    AddItems("Alpha", "Beta", "Gamma")

// Add with secondary text
list := components.NewList().
    AddItemWithSecondary("John Doe", "john@example.com").
    AddItemWithSecondary("Jane Smith", "jane@example.com")

// Add with reference data
list := components.NewList().
    AddItemWithRef("User 1", user1).
    AddItemWithRef("User 2", user2)
```

---

## ListItem Structure

```go
type ListItem struct {
    Text      string  // Main display text
    Secondary string  // Optional secondary text
    Reference any     // Optional reference data
}
```

---

## Events

```go
list := components.NewList().
    // Called when Enter is pressed
    SetOnSelect(func(index int, item components.ListItem) {
        openDetail(item.Reference)
    }).
    // Called when selection changes (navigation)
    SetOnChange(func(index int, item components.ListItem) {
        showPreview(item)
    })
```

---

## Methods

### Item Management

| Method | Description |
|--------|-------------|
| `AddItem(text)` | Add a simple item |
| `AddItems(...text)` | Add multiple items |
| `AddItemWithSecondary(text, secondary)` | Add item with subtitle |
| `AddItemWithRef(text, ref)` | Add item with reference data |
| `SetItems([]ListItem)` | Replace all items |
| `Clear()` | Remove all items |
| `GetItem(index)` | Get item at index |
| `GetItems()` | Get all items |
| `GetItemCount()` | Get number of items |

### Selection

| Method | Description |
|--------|-------------|
| `GetSelected()` | Get selected index and item |
| `SetSelected(index)` | Set selection by index |
| `SetOnSelect(handler)` | Handle Enter press |
| `SetOnChange(handler)` | Handle selection change |

### Navigation

| Method | Description |
|--------|-------------|
| `MoveUp()` | Move selection up |
| `MoveDown()` | Move selection down |
| `MoveToTop()` | Move to first item |
| `MoveToBottom()` | Move to last item |

### Display

| Method | Description |
|--------|-------------|
| `SetShowSecondary(bool)` | Show/hide secondary text |
| `SetWrapAround(bool)` | Enable wrap-around navigation |
| `SetHighlightFullLine(bool)` | Highlight full line |
| `Primitive()` | Get underlying `*tview.List` |

---

## Keyboard Navigation

Built-in vim-style navigation:

| Key | Action |
|-----|--------|
| `j` / `Down` | Move down |
| `k` / `Up` | Move up |
| `g` | Go to first item |
| `G` | Go to last item |
| `Enter` | Select item |

---

## Examples

### Simple Menu

```go
menu := components.NewList().
    AddItems("Dashboard", "Settings", "Help", "Quit").
    SetOnSelect(func(index int, item components.ListItem) {
        switch item.Text {
        case "Dashboard":
            app.Pages().Push(NewDashboard(app))
        case "Settings":
            app.Pages().Push(NewSettings(app))
        case "Help":
            app.Pages().Push(NewHelp(app))
        case "Quit":
            app.Stop()
        }
    })
```

### User List with Preview

```go
type UserListView struct {
    *components.ComponentBase
    list    *components.List
    preview *components.Label
}

func NewUserListView(app *layout.App, users []User) *UserListView {
    list := components.NewList()
    preview := components.NewLabel("Select a user")

    for _, u := range users {
        list.AddItemWithRef(u.Name, u)
    }

    list.SetOnChange(func(index int, item components.ListItem) {
        user := item.Reference.(User)
        preview.SetText(fmt.Sprintf("Email: %s\nRole: %s", user.Email, user.Role))
    })

    list.SetOnSelect(func(index int, item components.ListItem) {
        user := item.Reference.(User)
        app.Pages().Push(NewUserDetail(app, user))
    })

    layout := components.Row(
        list,
        preview,
    )

    v := &UserListView{list: list, preview: preview}
    v.ComponentBase = components.NewComponentBase(layout).
        SetName("users")

    return v
}
```

### Filtered List

```go
type FilteredList struct {
    list     *components.List
    allItems []components.ListItem
}

func (f *FilteredList) Filter(query string) {
    filtered := make([]components.ListItem, 0)
    for _, item := range f.allItems {
        if strings.Contains(strings.ToLower(item.Text), strings.ToLower(query)) {
            filtered = append(filtered, item)
        }
    }
    f.list.SetItems(filtered)
}
```

---

## Accessing the Primitive

For advanced tview features:

```go
list := components.NewList()
tviewList := list.Primitive()

// Use tview-specific features
tviewList.SetSelectedFocusOnly(true)
```
