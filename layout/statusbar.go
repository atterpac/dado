package layout

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/jig/components"
	"github.com/atterpac/jig/theme"
)

// StatusSection represents a single section in the status bar.
type StatusSection struct {
	Icon      string      // Icon (Nerd Font glyph)
	Text      string      // Text to display
	Color     tcell.Color // Static color for icon (0 = use Fg, ignored if ColorFunc set)
	ColorFunc func() tcell.Color // Dynamic color function (takes precedence over Color)
}

// StatusBar is a configurable status bar with multiple sections.
type StatusBar struct {
	*components.Panel
	content       *tview.TextView
	sections      []StatusSection // Left-aligned sections
	rightSections []StatusSection // Right-aligned sections
	title         string
	contentAlign  components.Align

	// Command mode
	commandMode  bool
	commandInput *tview.InputField
	onSubmit     func(text string)
	onCancel     func()

	// Completion support
	completions      []string // Current completion suggestions
	completionIndex  int      // Currently selected completion (-1 = none)
	completionList   *tview.List
	showCompletions  bool
	onComplete       func(input string) []string // Callback to get completions
	onCompletionDone func()                      // Called when completion popup closes

	// History support
	onHistoryPrev func(current string) string
	onHistoryNext func(current string) string

	// Inline suggestion (ghost text)
	suggestion string

	subs components.Subscriptions
}

// Subs returns the status bar's subscription set; release on teardown.
func (s *StatusBar) Subs() *components.Subscriptions { return &s.subs }

// NewStatusBar creates a new status bar.
func NewStatusBar() *StatusBar {
	s := &StatusBar{
		Panel:           components.NewPanel(),
		content:         tview.NewTextView(),
		sections:        make([]StatusSection, 0),
		contentAlign:    components.AlignCenter, // Default to center
		completionIndex: -1,
	}

	s.content.SetDynamicColors(true)
	s.content.SetTextAlign(tview.AlignLeft)
	s.content.SetBackgroundColor(theme.Bg())
	s.Panel.SetContent(s.content)

	// Setup command input
	s.commandInput = tview.NewInputField()
	s.commandInput.SetBackgroundColor(theme.Bg())
	s.commandInput.SetFieldBackgroundColor(theme.Bg())
	s.commandInput.SetFieldTextColor(theme.Fg())
	s.commandInput.SetLabelColor(theme.Accent())
	s.commandInput.SetLabel(": ")
	s.commandInput.SetPlaceholder("command...")
	s.commandInput.SetPlaceholderTextColor(theme.FgMuted())

	// Setup completion list
	s.completionList = tview.NewList()
	s.completionList.SetBackgroundColor(theme.BgLight())
	s.completionList.SetMainTextColor(theme.Fg())
	s.completionList.SetSelectedBackgroundColor(theme.Accent())
	s.completionList.SetSelectedTextColor(theme.Bg())
	s.completionList.ShowSecondaryText(false)
	s.completionList.SetHighlightFullLine(true)

	// Handle input events with custom capture for Tab/arrows
	s.commandInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			// First, try to accept inline suggestion
			if s.acceptSuggestion() {
				return nil
			}
			// Otherwise trigger completion popup
			if s.onComplete != nil {
				input := s.commandInput.GetText()
				s.completions = s.onComplete(input)
				if len(s.completions) > 0 {
					s.showCompletions = true
					s.completionIndex = 0
					s.updateCompletionList()
				}
			}
			return nil

		case tcell.KeyRight:
			// Accept inline suggestion if cursor is at end
			currentText := s.commandInput.GetText()
			// tview InputField doesn't expose cursor position, so we check if suggestion is valid
			if s.suggestion != "" && strings.HasPrefix(strings.ToLower(s.suggestion), strings.ToLower(currentText)) {
				if s.acceptSuggestion() {
					return nil
				}
			}
			return event

		case tcell.KeyBacktab:
			// Navigate completions backwards
			if s.showCompletions && len(s.completions) > 0 {
				s.completionIndex--
				if s.completionIndex < 0 {
					s.completionIndex = len(s.completions) - 1
				}
				s.updateCompletionList()
			}
			return nil

		case tcell.KeyUp:
			if s.showCompletions && len(s.completions) > 0 {
				// Navigate completions up
				s.completionIndex--
				if s.completionIndex < 0 {
					s.completionIndex = len(s.completions) - 1
				}
				s.updateCompletionList()
				return nil
			}
			// History previous
			if s.onHistoryPrev != nil {
				current := s.commandInput.GetText()
				prev := s.onHistoryPrev(current)
				s.commandInput.SetText(prev)
			}
			return nil

		case tcell.KeyDown:
			if s.showCompletions && len(s.completions) > 0 {
				// Navigate completions down
				s.completionIndex++
				if s.completionIndex >= len(s.completions) {
					s.completionIndex = 0
				}
				s.updateCompletionList()
				return nil
			}
			// History next
			if s.onHistoryNext != nil {
				current := s.commandInput.GetText()
				next := s.onHistoryNext(current)
				s.commandInput.SetText(next)
			}
			return nil

		case tcell.KeyEnter:
			if s.showCompletions && s.completionIndex >= 0 && s.completionIndex < len(s.completions) {
				// Accept completion
				s.acceptCompletion()
				return nil
			}
			// Submit command
			if s.onSubmit != nil {
				s.onSubmit(s.commandInput.GetText())
			}
			return nil

		case tcell.KeyEscape:
			if s.showCompletions {
				// Close completions
				s.hideCompletions()
				return nil
			}
			// Cancel command mode
			if s.onCancel != nil {
				s.onCancel()
			}
			return nil
		}

		// Close completions on any other key that modifies input
		if s.showCompletions && event.Key() != tcell.KeyRune {
			s.hideCompletions()
		}

		return event
	})

	// Register content for automatic theme updates (Panel registers itself)
	s.subs.Add(theme.Register(s.content))
	s.subs.Add(theme.Register(s.commandInput))
	s.subs.Add(theme.Register(s.completionList))

	return s
}

