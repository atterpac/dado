package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
	"github.com/atterpac/jig/theme/themes"
)

// Splash displays a centered splash screen with optional gradient logo,
// status text, and configurable dismissal behavior.
type Splash struct {
	*tview.Flex
	logo          string
	status        string
	content       tview.Primitive // Custom content (overrides logo/status when set)
	gradientType  theme.GradientType
	colors        []string
	onClose       func()
	autoDismiss   time.Duration
	dismissKeys   []DismissKey
	devMode       bool
	themeIndex    int
	logoText      *tview.TextView
	statusText    *tview.TextView
	logoContainer *tview.Flex
	topSpacer     *tview.Box
	bottomSpacer  *tview.Box
	leftSpacer    *tview.Box
	rightSpacer   *tview.Box
	logoWidth     int
	logoHeight    int
	statusHeight  int
	timer         *time.Timer
}

// DismissKey represents a key that can dismiss the splash screen.
type DismissKey struct {
	Key  tcell.Key
	Rune rune
}

// Common dismiss keys
var (
	DismissEscape = DismissKey{Key: tcell.KeyEscape}
	DismissEnter  = DismissKey{Key: tcell.KeyEnter}
	DismissSpace  = DismissKey{Rune: ' '}
	DismissQ      = DismissKey{Rune: 'q'}
	DismissAnyKey = DismissKey{Key: tcell.KeyRune, Rune: 0} // Special: any key
)

// DefaultDismissKeys returns the default set of dismiss keys.
func DefaultDismissKeys() []DismissKey {
	return []DismissKey{DismissEscape, DismissEnter, DismissSpace, DismissQ}
}

// NewSplash creates a new splash screen component.
func NewSplash() *Splash {
	s := &Splash{
		Flex:         tview.NewFlex().SetDirection(tview.FlexRow),
		gradientType: theme.GradientDiagonal,
		colors:       nil, // Will use theme defaults
		dismissKeys:  DefaultDismissKeys(),
		logoWidth:    90,
		logoHeight:   12,
		statusHeight: 5, // Default height for status + sponsor
	}
	return s
}

// SetLogo sets the ASCII art logo to display.
func (s *Splash) SetLogo(logo string) *Splash {
	s.logo = logo
	s.content = nil // Clear custom content
	s.calculateLogoDimensions()
	return s
}

// SetStatus sets the status/hint text below the logo.
func (s *Splash) SetStatus(status string) *Splash {
	s.status = status
	return s
}

// SetContent sets a custom primitive to display instead of logo/status.
// This overrides SetLogo and SetStatus.
func (s *Splash) SetContent(content tview.Primitive) *Splash {
	s.content = content
	return s
}

// SetGradient sets the gradient type for the logo.
func (s *Splash) SetGradient(gradientType theme.GradientType) *Splash {
	s.gradientType = gradientType
	return s
}

// SetColors sets custom gradient colors. If nil, uses theme defaults.
func (s *Splash) SetColors(colors []string) *Splash {
	s.colors = colors
	return s
}

// SetAutoDismiss sets automatic dismissal after the given duration.
// Set to 0 to disable auto-dismiss.
func (s *Splash) SetAutoDismiss(d time.Duration) *Splash {
	s.autoDismiss = d
	return s
}

// SetDismissKeys sets which keys can dismiss the splash.
// Pass nil or empty slice to disable key dismissal.
func (s *Splash) SetDismissKeys(keys []DismissKey) *Splash {
	s.dismissKeys = keys
	return s
}

// SetDevMode enables or disables dev mode (T/G key cycling).
func (s *Splash) SetDevMode(enabled bool) *Splash {
	s.devMode = enabled
	return s
}

// SetOnClose sets the callback when the splash is closed.
func (s *Splash) SetOnClose(fn func()) *Splash {
	s.onClose = fn
	return s
}

// SetLogoWidth sets the fixed width for the logo container.
func (s *Splash) SetLogoWidth(width int) *Splash {
	s.logoWidth = width
	return s
}

// SetLogoHeight sets the fixed height for the logo container.
func (s *Splash) SetLogoHeight(height int) *Splash {
	s.logoHeight = height
	return s
}

// SetStatusHeight sets the height for the status text area.
func (s *Splash) SetStatusHeight(height int) *Splash {
	s.statusHeight = height
	return s
}

