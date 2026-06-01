package layout

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// StatusSection represents a single section in the status bar.
type StatusSection struct {
	Icon      string             // Icon (Nerd Font glyph)
	Text      string             // Text to display
	Color     tcell.Color        // Static color for icon (0 = use Fg, ignored if ColorFunc set)
	ColorFunc func() tcell.Color // Dynamic color function (takes precedence over Color)
}

// StatusBar is a configurable status bar with multiple sections.
type StatusBar struct {
	*components.Panel
	content       *core.TextView
	sections      []StatusSection // Left-aligned sections
	rightSections []StatusSection // Right-aligned sections
	title         string
	contentAlign  components.Align

	// Command mode
	commandMode   bool
	commandInput  *components.TextField
	commandPrompt string // stored label for drawSuggestion
	onSubmit      func(text string)
	onCancel      func()
	onChange      func(text string)

	// Completion support
	completions      []string // Current completion suggestions
	completionIndex  int      // Currently selected completion (-1 = none)
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
		content:         core.NewTextView(),
		sections:        make([]StatusSection, 0),
		contentAlign:    components.AlignCenter, // Default to center
		completionIndex: -1,
		commandPrompt:   ": ",
	}

	s.content.SetDynamicColors(true)
	s.content.SetTextAlign(core.AlignLeft)
	s.content.SetBackgroundColor(theme.Bg())
	s.Panel.SetContent(s.content)

	// Setup command input
	s.commandInput = components.NewTextField("command").
		SetLabel(": ").
		SetPlaceholder("command...")
	s.commandInput.SetOnChange(func(ev *components.ChangeEvent[string]) {
		if s.onChange != nil {
			s.onChange(ev.NewValue)
		}
	})

	s.subs.Add(theme.RegisterFn(func(c tcell.Color) { s.content.SetBackgroundColor(c) }))
	s.subs.Add(theme.RegisterFn(func(c tcell.Color) { s.commandInput.SetBackgroundColor(c) }))

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
	// In command mode the bar "transforms" into a command line: the same
	// content area is reused to show prompt + input text inline (no nested
	// input box), so only the content swaps, not the frame.
	if s.commandMode {
		s.drawCommandMode(screen)
		return
	}

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
	s.content.SetTextAlign(core.AlignLeft)
	s.content.SetBackgroundColor(theme.Bg())

	s.Panel.Draw(screen)
}

// drawCommandMode renders the command line inline within the existing content
// area: prompt + input text + a block cursor, reusing the same TextView slot so
// the bar visually transforms rather than nesting a bordered input box.
func (s *StatusBar) drawCommandMode(screen tcell.Screen) {
	value := s.commandInput.GetValue()

	s.content.SetText(s.commandPrompt + value)
	s.content.SetTextAlign(core.AlignLeft)
	s.content.SetBackgroundColor(theme.Bg())

	s.Panel.Draw(screen)

	// Draw the text cursor as a reverse-video cell over the content row.
	cx, cy, cw, _ := s.content.GetRect()
	promptLen := len([]rune(s.commandPrompt))
	cursorPos := s.commandInput.CursorPos()
	cursorX := cx + promptLen + cursorPos

	if cursorX >= cx && cursorX < cx+cw {
		runes := []rune(value)
		cursorCh := ' '
		if cursorPos >= 0 && cursorPos < len(runes) {
			cursorCh = runes[cursorPos]
		}
		cursorStyle := tcell.StyleDefault.Background(theme.Fg()).Foreground(theme.Bg())
		screen.SetContent(cursorX, cy, cursorCh, nil, cursorStyle)
	}

	// Ghost-text suggestion and completion popup.
	if s.suggestion != "" {
		s.drawSuggestion(screen)
	}
	if s.showCompletions && len(s.completions) > 0 {
		s.drawCompletionPopup(screen)
	}
}

