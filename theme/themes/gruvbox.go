package themes

import "github.com/gdamore/tcell/v2"

// GruvboxDark - Retro groove dark variant
var GruvboxDark = &gruvboxDark{}

type gruvboxDark struct{}

func (g *gruvboxDark) Bg() tcell.Color          { return tcell.NewHexColor(0x282828) }
func (g *gruvboxDark) BgLight() tcell.Color     { return tcell.NewHexColor(0x3c3836) }
func (g *gruvboxDark) BgDark() tcell.Color      { return tcell.NewHexColor(0x1d2021) }
func (g *gruvboxDark) Fg() tcell.Color          { return tcell.NewHexColor(0xebdbb2) }
func (g *gruvboxDark) FgDim() tcell.Color       { return tcell.NewHexColor(0x928374) }
func (g *gruvboxDark) FgMuted() tcell.Color     { return tcell.NewHexColor(0x665c54) }
func (g *gruvboxDark) Accent() tcell.Color      { return tcell.NewHexColor(0xfe8019) }
func (g *gruvboxDark) AccentDim() tcell.Color   { return tcell.NewHexColor(0xd3869b) }
func (g *gruvboxDark) Highlight() tcell.Color   { return tcell.NewHexColor(0x504945) }
func (g *gruvboxDark) Success() tcell.Color     { return tcell.NewHexColor(0xb8bb26) }
func (g *gruvboxDark) Warning() tcell.Color     { return tcell.NewHexColor(0xfabd2f) }
func (g *gruvboxDark) Error() tcell.Color       { return tcell.NewHexColor(0xfb4934) }
func (g *gruvboxDark) Info() tcell.Color        { return tcell.NewHexColor(0x83a598) }
func (g *gruvboxDark) Border() tcell.Color      { return tcell.NewHexColor(0x504945) }
func (g *gruvboxDark) BorderFocus() tcell.Color { return tcell.NewHexColor(0xfe8019) }
func (g *gruvboxDark) Header() tcell.Color      { return tcell.NewHexColor(0x1d2021) }
func (g *gruvboxDark) Menu() tcell.Color        { return tcell.NewHexColor(0x282828) }
func (g *gruvboxDark) TableHeader() tcell.Color { return tcell.NewHexColor(0xfe8019) }
func (g *gruvboxDark) Key() tcell.Color         { return tcell.NewHexColor(0xd3869b) }
func (g *gruvboxDark) Crumb() tcell.Color       { return tcell.NewHexColor(0xfe8019) }
func (g *gruvboxDark) PanelBorder() tcell.Color { return tcell.NewHexColor(0x504945) }
func (g *gruvboxDark) PanelTitle() tcell.Color  { return tcell.NewHexColor(0xfe8019) }

// GruvboxLight - Retro groove light variant
var GruvboxLight = &gruvboxLight{}

type gruvboxLight struct{}

func (g *gruvboxLight) Bg() tcell.Color          { return tcell.NewHexColor(0xfbf1c7) }
func (g *gruvboxLight) BgLight() tcell.Color     { return tcell.NewHexColor(0xebdbb2) }
func (g *gruvboxLight) BgDark() tcell.Color      { return tcell.NewHexColor(0xf2e5bc) }
func (g *gruvboxLight) Fg() tcell.Color          { return tcell.NewHexColor(0x3c3836) }
func (g *gruvboxLight) FgDim() tcell.Color       { return tcell.NewHexColor(0x928374) }
func (g *gruvboxLight) FgMuted() tcell.Color     { return tcell.NewHexColor(0xa89984) }
func (g *gruvboxLight) Accent() tcell.Color      { return tcell.NewHexColor(0xaf3a03) }
func (g *gruvboxLight) AccentDim() tcell.Color   { return tcell.NewHexColor(0x8f3f71) }
func (g *gruvboxLight) Highlight() tcell.Color   { return tcell.NewHexColor(0xd5c4a1) }
func (g *gruvboxLight) Success() tcell.Color     { return tcell.NewHexColor(0x79740e) }
func (g *gruvboxLight) Warning() tcell.Color     { return tcell.NewHexColor(0xb57614) }
func (g *gruvboxLight) Error() tcell.Color       { return tcell.NewHexColor(0x9d0006) }
func (g *gruvboxLight) Info() tcell.Color        { return tcell.NewHexColor(0x076678) }
func (g *gruvboxLight) Border() tcell.Color      { return tcell.NewHexColor(0xd5c4a1) }
func (g *gruvboxLight) BorderFocus() tcell.Color { return tcell.NewHexColor(0xaf3a03) }
func (g *gruvboxLight) Header() tcell.Color      { return tcell.NewHexColor(0xf2e5bc) }
func (g *gruvboxLight) Menu() tcell.Color        { return tcell.NewHexColor(0xfbf1c7) }
func (g *gruvboxLight) TableHeader() tcell.Color { return tcell.NewHexColor(0xaf3a03) }
func (g *gruvboxLight) Key() tcell.Color         { return tcell.NewHexColor(0x8f3f71) }
func (g *gruvboxLight) Crumb() tcell.Color       { return tcell.NewHexColor(0xaf3a03) }
func (g *gruvboxLight) PanelBorder() tcell.Color { return tcell.NewHexColor(0xbdae93) }
func (g *gruvboxLight) PanelTitle() tcell.Color  { return tcell.NewHexColor(0xaf3a03) }
