package themes

import "github.com/gdamore/tcell/v2"

// Dracula is the Dracula theme with vibrant colors.
// A dark theme with purple accents and high contrast.
var Dracula = &draculaTheme{}

type draculaTheme struct{}

func (d *draculaTheme) Bg() tcell.Color          { return tcell.NewHexColor(0x282a36) }
func (d *draculaTheme) BgLight() tcell.Color     { return tcell.NewHexColor(0x44475a) }
func (d *draculaTheme) BgDark() tcell.Color      { return tcell.NewHexColor(0x21222c) }
func (d *draculaTheme) Fg() tcell.Color          { return tcell.NewHexColor(0xf8f8f2) }
func (d *draculaTheme) FgDim() tcell.Color       { return tcell.NewHexColor(0x6272a4) }
func (d *draculaTheme) FgMuted() tcell.Color     { return tcell.NewHexColor(0x44475a) }
func (d *draculaTheme) Accent() tcell.Color      { return tcell.NewHexColor(0xff79c6) }
func (d *draculaTheme) AccentDim() tcell.Color   { return tcell.NewHexColor(0xbd93f9) }
func (d *draculaTheme) Highlight() tcell.Color   { return tcell.NewHexColor(0x44475a) }
func (d *draculaTheme) Success() tcell.Color     { return tcell.NewHexColor(0x50fa7b) }
func (d *draculaTheme) Warning() tcell.Color     { return tcell.NewHexColor(0xf1fa8c) }
func (d *draculaTheme) Error() tcell.Color       { return tcell.NewHexColor(0xff5555) }
func (d *draculaTheme) Info() tcell.Color        { return tcell.NewHexColor(0x8be9fd) }
func (d *draculaTheme) Border() tcell.Color      { return tcell.NewHexColor(0x44475a) }
func (d *draculaTheme) BorderFocus() tcell.Color { return tcell.NewHexColor(0xff79c6) }
func (d *draculaTheme) Header() tcell.Color      { return tcell.NewHexColor(0x21222c) }
func (d *draculaTheme) Menu() tcell.Color        { return tcell.NewHexColor(0x282a36) }
func (d *draculaTheme) TableHeader() tcell.Color { return tcell.NewHexColor(0xff79c6) }
func (d *draculaTheme) Key() tcell.Color         { return tcell.NewHexColor(0xbd93f9) }
func (d *draculaTheme) Crumb() tcell.Color       { return tcell.NewHexColor(0xff79c6) }
func (d *draculaTheme) PanelBorder() tcell.Color { return tcell.NewHexColor(0x6272a4) }
func (d *draculaTheme) PanelTitle() tcell.Color  { return tcell.NewHexColor(0xff79c6) }

// DraculaLight is the light variant of Dracula
var DraculaLight = &draculaLight{}

type draculaLight struct{}

func (d *draculaLight) Bg() tcell.Color          { return tcell.NewHexColor(0xfffbeb) }
func (d *draculaLight) BgLight() tcell.Color     { return tcell.NewHexColor(0xcfcfde) }
func (d *draculaLight) BgDark() tcell.Color      { return tcell.NewHexColor(0xf5f5f0) }
func (d *draculaLight) Fg() tcell.Color          { return tcell.NewHexColor(0x1f1f1f) }
func (d *draculaLight) FgDim() tcell.Color       { return tcell.NewHexColor(0x6c664b) }
func (d *draculaLight) FgMuted() tcell.Color     { return tcell.NewHexColor(0x9c9889) }
func (d *draculaLight) Accent() tcell.Color      { return tcell.NewHexColor(0xa3144d) }
func (d *draculaLight) AccentDim() tcell.Color   { return tcell.NewHexColor(0x644ac9) }
func (d *draculaLight) Highlight() tcell.Color   { return tcell.NewHexColor(0xcfcfde) }
func (d *draculaLight) Success() tcell.Color     { return tcell.NewHexColor(0x14710a) }
func (d *draculaLight) Warning() tcell.Color     { return tcell.NewHexColor(0x846e15) }
func (d *draculaLight) Error() tcell.Color       { return tcell.NewHexColor(0xcb3a2a) }
func (d *draculaLight) Info() tcell.Color        { return tcell.NewHexColor(0x0e6f8e) }
func (d *draculaLight) Border() tcell.Color      { return tcell.NewHexColor(0xcfcfde) }
func (d *draculaLight) BorderFocus() tcell.Color { return tcell.NewHexColor(0xa3144d) }
func (d *draculaLight) Header() tcell.Color      { return tcell.NewHexColor(0xf5f5f0) }
func (d *draculaLight) Menu() tcell.Color        { return tcell.NewHexColor(0xfffbeb) }
func (d *draculaLight) TableHeader() tcell.Color { return tcell.NewHexColor(0xa3144d) }
func (d *draculaLight) Key() tcell.Color         { return tcell.NewHexColor(0x644ac9) }
func (d *draculaLight) Crumb() tcell.Color       { return tcell.NewHexColor(0xa3144d) }
func (d *draculaLight) PanelBorder() tcell.Color { return tcell.NewHexColor(0x6c664b) }
func (d *draculaLight) PanelTitle() tcell.Color  { return tcell.NewHexColor(0xa3144d) }
