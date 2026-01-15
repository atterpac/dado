---
label: Getting Started
icon: rocket
order: 90
---

# Getting Started

Build your first TUI application with Jig in about 15 minutes.

---

## Installation

```bash
go get github.com/atterpac/jig
```

---

## Your First App

Create `main.go`:

```go
package main

import (
    "log"

    "github.com/atterpac/jig/components"
    "github.com/atterpac/jig/input"
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
    content := components.NewLabel("Hello, Jig!\n\nPress 'q' to quit.").
        SetAlign(components.AlignCenter)

    panel := components.NewPanel().
        SetTitle("Welcome").
        SetContent(content)

    v := &HomeView{
        panel: panel,
        app:   app,
    }

    handler := input.NewKeyBindings().
        Bind('q', func() { v.app.Stop() }).
        Build()

    // Wrap panel with ComponentBase to implement nav.Component
    v.ComponentBase = components.NewComponentBase(panel).
        SetName("home").
        AddHint("q", "Quit").
        SetInputHandler(handler)

    return v
}
```

Run it:

```bash
go run main.go
```

You should see a themed panel with "Hello, Jig!" centered.

---

## Adding Navigation

Let's add a second view and navigate between them.

### Create SettingsView

```go
type SettingsView struct {
    *components.ComponentBase
    panel *components.Panel
    app   *layout.App
}

func NewSettingsView(app *layout.App) *SettingsView {
    content := components.NewLabel("Settings Page\n\nPress Esc to go back.")

    panel := components.NewPanel().
        SetTitle("Settings").
        SetContent(content)

    v := &SettingsView{
        panel: panel,
        app:   app,
    }

    handler := input.NewKeyBindings().
        Bind(input.KeyEscape, func() { v.app.Pages().Pop() }).
        Build()

    v.ComponentBase = components.NewComponentBase(panel).
        SetName("settings").
        AddHint("Esc", "Back").
        SetInputHandler(handler)

    return v
}
```

### Update HomeView

```go
// Update the handler in NewHomeView
handler := input.NewKeyBindings().
    Bind('s', func() { v.app.Pages().Push(NewSettingsView(v.app)) }).
    Bind('q', func() { v.app.Stop() }).
    Build()

v.ComponentBase = components.NewComponentBase(panel).
    SetName("home").
    AddHint("s", "Settings").
    AddHint("q", "Quit").
    SetInputHandler(handler)
```

Now pressing `s` navigates to settings, and `Esc` returns home.

---

## Working with Forms

Use the FormBuilder for easy form creation:

```go
import "github.com/atterpac/jig/validators"

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
    AsFormModal("Edit User", 60, 15)

app.Pages().Push(modal)
```

---

## Async Data Loading

Load data without blocking the UI:

```go
import (
    "context"
    "time"

    "github.com/atterpac/jig/async"
    "github.com/atterpac/jig/input"
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

    handler := input.NewKeyBindings().
        Bind('r', func() { v.loadUsers() }).
        Bind(input.KeyEscape, func() { v.app.Pages().Pop() }).
        Build()

    v.ComponentBase = components.NewComponentBase(table).
        SetName("users").
        AddHint("r", "Refresh").
        AddHint("Esc", "Back").
        SetOnStart(v.loadUsers).
        SetInputHandler(handler)

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

## Theming

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
// ... 13+ themes available
```

### Runtime Theme Switching

```go
var availableThemes = []theme.Theme{
    themes.TokyoNight(),
    themes.Catppuccin(),
    themes.Dracula(),
    themes.Nord(),
}

var themeIndex int

func cycleTheme() {
    themeIndex = (themeIndex + 1) % len(availableThemes)
    theme.SetProvider(availableThemes[themeIndex])
    // UI updates automatically
}
```

---

## Data Binding

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
connectionStatus := binding.NewValue("Disconnected")

unsubscribe := connectionStatus.BindToWithDraw(func(s string) {
    statusLabel.SetText("Status: " + s)
})
defer unsubscribe()

// Now any Set() call updates the label
connectionStatus.Set("Connected")
```

---

## Keyboard Input Helpers

### KeyBindings Builder

```go
import "github.com/atterpac/jig/input"

// Using key constants
handler := input.NewKeyBindings().
    Bind('q', func() { app.Stop() }).
    Bind('r', func() { v.refresh() }).
    Bind(input.KeyEnter, func() { v.select() }).
    Bind(input.KeyEscape, func() { app.Pages().Pop() }).
    Bind(input.KeyCtrlS, func() { v.save() }).
    Build()

// Or use string-based bindings
handler := input.NewKeyBindings().
    BindString("q", func() { app.Stop() }).
    BindString("enter", func() { v.select() }).
    BindString("escape", func() { app.Pages().Pop() }).
    BindString("ctrl+s", func() { v.save() }).
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

## Complete Example

Here's a full working example combining all concepts:

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/atterpac/jig/async"
    "github.com/atterpac/jig/components"
    "github.com/atterpac/jig/input"
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
    layout *components.Layout
    stats  *components.Label
    app    *layout.App
}

func NewDashboard(app *layout.App) *Dashboard {
    stats := components.NewLabel("Loading stats...").
        SetAlign(components.AlignCenter)

    panel := components.NewPanel().
        SetTitle("Dashboard").
        SetContent(stats)

    layout := components.Column(panel)

    d := &Dashboard{
        layout: layout,
        stats:  stats,
        app:    app,
    }

    handler := input.NewKeyBindings().
        Bind('r', func() { d.loadStats() }).
        Bind('t', func() { d.cycleTheme() }).
        Bind('q', func() { d.app.Stop() }).
        Build()

    d.ComponentBase = components.NewComponentBase(layout).
        SetName("dashboard").
        AddHint("r", "Refresh").
        AddHint("t", "Theme").
        AddHint("q", "Quit").
        SetOnStart(d.loadStats).
        SetInputHandler(handler)

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

func (d *Dashboard) cycleTheme() {
    themeIndex = (themeIndex + 1) % len(themeList)
    theme.SetProvider(themeList[themeIndex])
}
```

---

## Next Steps

[!ref icon="package" text="Component Reference"](components/index.md)
[!ref icon="book" text="Architecture Guide"](guides/architecture.md)
[!ref icon="paintbrush" text="Theming"](guides/theming.md)
[!ref icon="tools" text="Troubleshooting"](guides/troubleshooting.md)
