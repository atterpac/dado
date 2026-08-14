---
title: Inline Rendering
description: Render dynamic tcell-styled views without taking over the terminal.
---

The `inline.Renderer` owns a dynamic region in the terminal's normal screen
buffer. It preserves scrollback and leaves the final frame visible when it
closes. Unlike a full-screen Dado app, it does not initialize `tcell.Screen` or
enter the alternate screen.

## Render a frame

```go
package main

import (
    "os"

    "github.com/atterpac/dado/inline"
    "github.com/gdamore/tcell/v2"
)

func main() {
    renderer := inline.NewRenderer(inline.WithOutput(os.Stderr))
    defer renderer.Close()

    frame := inline.NewFrame(40, 2)
    frame.DrawString(0, 0, "Downloading dependencies", tcell.StyleDefault.Bold(true))
    frame.DrawString(0, 1, "████████░░░░ 67%", tcell.StyleDefault.Foreground(tcell.ColorAqua))

    if err := renderer.Render(frame); err != nil {
        panic(err)
    }
}
```

Only changed rows are redrawn. Rendering the same frame again produces no
terminal output. When output is redirected, the renderer suppresses interim
frames and writes the final frame as plain text during `Close`.

## Print above a live frame

Use `Println` or `Printf` for messages that should remain in scrollback:

```go
renderer.Println("downloaded github.com/gdamore/tcell/v2")
renderer.Printf("completed %d tasks", completed)
```

The renderer inserts the message above its managed region and redraws the live
frame below it.

## Render an existing Dado component

Existing components draw to `tcell.Screen`. `CaptureFrame` provides an
offscreen compatibility bridge without opening a real terminal screen:

```go
frame, err := inline.CaptureFrame(width, height, func(screen tcell.Screen) {
    widget.SetRect(0, 0, width, height)
    widget.Draw(screen)
})
if err != nil {
    return err
}
return renderer.Render(frame)
```

The capture bridge is useful while components adopt a smaller shared drawing
surface. Input handling remains the application's responsibility.

## Progress and activity

Use `Progress` for one determinate or indeterminate operation. It is safe to
update from worker goroutines and does not own a render loop:

```go
download := inline.NewProgress("Download assets", 100)
download.SetDetail("release.tar.zst")
download.Set(42)

if err := renderer.Render(download.Frame(width)); err != nil {
    return err
}
```

A non-positive total enables indeterminate mode. `WithProgressValues(true)`
shows `current/total` instead of a percentage. Progress can be completed,
failed, or cancelled and will render the corresponding stable final state.

The existing `Spinner` also supports a renderer-native mode:

```go
spinner := inline.NewSpinner("Resolve dependency graph")
spinner.Begin()
spinner.SetDetail("visiting modules")
renderer.Render(spinner.Frame(width))

spinner.Succeed("186 modules resolved")
renderer.Render(spinner.Frame(width))
```

`Begin` and `Frame` never start a goroutine. Spinner state changes only through
explicit calls such as `Succeed`, `Fail`, and `Cancel`.

`StatusTheme` centralizes markers, semantic styles, bar glyphs, and spinner
frames. `Progress`, `Spinner`, `MultiProgress`, and `Stepper` accept the same
theme through their respective options.

## Forms and input

`Session` adds raw keyboard input to the normal-screen renderer. It restores
the terminal when the form submits, is cancelled, or returns an error:

```go
form := inline.NewForm("Create release").Add(
    inline.NewTextField("name", "Release name").
        Required().
        SetPlaceholder("v1.0.0"),
    inline.NewSelectField("channel", "Channel",
        inline.NewChoice("stable", "Stable"),
        inline.NewChoice("preview", "Preview"),
    ).Required(),
    inline.NewMultiSelectField("targets", "Targets",
        inline.NewChoice("linux", "Linux"),
        inline.NewChoice("darwin", "macOS"),
        inline.NewChoice("windows", "Windows"),
    ).MinSelected(1),
)

result, err := inline.NewSession(renderer).Run(ctx, form)
if err != nil {
    return err
}
name := result["name"].(string)
targets := result["targets"].([]string)
```

Text fields support Unicode insertion, cursor movement, deletion, passwords,
placeholders, required values, and custom validation. Select fields skip
disabled choices. Multi-select fields enforce optional minimum and maximum
counts. `Tab` and `Shift+Tab` move focus; `Escape` and `Ctrl+C` cancel.

Multi-select markers are bracket-free and configurable independently from the
layout theme:

```go
targets.SetIndicator(inline.IndicatorTheme)   // theme glyphs
targets.SetIndicator(inline.IndicatorCheck)   // ✓ / ○
targets.SetIndicator(inline.IndicatorDot)     // ● / ○
targets.SetIndicator(inline.IndicatorDiamond) // ◆ / ◇
targets.SetIndicator(inline.IndicatorBox)     // ▣ / □
targets.SetIndicator(inline.IndicatorMinimal) // ✓ / ·
targets.SetIndicator(inline.IndicatorNone)    // selected style only

targets.SetIndicatorGlyphs("Y", "N", "-")
```

