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
	// ErrStepNotFound is returned when a step ID is not present.
	ErrStepNotFound = errors.New("inline: step not found")
	// ErrStepExists is returned when a step ID has already been added.
	ErrStepExists = errors.New("inline: step already exists")
	// ErrInvalidStepID is returned when a step ID is empty.
	ErrInvalidStepID = errors.New("inline: step ID cannot be empty")
	// ErrNoActiveStep is returned when Advance is called before a step is active.
	ErrNoActiveStep = errors.New("inline: no active step")
	// ErrStepperFinished is returned when Start has no pending step to activate.
	ErrStepperFinished = errors.New("inline: stepper is finished")
)

// StepState describes a step's lifecycle state.
type StepState int

const (
	StepPending StepState = iota
	StepActive
	StepComplete
	StepFailed
	StepSkipped
)

// String returns the lowercase step-state name.
func (state StepState) String() string {
	switch state {
	case StepActive:
		return "active"
	case StepComplete:
		return "complete"
	case StepFailed:
		return "failed"
	case StepSkipped:
		return "skipped"
	default:
		return "pending"
	}
}

// StepOrientation selects the stepper's layout.
type StepOrientation int

const (
	StepVertical StepOrientation = iota
	StepHorizontal
)

// StepSnapshot is an immutable view of a step's state.
type StepSnapshot struct {
	ID         string
	Label      string
	Detail     string
	State      StepState
	Error      string
	StartedAt  time.Time
	FinishedAt time.Time
}

// StepperOption configures a Stepper.
type StepperOption func(*Stepper)

// WithStepOrientation sets a vertical or horizontal layout.
func WithStepOrientation(orientation StepOrientation) StepperOption {
	return func(stepper *Stepper) { stepper.orientation = orientation }
}

// WithMaxVisibleSteps bounds the number of steps in a frame. The window
// follows the active or most recently finished step. Values below one show all.
func WithMaxVisibleSteps(maximum int) StepperOption {
	return func(stepper *Stepper) { stepper.maxVisible = max(maximum, 0) }
}

// WithStepDetails controls whether detail and error text is rendered.
func WithStepDetails(visible bool) StepperOption {
	return func(stepper *Stepper) { stepper.showDetails = visible }
}

// WithStepperTheme supplies shared semantic styles and glyphs.
func WithStepperTheme(theme StatusTheme) StepperOption {
	return func(stepper *Stepper) { stepper.theme = normalizedStatusTheme(theme) }
}

// WithStepperClock supplies lifecycle timestamps, primarily for tests.
func WithStepperClock(clock func() time.Time) StepperOption {
	return func(stepper *Stepper) {
		if clock != nil {
			stepper.clock = clock
		}
	}
}

// Stepper renders an ordered workflow without writing to the terminal. It is
// safe to update from one goroutine while another builds Frame snapshots.
type Stepper struct {
	mu sync.RWMutex

	title       string
	steps       []StepSnapshot
	byID        map[string]int
	orientation StepOrientation
	maxVisible  int
	showDetails bool
	theme       StatusTheme
	clock       func() time.Time
}

// NewStepper creates an empty ordered workflow.
func NewStepper(title string, options ...StepperOption) *Stepper {
	stepper := &Stepper{
		title:       title,
		byID:        make(map[string]int),
		showDetails: true,
		theme:       DefaultStatusTheme(),
		clock:       time.Now,
	}
	for _, option := range options {
		if option != nil {
			option(stepper)
		}
	}
	return stepper
}

// Add appends a pending step.
func (s *Stepper) Add(id, label string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidStepID
	}
	if label == "" {
		label = id
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byID[id]; exists {
		return fmt.Errorf("%w: %s", ErrStepExists, id)
	}
	s.byID[id] = len(s.steps)
	s.steps = append(s.steps, StepSnapshot{ID: id, Label: label})
	return nil
}

// Start activates the first pending step. Calling Start while a step is active
// is a no-op. ErrStepperFinished is returned when no pending step remains.
func (s *Stepper) Start() error {
	now := s.clock()
	s.mu.Lock()
	defer s.mu.Unlock()
	if activeStep(s.steps) >= 0 {
		return nil
	}
	index := nextPendingStep(s.steps, 0)
	if index < 0 {
		return ErrStepperFinished
	}
	activateStep(&s.steps[index], now)
	return nil
}

