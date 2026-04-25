package layout

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/theme"
)

// Menu is a bottom bar showing key hints and optional right-aligned text.
type Menu struct {
	*tview.Box
	hints     []components.KeyHint
	rightText string
	subs      components.Subscriptions
}

// Subs returns the menu's subscription set; release on app teardown to drop
// the theme registration.
func (m *Menu) Subs() *components.Subscriptions { return &m.subs }

// NewMenu creates a new menu bar.
func NewMenu() *Menu {
	m := &Menu{
		Box:   tview.NewBox(),
		hints: make([]components.KeyHint, 0),
	}

	m.Box.SetBackgroundColor(theme.Bg())

	m.subs.Add(theme.Register(m.Box))

	return m
}

// SetHints sets the key hints to display on the left.
func (m *Menu) SetHints(hints []components.KeyHint) *Menu {
	m.hints = hints
	return m
}

// SetRightText sets text on the right side (e.g., sponsor link).
func (m *Menu) SetRightText(text string) *Menu {
	m.rightText = text
	return m
}

// Clear clears all content.
func (m *Menu) Clear() *Menu {
	m.hints = make([]components.KeyHint, 0)
	m.rightText = ""
	return m
}

// Draw renders the menu with pill-style key hints.
func (m *Menu) Draw(screen tcell.Screen) {
	m.Box.DrawForSubclass(screen, m)

	x, y, width, height := m.GetInnerRect()
	if width < 1 || height < 1 {
		return
	}

	// Get colors from theme
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()

	// Styles
	bgStyle := tcell.StyleDefault.Background(bgColor)
	keyStyle := tcell.StyleDefault.Background(accentColor).Foreground(bgColor).Bold(true)
	descStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
	sepStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
	rightStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)

	// Clear the line
	for i := x; i < x+width; i++ {
		screen.SetContent(i, y, ' ', nil, bgStyle)
	}

	// Draw hints on the left with pill-style keys
	currentX := x + 1 // Small left padding

	for i, hint := range m.hints {
		if i > 0 {
			// Draw separator
			sep := " │ "
			for _, r := range sep {
				if currentX < x+width-len(m.rightText)-2 {
					screen.SetContent(currentX, y, r, nil, sepStyle)
					currentX++
				}
			}
		}

		// Draw key in pill style (highlighted background) with padding
		// Left padding
		if currentX < x+width-len(m.rightText)-2 {
			screen.SetContent(currentX, y, ' ', nil, keyStyle)
			currentX++
		}

		// Key text
		for _, r := range hint.Key {
			if currentX < x+width-len(m.rightText)-2 {
				screen.SetContent(currentX, y, r, nil, keyStyle)
				currentX++
			}
		}

		// Right padding
		if currentX < x+width-len(m.rightText)-2 {
			screen.SetContent(currentX, y, ' ', nil, keyStyle)
			currentX++
		}

		// Space after pill
		if currentX < x+width-len(m.rightText)-2 {
			screen.SetContent(currentX, y, ' ', nil, bgStyle)
			currentX++
		}

		// Draw description
		for _, r := range hint.Description {
			if currentX < x+width-len(m.rightText)-2 {
				screen.SetContent(currentX, y, r, nil, descStyle)
				currentX++
			}
		}
	}

	// Draw right text (sponsor link, etc.)
	if m.rightText != "" {
		rightX := x + width - len(m.rightText) - 1
		if rightX > currentX {
			for _, r := range m.rightText {
				screen.SetContent(rightX, y, r, nil, rightStyle)
				rightX++
			}
		}
	}
}

// GetPreferredHeight returns the preferred height (1 row).
func (m *Menu) GetPreferredHeight() int {
	return 1
}

// AddHint adds a single hint.
func (m *Menu) AddHint(key, description string) *Menu {
	m.hints = append(m.hints, components.KeyHint{
		Key:         key,
		Description: description,
	})
	return m
}

// GetHints returns the current hints.
func (m *Menu) GetHints() []components.KeyHint {
	result := make([]components.KeyHint, len(m.hints))
	copy(result, m.hints)
	return result
}
