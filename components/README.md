# Components Package

Core UI primitives that compose to build complex interfaces.

## Files

| File | Purpose |
|------|---------|
| `panel.go` | Rounded border container with optional title |
| `modal.go` | Configurable modal dialog base |
| `table.go` | Table wrapper with selection and styling |
| `hints.go` | Pill-style key hint bar |
| `empty.go` | Centered empty state display |

---

## panel.go

### Panel Component

A container with rounded borders and optional title. Delegates focus and input to content.

```go
type Panel struct {
    *tview.Box
    content    tview.Primitive
    title      string
    titleColor tcell.Color  // 0 means use theme default
}

func NewPanel() *Panel

// SetContent sets the inner content primitive
func (p *Panel) SetContent(content tview.Primitive) *Panel

// SetTitle sets the title displayed in the top border
func (p *Panel) SetTitle(title string) *Panel

// SetTitleColor overrides the title color (0 for theme default)
func (p *Panel) SetTitleColor(color tcell.Color) *Panel

// GetContent returns the inner content
func (p *Panel) GetContent() tview.Primitive
```

### Draw Implementation

```go
func (p *Panel) Draw(screen tcell.Screen) {
    // 1. Get current theme colors at draw time
    bgColor := theme.Bg()
    borderColor := theme.Border()
    titleColor := theme.Accent()  // or p.titleColor if set

    // 2. Draw rounded border
    // Top: ╭─── Title ───╮
    // Sides: │           │
    // Bottom: ╰──────────╯

    // 3. Draw content inside border
    if p.content != nil {
        // Set content rect to inner area (inset by 1 on each side)
        p.content.SetRect(x+1, y+1, width-2, height-2)
        p.content.Draw(screen)
    }
}
```

### Implementation Notes

From `internal/ui/panel.go`:
- Uses `tview.Box` as base for rect management
- Rounded corners: `╭ ╮ ╰ ╯`
- Title centered in top border with padding
- Delegates `Focus()`, `Blur()`, `HasFocus()`, `InputHandler()` to content
- Colors read at draw time from theme

---

## modal.go

### Modal Component

Configurable modal dialog with centered positioning, optional backdrop, and key hint bar.

```go
type ModalConfig struct {
    Title     string
    Width     int  // Fixed width (0 = use min/max)
    Height    int  // Fixed height (0 = use min/max)
    MinWidth  int
    MaxWidth  int
    MinHeight int
    MaxHeight int
    Backdrop  bool // Dark semi-transparent background
}

type Modal struct {
    *tview.Flex
    panel    *Panel
    hintBar  *KeyHintBar
    content  tview.Primitive
    config   ModalConfig
    onClose  func()
    onSubmit func()
    onCancel func()
}

func NewModal(config ModalConfig) *Modal

// SetContent sets the modal's main content
func (m *Modal) SetContent(content tview.Primitive) *Modal

// SetHints sets the key hints displayed at bottom
func (m *Modal) SetHints(hints []KeyHint) *Modal

// SetOnClose sets callback when modal closes
func (m *Modal) SetOnClose(fn func()) *Modal

// SetOnSubmit sets callback for submit action
func (m *Modal) SetOnSubmit(fn func()) *Modal

// SetOnCancel sets callback for cancel action
func (m *Modal) SetOnCancel(fn func()) *Modal

// Close triggers the close callback
func (m *Modal) Close()

// WrapInputHandler wraps custom handler with modal's base handler
func (m *Modal) WrapInputHandler(handler func(*tcell.EventKey) *tcell.EventKey) func(*tcell.EventKey, func(tview.Primitive))
```

### Layout Structure

```
Flex (row, centered)
└── Flex (column, centered)
    └── Panel (with title)
        └── Flex (column)
            ├── content (weight 0, flex)
            └── KeyHintBar (weight 1, fixed)
```

### Base Input Handler

```go
func (m *Modal) baseInputHandler(event *tcell.EventKey) *tcell.EventKey {
    switch event.Key() {
    case tcell.KeyEscape:
        m.Close()
        return nil
    }
    switch event.Rune() {
    case 'q':
        m.Close()
        return nil
    }
    return event
}
```

### Implementation Notes

From `internal/ui/modal_base.go`:
- Uses nested Flex for centering
- Backdrop draws semi-transparent dark overlay
- Dimensions calculated: `clamp(content, min, max)`
- Panel provides rounded border
- KeyHintBar at bottom shows available actions
- Fluent API for configuration

---

## table.go

### Table Component

Enhanced table wrapper with header support, selection, and theme-aware styling.