// Activate makes id the active step. A different active step returns to
// pending. Activating a failed step retries it and clears its failure.
func (s *Stepper) Activate(id string) error {
	now := s.clock()
	s.mu.Lock()
	defer s.mu.Unlock()
	index, exists := s.byID[id]
	if !exists {
		return fmt.Errorf("%w: %s", ErrStepNotFound, id)
	}
	for i := range s.steps {
		if i != index && s.steps[i].State == StepActive {
			s.steps[i].State = StepPending
		}
	}
	activateStep(&s.steps[index], now)
	return nil
}

// Advance completes the active step and activates the next pending step. The
// final advance completes the workflow without returning an error.
func (s *Stepper) Advance() error {
	now := s.clock()
	s.mu.Lock()
	defer s.mu.Unlock()
	index := activeStep(s.steps)
	if index < 0 {
		return ErrNoActiveStep
	}
	finishStep(&s.steps[index], StepComplete, now, "", s.steps[index].Detail)
	if next := nextPendingStep(s.steps, index+1); next >= 0 {
		activateStep(&s.steps[next], now)
	}
	return nil
}

// Complete marks id complete without activating another step.
func (s *Stepper) Complete(id string) error {
	return s.mutate(id, func(step *StepSnapshot, now time.Time) {
		finishStep(step, StepComplete, now, "", step.Detail)
	})
}

// Fail marks id failed and records err. Activate can retry the step.
func (s *Stepper) Fail(id string, err error) error {
	message := ""
	if err != nil {
		message = err.Error()
	}
	return s.mutate(id, func(step *StepSnapshot, now time.Time) {
		finishStep(step, StepFailed, now, message, step.Detail)
	})
}

// Skip marks id skipped with an optional reason.
func (s *Stepper) Skip(id, reason string) error {
	return s.mutate(id, func(step *StepSnapshot, now time.Time) {
		finishStep(step, StepSkipped, now, "", reason)
	})
}

// SetDetail sets secondary text rendered beneath a vertical step.
func (s *Stepper) SetDetail(id, detail string) error {
	return s.mutate(id, func(step *StepSnapshot, _ time.Time) { step.Detail = detail })
}

// Steps returns an insertion-ordered copy of the workflow state.
func (s *Stepper) Steps() []StepSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]StepSnapshot(nil), s.steps...)
}

// Frame builds the stepper's current renderer-native view.
func (s *Stepper) Frame(width int) *Frame {
	width = max(width, 0)
	s.mu.RLock()
	title := s.title
	steps := append([]StepSnapshot(nil), s.steps...)
	orientation := s.orientation
	maximum := s.maxVisible
	showDetails := s.showDetails
	theme := s.theme
	s.mu.RUnlock()

	visible, before, after := visibleSteps(steps, maximum)
	if orientation == StepHorizontal {
		return horizontalStepperFrame(width, title, visible, before, after, theme)
	}
	return verticalStepperFrame(width, title, visible, before, after, showDetails, theme)
}

func (s *Stepper) mutate(id string, mutate func(*StepSnapshot, time.Time)) error {
	now := s.clock()
	s.mu.Lock()
	defer s.mu.Unlock()
	index, exists := s.byID[id]
	if !exists {
		return fmt.Errorf("%w: %s", ErrStepNotFound, id)
	}
	mutate(&s.steps[index], now)
	return nil
}

func activateStep(step *StepSnapshot, now time.Time) {
	if step.StartedAt.IsZero() {
		step.StartedAt = now
	}
	step.State = StepActive
	step.FinishedAt = time.Time{}
	step.Error = ""
}

func finishStep(step *StepSnapshot, state StepState, now time.Time, failure, detail string) {
	if step.StartedAt.IsZero() {
		step.StartedAt = now
	}
	step.State = state
	step.FinishedAt = now
	step.Error = failure
	if detail != "" {
		step.Detail = detail
	}
}

func activeStep(steps []StepSnapshot) int {
	for i := range steps {
		if steps[i].State == StepActive {
			return i
		}
	}
	return -1
}

func nextPendingStep(steps []StepSnapshot, start int) int {
	for i := max(start, 0); i < len(steps); i++ {
		if steps[i].State == StepPending {
			return i
		}
	}
	return -1
}

func visibleSteps(steps []StepSnapshot, maximum int) (visible []StepSnapshot, before, after int) {
	if maximum < 1 || len(steps) <= maximum {
		return steps, 0, 0
	}
	focus := activeStep(steps)
	if focus < 0 {
		focus = 0
		for i := range steps {
			if steps[i].State != StepPending {
				focus = i
			}
		}
	}
	start := max(focus-maximum/2, 0)
	start = min(start, len(steps)-maximum)
	end := start + maximum
	return steps[start:end], start, len(steps) - end
}

