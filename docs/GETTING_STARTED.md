# Getting Started with Jig

Build your first TUI application with jig in about 15 minutes.

## 1. Installation

```bash
go get github.com/atterpac/jig
```

## 2. Your First App

Create `main.go`:

```go
package main

import (
    "log"

    "github.com/gdamore/tcell/v2"
    "github.com/rivo/tview"

    "github.com/atterpac/jig/components"
    "github.com/atterpac/jig/layout"
    "github.com/atterpac/jig/theme"
    "github.com/atterpac/jig/theme/themes"
)

func main() {
    // Set theme
    theme.SetProvider(themes.TokyoNight())

    // Create app with bottom menu bar
    app := layout.NewApp(layout.AppConfig{
        BottomBar: layout.NewMenu(),
    })

    // Create and push the home view
    app.Pages().Push(NewHomeView(app))

    // Run the application
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}

// HomeView is our main view
type HomeView struct {
    *components.ComponentBase
    panel *components.Panel
    app   *layout.App
}

func NewHomeView(app *layout.App) *HomeView {
    // Create a panel with welcome message
    content := tview.NewTextView().
        SetText("Hello, Jig!\n\nPress 'q' to quit.").
        SetTextAlign(tview.AlignCenter)

    panel := components.NewPanel().
        SetTitle("Welcome").
        SetContent(content)

    v := &HomeView{
        panel: panel,
        app:   app,
    }

    // Wrap panel with ComponentBase to implement nav.Component
    v.ComponentBase = components.NewComponentBase(panel).
        SetName("home").
        AddHint("q", "Quit").
        SetInputHandler(v.handleInput)

    return v
}

func (v *HomeView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    switch event.Rune() {
    case 'q':
        v.app.Stop()
        return nil // Event consumed
    }
    return event // Pass through
}
```

Run it:

```bash
go run main.go
```

You should see a themed panel with "Hello, Jig!" centered.

---

## 3. Adding Navigation

Let's add a second view and navigate between them.

```go
// Add SettingsView
type SettingsView struct {
    *components.ComponentBase
    panel *components.Panel
    app   *layout.App
}

func NewSettingsView(app *layout.App) *SettingsView {
    content := tview.NewTextView().
        SetText("Settings Page\n\nPress Esc to go back.")

    panel := components.NewPanel().
        SetTitle("Settings").
        SetContent(content)

    v := &SettingsView{
        panel: panel,
        app:   app,
    }

    v.ComponentBase = components.NewComponentBase(panel).
        SetName("settings").
        AddHint("Esc", "Back").
        SetInputHandler(v.handleInput)

    return v
}

func (v *SettingsView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    if event.Key() == tcell.KeyEscape {
        v.app.Pages().Pop()
        return nil
    }
    return event
}
```

Update HomeView to navigate:

```go
func (v *HomeView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    switch event.Rune() {
    case 's':
        v.app.Pages().Push(NewSettingsView(v.app))
        return nil
    case 'q':
        v.app.Stop()
        return nil
    }
    return event
}

// Update hints
v.ComponentBase = components.NewComponentBase(panel).
    SetName("home").
    AddHint("s", "Settings").
    AddHint("q", "Quit").
    SetInputHandler(v.handleInput)
```

Now pressing `s` navigates to settings, and `Esc` returns home.

---

## 4. Working with Forms

Use the FormBuilder for easy form creation:

```go
import (
    "github.com/atterpac/jig/validators"
)

type EditUserView struct {
    *components.ComponentBase
    form *components.Form
    app  *layout.App
}

func NewEditUserView(app *layout.App) *EditUserView {
    v := &EditUserView{app: app}

    form := components.NewFormBuilder().
        Text("name", "Name").
            Placeholder("Enter your name").
            Validate(validators.Required(), validators.MinLength(2)).
            Done().
        Text("email", "Email").
            Placeholder("user@example.com").
            Validate(validators.Required(), validators.Email()).
            Done().
        Select("role", "Role", []string{"Admin", "User", "Guest"}).
            Default("User").
            Done().
        Checkbox("notify", "Email notifications").
            Checked(true).
            Done().
        OnSubmit(func(values map[string]any) {
            // Save the user
            name := values["name"].(string)
            email := values["email"].(string)
            role := values["role"].(components.SelectOption)
            notify := values["notify"].(bool)

            log.Printf("Saved: %s, %s, %s, notify=%v", name, email, role.Value, notify)
            app.Pages().Pop()
        }).
        OnCancel(func() {
            app.Pages().Pop()
        }).
        Build()

    v.form = form
    v.ComponentBase = components.NewComponentBase(form).
        SetName("edit-user").
        SetHints(form.Hints())

    return v
}
```

