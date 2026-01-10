package components

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
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
	Align    int // tview.AlignLeft, AlignCenter, AlignRight
}

// StatusBar displays status information in a horizontal bar
type StatusBar struct {
	*tview.Box

	mu sync.RWMutex

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
	return &StatusBar{
		Box:       tview.NewBox(),
		separator: '│',
	}
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
	s.Box.DrawForSubclass(screen, s)
	x, y, width, height := s.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	bgColor := theme.BgLight()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()

	bgStyle := tcell.StyleDefault.Background(bgColor)
	sepStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)

	// Clear area
	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {
			screen.SetContent(col, row, ' ', nil, bgStyle)
		}
	}

	// Draw border if enabled
	contentY := y
	if s.showBorder && height > 1 {
		borderStyle := tcell.StyleDefault.Background(theme.Bg()).Foreground(fgDimColor)
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

		for _, r := range text {
			if col < x+width {
				screen.SetContent(col, contentY, r, nil, textStyle)
				col++
			}
		}
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
	*tview.Box

	mu sync.RWMutex

	// Input
	query       string
	placeholder string
	cursorPos   int

	// Results
	results       []SearchResult
	selectedIdx   int
	showResults   bool
	maxResults    int

	// Display
	icon          string
	showClear     bool
	showSpinner   bool
	spinnerFrame  int

	// Callbacks
	onSearch  func(query string)
	onSelect  func(result SearchResult)
	onCancel  func()
	onChange  func(query string)
}

// NewSearchBar creates a new search bar
func NewSearchBar() *SearchBar {
	return &SearchBar{
		Box:         tview.NewBox(),
		placeholder: "Search...",
		icon:        "🔍",
		maxResults:  10,
		selectedIdx: -1,
	}
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
func (s *SearchBar) Focus(delegate func(tview.Primitive)) {
	s.Box.Focus(delegate)
}

// Draw renders the search bar
func (s *SearchBar) Draw(screen tcell.Screen) {
	s.Box.DrawForSubclass(screen, s)
	x, y, width, height := s.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	bgColor := theme.Bg()
	inputBg := theme.BgLight()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()

	bgStyle := tcell.StyleDefault.Background(bgColor)
	inputStyle := tcell.StyleDefault.Background(inputBg).Foreground(fgColor)
	placeholderStyle := tcell.StyleDefault.Background(inputBg).Foreground(fgDimColor)
	iconStyle := tcell.StyleDefault.Background(inputBg).Foreground(accentColor)

	// Clear background
	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {
			screen.SetContent(col, row, ' ', nil, bgStyle)
		}
	}

	// Draw input box
	inputY := y
	inputHeight := 1

	// Draw input background
	for col := x; col < x+width; col++ {
		screen.SetContent(col, inputY, ' ', nil, inputStyle)
	}

	col := x + 1

	// Draw icon
	if s.icon != "" {
		for _, r := range s.icon {
			if col < x+width-1 {
				screen.SetContent(col, inputY, r, nil, iconStyle)
				col++
			}
		}
		col++ // space
	}

	// Draw query or placeholder
	inputStart := col
	if s.query == "" {
		for _, r := range s.placeholder {
			if col < x+width-1 {
				screen.SetContent(col, inputY, r, nil, placeholderStyle)
				col++
			}
		}
	} else {
		for _, r := range s.query {
			if col < x+width-1 {
				screen.SetContent(col, inputY, r, nil, inputStyle)
				col++
			}
		}
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

		resultBg := theme.BgLight()
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
			for col := x; col < x+width; col++ {
				screen.SetContent(col, rowY, ' ', nil, rowStyle)
			}

			col := x + 2

			// Icon
			if result.Icon != "" {
				for _, r := range result.Icon {
					if col < x+width-1 {
						screen.SetContent(col, rowY, r, nil, rowStyle)
						col++
					}
				}
				col++ // space
			}

			// Text
			for _, r := range result.Text {
				if col < x+width-1 {
					screen.SetContent(col, rowY, r, nil, rowStyle)
					col++
				}
			}

			// Description (dimmed, right-aligned)
			if result.Description != "" && col < x+width-len(result.Description)-2 {
				descCol := x + width - len(result.Description) - 2
				for _, r := range result.Description {
					if descCol < x+width-1 {
						screen.SetContent(descCol, rowY, r, nil, dimStyle)
						descCol++
					}
				}
			}
		}
	}
}

