package components

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
	"github.com/atterpac/dado/theme/themes"
)

// Splash displays a centered splash screen with optional gradient logo,
// status text, and configurable dismissal behavior.
type Splash struct {
	*core.Flex
	widgetBase
	logo          string
	status        string
	content       core.Widget // Custom content (overrides logo/status when set)
	gradientType  theme.GradientType
	colors        []string
	onClose       func()
	autoDismiss   time.Duration
	dismissKeys   []DismissKey
	devMode       bool
	themeIndex    int
	logoText      *core.TextView
	statusText    *core.TextView
	logoContainer *core.Flex
	topSpacer     *core.Box
	bottomSpacer  *core.Box
	leftSpacer    *core.Box
	rightSpacer   *core.Box
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
		Flex:         core.NewFlex(),
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
	s.updateStatus()
	return s
}

// SetContent sets a custom widget to display instead of logo/status.
// This overrides SetLogo and SetStatus.
func (s *Splash) SetContent(content core.Widget) *Splash {
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
		w := utf8.RuneCountInString(line)
		if w > maxWidth {
			maxWidth = w
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
	topSpacer := new(core.Box)
	topSpacer.SetBackgroundColor(theme.Bg())
	bottomSpacer := new(core.Box)
	bottomSpacer.SetBackgroundColor(theme.Bg())
	leftSpacer := new(core.Box)
	leftSpacer.SetBackgroundColor(theme.Bg())
	rightSpacer := new(core.Box)
	rightSpacer.SetBackgroundColor(theme.Bg())

	centerRow := core.NewFlex().SetDirection(core.Row).
		AddItem(leftSpacer, 0, 1, false).
		AddItem(s.content, 0, 0, true).
		AddItem(rightSpacer, 0, 1, false)
	centerRow.Box.SetBackgroundColor(theme.Bg())

	s.AddItem(topSpacer, 0, 1, false)
	s.AddItem(centerRow, 0, 1, true)
	s.AddItem(bottomSpacer, 0, 1, false)
	s.Flex.Box.SetBackgroundColor(theme.Bg())
}

func (s *Splash) buildLogoLayout() {
	s.logoText = core.NewTextView()
	s.logoText.SetDynamicColors(true)
	s.logoText.SetBackgroundColor(theme.Bg())

	s.statusText = core.NewTextView()
	s.statusText.SetDynamicColors(true)
	s.statusText.SetTextAlign(core.AlignCenter)
	s.statusText.SetBackgroundColor(theme.Bg())

	s.updateLogo()
	s.updateStatus()

	// Create spacer boxes
	s.topSpacer = new(core.Box)
	s.topSpacer.SetBackgroundColor(theme.Bg())
	s.bottomSpacer = new(core.Box)
	s.bottomSpacer.SetBackgroundColor(theme.Bg())
	s.leftSpacer = new(core.Box)
	s.leftSpacer.SetBackgroundColor(theme.Bg())
	s.rightSpacer = new(core.Box)
	s.rightSpacer.SetBackgroundColor(theme.Bg())

	// Logo container centered horizontally
	s.logoContainer = core.NewFlex().SetDirection(core.Row).
		AddItem(s.leftSpacer, 0, 1, false).
		AddItem(s.logoText, s.logoWidth, 0, false).
		AddItem(s.rightSpacer, 0, 1, false)
	s.logoContainer.Box.SetBackgroundColor(theme.Bg())

	// Status container centered horizontally (same width as logo)
	statusLeftSpacer := new(core.Box)
	statusLeftSpacer.SetBackgroundColor(theme.Bg())
	statusRightSpacer := new(core.Box)
	statusRightSpacer.SetBackgroundColor(theme.Bg())
	statusContainer := core.NewFlex().SetDirection(core.Row).
		AddItem(statusLeftSpacer, 0, 1, false).
		AddItem(s.statusText, s.logoWidth, 0, false).
		AddItem(statusRightSpacer, 0, 1, false)
	statusContainer.Box.SetBackgroundColor(theme.Bg())

	// Build vertical layout
	statusHeight := s.statusHeight
	if s.devMode && statusHeight < 5 {
		statusHeight = 5 // Extra line for dev mode hints
	}

	s.AddItem(s.topSpacer, 0, 1, false)
	s.AddItem(s.logoContainer, s.logoHeight, 0, false)
	s.AddItem(statusContainer, statusHeight, 0, false)
	s.AddItem(s.bottomSpacer, 0, 1, false)
	s.Flex.Box.SetBackgroundColor(theme.Bg())
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
		statusBuilder.WriteString(s.status + "\n")
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
		s.logoContainer.Box.SetBackgroundColor(bg)
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
	s.Flex.Box.SetBackgroundColor(bg)
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
	fillRect(screen, 0, 0, width, height, bgStyle)

	// Draw children on top
	s.Flex.Draw(screen)
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

// Blur handles blur.
func (s *Splash) Blur() {
	s.Flex.Box.Blur()
}

// Rect implements core.Widget.
func (s *Splash) Rect() (x, y, w, h int) {
	return s.Flex.Box.Rect()
}

// SetRect implements core.Widget.
func (s *Splash) SetRect(x, y, w, h int) {
	s.Flex.Box.SetRect(x, y, w, h)
}

// Focus handles focus.
func (s *Splash) Focus() {
	if s.content != nil {
		s.content.Focus()
	} else {
		s.Flex.Box.Focus()
	}
}

// HasFocus returns whether the splash has focus.
func (s *Splash) HasFocus() bool {
	if s.content != nil {
		return s.content.HasFocus()
	}
	return s.Flex.Box.HasFocus()
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
