package inline

import "github.com/gdamore/tcell/v2"

// StatusTheme contains the semantic styles and glyphs shared by renderer-native
// activity components. Zero-value fields are filled from DefaultStatusTheme.
type StatusTheme struct {
	PendingStyle   tcell.Style
	ActiveStyle    tcell.Style
	SuccessStyle   tcell.Style
	FailureStyle   tcell.Style
	CancelledStyle tcell.Style

	PendingMarker   string
	ActiveMarker    string
	SuccessMarker   string
	FailureMarker   string
	SkippedMarker   string
	CancelledMarker string
	BarFilled       string
	BarEmpty        string
	BarPartial      []string
	SpinnerFrames   []string
}

// DefaultStatusTheme returns Dado's default semantic inline presentation.
func DefaultStatusTheme() StatusTheme {
	return StatusTheme{
		PendingStyle:    tcell.StyleDefault.Dim(true),
		ActiveStyle:     tcell.StyleDefault.Foreground(tcell.ColorAqua),
		SuccessStyle:    tcell.StyleDefault.Foreground(tcell.ColorGreen),
		FailureStyle:    tcell.StyleDefault.Foreground(tcell.ColorRed),
		CancelledStyle:  tcell.StyleDefault.Foreground(tcell.ColorYellow),
		PendingMarker:   "○",
		ActiveMarker:    "●",
		SuccessMarker:   "✓",
		FailureMarker:   "✗",
		SkippedMarker:   "–",
		CancelledMarker: "■",
		BarFilled:       "█",
		BarEmpty:        "░",
		BarPartial:      []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉"},
		SpinnerFrames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

func normalizedStatusTheme(theme StatusTheme) StatusTheme {
	defaults := DefaultStatusTheme()
	if theme.PendingStyle == tcell.StyleDefault {
		theme.PendingStyle = defaults.PendingStyle
	}
	if theme.ActiveStyle == tcell.StyleDefault {
		theme.ActiveStyle = defaults.ActiveStyle
	}
	if theme.SuccessStyle == tcell.StyleDefault {
		theme.SuccessStyle = defaults.SuccessStyle
	}
	if theme.FailureStyle == tcell.StyleDefault {
		theme.FailureStyle = defaults.FailureStyle
	}
	if theme.CancelledStyle == tcell.StyleDefault {
		theme.CancelledStyle = defaults.CancelledStyle
	}
	if theme.PendingMarker == "" {
		theme.PendingMarker = defaults.PendingMarker
	}
	if theme.ActiveMarker == "" {
		theme.ActiveMarker = defaults.ActiveMarker
	}
	if theme.SuccessMarker == "" {
		theme.SuccessMarker = defaults.SuccessMarker
	}
	if theme.FailureMarker == "" {
		theme.FailureMarker = defaults.FailureMarker
	}
	if theme.SkippedMarker == "" {
		theme.SkippedMarker = defaults.SkippedMarker
	}
	if theme.CancelledMarker == "" {
		theme.CancelledMarker = defaults.CancelledMarker
	}
	if theme.BarFilled == "" {
		theme.BarFilled = defaults.BarFilled
	}
	if theme.BarEmpty == "" {
		theme.BarEmpty = defaults.BarEmpty
	}
	if len(theme.BarPartial) == 0 {
		theme.BarPartial = defaults.BarPartial
	} else {
		theme.BarPartial = append([]string(nil), theme.BarPartial...)
	}
	if len(theme.SpinnerFrames) == 0 {
		theme.SpinnerFrames = defaults.SpinnerFrames
	} else {
		theme.SpinnerFrames = append([]string(nil), theme.SpinnerFrames...)
	}
	return theme
}

func spinnerFrame(theme StatusTheme, nowMillis int64) string {
	frames := theme.SpinnerFrames
	if len(frames) == 0 {
		frames = DefaultStatusTheme().SpinnerFrames
	}
	return frames[(nowMillis/80)%int64(len(frames))]
}
