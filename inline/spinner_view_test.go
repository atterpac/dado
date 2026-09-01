package inline

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpinnerFrameLifecycle(t *testing.T) {
	t.Parallel()

	spinner := NewSpinner("Index packages")
	assert.Equal(t, SpinnerPending, spinner.Snapshot().State)
	spinner.Begin()
	spinner.SetDetail("reading module graph")
	assert.Equal(t, SpinnerActive, spinner.Snapshot().State)
	lines := plainFrameLines(spinner.Frame(44))
	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], "Index packages")
	assert.Contains(t, lines[0], "running")
	assert.Contains(t, lines[1], "reading module graph")

	spinner.Succeed("24 packages indexed")
	assert.Equal(t, SpinnerSucceeded, spinner.Snapshot().State)
	assert.Contains(t, plainFrameLines(spinner.Frame(44))[0], "done")
}

func TestSpinnerFailureAndCancellation(t *testing.T) {
	t.Parallel()

	spinner := NewSpinner("Connect")
	spinner.Begin()
	spinner.Fail(errors.New("connection refused"))
	lines := plainFrameLines(spinner.Frame(40))
	assert.Contains(t, lines[0], "failed")
	assert.Contains(t, lines[1], "connection refused")

	spinner.Cancel("interrupted")
	snapshot := spinner.Snapshot()
	assert.Equal(t, SpinnerCancelled, snapshot.State)
	assert.Equal(t, "interrupted", snapshot.Detail)
}

func TestSpinnerFrameAnimationUsesClock(t *testing.T) {
	t.Parallel()

	now := time.Unix(0, 0)
	spinner := NewSpinner("Work", WithSpinnerClock(func() time.Time { return now }))
	spinner.Begin()
	first := plainFrameLines(spinner.Frame(32))[0]
	now = now.Add(80 * time.Millisecond)
	second := plainFrameLines(spinner.Frame(32))[0]
	assert.NotEqual(t, first, second)
}

func TestSpinnerCustomFrames(t *testing.T) {
	t.Parallel()

	theme := DefaultStatusTheme()
	theme.SpinnerFrames = []string{"a", "b"}
	now := time.Unix(0, 0)
	spinner := NewSpinner("Custom", WithSpinnerTheme(theme), WithSpinnerClock(func() time.Time { return now }))
	spinner.Begin()
	assert.True(t, strings.HasPrefix(plainFrameLines(spinner.Frame(32))[0], "a"))
	now = now.Add(80 * time.Millisecond)
	assert.True(t, strings.HasPrefix(plainFrameLines(spinner.Frame(32))[0], "b"))
}

func TestSpinnerNarrowFramesStayWithinWidth(t *testing.T) {
	t.Parallel()

	spinner := NewSpinner("A very long spinner label")
	spinner.Begin()
	spinner.SetDetail("A very long spinner detail")
	for width := range 30 {
		for _, line := range plainFrameLines(spinner.Frame(width)) {
			assert.LessOrEqual(t, displayWidth(line), width, "width %d: %q", width, line)
		}
	}
}

func TestSpinnerConcurrentUpdatesAndFrames(t *testing.T) {
	t.Parallel()

	spinner := NewSpinner("Concurrent")
	spinner.Begin()
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 100 {
				spinner.SetDetail("working")
				_ = spinner.Frame(40)
			}
		}()
	}
	workers.Wait()
	assert.Equal(t, SpinnerActive, spinner.Snapshot().State)
}
