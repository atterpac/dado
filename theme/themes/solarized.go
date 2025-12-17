package themes

import "github.com/gdamore/tcell/v2"

// SolarizedDark - Precision colors for machines and people (dark)
var SolarizedDark = &solarizedDark{}

type solarizedDark struct{}

func (s *solarizedDark) Bg() tcell.Color          { return tcell.NewHexColor(0x002b36) }
func (s *solarizedDark) BgLight() tcell.Color     { return tcell.NewHexColor(0x073642) }
func (s *solarizedDark) BgDark() tcell.Color      { return tcell.NewHexColor(0x001e26) }
func (s *solarizedDark) Fg() tcell.Color          { return tcell.NewHexColor(0x839496) }
func (s *solarizedDark) FgDim() tcell.Color       { return tcell.NewHexColor(0x586e75) }
func (s *solarizedDark) FgMuted() tcell.Color     { return tcell.NewHexColor(0x073642) }
func (s *solarizedDark) Accent() tcell.Color      { return tcell.NewHexColor(0x268bd2) }
func (s *solarizedDark) AccentDim() tcell.Color   { return tcell.NewHexColor(0x6c71c4) }
func (s *solarizedDark) Highlight() tcell.Color   { return tcell.NewHexColor(0x073642) }
func (s *solarizedDark) Success() tcell.Color     { return tcell.NewHexColor(0x859900) }
func (s *solarizedDark) Warning() tcell.Color     { return tcell.NewHexColor(0xb58900) }
func (s *solarizedDark) Error() tcell.Color       { return tcell.NewHexColor(0xdc322f) }
func (s *solarizedDark) Info() tcell.Color        { return tcell.NewHexColor(0x2aa198) }
func (s *solarizedDark) Border() tcell.Color      { return tcell.NewHexColor(0x073642) }
func (s *solarizedDark) BorderFocus() tcell.Color { return tcell.NewHexColor(0x268bd2) }
func (s *solarizedDark) Header() tcell.Color      { return tcell.NewHexColor(0x001e26) }
func (s *solarizedDark) Menu() tcell.Color        { return tcell.NewHexColor(0x002b36) }
func (s *solarizedDark) TableHeader() tcell.Color { return tcell.NewHexColor(0x268bd2) }
func (s *solarizedDark) Key() tcell.Color         { return tcell.NewHexColor(0x6c71c4) }
func (s *solarizedDark) Crumb() tcell.Color       { return tcell.NewHexColor(0x268bd2) }
func (s *solarizedDark) PanelBorder() tcell.Color { return tcell.NewHexColor(0x586e75) }
func (s *solarizedDark) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x268bd2) }

// SolarizedLight - Precision colors for machines and people (light)
var SolarizedLight = &solarizedLight{}

type solarizedLight struct{}

func (s *solarizedLight) Bg() tcell.Color          { return tcell.NewHexColor(0xfdf6e3) }
func (s *solarizedLight) BgLight() tcell.Color     { return tcell.NewHexColor(0xeee8d5) }
func (s *solarizedLight) BgDark() tcell.Color      { return tcell.NewHexColor(0xf5efdc) }
func (s *solarizedLight) Fg() tcell.Color          { return tcell.NewHexColor(0x657b83) }
func (s *solarizedLight) FgDim() tcell.Color       { return tcell.NewHexColor(0x93a1a1) }
func (s *solarizedLight) FgMuted() tcell.Color     { return tcell.NewHexColor(0xc9c9bb) }
func (s *solarizedLight) Accent() tcell.Color      { return tcell.NewHexColor(0x268bd2) }
func (s *solarizedLight) AccentDim() tcell.Color   { return tcell.NewHexColor(0x6c71c4) }
func (s *solarizedLight) Highlight() tcell.Color   { return tcell.NewHexColor(0xeee8d5) }
func (s *solarizedLight) Success() tcell.Color     { return tcell.NewHexColor(0x859900) }
func (s *solarizedLight) Warning() tcell.Color     { return tcell.NewHexColor(0xb58900) }
func (s *solarizedLight) Error() tcell.Color       { return tcell.NewHexColor(0xdc322f) }
func (s *solarizedLight) Info() tcell.Color        { return tcell.NewHexColor(0x2aa198) }
func (s *solarizedLight) Border() tcell.Color      { return tcell.NewHexColor(0xeee8d5) }
func (s *solarizedLight) BorderFocus() tcell.Color { return tcell.NewHexColor(0x268bd2) }
func (s *solarizedLight) Header() tcell.Color      { return tcell.NewHexColor(0xf5efdc) }
func (s *solarizedLight) Menu() tcell.Color        { return tcell.NewHexColor(0xfdf6e3) }
func (s *solarizedLight) TableHeader() tcell.Color { return tcell.NewHexColor(0x268bd2) }
func (s *solarizedLight) Key() tcell.Color         { return tcell.NewHexColor(0x6c71c4) }
func (s *solarizedLight) Crumb() tcell.Color       { return tcell.NewHexColor(0x268bd2) }
func (s *solarizedLight) PanelBorder() tcell.Color { return tcell.NewHexColor(0x93a1a1) }
func (s *solarizedLight) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x268bd2) }
