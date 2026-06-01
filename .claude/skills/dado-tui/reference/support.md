# Supporting Subsystems

`async`, `effect`, `bus`, `style`, `help`, `validators`, `util`, `clipboard`, `recipes`, `testutil`, and the `dado` CLI.

**The single cross-cutting rule:** do work on goroutines, then call `theme.QueueUpdateDraw(func(){ ...mutate UI... })`. `async` and `effect` do this *for* you (callbacks/handlers run on the UI thread). `bus` handlers and `util.TaskRunner` callbacks do **not** — wrap UI updates yourself there.

---

## async — background loading + indicators

Generic builder (`async/loader.go`):
```go
func NewLoader[T any]() *Loader[T]
func (l *Loader[T]) WithTimeout(d) *Loader[T]   // default 30s; 0 = none
func (l *Loader[T]) WithContext(ctx) *Loader[T]
func (l *Loader[T]) WithIndicator(i LoadingIndicator) *Loader[T]
func (l *Loader[T]) OnSuccess(func(T)) / OnError(func(error)) / OnFinally(func()) *Loader[T]
func (l *Loader[T]) Run(fn LoadFunc[T]) *Loader[T] // LoadFunc = func(ctx) (T, error)
func (l *Loader[T]) Cancel() / IsRunning() bool
```
`Run` runs `fn` on a goroutine, delivers the result via `theme.QueueUpdateDraw`. **All callbacks execute on the UI thread — mutate widgets directly inside them.** A mutex makes a second concurrent `Run` a no-op. Convenience: `Load(fn, onSuccess, onError)`, `LoadWithIndicator(message, fn, onSuccess, onError)`.

`LoadingIndicator` interface (`Show/Hide/Success/Error(err)`) constructors: `Toast(msg)`, `StatusBar(msg, setStatus)`, `ProgressModal(...)`, `Primitive(p, pages, name)`, `Callback(...)`, `Noop()`, `Multi(...)`. `IndicatorConfig{Message, SuccessMessage, ShowSuccess, ShowError}`; `DefaultConfig(msg)` → `ShowSuccess=false, ShowError=true`.

---

## effect — Bubble Tea-style command/Msg layer

Opt-in Elm/Tea `Cmd→Msg` alternative to raw `async` (`effect/effect.go`):
```go
type Msg any
type Effect func(ctx context.Context) Msg // nil Msg = not delivered
func NewDispatcher() *Dispatcher
func (d *Dispatcher) Run(e Effect)
func (d *Dispatcher) Subscribe(matcher func(Msg) bool, h Handler) func()
func (d *Dispatcher) Shutdown(ctx) error
func On[T any](d *Dispatcher, h func(T)) func() // typed subscribe by concrete Msg type
```
`Run` executes on a goroutine; the returned Msg is delivered **on the UI thread via `theme.QueueUpdateDraw`**. Combinators: `Batch(cmds...)` (parallel), `Sequence(cmds...)` (serial), `Tick(dur, fn)` (one-shot timer; re-arm by returning a new Tick), `None()`. Pair with `components.Subscriptions`: `cb.Subs().Add(disp.Subscribe(...))`. The `layout.App` owns a dispatcher (`app.Effects()`). These are effects/commands, not animations.

---

## bus — opt-in pub/sub event bus

Observability/debug instrumentation (`bus/`). Disabled by default; an `atomic.Bool` gates every publish so the off-path is ~free.
```go
type Event struct { Kind, Source string; Payload any; Time time.Time; Seq uint64 }
type Bus interface {
    Publish(Event); Subscribe(filter func(Event) bool, h func(Event)) func()
    Recent(n int) []Event; Close()
}
func Enabled() bool / SetEnabled(on bool)
func Default() Bus / Publish(e Event) / New(ringSize, queueSize int) Bus // defaults 1024/256
```
`Publish` stamps Time+Seq, pushes into a circular `ring`, non-blocking-sends to a buffered channel; **overflow events are dropped** (`Dropped()` counter) so producers never block. A single goroutine fans out; **handlers must not block and must marshal UI touches via `theme.QueueUpdateDraw`.** `Recent(n)` returns newest-N chronologically — feeds the Ctrl+D debug overlay. Kinds: `binding.set/update`, `theme.switch`, `nav.push/pop/replace`, `loader.*`, `input.key`, `effect.msg`. Use for debugging/telemetry, not the primary data path (use `effect` for that).

---

## style — Lip Gloss-style immutable styles + borders

`Style` is an **immutable value** — every builder returns a copy (`style/`):
```go
s := style.New().Bold().
    ForegroundFn(func(t theme.Theme) tcell.Color { return t.Accent() }). // theme-reactive
    Border(style.RoundedBorder)
```
Attributes: `Bold/Italic/Underline/Reverse/Blink/Dim`. Colors: static `Foreground`/`Background`, or theme-reactive `ForegroundFn`/`BackgroundFn`/`BorderColorFn` (resolved at render). `PaddingX/Y` (affect `Render` only). Outputs:
- `Render(text) string` — tview color tags `[fg:bg:attrs]…[-:-:-]`; escapes user content via `tview.Escape`.
- `TcellStyle() tcell.Style` — for custom `Draw()`.
- `Apply(box) *tview.Box` — copies border + bg + border color onto a Box.

**Borders** (`style/border.go`): `BorderSet` (six runes). Presets `NormalBorder`, `RoundedBorder`, `ThickBorder`, `DoubleBorder`, `BlockBorder`. **Caveat:** custom glyphs via `Apply` are a **no-op** — tview draws borders from its global `tview.Borders` rune set. `Apply` only toggles the border on; for custom glyphs hand-draw in your own `Draw()`.

