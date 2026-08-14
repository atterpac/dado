package inline

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

var (
	// ErrTaskNotFound is returned when a task ID is not present in a MultiProgress.
	ErrTaskNotFound = errors.New("inline: task not found")
	// ErrTaskExists is returned when a task ID has already been added.
	ErrTaskExists = errors.New("inline: task already exists")
	// ErrInvalidTaskID is returned when a task ID is empty.
	ErrInvalidTaskID = errors.New("inline: task ID cannot be empty")
)

// TaskState describes the lifecycle state of a MultiProgress task.
type TaskState int

const (
	TaskPending TaskState = iota
	TaskRunning
	TaskComplete
	TaskFailed
	TaskSkipped
	TaskCancelled
)

// String returns the lowercase task-state name.
func (state TaskState) String() string {
	switch state {
	case TaskRunning:
		return "running"
	case TaskComplete:
		return "complete"
	case TaskFailed:
		return "failed"
	case TaskSkipped:
		return "skipped"
	case TaskCancelled:
		return "cancelled"
	default:
		return "pending"
	}
}

// TaskSnapshot is an immutable view of a task's current state.
type TaskSnapshot struct {
	ID         string
	Label      string
	Detail     string
	State      TaskState
	Current    int64
	Total      int64
	Error      string
	StartedAt  time.Time
	FinishedAt time.Time
}

// MultiProgressOption configures a MultiProgress component.
type MultiProgressOption func(*MultiProgress)

// WithMaxVisibleTasks bounds the number of task rows. Active and failed tasks
// are prioritized when not every task fits. A value less than one shows all.
func WithMaxVisibleTasks(maximum int) MultiProgressOption {
	return func(progress *MultiProgress) { progress.maxVisible = max(maximum, 0) }
}

// WithCompletedTasksCollapsed groups completed tasks into a single summary row.
func WithCompletedTasksCollapsed(collapsed bool) MultiProgressOption {
	return func(progress *MultiProgress) { progress.collapseCompleted = collapsed }
}

// WithAggregateProgress controls whether an overall progress row is shown.
func WithAggregateProgress(visible bool) MultiProgressOption {
	return func(progress *MultiProgress) { progress.showAggregate = visible }
}

// WithMultiProgressTheme supplies shared semantic styles and glyphs.
func WithMultiProgressTheme(theme StatusTheme) MultiProgressOption {
	return func(progress *MultiProgress) { progress.theme = normalizedStatusTheme(theme) }
}

// WithMultiProgressClock supplies the clock used for animation and task
// timestamps. It is primarily useful for deterministic tests.
func WithMultiProgressClock(clock func() time.Time) MultiProgressOption {
	return func(progress *MultiProgress) {
		if clock != nil {
			progress.clock = clock
		}
	}
}

// MultiProgress renders a group of concurrent tasks. It owns task state but
// never writes to the terminal; callers render its Frame through Renderer.
// State mutation and snapshotting are safe from multiple goroutines.
type MultiProgress struct {
	mu sync.RWMutex

	title             string
	tasks             []TaskSnapshot
	byID              map[string]int
	maxVisible        int
	collapseCompleted bool
	showAggregate     bool
	theme             StatusTheme
	clock             func() time.Time
}

// NewMultiProgress creates an empty concurrent task group.
func NewMultiProgress(title string, options ...MultiProgressOption) *MultiProgress {
	progress := &MultiProgress{
		title:         title,
		byID:          make(map[string]int),
		showAggregate: true,
		theme:         DefaultStatusTheme(),
		clock:         time.Now,
	}
	for _, option := range options {
		if option != nil {
			option(progress)
		}
	}
	return progress
}

// Add adds a pending task. A non-positive total creates an indeterminate task.
func (m *MultiProgress) Add(id, label string, total int64) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidTaskID
	}
	if total < 0 {
		total = 0
	}
	if label == "" {
		label = id
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.byID[id]; exists {
		return fmt.Errorf("%w: %s", ErrTaskExists, id)
	}
	m.byID[id] = len(m.tasks)
	m.tasks = append(m.tasks, TaskSnapshot{ID: id, Label: label, Total: total})
	return nil
}

