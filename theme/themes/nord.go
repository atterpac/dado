package themes

import "github.com/gdamore/tcell/v2"

// Nord is the Nord theme with arctic, bluish colors.
// Clean and elegant with a focus on readability.
var Nord = &nordTheme{}

type nordTheme struct{}

func (n *nordTheme) Bg() tcell.Color          { return tcell.NewHexColor(0x2e3440) }
func (n *nordTheme) BgLight() tcell.Color     { return tcell.NewHexColor(0x3b4252) }
func (n *nordTheme) BgDark() tcell.Color      { return tcell.NewHexColor(0x242933) }
func (n *nordTheme) Fg() tcell.Color          { return tcell.NewHexColor(0xd8dee9) }
func (n *nordTheme) FgDim() tcell.Color       { return tcell.NewHexColor(0x4c566a) }
func (n *nordTheme) FgMuted() tcell.Color     { return tcell.NewHexColor(0x3b4252) }
func (n *nordTheme) Accent() tcell.Color      { return tcell.NewHexColor(0x88c0d0) }
func (n *nordTheme) AccentDim() tcell.Color   { return tcell.NewHexColor(0x81a1c1) }
func (n *nordTheme) Highlight() tcell.Color   { return tcell.NewHexColor(0x434c5e) }
func (n *nordTheme) Success() tcell.Color     { return tcell.NewHexColor(0xa3be8c) }
func (n *nordTheme) Warning() tcell.Color     { return tcell.NewHexColor(0xebcb8b) }
func (n *nordTheme) Error() tcell.Color       { return tcell.NewHexColor(0xbf616a) }
func (n *nordTheme) Info() tcell.Color        { return tcell.NewHexColor(0x5e81ac) }
func (n *nordTheme) Border() tcell.Color      { return tcell.NewHexColor(0x4c566a) }
func (n *nordTheme) BorderFocus() tcell.Color { return tcell.NewHexColor(0x88c0d0) }
func (n *nordTheme) Header() tcell.Color      { return tcell.NewHexColor(0x242933) }
func (n *nordTheme) Menu() tcell.Color        { return tcell.NewHexColor(0x2e3440) }
func (n *nordTheme) TableHeader() tcell.Color { return tcell.NewHexColor(0x88c0d0) }
func (n *nordTheme) Key() tcell.Color         { return tcell.NewHexColor(0x81a1c1) }
func (n *nordTheme) Crumb() tcell.Color       { return tcell.NewHexColor(0x88c0d0) }
func (n *nordTheme) PanelBorder() tcell.Color { return tcell.NewHexColor(0x434c5e) }
func (n *nordTheme) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x88c0d0) }
