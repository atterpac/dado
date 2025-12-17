package themes

import "github.com/gdamore/tcell/v2"

// CatppuccinMocha - The darkest Catppuccin variant
var CatppuccinMocha = &catppuccinMocha{}

type catppuccinMocha struct{}

func (c *catppuccinMocha) Bg() tcell.Color          { return tcell.NewHexColor(0x1e1e2e) }
func (c *catppuccinMocha) BgLight() tcell.Color     { return tcell.NewHexColor(0x313244) }
func (c *catppuccinMocha) BgDark() tcell.Color      { return tcell.NewHexColor(0x181825) }
func (c *catppuccinMocha) Fg() tcell.Color          { return tcell.NewHexColor(0xcdd6f4) }
func (c *catppuccinMocha) FgDim() tcell.Color       { return tcell.NewHexColor(0x6c7086) }
func (c *catppuccinMocha) FgMuted() tcell.Color     { return tcell.NewHexColor(0x45475a) }
func (c *catppuccinMocha) Accent() tcell.Color      { return tcell.NewHexColor(0xf5c2e7) }
func (c *catppuccinMocha) AccentDim() tcell.Color   { return tcell.NewHexColor(0xcba6f7) }
func (c *catppuccinMocha) Highlight() tcell.Color   { return tcell.NewHexColor(0x585b70) }
func (c *catppuccinMocha) Success() tcell.Color     { return tcell.NewHexColor(0xa6e3a1) }
func (c *catppuccinMocha) Warning() tcell.Color     { return tcell.NewHexColor(0xf9e2af) }
func (c *catppuccinMocha) Error() tcell.Color       { return tcell.NewHexColor(0xf38ba8) }
func (c *catppuccinMocha) Info() tcell.Color        { return tcell.NewHexColor(0x89dceb) }
func (c *catppuccinMocha) Border() tcell.Color      { return tcell.NewHexColor(0x45475a) }
func (c *catppuccinMocha) BorderFocus() tcell.Color { return tcell.NewHexColor(0xf5c2e7) }
func (c *catppuccinMocha) Header() tcell.Color      { return tcell.NewHexColor(0x181825) }
func (c *catppuccinMocha) Menu() tcell.Color        { return tcell.NewHexColor(0x1e1e2e) }
func (c *catppuccinMocha) TableHeader() tcell.Color { return tcell.NewHexColor(0xf5c2e7) }
func (c *catppuccinMocha) Key() tcell.Color         { return tcell.NewHexColor(0xcba6f7) }
func (c *catppuccinMocha) Crumb() tcell.Color       { return tcell.NewHexColor(0xf5c2e7) }
func (c *catppuccinMocha) PanelBorder() tcell.Color { return tcell.NewHexColor(0x585b70) }
func (c *catppuccinMocha) PanelTitle() tcell.Color  { return tcell.NewHexColor(0xf5c2e7) }

// CatppuccinMacchiato - Medium-dark variant
var CatppuccinMacchiato = &catppuccinMacchiato{}

type catppuccinMacchiato struct{}

func (c *catppuccinMacchiato) Bg() tcell.Color          { return tcell.NewHexColor(0x24273a) }
func (c *catppuccinMacchiato) BgLight() tcell.Color     { return tcell.NewHexColor(0x363a4f) }
func (c *catppuccinMacchiato) BgDark() tcell.Color      { return tcell.NewHexColor(0x1e2030) }
func (c *catppuccinMacchiato) Fg() tcell.Color          { return tcell.NewHexColor(0xcad3f5) }
func (c *catppuccinMacchiato) FgDim() tcell.Color       { return tcell.NewHexColor(0x6e738d) }
func (c *catppuccinMacchiato) FgMuted() tcell.Color     { return tcell.NewHexColor(0x494d64) }
func (c *catppuccinMacchiato) Accent() tcell.Color      { return tcell.NewHexColor(0xf5bde6) }
func (c *catppuccinMacchiato) AccentDim() tcell.Color   { return tcell.NewHexColor(0xc6a0f6) }
func (c *catppuccinMacchiato) Highlight() tcell.Color   { return tcell.NewHexColor(0x5b6078) }
func (c *catppuccinMacchiato) Success() tcell.Color     { return tcell.NewHexColor(0xa6da95) }
func (c *catppuccinMacchiato) Warning() tcell.Color     { return tcell.NewHexColor(0xeed49f) }
func (c *catppuccinMacchiato) Error() tcell.Color       { return tcell.NewHexColor(0xed8796) }
func (c *catppuccinMacchiato) Info() tcell.Color        { return tcell.NewHexColor(0x91d7e3) }
func (c *catppuccinMacchiato) Border() tcell.Color      { return tcell.NewHexColor(0x494d64) }
func (c *catppuccinMacchiato) BorderFocus() tcell.Color { return tcell.NewHexColor(0xf5bde6) }
func (c *catppuccinMacchiato) Header() tcell.Color      { return tcell.NewHexColor(0x1e2030) }
func (c *catppuccinMacchiato) Menu() tcell.Color        { return tcell.NewHexColor(0x24273a) }
func (c *catppuccinMacchiato) TableHeader() tcell.Color { return tcell.NewHexColor(0xf5bde6) }
func (c *catppuccinMacchiato) Key() tcell.Color         { return tcell.NewHexColor(0xc6a0f6) }
func (c *catppuccinMacchiato) Crumb() tcell.Color       { return tcell.NewHexColor(0xf5bde6) }
func (c *catppuccinMacchiato) PanelBorder() tcell.Color { return tcell.NewHexColor(0x5b6078) }
func (c *catppuccinMacchiato) PanelTitle() tcell.Color  { return tcell.NewHexColor(0xf5bde6) }