// Start marks a task as running.
func (m *MultiProgress) Start(id string) error {
	return m.mutate(id, func(task *TaskSnapshot, now time.Time) {
		startTask(task, now)
	})
}

// Set updates a task's current value, clamped to its total. Pending tasks are
// started automatically; reaching a positive total completes the task.
func (m *MultiProgress) Set(id string, current int64) error {
	return m.mutate(id, func(task *TaskSnapshot, now time.Time) {
		if taskFinished(task.State) {
			return
		}
		startTask(task, now)
		current, complete := clampProgress(current, task.Total)
		task.Current = current
		if complete {
			finishTask(task, TaskComplete, now, "", "")
			return
		}
	})
}

// Advance increments a task's current value.
func (m *MultiProgress) Advance(id string, delta int64) error {
	return m.mutate(id, func(task *TaskSnapshot, now time.Time) {
		if taskFinished(task.State) {
			return
		}
		startTask(task, now)
		current, complete := clampProgress(task.Current+delta, task.Total)
		task.Current = current
		if complete {
			finishTask(task, TaskComplete, now, "", "")
			return
		}
	})
}

// SetTotal changes a task's total. Non-positive values make it indeterminate.
func (m *MultiProgress) SetTotal(id string, total int64) error {
	return m.mutate(id, func(task *TaskSnapshot, now time.Time) {
		if taskFinished(task.State) {
			return
		}
		task.Total = max(total, 0)
		if current, complete := clampProgress(task.Current, task.Total); complete {
			task.Current = current
			finishTask(task, TaskComplete, now, "", "")
		}
	})
}

// SetDetail sets secondary text displayed alongside a task when width allows.
func (m *MultiProgress) SetDetail(id, detail string) error {
	return m.mutate(id, func(task *TaskSnapshot, _ time.Time) { task.Detail = detail })
}

// Complete marks a task successful and fills determinate progress.
func (m *MultiProgress) Complete(id string) error {
	return m.mutate(id, func(task *TaskSnapshot, now time.Time) {
		if task.Total > 0 {
			task.Current = task.Total
		}
		finishTask(task, TaskComplete, now, "", task.Detail)
	})
}

// Fail marks a task failed and records err for snapshots and detail rendering.
func (m *MultiProgress) Fail(id string, err error) error {
	message := ""
	if err != nil {
		message = err.Error()
	}
	return m.mutate(id, func(task *TaskSnapshot, now time.Time) {
		finishTask(task, TaskFailed, now, message, task.Detail)
	})
}

// Skip marks a task skipped with an optional reason.
func (m *MultiProgress) Skip(id, reason string) error {
	return m.mutate(id, func(task *TaskSnapshot, now time.Time) {
		finishTask(task, TaskSkipped, now, "", reason)
	})
}

// Cancel marks a task cancelled with an optional reason.
func (m *MultiProgress) Cancel(id, reason string) error {
	return m.mutate(id, func(task *TaskSnapshot, now time.Time) {
		finishTask(task, TaskCancelled, now, "", reason)
	})
}

// Tasks returns an insertion-ordered copy of the task state.
func (m *MultiProgress) Tasks() []TaskSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]TaskSnapshot(nil), m.tasks...)
}

// Frame builds the component's current renderer-native view.
func (m *MultiProgress) Frame(width int) *Frame {
	if width < 0 {
		width = 0
	}
	m.mu.RLock()
	title := m.title
	tasks := append([]TaskSnapshot(nil), m.tasks...)
	maximum := m.maxVisible
	collapse := m.collapseCompleted
	showAggregate := m.showAggregate
	theme := m.theme
	clock := m.clock
	m.mu.RUnlock()
	now := clock()

	rows := taskRows(tasks, collapse, maximum)
	height := len(rows)
	if title != "" {
		height++
	}
	if showAggregate && len(tasks) > 0 {
		height++
	}
	frame := NewFrame(width, height)
	y := 0
	if title != "" {
		drawClipped(frame, 0, y, title, width, tcell.StyleDefault.Bold(true))
		y++
	}
	if showAggregate && len(tasks) > 0 {
		drawAggregateRow(frame, y, width, tasks, theme)
		y++
	}
	for _, row := range rows {
		drawTaskRow(frame, y, width, row, now, theme)
		y++
	}
	return frame
}