// drawSuggestion draws the inline ghost text after the current input.
func (s *StatusBar) drawSuggestion(screen tcell.Screen) {
	currentText := s.commandInput.GetValue()
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

	// Position relative to the inline content row: content x + prompt + typed.
	cx, cy, cw, _ := s.content.GetRect()
	promptLen := len([]rune(s.commandPrompt))
	inputLen := len([]rune(currentText))
	startX := cx + promptLen + inputLen

	// Draw the ghost text in muted color, clipped to the content width.
	ghostStyle := tcell.StyleDefault.Background(theme.Bg()).Foreground(theme.FgMuted())

	col := startX
	for _, r := range suffix {
		if col >= cx+cw {
			break
		}
		screen.SetContent(col, cy, r, nil, ghostStyle)
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

// EnterCommandMode switches the status bar to show a command input. The command
// line is rendered inline in the existing content area (see drawCommandMode);
// the Panel content stays s.content so only the text swaps, not the frame.
func (s *StatusBar) EnterCommandMode() *StatusBar {
	s.commandMode = true
	s.commandInput.SetValue("")
	return s
}

// ExitCommandMode switches back to showing status sections.
func (s *StatusBar) ExitCommandMode() *StatusBar {
	s.commandMode = false
	s.commandInput.SetValue("")
	return s
}

// IsCommandMode returns whether command mode is active.
func (s *StatusBar) IsCommandMode() bool {
	return s.commandMode
}

// SetCommandPrompt sets the prompt shown before the input (default ": ").
func (s *StatusBar) SetCommandPrompt(prompt string) *StatusBar {
	s.commandPrompt = prompt
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

// SetOnCommandChange sets the callback invoked whenever the command input text
// changes (live). Pass nil to clear it.
func (s *StatusBar) SetOnCommandChange(fn func(text string)) *StatusBar {
	s.onChange = fn
	return s
}

// SetCommandText sets the current command input text and moves the cursor to
// the end. Useful for seeding a search box with an existing query.
func (s *StatusBar) SetCommandText(text string) *StatusBar {
	s.commandInput.SetValue(text)
	return s
}

// CommandText returns the current command input text.
func (s *StatusBar) CommandText() string {
	return s.commandInput.GetValue()
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

	currentText := s.commandInput.GetValue()
	if currentText == "" {
		return false
	}

	// Check if suggestion starts with current input (case-insensitive)
	if !strings.HasPrefix(strings.ToLower(s.suggestion), strings.ToLower(currentText)) {
		return false
	}

	// Accept the suggestion
	s.commandInput.SetValue(s.suggestion)
	s.suggestion = ""
	return true
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

// HandleKey handles keyboard input for command mode.
func (s *StatusBar) HandleKey(ev *tcell.EventKey) bool {
	if !s.commandMode {
		return false
	}

	switch ev.Key() {
	case tcell.KeyTab:
		// First, try to accept inline suggestion
		if s.acceptSuggestion() {
			return true
		}
		// Otherwise trigger completion popup
		if s.onComplete != nil {
			input := s.commandInput.GetValue()
			s.completions = s.onComplete(input)
			if len(s.completions) > 0 {
				s.showCompletions = true
				s.completionIndex = 0
			}
		}
		return true

	case tcell.KeyRight:
		// Accept inline suggestion if one is pending
		currentText := s.commandInput.GetValue()
		if s.suggestion != "" && strings.HasPrefix(strings.ToLower(s.suggestion), strings.ToLower(currentText)) {
			if s.acceptSuggestion() {
				return true
			}
		}
		return s.commandInput.HandleKey(ev)

	case tcell.KeyBacktab:
		// Navigate completions backwards
		if s.showCompletions && len(s.completions) > 0 {
			s.completionIndex--
			if s.completionIndex < 0 {
				s.completionIndex = len(s.completions) - 1
			}
		}
		return true

	case tcell.KeyUp:
		if s.showCompletions && len(s.completions) > 0 {
			s.completionIndex--
			if s.completionIndex < 0 {
				s.completionIndex = len(s.completions) - 1
			}
			return true
		}
		// History previous
		if s.onHistoryPrev != nil {
			current := s.commandInput.GetValue()
			prev := s.onHistoryPrev(current)
			s.commandInput.SetValue(prev)
		}
		return true

	case tcell.KeyDown:
		if s.showCompletions && len(s.completions) > 0 {
			s.completionIndex++
			if s.completionIndex >= len(s.completions) {
				s.completionIndex = 0
			}
			return true
		}
		// History next
		if s.onHistoryNext != nil {
			current := s.commandInput.GetValue()
			next := s.onHistoryNext(current)
			s.commandInput.SetValue(next)
		}
		return true

	case tcell.KeyEnter:
		if s.showCompletions && s.completionIndex >= 0 && s.completionIndex < len(s.completions) {
			s.acceptCompletion()
			return true
		}
		// Submit command
		if s.onSubmit != nil {
			s.onSubmit(s.commandInput.GetValue())
		}
		return true

	case tcell.KeyEscape:
		if s.showCompletions {
			s.hideCompletions()
			return true
		}
		// Cancel command mode
		if s.onCancel != nil {
			s.onCancel()
		}
		return true
	}

	// Close completions on any non-rune key that modifies input
	if s.showCompletions && ev.Key() != tcell.KeyRune {
		s.hideCompletions()
	}

	return s.commandInput.HandleKey(ev)
}

// acceptCompletion inserts the selected completion into the input.
func (s *StatusBar) acceptCompletion() {
	if s.completionIndex < 0 || s.completionIndex >= len(s.completions) {
		return
	}

	completion := s.completions[s.completionIndex]
	currentText := s.commandInput.GetValue()

	// Find the last word to replace
	parts := strings.Fields(currentText)
	if len(parts) == 0 {
		s.commandInput.SetValue(completion + " ")
	} else {
		// Replace last partial word with completion
		parts[len(parts)-1] = completion
		s.commandInput.SetValue(strings.Join(parts, " ") + " ")
	}

	s.hideCompletions()
}

// hideCompletions closes the completion popup.
func (s *StatusBar) hideCompletions() {
	s.showCompletions = false
	s.completionIndex = -1
	s.completions = nil
	if s.onCompletionDone != nil {
		s.onCompletionDone()
	}
}

// visibleLength calculates the visible length of a string with color tags.
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
