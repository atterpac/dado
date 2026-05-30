package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// KeyHint represents a single key binding hint.
type KeyHint struct {
	Key         string // e.g., "Enter", "Esc", "Space", "j/k"
	Description string // e.g., "Select", "Close", "Toggle"
}

// KeyHintBar displays key hints in a pill style at the bottom of views/modals.
type KeyHintBar struct {
	widgetBase
	Hints []KeyHint
}

// NewKeyHintBar creates a new key hint bar.
func NewKeyHintBar() *KeyHintBar {
	k := &KeyHintBar{
		Hints: make([]KeyHint, 0),
	}

	k.initWidget(tview.NewBox())

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
	th := k.th()
	bgColor := th.Bg()
	fgColor := th.Fg()
	accentColor := th.Accent()

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
	fillLine(screen, x, drawY, width, bgStyle)

	// Draw hints
	for _, part := range parts {
		style := descStyle
		if part.isKey {
			style = keyStyle
		}

		currentX = drawText(screen, currentX, drawY, x+width-currentX, part.text, style)
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