// calculateLogoDimensions calculates width/height from logo content.
func (s *Splash) calculateLogoDimensions() {
	if s.logo == "" {
		return
	}
	lines := strings.Split(s.logo, "\n")
	s.logoHeight = len(lines) + 2 // Add padding

	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	s.logoWidth = maxWidth + 2 // Add padding
}

// Build initializes the splash layout. Call this after setting all options.
func (s *Splash) Build() *Splash {
	// Clear existing items
	s.Clear()

	if s.content != nil {
		// Custom content mode - just center it
		s.buildCustomContentLayout()
	} else {
		// Logo + status mode
		s.buildLogoLayout()
	}

	// Start auto-dismiss timer if configured
	if s.autoDismiss > 0 {
		s.startTimer()
	}

	return s
}

func (s *Splash) buildCustomContentLayout() {
	topSpacer := tview.NewBox().SetBackgroundColor(theme.Bg())
	bottomSpacer := tview.NewBox().SetBackgroundColor(theme.Bg())
	leftSpacer := tview.NewBox().SetBackgroundColor(theme.Bg())
	rightSpacer := tview.NewBox().SetBackgroundColor(theme.Bg())

	centerRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftSpacer, 0, 1, false).
		AddItem(s.content, 0, 0, true).
		AddItem(rightSpacer, 0, 1, false)
	centerRow.SetBackgroundColor(theme.Bg())

	s.AddItem(topSpacer, 0, 1, false)
	s.AddItem(centerRow, 0, 1, true)
	s.AddItem(bottomSpacer, 0, 1, false)
	s.SetBackgroundColor(theme.Bg())
}

func (s *Splash) buildLogoLayout() {
	s.logoText = tview.NewTextView()
	s.logoText.SetDynamicColors(true)
	s.logoText.SetTextAlign(tview.AlignLeft)
	s.logoText.SetBackgroundColor(theme.Bg())

	s.statusText = tview.NewTextView()
	s.statusText.SetDynamicColors(true)
	s.statusText.SetTextAlign(tview.AlignCenter)
	s.statusText.SetBackgroundColor(theme.Bg())

	s.updateLogo()
	s.updateStatus()

	// Create spacer boxes
	s.topSpacer = tview.NewBox().SetBackgroundColor(theme.Bg())
	s.bottomSpacer = tview.NewBox().SetBackgroundColor(theme.Bg())
	s.leftSpacer = tview.NewBox().SetBackgroundColor(theme.Bg())
	s.rightSpacer = tview.NewBox().SetBackgroundColor(theme.Bg())

	// Logo container centered horizontally
	s.logoContainer = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(s.leftSpacer, 0, 1, false).
		AddItem(s.logoText, s.logoWidth, 0, false).
		AddItem(s.rightSpacer, 0, 1, false)
	s.logoContainer.SetBackgroundColor(theme.Bg())

	// Build vertical layout
	statusHeight := s.statusHeight
	if s.devMode && statusHeight < 5 {
		statusHeight = 5 // Extra line for dev mode hints
	}

	s.AddItem(s.topSpacer, 0, 1, false)
	s.AddItem(s.logoContainer, s.logoHeight, 0, false)
	s.AddItem(s.statusText, statusHeight, 0, false)
	s.AddItem(s.bottomSpacer, 0, 1, false)
	s.SetBackgroundColor(theme.Bg())
}

func (s *Splash) updateLogo() {
	if s.logoText == nil || s.logo == "" {
		return
	}

	colors := s.colors
	if colors == nil {
		colors = theme.DefaultGradientColors()
	}

	gradientLogo := theme.ApplyGradient(s.logo, s.gradientType, colors)
	s.logoText.SetText(gradientLogo)
}

func (s *Splash) updateStatus() {
	if s.statusText == nil {
		return
	}

	var statusBuilder strings.Builder

	if s.status != "" {
		statusBuilder.WriteString(fmt.Sprintf("[%s]%s[-]\n", theme.TagFg(), s.status))
	}

	if s.devMode {
		themeName := s.getCurrentThemeName()
		statusBuilder.WriteString(fmt.Sprintf(
			"[%s]Theme: [%s::b]%s[-:-:-]  [%s]Gradient: [%s::b]%s[-:-:-]\n[%s][T] Theme  [G] Gradient  [Esc] Close[-]",
			theme.TagFgDim(), theme.TagAccent(), themeName,
			theme.TagFgDim(), theme.TagAccent(), s.gradientType.String(),
			theme.TagFgDim(),
		))
	}

	s.statusText.SetText(statusBuilder.String())
}