func (m *MultiProgress) mutate(id string, mutate func(*TaskSnapshot, time.Time)) error {
	now := m.clock()
	m.mu.Lock()
	defer m.mu.Unlock()
	index, exists := m.byID[id]
	if !exists {
		return fmt.Errorf("%w: %s", ErrTaskNotFound, id)
	}
	mutate(&m.tasks[index], now)
	return nil
}

func startTask(task *TaskSnapshot, now time.Time) {
	if task.StartedAt.IsZero() {
		task.StartedAt = now
	}
	if task.State == TaskPending {
		task.State = TaskRunning
	}
}

func finishTask(task *TaskSnapshot, state TaskState, now time.Time, failure, detail string) {
	if task.StartedAt.IsZero() {
		task.StartedAt = now
	}
	task.State = state
	task.FinishedAt = now
	task.Error = failure
	if detail != "" {
		task.Detail = detail
	}
}

func taskFinished(state TaskState) bool {
	return semanticFinished(taskSemanticStatus(state))
}

type taskRow struct {
	task      *TaskSnapshot
	completed int
	overflow  int
}

func taskRows(tasks []TaskSnapshot, collapse bool, maximum int) []taskRow {
	rows := make([]taskRow, 0, len(tasks))
	completed := 0
	for i := range tasks {
		if collapse && tasks[i].State == TaskComplete {
			completed++
			continue
		}
		rows = append(rows, taskRow{task: &tasks[i]})
	}
	if completed > 0 {
		rows = append(rows, taskRow{completed: completed})
	}
	if maximum < 1 || len(rows) <= maximum {
		return rows
	}

	prioritized := make([]taskRow, 0, len(rows))
	for _, state := range []TaskState{TaskFailed, TaskRunning, TaskPending, TaskCancelled, TaskSkipped, TaskComplete} {
		for _, row := range rows {
			if row.task != nil && row.task.State == state {
				prioritized = append(prioritized, row)
			}
		}
	}
	for _, row := range rows {
		if row.task == nil {
			prioritized = append(prioritized, row)
		}
	}
	visible := max(maximum-1, 0)
	result := append([]taskRow(nil), prioritized[:visible]...)
	result = append(result, taskRow{overflow: len(rows) - visible})
	return result
}

func drawAggregateRow(frame *Frame, y, width int, tasks []TaskSnapshot, theme StatusTheme) {
	completed := 0
	var current, total int64
	for _, task := range tasks {
		if task.State == TaskComplete || task.State == TaskSkipped {
			completed++
		}
		// Skipped work is complete for the task count but contributes neither
		// unfinished work nor a zero-valued total to the aggregate percentage.
		if task.Total > 0 && task.State != TaskSkipped {
			total += task.Total
			current += min(task.Current, task.Total)
		}
	}
	label := fmt.Sprintf("Overall %d/%d", completed, len(tasks))
	fraction := 0.0
	if total > 0 {
		fraction = float64(current) / float64(total)
	}
	status := ""
	if total > 0 {
		status = formatProgressPercent(fraction)
	}
	drawProgressRowWithTheme(frame, y, width, "Σ", label, fraction, total > 0, theme.ActiveStyle, status, theme)
}

