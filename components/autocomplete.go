package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// Suggestion represents an autocomplete suggestion.
type Suggestion struct {
	Text        string // Display text
	InsertText  string // Text to insert when selected (defaults to Text if empty)
	Description string // Optional description
	Category    string // Category for grouping (e.g., "Field", "Operator", "Value")
	Data        any    // Optional user data
}

// SuggestionProvider is a function that returns suggestions based on input and cursor position.
type SuggestionProvider func(text string, cursorPos int) []Suggestion

// HistoryProvider is a function that returns history entries.
// Direction: -1 for previous, +1 for next.
// Returns the history entry string, or empty string if none.
type HistoryProvider func(direction int) string

// AutocompleteInput is an input field with autocomplete support.
type AutocompleteInput struct {
	*tview.Box
	text           string
	cursorPos      int
	filteredSuggs  []Suggestion
	selectedIndex  int
	showSuggestion bool
	maxSuggestions int

	// Configuration
	prompt      string
	placeholder string
	title       string

	// Callbacks
	onSubmit func(text string)
	onCancel func()
	onChange func(text string)
	onSelect func(suggestion Suggestion)

	// Providers
	suggestionFn SuggestionProvider
	historyFn    HistoryProvider
}

// NewAutocompleteInput creates a new autocomplete input field.
func NewAutocompleteInput() *AutocompleteInput {
	ai := &AutocompleteInput{
		Box:            tview.NewBox(),
		maxSuggestions: 8,
		prompt:         "> ",
		placeholder:    "Type to search...",
	}
	ai.SetBackgroundColor(theme.Bg())

	// Register for automatic theme updates
	theme.Register(ai.Box)

	return ai
}

// SetPrompt sets the prompt displayed before the input.
func (ai *AutocompleteInput) SetPrompt(prompt string) *AutocompleteInput {
	ai.prompt = prompt
	return ai
}

// SetPlaceholder sets the placeholder text shown when input is empty.
func (ai *AutocompleteInput) SetPlaceholder(placeholder string) *AutocompleteInput {
	ai.placeholder = placeholder
	return ai
}

// SetTitle sets the title displayed in the border.
func (ai *AutocompleteInput) SetTitle(title string) *AutocompleteInput {
	ai.title = title
	return ai
}

// SetMaxSuggestions sets the maximum number of suggestions to display.
func (ai *AutocompleteInput) SetMaxSuggestions(max int) *AutocompleteInput {
	ai.maxSuggestions = max
	return ai
}

// SetText sets the input text.
func (ai *AutocompleteInput) SetText(text string) *AutocompleteInput {
	ai.text = text
	ai.cursorPos = len(text)
	ai.updateSuggestions()
	if ai.onChange != nil {
		ai.onChange(text)
	}
	return ai
}

// GetText returns the current input text.
func (ai *AutocompleteInput) GetText() string {
	return ai.text
}

// Clear clears the input text.
func (ai *AutocompleteInput) Clear() *AutocompleteInput {
	ai.text = ""
	ai.cursorPos = 0
	ai.showSuggestion = false
	ai.filteredSuggs = nil
	if ai.onChange != nil {
		ai.onChange("")
	}
	return ai
}

// SetOnSubmit sets the submit callback (Enter pressed).
func (ai *AutocompleteInput) SetOnSubmit(fn func(text string)) *AutocompleteInput {
	ai.onSubmit = fn
	return ai
}

// SetOnCancel sets the cancel callback (Esc pressed).
func (ai *AutocompleteInput) SetOnCancel(fn func()) *AutocompleteInput {
	ai.onCancel = fn
	return ai
}

// SetOnChange sets the change callback (input changed).
func (ai *AutocompleteInput) SetOnChange(fn func(text string)) *AutocompleteInput {
	ai.onChange = fn
	return ai
}

// SetOnSelect sets the selection callback for when a suggestion is chosen.
func (ai *AutocompleteInput) SetOnSelect(fn func(suggestion Suggestion)) *AutocompleteInput {
	ai.onSelect = fn
	return ai
}

// SetSuggestionProvider sets a function to provide suggestions.
func (ai *AutocompleteInput) SetSuggestionProvider(fn SuggestionProvider) *AutocompleteInput {
	ai.suggestionFn = fn
	return ai
}

// SetHistoryProvider sets a function to provide history entries.
// The function receives direction (-1 for previous, +1 for next) and returns the entry.
func (ai *AutocompleteInput) SetHistoryProvider(fn HistoryProvider) *AutocompleteInput {
	ai.historyFn = fn
	return ai
}

