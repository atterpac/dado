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

func TestStepperLifecycle(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 14, 9, 0, 0, 0, time.UTC)
	stepper := NewStepper("Release", WithStepperClock(func() time.Time { return now }))
	require.NoError(t, stepper.Add("configure", "Configure"))
	require.NoError(t, stepper.Add("build", "Build"))
	require.NoError(t, stepper.Add("deploy", "Deploy"))

	require.NoError(t, stepper.Start())
	require.NoError(t, stepper.SetDetail("configure", "workspace ready"))
	now = now.Add(time.Second)
	require.NoError(t, stepper.Advance())
	require.NoError(t, stepper.Fail("build", errors.New("tests failed")))

	steps := stepper.Steps()
	assert.Equal(t, StepComplete, steps[0].State)
	assert.Equal(t, StepFailed, steps[1].State)
	assert.Equal(t, "tests failed", steps[1].Error)
	assert.Equal(t, StepPending, steps[2].State)
	assert.False(t, steps[0].FinishedAt.IsZero())

	require.NoError(t, stepper.Activate("build"))
	assert.Equal(t, StepActive, stepper.Steps()[1].State)
	assert.Empty(t, stepper.Steps()[1].Error)
	require.NoError(t, stepper.Advance())
	assert.Equal(t, StepActive, stepper.Steps()[2].State)
	require.NoError(t, stepper.Advance())
	assert.Equal(t, StepComplete, stepper.Steps()[2].State)
	assert.ErrorIs(t, stepper.Start(), ErrStepperFinished)
}

func TestStepperActivateMovesCurrentStep(t *testing.T) {
	t.Parallel()

	stepper := NewStepper("")
	require.NoError(t, stepper.Add("one", "One"))
	require.NoError(t, stepper.Add("two", "Two"))
	require.NoError(t, stepper.Start())
	require.NoError(t, stepper.Activate("two"))

	steps := stepper.Steps()
	assert.Equal(t, StepPending, steps[0].State)
	assert.Equal(t, StepActive, steps[1].State)
}

func TestStepperVerticalFrame(t *testing.T) {
	t.Parallel()

	stepper := NewStepper("Ship release")
	require.NoError(t, stepper.Add("prepare", "Prepare"))
	require.NoError(t, stepper.Add("verify", "Verify"))
	require.NoError(t, stepper.Add("publish", "Publish"))
	require.NoError(t, stepper.Complete("prepare"))
	require.NoError(t, stepper.SetDetail("prepare", "version selected"))
	require.NoError(t, stepper.Activate("verify"))
	require.NoError(t, stepper.SetDetail("verify", "running checks"))

	lines := plainFrameLines(stepper.Frame(40))
	assert.Equal(t, "Ship release", lines[0])
	assert.Contains(t, lines[1], "✓ Prepare")
	assert.Contains(t, lines[2], "│ version selected")
	assert.Contains(t, lines[3], "● Verify")
	assert.Contains(t, lines[4], "│ running checks")
	assert.Contains(t, lines[5], "○ Publish")
}

func TestStepperHorizontalFrame(t *testing.T) {
	t.Parallel()

	stepper := NewStepper("Deploy", WithStepOrientation(StepHorizontal))
	for _, id := range []string{"plan", "apply", "verify"} {
		require.NoError(t, stepper.Add(id, strings.ToUpper(id)))
	}
	require.NoError(t, stepper.Complete("plan"))
	require.NoError(t, stepper.Activate("apply"))

	lines := plainFrameLines(stepper.Frame(60))
	require.Len(t, lines, 2)
	assert.Contains(t, lines[1], "✓ PLAN")
	assert.Contains(t, lines[1], "● APPLY")
	assert.Contains(t, lines[1], "○ VERIFY")
	assert.Contains(t, lines[1], "Step 2 of 3")
	assert.NotContains(t, lines[1], "─")
}

