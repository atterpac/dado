---
label: Theming
icon: paintbrush
order: 90
---

# Theming

Jig's theme system provides 13+ built-in themes with runtime switching support.

---

## Quick Start

```go
import (
    "github.com/atterpac/jig/theme"
    "github.com/atterpac/jig/theme/themes"
)

// Set theme at startup
theme.SetProvider(themes.TokyoNight())

// Switch theme at runtime
theme.SetProvider(themes.Dracula())
```

---

## Built-in Themes

| Theme | Function |
|-------|----------|
| Tokyo Night | `themes.TokyoNight()` |
| Catppuccin | `themes.Catppuccin()` |
| Catppuccin Mocha | `themes.CatppuccinMocha()` |
| Dracula | `themes.Dracula()` |
| Nord | `themes.Nord()` |
| Gruvbox | `themes.Gruvbox()` |
| Gruvbox Dark | `themes.GruvboxDark()` |
| Kanagawa | `themes.Kanagawa()` |
| Rosepine | `themes.Rosepine()` |
| Solarized | `themes.Solarized()` |
| Solarized Dark | `themes.SolarizedDark()` |
| OneDark | `themes.OneDark()` |
| Everforest | `themes.Everforest()` |
| GitHub | `themes.GitHub()` |
| Monokai | `themes.Monokai()` |

---

## Lock-Free Design

The theme system uses `atomic.Value` for lock-free reads:

```go
// In theme/provider.go
var activeTheme atomic.Value  // stores *themeHolder

// Safe to call from any goroutine, including Draw()
func Get() Theme {
    if holder := activeTheme.Load(); holder != nil {
        return holder.(*themeHolder).theme
    }
    return nil
}
```

!!!success Why Lock-Free Matters
`Draw()` is called frequently and must never block. Theme colors are read during every draw cycle. A mutex would cause deadlocks when `SetProvider()` is called during drawing.
!!!

---

## Reading Theme Colors

!!!danger Critical
Always read colors at draw time, not at creation time.
!!!

+++ Correct
```go
// CORRECT: Read at draw time
func (c *MyComponent) Draw(screen tcell.Screen) {
    style := tcell.StyleDefault.
        Background(theme.Bg()).
        Foreground(theme.Fg())
    // ... draw with style
}
```
+++ Incorrect
```go
// INCORRECT: Cache colors at creation
func NewMyComponent() *MyComponent {
    return &MyComponent{
        bgColor: theme.Bg(),  // Won't update on theme change!
    }
}
```
+++

---

## Theme Interface

All themes implement this interface:

```go
type Theme interface {
    // Base colors
    Bg() tcell.Color
    BgLight() tcell.Color
    BgDark() tcell.Color
    Fg() tcell.Color
    FgDim() tcell.Color
    FgMuted() tcell.Color

    // Accent colors
    Accent() tcell.Color
    AccentDim() tcell.Color
    Highlight() tcell.Color

    // Semantic colors
    Success() tcell.Color
    Warning() tcell.Color
    Error() tcell.Color
    Info() tcell.Color

    // Border colors
    Border() tcell.Color
    BorderFocus() tcell.Color

    // UI element colors
    Header() tcell.Color
    Menu() tcell.Color
    TableHeader() tcell.Color
    Key() tcell.Color
    Crumb() tcell.Color
    PanelBorder() tcell.Color
    PanelTitle() tcell.Color
}
```

---

## Registration Patterns

### Auto-Background Update

Automatically calls `SetBackgroundColor(theme.Bg())` on theme change:

```go
theme.Register(myBox)
```

### Custom Refresh Logic

Calls `myComponent.RefreshTheme()` on theme change:

```go
unregister := theme.RegisterRefreshable(myComponent)
defer unregister()  // Important: prevent memory leak
```

### Callback Subscription

```go
unsubscribe := theme.OnThemeChange(func() {
    // Custom handling
    updateStyles()
})
defer unsubscribe()
```

---

## Runtime Theme Switching

```go
var themeIndex int
var availableThemes = []theme.Theme{
    themes.TokyoNight(),
    themes.Catppuccin(),
    themes.Dracula(),
    themes.Nord(),
}

func cycleTheme() {
    themeIndex = (themeIndex + 1) % len(availableThemes)
    theme.SetProvider(availableThemes[themeIndex])
    // All registered components update automatically
}
```

### Theme Picker Example

