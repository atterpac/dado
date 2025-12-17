package themes

import "github.com/gdamore/tcell/v2"

// RosePine - All natural pine, faux fur and a bit of soho vibes
var RosePine = &rosePine{}

type rosePine struct{}

func (r *rosePine) Bg() tcell.Color          { return tcell.NewHexColor(0x191724) }
func (r *rosePine) BgLight() tcell.Color     { return tcell.NewHexColor(0x26233a) }
func (r *rosePine) BgDark() tcell.Color      { return tcell.NewHexColor(0x1f1d2e) }
func (r *rosePine) Fg() tcell.Color          { return tcell.NewHexColor(0xe0def4) }
func (r *rosePine) FgDim() tcell.Color       { return tcell.NewHexColor(0x6e6a86) }
func (r *rosePine) FgMuted() tcell.Color     { return tcell.NewHexColor(0x524f67) }
func (r *rosePine) Accent() tcell.Color      { return tcell.NewHexColor(0xebbcba) }
func (r *rosePine) AccentDim() tcell.Color   { return tcell.NewHexColor(0xc4a7e7) }
func (r *rosePine) Highlight() tcell.Color   { return tcell.NewHexColor(0x403d52) }
func (r *rosePine) Success() tcell.Color     { return tcell.NewHexColor(0x9ccfd8) }
func (r *rosePine) Warning() tcell.Color     { return tcell.NewHexColor(0xf6c177) }
func (r *rosePine) Error() tcell.Color       { return tcell.NewHexColor(0xeb6f92) }
func (r *rosePine) Info() tcell.Color        { return tcell.NewHexColor(0x31748f) }
func (r *rosePine) Border() tcell.Color      { return tcell.NewHexColor(0x26233a) }
func (r *rosePine) BorderFocus() tcell.Color { return tcell.NewHexColor(0xebbcba) }
func (r *rosePine) Header() tcell.Color      { return tcell.NewHexColor(0x1f1d2e) }
func (r *rosePine) Menu() tcell.Color        { return tcell.NewHexColor(0x191724) }
func (r *rosePine) TableHeader() tcell.Color { return tcell.NewHexColor(0xebbcba) }
func (r *rosePine) Key() tcell.Color         { return tcell.NewHexColor(0xc4a7e7) }
func (r *rosePine) Crumb() tcell.Color       { return tcell.NewHexColor(0xebbcba) }
func (r *rosePine) PanelBorder() tcell.Color { return tcell.NewHexColor(0x403d52) }
func (r *rosePine) PanelTitle() tcell.Color  { return tcell.NewHexColor(0xebbcba) }

// RosePineMoon - Rosé Pine variant with more contrast
var RosePineMoon = &rosePineMoon{}

type rosePineMoon struct{}

