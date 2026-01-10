---
label: Progress
icon: meter
order: 40
---

# Progress

Progress indicator with percentage display.

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

progress := components.NewProgress().
    SetLabel("Downloading...").
    SetProgress(0.5)  // 50%
```

---

## Configuration

```go
progress := components.NewProgress().
    SetLabel("Processing files...").
    SetProgress(0.0).
    SetShowPercentage(true)

// Update progress
progress.SetProgress(0.25)  // 25%
progress.SetProgress(0.50)  // 50%
progress.SetProgress(0.75)  // 75%
progress.Complete()         // 100%
```

---

## Methods

| Method | Description |
|--------|-------------|
| `SetLabel(string)` | Set progress label |
| `SetProgress(float64)` | Set progress (0.0-1.0) |
| `GetProgress()` | Get current progress |
| `SetShowPercentage(bool)` | Show/hide percentage |
| `Complete()` | Set to 100% |
| `Reset()` | Reset to 0% |

---

## Example: File Upload

```go
func uploadFile(app *layout.App, file *os.File) {
    progress := components.NewProgress().
        SetLabel("Uploading " + file.Name()).
        SetShowPercentage(true)

    panel := components.NewPanel().
        SetTitle("Upload Progress").
        SetContent(progress)

    modal := components.NewModal(components.ModalConfig{
        Title:  "Uploading",
        Width:  50,
        Height: 5,
    }).SetContent(progress).
        SetBlockUntilDismissed(true)

    app.Pages().Push(modal)

    go func() {
        stat, _ := file.Stat()
        totalSize := stat.Size()
        var uploaded int64

        reader := &progressReader{
            reader: file,
            onProgress: func(n int64) {
                uploaded += n
                pct := float64(uploaded) / float64(totalSize)
                app.QueueUpdateDraw(func() {
                    progress.SetProgress(pct)
                })
            },
        }

        // Upload with progress tracking
        err := api.Upload(reader)

        app.QueueUpdateDraw(func() {
            app.Pages().Pop()
            if err != nil {
                components.ShowToast(app, "Upload failed: "+err.Error(), 3*time.Second)
            } else {
                components.ShowToast(app, "Upload complete!", 2*time.Second)
            }
        })
    }()
}
```

---

## Progress Modal

For modal progress dialogs:

```go
func ShowProgressModal(app *layout.App, title string) *components.ProgressModal {
    modal := components.NewProgressModal(components.ModalConfig{
        Title:  title,
        Width:  50,
        Height: 5,
    }).SetBlockUntilDismissed(true)

    app.Pages().Push(modal)
    return modal
}

// Usage
progress := ShowProgressModal(app, "Processing...")

go func() {
    for i := 0; i <= 100; i++ {
        app.QueueUpdateDraw(func() {
            progress.SetProgress(float64(i) / 100)
            progress.SetLabel(fmt.Sprintf("Step %d of 100", i))
        })
        time.Sleep(50 * time.Millisecond)
    }

    app.QueueUpdateDraw(func() {
        app.Pages().Pop()
    })
}()
```

---

## Indeterminate Progress

For operations with unknown duration:

```go
// Use async.Indicator for indeterminate progress
async.NewLoader[Result]().
    WithIndicator(async.Toast("Loading...")).
    OnSuccess(func(r Result) { /* ... */ }).
    Run(func(ctx context.Context) (Result, error) {
        return longOperation(ctx)
    })
```
