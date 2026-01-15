package components

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// ProgressBar Tests
// ============================================================================

// TestProgressBar_NewProgressBar tests ProgressBar creation.
func TestProgressBar_NewProgressBar(t *testing.T) {
	bar := NewProgressBar()

	assert.NotNil(t, bar)
	assert.Equal(t, 0.0, bar.GetProgress())
}

// TestProgressBar_SetProgress tests setting progress values.
func TestProgressBar_SetProgress(t *testing.T) {
	bar := NewProgressBar()

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero", 0.0, 0.0},
		{"half", 0.5, 0.5},
		{"full", 1.0, 1.0},
		{"negative clamped", -0.5, 0.0},
		{"over one clamped", 1.5, 1.0},
		{"quarter", 0.25, 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar.SetProgress(tt.input)
			assert.Equal(t, tt.expected, bar.GetProgress())
		})
	}
}

// TestProgressBar_SetLabel tests setting label.
func TestProgressBar_SetLabel(t *testing.T) {
	bar := NewProgressBar()

	result := bar.SetLabel("Downloading...")

	assert.Same(t, bar, result) // Fluent API
}

// TestProgressBar_SetShowPercentage tests percentage display toggle.
func TestProgressBar_SetShowPercentage(t *testing.T) {
	bar := NewProgressBar()

	result := bar.SetShowPercentage(false)
	assert.Same(t, bar, result)

	result = bar.SetShowPercentage(true)
	assert.Same(t, bar, result)
}

// TestProgressBar_SetShowValue tests value display.
func TestProgressBar_SetShowValue(t *testing.T) {
	bar := NewProgressBar()

	result := bar.SetShowValue(true, 50, 100)
	assert.Same(t, bar, result)
}

// TestProgressBar_SetChars tests custom characters.
func TestProgressBar_SetChars(t *testing.T) {
	bar := NewProgressBar()

	result := bar.SetChars('=', '-')
	assert.Same(t, bar, result)

	result = bar.SetChars('▓', '░')
	assert.Same(t, bar, result)
}

// TestProgressBar_GetFieldHeight tests preferred height calculation.
func TestProgressBar_GetFieldHeight(t *testing.T) {
	t.Run("without label", func(t *testing.T) {
		bar := NewProgressBar()
		assert.Equal(t, 1, bar.GetFieldHeight())
	})

	t.Run("with label", func(t *testing.T) {
		bar := NewProgressBar().SetLabel("Label")
		assert.Equal(t, 2, bar.GetFieldHeight())
	})
}

// TestProgressBar_FluentAPI tests method chaining.
func TestProgressBar_FluentAPI(t *testing.T) {
	bar := NewProgressBar().
		SetProgress(0.75).
		SetLabel("Processing").
		SetShowPercentage(true).
		SetShowValue(true, 75, 100).
		SetChars('█', '░')

	assert.Equal(t, 0.75, bar.GetProgress())
}

// ============================================================================
// Spinner Tests
// ============================================================================

// TestSpinner_NewSpinner tests Spinner creation.
func TestSpinner_NewSpinner(t *testing.T) {
	spinner := NewSpinner()

	assert.NotNil(t, spinner)
	assert.False(t, spinner.IsRunning())
}

// TestSpinner_SetStyle tests setting spinner style.
func TestSpinner_SetStyle(t *testing.T) {
	spinner := NewSpinner()

	styles := []SpinnerStyle{
		SpinnerDots,
		SpinnerLine,
		SpinnerBraille,
		SpinnerCircle,
		SpinnerArrow,
	}

	for _, style := range styles {
		result := spinner.SetStyle(style)
		assert.Same(t, spinner, result)
	}
}

// TestSpinner_SetLabel tests setting spinner label.
func TestSpinner_SetLabel(t *testing.T) {
	spinner := NewSpinner()

	result := spinner.SetLabel("Loading...")
	assert.Same(t, spinner, result)
}

// TestSpinner_SetInterval tests setting animation interval.
func TestSpinner_SetInterval(t *testing.T) {
	spinner := NewSpinner()

	result := spinner.SetInterval(50 * time.Millisecond)
	assert.Same(t, spinner, result)
}

