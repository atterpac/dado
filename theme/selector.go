package theme

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ThemeSelectorModal displays a modal for selecting themes with live preview.
type ThemeSelectorModal struct {
	*tview.Flex
	table         *tview.Table
	themes        []string
	originalTheme string
	currentIdx    int
	onSelect      func(name string)
	onCancel      func()
	onPreview     func(name string) // Called when navigating for live preview
	// Internal components for theme updates
	innerFlex *tview.Flex
	centerRow *tview.Flex
	title     *tview.TextView
	hints     *tview.TextView
}

// NewThemeSelectorModal creates a new theme selector modal.
// The themes parameter should be the list of available theme names (e.g., from themes.Names()).
// The currentTheme parameter is the name of the currently active theme.
func NewThemeSelectorModal(themes []string, currentTheme string) *ThemeSelectorModal {
	flex := tview.NewFlex()
	flex.SetBackgroundColor(Bg())
	table := tview.NewTable()
	table.SetBackgroundColor(Bg())

	tsm := &ThemeSelectorModal{
		Flex:          flex,
		table:         table,
		themes:        themes,
		originalTheme: currentTheme,
	}

	// Register for automatic theme updates
	Register(flex)
	Register(table)

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
// This is called whenever the selection changes (j/k navigation).
func (tsm *ThemeSelectorModal) SetOnPreview(fn func(name string)) *ThemeSelectorModal {
	tsm.onPreview = fn
	return tsm
}

func (tsm *ThemeSelectorModal) setup() {
	// Configure table
	tsm.table.SetBackgroundColor(tcell.ColorDefault)
	tsm.table.SetSelectable(true, false)
	tsm.table.SetSelectedStyle(tcell.StyleDefault.
		Foreground(Fg()).
		Background(Highlight()).
		Bold(true))

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
			// Trigger live preview
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

	// Input capture for navigation
	tsm.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			tsm.restoreAndCancel()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				tsm.moveDown()
				return nil
			case 'k':
				tsm.moveUp()
				return nil
			case 'q':
				tsm.restoreAndCancel()
				return nil
			}
		}
		return event
	})

	// Build modal layout
	tsm.buildLayout()
}

func (tsm *ThemeSelectorModal) buildLayout() {
	// Calculate dimensions
	height := len(tsm.themes) + 4 // +4 for border and hints
	if height < 10 {
		height = 10
	}
	if height > 20 {
		height = 20
	}
	width := 40

	bg := Bg()

	// Title bar
	tsm.title = tview.NewTextView().
		SetText(" Select Theme ").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	tsm.title.SetBackgroundColor(bg)

	// Hints bar
	tsm.hints = tview.NewTextView().
		SetText(" j/k:Navigate  Enter:Select  Esc:Cancel ").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	tsm.hints.SetBackgroundColor(bg)

	// Inner content
	tsm.innerFlex = tview.NewFlex().SetDirection(tview.FlexRow)
	tsm.innerFlex.SetBackgroundColor(bg)
	tsm.innerFlex.AddItem(tsm.title, 1, 0, false)
	tsm.innerFlex.AddItem(tsm.table, 0, 1, true)
	tsm.innerFlex.AddItem(tsm.hints, 1, 0, false)

	// Center the modal
	tsm.Flex.SetDirection(tview.FlexRow)
	tsm.Flex.AddItem(nil, 0, 1, false) // Top spacer

	tsm.centerRow = tview.NewFlex().SetDirection(tview.FlexColumn)
	tsm.centerRow.SetBackgroundColor(bg)
	tsm.centerRow.AddItem(nil, 0, 1, false)          // Left spacer
	tsm.centerRow.AddItem(tsm.innerFlex, width, 0, true)
	tsm.centerRow.AddItem(nil, 0, 1, false)          // Right spacer

	tsm.Flex.AddItem(tsm.centerRow, height, 0, true)
	tsm.Flex.AddItem(nil, 0, 1, false) // Bottom spacer

	// Register inner components for automatic theme updates
	Register(tsm.title)
	Register(tsm.hints)
	Register(tsm.innerFlex)
	Register(tsm.centerRow)
}

func (tsm *ThemeSelectorModal) rebuildTable() {
	tsm.table.Clear()

	for i, name := range tsm.themes {
		marker := "  "
		if name == tsm.originalTheme {
			marker = "✓ "
		}
		cell := tview.NewTableCell(marker + name).
			SetTextColor(tcell.ColorDefault).
			SetBackgroundColor(tcell.ColorDefault)
		tsm.table.SetCell(i, 0, cell)
	}
}

func (tsm *ThemeSelectorModal) moveDown() {
	if tsm.currentIdx < len(tsm.themes)-1 {
		tsm.currentIdx++
		tsm.table.Select(tsm.currentIdx, 0)
	}
}

func (tsm *ThemeSelectorModal) moveUp() {
	if tsm.currentIdx > 0 {
		tsm.currentIdx--
		tsm.table.Select(tsm.currentIdx, 0)
	}
}

func (tsm *ThemeSelectorModal) restoreAndCancel() {
	if tsm.onCancel != nil {
		tsm.onCancel()
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

// Draw updates colors dynamically and draws the modal.
func (tsm *ThemeSelectorModal) Draw(screen tcell.Screen) {
	// Update all backgrounds from theme
	bg := Bg()
	fg := Fg()
	accent := Accent()

	// Update all flex containers
	tsm.Flex.SetBackgroundColor(bg)
	tsm.centerRow.SetBackgroundColor(bg)
	tsm.innerFlex.SetBackgroundColor(bg)

	// Update title and hints
	tsm.title.SetBackgroundColor(bg)
	tsm.title.SetTextColor(accent)
	tsm.hints.SetBackgroundColor(bg)
	tsm.hints.SetTextColor(FgDim())

	// Update table
	tsm.table.SetBackgroundColor(bg)
	tsm.table.SetSelectedStyle(tcell.StyleDefault.
		Foreground(fg).
		Background(Highlight()).
		Bold(true))

	// Update cell colors
	for row := 0; row < tsm.table.GetRowCount(); row++ {
		if cell := tsm.table.GetCell(row, 0); cell != nil {
			cell.SetTextColor(fg)
			cell.SetBackgroundColor(bg)
		}
	}

	tsm.Flex.Draw(screen)
}


// Focus delegates focus to the table.
func (tsm *ThemeSelectorModal) Focus(delegate func(p tview.Primitive)) {
	delegate(tsm.table)
}

// HasFocus returns whether the table has focus.
func (tsm *ThemeSelectorModal) HasFocus() bool {
	return tsm.table.HasFocus()
}
