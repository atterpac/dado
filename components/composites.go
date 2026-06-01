package components

import (
	"github.com/gdamore/tcell/v2"
)

// =============================================================================
// StatusBar - Application status bar with sections
// =============================================================================

// StatusSection represents a section of the status bar
type StatusSection struct {
	Text     string
	Icon     string
	Color    tcell.Color // 0 = default
	MinWidth int
	Align    int // core.AlignLeft, AlignCenter, AlignRight
}

// StatusBar displays status information in a horizontal bar
type StatusBar struct {
	widgetBase

	// Sections
	leftSections   []StatusSection
	centerSections []StatusSection
	rightSections  []StatusSection

	// Separator
	separator rune

	// Styling
	showBorder bool
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	s := &StatusBar{
		separator: '│',
	}
	s.initWidget()
	return s
}

// SetLeft sets left-aligned sections
func (s *StatusBar) SetLeft(sections ...StatusSection) *StatusBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.leftSections = sections
	return s
}

// SetCenter sets center-aligned sections
func (s *StatusBar) SetCenter(sections ...StatusSection) *StatusBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.centerSections = sections
	return s
}

// SetRight sets right-aligned sections
func (s *StatusBar) SetRight(sections ...StatusSection) *StatusBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rightSections = sections
	return s
}

// AddLeft adds a left section
func (s *StatusBar) AddLeft(section StatusSection) *StatusBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.leftSections = append(s.leftSections, section)
	return s
}

// AddRight adds a right section
func (s *StatusBar) AddRight(section StatusSection) *StatusBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rightSections = append(s.rightSections, section)
	return s
}

// UpdateSection updates a section by matching text prefix
func (s *StatusBar) UpdateSection(prefix, newText string) *StatusBar {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.leftSections {
		if len(s.leftSections[i].Text) >= len(prefix) && s.leftSections[i].Text[:len(prefix)] == prefix {
			s.leftSections[i].Text = newText
			return s
		}
	}
	for i := range s.centerSections {
		if len(s.centerSections[i].Text) >= len(prefix) && s.centerSections[i].Text[:len(prefix)] == prefix {
			s.centerSections[i].Text = newText
			return s
		}
	}
	for i := range s.rightSections {
		if len(s.rightSections[i].Text) >= len(prefix) && s.rightSections[i].Text[:len(prefix)] == prefix {
			s.rightSections[i].Text = newText
			return s
		}
	}
	return s
}

// SetSeparator sets the section separator character
func (s *StatusBar) SetSeparator(sep rune) *StatusBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.separator = sep
	return s
}

// SetShowBorder enables/disables the top border
func (s *StatusBar) SetShowBorder(show bool) *StatusBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.showBorder = show
	return s
}

// Clear removes all sections
func (s *StatusBar) Clear() *StatusBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.leftSections = nil
	s.centerSections = nil
	s.rightSections = nil
	return s
}

func (s *StatusBar) renderSection(section StatusSection) string {
	result := ""
	if section.Icon != "" {
		result = section.Icon + " "
	}
	result += section.Text
	return result
}