// SetTitle sets the title shown in the border.
func (s *StatusBar) SetTitle(title string) *StatusBar {
	s.title = title
	s.Panel.SetTitle(title)
	return s
}

// SetTitleAlign sets the title alignment (Left, Center, or Right).
func (s *StatusBar) SetTitleAlign(align components.TitleAlign) *StatusBar {
	s.Panel.SetTitleAlign(align)
	return s
}

// SetContentAlign sets the content alignment (AlignLeft, AlignCenter, or AlignRight).
func (s *StatusBar) SetContentAlign(align components.Align) *StatusBar {
	s.contentAlign = align
	return s
}

// SetSections sets all status sections.
func (s *StatusBar) SetSections(sections []StatusSection) *StatusBar {
	s.sections = sections
	return s
}

// AddSection adds a status section.
func (s *StatusBar) AddSection(section StatusSection) *StatusBar {
	s.sections = append(s.sections, section)
	return s
}

// ClearSections removes all left-aligned sections.
func (s *StatusBar) ClearSections() *StatusBar {
	s.sections = make([]StatusSection, 0)
	return s
}

// AddRightSection adds a right-aligned status section.
func (s *StatusBar) AddRightSection(section StatusSection) *StatusBar {
	s.rightSections = append(s.rightSections, section)
	return s
}

// ClearRightSections removes all right-aligned sections.
func (s *StatusBar) ClearRightSections() *StatusBar {
	s.rightSections = make([]StatusSection, 0)
	return s
}

// SetRightSections sets all right-aligned status sections.
func (s *StatusBar) SetRightSections(sections []StatusSection) *StatusBar {
	s.rightSections = sections
	return s
}

// UpdateSection updates a specific section by index.
// Does nothing if index is out of range.
func (s *StatusBar) UpdateSection(index int, section StatusSection) *StatusBar {
	if index >= 0 && index < len(s.sections) {
		s.sections[index] = section
	}
	return s
}

// GetSection returns a section by index, or empty section if out of range.
func (s *StatusBar) GetSection(index int) StatusSection {
	if index >= 0 && index < len(s.sections) {
		return s.sections[index]
	}
	return StatusSection{}
}

// SectionCount returns the number of sections.
func (s *StatusBar) SectionCount() int {
	return len(s.sections)
}

