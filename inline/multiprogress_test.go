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

func TestMultiProgressTaskLifecycle(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 14, 8, 0, 0, 0, time.UTC)
	progress := NewMultiProgress("Release", WithMultiProgressClock(func() time.Time { return now }))
	require.NoError(t, progress.Add("compile", "Compile", 100))
	require.NoError(t, progress.Add("upload", "Upload", 0))
	require.NoError(t, progress.Add("package", "Package", 50))

	require.NoError(t, progress.Set("compile", 25))
	now = now.Add(time.Second)
	require.NoError(t, progress.Start("upload"))
	require.NoError(t, progress.SetDetail("upload", "registry.example.com"))
	now = now.Add(time.Second)
	require.NoError(t, progress.Fail("upload", errors.New("connection reset")))
	require.NoError(t, progress.Complete("package"))

	tasks := progress.Tasks()
	require.Len(t, tasks, 3)
	assert.Equal(t, TaskRunning, tasks[0].State)
	assert.Equal(t, int64(25), tasks[0].Current)
	assert.Equal(t, TaskFailed, tasks[1].State)
	assert.Equal(t, "connection reset", tasks[1].Error)
	assert.Equal(t, TaskComplete, tasks[2].State)
	assert.Equal(t, int64(50), tasks[2].Current)
	assert.False(t, tasks[1].FinishedAt.IsZero())

	lines := plainFrameLines(progress.Frame(64))
	assert.Equal(t, "Release", lines[0])
	assert.Contains(t, lines[1], "Overall 1/3")
	assert.Contains(t, lines[2], "Compile")
	assert.Contains(t, lines[2], "25%")
	assert.Contains(t, lines[3], "Upload · connection reset")
	assert.Contains(t, lines[3], "failed")
	assert.Contains(t, lines[4], "Package")
	assert.Contains(t, lines[4], "done")
}

func TestMultiProgressCompletesAtTotalAndIgnoresLateUpdates(t *testing.T) {
	t.Parallel()

	progress := NewMultiProgress("")
	require.NoError(t, progress.Add("download", "Download", 10))
	require.NoError(t, progress.Advance("download", 12))
	require.NoError(t, progress.Set("download", 3))

	task := progress.Tasks()[0]
	assert.Equal(t, TaskComplete, task.State)
	assert.Equal(t, int64(10), task.Current)
}

func TestMultiProgressCollapseAndOverflow(t *testing.T) {
	t.Parallel()

	progress := NewMultiProgress("Tasks",
		WithCompletedTasksCollapsed(true),
		WithMaxVisibleTasks(3),
	)
	for _, id := range []string{"one", "two", "three", "four", "five"} {
		require.NoError(t, progress.Add(id, id, 10))
	}
	require.NoError(t, progress.Complete("one"))
	require.NoError(t, progress.Complete("two"))
	require.NoError(t, progress.Start("four"))
	require.NoError(t, progress.Fail("five", errors.New("boom")))

	lines := plainFrameLines(progress.Frame(48))
	// Title + aggregate + two prioritized task rows + overflow row.
	require.Len(t, lines, 5)
	assert.Contains(t, lines[2], "five · boom")
	assert.Contains(t, lines[3], "four")
	assert.Contains(t, lines[4], "and 2 more tasks")
}

func TestMultiProgressCollapsedCompletedSummary(t *testing.T) {
	t.Parallel()

	progress := NewMultiProgress("", WithCompletedTasksCollapsed(true), WithAggregateProgress(false))
	require.NoError(t, progress.Add("one", "one", 1))
	require.NoError(t, progress.Add("two", "two", 1))
	require.NoError(t, progress.Complete("one"))
	require.NoError(t, progress.Complete("two"))

	lines := plainFrameLines(progress.Frame(30))
	require.Len(t, lines, 1)
	assert.Contains(t, lines[0], "2 completed")
}

func TestMultiProgressNarrowFramesStayWithinWidth(t *testing.T) {
	t.Parallel()

	progress := NewMultiProgress("Long task group title")
	require.NoError(t, progress.Add("one", "A very long task label", 100))
	require.NoError(t, progress.Set("one", 42))

	for width := range 30 {
		frame := progress.Frame(width)
		for _, line := range plainFrameLines(frame) {
			assert.LessOrEqual(t, displayWidth(line), width, "width %d: %q", width, line)
		}
	}
}

func TestMultiProgressAnimationUsesInjectedClock(t *testing.T) {
	t.Parallel()

	now := time.Unix(0, 0)
	progress := NewMultiProgress("", WithAggregateProgress(false), WithMultiProgressClock(func() time.Time { return now }))
	require.NoError(t, progress.Add("work", "Work", 0))
	require.NoError(t, progress.Start("work"))
	first := plainFrameLines(progress.Frame(30))[0]

	now = now.Add(80 * time.Millisecond)
	second := plainFrameLines(progress.Frame(30))[0]
	assert.NotEqual(t, first, second)
}

func TestMultiProgressCustomTheme(t *testing.T) {
	t.Parallel()

	theme := DefaultStatusTheme()
	theme.FailureMarker = "!"
	progress := NewMultiProgress("", WithAggregateProgress(false), WithMultiProgressTheme(theme))
	require.NoError(t, progress.Add("work", "Work", 1))
	require.NoError(t, progress.Fail("work", errors.New("boom")))
	assert.True(t, strings.HasPrefix(plainFrameLines(progress.Frame(30))[0], "!"))
}

func TestMultiProgressValidationErrors(t *testing.T) {
	t.Parallel()

	progress := NewMultiProgress("")
	assert.ErrorIs(t, progress.Add("", "bad", 1), ErrInvalidTaskID)
	require.NoError(t, progress.Add("task", "", -1))
	assert.ErrorIs(t, progress.Add("task", "duplicate", 1), ErrTaskExists)
	assert.ErrorIs(t, progress.Start("missing"), ErrTaskNotFound)
	assert.Equal(t, "task", progress.Tasks()[0].Label)
	assert.Zero(t, progress.Tasks()[0].Total)
}

func TestMultiProgressConcurrentUpdates(t *testing.T) {
	t.Parallel()

	progress := NewMultiProgress("Concurrent")
	const taskCount = 20
	for i := range taskCount {
		require.NoError(t, progress.Add(string(rune('a'+i)), "worker", 100))
	}

	var wg sync.WaitGroup
	for i := range taskCount {
		id := string(rune('a' + i))
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				assert.NoError(t, progress.Advance(id, 1))
				_ = progress.Frame(60)
			}
		}()
	}
	wg.Wait()
	for _, task := range progress.Tasks() {
		assert.Equal(t, TaskComplete, task.State)
		assert.Equal(t, int64(100), task.Current)
	}
}

func plainFrameLines(frame *Frame) []string {
	plain := strings.TrimSuffix(plainSnapshot(snapshotFrame(frame)), "\n")
	if plain == "" {
		return nil
	}
	lines := strings.Split(plain, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return lines
}