// Draw renders the status bar
func (s *StatusBar) Draw(screen tcell.Screen) {
	s.Box.DrawForSubclass(screen)
	x, y, width, height := s.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	th := s.th()
	bgColor := th.BgLight()
	fgColor := th.Fg()
	fgDimColor := th.FgDim()

	bgStyle := tcell.StyleDefault.Background(bgColor)
	sepStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)

	// Clear area
	fillRect(screen, x, y, width, height, bgStyle)

	// Draw border if enabled
	contentY := y
	if s.showBorder && height > 1 {
		borderStyle := tcell.StyleDefault.Background(th.Bg()).Foreground(fgDimColor)
		for col := x; col < x+width; col++ {
			screen.SetContent(col, y, '─', nil, borderStyle)
		}
		contentY++
	}

	// Render left sections
	col := x + 1
	for i, section := range s.leftSections {
		if i > 0 {
			screen.SetContent(col, contentY, ' ', nil, bgStyle)
			col++
			screen.SetContent(col, contentY, s.separator, nil, sepStyle)
			col++
			screen.SetContent(col, contentY, ' ', nil, bgStyle)
			col++
		}

		text := s.renderSection(section)
		textColor := fgColor
		if section.Color != 0 {
			textColor = section.Color
		}
		textStyle := tcell.StyleDefault.Background(bgColor).Foreground(textColor)

		col = drawText(screen, col, contentY, x+width-col, text, textStyle)
	}

	// Render right sections (from right edge)
	rightCol := x + width - 1
	for i := len(s.rightSections) - 1; i >= 0; i-- {
		section := s.rightSections[i]
		text := s.renderSection(section)

		textColor := fgColor
		if section.Color != 0 {
			textColor = section.Color
		}
		textStyle := tcell.StyleDefault.Background(bgColor).Foreground(textColor)

		// Draw from right
		for j := len(text) - 1; j >= 0; j-- {
			if rightCol > col { // Don't overlap with left
				screen.SetContent(rightCol, contentY, rune(text[j]), nil, textStyle)
				rightCol--
			}
		}

		if i > 0 && rightCol > col {
			screen.SetContent(rightCol, contentY, ' ', nil, bgStyle)
			rightCol--
			screen.SetContent(rightCol, contentY, s.separator, nil, sepStyle)
			rightCol--
			screen.SetContent(rightCol, contentY, ' ', nil, bgStyle)
			rightCol--
		}
	}

	// Render center sections
	if len(s.centerSections) > 0 {
		// Calculate total center width
		centerText := ""
		for i, section := range s.centerSections {
			if i > 0 {
				centerText += " " + string(s.separator) + " "
			}
			centerText += s.renderSection(section)
		}

		centerStart := x + (width-len(centerText))/2
		centerCol := centerStart

		for i, section := range s.centerSections {
			if i > 0 {
				if centerCol < rightCol && centerCol > col {
					screen.SetContent(centerCol, contentY, ' ', nil, bgStyle)
					centerCol++
					screen.SetContent(centerCol, contentY, s.separator, nil, sepStyle)
					centerCol++
					screen.SetContent(centerCol, contentY, ' ', nil, bgStyle)
					centerCol++
				}
			}

			text := s.renderSection(section)
			textColor := fgColor
			if section.Color != 0 {
				textColor = section.Color
			}
			textStyle := tcell.StyleDefault.Background(bgColor).Foreground(textColor)

			for _, r := range text {
				if centerCol < rightCol && centerCol > col {
					screen.SetContent(centerCol, contentY, r, nil, textStyle)
					centerCol++
				}
			}
		}
	}
}

// GetFieldHeight returns preferred height
func (s *StatusBar) GetFieldHeight() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.showBorder {
		return 2
	}
	return 1
}

// =============================================================================
// SearchBar - Search input with results and filters
// =============================================================================

// SearchResult represents a search result item
type SearchResult struct {
	Text        string
	Description string
	Icon        string
	Data        any
}

// SearchBar provides search input with optional results dropdown
type SearchBar struct {
	widgetBase

	// Input
	query       string
	placeholder string
	cursorPos   int

	// Results
	results     []SearchResult
	selectedIdx int
	showResults bool
	maxResults  int

	// Display
	icon         string
	showClear    bool
	showSpinner  bool
	spinnerFrame int

	// Callbacks
	onSearch func(query string)
	onSelect func(result SearchResult)
	onCancel func()
	onChange func(query string)
}

// NewSearchBar creates a new search bar
func NewSearchBar() *SearchBar {
	s := &SearchBar{
		placeholder: "Search...",
		icon:        "🔍",
		maxResults:  10,
		selectedIdx: -1,
	}
	s.initWidget()
	return s
}

// SetPlaceholder sets the placeholder text
func (s *SearchBar) SetPlaceholder(text string) *SearchBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.placeholder = text
	return s
}

// SetIcon sets the search icon
func (s *SearchBar) SetIcon(icon string) *SearchBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.icon = icon
	return s
}

// SetQuery sets the search query
func (s *SearchBar) SetQuery(query string) *SearchBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.query = query
	s.cursorPos = len(query)
	return s
}

// GetQuery returns the current query
func (s *SearchBar) GetQuery() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.query
}

// SetResults sets the search results
func (s *SearchBar) SetResults(results []SearchResult) *SearchBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results = results
	s.showResults = len(results) > 0
	s.selectedIdx = -1
	if len(results) > 0 {
		s.selectedIdx = 0
	}
	return s
}

// ClearResults removes all results
func (s *SearchBar) ClearResults() *SearchBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results = nil
	s.showResults = false
	s.selectedIdx = -1
	return s
}

// SetShowSpinner shows/hides the loading spinner
func (s *SearchBar) SetShowSpinner(show bool) *SearchBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.showSpinner = show
	return s
}

// SetMaxResults sets maximum results to display
func (s *SearchBar) SetMaxResults(max int) *SearchBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxResults = max
	return s
}

// SetOnSearch sets the search callback
func (s *SearchBar) SetOnSearch(fn func(query string)) *SearchBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onSearch = fn
	return s
}

// SetOnSelect sets the result selection callback
func (s *SearchBar) SetOnSelect(fn func(result SearchResult)) *SearchBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onSelect = fn
	return s
}

