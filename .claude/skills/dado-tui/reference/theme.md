# Theming System (`github.com/atterpac/dado/theme`)

Core principle: **read colors at draw time so live switching just works.** A `Provider` owns the active theme and its observers; package-level funcs forward to `Default()`.

---

## The Theme interface (color contract)

Every theme implements `theme.Theme`. Components read these each frame:

```go
type Theme interface {
    Bg(); BgLight(); BgDark() tcell.Color        // base backgrounds
    Fg(); FgDim(); FgMuted() tcell.Color         // foreground/text
    Accent(); AccentDim(); Highlight() tcell.Color
    Success(); Warning(); Error(); Info() tcell.Color // semantic
    Border(); BorderFocus() tcell.Color
    Header(); Menu(); TableHeader(); Key();
    Crumb(); PanelBorder(); PanelTitle() tcell.Color  // UI elements
}
```

## Provider & Default()

```go
func NewProvider() *Provider   // auto-refresh enabled
func Default() *Provider         // package singleton
func (p *Provider) SetTheme(t Theme)
func (p *Provider) Theme() Theme  // active theme, nil if unset
```
Active theme is stored in an `atomic.Value` for lock-free Draw-path reads. A separate `*Provider` scopes a different theme to a subtree (preview pane, modal, test fixture) via `component.SetTheme(p)`.

### Color accessors — full palette
Each exists on `*Provider` **and** as a package func forwarding to `Default()`:

| Group | Accessors |
|---|---|
| Base | `Bg`, `BgLight`, `BgDark`, `Fg`, `FgDim`, `FgMuted` |
| Accent | `Accent`, `AccentDim`, `Highlight` |
| Semantic | `Success`, `Warning`, `Error`, `Info` |
| Border | `Border`, `BorderFocus` |
| UI element | `Header`, `Menu`, `TableHeader`, `Key`, `Crumb`, `PanelBorder`, `PanelTitle` |

Each reads the active theme at call time with a fallback:
```go
func (p *Provider) Bg() tcell.Color { return p.colorOr(Theme.Bg, tcell.ColorDefault) }
```

**Tag variants** for tview color-tag strings: `TagBg()`, `TagFg()`, `TagAccent()`, … each returns `"#RRGGBB"`. Helper: `ColorToHex(tcell.Color) string`.

**Selection / list helpers:** `SelectionBg()` (=Accent), `SelectionFg()` (=black), `SelectionStyle()`, `InactiveSelectionStyle()`, and `ConfigureList`/`ConfigureListActive`/`ConfigureListInactive(*tview.List)`.

### Setting a theme
```go
theme.Default().SetTheme(t) // or back-compat theme.SetProvider(t)
```
`SetTheme`: stores the theme atomically; on the default provider applies global tview styles (`tview.Styles.*`); fires change callbacks; publishes `bus.KindThemeSwitch` if the bus is enabled; if auto-refresh is on, spawns a goroutine calling `QueueUpdateDraw` to update registered primitive backgrounds + call `RefreshTheme()`. The goroutine hop avoids deadlocking `QueueUpdateDraw` when called from inside a tview callback.

---

## Runtime switching — subscriptions

Three observer kinds, each returns an unregister `func()`:

```go
func (p *Provider) Register(prim Primitive) func()           // auto bg update on switch
func (p *Provider) RegisterRefreshable(r Refreshable) func()  // calls r.RefreshTheme()
func (p *Provider) OnThemeChange(fn func()) func()            // arbitrary callback (fires synchronously in SetTheme)
```
- `Primitive` = anything with `SetBackgroundColor(tcell.Color) *tview.Box`.
- Package forwarders: `theme.Register`, `theme.RegisterRefreshable`, `theme.OnThemeChange`, `Unregister`, `UnregisterRefreshable`.
- App wiring: `theme.SetApp(*tview.Application)`; `theme.QueueUpdate`/`QueueUpdateDraw` marshal onto the UI thread (fall back to immediate execution in test mode with no app).

**Best practice — Register vs Subscribe vs nothing:**
- **`Register(prim)`** — for a tview primitive whose **background** should follow `Bg()`. `widgetBase.initWidget` does exactly this. Store the unregister in `Subs()`.
- **`RegisterRefreshable(r)`** — when switch needs custom recolor/relayout (implement `RefreshTheme()`).
- **`OnThemeChange(fn)`** — arbitrary side effects.
- **Nothing** — if you read colors fresh in `Draw` via `th()`, foreground colors need no registration; the redraw queued by `SetTheme` re-runs Draw with new values.

Always add register results to a `components.Subscriptions` (`cb.Subs().Add(...)`/`w.subs.Add(...)`); `ComponentBase.Stop()` releases them.

**Built-in selector (downstream apps get switching free):** `app.EnableThemes(layout.ThemeOptions{Key, Themes, Names, Default, OnChange})` binds a key (default `Ctrl+T`) opening a live-preview `ThemeSelectorModal`. Preview/select call `theme.Default().SetTheme(...)`, auto-refreshing every registered primitive. `OnChange func(name string)` is for persistence.

---

## Status colors (`theme/status.go`)

> The stale `components/README.md` references `theme.StatusColor`/`HasStatus`/`RegisterStatus` — **these do not exist.** Use typed `*Status` handles.