func (r *rosePineMoon) Bg() tcell.Color          { return tcell.NewHexColor(0x232136) }
func (r *rosePineMoon) BgLight() tcell.Color     { return tcell.NewHexColor(0x2a273f) }
func (r *rosePineMoon) BgDark() tcell.Color      { return tcell.NewHexColor(0x393552) }
func (r *rosePineMoon) Fg() tcell.Color          { return tcell.NewHexColor(0xe0def4) }
func (r *rosePineMoon) FgDim() tcell.Color       { return tcell.NewHexColor(0x6e6a86) }
func (r *rosePineMoon) FgMuted() tcell.Color     { return tcell.NewHexColor(0x524f67) }
func (r *rosePineMoon) Accent() tcell.Color      { return tcell.NewHexColor(0xea9a97) }
func (r *rosePineMoon) AccentDim() tcell.Color   { return tcell.NewHexColor(0xc4a7e7) }
func (r *rosePineMoon) Highlight() tcell.Color   { return tcell.NewHexColor(0x44415a) }
func (r *rosePineMoon) Success() tcell.Color     { return tcell.NewHexColor(0x9ccfd8) }
func (r *rosePineMoon) Warning() tcell.Color     { return tcell.NewHexColor(0xf6c177) }
func (r *rosePineMoon) Error() tcell.Color       { return tcell.NewHexColor(0xeb6f92) }
func (r *rosePineMoon) Info() tcell.Color        { return tcell.NewHexColor(0x3e8fb0) }
func (r *rosePineMoon) Border() tcell.Color      { return tcell.NewHexColor(0x2a273f) }
func (r *rosePineMoon) BorderFocus() tcell.Color { return tcell.NewHexColor(0xea9a97) }
func (r *rosePineMoon) Header() tcell.Color      { return tcell.NewHexColor(0x393552) }
func (r *rosePineMoon) Menu() tcell.Color        { return tcell.NewHexColor(0x232136) }
func (r *rosePineMoon) TableHeader() tcell.Color { return tcell.NewHexColor(0xea9a97) }
func (r *rosePineMoon) Key() tcell.Color         { return tcell.NewHexColor(0xc4a7e7) }
func (r *rosePineMoon) Crumb() tcell.Color       { return tcell.NewHexColor(0xea9a97) }
func (r *rosePineMoon) PanelBorder() tcell.Color { return tcell.NewHexColor(0x44415a) }
func (r *rosePineMoon) PanelTitle() tcell.Color  { return tcell.NewHexColor(0xea9a97) }

// RosePineDawn - Light variant of Rosé Pine
var RosePineDawn = &rosePineDawn{}

type rosePineDawn struct{}

func (r *rosePineDawn) Bg() tcell.Color          { return tcell.NewHexColor(0xfaf4ed) }
func (r *rosePineDawn) BgLight() tcell.Color     { return tcell.NewHexColor(0xf2e9e1) }
func (r *rosePineDawn) BgDark() tcell.Color      { return tcell.NewHexColor(0xfffaf3) }
func (r *rosePineDawn) Fg() tcell.Color          { return tcell.NewHexColor(0x575279) }
func (r *rosePineDawn) FgDim() tcell.Color       { return tcell.NewHexColor(0x9893a5) }
func (r *rosePineDawn) FgMuted() tcell.Color     { return tcell.NewHexColor(0xb4b0c4) }
func (r *rosePineDawn) Accent() tcell.Color      { return tcell.NewHexColor(0xd7827e) }
func (r *rosePineDawn) AccentDim() tcell.Color   { return tcell.NewHexColor(0x907aa9) }
func (r *rosePineDawn) Highlight() tcell.Color   { return tcell.NewHexColor(0xf2e9e1) }
func (r *rosePineDawn) Success() tcell.Color     { return tcell.NewHexColor(0x56949f) }
func (r *rosePineDawn) Warning() tcell.Color     { return tcell.NewHexColor(0xea9d34) }
func (r *rosePineDawn) Error() tcell.Color       { return tcell.NewHexColor(0xb4637a) }
func (r *rosePineDawn) Info() tcell.Color        { return tcell.NewHexColor(0x286983) }
func (r *rosePineDawn) Border() tcell.Color      { return tcell.NewHexColor(0xdfdad9) }
func (r *rosePineDawn) BorderFocus() tcell.Color { return tcell.NewHexColor(0xd7827e) }
func (r *rosePineDawn) Header() tcell.Color      { return tcell.NewHexColor(0xfffaf3) }
func (r *rosePineDawn) Menu() tcell.Color        { return tcell.NewHexColor(0xfaf4ed) }
func (r *rosePineDawn) TableHeader() tcell.Color { return tcell.NewHexColor(0xd7827e) }
func (r *rosePineDawn) Key() tcell.Color         { return tcell.NewHexColor(0x907aa9) }
func (r *rosePineDawn) Crumb() tcell.Color       { return tcell.NewHexColor(0xd7827e) }
func (r *rosePineDawn) PanelBorder() tcell.Color { return tcell.NewHexColor(0xcecacd) }
func (r *rosePineDawn) PanelTitle() tcell.Color  { return tcell.NewHexColor(0xd7827e) }
