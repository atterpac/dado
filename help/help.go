package help

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/input"
	"github.com/atterpac/dado/theme"
)

// Section represents a group of key bindings.
type Section struct {
	Name    string
	Actions []ActionInfo
}

// ActionInfo contains action details for display.
type ActionInfo struct {
	Key         string
	Description string
}

// Help manages application help information.
type Help struct {
	appName  string
	version  string
	sections []Section
}

// New creates a new Help instance.
func New() *Help {
	return &Help{}
}

// SetAppName sets the application name.
func (h *Help) SetAppName(name string) *Help {
	h.appName = name
	return h
}

// SetVersion sets the application version.
func (h *Help) SetVersion(version string) *Help {
	h.version = version
	return h
}

// AddSection adds a section with actions.
func (h *Help) AddSection(name string, actions []ActionInfo) *Help {
	h.sections = append(h.sections, Section{
		Name:    name,
		Actions: actions,
	})
	return h
}

// AddRegistry adds a section from an ActionRegistry.
func (h *Help) AddRegistry(name string, registry *input.ActionRegistry) *Help {
	hints := registry.Hints()
	actions := make([]ActionInfo, len(hints))
	for i, hint := range hints {
		actions[i] = ActionInfo{
			Key:         hint.Key,
			Description: hint.Description,
		}
	}
	return h.AddSection(name, actions)
}

// Clear removes all sections.
func (h *Help) Clear() *Help {
	h.sections = nil
	return h
}

// Modal returns a help modal component.
func (h *Help) Modal() *HelpModal {
	return NewHelpModal(h)
}

// GetSections returns all sections.
func (h *Help) GetSections() []Section {
	return h.sections
}

// ContextHints returns hints for a specific section.
func (h *Help) ContextHints(sectionName string) []components.KeyHint {
	for _, section := range h.sections {
		if section.Name == sectionName {
			hints := make([]components.KeyHint, len(section.Actions))
			for i, action := range section.Actions {
				hints[i] = components.KeyHint{
					Key:         action.Key,
					Description: action.Description,
				}
			}
			return hints
		}
	}
	return nil
}

