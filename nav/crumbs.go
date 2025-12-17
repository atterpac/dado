package nav

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/jig/theme"
)

// Crumbs displays a breadcrumb navigation trail.
type Crumbs struct {
	*tview.TextView
	path      []string
	separator string
}

// NewCrumbs creates a new breadcrumb component.
func NewCrumbs() *Crumbs {
	tv := tview.NewTextView()
	tv.SetBackgroundColor(theme.Bg())

	c := &Crumbs{
		TextView:  tv,
		path:      make([]string, 0),
		separator: " > ",
	}

	c.TextView.SetDynamicColors(true)
	c.TextView.SetTextAlign(tview.AlignLeft)

	// Register for automatic theme updates
	theme.Register(tv)

	return c
}

// SetPath sets the full breadcrumb path.
func (c *Crumbs) SetPath(path []string) *Crumbs {
	c.path = path
	c.refresh()
	return c
}

// Push adds a crumb to the path.
func (c *Crumbs) Push(crumb string) *Crumbs {
	c.path = append(c.path, crumb)
	c.refresh()
	return c
}

// Pop removes the last crumb.
func (c *Crumbs) Pop() *Crumbs {
	if len(c.path) > 0 {
		c.path = c.path[:len(c.path)-1]
		c.refresh()
	}
	return c
}

// Clear removes all crumbs.
func (c *Crumbs) Clear() *Crumbs {
	c.path = make([]string, 0)
	c.refresh()
	return c
}

// SetSeparator sets the separator string (default: " > ").
func (c *Crumbs) SetSeparator(sep string) *Crumbs {
	c.separator = sep
	c.refresh()
	return c
}

// GetPath returns the current path.
func (c *Crumbs) GetPath() []string {
	result := make([]string, len(c.path))
	copy(result, c.path)
	return result
}

// refresh rebuilds the display text with colors.
func (c *Crumbs) refresh() {
	if len(c.path) == 0 {
		c.TextView.SetText("")
		return
	}

	fgDim := theme.TagFgDim()
	fgMuted := theme.TagFgMuted()
	accent := theme.TagAccent()

	var parts []string
	for i, crumb := range c.path {
		if i == len(c.path)-1 {
			// Last item: accent color
			parts = append(parts, "["+accent+"]"+crumb+"[-]")
		} else {
			// Previous items: dim color
			parts = append(parts, "["+fgDim+"]"+crumb+"[-]")
		}
	}

	// Join with styled separator
	sepStyled := "[" + fgMuted + "]" + c.separator + "[-]"
	c.TextView.SetText(strings.Join(parts, sepStyled))
}

// Draw renders the breadcrumbs with current theme colors.
func (c *Crumbs) Draw(screen tcell.Screen) {
	// Refresh colors before drawing (theme may have changed)
	c.refresh()
	c.TextView.SetBackgroundColor(theme.Bg())
	c.TextView.Draw(screen)
}

// GetPreferredHeight returns the preferred height for the crumbs.
func (c *Crumbs) GetPreferredHeight() int {
	return 1
}
