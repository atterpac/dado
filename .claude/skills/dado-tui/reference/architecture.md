# Component Architecture

The `components` package builds a structured layer over `rivo/tview` + `gdamore/tcell`. Two orthogonal concerns:

1. **View/navigation bases** — `ComponentBase`, `StatefulComponentBase[T]` wrap a finished primitive and add a `nav.Component` lifecycle.
2. **Leaf-widget base** — `widgetBase` is embedded into a hand-rolled widget that draws itself on a `tview.Box`.

Theme reads happen **at draw time**; all teardown funnels through a `Subscriptions` set released on `Stop()`.

---

## ComponentBase (`components/base.go`)

Wraps a `tview.Primitive` by **composition** (a field, not embedding) so callers keep type-safe access to the concrete primitive. Implements `tview.Primitive` (by delegation) and `nav.Component`.

```go
type ComponentBase struct {
    mu        sync.RWMutex
    primitive tview.Primitive
    name      string
    id        uint64        // unique, from atomic counter
    hints     []KeyHint
    onStart   func()
    onStop    func()
    subs      Subscriptions
    inputHandler func(*tcell.EventKey, func(tview.Primitive)) bool // optional
    drawOverlay  func(screen tcell.Screen)                          // optional
    themeP       *theme.Provider                                    // optional scope
}

func NewComponentBase(p tview.Primitive) *ComponentBase
```

### Fluent API (each returns `*ComponentBase`)
- `SetName(string)` / `Name() string`, `ID() uint64`
- `SetHints([]KeyHint)`, `AddHint(key, desc string)` — `Hints()` returns a defensive copy.
- `SetOnStart(func())`, `SetOnStop(func())`
- `SetInputHandler(func(*tcell.EventKey, func(tview.Primitive)) bool)` — **bool-return**: `true` consumes, `false` delegates to the wrapped primitive.
- `SetDrawOverlay(func(screen))` — runs *after* the primitive draws.
- `SetTheme(*theme.Provider)` — scopes a provider to this subtree.

### Accessors
- `Primitive() tview.Primitive`
- `Typed[P tview.Primitive](cb) P` — generic cast of the stored primitive (`tbl := components.Typed[*Table](cb)`). Panics on mismatch (programmer error).
- `Subs() *Subscriptions` — register theme/binding unsubscribers here.
- `Theme() *theme.Provider` — the scoped provider, else `theme.Default()`.

