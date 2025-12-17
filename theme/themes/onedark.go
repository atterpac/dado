package themes

import "github.com/gdamore/tcell/v2"

// OneDark - Atom One Dark inspired theme
var OneDark = &oneDark{}

type oneDark struct{}

func (o *oneDark) Bg() tcell.Color          { return tcell.NewHexColor(0x282c34) }
func (o *oneDark) BgLight() tcell.Color     { return tcell.NewHexColor(0x3e4451) }
func (o *oneDark) BgDark() tcell.Color      { return tcell.NewHexColor(0x21252b) }
func (o *oneDark) Fg() tcell.Color          { return tcell.NewHexColor(0xabb2bf) }
func (o *oneDark) FgDim() tcell.Color       { return tcell.NewHexColor(0x5c6370) }
func (o *oneDark) FgMuted() tcell.Color     { return tcell.NewHexColor(0x4b5263) }
func (o *oneDark) Accent() tcell.Color      { return tcell.NewHexColor(0x61afef) }
func (o *oneDark) AccentDim() tcell.Color   { return tcell.NewHexColor(0xc678dd) }
func (o *oneDark) Highlight() tcell.Color   { return tcell.NewHexColor(0x3e4451) }
func (o *oneDark) Success() tcell.Color     { return tcell.NewHexColor(0x98c379) }
func (o *oneDark) Warning() tcell.Color     { return tcell.NewHexColor(0xe5c07b) }
func (o *oneDark) Error() tcell.Color       { return tcell.NewHexColor(0xe06c75) }
func (o *oneDark) Info() tcell.Color        { return tcell.NewHexColor(0x56b6c2) }
func (o *oneDark) Border() tcell.Color      { return tcell.NewHexColor(0x3e4451) }
func (o *oneDark) BorderFocus() tcell.Color { return tcell.NewHexColor(0x61afef) }
func (o *oneDark) Header() tcell.Color      { return tcell.NewHexColor(0x21252b) }
func (o *oneDark) Menu() tcell.Color        { return tcell.NewHexColor(0x282c34) }
func (o *oneDark) TableHeader() tcell.Color { return tcell.NewHexColor(0x61afef) }
func (o *oneDark) Key() tcell.Color         { return tcell.NewHexColor(0xc678dd) }
func (o *oneDark) Crumb() tcell.Color       { return tcell.NewHexColor(0x61afef) }
func (o *oneDark) PanelBorder() tcell.Color { return tcell.NewHexColor(0x3e4451) }
func (o *oneDark) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x61afef) }

// OneLight - Atom One Light inspired theme
var OneLight = &oneLight{}

type oneLight struct{}

func (o *oneLight) Bg() tcell.Color          { return tcell.NewHexColor(0xfafafa) }
func (o *oneLight) BgLight() tcell.Color     { return tcell.NewHexColor(0xeaeaeb) }
func (o *oneLight) BgDark() tcell.Color      { return tcell.NewHexColor(0xf0f0f0) }
func (o *oneLight) Fg() tcell.Color          { return tcell.NewHexColor(0x383a42) }
func (o *oneLight) FgDim() tcell.Color       { return tcell.NewHexColor(0xa0a1a7) }
func (o *oneLight) FgMuted() tcell.Color     { return tcell.NewHexColor(0xc8c9cb) }
func (o *oneLight) Accent() tcell.Color      { return tcell.NewHexColor(0x4078f2) }
func (o *oneLight) AccentDim() tcell.Color   { return tcell.NewHexColor(0xa626a4) }
func (o *oneLight) Highlight() tcell.Color   { return tcell.NewHexColor(0xe5e5e6) }
func (o *oneLight) Success() tcell.Color     { return tcell.NewHexColor(0x50a14f) }
func (o *oneLight) Warning() tcell.Color     { return tcell.NewHexColor(0xc18401) }
func (o *oneLight) Error() tcell.Color       { return tcell.NewHexColor(0xe45649) }
func (o *oneLight) Info() tcell.Color        { return tcell.NewHexColor(0x0184bc) }
func (o *oneLight) Border() tcell.Color      { return tcell.NewHexColor(0xd3d3d4) }
func (o *oneLight) BorderFocus() tcell.Color { return tcell.NewHexColor(0x4078f2) }
func (o *oneLight) Header() tcell.Color      { return tcell.NewHexColor(0xf0f0f0) }
func (o *oneLight) Menu() tcell.Color        { return tcell.NewHexColor(0xfafafa) }
func (o *oneLight) TableHeader() tcell.Color { return tcell.NewHexColor(0x4078f2) }
func (o *oneLight) Key() tcell.Color         { return tcell.NewHexColor(0xa626a4) }
func (o *oneLight) Crumb() tcell.Color       { return tcell.NewHexColor(0x4078f2) }
func (o *oneLight) PanelBorder() tcell.Color { return tcell.NewHexColor(0xc8c8c8) }
func (o *oneLight) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x4078f2) }