func drawTaskRow(frame *Frame, y, width int, row taskRow, now time.Time, theme StatusTheme) {
	if row.completed > 0 {
		drawClipped(frame, 0, y, theme.SuccessMarker, 1, theme.SuccessStyle)
		drawClipped(frame, 2, y, fmt.Sprintf("%d completed", row.completed), max(width-2, 0), tcell.StyleDefault.Dim(true))
		return
	}
	if row.overflow > 0 {
		drawClipped(frame, 0, y, "…", 1, tcell.StyleDefault.Dim(true))
		drawClipped(frame, 2, y, fmt.Sprintf("and %d more tasks", row.overflow), max(width-2, 0), tcell.StyleDefault.Dim(true))
		return
	}

	task := *row.task
	marker, style, status := taskAppearance(task, now, theme)
	label := task.Label
	detail := task.Detail
	if task.State == TaskFailed && task.Error != "" {
		detail = task.Error
	}
	if detail != "" {
		label += " · " + detail
	}
	determinate := task.Total > 0
	fraction := 0.0
	if determinate {
		fraction = float64(min(task.Current, task.Total)) / float64(task.Total)
		if task.State == TaskPending || task.State == TaskRunning {
			status = formatProgressPercent(fraction)
		}
	}
	drawProgressRowWithTheme(frame, y, width, marker, label, fraction, determinate, style, status, theme)
}

func taskAppearance(task TaskSnapshot, now time.Time, theme StatusTheme) (marker string, style tcell.Style, status string) {
	return semanticAppearance(taskSemanticStatus(task.State), true, theme, now.UnixMilli())
}

func taskSemanticStatus(state TaskState) semanticStatus {
	switch state {
	case TaskRunning:
		return statusActive
	case TaskComplete:
		return statusSucceeded
	case TaskFailed:
		return statusFailed
	case TaskSkipped:
		return statusSkipped
	case TaskCancelled:
		return statusCancelled
	default:
		return statusPending
	}
}

func drawProgressRowWithTheme(frame *Frame, y, width int, marker, label string, fraction float64, determinate bool, style tcell.Style, status string, theme StatusTheme) {
	if width <= 0 {
		return
	}
	drawClipped(frame, 0, y, marker, 1, style)
	if width <= 3 {
		return
	}
	statusText := status
	statusWidth := 0
	if statusText != "" {
		statusWidth = min(max(displayWidth(statusText), 4), max(width/3, 4))
	}
	if width < 24 {
		labelWidth := max(width-statusWidth-3, 0)
		drawClipped(frame, 2, y, label, labelWidth, tcell.StyleDefault)
		drawRight(frame, y, width, statusText, style)
		return
	}

	labelWidth := min(max(width*2/5, 10), 32)
	barX := 2 + labelWidth + 1
	barWidth := width - barX - statusWidth - 1
	if barWidth < 4 {
		labelWidth = max(labelWidth-(4-barWidth), 4)
		barX = 2 + labelWidth + 1
		barWidth = max(width-barX-statusWidth-1, 0)
	}
	drawClipped(frame, 2, y, label, labelWidth, tcell.StyleDefault)
	if barWidth > 0 {
		filled := 0
		partial := ""
		if determinate {
			exact := fraction * float64(barWidth)
			filled = min(int(exact), barWidth)
			if filled < barWidth && len(theme.BarPartial) > 0 {
				remainder := exact - float64(filled)
				if remainder > 0 {
					index := min(int(remainder*float64(len(theme.BarPartial))), len(theme.BarPartial)-1)
					partial = theme.BarPartial[index]
				}
			}
		}
		for x := range barWidth {
			character := theme.BarEmpty
			barStyle := theme.PendingStyle
			if determinate && x < filled {
				character = theme.BarFilled
				barStyle = style
			} else if determinate && x == filled && partial != "" {
				character = partial
				barStyle = style
			}
			frame.DrawString(barX+x, y, character, barStyle)
		}
	}
	drawRight(frame, y, width, statusText, style)
}

func drawRight(frame *Frame, y, width int, text string, style tcell.Style) {
	textWidth := displayWidth(text)
	drawClipped(frame, max(width-textWidth, 0), y, text, min(textWidth, width), style)
}

func drawClipped(frame *Frame, x, y int, text string, maximum int, style tcell.Style) {
	if maximum <= 0 || x < 0 || y < 0 || y >= frame.height {
		return
	}
	used := 0
	for text != "" && used < maximum {
		rest, width := frame.Put(x+used, y, text, style)
		if width <= 0 || rest == text {
			break
		}
		if used+width > maximum {
			frame.Put(x+used, y, " ", tcell.StyleDefault)
			break
		}
		used += width
		text = rest
	}
}
