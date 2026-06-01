# Input, Navigation, Layout & Binding

Packages: `input` (keys/actions/commandbar), `nav` (page stack, breadcrumbs, modals), `layout` (app shell, menu, statusbar), `binding` (observable values, form/table binding).

---

## 1. KeyBindings (`input/keybindings.go`, `input/keys.go`)

Fluent input handler builder. `NewKeyBindings()` → `*KeyBindings`; all methods chain.

`type Handler func(event *tcell.EventKey) bool` (`true` = consumed).

```go
func (kb *KeyBindings) On(key tcell.Key, h Handler) *KeyBindings
func (kb *KeyBindings) OnRune(r rune, h Handler) *KeyBindings
func (kb *KeyBindings) OnRunes(runes string, h Handler) *KeyBindings
func (kb *KeyBindings) OnMod(key tcell.Key, mod tcell.ModMask, h Handler) *KeyBindings
func (kb *KeyBindings) OnCtrl/OnAlt/OnShift(key tcell.Key, h Handler) *KeyBindings
func (kb *KeyBindings) OnCtrlRune(r rune, h Handler) *KeyBindings // 'a'..'z' -> KeyCtrlA..Z
func (kb *KeyBindings) SetFallback(h Handler) *KeyBindings
```
`OnCtrlRune('s', h)` maps the letter to `tcell.KeyCtrlS` (tcell encodes Ctrl+letter as a distinct key). Simpler `Bind*` family (`keys.go`) wraps a `func()` (always consumes):
```go
func (kb *KeyBindings) Bind(key any, h func()) *KeyBindings           // Key, rune, or byte
func (kb *KeyBindings) BindString(keyStr string, h func()) *KeyBindings // "enter","ctrl+s","q","f5"
func (kb *KeyBindings) BindCtrl(r rune, h func()) *KeyBindings
```
`keys.go` re-exports `tcell.Key` as `input.Key` + constants (`input.KeyEnter`, `input.KeyCtrlS`) so callers needn't import tcell.

**Dispatch order:** modifier combos → specific keys → runes (only when `event.Key()==KeyRune`) → fallback. `Handle()` wraps dispatch and also publishes an input event to the bus.

**Build vs BuildBool — the key distinction:**
```go
func (kb *KeyBindings) Build() func(*tcell.EventKey, func(tview.Primitive)) *tcell.EventKey // tview: returns event (nil=consumed)
func (kb *KeyBindings) BuildBool() func(*tcell.EventKey, func(tview.Primitive)) bool        // dado ComponentBase: returns bool
```
- `Build()` → raw tview `SetInputHandler`/`SetInputCapture` (returning `nil` swallows the event).
- `BuildBool()` → dado `components.ComponentBase.SetInputHandler`.
- `BuildWithFocus(focusHandler)` — `Build()` variant exposing tview's `setFocus`.

Composition: `Merge(other)` (other wins on conflicts), `Clone()`.

---

## 2. Vim navigation (`input/vim.go`)

```go
type VimNavigator struct {
    Up, Down, Left, Right func() // k/↑, j/↓, h/←, l/→
    Top, Bottom           func() // Home / G,End
    PageUp, PageDown      func() // Ctrl+U,PgUp / Ctrl+D,PgDn
    Select                func() // Enter, Space
    Back                  func() // Escape, q
}
func (kb *KeyBindings) AddVimNavigation(nav VimNavigator) *KeyBindings // binds only non-nil callbacks
```
List helper: `VimListBindings(list ListNavigator) *KeyBindings` (binds Up/Down/Top/Bottom from a `ListNavigator` with `MoveUp/MoveDown/MoveToTop/MoveToBottom`), then chain extras.

`gg` is NOT bound by `AddVimNavigation` (Top = Home only). Use `AddGG(onGG)` or standalone `GGHandler(onGG) (handler, reset)`.

> Note: `nav/navigator.go` has a *different*, richer `ListNavigator` (adds `GetSelectedIndex/SetSelectedIndex/GetItemCount`) with `TableNavigator`/`TextViewNavigator` adapters and `NavigationInputHandler(nav)` (bakes in j/k/g/G/arrows/PgUp/PgDn). Two interfaces named `ListNavigator` — don't conflate.

---

## 3. ActionRegistry (`input/action.go`)

Named, ordered key→handler registry that also produces display hints.
```go
type KeyAction struct { Key tcell.Key; Rune rune; Modifiers tcell.ModMask; Description string; Handler func() }
func NewActionRegistry() *ActionRegistry
func (r *ActionRegistry) AddSimple(name string, key rune, desc string, h func()) *ActionRegistry
func (r *ActionRegistry) AddKey(name string, key tcell.Key, desc string, h func()) *ActionRegistry
func (r *ActionRegistry) AddCtrl(name string, key rune, desc string, h func()) *ActionRegistry
func (r *ActionRegistry) Handle(event *tcell.EventKey) bool
func (r *ActionRegistry) Hints() []components.KeyHint // insertion order
```
Plus `Add/Remove/Clear/Merge/Has/Get`, `FormatKey(action)`.

