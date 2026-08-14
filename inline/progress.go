package inline

import (
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

// ProgressState describes a renderer-native Progress lifecycle.
type ProgressState int

const (
	ProgressPending ProgressState = iota
	ProgressActive
	ProgressSucceeded
	ProgressFailed
	ProgressCancelled
)

// String returns the lowercase progress-state name.
func (state ProgressState) String() string {
	switch state {
	case ProgressActive:
		return "active"
	case ProgressSucceeded:
		return "succeeded"
	case ProgressFailed:
		return "failed"
	case ProgressCancelled:
		return "cancelled"
	default:
		return "pending"
	}
}

// ProgressSnapshot is an immutable view of a Progress value.
type ProgressSnapshot struct {
	Label   string
	Detail  string
	State   ProgressState
	Current int64
	Total   int64
	Error   string
}

// ProgressOption configures a Progress component.
type ProgressOption func(*Progress)

// WithProgressPercentage controls percentage display for determinate work.
func WithProgressPercentage(visible bool) ProgressOption {
	return func(progress *Progress) { progress.showPercentage = visible }
}

// WithProgressValues displays current/total instead of a percentage.
func WithProgressValues(visible bool) ProgressOption {
	return func(progress *Progress) { progress.showValues = visible }
}

// WithProgressTheme supplies semantic styles and glyphs.
func WithProgressTheme(theme StatusTheme) ProgressOption {
	return func(progress *Progress) { progress.theme = normalizedStatusTheme(theme) }
}

// WithProgressClock supplies the animation clock, primarily for tests.
func WithProgressClock(clock func() time.Time) ProgressOption {
	return func(progress *Progress) {
		if clock != nil {
			progress.clock = clock
		}
	}
}

// Progress renders one determinate or indeterminate operation. It owns state
// but never starts a goroutine or writes to the terminal.
type Progress struct {
	mu sync.RWMutex

	label          string
	detail         string
	state          ProgressState
	current        int64
	total          int64
	failure        string
	showPercentage bool
	showValues     bool
	theme          StatusTheme
	clock          func() time.Time
}

// NewProgress creates a pending operation. A non-positive total is
// indeterminate.
func NewProgress(label string, total int64, options ...ProgressOption) *Progress {
	progress := &Progress{
		label:          label,
		total:          max(total, 0),
		showPercentage: true,
		theme:          DefaultStatusTheme(),
		clock:          time.Now,
	}
	for _, option := range options {
		if option != nil {
			option(progress)
		}
	}
	progress.theme = normalizedStatusTheme(progress.theme)
	return progress
}

// Start marks the operation active.
func (p *Progress) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !progressFinished(p.state) {
		p.state = ProgressActive
	}
}

// Set updates current, clamps it to total, and starts pending work. Reaching a
// positive total succeeds the operation.
func (p *Progress) Set(current int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if progressFinished(p.state) {
		return
	}
	p.state = ProgressActive
	current, complete := clampProgress(current, p.total)
	p.current = current
	if complete {
		p.state = ProgressSucceeded
	}
}

// Advance increments current.
func (p *Progress) Advance(delta int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if progressFinished(p.state) {
		return
	}
	p.state = ProgressActive
	current, complete := clampProgress(p.current+delta, p.total)
	p.current = current
	if complete {
		p.state = ProgressSucceeded
	}
}

// SetTotal changes total. A non-positive total makes the operation
// indeterminate.
func (p *Progress) SetTotal(total int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if progressFinished(p.state) {
		return
	}
	p.total = max(total, 0)
	if current, complete := clampProgress(p.current, p.total); complete {
		p.current = current
		p.state = ProgressSucceeded
	}
}

// SetDetail changes the secondary line.
func (p *Progress) SetDetail(detail string) {
	p.mu.Lock()
	p.detail = detail
	p.mu.Unlock()
}

// Complete marks the operation successful and fills determinate progress.
func (p *Progress) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.total > 0 {
		p.current = p.total
	}
	p.state = ProgressSucceeded
	p.failure = ""
}

// Fail marks the operation failed.
func (p *Progress) Fail(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = ProgressFailed
	if err == nil {
		p.failure = ""
	} else {
		p.failure = err.Error()
	}
}

// Cancel marks the operation cancelled with an optional reason.
func (p *Progress) Cancel(reason string) {
	p.mu.Lock()
	p.state = ProgressCancelled
	p.detail = reason
	p.mu.Unlock()
}

// Snapshot returns a copy of the operation state.
func (p *Progress) Snapshot() ProgressSnapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return ProgressSnapshot{
		Label: p.label, Detail: p.detail, State: p.state,
		Current: p.current, Total: p.total, Error: p.failure,
	}
}

// Frame builds the operation's current renderer-native view.
func (p *Progress) Frame(width int) *Frame {
	p.mu.RLock()
	snapshot := ProgressSnapshot{
		Label: p.label, Detail: p.detail, State: p.state,
		Current: p.current, Total: p.total, Error: p.failure,
	}
	showPercentage := p.showPercentage
	showValues := p.showValues
	theme := p.theme
	clock := p.clock
	p.mu.RUnlock()

	detail := snapshot.Detail
	if snapshot.State == ProgressFailed && snapshot.Error != "" {
		detail = snapshot.Error
	}
	height := 1
	if detail != "" {
		height++
	}
	frame := NewFrame(max(width, 0), height)
	marker, style, status := progressAppearance(snapshot, theme, clock())
	determinate := snapshot.Total > 0
	fraction := 0.0
	if determinate {
		fraction = float64(min(snapshot.Current, snapshot.Total)) / float64(snapshot.Total)
	}
	if determinate && snapshot.State != ProgressSucceeded && snapshot.State != ProgressFailed && snapshot.State != ProgressCancelled {
		switch {
		case showValues:
			status = formatProgressValues(snapshot.Current, snapshot.Total)
		case showPercentage:
			status = formatProgressPercent(fraction)
		default:
			status = ""
		}
	}
	drawProgressRowWithTheme(frame, 0, max(width, 0), marker, snapshot.Label, fraction, determinate, style, status, theme)
	if detail != "" {
		drawClipped(frame, 2, 1, detail, max(width-2, 0), tcell.StyleDefault.Dim(true))
	}
	return frame
}

func progressAppearance(snapshot ProgressSnapshot, theme StatusTheme, now time.Time) (string, tcell.Style, string) {
	return semanticAppearance(progressSemanticStatus(snapshot.State), true, theme, now.UnixMilli())
}

func progressSemanticStatus(state ProgressState) semanticStatus {
	switch state {
	case ProgressActive:
		return statusActive
	case ProgressSucceeded:
		return statusSucceeded
	case ProgressFailed:
		return statusFailed
	case ProgressCancelled:
		return statusCancelled
	default:
		return statusPending
	}
}

func progressFinished(state ProgressState) bool {
	return semanticFinished(progressSemanticStatus(state))
}

func formatProgressPercent(fraction float64) string {
	return fmt.Sprintf("%3.0f%%", fraction*100)
}

func formatProgressValues(current, total int64) string {
	return fmt.Sprintf("%d/%d", min(max(current, 0), total), total)
}