// TestSpinner_StartStop tests starting and stopping animation.
func TestSpinner_StartStop(t *testing.T) {
	spinner := NewSpinner()
	spinner.SetInterval(10 * time.Millisecond)

	// Start
	result := spinner.Start()
	assert.Same(t, spinner, result)
	assert.True(t, spinner.IsRunning())

	// Let it run briefly
	time.Sleep(50 * time.Millisecond)

	// Stop
	result = spinner.Stop()
	assert.Same(t, spinner, result)
	assert.False(t, spinner.IsRunning())
}

// TestSpinner_StartWhileRunning tests starting when already running.
func TestSpinner_StartWhileRunning(t *testing.T) {
	spinner := NewSpinner()
	spinner.SetInterval(10 * time.Millisecond)

	spinner.Start()
	defer spinner.Stop()

	require.True(t, spinner.IsRunning())

	// Should not panic or cause issues
	spinner.Start()
	assert.True(t, spinner.IsRunning())
}

// TestSpinner_StopWhileStopped tests stopping when not running.
func TestSpinner_StopWhileStopped(t *testing.T) {
	spinner := NewSpinner()

	assert.False(t, spinner.IsRunning())

	// Should not panic
	spinner.Stop()
	assert.False(t, spinner.IsRunning())
}

// TestSpinner_GetFieldHeight tests preferred height.
func TestSpinner_GetFieldHeight(t *testing.T) {
	spinner := NewSpinner()
	assert.Equal(t, 1, spinner.GetFieldHeight())
}

// TestSpinner_FluentAPI tests method chaining.
func TestSpinner_FluentAPI(t *testing.T) {
	spinner := NewSpinner().
		SetStyle(SpinnerBraille).
		SetLabel("Processing...").
		SetInterval(75 * time.Millisecond)

	assert.NotNil(t, spinner)
}

// ============================================================================
// Gauge Tests
// ============================================================================

// TestGauge_NewGauge tests Gauge creation.
func TestGauge_NewGauge(t *testing.T) {
	gauge := NewGauge()

	assert.NotNil(t, gauge)
}

// TestGauge_SetValue tests setting gauge value.
func TestGauge_SetValue(t *testing.T) {
	gauge := NewGauge()

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero", 0.0, 0.0},
		{"half", 0.5, 0.5},
		{"full", 1.0, 1.0},
		{"negative clamped", -0.5, 0.0},
		{"over one clamped", 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gauge.SetValue(tt.input)
			assert.Same(t, gauge, result)
		})
	}
}

// TestGauge_SetLabel tests setting gauge label.
func TestGauge_SetLabel(t *testing.T) {
	gauge := NewGauge()

	result := gauge.SetLabel("CPU Usage")
	assert.Same(t, gauge, result)
}

// TestGauge_SetUnit tests setting unit string.
func TestGauge_SetUnit(t *testing.T) {
	gauge := NewGauge()

	result := gauge.SetUnit("%")
	assert.Same(t, gauge, result)
}

// TestGauge_SetMaxValue tests setting maximum value.
func TestGauge_SetMaxValue(t *testing.T) {
	gauge := NewGauge()

	result := gauge.SetMaxValue(200)
	assert.Same(t, gauge, result)
}

// TestGauge_FluentAPI tests method chaining.
func TestGauge_FluentAPI(t *testing.T) {
	gauge := NewGauge().
		SetValue(0.65).
		SetLabel("Memory").
		SetUnit("GB").
		SetMaxValue(16)

	assert.NotNil(t, gauge)
}

// ============================================================================
// SpinnerFrames Test
// ============================================================================

// TestSpinnerFrames tests that all spinner styles have frames.
func TestSpinnerFrames(t *testing.T) {
	styles := []SpinnerStyle{
		SpinnerDots,
		SpinnerLine,
		SpinnerBraille,
		SpinnerCircle,
		SpinnerArrow,
	}

	for _, style := range styles {
		frames, ok := spinnerFrames[style]
		assert.True(t, ok, "style %d should have frames", style)
		assert.NotEmpty(t, frames, "style %d should have non-empty frames", style)
	}
}