```go
type ColorFunc func() tcell.Color
func DefineStatus(name string, colorFunc ColorFunc, icon string) *Status      // dynamic, theme-aware
func DefineStatusStatic(name string, color tcell.Color, icon string) *Status  // fixed color
```
A `Status` couples a name + dynamic color + Nerd Font icon; the color resolves at call time so it tracks switches:
```go
StatusRunning := theme.DefineStatus("Running", theme.Info, theme.IconRunning)
StatusRunning.Color()     // tcell.Color
StatusRunning.ColorTag()  // "#RRGGBB"
StatusRunning.Icon()      // glyph
StatusRunning.String()    // "<icon> Running"
StatusRunning.Format()    // "[#RRGGBB]<icon> Running[-]"
```
Built-ins: `StatusSuccess`, `StatusError`, `StatusWarning`, `StatusInfo`, `StatusPending`, `StatusRunning`. Config themes emit static statuses via `(*ConfigTheme).CreateStatuses()`.

---

## Icons (`theme/icons.go`)

Flat list of `const` string glyphs — reference by name, no functions:
```go
theme.IconCheck, theme.IconError, theme.IconWarning, theme.IconRunning, theme.IconFolder,
theme.IconSearch, theme.IconSettings, theme.IconTreeBranch, theme.IconBarFull, ...
```
Categories: status, navigation chevrons/arrows, tree/box-drawing, UI elements, connection, time, actions, domain (`IconWorkflow`, `IconActivity`), progress-bar blocks (`IconBarFull`/`IconBarEmpty`). Most are Nerd Font private-use codepoints; box/tree ones are plain Unicode.

---

## Theme YAML schema & loading

**Two paths:**

1. **Built-in (embedded → generated).** Source of truth: `theme/themes/defs/*.yaml`. `go generate ./...` runs `theme/themes/gen` to produce `themes_gen.go` declaring each theme as `theme.MustFromColors(theme.ThemeColors{...})` (e.g. `themes.TokyoNightNight`). Registry: `themes.All() map[string]theme.Theme`, `themes.Names()`, `themes.Get(name)`, `themes.Default()`, `const DefaultName = "tokyonight-night"`. 26 built-ins.

2. **Runtime config files** (`theme/loader.go`): `LoadFromFile(path)`, `LoadFromJSON`, `LoadFromYAML` → `*ConfigTheme`. YAML parsed by a dependency-free `parseSimpleYAML` (indentation-based; supports `name`, `colors:`, `statuses:`).

**Schema** (`ColorConfig`, snake_case):
- Required: `bg`, `fg`, `fg_dim`, `fg_muted`, `accent`, `highlight`, `success`, `warning`, `error`, `info`, `border`, `border_focus`.
- Optional (auto-derived): `bg_light` (lighten bg 10%), `bg_dark` (darken 10%), `accent_dim` (darken accent 20%), `header` (=bg_dark), `menu` (=bg), `table_header`/`crumb`/`panel_title` (=accent), `key` (=accent_dim), `panel_border` (=border).
- `statuses:` — list of `{name, color, icon}` → `DefineStatusStatic`.

```yaml
name: tokyonight-night
ident: TokyoNightNight   # Go var name (defs only)
default: true            # defs only
colors:
  bg: "#1a1b26"
  fg: "#c0caf5"
  accent: "#7aa2f7"
  # ...
```

---

## Programmatic builder & gradients

### Builder (`theme/builder.go`)
`theme.C(v any) Color` accepts `tcell.Color`, hex (`"#282828"`/`"282828"`), or `int`/`int32` hex.
```go
t, err := theme.FromColors(theme.ThemeColors{
    Bg: theme.C("#282828"), Fg: theme.C("#ebdbb2"), Accent: theme.C("#83a598"),
    Success: ..., Warning: ..., Error: ..., Info: ...,
})
t := theme.MustFromColors(...) // panics on error

t := theme.NewBuilder().Bg("#282828").Fg("#ebdbb2").Accent("#83a598").
    Success("#b8bb26").Warning("#fabd2f").Error("#fb4934").Info("#83a598").MustBuild()

variant := theme.FromTheme(existing) // clone for tweaks
```
Required builder fields: `Bg`, `Fg`, `Accent`, `Success`, `Warning`, `Error`, `Info`; optionals auto-derived.

### Gradients (`theme/gradient.go`)
For coloring ASCII art / banners with tview tags:
```go
type GradientType: GradientDiagonal | GradientHorizontal | GradientVertical | GradientReverseDiagonal
theme.DefaultGradientColors() // Accent→Success→FgDim
theme.AccentGradientColors()  // Accent→Highlight→Success
theme.InterpolateHex(h1, h2, ratio) string
theme.ApplyGradient(text, gradientType, colors) string
// + ApplyHorizontal/Vertical/Diagonal/ReverseDiagonalGradient
```

---

## Consuming theme in a component (best practice)

Read via the base-routed scoped accessor so per-subtree `SetTheme` overrides work:
```go
// widget: lock-free, on the Draw hot path
func (w *widgetBase) th() *theme.Provider {
    if p := w.themeP.Load(); p != nil { return p }
    return theme.Default()
}
// in Draw: bg := w.th().Bg()
```
`ComponentBase.Theme()` does the same. The package-level `theme.Bg()` always reads `Default()` — use the scoped accessor in components that may be themed independently.