---

## help — help registry + overlay modal

```go
h := help.New().SetAppName("MyApp").SetVersion("1.0").
    AddSection("Navigation", []help.ActionInfo{{Key:"j", Description:"Down"}}).
    AddRegistry("Actions", actionRegistry) // pulls hints from *input.ActionRegistry
```
`ExportMarkdown(path)`, `ExportManPage(path)`, `ContextHints(section) []components.KeyHint`. `h.Modal()` → `*HelpModal`, a self-contained hand-drawn scrollable overlay (`Show/Hide/IsVisible`; Esc/Enter/q close, j/k/g/G/PgUp/PgDn scroll; captures focus while visible).

---

## validators — form validators

`type Validator func(value any) error` (`validators/validators.go`). Built-ins: `Required()`, `MinLength(n)`, `MaxLength(n)`, `Email()`, `URL()`, `Alphanumeric()`, `NoWhitespace()`, `Pattern(re)`/`PatternWithMessage`, `Range(min,max)`, `Min`, `Max`, `OneOf(opts...)`. Most string validators treat `""` as valid — compose with `Required()` for presence. Combinators: `Custom(fn)`, `All(...)` (AND), `Any(...)` (OR), `WithMessage(v, msg)`.

> `components.TextField` takes a `func(string) error` validator; `validators.Validator` is `func(any) error` — adapt when wiring to string fields.

---

## util & clipboard

**`util/taskrunner.go`** — multiple tracked background tasks with progress. `TaskStatus` (Pending/Running/Completed/Failed/Cancelled). `TaskFunc = func(ctx, progress func(pct float64, msg string)) (any, error)`.
```go
r := util.NewTaskRunner().SetMaxParallel(4).
    SetOnStart(fn).SetOnProgress(fn).SetOnComplete(fn).SetOnError(fn)
task := r.Run("name", taskFunc) // async; returns *Task now
r.Cancel(id) / CancelAll() / Wait(id) / WaitAll() / Get / GetAll / GetRunning / Count
```
Queues when over `maxParallel`; indeterminate progress = `-1`. **Caveat:** callbacks fire on the task goroutine — wrap UI mutations in `theme.QueueUpdateDraw` yourself.

**`util/gradient.go`** — `ApplyHorizontal/Vertical/Diagonal/ReverseDiagonalGradient(text, colors []string)` (hex), `InterpolateColor`, `InterpolateColors`, `DefaultGradientColors`, `AccentGradientColors`.

**`clipboard` package (preferred)** — `Copy(text)`, `Paste() (string,error)`, `Available()`, `CopyBytes`, `Clear()`, `WriteAndPaste`. Cross-platform shell-out (pbcopy/wl-copy/xclip/xsel/clip). Prefer over the older `util/clipboard.go` (`CopyToClipboard`/`ReadFromClipboard`/`HasClipboard`).

---

## recipes — pre-built full views

Each implements `nav.Component`, so they drop straight into the nav stack — good copy/extend starting points.

- **`Dashboard`** — multi-pane grid (`SetColumns` 1–6). `AddSection`, `AddWidget`, typed `AddSparkline/AddGauge/AddProgressBar` returning the live widget. Per-section auto-refresh (`Refresh time.Duration`+`OnRefresh`); Tab/Shift+Tab/`hjkl` move focus, `r` refresh all.
- **`LogViewer`** — streaming log pane. `Append(line)`/`AppendLines` (auto-parses ts+level), `maxLines` ring-trim (default 10000), `/` search via `input.CommandBar` with highlight + `n`/`N`, toggles `f`(follow)`p`(pause)`t`(timestamps)`w`(wrap)`c`(clear), `g/G`.
- **`ResourceList[T]`** — generic K9s-style filterable table. `SetColumns`, `SetRowMapper(func(T)[]string)`, `SetFetcher`, `SetRefreshInterval`, `SetOnSelect`. `AddAction(key,desc,func(T))`, `AddBulkAction(key,desc,func([]T))`. `/` filters live, `r` refreshes, Enter selects, Space multi-selects.

---

## testutil — component/snapshot testing

Render to a tcell `SimulationScreen`, assert on cells, drive synthetic input.
- **`screen.go`** — `TestScreen` wraps `tcell.SimulationScreen`. `NewTestScreen(w,h)`, `DrawPrimitive(p)`. Read: `GetContent/GetRow/GetRect`, `ContainsText`, `FindText`, `GetStyleAt`/`GetForegroundAt`/`GetBackgroundAt`. Golden: `Dump()`/`DumpTrimmed()`. Assertions: `AssertContains`, `AssertNotContains`, `AssertTextAt`.
- **`helpers.go`** — `EventCollector[T]` (`Collect/Events/Last/Count/WaitForCount`), `LifecycleRecorder`. Synthetic input: `SimulateKey/SimulateRune/SimulateEnter/Escape/Tab/Up/Down/...`, `TypeString(handler, s)`. Mocks: `MockComponent` (tracks `StartCalled/StopCalled/...`), `MockModal`.
- **`builders.go`** — fluent test builders: `NewTestTextField/Checkbox/Select/RadioGroup/MultiSelect/Form/ComponentBase`, each `WithX(...).Build()`.

Golden snapshot tests live at `components/testdata/draw_golden/*.golden` (driven by `draw_snapshot_test.go`).

---

## dado CLI (`cmd/dado`)

**Inspection/info tool only — it does NOT scaffold projects.** No `dado new`. Subcommands: `theme list|preview`, `component[s] list` (static catalog grouped Core/Forms/Progress/Recipes), `help`, `version`. Bootstrap projects by hand.
