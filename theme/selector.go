package theme

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ThemeSelectorModal displays a modal for selecting themes with live preview.
type ThemeSelectorModal struct {
	*tview.Box
	table         *tview.Table
	themes        []string
	originalTheme string
	currentIdx    int
	onSelect      func(name string)
	onCancel      func()
	onPreview     func(name string)
}

// NewThemeSelectorModal creates a new theme selector modal.
func NewThemeSelectorModal(themes []string, currentTheme string) *ThemeSelectorModal {
	tsm := &ThemeSelectorModal{
		Box:           tview.NewBox(),
		table:         tview.NewTable(),
		themes:        themes,
		originalTheme: currentTheme,
	}
	tsm.setup()
	return tsm
}

// SetOnSelect sets the callback when a theme is selected (Enter pressed).
func (tsm *ThemeSelectorModal) SetOnSelect(fn func(name string)) *ThemeSelectorModal {
	tsm.onSelect = fn
	return tsm
}

// SetOnCancel sets the callback when selection is cancelled (Esc pressed).
func (tsm *ThemeSelectorModal) SetOnCancel(fn func()) *ThemeSelectorModal {
	tsm.onCancel = fn
	return tsm
}

// SetOnPreview sets the callback for live theme preview during navigation.
func (tsm *ThemeSelectorModal) SetOnPreview(fn func(name string)) *ThemeSelectorModal {
	tsm.onPreview = fn
	return tsm
}

func (tsm *ThemeSelectorModal) setup() {
	// Configure table
	tsm.table.SetSelectable(true, false)
	tsm.table.SetSelectedStyle(SelectionStyle())

	// Populate table
	tsm.rebuildTable()

	// Find and select current theme row
	for i, name := range tsm.themes {
		if name == tsm.originalTheme {
			tsm.currentIdx = i
			tsm.table.Select(i, 0)
			break
		}
	}

	// Preview theme on selection change
	tsm.table.SetSelectionChangedFunc(func(row, col int) {
		if row >= 0 && row < len(tsm.themes) {
			tsm.currentIdx = row
			if tsm.onPreview != nil {
				tsm.onPreview(tsm.themes[row])
			}
		}
	})

	// Handle final selection
	tsm.table.SetSelectedFunc(func(row, col int) {
		if tsm.onSelect != nil {
			selected := tsm.GetSelectedTheme()
			if selected != "" {
				tsm.onSelect(selected)
			}
		}
	})
}

func (tsm *ThemeSelectorModal) rebuildTable() {
	tsm.table.Clear()

	for i, name := range tsm.themes {
		marker := "  "
		if name == tsm.originalTheme {
			marker = "✓ "
		}
		cell := tview.NewTableCell(marker + name).
			SetExpansion(1)
		tsm.table.SetCell(i, 0, cell)
	}
}

// GetSelectedTheme returns the currently highlighted theme name.
func (tsm *ThemeSelectorModal) GetSelectedTheme() string {
	if tsm.currentIdx >= 0 && tsm.currentIdx < len(tsm.themes) {
		return tsm.themes[tsm.currentIdx]
	}
	return ""
}

// GetOriginalTheme returns the theme that was active when the modal opened.
func (tsm *ThemeSelectorModal) GetOriginalTheme() string {
	return tsm.originalTheme
}