```go
type ThemePicker struct {
    *components.ComponentBase
    list *tview.List
    app  *layout.App
}

func NewThemePicker(app *layout.App) *ThemePicker {
    list := tview.NewList()

    themeList := []struct {
        name  string
        theme theme.Theme
    }{
        {"Tokyo Night", themes.TokyoNight()},
        {"Catppuccin", themes.Catppuccin()},
        {"Dracula", themes.Dracula()},
        {"Nord", themes.Nord()},
        {"Gruvbox", themes.Gruvbox()},
    }

    for i, t := range themeList {
        idx := i
        list.AddItem(t.name, "", 0, func() {
            theme.SetProvider(themeList[idx].theme)
        })
    }

    panel := components.NewPanel().
        SetTitle("Select Theme").
        SetContent(list)

    v := &ThemePicker{list: list, app: app}
    v.ComponentBase = components.NewComponentBase(panel).
        SetName("theme-picker").
        AddHint("Enter", "Select").
        AddHint("Esc", "Close")

    return v
}
```

---

## Custom Theme Creation

### Using the Builder

```go
custom := theme.NewBuilder().
    SetBg(tcell.ColorBlack).
    SetBgLight(tcell.NewRGBColor(30, 30, 30)).
    SetBgDark(tcell.NewRGBColor(10, 10, 10)).
    SetFg(tcell.ColorWhite).
    SetFgDim(tcell.NewRGBColor(180, 180, 180)).
    SetFgMuted(tcell.NewRGBColor(100, 100, 100)).
    SetAccent(tcell.ColorBlue).
    SetSuccess(tcell.ColorGreen).
    SetWarning(tcell.ColorYellow).
    SetError(tcell.ColorRed).
    SetInfo(tcell.ColorCyan).
    Build()

theme.SetProvider(custom)
```

### From JSON File

```go
custom, err := theme.LoadFromFile("mytheme.json")
if err != nil {
    log.Fatal(err)
}
theme.SetProvider(custom)
```

JSON format:

```json
{
  "bg": "#1a1b26",
  "bgLight": "#24283b",
  "bgDark": "#16161e",
  "fg": "#c0caf5",
  "fgDim": "#a9b1d6",
  "fgMuted": "#565f89",
  "accent": "#7aa2f7",
  "success": "#9ece6a",
  "warning": "#e0af68",
  "error": "#f7768e",
  "info": "#7dcfff"
}
```

### Implement Interface Directly

```go
type MyTheme struct{}

func (t *MyTheme) Bg() tcell.Color        { return tcell.ColorBlack }
func (t *MyTheme) BgLight() tcell.Color   { return tcell.NewRGBColor(30, 30, 30) }
func (t *MyTheme) BgDark() tcell.Color    { return tcell.NewRGBColor(10, 10, 10) }
func (t *MyTheme) Fg() tcell.Color        { return tcell.ColorWhite }
func (t *MyTheme) FgDim() tcell.Color     { return tcell.NewRGBColor(180, 180, 180) }
func (t *MyTheme) FgMuted() tcell.Color   { return tcell.NewRGBColor(100, 100, 100) }
func (t *MyTheme) Accent() tcell.Color    { return tcell.ColorBlue }
func (t *MyTheme) AccentDim() tcell.Color { return tcell.NewRGBColor(80, 80, 200) }
// ... implement all methods

theme.SetProvider(&MyTheme{})
```

---

## Using Theme Colors

### In Components

```go
// Direct access
bgColor := theme.Bg()
fgColor := theme.Fg()
accentColor := theme.Accent()

// Apply to tview primitives
textView.SetBackgroundColor(theme.Bg())
textView.SetTextColor(theme.Fg())
```

### Semantic Colors

```go
// Use semantic colors for meaning
successLabel.SetTextColor(theme.Success())  // Green
errorLabel.SetTextColor(theme.Error())      // Red
warningLabel.SetTextColor(theme.Warning())  // Yellow
infoLabel.SetTextColor(theme.Info())        // Cyan
```

### In Custom Draw Methods

```go
func (c *CustomComponent) Draw(screen tcell.Screen) {
    // Read theme colors at draw time
    style := tcell.StyleDefault.
        Background(theme.Bg()).
        Foreground(theme.Accent())

    x, y, width, height := c.GetInnerRect()
    for i := 0; i < width; i++ {
        for j := 0; j < height; j++ {
            screen.SetContent(x+i, y+j, ' ', nil, style)
        }
    }
}
```