// Draw renders the status bar with current theme colors.
func (s *StatusBar) Draw(screen tcell.Screen) {
	separator := "  [" + theme.TagFgMuted() + "]•[-]  "

	// Helper to build section text
	buildSectionText := func(section StatusSection) string {
		var sectionColor string
		if section.ColorFunc != nil {
			sectionColor = theme.ColorToHex(section.ColorFunc())
		} else if section.Color != 0 {
			sectionColor = theme.ColorToHex(section.Color)
		} else {
			sectionColor = theme.TagFg()
		}

		var part string
		if section.Icon != "" {
			part = "[" + sectionColor + "]" + section.Icon + "[-] "
		}
		part += "[" + sectionColor + "]" + section.Text + "[-]"
		return part
	}

	// Build left sections
	var leftParts []string
	for _, section := range s.sections {
		leftParts = append(leftParts, buildSectionText(section))
	}
	leftText := strings.Join(leftParts, separator)

	// Build right sections
	var rightParts []string
	for _, section := range s.rightSections {
		rightParts = append(rightParts, buildSectionText(section))
	}
	rightText := strings.Join(rightParts, separator)

	// Calculate padding between left and right
	_, _, width, _ := s.content.GetInnerRect()
	if width <= 0 {
		// Fallback if not yet laid out
		width = 100
	}

	// Calculate visible lengths (strip color tags for length calculation)
	leftLen := visibleLength(leftText)
	rightLen := visibleLength(rightText)

	// Build final text with padding
	var finalText string
	if len(s.rightSections) > 0 {
		padding := width - leftLen - rightLen
		if padding < 2 {
			padding = 2
		}
		finalText = leftText + strings.Repeat(" ", padding) + rightText
	} else {
		finalText = leftText
	}

	s.content.SetText(finalText)
	s.content.SetTextAlign(tview.AlignLeft)
	s.content.SetBackgroundColor(theme.Bg())

	s.Panel.Draw(screen)

	// Draw inline suggestion (ghost text) if in command mode
	if s.commandMode && s.suggestion != "" {
		s.drawSuggestion(screen)
	}

	// Draw completion popup if showing
	if s.showCompletions && len(s.completions) > 0 && s.commandMode {
		s.drawCompletionPopup(screen)
	}
}

// drawSuggestion draws the inline ghost text after the current input.
func (s *StatusBar) drawSuggestion(screen tcell.Screen) {
	currentText := s.commandInput.GetText()
	if currentText == "" {
		return
	}

	// Check if suggestion starts with current input (case-insensitive)
	if !strings.HasPrefix(strings.ToLower(s.suggestion), strings.ToLower(currentText)) {
		return
	}

	// Get the suffix to display (the part after what user typed)
	suffix := s.suggestion[len(currentText):]
	if suffix == "" {
		return
	}

	// Get the input field's actual position on screen
	inputX, inputY, _, _ := s.commandInput.GetRect()

	// Calculate x position: input field x + label length + input text length
	label := s.commandInput.GetLabel()
	labelLen := len([]rune(label))
	inputLen := len([]rune(currentText))
	startX := inputX + labelLen + inputLen

	// Draw the ghost text in muted color
	ghostStyle := tcell.StyleDefault.Background(theme.Bg()).Foreground(theme.FgMuted())

	col := startX
	for _, r := range suffix {
		screen.SetContent(col, inputY, r, nil, ghostStyle)
		col++
	}
}