func (s *Splash) getCurrentThemeName() string {
	names := themes.Names()
	if s.themeIndex >= 0 && s.themeIndex < len(names) {
		return names[s.themeIndex]
	}
	return "unknown"
}

func (s *Splash) cycleGradient() {
	s.gradientType = s.gradientType.Next()
	s.updateLogo()
	s.updateStatus()
}

func (s *Splash) cycleTheme() {
	names := themes.Names()
	if len(names) == 0 {
		return
	}

	s.themeIndex = (s.themeIndex + 1) % len(names)
	newTheme := themes.Get(names[s.themeIndex])
	if newTheme != nil {
		theme.SetProvider(newTheme)
		s.refreshColors()
		s.updateLogo()
		s.updateStatus()
	}
}

func (s *Splash) refreshColors() {
	bg := theme.Bg()

	if s.logoText != nil {
		s.logoText.SetBackgroundColor(bg)
	}
	if s.statusText != nil {
		s.statusText.SetBackgroundColor(bg)
	}
	if s.logoContainer != nil {
		s.logoContainer.SetBackgroundColor(bg)
	}
	if s.topSpacer != nil {
		s.topSpacer.SetBackgroundColor(bg)
	}
	if s.bottomSpacer != nil {
		s.bottomSpacer.SetBackgroundColor(bg)
	}
	if s.leftSpacer != nil {
		s.leftSpacer.SetBackgroundColor(bg)
	}
	if s.rightSpacer != nil {
		s.rightSpacer.SetBackgroundColor(bg)
	}
	s.SetBackgroundColor(bg)
}

func (s *Splash) startTimer() {
	s.timer = time.AfterFunc(s.autoDismiss, func() {
		theme.QueueUpdateDraw(func() {
			s.Close()
		})
	})
}

// Close triggers the close callback and stops any running timer.
func (s *Splash) Close() {
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	if s.onClose != nil {
		s.onClose()
	}
}

// Draw fills the entire background before drawing children.
func (s *Splash) Draw(screen tcell.Screen) {
	// Fill entire screen with background color
	width, height := screen.Size()
	bgStyle := tcell.StyleDefault.Background(theme.Bg())
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			screen.SetContent(col, row, ' ', nil, bgStyle)
		}
	}

	// Draw children on top
	s.Flex.Draw(screen)
}

// InputHandler handles keyboard input.
func (s *Splash) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return s.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		// Dev mode controls
		if s.devMode {
			switch event.Rune() {
			case 'T':
				s.cycleTheme()
				return
			case 'G', 'g':
				s.cycleGradient()
				return
			}
		}

		// Check dismiss keys
		if s.shouldDismiss(event) {
			s.Close()
			return
		}

		// Delegate to content if present
		if s.content != nil {
			if handler := s.content.InputHandler(); handler != nil {
				handler(event, setFocus)
			}
		}
	})
}

func (s *Splash) shouldDismiss(event *tcell.EventKey) bool {
	for _, dk := range s.dismissKeys {
		// Special case: DismissAnyKey matches any key press
		if dk.Key == tcell.KeyRune && dk.Rune == 0 {
			return true
		}

		// Match specific key
		if dk.Key != tcell.KeyRune && event.Key() == dk.Key {
			return true
		}

		// Match specific rune
		if dk.Rune != 0 && event.Rune() == dk.Rune {
			return true
		}
	}
	return false
}

// Focus handles focus.
func (s *Splash) Focus(delegate func(p tview.Primitive)) {
	if s.content != nil {
		delegate(s.content)
	} else {
		s.Flex.Focus(delegate)
	}
}

// HasFocus returns whether the splash has focus.
func (s *Splash) HasFocus() bool {
	if s.content != nil {
		return s.content.HasFocus()
	}
	return s.Flex.HasFocus()
}

// GetGradientType returns the current gradient type.
func (s *Splash) GetGradientType() theme.GradientType {
	return s.gradientType
}

// GetThemeIndex returns the current theme index (for dev mode).
func (s *Splash) GetThemeIndex() int {
	return s.themeIndex
}

// SetThemeIndex sets the current theme index (for dev mode).
func (s *Splash) SetThemeIndex(index int) *Splash {
	names := themes.Names()
	if index >= 0 && index < len(names) {
		s.themeIndex = index
	}
	return s
}