func TestStepperHorizontalFrameUsesCompactWizardHeader(t *testing.T) {
	t.Parallel()

	stepper := NewStepper("", WithStepOrientation(StepHorizontal))
	for _, step := range []struct{ id, label string }{
		{"connector", "Connector"},
		{"configuration", "Configuration"},
		{"review", "Review"},
	} {
		require.NoError(t, stepper.Add(step.id, step.label))
	}
	require.NoError(t, stepper.Activate("connector"))

	line := plainFrameLines(stepper.Frame(64))[0]
	assert.Equal(t, "● Connector   ○ Configuration   ○ Review             Step 1 of 3", line)
}

func TestStepperHorizontalCounterTracksFinishedWorkflow(t *testing.T) {
	t.Parallel()

	stepper := NewStepper("", WithStepOrientation(StepHorizontal))
	require.NoError(t, stepper.Add("one", "One"))
	require.NoError(t, stepper.Add("two", "Two"))
	require.NoError(t, stepper.Complete("one"))
	require.NoError(t, stepper.Complete("two"))

	assert.Contains(t, plainFrameLines(stepper.Frame(40))[0], "Step 2 of 2")
}

func TestStepperCustomTheme(t *testing.T) {
	t.Parallel()

	theme := DefaultStatusTheme()
	theme.ActiveMarker = ">"
	stepper := NewStepper("", WithStepperTheme(theme))
	require.NoError(t, stepper.Add("work", "Work"))
	require.NoError(t, stepper.Start())
	assert.True(t, strings.HasPrefix(plainFrameLines(stepper.Frame(30))[0], ">"))
}

func TestStepperVisibleWindowFollowsActiveStep(t *testing.T) {
	t.Parallel()

	stepper := NewStepper("", WithMaxVisibleSteps(3))
	for i := range 8 {
		id := fmt.Sprintf("step-%d", i)
		require.NoError(t, stepper.Add(id, id))
	}
	require.NoError(t, stepper.Activate("step-5"))

	plain := strings.Join(plainFrameLines(stepper.Frame(40)), "\n")
	assert.Contains(t, plain, "4 earlier steps")
	assert.Contains(t, plain, "step-5")
	assert.Contains(t, plain, "1 later steps")
	assert.NotContains(t, plain, "step-0")
}

func TestStepperValidation(t *testing.T) {
	t.Parallel()

	stepper := NewStepper("")
	assert.ErrorIs(t, stepper.Add("", "Bad"), ErrInvalidStepID)
	require.NoError(t, stepper.Add("one", ""))
	assert.ErrorIs(t, stepper.Add("one", "Again"), ErrStepExists)
	assert.ErrorIs(t, stepper.Activate("missing"), ErrStepNotFound)
	assert.ErrorIs(t, stepper.Advance(), ErrNoActiveStep)
	assert.Equal(t, "one", stepper.Steps()[0].Label)
}

func TestStepperNarrowFramesStayWithinWidth(t *testing.T) {
	t.Parallel()

	for _, orientation := range []StepOrientation{StepVertical, StepHorizontal} {
		stepper := NewStepper("A long workflow title", WithStepOrientation(orientation))
		require.NoError(t, stepper.Add("one", "A very long first step"))
		require.NoError(t, stepper.Add("two", "A very long second step"))
		for width := range 30 {
			for _, line := range plainFrameLines(stepper.Frame(width)) {
				assert.LessOrEqual(t, displayWidth(line), width, "width %d: %q", width, line)
			}
		}
	}
}

func TestStepperConcurrentSnapshots(t *testing.T) {
	t.Parallel()

	stepper := NewStepper("Concurrent")
	require.NoError(t, stepper.Add("work", "Work"))
	require.NoError(t, stepper.Start())
	var workers sync.WaitGroup
	for worker := range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for update := range 100 {
				require.NoError(t, stepper.SetDetail("work", fmt.Sprintf("%d:%d", worker, update)))
				_ = stepper.Frame(40)
			}
		}()
	}
	workers.Wait()
	assert.Equal(t, StepActive, stepper.Steps()[0].State)
}
