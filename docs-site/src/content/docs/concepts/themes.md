---
title: Themes
description: How dado's theme system works and how to switch or customise themes.
---

dado ships with 26 built-in themes and a `theme.Provider` that resolves colours at
runtime. Every component reads from the active provider — switching the theme
refreshes every registered primitive and redraws the whole app instantly, with no
per-component wiring.

## How built-in themes are defined

Built-in themes are authored as YAML in `theme/themes/defs/*.yaml`. A code generator
compiles those defs into the type-safe Go registry at `theme/themes/themes_gen.go`.

```yaml
# theme/themes/defs/nord.yaml
name: nord
ident: Nord
colors:
  bg: "#2e3440"
  fg: "#d8dee9"
  accent: "#88c0d0"
  success: "#a3be8c"
  # … remaining colour keys
```

To add or change a built-in theme, edit a def and regenerate:

```sh
go generate ./...
```

Each def becomes an exported `theme.Theme` value in the `theme/themes` package
(`themes.Nord`, `themes.CatppuccinMocha`, …) plus registry helpers: `themes.All()`,
`themes.Get(name)`, `themes.Default()`, and `themes.DefaultName` (`"tokyonight-night"`).

## Built-in themes

| Theme | Variants |
|---|---|
| Catppuccin | Latte, Frappé, Macchiato, Mocha |
| Tokyo Night | Night, Storm, Day, Moon |
| Rosé Pine | Default, Moon, Dawn |
| Dracula | Default, Light |
| Gruvbox | Dark, Light |
| Everforest | Dark, Light |
| Solarized | Dark, Light |
| Nord | — |
| Kanagawa | — |
| Monokai | — |
| One | Dark, Light |
| GitHub | Dark, Light |

## Setting a theme

The built-in vars live in the `theme/themes` package. Apply one globally with
`theme.SetProvider`:

```go
import (
    "github.com/atterpac/dado/theme"
    "github.com/atterpac/dado/theme/themes"
)

theme.SetProvider(themes.Nord)
theme.SetProvider(themes.CatppuccinMocha)
theme.SetProvider(themes.TokyoNightStorm)
```

## The theme selector (recommended)

`App.EnableThemes` applies a default theme and wires a live-preview selector
(default key: **Ctrl+T**). Selecting a theme switches it app-wide and persists the
choice via the `OnChange` callback.

```go
app.EnableThemes(layout.ThemeOptions{
    Default:  "nord",            // name from themes.All()
    OnChange: persistThemeName,  // optional: save across runs
    // Key:    tcell.KeyCtrlT,   // override the selector key
    // Themes: customSet,        // override the offered set (defaults to themes.All())
})
```

## Custom themes

Three ways to build a `theme.Theme` at runtime, in increasing convenience:

**From a colour struct.** Only `Bg`, `Fg`, `Accent`, `Success`, `Warning`, `Error`,
and `Info` are required — the rest are derived (e.g. `BgLight` lightens `Bg`):

```go
mine := theme.MustFromColors(theme.ThemeColors{
    Bg:      theme.C("#1e1e2e"),
    Fg:      theme.C("#cdd6f4"),
    Accent:  theme.C("#89b4fa"),
    Success: theme.C("#a6e3a1"),
    Warning: theme.C("#f9e2af"),
    Error:   theme.C("#f38ba8"),
    Info:    theme.C("#89dceb"),
})
theme.SetProvider(mine)
```

**With the fluent builder**, optionally deriving from an existing theme:

```go
mine := theme.FromTheme(themes.TokyoNightNight).
    Accent("#ff7a93").
    BorderFocus("#ff7a93").
    MustBuild()
```

**From a config file** loaded at runtime — `LoadFromFile`, `LoadFromYAML`, and
`LoadFromJSON` return a `theme.Theme` you can pass straight to `SetProvider`:

```go
mine, err := theme.LoadFromFile("mytheme.yaml")
if err != nil {
    log.Fatal(err)
}
theme.SetProvider(mine)
```

`theme.C` accepts a hex string (`"#1e1e2e"`), a `tcell.Color`, or an `int32` hex value.

## Theme in components

During each `Draw`, components read the active colours through the package-level
helpers (`theme.Bg()`, `theme.Accent()`, …) or the resolved `theme.Theme`. These
read the default provider, so a `SetProvider` call takes effect on the next render
cycle. The `theme.Theme` interface exposes one method per colour role — `Bg()`,
`Fg()`, `Accent()`, `Success()`, `Border()`, and so on — all returning `tcell.Color`.