// IsSuggestionsVisible returns whether suggestions are currently displayed.
func (ai *AutocompleteInput) IsSuggestionsVisible() bool {
	return ai.showSuggestion && len(ai.filteredSuggs) > 0
}

// GetSuggestions returns the current filtered suggestions.
func (ai *AutocompleteInput) GetSuggestions() []Suggestion {
	return ai.filteredSuggs
}

// GetSelectedSuggestionIndex returns the currently selected suggestion index.
func (ai *AutocompleteInput) GetSelectedSuggestionIndex() int {
	return ai.selectedIndex
}

// updateSuggestions updates the filtered suggestions based on current input.
func (ai *AutocompleteInput) updateSuggestions() {
	if ai.suggestionFn != nil {
		ai.filteredSuggs = ai.suggestionFn(ai.text, ai.cursorPos)
	} else {
		ai.filteredSuggs = nil
	}
	ai.selectedIndex = 0
	ai.showSuggestion = len(ai.filteredSuggs) > 0
}

// acceptSuggestion inserts the selected suggestion.
func (ai *AutocompleteInput) acceptSuggestion() {
	if ai.selectedIndex < 0 || ai.selectedIndex >= len(ai.filteredSuggs) {
		return
	}

	sugg := ai.filteredSuggs[ai.selectedIndex]
	insertText := sugg.InsertText
	if insertText == "" {
		insertText = sugg.Text
	}

	// Find where to insert (replace current token)
	textUpToCursor := ai.text
	if ai.cursorPos < len(ai.text) {
		textUpToCursor = ai.text[:ai.cursorPos]
	}

	lastSpace := strings.LastIndexAny(textUpToCursor, " ()")
	prefix := ""
	if lastSpace >= 0 {
		prefix = textUpToCursor[:lastSpace+1]
	}

	suffix := ""
	if ai.cursorPos < len(ai.text) {
		suffix = ai.text[ai.cursorPos:]
	}

	// Build new text
	ai.text = prefix + insertText + suffix
	ai.cursorPos = len(prefix) + len(insertText)

	ai.showSuggestion = false
	ai.filteredSuggs = nil

	if ai.onSelect != nil {
		ai.onSelect(sugg)
	}
	if ai.onChange != nil {
		ai.onChange(ai.text)
	}
}