// ExportMarkdown exports help to a Markdown file.
func (h *Help) ExportMarkdown(path string) error {
	var sb strings.Builder

	// Header
	if h.appName != "" {
		sb.WriteString(fmt.Sprintf("# %s", h.appName))
		if h.version != "" {
			sb.WriteString(fmt.Sprintf(" v%s", h.version))
		}
		sb.WriteString("\n\n")
	}

	sb.WriteString("## Keyboard Shortcuts\n\n")

	// Sections
	for _, section := range h.sections {
		sb.WriteString(fmt.Sprintf("### %s\n\n", section.Name))
		sb.WriteString("| Key | Description |\n")
		sb.WriteString("|-----|-------------|\n")
		for _, action := range section.Actions {
			sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", action.Key, action.Description))
		}
		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// ExportManPage exports help to a man page format.
func (h *Help) ExportManPage(path string) error {
	var sb strings.Builder

	// Man page header
	name := h.appName
	if name == "" {
		name = "app"
	}

	sb.WriteString(fmt.Sprintf(".TH %s 1\n", strings.ToUpper(name)))
	sb.WriteString(".SH NAME\n")
	sb.WriteString(fmt.Sprintf("%s \\- terminal user interface application\n", name))

	sb.WriteString(".SH SYNOPSIS\n")
	sb.WriteString(fmt.Sprintf(".B %s\n", name))

	sb.WriteString(".SH DESCRIPTION\n")
	sb.WriteString(fmt.Sprintf("%s is a terminal user interface application.\n", name))

	sb.WriteString(".SH KEYBOARD SHORTCUTS\n")

	for _, section := range h.sections {
		sb.WriteString(fmt.Sprintf(".SS %s\n", section.Name))
		for _, action := range section.Actions {
			sb.WriteString(fmt.Sprintf(".TP\n.B %s\n%s\n", action.Key, action.Description))
		}
	}

	if h.version != "" {
		sb.WriteString(".SH VERSION\n")
		sb.WriteString(fmt.Sprintf("%s\n", h.version))
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// HelpModal is a modal dialog displaying help information.
type HelpModal struct {
	*tview.Box

	help         *Help
	scrollOffset int
	totalLines   int
	visible      bool
}

// NewHelpModal creates a new HelpModal.
func NewHelpModal(help *Help) *HelpModal {
	return &HelpModal{
		Box:  tview.NewBox(),
		help: help,
	}
}

// Show makes the modal visible.
func (m *HelpModal) Show() *HelpModal {
	m.visible = true
	m.scrollOffset = 0
	return m
}

// Hide makes the modal invisible.
func (m *HelpModal) Hide() *HelpModal {
	m.visible = false
	return m
}

// IsVisible returns whether the modal is visible.
func (m *HelpModal) IsVisible() bool {
	return m.visible
}

// Draw renders the help modal.
func (m *HelpModal) Draw(screen tcell.Screen) {
	if !m.visible {
		return
	}

	screenWidth, screenHeight := screen.Size()

	// Modal dimensions
	modalWidth := screenWidth * 3 / 4
	if modalWidth > 80 {
		modalWidth = 80
	}
	if modalWidth < 40 {
		modalWidth = 40
	}

	modalHeight := screenHeight * 3 / 4
	if modalHeight > 30 {
		modalHeight = 30
	}
	if modalHeight < 10 {
		modalHeight = 10
	}

	// Center modal
	x := (screenWidth - modalWidth) / 2
	y := (screenHeight - modalHeight) / 2

	// Get colors
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()
	borderColor := theme.BorderFocus()

	borderStyle := tcell.StyleDefault.Background(bgColor).Foreground(borderColor)
	titleStyle := tcell.StyleDefault.Background(bgColor).Foreground(accentColor)
	textStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
	dimStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
	keyStyle := tcell.StyleDefault.Background(bgColor).Foreground(accentColor)

	// Draw dark backdrop
	backdropStyle := tcell.StyleDefault.Background(tcell.ColorBlack)
	for row := 0; row < screenHeight; row++ {
		for col := 0; col < screenWidth; col++ {
			screen.SetContent(col, row, ' ', nil, backdropStyle)
		}
	}

	// Clear modal area
	clearStyle := tcell.StyleDefault.Background(bgColor)
	for row := y; row < y+modalHeight; row++ {
		for col := x; col < x+modalWidth; col++ {
			screen.SetContent(col, row, ' ', nil, clearStyle)
		}
	}

	// Draw border
	screen.SetContent(x, y, '╭', nil, borderStyle)
	screen.SetContent(x+modalWidth-1, y, '╮', nil, borderStyle)
	screen.SetContent(x, y+modalHeight-1, '╰', nil, borderStyle)
	screen.SetContent(x+modalWidth-1, y+modalHeight-1, '╯', nil, borderStyle)

	for col := x + 1; col < x+modalWidth-1; col++ {
		screen.SetContent(col, y, '─', nil, borderStyle)
		screen.SetContent(col, y+modalHeight-1, '─', nil, borderStyle)
	}

	for row := y + 1; row < y+modalHeight-1; row++ {
		screen.SetContent(x, row, '│', nil, borderStyle)
		screen.SetContent(x+modalWidth-1, row, '│', nil, borderStyle)
	}

	// Draw title
	title := " Help "
	if m.help.appName != "" {
		title = fmt.Sprintf(" %s Help ", m.help.appName)
	}
	titleX := x + (modalWidth-len(title))/2
	for i, r := range title {
		screen.SetContent(titleX+i, y, r, nil, titleStyle)
	}

	// Build content lines
	var lines []struct {
		text    string
		style   tcell.Style
		isKey   bool
		keyText string
	}

	for _, section := range m.help.sections {
		// Section header
		lines = append(lines, struct {
			text    string
			style   tcell.Style
			isKey   bool
			keyText string
		}{text: section.Name, style: titleStyle})

		// Separator
		lines = append(lines, struct {
			text    string
			style   tcell.Style
			isKey   bool
			keyText string
		}{text: strings.Repeat("─", modalWidth-4), style: dimStyle})

		// Actions
		for _, action := range section.Actions {
			lines = append(lines, struct {
				text    string
				style   tcell.Style
				isKey   bool
				keyText string
			}{
				text:    action.Description,
				style:   textStyle,
				isKey:   true,
				keyText: action.Key,
			})
		}

		// Empty line between sections
		lines = append(lines, struct {
			text    string
			style   tcell.Style
			isKey   bool
			keyText string
		}{text: "", style: textStyle})
	}

	m.totalLines = len(lines)

	// Draw content
	contentHeight := modalHeight - 4 // borders + footer
	contentY := y + 2
	contentX := x + 2

	// Adjust scroll offset
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	maxOffset := m.totalLines - contentHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}

	for i := 0; i < contentHeight && m.scrollOffset+i < len(lines); i++ {
		line := lines[m.scrollOffset+i]
		row := contentY + i

		if line.isKey {
			// Draw key in brackets
			col := contentX
			keyFormatted := fmt.Sprintf("[%s]", line.keyText)
			for _, r := range keyFormatted {
				if col < contentX+15 { // Fixed width for keys
					screen.SetContent(col, row, r, nil, keyStyle)
					col++
				}
			}
			// Pad to fixed width
			for col < contentX+15 {
				screen.SetContent(col, row, ' ', nil, clearStyle)
				col++
			}
			// Draw description
			for _, r := range line.text {
				if col < x+modalWidth-2 {
					screen.SetContent(col, row, r, nil, line.style)
					col++
				}
			}
		} else {
			col := contentX
			for _, r := range line.text {
				if col < x+modalWidth-2 {
					screen.SetContent(col, row, r, nil, line.style)
					col++
				}
			}
		}
	}

	// Draw scroll indicators
	if m.scrollOffset > 0 {
		screen.SetContent(x+modalWidth-2, y+1, '▲', nil, dimStyle)
	}
	if m.scrollOffset < maxOffset {
		screen.SetContent(x+modalWidth-2, y+modalHeight-3, '▼', nil, dimStyle)
	}

	// Draw footer
	footer := "[Esc] Close  [j/k] Scroll"
	footerX := x + (modalWidth-len(footer))/2
	for i, r := range footer {
		screen.SetContent(footerX+i, y+modalHeight-2, r, nil, dimStyle)
	}
}

// InputHandler handles keyboard input.
func (m *HelpModal) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if !m.visible {
			return
		}

		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyEnter:
			m.Hide()
		case tcell.KeyUp:
			m.scrollOffset--
		case tcell.KeyDown:
			m.scrollOffset++
		case tcell.KeyPgUp:
			m.scrollOffset -= 10
		case tcell.KeyPgDn:
			m.scrollOffset += 10
		case tcell.KeyHome:
			m.scrollOffset = 0
		case tcell.KeyEnd:
			m.scrollOffset = m.totalLines
		case tcell.KeyRune:
			switch event.Rune() {
			case 'k':
				m.scrollOffset--
			case 'j':
				m.scrollOffset++
			case 'g':
				m.scrollOffset = 0
			case 'G':
				m.scrollOffset = m.totalLines
			case 'q':
				m.Hide()
			}
		}
	}
}

// Focus handles focus.
func (m *HelpModal) Focus(delegate func(tview.Primitive)) {
	// Modal captures focus
}

// HasFocus returns whether the modal has focus.
func (m *HelpModal) HasFocus() bool {
	return m.visible
}

// MouseHandler handles mouse input.
func (m *HelpModal) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		if !m.visible {
			return false, nil
		}

		switch action {
		case tview.MouseScrollUp:
			m.scrollOffset--
			return true, m
		case tview.MouseScrollDown:
			m.scrollOffset++
			return true, m
		case tview.MouseLeftClick:
			// Check if click is outside modal to close
			screenWidth, screenHeight := event.Position()
			_ = screenWidth
			_ = screenHeight
			// For simplicity, any click closes
			m.Hide()
			return true, m
		}

		return true, m
	}
}
