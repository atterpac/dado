package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/dado/theme"
)

// EmptyState displays a centered empty/loading/error state.
type EmptyState struct {
	*tview.Flex
	iconView    *tview.TextView
	titleView   *tview.TextView
	messageView *tview.TextView
	icon        string
	title       string
	message     string
	hasFocus    bool
}

// NewEmptyState creates a new empty state display.
func NewEmptyState() *EmptyState {
	e := &EmptyState{
		Flex:        tview.NewFlex(),
		iconView:    tview.NewTextView(),
		titleView:   tview.NewTextView(),
		messageView: tview.NewTextView(),
	}

	// Configure text views
	e.iconView.SetTextAlign(tview.AlignCenter).SetDynamicColors(true)
	e.titleView.SetTextAlign(tview.AlignCenter).SetDynamicColors(true)
	e.messageView.SetTextAlign(tview.AlignCenter).SetDynamicColors(true)

	e.setupLayout()
	return e
}

// setupLayout builds the centered layout structure.
func (e *EmptyState) setupLayout() {
	// Content column
	content := tview.NewFlex().SetDirection(tview.FlexRow)
	content.AddItem(e.iconView, 2, 0, false)
	content.AddItem(e.titleView, 1, 0, false)
	content.AddItem(e.messageView, 1, 0, false)

	// Center horizontally — use proportional sizing so long messages aren't truncated
	hCenter := tview.NewFlex().SetDirection(tview.FlexColumn)
	hCenter.AddItem(nil, 0, 1, false)
	hCenter.AddItem(content, 0, 2, false)
	hCenter.AddItem(nil, 0, 1, false)

	// Center vertically
	e.Flex.SetDirection(tview.FlexRow)
	e.Flex.AddItem(nil, 0, 1, false)
	e.Flex.AddItem(hCenter, 5, 0, false)
	e.Flex.AddItem(nil, 0, 1, false)
}

// SetIcon sets the icon (Nerd Font glyph).
func (e *EmptyState) SetIcon(icon string) *EmptyState {
	e.icon = icon
	return e
}

// SetTitle sets the main title text.
func (e *EmptyState) SetTitle(title string) *EmptyState {
	e.title = title
	return e
}

// SetMessage sets the secondary message text.
func (e *EmptyState) SetMessage(message string) *EmptyState {
	e.message = message
	return e
}

// Configure sets all fields at once.
func (e *EmptyState) Configure(icon, title, message string) *EmptyState {
	e.icon = icon
	e.title = title
	e.message = message
	return e
}

// Draw renders the empty state with theme colors.
func (e *EmptyState) Draw(screen tcell.Screen) {
	// Apply theme colors at draw time
	accentHex := theme.TagAccent()
	fgHex := theme.TagFg()
	fgDimHex := theme.TagFgDim()
	bgColor := theme.Bg()

	// Update text views with colored content
	e.iconView.SetText("[" + accentHex + "]" + e.icon + "[-]")
	e.iconView.SetBackgroundColor(bgColor)

	e.titleView.SetText("[" + fgHex + "]" + e.title + "[-]")
	e.titleView.SetBackgroundColor(bgColor)

	e.messageView.SetText("[" + fgDimHex + "]" + e.message + "[-]")
	e.messageView.SetBackgroundColor(bgColor)

	e.Flex.Draw(screen)
}

// Focus is called when the empty state receives focus.
func (e *EmptyState) Focus(delegate func(p tview.Primitive)) {
	e.hasFocus = true
	// Don't delegate to children - EmptyState handles focus itself
}

// Blur is called when the empty state loses focus.
func (e *EmptyState) Blur() {
	e.hasFocus = false
}

// HasFocus returns true if the empty state currently has focus.
func (e *EmptyState) HasFocus() bool {
	return e.hasFocus
}