### Form as Modal

Wrap a form in a modal:

```go
modal := components.NewFormBuilder().
    Text("name", "Name").
        Validate(validators.Required()).
        Done().
    OnSubmit(func(values map[string]any) {
        // Handle submit
    }).
    AsFormModal("Edit User", 60, 15)  // Doesn't dismiss on Escape

app.Pages().Push(modal)
```

---

## 5. Async Data Loading

Load data without blocking the UI:

```go
import (
    "context"
    "time"

    "github.com/atterpac/jig/async"
)

type User struct {
    Name  string
    Email string
    Role  string
}

type UsersView struct {
    *components.ComponentBase
    table *components.Table
    app   *layout.App
}

func NewUsersView(app *layout.App) *UsersView {
    table := components.NewTable()
    table.SetHeaders("Name", "Email", "Role")

    v := &UsersView{
        table: table,
        app:   app,
    }

    v.ComponentBase = components.NewComponentBase(table).
        SetName("users").
        AddHint("r", "Refresh").
        AddHint("Esc", "Back").
        SetOnStart(v.loadUsers).  // Load data when view becomes active
        SetInputHandler(v.handleInput)

    return v
}

func (v *UsersView) loadUsers() {
    async.NewLoader[[]User]().
        WithTimeout(10 * time.Second).
        WithIndicator(async.Toast("Loading users...")).
        OnSuccess(func(users []User) {
            v.table.Clear()
            for _, u := range users {
                v.table.AddRow(u.Name, u.Email, u.Role)
            }
        }).
        OnError(func(err error) {
            // Show error to user
            log.Printf("Failed to load users: %v", err)
        }).
        Run(func(ctx context.Context) ([]User, error) {
            // Simulate API call
            time.Sleep(1 * time.Second)
            return []User{
                {"Alice", "alice@example.com", "Admin"},
                {"Bob", "bob@example.com", "User"},
            }, nil
        })
}

func (v *UsersView) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    switch {
    case event.Rune() == 'r':
        v.loadUsers()
        return nil
    case event.Key() == tcell.KeyEscape:
        v.app.Pages().Pop()
        return nil
    }
    return event
}
```

### Simple One-Shot Loading

For simple cases:

```go
async.Load(
    func(ctx context.Context) ([]User, error) {
        return fetchUsers(ctx)
    },
    func(users []User) { v.setUsers(users) },
    func(err error) { v.showError(err) },
)
```

---

## 6. Theming

### Built-in Themes

```go
import "github.com/atterpac/jig/theme/themes"

// Available themes
theme.SetProvider(themes.TokyoNight())
theme.SetProvider(themes.Catppuccin())
theme.SetProvider(themes.CatppuccinMocha())
theme.SetProvider(themes.Dracula())
theme.SetProvider(themes.Nord())
theme.SetProvider(themes.Gruvbox())
theme.SetProvider(themes.GruvboxDark())
theme.SetProvider(themes.Solarized())
theme.SetProvider(themes.SolarizedDark())
theme.SetProvider(themes.Rosepine())
theme.SetProvider(themes.OneDark())
// ... 20+ themes available
```

### Runtime Theme Switching

```go
type ThemePickerView struct {
    *components.ComponentBase
    app        *layout.App
    themeIndex int
}

var availableThemes = []theme.Theme{
    themes.TokyoNight(),
    themes.Catppuccin(),
    themes.Dracula(),
    themes.Nord(),
}

func (v *ThemePickerView) cycleTheme() {
    v.themeIndex = (v.themeIndex + 1) % len(availableThemes)
    theme.SetProvider(availableThemes[v.themeIndex])
    // UI updates automatically - all registered components refresh
}
```

### Using Theme Colors

Read colors at draw time for live switching:

```go
func (c *MyComponent) Draw(screen tcell.Screen) {
    // Always read theme colors in Draw, not at construction
    style := tcell.StyleDefault.
        Background(theme.Bg()).
        Foreground(theme.Accent())

    // Draw with style...
}
```

### Register Components for Theme Updates

```go
func NewMyComponent() *MyComponent {
    box := tview.NewBox()

    // Auto-update background on theme change
    theme.Register(box)

    return &MyComponent{Box: box}
}
```

---

## 7. Data Binding

### Observable Values

