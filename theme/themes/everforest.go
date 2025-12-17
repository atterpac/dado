package themes

import "github.com/gdamore/tcell/v2"

// EverforestDark - Comfortable and pleasant green forest theme (dark)
var EverforestDark = &everforestDark{}

type everforestDark struct{}

func (e *everforestDark) Bg() tcell.Color          { return tcell.NewHexColor(0x2d353b) }
func (e *everforestDark) BgLight() tcell.Color     { return tcell.NewHexColor(0x3d484d) }
func (e *everforestDark) BgDark() tcell.Color      { return tcell.NewHexColor(0x232a2e) }
func (e *everforestDark) Fg() tcell.Color          { return tcell.NewHexColor(0xd3c6aa) }
func (e *everforestDark) FgDim() tcell.Color       { return tcell.NewHexColor(0x859289) }
func (e *everforestDark) FgMuted() tcell.Color     { return tcell.NewHexColor(0x5d6b66) }
func (e *everforestDark) Accent() tcell.Color      { return tcell.NewHexColor(0xa7c080) }
func (e *everforestDark) AccentDim() tcell.Color   { return tcell.NewHexColor(0xd699b6) }
func (e *everforestDark) Highlight() tcell.Color   { return tcell.NewHexColor(0x475258) }
func (e *everforestDark) Success() tcell.Color     { return tcell.NewHexColor(0xa7c080) }
func (e *everforestDark) Warning() tcell.Color     { return tcell.NewHexColor(0xdbbc7f) }
func (e *everforestDark) Error() tcell.Color       { return tcell.NewHexColor(0xe67e80) }
func (e *everforestDark) Info() tcell.Color        { return tcell.NewHexColor(0x7fbbb3) }
func (e *everforestDark) Border() tcell.Color      { return tcell.NewHexColor(0x475258) }
func (e *everforestDark) BorderFocus() tcell.Color { return tcell.NewHexColor(0xa7c080) }
func (e *everforestDark) Header() tcell.Color      { return tcell.NewHexColor(0x232a2e) }
func (e *everforestDark) Menu() tcell.Color        { return tcell.NewHexColor(0x2d353b) }
func (e *everforestDark) TableHeader() tcell.Color { return tcell.NewHexColor(0xa7c080) }
func (e *everforestDark) Key() tcell.Color         { return tcell.NewHexColor(0xd699b6) }
func (e *everforestDark) Crumb() tcell.Color       { return tcell.NewHexColor(0xa7c080) }
func (e *everforestDark) PanelBorder() tcell.Color { return tcell.NewHexColor(0x475258) }
func (e *everforestDark) PanelTitle() tcell.Color  { return tcell.NewHexColor(0xa7c080) }

// EverforestLight - Comfortable and pleasant green forest theme (light)
var EverforestLight = &everforestLight{}

type everforestLight struct{}

func (e *everforestLight) Bg() tcell.Color          { return tcell.NewHexColor(0xfdf6e3) }
func (e *everforestLight) BgLight() tcell.Color     { return tcell.NewHexColor(0xefebd4) }
func (e *everforestLight) BgDark() tcell.Color      { return tcell.NewHexColor(0xf4f0d9) }
func (e *everforestLight) Fg() tcell.Color          { return tcell.NewHexColor(0x5c6a72) }
func (e *everforestLight) FgDim() tcell.Color       { return tcell.NewHexColor(0x939f91) }
func (e *everforestLight) FgMuted() tcell.Color     { return tcell.NewHexColor(0xa6b0a0) }
func (e *everforestLight) Accent() tcell.Color      { return tcell.NewHexColor(0x8da101) }
func (e *everforestLight) AccentDim() tcell.Color   { return tcell.NewHexColor(0xdf69ba) }
func (e *everforestLight) Highlight() tcell.Color   { return tcell.NewHexColor(0xe1ddc9) }
func (e *everforestLight) Success() tcell.Color     { return tcell.NewHexColor(0x8da101) }
func (e *everforestLight) Warning() tcell.Color     { return tcell.NewHexColor(0xdfa000) }
func (e *everforestLight) Error() tcell.Color       { return tcell.NewHexColor(0xf85552) }
func (e *everforestLight) Info() tcell.Color        { return tcell.NewHexColor(0x3a94c5) }
func (e *everforestLight) Border() tcell.Color      { return tcell.NewHexColor(0xe1ddc9) }
func (e *everforestLight) BorderFocus() tcell.Color { return tcell.NewHexColor(0x8da101) }
func (e *everforestLight) Header() tcell.Color      { return tcell.NewHexColor(0xf4f0d9) }
func (e *everforestLight) Menu() tcell.Color        { return tcell.NewHexColor(0xfdf6e3) }
func (e *everforestLight) TableHeader() tcell.Color { return tcell.NewHexColor(0x8da101) }
func (e *everforestLight) Key() tcell.Color         { return tcell.NewHexColor(0xdf69ba) }
func (e *everforestLight) Crumb() tcell.Color       { return tcell.NewHexColor(0x8da101) }
func (e *everforestLight) PanelBorder() tcell.Color { return tcell.NewHexColor(0xd8d3ba) }
func (e *everforestLight) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x8da101) }