**When vs raw KeyBindings:** use `ActionRegistry` when each binding needs a **name + description** that feeds the menu/status hint bar (pairs with `components.KeyHint` and `Menu.SetHints`). Use raw `KeyBindings` for behavior-only (ordered specificity, event access, fallback, merge/clone, vim helpers).

---

## 4. CommandBar (`input/commandbar.go`)

K9s-style single-line command/filter input (draws itself on a `tview.Box`). Multiple command types each with a prompt glyph + placeholder:
```go
const ( CommandTypeFilter; CommandTypeAction; CommandTypeSearch; CommandTypeCustom )
```
Defaults: Filter `/`, Action `:`, Search `?`. Lifecycle: `Show(cmdType)` / `Hide()`. Callbacks: `SetOnSubmit(func(cmdType, input))`, `SetOnCancel(func())`, `SetOnChange(func(input))`. Full line editor: cursor moves, Home/End (Ctrl+A/E), Backspace/Delete, Ctrl+U/K/W. `GetPreferredHeight()` = 1.

> For completion + history, `layout.StatusBar`'s command mode is the heavier alternative (built on a real `tview.InputField`, with Tab completion popups, ghost-text suggestions, history).

---

## 5. nav — routing & page stack

**Component interface** (`nav/component.go`) — everything pushed implements:
```go
type Component interface {
    tview.Primitive
    Name() string                // breadcrumb label
    Start(); Stop()              // active / inactive lifecycle
    Hints() []components.KeyHint  // feeds the menu
}
```

**Pages** (`nav/pages.go`) — stack router embedding `*tview.Pages`:
```go
func NewPages() *Pages
func (p *Pages) Push(c Component)    // Stop() prev, AddPage, Start() new, update crumbs, publish bus
func (p *Pages) Pop() bool           // false if depth<=1 or modal OnDismiss() returns false
func (p *Pages) Replace(c Component) // swap current at same depth
func (p *Pages) Current() Component
func (p *Pages) Clear() / StackDepth() / CanPop() / GetStack()
func (p *Pages) SetOnChange(func(Component))
func (p *Pages) SetCrumbs(*Crumbs)         // auto-updates breadcrumbs (modals excluded)
func (p *Pages) SetApplication(*tview.Application) // for focus restore
```
A **stack push/pop model**, not URL routes. Every change rebuilds the breadcrumb path from non-modal entries, fires `onChange`, publishes `KindNavPush/Pop/Replace`.

**Modals** (`nav/modal.go`): a `Modal` is a `Component` + `ModalBehavior() components.ModalBehavior` + `OnDismiss() bool`. Helpers `IsModal`, `AsModal`, `GetModalBehavior`. Pages specifics:
- On `Push` of a modal, current focus is saved to `focusStack`.
- `Pop` of a modal calls `OnDismiss()` (`false` cancels — e.g. unsaved-changes guard), fires `onModalDismiss`, restores focus if `behavior.RestoreFocusOnDismiss`.
- `hasBlockingModal()` (`behavior.BlockUntilDismissed`) blocks pushing/replacing.
- `DismissModal()`, `CurrentIsModal()`, `CurrentModalBehavior()`, `HasModal()`, `ModalCount()`.
- Esc-to-dismiss wiring lives in `layout/app.go` (checks `behavior.DismissOnEsc`).

**Breadcrumbs** (`nav/crumbs.go`): `*Crumbs` embeds `*tview.TextView`. `SetPath([]string)`/`Push`/`Pop`/`Clear`/`SetSeparator` (default `" > "`). Last crumb accent, earlier ones dim. Driven automatically once `SetCrumbs` is set.

---

## 6. layout — app shell

**`AppConfig`** (`layout/app.go`): `TopBar` (+`TopBarHeight`, default 3), `ShowCrumbs`, `BottomBar` (+`BottomBarHeight`, default 1), `OnComponentChange`, `Debug`/`DebugKey` (default Ctrl+D → `DebugOverlay`), `EffectShutdownTimeout`.

`NewApp(config)` builds a vertical root Flex, a `tview.Application`, a `nav.Pages`, and an `effect.Dispatcher`; registers the flex with `theme`; calls `theme.SetApp(app)` and `pages.SetApplication(app)`. Layout rows: TopBar → Crumbs → Pages (focused main area) → BottomBar. If `BottomBar` is a `*Menu`, hints auto-update. Global `SetInputCapture` handles Debug key, theme-selector key, and modal auto-dismiss (Esc via a deferred `QueueUpdateDraw` goroutine to dodge tview deadlocks), then delegates to user capture.