// InputHandler handles keyboard input
func (s *SearchBar) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return s.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		s.mu.Lock()

		switch event.Key() {
		case tcell.KeyEscape:
			onCancel := s.onCancel
			s.mu.Unlock()
			if onCancel != nil {
				onCancel()
			}
			return

		case tcell.KeyEnter:
			if s.selectedIdx >= 0 && s.selectedIdx < len(s.results) {
				result := s.results[s.selectedIdx]
				onSelect := s.onSelect
				s.mu.Unlock()
				if onSelect != nil {
					onSelect(result)
				}
				return
			} else if s.query != "" {
				query := s.query
				onSearch := s.onSearch
				s.mu.Unlock()
				if onSearch != nil {
					onSearch(query)
				}
				return
			}

		case tcell.KeyDown:
			if s.showResults && s.selectedIdx < len(s.results)-1 {
				s.selectedIdx++
			}

		case tcell.KeyUp:
			if s.showResults && s.selectedIdx > 0 {
				s.selectedIdx--
			}

		case tcell.KeyLeft:
			if s.cursorPos > 0 {
				s.cursorPos--
			}

		case tcell.KeyRight:
			if s.cursorPos < len(s.query) {
				s.cursorPos++
			}

		case tcell.KeyHome:
			s.cursorPos = 0

		case tcell.KeyEnd:
			s.cursorPos = len(s.query)

		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if s.cursorPos > 0 {
				s.query = s.query[:s.cursorPos-1] + s.query[s.cursorPos:]
				s.cursorPos--
				onChange := s.onChange
				query := s.query
				s.mu.Unlock()
				if onChange != nil {
					onChange(query)
				}
				return
			}

		case tcell.KeyDelete:
			if s.cursorPos < len(s.query) {
				s.query = s.query[:s.cursorPos] + s.query[s.cursorPos+1:]
				onChange := s.onChange
				query := s.query
				s.mu.Unlock()
				if onChange != nil {
					onChange(query)
				}
				return
			}

		case tcell.KeyRune:
			r := event.Rune()
			s.query = s.query[:s.cursorPos] + string(r) + s.query[s.cursorPos:]
			s.cursorPos++
			onChange := s.onChange
			query := s.query
			s.mu.Unlock()
			if onChange != nil {
				onChange(query)
			}
			return

		case tcell.KeyCtrlU:
			s.query = s.query[s.cursorPos:]
			s.cursorPos = 0
			onChange := s.onChange
			query := s.query
			s.mu.Unlock()
			if onChange != nil {
				onChange(query)
			}
			return

		case tcell.KeyCtrlK:
			s.query = s.query[:s.cursorPos]
			onChange := s.onChange
			query := s.query
			s.mu.Unlock()
			if onChange != nil {
				onChange(query)
			}
			return

		case tcell.KeyCtrlW:
			// Delete word backward
			if s.cursorPos > 0 {
				pos := s.cursorPos - 1
				for pos > 0 && s.query[pos] == ' ' {
					pos--
				}
				for pos > 0 && s.query[pos-1] != ' ' {
					pos--
				}
				s.query = s.query[:pos] + s.query[s.cursorPos:]
				s.cursorPos = pos
				onChange := s.onChange
				query := s.query
				s.mu.Unlock()
				if onChange != nil {
					onChange(query)
				}
				return
			}
		}

		s.mu.Unlock()
	})
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