func verticalStepperFrame(width int, title string, steps []StepSnapshot, before, after int, showDetails bool, theme StatusTheme) *Frame {
	height := 0
	if title != "" {
		height++
	}
	if before > 0 {
		height++
	}
	for i, step := range steps {
		height++
		detail := stepDetail(step, showDetails)
		if detail != "" || i < len(steps)-1 || after > 0 {
			height++
		}
	}
	if after > 0 {
		height++
	}
	frame := NewFrame(width, height)
	y := 0
	if title != "" {
		drawClipped(frame, 0, y, title, width, tcell.StyleDefault.Bold(true))
		y++
	}
	if before > 0 {
		drawClipped(frame, 0, y, "…", 1, tcell.StyleDefault.Dim(true))
		drawClipped(frame, 2, y, fmt.Sprintf("%d earlier steps", before), max(width-2, 0), tcell.StyleDefault.Dim(true))
		y++
	}
	for i, step := range steps {
		marker, style := stepAppearance(step.State, theme)
		drawClipped(frame, 0, y, marker, 1, style)
		drawClipped(frame, 2, y, step.Label, max(width-2, 0), stepLabelStyle(step.State))
		y++
		detail := stepDetail(step, showDetails)
		if detail != "" || i < len(steps)-1 || after > 0 {
			drawClipped(frame, 0, y, "│", 1, tcell.StyleDefault.Dim(true))
			drawClipped(frame, 2, y, detail, max(width-2, 0), tcell.StyleDefault.Dim(true))
			y++
		}
	}
	if after > 0 {
		drawClipped(frame, 0, y, "…", 1, tcell.StyleDefault.Dim(true))
		drawClipped(frame, 2, y, fmt.Sprintf("%d later steps", after), max(width-2, 0), tcell.StyleDefault.Dim(true))
	}
	return frame
}

func horizontalStepperFrame(width int, title string, steps []StepSnapshot, before, after int, theme StatusTheme) *Frame {
	height := 1
	if title != "" {
		height++
	}
	frame := NewFrame(width, height)
	y := 0
	if title != "" {
		drawClipped(frame, 0, y, title, width, tcell.StyleDefault.Bold(true))
		y++
	}
	parts := len(steps)
	if before > 0 {
		parts++
	}
	if after > 0 {
		parts++
	}
	if parts == 0 {
		return frame
	}
	separator := " ─ "
	separatorWidth := displayWidth(separator)
	partWidth := max((width-separatorWidth*(parts-1))/parts, 1)
	x := 0
	drawPart := func(marker, label string, state StepState, style tcell.Style) {
		if x > 0 {
			drawClipped(frame, x, y, separator, separatorWidth, tcell.StyleDefault.Dim(true))
			x += separatorWidth
		}
		drawClipped(frame, x, y, marker, 1, style)
		drawClipped(frame, x+2, y, label, max(partWidth-2, 0), stepLabelStyle(state))
		x += partWidth
	}
	if before > 0 {
		drawPart("…", fmt.Sprintf("%d earlier", before), StepPending, tcell.StyleDefault.Dim(true))
	}
	for _, step := range steps {
		marker, style := stepAppearance(step.State, theme)
		drawPart(marker, step.Label, step.State, style)
	}
	if after > 0 {
		drawPart("…", fmt.Sprintf("%d later", after), StepPending, tcell.StyleDefault.Dim(true))
	}
	return frame
}

func stepDetail(step StepSnapshot, visible bool) string {
	if !visible {
		return ""
	}
	if step.State == StepFailed && step.Error != "" {
		return step.Error
	}
	return step.Detail
}

func stepAppearance(state StepState, theme StatusTheme) (string, tcell.Style) {
	switch state {
	case StepActive:
		return theme.ActiveMarker, theme.ActiveStyle
	case StepComplete:
		return theme.SuccessMarker, theme.SuccessStyle
	case StepFailed:
		return theme.FailureMarker, theme.FailureStyle
	case StepSkipped:
		return theme.SkippedMarker, theme.PendingStyle
	default:
		return theme.PendingMarker, theme.PendingStyle
	}
}

func stepLabelStyle(state StepState) tcell.Style {
	if state == StepActive {
		return tcell.StyleDefault.Bold(true)
	}
	if state == StepPending || state == StepSkipped {
		return tcell.StyleDefault.Dim(true)
	}
	return tcell.StyleDefault
}