// drawCompletionPopup draws the completion suggestions below the status bar.
func (s *StatusBar) drawCompletionPopup(screen tcell.Screen) {
	x, y, width, _ := s.GetRect()

	// Calculate popup dimensions
	maxItems := 8
	items := len(s.completions)
	if items > maxItems {
		items = maxItems
	}

	popupHeight := items + 2 // +2 for border
	popupWidth := 40
	if popupWidth > width-4 {
		popupWidth = width - 4
	}

	// Position popup below the status bar, aligned to left with some padding
	popupX := x + 2
	popupY := y + 3 // Below status bar

	// Draw popup background and border
	bgStyle := tcell.StyleDefault.Background(theme.BgLight()).Foreground(theme.Fg())
	borderStyle := tcell.StyleDefault.Background(theme.BgLight()).Foreground(theme.FgMuted())

	// Fill background
	for row := popupY; row < popupY+popupHeight; row++ {
		for col := popupX; col < popupX+popupWidth; col++ {
			screen.SetContent(col, row, ' ', nil, bgStyle)
		}
	}

	// Draw border
	// Top border
	screen.SetContent(popupX, popupY, '┌', nil, borderStyle)
	for col := popupX + 1; col < popupX+popupWidth-1; col++ {
		screen.SetContent(col, popupY, '─', nil, borderStyle)
	}
	screen.SetContent(popupX+popupWidth-1, popupY, '┐', nil, borderStyle)

	// Bottom border
	screen.SetContent(popupX, popupY+popupHeight-1, '└', nil, borderStyle)
	for col := popupX + 1; col < popupX+popupWidth-1; col++ {
		screen.SetContent(col, popupY+popupHeight-1, '─', nil, borderStyle)
	}
	screen.SetContent(popupX+popupWidth-1, popupY+popupHeight-1, '┘', nil, borderStyle)

	// Side borders
	for row := popupY + 1; row < popupY+popupHeight-1; row++ {
		screen.SetContent(popupX, row, '│', nil, borderStyle)
		screen.SetContent(popupX+popupWidth-1, row, '│', nil, borderStyle)
	}

	// Draw completion items
	for i := 0; i < items; i++ {
		row := popupY + 1 + i
		text := s.completions[i]

		// Truncate if too long
		maxLen := popupWidth - 4
		if len(text) > maxLen {
			text = text[:maxLen-1] + "…"
		}

		// Determine style (highlight selected)
		style := bgStyle
		if i == s.completionIndex {
			style = tcell.StyleDefault.Background(theme.Accent()).Foreground(theme.Bg())
		}

		// Fill row background for selected item
		if i == s.completionIndex {
			for col := popupX + 1; col < popupX+popupWidth-1; col++ {
				screen.SetContent(col, row, ' ', nil, style)
			}
		}

		// Draw text
		col := popupX + 2
		for _, r := range text {
			if col >= popupX+popupWidth-2 {
				break
			}
			screen.SetContent(col, row, r, nil, style)
			col++
		}
	}

	// Show scroll indicator if more items
	if len(s.completions) > maxItems {
		indicator := fmt.Sprintf("↓ %d more", len(s.completions)-maxItems)
		col := popupX + popupWidth - 2 - len(indicator)
		for _, r := range indicator {
			screen.SetContent(col, popupY+popupHeight-1, r, nil, borderStyle)
			col++
		}
	}
}

// GetPreferredHeight returns the preferred height (3 rows for padding).
func (s *StatusBar) GetPreferredHeight() int {
	return 3
}

// SetConnectionStatus is a convenience method for showing connection state.
func (s *StatusBar) SetConnectionStatus(connected bool, name string) *StatusBar {
	icon := theme.IconDisconnected
	colorFunc := theme.Error // Dynamic color function
	if connected {
		icon = theme.IconConnected
		colorFunc = theme.Success // Dynamic color function
	}

	// Update or add connection section at index 0
	section := StatusSection{
		Icon:      icon,
		Text:      name,
		ColorFunc: colorFunc, // Use dynamic color that updates with theme
	}

	if len(s.sections) > 0 {
		s.sections[0] = section
	} else {
		s.sections = append(s.sections, section)
	}

	return s
}

// -----------------------------------------------------------------------------
// Command Mode
// -----------------------------------------------------------------------------

// EnterCommandMode switches the status bar to show a command input.
func (s *StatusBar) EnterCommandMode() *StatusBar {
	s.commandMode = true
	s.commandInput.SetText("")
	s.Panel.SetContent(s.commandInput)
	return s
}

// ExitCommandMode switches back to showing status sections.
func (s *StatusBar) ExitCommandMode() *StatusBar {
	s.commandMode = false
	s.commandInput.SetText("")
	s.Panel.SetContent(s.content)
	return s
}

// IsCommandMode returns whether command mode is active.
func (s *StatusBar) IsCommandMode() bool {
	return s.commandMode
}

// SetCommandPrompt sets the prompt shown before the input (default ": ").
func (s *StatusBar) SetCommandPrompt(prompt string) *StatusBar {
	s.commandInput.SetLabel(prompt)
	return s
}

// SetCommandPlaceholder sets the placeholder text.
func (s *StatusBar) SetCommandPlaceholder(placeholder string) *StatusBar {
	s.commandInput.SetPlaceholder(placeholder)
	return s
}

// SetOnCommandSubmit sets the callback when Enter is pressed in command mode.
func (s *StatusBar) SetOnCommandSubmit(fn func(text string)) *StatusBar {
	s.onSubmit = fn
	return s
}

// SetOnCommandCancel sets the callback when Escape is pressed in command mode.
func (s *StatusBar) SetOnCommandCancel(fn func()) *StatusBar {
	s.onCancel = fn
	return s
}