Disabled choices use a dedicated disabled glyph. The markerless preset applies
reverse-video accent styling to selected labels, so selection remains visible
without brackets or leading symbols.

Sessions do not fall back to line-oriented prompts. Non-terminal input returns
`ErrNonInteractive` unless explicitly supplied with `WithSessionInput`, which
also enables deterministic tests and scripted demos.

### Visual presets

Forms default to `RoundedInlineTheme`, which provides rounded focused controls,
semantic color, required and validation markers, styled key hints, and a boxed
submission receipt. Two additional presets make the visual system easy to
compare:

```go
form.SetTheme(inline.RoundedInlineTheme())
form.SetTheme(inline.SquareInlineTheme())
form.SetTheme(inline.ASCIIInlineTheme())
```

`InlineTheme` exposes text, label, muted, accent, border, error, and success
styles alongside border and form glyph sets. Its embedded `Status` theme can be
passed to progress, spinner, multi-progress, and stepper components for a
consistent application-wide treatment.

## Concurrent task progress

`MultiProgress` is a renderer-native component for work that can run and finish
out of order. Worker goroutines may update it safely while one render loop
produces frames:

```go
tasks := inline.NewMultiProgress(
    "Building targets",
    inline.WithMaxVisibleTasks(6),
    inline.WithCompletedTasksCollapsed(true),
)

tasks.Add("linux", "linux/amd64", 100)
tasks.Add("darwin", "darwin/arm64", 100)
tasks.Add("package", "Package artifacts", 0) // indeterminate

tasks.Start("linux")
tasks.Set("linux", 42)
tasks.Complete("linux")
tasks.Fail("darwin", err)
tasks.Skip("package", "a target failed")

if err := renderer.Render(tasks.Frame(width)); err != nil {
    return err
}
```

The component supports pending, running, complete, failed, skipped, and
cancelled states. When height is bounded, running and failed tasks are kept
visible ahead of pending and completed work. It does not start a rendering
goroutine of its own.

## Ordered workflows

`Stepper` presents sequential work separately from concurrent task progress.
It supports activation, advancement, failure and retry, skipped steps, detail
text, and bounded windows for long workflows:

```go
steps := inline.NewStepper(
    "Deploy preview",
    inline.WithMaxVisibleSteps(4),
)

steps.Add("build", "Build artifacts")
steps.Add("verify", "Run verification")
steps.Add("deploy", "Deploy preview")

steps.Start()                         // activates build
steps.SetDetail("build", "18/24 packages")
steps.Advance()                       // completes build, activates verify
steps.Fail("verify", err)
steps.Activate("verify")              // retries and clears the failure
steps.Advance()
steps.Skip("deploy", "dry run only")

if err := renderer.Render(steps.Frame(width)); err != nil {
    return err
}
```

Vertical layout is the default. Pass
`inline.WithStepOrientation(inline.StepHorizontal)` for a compact single-row
stepper. State mutation and frame snapshots are safe across goroutines, while
the application retains ownership of the render loop.

## Cursor and cleanup

Call `frame.ShowCursor(x, y)` for editable views. The renderer restores its
baseline before every update and returns the cursor to a normal visible state
on `Close`. Call `Clear` first if the final live region should be removed.

For remote PTYs or tests where the output writer is not an `*os.File`, use
`inline.WithTerminalOutput(true)`. Synchronized output can be enabled for a
known-compatible terminal with `inline.WithSynchronizedOutput(true)`.

## Run the demos

All inline showcases live in one command. Run the renderer, native progress and
spinner, concurrent task group, and stepper sequentially:

```sh
go run ./cmd/inlinerenderdemo -demo all
```

Select an individual showcase with `-demo`:

```sh
go run ./cmd/inlinerenderdemo -demo renderer
go run ./cmd/inlinerenderdemo -demo progress
go run ./cmd/inlinerenderdemo -demo spinner
go run ./cmd/inlinerenderdemo -demo multi
go run ./cmd/inlinerenderdemo -demo stepper
go run ./cmd/inlinerenderdemo -demo text
go run ./cmd/inlinerenderdemo -demo select
go run ./cmd/inlinerenderdemo -demo multiselect
go run ./cmd/inlinerenderdemo -demo form
```

Use `-width` to cap the rendered width and `-delay` to adjust animation speed.
Redirected output automatically skips animation and prints one final frame per
selected showcase.

Use `-style rounded`, `-style square`, or `-style ascii` to compare the visual
presets:

```sh
go run ./cmd/inlinerenderdemo -demo form -style rounded
go run ./cmd/inlinerenderdemo -demo form -style square
go run ./cmd/inlinerenderdemo -demo form -style ascii
```

Use `-indicator` to compare multiselect marker presets:

```sh
go run ./cmd/inlinerenderdemo -demo multiselect -indicator check
go run ./cmd/inlinerenderdemo -demo multiselect -indicator dot
go run ./cmd/inlinerenderdemo -demo multiselect -indicator diamond
go run ./cmd/inlinerenderdemo -demo multiselect -indicator box
go run ./cmd/inlinerenderdemo -demo multiselect -indicator minimal
go run ./cmd/inlinerenderdemo -demo multiselect -indicator none
```
