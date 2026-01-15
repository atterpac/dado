# Component Reference

Quick reference for all jig components with usage examples.

## Component Categories

### Input Components

| Component | Description | Value Type |
|-----------|-------------|------------|
| [TextField](#textfield) | Single-line text input | `string` |
| [TextArea](#textarea) | Multi-line text input | `string` |
| [Checkbox](#checkbox) | Boolean toggle | `bool` |
| [Select](#select) | Dropdown selection | `SelectOption` |
| [MultiSelect](#multiselect) | Multi-option selection | `[]SelectOption` |
| [RadioGroup](#radiogroup) | Single selection from options | `string` |
| [Autocomplete](#autocomplete) | Text input with suggestions | `string` |

### Container Components

| Component | Description |
|-----------|-------------|
| [Panel](#panel) | Bordered container with title |
| [Modal](#modal) | Centered dialog overlay |
| [Tabs](#tabs) | Tabbed container |
| [Split](#split) | Resizable split panes |
| [MasterDetail](#masterdetail) | List-detail pattern |

### Display Components

| Component | Description |
|-----------|-------------|
| [Table](#table) | Data table with selection |
| [Tree](#tree) | Hierarchical tree view |
| [VirtualList](#virtuallist) | Virtualized scrolling list |
| [Progress](#progress) | Progress indicator |
| [Toast](#toast) | Notification messages |
| [DiffViewer](#diffviewer) | Side-by-side diff view |

### Navigation Components

| Component | Description |
|-----------|-------------|
| [ComponentBase](#componentbase) | Base wrapper for nav.Component |
| [KeyHintBar](#keyhintbar) | Keyboard shortcut display |

---

## Input Components

### TextField

Single-line text input with validation and placeholder support.

```go
import "github.com/atterpac/jig/components"
import "github.com/atterpac/jig/validators"

field := components.NewTextField("email").
    SetLabel("Email Address").
    SetPlaceholder("user@example.com").
    SetValidator(func(value string) error {
        return validators.Email()(value)
    }).
    SetOnChange(func(e *components.ChangeEvent[string]) {
        log.Printf("Changed: %s -> %s", e.OldValue, e.NewValue)
    }).
    SetOnSubmit(func(e *components.SubmitEvent) {
        log.Printf("Submitted: %s", e.Value)
    })

// Get/set value
value := field.Value()
field.SetValue("test@example.com")

// Validate
if err := field.Validate(); err != nil {
    log.Printf("Validation error: %v", err)
}
```

### TextArea

Multi-line text input.

```go
area := components.NewTextArea("description").
    SetLabel("Description").
    SetPlaceholder("Enter description...").
    SetMaxLines(10).
    SetOnChange(func(e *components.ChangeEvent[string]) {
        log.Printf("Text changed")
    })

text := area.Value()
area.SetValue("Multi-line\ntext content")
```

### Checkbox

Boolean toggle with label.

```go
checkbox := components.NewCheckbox("active").
    SetLabel("Enable notifications").
    SetChecked(true).
    SetOnChange(func(e *components.ChangeEvent[bool]) {
        log.Printf("Checked: %v", e.NewValue)
    })

isChecked := checkbox.Value()  // bool
checkbox.SetChecked(false)
```

### Select

Dropdown selection with single choice.

```go
// Simple string options
sel := components.NewSelect("role").
    SetLabel("Role").
    SetOptions([]string{"Admin", "User", "Guest"}).
    SetPlaceholder("Select a role").
    SetDefault("User").
    SetOnChange(func(e *components.ChangeEvent[components.SelectOption]) {
        log.Printf("Selected: %s (value: %s)", e.NewValue.Label, e.NewValue.Value)
    })

// Custom label/value pairs
sel := components.NewSelect("status").
    SetOptionsWithValues([]components.SelectOption{
        {Label: "Active", Value: "active"},
        {Label: "Inactive", Value: "inactive"},
        {Label: "Pending", Value: "pending"},
    })

option := sel.Value()  // SelectOption{Label, Value}
index := sel.SelectedIndex()
```

### MultiSelect

Multiple selection from options.

```go
multi := components.NewMultiSelect("tags").
    SetLabel("Tags").
    SetOptions([]string{"Bug", "Feature", "Enhancement", "Documentation"}).
    SetSelected([]int{0, 2}).  // Pre-select indices
    SetOnChange(func(e *components.ChangeEvent[[]components.SelectOption]) {
        for _, opt := range e.NewValue {
            log.Printf("Selected: %s", opt.Label)
        }
    })

options := multi.Values()         // []SelectOption
indices := multi.SelectedIndices()  // []int
```

### RadioGroup

Single selection from visible options.

```go
radio := components.NewRadioGroup("priority").
    SetLabel("Priority").
    SetOptions([]string{"Low", "Medium", "High", "Critical"}).
    SetSelected(1).  // Default to "Medium"
    SetOnChange(func(e *components.ChangeEvent[string]) {
        log.Printf("Priority: %s", e.NewValue)
    })

selected := radio.Value()  // "Medium"
index := radio.SelectedIndex()  // 1
```

---

## Container Components

### Panel

Bordered container with rounded corners and optional title.

```go
content := tview.NewTextView().SetText("Hello, World!")

panel := components.NewPanel().
    SetTitle("Welcome").
    SetTitleColor(theme.Accent()).
    SetTitleAlign(components.TitleAlignLeft).
    SetContent(content).
    SetFocused(true)  // Highlight border

// Access content
inner := panel.GetContent()
```

### Modal

Centered dialog overlay with configurable behavior.

```go
content := tview.NewTextView().SetText("Are you sure?")

modal := components.NewModal(components.ModalConfig{
    Title:     "Confirm",
    Width:     50,
    Height:    10,
    MinWidth:  40,
    MaxWidth:  80,
    Backdrop:  true,
}).
    SetContent(content).
    SetHints([]components.KeyHint{
        {Key: "Enter", Description: "Confirm"},
        {Key: "Esc", Description: "Cancel"},
    }).
    SetOnSubmit(func() {
        log.Println("Confirmed!")
    }).
    SetOnCancel(func() {
        log.Println("Cancelled")
    }).
    SetDismissOnEsc(true).
    SetBlockUntilDismissed(false)

// Use with Pages
app.Pages().Push(modal)
```

See [modal.md](modal.md) for advanced modal patterns.

### Tabs

Tabbed container for organizing content.

```go
tabs := components.NewTabs().
    AddTab("Overview", overviewContent).
    AddTab("Details", detailsContent).
    AddTab("Settings", settingsContent).
    SetActiveTab(0).
    SetOnTabChange(func(index int, name string) {
        log.Printf("Switched to tab: %s", name)
    })

// Navigation
tabs.NextTab()
tabs.PrevTab()
tabs.SetActiveTab(2)
```

### Split

Resizable split pane container.

```go
split := components.NewSplit().
    SetDirection(components.SplitHorizontal).
    SetPrimary(leftPanel).
    SetSecondary(rightPanel).
    SetRatio(0.3).  // 30% / 70% split
    SetMinSize(20). // Minimum panel size
    SetResizable(true)

// Toggle orientation
split.ToggleDirection()
```

### MasterDetail

List-detail navigation pattern.

```go
master := components.NewMasterDetail().
    SetMaster(listView).
    SetDetail(detailView).
    SetRatio(0.3).
    SetOnSelectionChange(func(index int) {
        loadDetail(index)
    })

// Focus management
master.FocusMaster()
master.FocusDetail()
```

---

## Display Components

### Table

Enhanced table with headers, selection, and styling.

```go
table := components.NewTable()
table.SetHeaders("Name", "Status", "Created")
table.SetMultiSelect(true)

// Add rows
table.AddRow("Alice", "Active", "2024-01-15")
table.AddRow("Bob", "Inactive", "2024-01-10")

// Add colored row
table.AddColoredRow(
    []string{"Charlie", "Pending", "2024-01-20"},
    []tcell.Color{0, theme.Warning(), 0},
)

// Add row with status
table.AddRowWithStatus(
    theme.StatusSuccess(),
    1,  // Status column
    "Dave", "Active", "2024-01-22",
)

// Selection handling
table.SetOnSelect(func(row int) {
    data := table.GetRowData(row)
    log.Printf("Selected: %v", data)
})

// Multi-select
table.ToggleSelection()   // Toggle current row
table.SelectAll()
table.ClearSelection()
rows := table.GetSelectedRows()

// Row manipulation
table.UpdateRow(0, "Alice Updated", "Active", "2024-01-16")
table.InsertRowAt(1, "New User", "Pending", "2024-01-25")
table.RemoveRowAt(2)
table.ClearRows()  // Keep headers
```

See [table.md](table.md) for advanced table features.

### Tree

Hierarchical tree view.

```go
tree := components.NewTree().
    SetRoot(components.NewTreeNode("Root").
        AddChild(components.NewTreeNode("Child 1").
            AddChild(components.NewTreeNode("Grandchild 1")).
            AddChild(components.NewTreeNode("Grandchild 2"))).
        AddChild(components.NewTreeNode("Child 2"))).
    SetOnSelect(func(node *components.TreeNode) {
        log.Printf("Selected: %s", node.Text())
    })

// Navigation
tree.ExpandAll()
tree.CollapseAll()
```

### VirtualList

Virtualized list for large datasets.

```go
list := components.NewVirtualList().
    SetItemCount(10000).
    SetItemHeight(1).
    SetDrawItem(func(index int, x, y, width int, screen tcell.Screen) {
        text := fmt.Sprintf("Item %d", index)
        // Draw item...
    }).
    SetOnSelect(func(index int) {
        log.Printf("Selected item: %d", index)
    })
```

### Progress

Progress indicator with percentage.

```go
progress := components.NewProgress().
    SetLabel("Downloading...").
    SetProgress(0.5).  // 50%
    SetShowPercentage(true)

// Update progress
progress.SetProgress(0.75)
progress.Complete()  // Sets to 100%
```

### Toast

Notification messages.

```go
// Show a toast (auto-dismisses)
components.ShowToast(app, "Operation complete!", 3*time.Second)

// Toast with styling
components.ShowToastWithStyle(app, "Error occurred", components.ToastError, 5*time.Second)
```

---

## Navigation Components

### ComponentBase

Wrapper to make any `tview.Primitive` implement `nav.Component`.

```go
type MyView struct {
    *components.ComponentBase
    table *components.Table
}

func NewMyView() *MyView {
    table := components.NewTable()

    v := &MyView{table: table}
    v.ComponentBase = components.NewComponentBase(table).
        SetName("my-view").
        AddHint("r", "Refresh").
        AddHint("Enter", "Select").
        SetOnStart(v.loadData).
        SetOnStop(v.cleanup).
        SetInputHandler(v.handleInput)

    return v
}

func (v *MyView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if event.Rune() == 'r' {
        v.loadData()
        return nil  // Consumed
    }
    return event  // Pass through
}

func (v *MyView) loadData() {
    // Load data when view becomes active
}

func (v *MyView) cleanup() {
    // Cleanup when view becomes inactive
}
```

### KeyHintBar

Displays keyboard shortcuts at bottom of screen.

```go
hintBar := components.NewKeyHintBar().
    SetHints([]components.KeyHint{
        {Key: "Enter", Description: "Select"},
        {Key: "r", Description: "Refresh"},
        {Key: "q", Description: "Quit"},
    })

// Used in layouts
flex := tview.NewFlex().SetDirection(tview.FlexRow)
flex.AddItem(content, 0, 1, true)
flex.AddItem(hintBar, 1, 0, false)
```

---

## Form Builder

Fluent API for building forms with validation.

```go
form := components.NewFormBuilder().
    Text("name", "Name").
        Placeholder("Enter name").
        Validate(validators.Required(), validators.MinLength(2)).
        Done().
    Text("email", "Email").
        Validate(validators.Required(), validators.Email()).
        Done().
    Select("role", "Role", []string{"Admin", "User", "Guest"}).
        Default("User").
        Done().
    Checkbox("notify", "Email notifications").
        Checked(true).
        Done().
    TextArea("bio", "Bio").
        Placeholder("Tell us about yourself").
        MaxLines(5).
        Done().
    OnSubmit(func(values map[string]any) {
        name := values["name"].(string)
        email := values["email"].(string)
        role := values["role"].(components.SelectOption)
        notify := values["notify"].(bool)
        bio := values["bio"].(string)

        log.Printf("Submitted: %s, %s, %s, %v, %s", name, email, role.Value, notify, bio)
    }).
    OnCancel(func() {
        app.Pages().Pop()
    }).
    Build()

// As modal
modal := components.NewFormBuilder().
    Text("name", "Name").Done().
    OnSubmit(func(values map[string]any) { /* ... */ }).
    AsFormModal("Edit User", 60, 20)
```

See [form.md](form.md) for form patterns and validation.

---

## Common Interfaces

### ValueProvider

Implemented by all input components:

```go
type ValueProvider[T any] interface {
    Value() T
    SetValue(T) ValueProvider[T]
    HasValue() bool
    Clear()
}
```

### Validatable

Implemented by components with validation:

```go
type Validatable interface {
    Validate() error
    HasError() bool
    GetError() string
}
```

### Named

Implemented by named components:

```go
type Named interface {
    GetName() string
}
```

---

## Next Steps

- [Form Patterns](form.md) - Validation, binding, custom fields
- [Table Patterns](table.md) - Key-based selection, status colors
- [Modal Patterns](modal.md) - Blocking modals, confirm dialogs
