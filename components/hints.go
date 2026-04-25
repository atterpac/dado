package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/jig/theme"
)

// KeyHint represents a single key binding hint.
type KeyHint struct {
	Key         string // e.g., "Enter", "Esc", "Space", "j/k"
	Description string // e.g., "Select", "Close", "Toggle"
}

// KeyHintBar displays key hints in a pill style at the bottom of views/modals.
type KeyHintBar struct {
	*tview.Box
	Hints []KeyHint
	subs  Subscriptions
}

// Subs returns the widget's subscription set; released by ComponentBase.Stop.
func (k *KeyHintBar) Subs() *Subscriptions { return &k.subs }

// NewKeyHintBar creates a new key hint bar.
func NewKeyHintBar() *KeyHintBar {
	box := tview.NewBox()
	box.SetBackgroundColor(theme.Bg())

	k := &KeyHintBar{
		Box:   box,
		Hints: make([]KeyHint, 0),
	}

	k.subs.Add(theme.Register(box))

	return k
}

// SetHints sets the hints to display.
func (k *KeyHintBar) SetHints(hints []KeyHint) *KeyHintBar {
	k.Hints = hints
	return k
}

// AddHint adds a single hint.
func (k *KeyHintBar) AddHint(key, description string) *KeyHintBar {
	k.Hints = append(k.Hints, KeyHint{Key: key, Description: description})
	return k
}

// Clear removes all hints.
func (k *KeyHintBar) Clear() *KeyHintBar {
	k.Hints = make([]KeyHint, 0)
	return k
}

// Draw renders the key hint bar.
func (k *KeyHintBar) Draw(screen tcell.Screen) {
	k.Box.DrawForSubclass(screen, k)

	x, y, width, height := k.GetInnerRect()
	if width < 1 || height < 1 || len(k.Hints) == 0 {
		return
	}

	// Get colors from theme at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	accentColor := theme.Accent()

	// Build the hint string and calculate positions
	var parts []hintPart
	totalWidth := 0

	for i, hint := range k.Hints {
		if i > 0 {
			// Add separator
			parts = append(parts, hintPart{text: "   ", isKey: false})
			totalWidth += 3
		}

		// Key part: [Key]
		keyText := "[" + hint.Key + "]"
		parts = append(parts, hintPart{text: keyText, isKey: true})
		totalWidth += len(keyText)

		// Space
		parts = append(parts, hintPart{text: " ", isKey: false})
		totalWidth += 1

		// Description
		parts = append(parts, hintPart{text: hint.Description, isKey: false})
		totalWidth += len(hint.Description)
	}

	// Center horizontally
	startX := x + (width-totalWidth)/2
	if startX < x {
		startX = x
	}

	// Draw each part
	currentX := startX
	drawY := y + height/2 // Center vertically if height > 1

	bgStyle := tcell.StyleDefault.Background(bgColor)
	keyStyle := tcell.StyleDefault.Background(accentColor).Foreground(bgColor)
	descStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)

	// Clear the line first
	for i := x; i < x+width; i++ {
		screen.SetContent(i, drawY, ' ', nil, bgStyle)
	}

	// Draw hints
	for _, part := range parts {
		style := descStyle
		if part.isKey {
			style = keyStyle
		}

		for _, r := range part.text {
			if currentX < x+width {
				screen.SetContent(currentX, drawY, r, nil, style)
				currentX++
			}
		}
	}
}

// hintPart is a segment of the hint bar (either key or description).
type hintPart struct {
	text  string
	isKey bool
}

// GetPreferredHeight returns the preferred height for the hint bar.
func (k *KeyHintBar) GetPreferredHeight() int {
	return 1
}

// HintsToString converts hints to a simple string representation.
func HintsToString(hints []KeyHint) string {
	var parts []string
	for _, h := range hints {
		parts = append(parts, "["+h.Key+"] "+h.Description)
	}
	return strings.Join(parts, "   ")
}
