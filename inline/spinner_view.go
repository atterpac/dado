package inline

import (
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

// SpinnerState describes the lifecycle of a Spinner frame view.
type SpinnerState int

const (
	SpinnerPending SpinnerState = iota
	SpinnerActive
	SpinnerSucceeded
	SpinnerFailed
	SpinnerCancelled
)

// Spinner is a renderer-native animated activity indicator. It never starts a
// goroutine or writes output; callers sample Frame from their render loop.
type Spinner struct {
	mu sync.RWMutex

	label   string
	detail  string
	failure string
	state   SpinnerState
	theme   StatusTheme
	clock   func() time.Time
}

// NewSpinner creates a pending renderer-native spinner.
func NewSpinner(label string, options ...SpinnerOption) *Spinner {
	spinner := &Spinner{
		label: label,
		theme: DefaultStatusTheme(),
		clock: time.Now,
	}
	for _, option := range options {
		if option != nil {
			option(spinner)
		}
	}
	spinner.theme = normalizedStatusTheme(spinner.theme)
	return spinner
}

// String returns the lowercase spinner-state name.
func (state SpinnerState) String() string {
	switch state {
	case SpinnerActive:
		return "active"
	case SpinnerSucceeded:
		return "succeeded"
	case SpinnerFailed:
		return "failed"
	case SpinnerCancelled:
		return "cancelled"
	default:
		return "pending"
	}
}

// SpinnerSnapshot is an immutable view of Spinner state.
type SpinnerSnapshot struct {
	Label  string
	Detail string
	State  SpinnerState
	Error  string
}

// SpinnerOption configures a Spinner's renderer-native view.
type SpinnerOption func(*Spinner)

// WithSpinnerTheme supplies semantic styles, markers, and animation frames.
func WithSpinnerTheme(theme StatusTheme) SpinnerOption {
	return func(spinner *Spinner) { spinner.theme = normalizedStatusTheme(theme) }
}

// WithSpinnerClock supplies the animation clock, primarily for tests.
func WithSpinnerClock(clock func() time.Time) SpinnerOption {
	return func(spinner *Spinner) {
		if clock != nil {
			spinner.clock = clock
		}
	}
}

// Begin marks the spinner active without starting a goroutine or writing
// output. Render Frame from the application's existing loop.
func (s *Spinner) Begin() {
	s.mu.Lock()
	s.state = SpinnerActive
	s.failure = ""
	s.mu.Unlock()
}

// SetDetail changes the secondary line shown by Frame.
func (s *Spinner) SetDetail(detail string) {
	s.mu.Lock()
	s.detail = detail
	s.mu.Unlock()
}

// Succeed marks the spinner successful.
func (s *Spinner) Succeed(detail string) {
	s.mu.Lock()
	s.state = SpinnerSucceeded
	s.failure = ""
	s.detail = detail
	s.mu.Unlock()
}

// Fail marks the spinner failed.
func (s *Spinner) Fail(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = SpinnerFailed
	if err == nil {
		s.failure = ""
	} else {
		s.failure = err.Error()
	}
}

// Cancel marks the spinner cancelled with an optional reason.
func (s *Spinner) Cancel(reason string) {
	s.mu.Lock()
	s.state = SpinnerCancelled
	s.detail = reason
	s.mu.Unlock()
}

// Snapshot returns a copy of the spinner state.
func (s *Spinner) Snapshot() SpinnerSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return SpinnerSnapshot{Label: s.label, Detail: s.detail, State: s.state, Error: s.failure}
}

// Frame builds the spinner's current view. Animation is sampled from the
// configured clock; Frame never starts its own goroutine.
func (s *Spinner) Frame(width int) *Frame {
	s.mu.RLock()
	snapshot := SpinnerSnapshot{Label: s.label, Detail: s.detail, State: s.state, Error: s.failure}
	theme := s.theme
	clock := s.clock
	s.mu.RUnlock()

	detail := snapshot.Detail
	if snapshot.State == SpinnerFailed && snapshot.Error != "" {
		detail = snapshot.Error
	}
	height := 1
	if detail != "" {
		height++
	}
	frame := NewFrame(max(width, 0), height)
	marker, style, status := spinnerAppearance(snapshot.State, theme, clock())
	drawClipped(frame, 0, 0, marker, 1, style)
	drawClipped(frame, 2, 0, snapshot.Label, max(width-displayWidth(status)-3, 0), spinnerLabelStyle(snapshot.State))
	if status != "" {
		drawRight(frame, 0, max(width, 0), status, style)
	}
	if detail != "" {
		drawClipped(frame, 2, 1, detail, max(width-2, 0), tcell.StyleDefault.Dim(true))
	}
	return frame
}

func spinnerAppearance(state SpinnerState, theme StatusTheme, now time.Time) (string, tcell.Style, string) {
	return semanticAppearance(spinnerSemanticStatus(state), true, theme, now.UnixMilli())
}

func spinnerSemanticStatus(state SpinnerState) semanticStatus {
	switch state {
	case SpinnerActive:
		return statusActive
	case SpinnerSucceeded:
		return statusSucceeded
	case SpinnerFailed:
		return statusFailed
	case SpinnerCancelled:
		return statusCancelled
	default:
		return statusPending
	}
}

func spinnerLabelStyle(state SpinnerState) tcell.Style {
	if state == SpinnerActive {
		return tcell.StyleDefault.Bold(true)
	}
	if state == SpinnerPending {
		return tcell.StyleDefault.Dim(true)
	}
	return tcell.StyleDefault
}