```go
import "github.com/atterpac/jig/binding"

// Create observable
status := binding.NewValue("Ready")

// Subscribe to changes
unsubscribe := status.Subscribe(func(old, new string) {
    log.Printf("Status: %s -> %s", old, new)
})
defer unsubscribe()

// Update value
status.Set("Loading")

// Update and redraw
status.SetAndDraw("Complete")
```

### Bind to UI

```go
// One-way binding: value changes update UI automatically
connectionStatus := binding.NewValue("Disconnected")

unsubscribe := connectionStatus.BindToWithDraw(func(s string) {
    statusLabel.SetText("Status: " + s)
})
defer unsubscribe()

// Now any Set() call updates the label
connectionStatus.Set("Connected")
```

### Computed Values

```go
count := binding.NewValue(0)

// Derive a display value
displayText := binding.ComputedTo(count, func(n int) string {
    return fmt.Sprintf("Count: %d", n)
})

// displayText.Get() returns "Count: 0"
count.Set(5)
// displayText.Get() now returns "Count: 5"
```

---

## 8. Keyboard Input Helpers

### KeyBindings Builder

```go
import "github.com/atterpac/jig/input"

handler := input.NewKeyBindings().
    OnRune('q', func() { app.Stop() }).
    OnRune('r', func() { v.refresh() }).
    On(tcell.KeyEnter, func() { v.select() }).
    On(tcell.KeyEscape, func() { app.Pages().Pop() }).
    OnCtrlRune('s', func() { v.save() }).
    Build()

v.ComponentBase = components.NewComponentBase(content).
    SetInputHandler(handler)
```

### Vim Navigation

Add vim-style navigation to lists:

```go
// j/k to move, gg/G for top/bottom
input.VimListBindings(v.list)
```

---

## 9. Complete Example

Here's a full working example combining all concepts:

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/gdamore/tcell/v2"
    "github.com/rivo/tview"

    "github.com/atterpac/jig/async"
    "github.com/atterpac/jig/components"
    "github.com/atterpac/jig/layout"
    "github.com/atterpac/jig/theme"
    "github.com/atterpac/jig/theme/themes"
)

func main() {
    theme.SetProvider(themes.TokyoNight())

    app := layout.NewApp(layout.AppConfig{
        BottomBar:  layout.NewMenu(),
        ShowCrumbs: true,
    })

    app.Pages().Push(NewDashboard(app))

    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}

type Dashboard struct {
    *components.ComponentBase
    flex  *tview.Flex
    stats *tview.TextView
    app   *layout.App
}

func NewDashboard(app *layout.App) *Dashboard {
    stats := tview.NewTextView().
        SetText("Loading stats...").
        SetTextAlign(tview.AlignCenter)

    panel := components.NewPanel().
        SetTitle("Dashboard").
        SetContent(stats)

    d := &Dashboard{
        flex:  tview.NewFlex().AddItem(panel, 0, 1, true),
        stats: stats,
        app:   app,
    }

    d.ComponentBase = components.NewComponentBase(d.flex).
        SetName("dashboard").
        AddHint("r", "Refresh").
        AddHint("t", "Theme").
        AddHint("q", "Quit").
        SetOnStart(d.loadStats).
        SetInputHandler(d.handleInput)

    return d
}

func (d *Dashboard) loadStats() {
    async.NewLoader[string]().
        WithIndicator(async.Toast("Loading...")).
        OnSuccess(func(stats string) {
            d.stats.SetText(stats)
        }).
        Run(func(ctx context.Context) (string, error) {
            time.Sleep(500 * time.Millisecond)
            return "Users: 1,234\nActive: 567\nOrders: 89", nil
        })
}

var themeIndex = 0
var themeList = []theme.Theme{
    themes.TokyoNight(),
    themes.Catppuccin(),
    themes.Dracula(),
}

func (d *Dashboard) handleInput(event *tcell.EventKey, setFocus func(tview.Primitive)) *tcell.EventKey {
    switch event.Rune() {
    case 'r':
        d.loadStats()
        return nil
    case 't':
        themeIndex = (themeIndex + 1) % len(themeList)
        theme.SetProvider(themeList[themeIndex])
        return nil
    case 'q':
        d.app.Stop()
        return nil
    }
    return event
}
```

---

## Next Steps

- Read the [Architecture Guide](ARCHITECTURE.md) to understand threading and lifecycle
- Explore the [Component Reference](components/README.md) for all available components
- Run the interactive tutorial: `go run ./cmd/tutorial`
- Check [Troubleshooting](TROUBLESHOOTING.md) if you encounter issues
