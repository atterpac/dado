package inline

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressLifecycle(t *testing.T) {
	t.Parallel()

	progress := NewProgress("Download", 100)
	assert.Equal(t, ProgressPending, progress.Snapshot().State)
	progress.SetDetail("registry.example.com")
	progress.Set(42)
	snapshot := progress.Snapshot()
	assert.Equal(t, ProgressActive, snapshot.State)
	assert.Equal(t, int64(42), snapshot.Current)

	lines := plainFrameLines(progress.Frame(56))
	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], "Download")
	assert.Contains(t, lines[0], "42%")
	assert.Contains(t, lines[1], "registry.example.com")

	progress.Advance(60)
	snapshot = progress.Snapshot()
	assert.Equal(t, ProgressSucceeded, snapshot.State)
	assert.Equal(t, int64(100), snapshot.Current)
	assert.Contains(t, plainFrameLines(progress.Frame(56))[0], "done")
}

func TestProgressStatesAndValues(t *testing.T) {
	t.Parallel()

	progress := NewProgress("Upload", 12, WithProgressValues(true))
	progress.Set(5)
	assert.Contains(t, plainFrameLines(progress.Frame(48))[0], "5/12")
	progress.Fail(errors.New("connection reset"))
	lines := plainFrameLines(progress.Frame(48))
	assert.Contains(t, lines[0], "failed")
	assert.Contains(t, lines[1], "connection reset")

	cancelled := NewProgress("Publish", 0)
	cancelled.Start()
	cancelled.Cancel("interrupted")
	assert.Equal(t, ProgressCancelled, cancelled.Snapshot().State)
	assert.Contains(t, plainFrameLines(cancelled.Frame(40))[0], "cancelled")
}

func TestProgressIndeterminateAnimationUsesClock(t *testing.T) {
	t.Parallel()

	now := time.Unix(0, 0)
	progress := NewProgress("Resolve", 0, WithProgressClock(func() time.Time { return now }))
	progress.Start()
	first := plainFrameLines(progress.Frame(40))[0]
	now = now.Add(80 * time.Millisecond)
	second := plainFrameLines(progress.Frame(40))[0]
	assert.NotEqual(t, first, second)
}

func TestProgressCustomTheme(t *testing.T) {
	t.Parallel()

	theme := DefaultStatusTheme()
	theme.SuccessMarker = "!"
	theme.BarFilled = "#"
	theme.BarEmpty = "-"
	progress := NewProgress("Custom", 4, WithProgressTheme(theme))
	progress.Set(2)
	assert.Contains(t, plainFrameLines(progress.Frame(36))[0], "#")
	progress.Complete()
	assert.True(t, strings.HasPrefix(plainFrameLines(progress.Frame(36))[0], "!"))
}

func TestProgressUsesFractionalBarGlyphs(t *testing.T) {
	t.Parallel()
	progress := NewProgress("Fractional", 3)
	progress.Set(1)
	found := false
	for width := 40; width < 52; width++ {
		line := plainFrameLines(progress.Frame(width))[0]
		found = found || strings.ContainsAny(line, "▏▎▍▌▋▊▉")
	}
	assert.True(t, found)
}

func TestProgressNarrowFramesStayWithinWidth(t *testing.T) {
	t.Parallel()

	progress := NewProgress("A very long operation label", 100)
	progress.SetDetail("A very long operation detail")
	progress.Set(55)
	for width := range 32 {
		for _, line := range plainFrameLines(progress.Frame(width)) {
			assert.LessOrEqual(t, displayWidth(line), width, "width %d: %q", width, line)
		}
	}
}

func TestProgressConcurrentAdvanceAndFrames(t *testing.T) {
	t.Parallel()

	progress := NewProgress("Concurrent", 800)
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 100 {
				progress.Advance(1)
				_ = progress.Frame(48)
			}
		}()
	}
	workers.Wait()
	snapshot := progress.Snapshot()
	assert.Equal(t, int64(800), snapshot.Current, fmt.Sprintf("snapshot: %#v", snapshot))
	assert.Equal(t, ProgressSucceeded, snapshot.State)
}
