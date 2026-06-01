package core

import "github.com/gdamore/tcell/v2"

type page struct {
	name   string
	widget Widget
}

// Pages is a container that shows one child Widget at a time.
// It tracks a named list of pages and a "front" (current) page.
type Pages struct {
	Box
	pages   []page
	current int // index into pages; -1 = empty
}

// NewPages returns an empty Pages container.
func NewPages() *Pages { return &Pages{current: -1} }

// AddPage adds a named page. If show is true the page becomes the front.
// The resize parameter is accepted for API compatibility but is ignored
// (all pages resize to the container rect on Draw).
func (p *Pages) AddPage(name string, w Widget, _ bool, show bool) {
	p.pages = append(p.pages, page{name, w})
	if show || p.current == -1 {
		p.current = len(p.pages) - 1
	}
}

// RemovePage removes the named page. If the current page is removed, the
// previous page becomes current (or -1 if no pages remain).
func (p *Pages) RemovePage(name string) {
	for i, pg := range p.pages {
		if pg.name == name {
			p.pages = append(p.pages[:i], p.pages[i+1:]...)
			if p.current >= len(p.pages) {
				p.current = len(p.pages) - 1
			}
			return
		}
	}
}

// ShowPage makes the named page the front.
func (p *Pages) ShowPage(name string) {
	for i, pg := range p.pages {
		if pg.name == name {
			p.current = i
			return
		}
	}
}

// GetFrontPage returns the name and widget of the current front page.
// Returns ("", nil) if no pages exist.
func (p *Pages) GetFrontPage() (string, Widget) {
	if p.current < 0 || p.current >= len(p.pages) {
		return "", nil
	}
	pg := p.pages[p.current]
	return pg.name, pg.widget
}

// GetPageNames returns all page names in insertion order.
func (p *Pages) GetPageNames() []string {
	names := make([]string, len(p.pages))
	for i, pg := range p.pages {
		names[i] = pg.name
	}
	return names
}

// Draw renders the front page to the full container rect.
func (p *Pages) Draw(screen tcell.Screen) {
	p.Box.Draw(screen)
	if p.current < 0 || p.current >= len(p.pages) {
		return
	}
	w := p.pages[p.current].widget
	x, y, pw, ph := p.Rect()
	w.SetRect(x, y, pw, ph)
	w.Draw(screen)
}

// HandleKey routes to the front page (implements KeyHandler).
func (p *Pages) HandleKey(ev *tcell.EventKey) bool {
	if p.current < 0 || p.current >= len(p.pages) {
		return false
	}
	w := p.pages[p.current].widget
	if kh, ok := w.(KeyHandler); ok {
		return kh.HandleKey(ev)
	}
	return false
}

// Children returns the front page as the only child (implements Container).
func (p *Pages) Children() []Widget {
	if p.current < 0 || p.current >= len(p.pages) {
		return nil
	}
	return []Widget{p.pages[p.current].widget}
}

// DescendantsAt returns descendants of the front page at (x, y).
func (p *Pages) DescendantsAt(x, y int) []Widget {
	if p.current < 0 {
		return nil
	}
	w := p.pages[p.current].widget
	wx, wy, ww, wh := w.Rect()
	if x < wx || x >= wx+ww || y < wy || y >= wy+wh {
		return nil
	}
	if c, ok := w.(Container); ok {
		return append(c.DescendantsAt(x, y), w)
	}
	return []Widget{w}
}