// Draw renders the autocomplete input.
func (ai *AutocompleteInput) Draw(screen tcell.Screen) {
	ai.Box.DrawForSubclass(screen, ai)

	x, y, width, height := ai.GetInnerRect()
	if width <= 0 || height < 1 {
		return
	}

	// Colors
	borderStyle := tcell.StyleDefault.Foreground(theme.PanelBorder()).Background(theme.Bg())
	titleStyle := tcell.StyleDefault.Foreground(theme.Accent()).Background(theme.Bg()).Bold(true)
	textStyle := tcell.StyleDefault.Foreground(theme.Fg()).Background(theme.Bg())
	promptStyle := tcell.StyleDefault.Foreground(theme.Accent()).Background(theme.Bg())
	placeholderStyle := tcell.StyleDefault.Foreground(theme.FgDim()).Background(theme.Bg())
	hintStyle := tcell.StyleDefault.Foreground(theme.FgDim()).Background(theme.Bg())
	suggestionStyle := tcell.StyleDefault.Foreground(theme.Fg()).Background(theme.BgLight())
	selectedStyle := tcell.StyleDefault.Foreground(theme.Bg()).Background(theme.Accent())
	categoryStyle := tcell.StyleDefault.Foreground(theme.FgDim()).Background(theme.BgLight())

	// Calculate content area (inside border)
	inputRows := 3

	// Draw input box border
	screen.SetContent(x, y, []rune(theme.IconCornerTL)[0], nil, borderStyle)
	screen.SetContent(x+width-1, y, []rune(theme.IconCornerTR)[0], nil, borderStyle)
	screen.SetContent(x, y+2, []rune(theme.IconCornerBL)[0], nil, borderStyle)
	screen.SetContent(x+width-1, y+2, []rune(theme.IconCornerBR)[0], nil, borderStyle)

	for i := x + 1; i < x+width-1; i++ {
		screen.SetContent(i, y, '─', nil, borderStyle)
		screen.SetContent(i, y+2, '─', nil, borderStyle)
	}
	screen.SetContent(x, y+1, '│', nil, borderStyle)
	screen.SetContent(x+width-1, y+1, '│', nil, borderStyle)

	// Draw title if set
	if ai.title != "" {
		title := " " + ai.title + " "
		titleRunes := []rune(title)
		titleX := x + 2
		for i, r := range titleRunes {
			if titleX+i >= x+width-1 {
				break
			}
			screen.SetContent(titleX+i, y, r, nil, titleStyle)
		}
	}

	// Draw prompt and input
	contentY := y + 1
	contentX := x + 2

	for _, r := range []rune(ai.prompt) {
		if contentX >= x+width-2 {
			break
		}
		screen.SetContent(contentX, contentY, r, nil, promptStyle)
		contentX++
	}

	// Draw input text or placeholder
	if ai.text == "" && ai.placeholder != "" {
		for i, r := range []rune(ai.placeholder) {
			if contentX+i >= x+width-2 {
				break
			}
			screen.SetContent(contentX+i, contentY, r, nil, placeholderStyle)
		}
	} else {
		inputRunes := []rune(ai.text)
		for i, r := range inputRunes {
			if contentX+i >= x+width-2 {
				break
			}
			screen.SetContent(contentX+i, contentY, r, nil, textStyle)
		}
	}

	// Draw cursor
	cursorX := contentX + ai.cursorPos
	if cursorX < x+width-2 {
		cursorStyle := tcell.StyleDefault.Foreground(theme.Bg()).Background(theme.Fg())
		if ai.cursorPos < len(ai.text) {
			r := []rune(ai.text)[ai.cursorPos]
			screen.SetContent(cursorX, contentY, r, nil, cursorStyle)
		} else {
			screen.SetContent(cursorX, contentY, ' ', nil, cursorStyle)
		}
	}

	// Draw hint on right side
	hint := "[Tab] Complete  [" + theme.IconArrowUp + "/" + theme.IconArrowDown + "] Navigate"
	hintX := x + width - len([]rune(hint)) - 3
	if hintX > contentX+len(ai.text)+2 {
		for i, r := range []rune(hint) {
			screen.SetContent(hintX+i, contentY, r, nil, hintStyle)
		}
	}

	// Draw suggestions dropdown below
	if ai.showSuggestion && len(ai.filteredSuggs) > 0 && height > inputRows {
		suggY := y + 3
		suggWidth := width - 4
		if suggWidth < 30 {
			suggWidth = min(width-2, 40)
		}
		suggX := x + 2

		// Draw suggestion box
		numSuggs := min(len(ai.filteredSuggs), ai.maxSuggestions)
		if suggY+numSuggs+1 <= y+height {
			// Top border
			screen.SetContent(suggX, suggY, '┌', nil, borderStyle)
			screen.SetContent(suggX+suggWidth-1, suggY, '┐', nil, borderStyle)
			for i := suggX + 1; i < suggX+suggWidth-1; i++ {
				screen.SetContent(i, suggY, '─', nil, borderStyle)
			}

			// Suggestions
			for i, sugg := range ai.filteredSuggs {
				if i >= numSuggs {
					break
				}
				rowY := suggY + 1 + i
				if rowY >= y+height-1 {
					break
				}

				// Left border
				screen.SetContent(suggX, rowY, '│', nil, borderStyle)

				// Content
				style := suggestionStyle
				catStyle := categoryStyle
				if i == ai.selectedIndex {
					style = selectedStyle
					catStyle = selectedStyle
				}

				// Clear row
				for cx := suggX + 1; cx < suggX+suggWidth-1; cx++ {
					screen.SetContent(cx, rowY, ' ', nil, style)
				}

				// Draw category tag if present
				textOffset := suggX + 2
				if sugg.Category != "" {
					catTag := "[" + sugg.Category + "]"
					for ci, r := range []rune(catTag) {
						if suggX+2+ci < suggX+suggWidth-1 {
							screen.SetContent(suggX+2+ci, rowY, r, nil, catStyle)
						}
					}
					textOffset = suggX + 2 + len([]rune(catTag)) + 1
				}

				// Draw suggestion text
				for ti, r := range []rune(sugg.Text) {
					if textOffset+ti < suggX+suggWidth-1 {
						screen.SetContent(textOffset+ti, rowY, r, nil, style)
					}
				}

				// Draw description if space permits
				if sugg.Description != "" {
					descOffset := textOffset + len([]rune(sugg.Text)) + 2
					if descOffset < suggX+suggWidth-10 {
						desc := "- " + sugg.Description
						descStyle := hintStyle
						if i == ai.selectedIndex {
							descStyle = selectedStyle
						}
						for di, r := range []rune(desc) {
							if descOffset+di < suggX+suggWidth-2 {
								screen.SetContent(descOffset+di, rowY, r, nil, descStyle)
							}
						}
					}
				}

				// Right border
				screen.SetContent(suggX+suggWidth-1, rowY, '│', nil, borderStyle)
			}

			// Bottom border
			bottomY := suggY + 1 + numSuggs
			if bottomY < y+height {
				screen.SetContent(suggX, bottomY, '└', nil, borderStyle)
				screen.SetContent(suggX+suggWidth-1, bottomY, '┘', nil, borderStyle)
				for i := suggX + 1; i < suggX+suggWidth-1; i++ {
					screen.SetContent(i, bottomY, '─', nil, borderStyle)
				}
			}
		}
	}
}