- `Run()` → `app.Run()`.
- `Stop()` releases all theme subs, shuts down the effect dispatcher (bounded by `EffectShutdownTimeout`; negative = fire-and-forget), then `app.Stop()`.
- Accessors: `Pages()`, `Menu()`, `Crumbs()`, `TopBar()`, `Effects()`, `GetApplication()`, `SetTopBar`/`SetBottomBar`, `SetFocus`, `QueueUpdate`/`QueueUpdateDraw`/`Draw`, `SetInputCapture`, `Suspend(fn)` (external commands), `ShowModal(*components.Modal)`, `UpdateMenuHints`.
- `EnableThemes(ThemeOptions{...})` → Ctrl+T live theme picker.

**Menu** (`layout/menu.go`): bottom bar. `SetHints([]KeyHint)`/`AddHint(key,desc)`/`SetRightText`/`Clear`. Pill-styled keys + descriptions; height 1.

**StatusBar** (`layout/statusbar.go`): `components.Panel` + `TextView`. Left/right `StatusSection`s (`Icon`, `Text`, static `Color` or `ColorFunc`). `AddSection`/`SetSections`/`SetRightSections`/`UpdateSection`, `SetConnectionStatus(connected,name)`. Full command mode (`EnterCommandMode`/`ExitCommandMode`) with Tab completion popup, ghost-text suggestion, history. Height 3.

All these expose `Subs()` so the App releases their theme registrations on teardown.

---

## 7. binding — observable values & data binding

**`Value[T]`** (`binding/value.go`) — thread-safe observable cell:
```go
func NewValue[T any](initial T) *Value[T]
func (v *Value[T]) Get() T
func (v *Value[T]) Set(newVal T)        // notifies + publishes to bus
func (v *Value[T]) SetAndDraw(newVal T) // Set + QueueUpdateDraw, skipped if DeepEqual unchanged
func (v *Value[T]) Update(fn func(T) T) / UpdateAndDraw(...)
func (v *Value[T]) Subscribe(fn func(old, new T)) func() // returns unsubscribe
func (v *Value[T]) BindTo(setter func(T)) func()         // one-way; sets now + on change (QueueUpdate)
func (v *Value[T]) BindToWithDraw(setter func(T)) func()
func (v *Value[T]) Computed(compute func(T) T) *Value[T]
func ComputedTo[T, U any](src *Value[T], compute func(T) U) *Value[U] // derived, different type
```
Use for reactive scalar state; `*AndDraw` variants are the "update state and re-render" path. Store the unsubscribe and call it on teardown.

**`FormBinding[T]`** (`binding/form.go`) — two-way struct↔Form via `form:"name"` tags:
```go
func NewFormBinding[T any](form *components.Form, target *T) *FormBinding[T]
func (fb *FormBinding[T]) Sync() error              // struct -> form fields
func (fb *FormBinding[T]) Collect() error           // form -> struct (runs validators)
func (fb *FormBinding[T]) CollectAndNotify() error
// Validate / Clear / SetValidator(field, fn) / SetOnChange / SetTarget
```
`Sync` type-switches field kinds (TextField/TextArea/Checkbox/Select/RadioGroup/MultiSelect) writing struct values in; `Collect` reads `form.GetValues()` back, runs per-field validators (`func(any) error`), assigns via reflection. Helpers (`helpers.go`): `RequiredValidator`, `MinLengthValidator`, `MaxLengthValidator`, `ChainValidators`. `BindFormToStruct(form, target, opts)` convenience (AutoCollect is a stub — collect manually in submit).

**`TableBinding[T]`** (`binding/table.go`) — binds `[]T` to a `components.Table`:
```go
func NewTableBinding[T any](table *components.Table) *TableBinding[T]
SetMapper(func(T) []string)            // required: item -> row cells
SetColorMapper(func(T) []tcell.Color)
SetKeyMapper(func(T) string)           // stable identity for Update/Remove/selection
SetFilter(func(T, string) bool)
SetFetcher(func() ([]T, error)) / SetRefreshInterval(d) / SetOnRefresh(...)
SetOnSelect(func(T))
// data ops: SetData / AddItem / UpdateItem / RemoveItem(key) / Filter(q) / ClearFilter
// lifecycle: Refresh() / RefreshAsync() / Start() / Stop()
// access: GetSelectedValue / GetSelectedItems / GetItem(row) / Count / TotalCount
```
`Start()` does an initial async fetch and (if interval>0) spins a ticker calling `RefreshAsync`; `Stop()` halts it. `applyFilter` rebuilds then `theme.QueueUpdateDraw(rebuildTable)`. Data row = table row − 1 (header). Convenience: `BindTableToSlice`, `BindTableWithKey`; filters `DefaultStringFilter`, `FieldFilter`, `ExactMatchFilter`, `PrefixFilter`.