// SetOnCancel sets the cancel callback
func (s *SearchBar) SetOnCancel(fn func()) *SearchBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onCancel = fn
	return s
}

// SetOnChange sets the query change callback
func (s *SearchBar) SetOnChange(fn func(query string)) *SearchBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onChange = fn
	return s
}

// Clear clears the query and results
func (s *SearchBar) Clear() *SearchBar {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.query = ""
	s.cursorPos = 0
	s.results = nil
	s.showResults = false
	s.selectedIdx = -1
	return s
}

// Focus gives focus to the search bar
// Draw renders the search bar
func (s *SearchBar) Draw(screen tcell.Screen) {
	s.Box.DrawForSubclass(screen)
	x, y, width, height := s.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	th := s.th()
	bgColor := th.Bg()
	inputBg := th.BgLight()
	fgColor := th.Fg()
	fgDimColor := th.FgDim()
	accentColor := th.Accent()

	bgStyle := tcell.StyleDefault.Background(bgColor)
	inputStyle := tcell.StyleDefault.Background(inputBg).Foreground(fgColor)
	placeholderStyle := tcell.StyleDefault.Background(inputBg).Foreground(fgDimColor)
	iconStyle := tcell.StyleDefault.Background(inputBg).Foreground(accentColor)

	// Clear background
	fillRect(screen, x, y, width, height, bgStyle)

	// Draw input box
	inputY := y
	inputHeight := 1

	// Draw input background
	fillLine(screen, x, inputY, width, inputStyle)

	col := x + 1

	// Draw icon
	if s.icon != "" {
		col = drawText(screen, col, inputY, x+width-1-col, s.icon, iconStyle)
		col++ // space
	}

	// Draw query or placeholder
	inputStart := col
	if s.query == "" {
		col = drawText(screen, col, inputY, x+width-1-col, s.placeholder, placeholderStyle)
	} else {
		col = drawText(screen, col, inputY, x+width-1-col, s.query, inputStyle)
	}

	// Draw cursor if focused
	if s.HasFocus() {
		cursorX := inputStart + s.cursorPos
		if cursorX < x+width-1 {
			cursorStyle := tcell.StyleDefault.Background(fgColor).Foreground(inputBg)
			char := ' '
			if s.cursorPos < len(s.query) {
				char = rune(s.query[s.cursorPos])
			}
			screen.SetContent(cursorX, inputY, char, nil, cursorStyle)
		}
	}

	// Draw spinner if loading
	if s.showSpinner {
		spinnerChars := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
		spinnerStyle := tcell.StyleDefault.Background(inputBg).Foreground(accentColor)
		screen.SetContent(x+width-2, inputY, spinnerChars[s.spinnerFrame%len(spinnerChars)], nil, spinnerStyle)
	}

	// Draw results dropdown if visible
	if s.showResults && len(s.results) > 0 && height > inputHeight {
		resultsY := inputY + inputHeight
		maxVisible := height - inputHeight
		if maxVisible > s.maxResults {
			maxVisible = s.maxResults
		}
		if maxVisible > len(s.results) {
			maxVisible = len(s.results)
		}

		resultBg := th.BgLight()
		selectedBg := accentColor

		for i := 0; i < maxVisible; i++ {
			result := s.results[i]
			rowY := resultsY + i
			isSelected := i == s.selectedIdx

			rowBg := resultBg
			rowFg := fgColor
			if isSelected {
				rowBg = selectedBg
				rowFg = bgColor
			}

			rowStyle := tcell.StyleDefault.Background(rowBg).Foreground(rowFg)
			dimStyle := tcell.StyleDefault.Background(rowBg).Foreground(fgDimColor)
			if isSelected {
				dimStyle = rowStyle
			}

			// Clear row
			fillLine(screen, x, rowY, width, rowStyle)

			col := x + 2

			// Icon
			if result.Icon != "" {
				col = drawText(screen, col, rowY, x+width-1-col, result.Icon, rowStyle)
				col++ // space
			}

			// Text
			col = drawText(screen, col, rowY, x+width-1-col, result.Text, rowStyle)

			// Description (dimmed, right-aligned)
			if result.Description != "" && col < x+width-len(result.Description)-2 {
				descCol := x + width - len(result.Description) - 2
				drawText(screen, descCol, rowY, x+width-1-descCol, result.Description, dimStyle)
			}
		}
	}
}

// GetFieldHeight returns preferred height
func (s *SearchBar) GetFieldHeight() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.showResults && len(s.results) > 0 {
		visible := len(s.results)
		if visible > s.maxResults {
			visible = s.maxResults
		}
		return 1 + visible
	}
	return 1
}