// InputHandler handles keyboard input.
func (ai *AutocompleteInput) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return ai.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyEnter:
			if ai.showSuggestion && ai.selectedIndex >= 0 && ai.selectedIndex < len(ai.filteredSuggs) {
				// Accept suggestion on Enter if dropdown is visible
				ai.acceptSuggestion()
			} else if ai.onSubmit != nil {
				ai.onSubmit(ai.text)
			}

		case tcell.KeyEscape:
			if ai.showSuggestion {
				// Close suggestions first
				ai.showSuggestion = false
			} else if ai.onCancel != nil {
				ai.onCancel()
			}

		case tcell.KeyTab:
			// Accept current suggestion
			if ai.showSuggestion && len(ai.filteredSuggs) > 0 {
				ai.acceptSuggestion()
				ai.updateSuggestions() // Show new suggestions
			}

		case tcell.KeyUp:
			if ai.showSuggestion && len(ai.filteredSuggs) > 0 {
				ai.selectedIndex--
				if ai.selectedIndex < 0 {
					ai.selectedIndex = len(ai.filteredSuggs) - 1
				}
			} else if ai.historyFn != nil {
				// Navigate to previous history entry
				if historyEntry := ai.historyFn(-1); historyEntry != "" {
					ai.text = historyEntry
					ai.cursorPos = len(historyEntry)
					ai.updateSuggestions()
					if ai.onChange != nil {
						ai.onChange(ai.text)
					}
				}
			}

		case tcell.KeyDown:
			if ai.showSuggestion && len(ai.filteredSuggs) > 0 {
				ai.selectedIndex++
				if ai.selectedIndex >= len(ai.filteredSuggs) {
					ai.selectedIndex = 0
				}
			} else if ai.historyFn != nil {
				// Navigate to next history entry
				historyEntry := ai.historyFn(+1)
				ai.text = historyEntry
				ai.cursorPos = len(historyEntry)
				ai.updateSuggestions()
				if ai.onChange != nil {
					ai.onChange(ai.text)
				}
			}

		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if ai.cursorPos > 0 {
				runes := []rune(ai.text)
				ai.text = string(runes[:ai.cursorPos-1]) + string(runes[ai.cursorPos:])
				ai.cursorPos--
				ai.updateSuggestions()
				if ai.onChange != nil {
					ai.onChange(ai.text)
				}
			}

		case tcell.KeyDelete:
			runes := []rune(ai.text)
			if ai.cursorPos < len(runes) {
				ai.text = string(runes[:ai.cursorPos]) + string(runes[ai.cursorPos+1:])
				ai.updateSuggestions()
				if ai.onChange != nil {
					ai.onChange(ai.text)
				}
			}

		case tcell.KeyLeft:
			if ai.cursorPos > 0 {
				ai.cursorPos--
			}

		case tcell.KeyRight:
			if ai.cursorPos < len(ai.text) {
				ai.cursorPos++
			}

		case tcell.KeyHome, tcell.KeyCtrlA:
			ai.cursorPos = 0

		case tcell.KeyEnd, tcell.KeyCtrlE:
			ai.cursorPos = len(ai.text)

		case tcell.KeyCtrlU:
			// Clear line
			ai.text = ""
			ai.cursorPos = 0
			ai.updateSuggestions()
			if ai.onChange != nil {
				ai.onChange(ai.text)
			}

		case tcell.KeyCtrlW:
			// Delete word backward
			if ai.cursorPos > 0 {
				runes := []rune(ai.text)
				pos := ai.cursorPos - 1
				// Skip trailing spaces
				for pos > 0 && runes[pos] == ' ' {
					pos--
				}
				// Skip word
				for pos > 0 && runes[pos-1] != ' ' {
					pos--
				}
				ai.text = string(runes[:pos]) + string(runes[ai.cursorPos:])
				ai.cursorPos = pos
				ai.updateSuggestions()
				if ai.onChange != nil {
					ai.onChange(ai.text)
				}
			}

		case tcell.KeyRune:
			r := event.Rune()
			runes := []rune(ai.text)
			ai.text = string(runes[:ai.cursorPos]) + string(r) + string(runes[ai.cursorPos:])
			ai.cursorPos++
			ai.updateSuggestions()
			if ai.onChange != nil {
				ai.onChange(ai.text)
			}
		}
	})
}

