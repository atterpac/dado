package nav

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	// TODO: Update import path when extracted to separate repo
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

// Crumbs displays a breadcrumb navigation trail.
type Crumbs struct {
	core.TextView
	path      []string
	separator string
	subs      components.Subscriptions
}

// Subs returns the widget's subscription set; release on teardown.
func (c *Crumbs) Subs() *components.Subscriptions { return &c.subs }

// NewCrumbs creates a new breadcrumb component.
func NewCrumbs() *Crumbs {
	c := &Crumbs{
		path:      make([]string, 0),
		separator: " > ",
	}
	c.TextView.SetDynamicColors(true)
	c.TextView.Box.SetBackgroundColor(theme.Bg())
	c.subs.Add(theme.RegisterFn(func(col tcell.Color) {
		c.TextView.Box.SetBackgroundColor(col)
	}))
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
			parts = append(parts, "["+accent+"]"+crumb+"[-]")
		} else {
			parts = append(parts, "["+fgDim+"]"+crumb+"[-]")
		}
	}

	sepStyled := "[" + fgMuted + "]" + c.separator + "[-]"
	c.TextView.SetText(strings.Join(parts, sepStyled))
}

// Draw renders the breadcrumbs with current theme colors.
func (c *Crumbs) Draw(screen tcell.Screen) {
	c.refresh()
	c.TextView.Draw(screen)
}

// GetPreferredHeight returns the preferred height for the crumbs.
func (c *Crumbs) GetPreferredHeight() int {
	return 1
}