// GetCommandInput returns the input field for focusing.
func (s *StatusBar) GetCommandInput() *tview.InputField {
	return s.commandInput
}

// SetOnComplete sets the callback to get completions for current input.
// The callback receives the current input and should return a list of completion strings.
func (s *StatusBar) SetOnComplete(fn func(input string) []string) *StatusBar {
	s.onComplete = fn
	return s
}

// SetOnHistoryPrev sets the callback for getting the previous history entry.
func (s *StatusBar) SetOnHistoryPrev(fn func(current string) string) *StatusBar {
	s.onHistoryPrev = fn
	return s
}

// SetOnHistoryNext sets the callback for getting the next history entry.
func (s *StatusBar) SetOnHistoryNext(fn func(current string) string) *StatusBar {
	s.onHistoryNext = fn
	return s
}

// SetSuggestion sets the inline suggestion (ghost text) that appears after the input.
// The suggestion should be the full text, including what the user has already typed.
// Only the portion after the current input will be shown as ghost text.
func (s *StatusBar) SetSuggestion(suggestion string) *StatusBar {
	s.suggestion = suggestion
	return s
}

// GetSuggestion returns the current inline suggestion.
func (s *StatusBar) GetSuggestion() string {
	return s.suggestion
}

// ClearSuggestion clears the inline suggestion.
func (s *StatusBar) ClearSuggestion() *StatusBar {
	s.suggestion = ""
	return s
}

// acceptSuggestion accepts the current inline suggestion.
// Returns true if a suggestion was accepted, false otherwise.
func (s *StatusBar) acceptSuggestion() bool {
	if s.suggestion == "" {
		return false
	}

	currentText := s.commandInput.GetText()
	if currentText == "" {
		return false
	}

	// Check if suggestion starts with current input (case-insensitive)
	if !strings.HasPrefix(strings.ToLower(s.suggestion), strings.ToLower(currentText)) {
		return false
	}

	// Accept the suggestion
	s.commandInput.SetText(s.suggestion)
	s.suggestion = ""
	return true
}

// GetCompletionList returns the completion list for rendering.
func (s *StatusBar) GetCompletionList() *tview.List {
	return s.completionList
}

// IsShowingCompletions returns whether completions are currently displayed.
func (s *StatusBar) IsShowingCompletions() bool {
	return s.showCompletions
}

// GetCompletions returns the current completion suggestions.
func (s *StatusBar) GetCompletions() []string {
	return s.completions
}

// GetCompletionIndex returns the currently selected completion index.
func (s *StatusBar) GetCompletionIndex() int {
	return s.completionIndex
}

// updateCompletionList updates the list widget with current completions.
func (s *StatusBar) updateCompletionList() {
	s.completionList.Clear()
	for _, c := range s.completions {
		s.completionList.AddItem(c, "", 0, nil)
	}
	if s.completionIndex >= 0 && s.completionIndex < len(s.completions) {
		s.completionList.SetCurrentItem(s.completionIndex)
	}
}

// acceptCompletion inserts the selected completion into the input.
func (s *StatusBar) acceptCompletion() {
	if s.completionIndex < 0 || s.completionIndex >= len(s.completions) {
		return
	}

	completion := s.completions[s.completionIndex]
	currentText := s.commandInput.GetText()

	// Find the last word to replace
	parts := strings.Fields(currentText)
	if len(parts) == 0 {
		s.commandInput.SetText(completion + " ")
	} else {
		// Replace last partial word with completion
		parts[len(parts)-1] = completion
		s.commandInput.SetText(strings.Join(parts, " ") + " ")
	}

	s.hideCompletions()
}

// hideCompletions closes the completion popup.
func (s *StatusBar) hideCompletions() {
	s.showCompletions = false
	s.completionIndex = -1
	s.completions = nil
	s.completionList.Clear()
	if s.onCompletionDone != nil {
		s.onCompletionDone()
	}
}

// visibleLength calculates the visible length of a string with tview color tags.
// It strips [color]text[-] style tags to get the actual display length.
func visibleLength(s string) int {
	length := 0
	inTag := false
	for _, r := range s {
		if r == '[' {
			inTag = true
			continue
		}
		if r == ']' && inTag {
			inTag = false
			continue
		}
		if !inTag {
			length++
		}
	}
	return length
}