// Focus sets focus to this input.
func (ai *AutocompleteInput) Focus(delegate func(p tview.Primitive)) {
	ai.Box.Focus(delegate)
}

// HasFocus returns whether this input has focus.
func (ai *AutocompleteInput) HasFocus() bool {
	return ai.Box.HasFocus()
}

// GetPreferredHeight returns the height needed to display the input and suggestions.
func (ai *AutocompleteInput) GetPreferredHeight() int {
	if !ai.showSuggestion || len(ai.filteredSuggs) == 0 {
		return 3 // Just the input box
	}
	return 3 + min(len(ai.filteredSuggs), ai.maxSuggestions) + 2
}

// =============================================================================
// Helper Functions for Common Suggestion Patterns
// =============================================================================

// PrefixMatcher creates a suggestion provider that filters by prefix.
func PrefixMatcher(suggestions []Suggestion) SuggestionProvider {
	return func(text string, cursorPos int) []Suggestion {
		if text == "" {
			return suggestions
		}

		// Get the current word being typed
		textUpToCursor := text
		if cursorPos < len(text) {
			textUpToCursor = text[:cursorPos]
		}
		lastSpace := strings.LastIndexAny(textUpToCursor, " ()")
		currentToken := textUpToCursor
		if lastSpace >= 0 {
			currentToken = textUpToCursor[lastSpace+1:]
		}

		if currentToken == "" {
			return suggestions
		}

		currentTokenLower := strings.ToLower(currentToken)
		var filtered []Suggestion
		for _, s := range suggestions {
			if strings.HasPrefix(strings.ToLower(s.Text), currentTokenLower) {
				filtered = append(filtered, s)
			}
		}
		return filtered
	}
}

// FuzzyMatcher creates a suggestion provider that does fuzzy matching.
func FuzzyMatcher(suggestions []Suggestion) SuggestionProvider {
	return func(text string, cursorPos int) []Suggestion {
		if text == "" {
			return suggestions
		}

		// Get the current word being typed
		textUpToCursor := text
		if cursorPos < len(text) {
			textUpToCursor = text[:cursorPos]
		}
		lastSpace := strings.LastIndexAny(textUpToCursor, " ()")
		currentToken := textUpToCursor
		if lastSpace >= 0 {
			currentToken = textUpToCursor[lastSpace+1:]
		}

		if currentToken == "" {
			return suggestions
		}

		currentTokenLower := strings.ToLower(currentToken)
		var filtered []Suggestion
		for _, s := range suggestions {
			if fuzzyMatch(strings.ToLower(s.Text), currentTokenLower) {
				filtered = append(filtered, s)
			}
		}
		return filtered
	}
}

// fuzzyMatch checks if pattern chars appear in text in order.
func fuzzyMatch(text, pattern string) bool {
	pi := 0
	for ti := 0; ti < len(text) && pi < len(pattern); ti++ {
		if text[ti] == pattern[pi] {
			pi++
		}
	}
	return pi == len(pattern)
}

// StaticSuggestions returns a provider that always returns the same suggestions.
func StaticSuggestions(suggestions []Suggestion) SuggestionProvider {
	return func(text string, cursorPos int) []Suggestion {
		return suggestions
	}
}

// ChainedProvider chains multiple suggestion providers together.
func ChainedProvider(providers ...SuggestionProvider) SuggestionProvider {
	return func(text string, cursorPos int) []Suggestion {
		var all []Suggestion
		for _, p := range providers {
			all = append(all, p(text, cursorPos)...)
		}
		return all
	}
}