// Draw renders the modal with border and centered content.
func (tsm *ThemeSelectorModal) Draw(screen tcell.Screen) {
	// Don't call Box.DrawForSubclass - that would fill the entire screen as backdrop
	// Only draw the centered modal box

	// Get theme colors
	bg := Bg()
	fg := Fg()
	borderColor := PanelBorder()
	titleColor := PanelTitle()

	// Calculate modal dimensions
	screenWidth, screenHeight := screen.Size()
	modalWidth := 42
	modalHeight := len(tsm.themes) + 4
	if modalHeight < 10 {
		modalHeight = 10
	}
	if modalHeight > 22 {
		modalHeight = 22
	}

	// Center the modal
	x := (screenWidth - modalWidth) / 2
	y := (screenHeight - modalHeight) / 2

	// Draw modal background only (no full-screen backdrop)
	bgStyle := tcell.StyleDefault.Background(bg)
	for row := y; row < y+modalHeight; row++ {
		for col := x; col < x+modalWidth; col++ {
			screen.SetContent(col, row, ' ', nil, bgStyle)
		}
	}

	// Draw border
	borderStyle := tcell.StyleDefault.Background(bg).Foreground(borderColor)
	titleStyle := tcell.StyleDefault.Background(bg).Foreground(titleColor)

	// Corners
	screen.SetContent(x, y, '╭', nil, borderStyle)
	screen.SetContent(x+modalWidth-1, y, '╮', nil, borderStyle)
	screen.SetContent(x, y+modalHeight-1, '╰', nil, borderStyle)
	screen.SetContent(x+modalWidth-1, y+modalHeight-1, '╯', nil, borderStyle)

	// Horizontal borders
	for i := x + 1; i < x+modalWidth-1; i++ {
		screen.SetContent(i, y, '─', nil, borderStyle)
		screen.SetContent(i, y+modalHeight-1, '─', nil, borderStyle)
	}

	// Vertical borders
	for i := y + 1; i < y+modalHeight-1; i++ {
		screen.SetContent(x, i, '│', nil, borderStyle)
		screen.SetContent(x+modalWidth-1, i, '│', nil, borderStyle)
	}

	// Title
	title := " Select Theme "
	titleRunes := []rune(title)
	titleStart := x + (modalWidth-len(titleRunes))/2
	for i, r := range titleRunes {
		screen.SetContent(titleStart+i, y, r, nil, titleStyle)
	}

	// Hints at bottom (left-aligned)
	hints := "j/k:Navigate  Enter:Select  Esc:Cancel"
	hintsStyle := tcell.StyleDefault.Background(bg).Foreground(FgDim())
	hintsStart := x + 2
	for i, r := range hints {
		if hintsStart+i < x+modalWidth-1 {
			screen.SetContent(hintsStart+i, y+modalHeight-2, r, nil, hintsStyle)
		}
	}

	// Update table colors
	tsm.table.SetBackgroundColor(bg)
	tsm.table.SetSelectedStyle(SelectionStyle())

	for row := 0; row < tsm.table.GetRowCount(); row++ {
		if cell := tsm.table.GetCell(row, 0); cell != nil {
			cell.SetTextColor(fg)
			cell.SetBackgroundColor(bg)
		}
	}

	// Draw table inside the border
	tableX := x + 2
	tableY := y + 1
	tableWidth := modalWidth - 4
	tableHeight := modalHeight - 4

	tsm.table.SetRect(tableX, tableY, tableWidth, tableHeight)
	tsm.table.Draw(screen)
}

// Focus delegates focus to the table.
func (tsm *ThemeSelectorModal) Focus(delegate func(p tview.Primitive)) {
	delegate(tsm.table)
}

// HasFocus returns whether the table has focus.
func (tsm *ThemeSelectorModal) HasFocus() bool {
	return tsm.table.HasFocus()
}

// InputHandler handles keyboard input.
func (tsm *ThemeSelectorModal) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return tsm.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyEscape:
			if tsm.onCancel != nil {
				tsm.onCancel()
			}
			return
		case tcell.KeyEnter:
			if tsm.onSelect != nil {
				selected := tsm.GetSelectedTheme()
				if selected != "" {
					tsm.onSelect(selected)
				}
			}
			return
		case tcell.KeyDown:
			if tsm.currentIdx < len(tsm.themes)-1 {
				tsm.currentIdx++
				tsm.table.Select(tsm.currentIdx, 0)
			}
			return
		case tcell.KeyUp:
			if tsm.currentIdx > 0 {
				tsm.currentIdx--
				tsm.table.Select(tsm.currentIdx, 0)
			}
			return
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				if tsm.currentIdx < len(tsm.themes)-1 {
					tsm.currentIdx++
					tsm.table.Select(tsm.currentIdx, 0)
				}
				return
			case 'k':
				if tsm.currentIdx > 0 {
					tsm.currentIdx--
					tsm.table.Select(tsm.currentIdx, 0)
				}
				return
			case 'q':
				if tsm.onCancel != nil {
					tsm.onCancel()
				}
				return
			}
		}

		// Delegate to table for other input
		if handler := tsm.table.InputHandler(); handler != nil {
			handler(event, setFocus)
		}
	})
}

// MouseHandler delegates mouse events.
func (tsm *ThemeSelectorModal) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return tsm.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		if handler := tsm.table.MouseHandler(); handler != nil {
			return handler(action, event, setFocus)
		}
		return false, nil
	})
}