### Lifecycle (the `nav.Component` contract)
```go
func (cb *ComponentBase) Start() // calls onStart if set
func (cb *ComponentBase) Stop()  // onStop, then release subscriptions
```
`Stop()` ordering (tested in `subscriptions_test.go`):
1. Run user `onStop`.
2. If the wrapped primitive exposes `Subs() *Subscriptions`, release **its** subs too (tears down theme hooks attached inside a leaf widget's constructor).
3. Release `cb.subs`.

Nil-safe; config fields guarded by `mu`. `Draw`, `GetRect`/`SetRect`, `Focus`/`Blur`/`HasFocus`, `MouseHandler`, `PasteHandler`, `InputHandler` all delegate to the wrapped primitive.

---

## widgetBase (`components/widget_base.go`) — for authoring leaf widgets

Embedded **by value** into a concrete widget. Bundles box plumbing, a state mutex, a `Subscriptions`, and a lock-free scoped theme pointer.

```go
type widgetBase struct {
    *tview.Box
    mu     sync.RWMutex
    subs   Subscriptions
    themeP atomic.Pointer[theme.Provider] // read on the Draw hot path
}

func (w *widgetBase) initWidget(box *tview.Box) // call in constructor
func (w *widgetBase) Subs() *Subscriptions
func (w *widgetBase) SetTheme(p *theme.Provider)
func (w *widgetBase) th() *theme.Provider // scoped, else theme.Default()
```

`initWidget` sets the box background and registers it for auto-background updates, storing the unregister in `subs`:
```go
func (w *widgetBase) initWidget(box *tview.Box) {
    w.Box = box
    box.SetBackgroundColor(theme.Bg())
    w.subs.Add(theme.Register(box))
}
```

**Why `themeP` is atomic, not under `mu`:** Draw methods hold `mu` for their own state and then call `th()`; routing the theme read through `mu` would self-deadlock. Theme reads are lock-free.

> `Layout` and `Label` wrap `*tview.Flex`/`*tview.TextView` rather than `tview.Box`, so they don't embed `widgetBase`. They replicate the convention by hand: a `subs Subscriptions` field, a `Subs()` accessor, and `subs.Add(theme.Register(prim))` in the constructor.

---

## StatefulComponentBase[T] (`components/stateful_base.go`)

Embeds `*ComponentBase` and adds typed async-data state. Use for a view backed by a fetch (loading spinner → error → render).

```go
type StatefulComponentBase[T any] struct {
    *ComponentBase
    // data T, loadState LoadState (Idle|Loading|Error|Success), err error, callbacks
}

func NewStatefulComponentBase[T any](p tview.Primitive) *StatefulComponentBase[T]
```

State API (fluent; callbacks fire while NOT holding the lock): `SetData(T)`→Success, `SetLoadState`, `SetError(err)`→Error, `Reset()`→Idle, `UpdateData(func(T) T)`, plus `Data()`, `LoadState()`, `IsLoading()`, `HasError()`, `IsReady()`, `Error()`, `SetOnStateChange`, `SetOnDataChange`. Re-declares the fluent config setters so they chain returning `*StatefulComponentBase[T]`.

---

## Subscriptions (`components/base.go`)

The teardown aggregator. Collect every unregister func here.

```go
type Subscriptions struct { /* mu, funcs */ }
func (s *Subscriptions) Add(unsub func()) // nil ignored
func (s *Subscriptions) Release()          // LIFO, idempotent, clears list
func (s *Subscriptions) Len() int
```

---

## The Draw split: prepare + paint

For widgets that **mutate state during Draw** (scrolling/virtualization: fit viewport, clamp cursor, prefetch, recompute column widths), split Draw into a state-mutating `prepare` that returns an immutable **snapshot**, and a pure `paint(screen, rect, snap)`. Simpler widgets (Badge, Chip, Divider, StatusBar) don't need it — single `Draw` with an `RLock`.

```go
func (v *VirtualList) Draw(screen tcell.Screen) {
    v.Box.DrawForSubclass(screen, v)
    x, y, width, height := v.GetInnerRect()
    if width <= 0 || height <= 0 || v.totalCount == 0 { return }
    snap := v.prepareDraw(width, height)        // mutates state, reads theme ONCE
    v.paint(screen, x, y, width, height, snap)  // writes screen only
}

type vlSnapshot struct {
    contentWidth, indexWidth int
    bg, fg, fgDim, accent    tcell.Color // resolved theme colors
}
```

DataGrid uses `*Locked` variants (`prepareDrawLocked`/`paintLocked`) that run while `dg.mu` is held the whole time. The snapshot carries resolved geometry + resolved colors so paint is a pure function — golden snapshot tests (`draw_snapshot_test.go`) pin exact rune+fg+bg per cell, and colors are read once per frame rather than per cell.

---

## Value / event / handler interfaces

### Value interfaces (`components/value.go`)
```go
type ValueProvider[T any] interface {
    Value() T
    SetValue(value T) ValueProvider[T] // fluent
    HasValue() bool
    Clear()
}
type IndexedValueProvider[T any] interface { // Select, RadioGroup
    ValueProvider[T]; SelectedIndex() int; SetSelectedIndex(int) error
}
type MultiValueProvider[T any] interface { Values() []T; SetValues([]T) error /*...*/ } // MultiSelect

type Validatable interface { Validate() error; HasError() bool; GetError() string }
type Focusable interface { HasFocus() bool }
type Named interface { GetName() string }
```
These give form fields a uniform value/validation contract so `FormBuilder`/`binding` can treat heterogeneous fields uniformly.

### Events (`components/events.go`) and handlers (`components/handlers.go`)
A typed hierarchy over `BaseEvent` (type + source + id): `ChangeEvent[T]`, `SubmitEvent`, `CancelEvent`, `FocusEvent`, `SelectEvent[T]`, `ActivateEvent`, `KeyEvent`. Handler typedefs `ChangeHandler[T]`, `SubmitHandler`, …, `GenericHandler`. `BaseEventEmitter` provides thread-safe `OnEvent`/`EmitEvent` (snapshots handlers under `RLock` before dispatch).

### Two input-handler conventions (intentional)
- **tview-native:** `InputHandler() func(*tcell.EventKey, func(tview.Primitive))`; tview's own pre-processing uses `*tcell.EventKey`-return.
- **dado `ComponentBase.SetInputHandler`:** `... bool` — `true` consumed, `false` delegate. To rebind keys before delegating, mutate the event in place and return `false`.
- **`WrapInputHandler`** (on `Modal`, `Drawer`, `Table`, `TextField`, …) runs the widget's base handler first; if it consumes, returns; else runs the custom handler:
```go
func (m *Modal) WrapInputHandler(handler func(*tcell.EventKey, func(tview.Primitive))) func(*tcell.EventKey, func(tview.Primitive)) {
    return func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
        if m.handleBaseInput(event) { return }
        handler(event, setFocus)
    }
}
```

---

## Canonical recipe: authoring a NEW component

### A. A self-drawing leaf widget (common case)
Embed `widgetBase` by value, call `initWidget`, guard state with `w.mu`, read theme via `w.th()` inside `Draw`.

```go
type Badge struct {
    widgetBase
    text    string
    variant BadgeVariant
}

func NewBadge(text string) *Badge {
    b := &Badge{text: text, variant: BadgeDefault}
    b.initWidget(tview.NewBox()) // sets bg, registers theme auto-update into b.subs
    return b
}

func (b *Badge) SetVariant(v BadgeVariant) *Badge { // fluent setters lock + return self
    b.mu.Lock(); defer b.mu.Unlock()
    b.variant = v
    return b
}

func (b *Badge) Draw(screen tcell.Screen) {
    b.Box.DrawForSubclass(screen, b)          // 1. tview draws box/border
    x, y, width, height := b.GetInnerRect()   // 2. inner rect
    if width <= 0 || height <= 0 { return }    // 3. zero-area guard
    b.mu.RLock(); defer b.mu.RUnlock()         // 4. lock state
    t := b.th()                                // 5. read theme AT DRAW TIME
    style := tcell.StyleDefault.Background(t.BgLight()).Foreground(t.Fg())
    fillLine(screen, x, y, width, tcell.StyleDefault.Background(t.Bg()))
    drawCentered(screen, x, y, width, b.text, style) // 6. shared draw helpers
}
```

Rules of thumb:
- Always `DrawForSubclass(screen, self)` first, then `GetInnerRect()`, then the zero-area guard.
- Use shared helpers in `draw.go` — `fillLine`, `fillRect`, `drawText` (clips at `x+maxW`), `drawCentered`, `runeLen` — not hand-rolled emit loops.
- Read colors fresh every Draw; never store a `tcell.Color` in a field.
- For custom keys, implement `InputHandler()` via `WrapInputHandler(...)` (if there's a base handler) or return a closure.
- Optionally implement `GetFieldHeight() int` / `Width() int` for layout sizing.
- If Draw mutates state, apply the prepare/paint split.
- If you wrap a tview-native primitive instead of `tview.Box`, add a `subs Subscriptions` field, a `Subs()` accessor, and `subs.Add(theme.Register(prim))` in the constructor.

### B. A new view (nav.Component)
Hold a `*ComponentBase` (or `*StatefulComponentBase[T]`) field and wire it fluently. Both its `Subs()` and the wrapped widget's `Subs()` release on `Stop()`.

```go
type MyView struct {
    *ComponentBase
    table *Table
}

func NewMyView() *MyView {
    table := NewTable()
    v := &MyView{table: table}
    v.ComponentBase = NewComponentBase(table).
        SetName("my-view").
        AddHint("Enter", "Select").
        SetOnStart(v.loadData)
    return v
}
```
For async data use `StatefulComponentBase[T]`: drive `SetLoadState(LoadStateLoading)` → `SetData(...)`/`SetError(...)`, render off `SetOnStateChange`.

---

## Conventions checklist
1. Composition over embedding for `ComponentBase` (type-safe primitive access).
2. `widgetBase` + `initWidget` is the only correct start for a self-drawing widget; it wires theme auto-update into `Subs()`.
3. Theme = read at draw via `th()`/`Theme()`; subscribe (don't cache) for tview-native primitives.
4. Teardown is automatic *iff* unsubscribers go into a `Subscriptions` exposed via `Subs()`.
5. dado input handlers are bool-return (consume/delegate); tview's are EventKey-return. Use `WrapInputHandler` to run a base handler first.
6. Split Draw into prepare(mutate→snapshot) + paint(pure) only when Draw mutates state.