// CatppuccinFrappe - Medium variant
var CatppuccinFrappe = &catppuccinFrappe{}

type catppuccinFrappe struct{}

func (c *catppuccinFrappe) Bg() tcell.Color          { return tcell.NewHexColor(0x303446) }
func (c *catppuccinFrappe) BgLight() tcell.Color     { return tcell.NewHexColor(0x414559) }
func (c *catppuccinFrappe) BgDark() tcell.Color      { return tcell.NewHexColor(0x292c3c) }
func (c *catppuccinFrappe) Fg() tcell.Color          { return tcell.NewHexColor(0xc6d0f5) }
func (c *catppuccinFrappe) FgDim() tcell.Color       { return tcell.NewHexColor(0x737994) }
func (c *catppuccinFrappe) FgMuted() tcell.Color     { return tcell.NewHexColor(0x51576d) }
func (c *catppuccinFrappe) Accent() tcell.Color      { return tcell.NewHexColor(0xf4b8e4) }
func (c *catppuccinFrappe) AccentDim() tcell.Color   { return tcell.NewHexColor(0xca9ee6) }
func (c *catppuccinFrappe) Highlight() tcell.Color   { return tcell.NewHexColor(0x626880) }
func (c *catppuccinFrappe) Success() tcell.Color     { return tcell.NewHexColor(0xa6d189) }
func (c *catppuccinFrappe) Warning() tcell.Color     { return tcell.NewHexColor(0xe5c890) }
func (c *catppuccinFrappe) Error() tcell.Color       { return tcell.NewHexColor(0xe78284) }
func (c *catppuccinFrappe) Info() tcell.Color        { return tcell.NewHexColor(0x85c1dc) }
func (c *catppuccinFrappe) Border() tcell.Color      { return tcell.NewHexColor(0x51576d) }
func (c *catppuccinFrappe) BorderFocus() tcell.Color { return tcell.NewHexColor(0xf4b8e4) }
func (c *catppuccinFrappe) Header() tcell.Color      { return tcell.NewHexColor(0x292c3c) }
func (c *catppuccinFrappe) Menu() tcell.Color        { return tcell.NewHexColor(0x303446) }
func (c *catppuccinFrappe) TableHeader() tcell.Color { return tcell.NewHexColor(0xf4b8e4) }
func (c *catppuccinFrappe) Key() tcell.Color         { return tcell.NewHexColor(0xca9ee6) }
func (c *catppuccinFrappe) Crumb() tcell.Color       { return tcell.NewHexColor(0xf4b8e4) }
func (c *catppuccinFrappe) PanelBorder() tcell.Color { return tcell.NewHexColor(0x626880) }
func (c *catppuccinFrappe) PanelTitle() tcell.Color  { return tcell.NewHexColor(0xf4b8e4) }

// CatppuccinLatte - Light variant
var CatppuccinLatte = &catppuccinLatte{}

type catppuccinLatte struct{}

func (c *catppuccinLatte) Bg() tcell.Color          { return tcell.NewHexColor(0xeff1f5) }
func (c *catppuccinLatte) BgLight() tcell.Color     { return tcell.NewHexColor(0xccd0da) }
func (c *catppuccinLatte) BgDark() tcell.Color      { return tcell.NewHexColor(0xe6e9ef) }
func (c *catppuccinLatte) Fg() tcell.Color          { return tcell.NewHexColor(0x4c4f69) }
func (c *catppuccinLatte) FgDim() tcell.Color       { return tcell.NewHexColor(0x6c6f85) }
func (c *catppuccinLatte) FgMuted() tcell.Color     { return tcell.NewHexColor(0x9ca0b0) }
func (c *catppuccinLatte) Accent() tcell.Color      { return tcell.NewHexColor(0xea76cb) }
func (c *catppuccinLatte) AccentDim() tcell.Color   { return tcell.NewHexColor(0x8839ef) }
func (c *catppuccinLatte) Highlight() tcell.Color   { return tcell.NewHexColor(0xacb0be) }
func (c *catppuccinLatte) Success() tcell.Color     { return tcell.NewHexColor(0x40a02b) }
func (c *catppuccinLatte) Warning() tcell.Color     { return tcell.NewHexColor(0xdf8e1d) }
func (c *catppuccinLatte) Error() tcell.Color       { return tcell.NewHexColor(0xd20f39) }
func (c *catppuccinLatte) Info() tcell.Color        { return tcell.NewHexColor(0x04a5e5) }
func (c *catppuccinLatte) Border() tcell.Color      { return tcell.NewHexColor(0xbcc0cc) }
func (c *catppuccinLatte) BorderFocus() tcell.Color { return tcell.NewHexColor(0xea76cb) }
func (c *catppuccinLatte) Header() tcell.Color      { return tcell.NewHexColor(0xe6e9ef) }
func (c *catppuccinLatte) Menu() tcell.Color        { return tcell.NewHexColor(0xeff1f5) }
func (c *catppuccinLatte) TableHeader() tcell.Color { return tcell.NewHexColor(0xea76cb) }
func (c *catppuccinLatte) Key() tcell.Color         { return tcell.NewHexColor(0x8839ef) }
func (c *catppuccinLatte) Crumb() tcell.Color       { return tcell.NewHexColor(0xea76cb) }
func (c *catppuccinLatte) PanelBorder() tcell.Color { return tcell.NewHexColor(0xacb0be) }
func (c *catppuccinLatte) PanelTitle() tcell.Color  { return tcell.NewHexColor(0xea76cb) }
