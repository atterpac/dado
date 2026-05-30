package theme

import "github.com/gdamore/tcell/v2"

// Color accessors on Provider. Each reads the provider's active theme at call
// time (for live switching) and falls back to a sane tcell color when no theme
// is set. The package-level convenience funcs in colors.go forward here against
// the default provider; scoped providers (ComponentBase/widget th()) read their
// own theme through the same methods, which is what makes per-subtree theming
// take effect at draw time.

func (p *Provider) colorOr(pick func(Theme) tcell.Color, fallback tcell.Color) tcell.Color {
	if t := p.Theme(); t != nil {
		return pick(t)
	}
	return fallback
}

// Base colors

func (p *Provider) Bg() tcell.Color      { return p.colorOr(Theme.Bg, tcell.ColorDefault) }
func (p *Provider) BgLight() tcell.Color { return p.colorOr(Theme.BgLight, tcell.ColorDefault) }
func (p *Provider) BgDark() tcell.Color  { return p.colorOr(Theme.BgDark, tcell.ColorDefault) }
func (p *Provider) Fg() tcell.Color      { return p.colorOr(Theme.Fg, tcell.ColorDefault) }
func (p *Provider) FgDim() tcell.Color   { return p.colorOr(Theme.FgDim, tcell.ColorGray) }
func (p *Provider) FgMuted() tcell.Color { return p.colorOr(Theme.FgMuted, tcell.ColorDarkGray) }

// Accent colors

func (p *Provider) Accent() tcell.Color    { return p.colorOr(Theme.Accent, tcell.ColorBlue) }
func (p *Provider) AccentDim() tcell.Color { return p.colorOr(Theme.AccentDim, tcell.ColorBlue) }
func (p *Provider) Highlight() tcell.Color { return p.colorOr(Theme.Highlight, tcell.ColorYellow) }

// Semantic colors

func (p *Provider) Success() tcell.Color { return p.colorOr(Theme.Success, tcell.ColorGreen) }
func (p *Provider) Warning() tcell.Color { return p.colorOr(Theme.Warning, tcell.ColorYellow) }
func (p *Provider) Error() tcell.Color   { return p.colorOr(Theme.Error, tcell.ColorRed) }
func (p *Provider) Info() tcell.Color    { return p.colorOr(Theme.Info, tcell.ColorBlue) }

// Border colors

func (p *Provider) Border() tcell.Color      { return p.colorOr(Theme.Border, tcell.ColorGray) }
func (p *Provider) BorderFocus() tcell.Color { return p.colorOr(Theme.BorderFocus, tcell.ColorWhite) }

// UI element colors

func (p *Provider) Header() tcell.Color      { return p.colorOr(Theme.Header, tcell.ColorDefault) }
func (p *Provider) Menu() tcell.Color        { return p.colorOr(Theme.Menu, tcell.ColorDefault) }
func (p *Provider) TableHeader() tcell.Color { return p.colorOr(Theme.TableHeader, tcell.ColorBlue) }
func (p *Provider) Key() tcell.Color         { return p.colorOr(Theme.Key, tcell.ColorBlue) }
func (p *Provider) Crumb() tcell.Color       { return p.colorOr(Theme.Crumb, tcell.ColorBlue) }
func (p *Provider) PanelBorder() tcell.Color { return p.colorOr(Theme.PanelBorder, tcell.ColorGray) }
func (p *Provider) PanelTitle() tcell.Color  { return p.colorOr(Theme.PanelTitle, tcell.ColorBlue) }