```go
type Table struct {
    *tview.Table
    headers            []string
    hasHeader          bool
    multiSelect        bool
    selectedRows       map[int]bool
    onSelect           func(row int)
    onSelectionChange  func(rows []int)
    actions            *ActionRegistry
}

func NewTable() *Table

// SetHeaders sets column headers (row 0, fixed)
func (t *Table) SetHeaders(headers ...string) *Table

// AddRow adds a data row with automatic coloring
func (t *Table) AddRow(cells ...string) *Table

// AddColoredRow adds a row with explicit color per cell
func (t *Table) AddColoredRow(cells []string, colors []tcell.Color) *Table

// AddStyledRow adds a row with full cell styling
func (t *Table) AddStyledRow(cells []TableCell) *Table

// ClearRows removes all data rows (keeps headers)
func (t *Table) ClearRows() *Table

// SetMultiSelect enables/disables multi-selection
func (t *Table) SetMultiSelect(enabled bool) *Table

// ToggleSelection toggles selection of current row
func (t *Table) ToggleSelection()

// SelectAll selects all data rows
func (t *Table) SelectAll()

// ClearSelection deselects all rows
func (t *Table) ClearSelection()

// GetSelectedRows returns indices of selected rows
func (t *Table) GetSelectedRows() []int

// SetOnSelect sets callback for row selection (Enter key)
func (t *Table) SetOnSelect(fn func(row int)) *Table

// SetOnSelectionChange sets callback when selection changes
func (t *Table) SetOnSelectionChange(fn func(rows []int)) *Table

// SetActions sets the action registry for key bindings
func (t *Table) SetActions(actions *ActionRegistry) *Table
```

### TableCell Type

```go
type TableCell struct {
    Text       string
    Color      tcell.Color  // 0 = auto-detect from text
    Align      int          // tview.AlignLeft, etc.
    Expansion  int
    MaxWidth   int
    Selectable bool
}
```

### Status Auto-Detection

```go
func (t *Table) detectCellColor(text string) tcell.Color {
    // Check if text matches a registered status
    if theme.HasStatus(text) {
        return theme.StatusColor(text)
    }
    return theme.Fg()
}
```

### Draw Override

```go
func (t *Table) Draw(screen tcell.Screen) {
    // Refresh header colors from theme
    if t.hasHeader {
        for col := 0; col < t.GetColumnCount(); col++ {
            cell := t.GetCell(0, col)
            cell.SetTextColor(theme.Accent())
            cell.SetBackgroundColor(theme.Bg())
        }
    }

    // Refresh selection highlight colors
    t.SetSelectedStyle(tcell.StyleDefault.
        Background(theme.Highlight()).
        Foreground(theme.Bg()))

    t.Table.Draw(screen)
}
```

### Implementation Notes

From `internal/ui/table.go`:
- Wraps `tview.Table` for enhanced functionality
- Header row is fixed (not selectable)
- Multi-select with space/ctrl+a
- Status text auto-coloring via registry
- Colors refreshed at draw time
- Selection state persisted across redraws

---

## hints.go

### KeyHintBar Component

Pill-style key hint display typically shown at bottom of modals/views.

```go
type KeyHint struct {
    Key         string  // e.g., "Enter", "Esc", "Space"
    Description string  // e.g., "Select", "Close", "Toggle"
}

type KeyHintBar struct {
    *tview.Box
    hints []KeyHint
}

func NewKeyHintBar() *KeyHintBar

// SetHints sets the hints to display
func (k *KeyHintBar) SetHints(hints []KeyHint) *KeyHintBar

// AddHint adds a single hint
func (k *KeyHintBar) AddHint(key, description string) *KeyHintBar

// Clear removes all hints
func (k *KeyHintBar) Clear() *KeyHintBar
```

### Rendering Format

```
[Key] Description   [Key] Description   [Key] Description
 ^^^                 ^^^                 ^^^
 Accent bg           Accent bg           Accent bg
```

### Draw Implementation

```go
func (k *KeyHintBar) Draw(screen tcell.Screen) {
    // 1. Calculate total width needed
    // 2. Center hints in available width
    // 3. For each hint:
    //    - Draw key with accent background, bg foreground
    //    - Draw space
    //    - Draw description with fg color
    //    - Draw separator spaces
}
```

### Implementation Notes

From `internal/ui/key_hint_bar.go`:
- Single line height
- Pills: `[Key]` with accent background
- Centered horizontally
- Spacing between hint groups
- Dynamic color from theme at draw time

---

## empty.go

### EmptyState Component

Centered display for empty/loading/error states with icon, title, and message.

```go
type EmptyState struct {
    *tview.Flex
    icon    string
    title   string
    message string
}

func NewEmptyState() *EmptyState

// SetIcon sets the icon (Nerd Font glyph)
func (e *EmptyState) SetIcon(icon string) *EmptyState

// SetTitle sets the main title text
func (e *EmptyState) SetTitle(title string) *EmptyState

// SetMessage sets the secondary message text
func (e *EmptyState) SetMessage(message string) *EmptyState

// Configure sets all fields at once
func (e *EmptyState) Configure(icon, title, message string) *EmptyState
```

### Layout

```
         (vertical centering)
            [Icon]
            Title
           Message
         (vertical centering)
```

### Draw Implementation

```go
func (e *EmptyState) Draw(screen tcell.Screen) {
    // Icon: large, accent color
    // Title: fg color, bold
    // Message: fg-dim color

    // All centered horizontally and vertically
}
```

### Factory Functions (Optional)

Apps can create their own factory functions:

```go
// Example app-specific factories (NOT in primitives)
func EmptyStateNoData() *EmptyState {
    return NewEmptyState().Configure(
        theme.IconFolder,
        "No Data",
        "There's nothing here yet",
    )
}
```

### Implementation Notes

From `internal/ui/empty_state.go`:
- Uses Flex for centering
- Three text views stacked vertically
- Colors applied at draw time
- Icon typically uses Nerd Font glyph
- Presets are app-specific (not in primitives)
