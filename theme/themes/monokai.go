package themes

import "github.com/gdamore/tcell/v2"

// Monokai - Monokai Pro inspired theme
var Monokai = &monokai{}

type monokai struct{}

func (m *monokai) Bg() tcell.Color          { return tcell.NewHexColor(0x2d2a2e) }
func (m *monokai) BgLight() tcell.Color     { return tcell.NewHexColor(0x403e41) }
func (m *monokai) BgDark() tcell.Color      { return tcell.NewHexColor(0x221f22) }
func (m *monokai) Fg() tcell.Color          { return tcell.NewHexColor(0xfcfcfa) }
func (m *monokai) FgDim() tcell.Color       { return tcell.NewHexColor(0x727072) }
func (m *monokai) FgMuted() tcell.Color     { return tcell.NewHexColor(0x5b595c) }
func (m *monokai) Accent() tcell.Color      { return tcell.NewHexColor(0x78dce8) }
func (m *monokai) AccentDim() tcell.Color   { return tcell.NewHexColor(0xab9df2) }
func (m *monokai) Highlight() tcell.Color   { return tcell.NewHexColor(0x5b595c) }
func (m *monokai) Success() tcell.Color     { return tcell.NewHexColor(0xa9dc76) }
func (m *monokai) Warning() tcell.Color     { return tcell.NewHexColor(0xffd866) }
func (m *monokai) Error() tcell.Color       { return tcell.NewHexColor(0xff6188) }
func (m *monokai) Info() tcell.Color        { return tcell.NewHexColor(0x78dce8) }
func (m *monokai) Border() tcell.Color      { return tcell.NewHexColor(0x5b595c) }
func (m *monokai) BorderFocus() tcell.Color { return tcell.NewHexColor(0x78dce8) }
func (m *monokai) Header() tcell.Color      { return tcell.NewHexColor(0x221f22) }
func (m *monokai) Menu() tcell.Color        { return tcell.NewHexColor(0x2d2a2e) }
func (m *monokai) TableHeader() tcell.Color { return tcell.NewHexColor(0x78dce8) }
func (m *monokai) Key() tcell.Color         { return tcell.NewHexColor(0xab9df2) }
func (m *monokai) Crumb() tcell.Color       { return tcell.NewHexColor(0x78dce8) }
func (m *monokai) PanelBorder() tcell.Color { return tcell.NewHexColor(0x5b595c) }
func (m *monokai) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x78dce8) }
