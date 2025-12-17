package themes

import "github.com/gdamore/tcell/v2"

// Kanagawa (Wave) - Dark theme inspired by the famous wave painting
var Kanagawa = &kanagawa{}

type kanagawa struct{}

func (k *kanagawa) Bg() tcell.Color          { return tcell.NewHexColor(0x1f1f28) }
func (k *kanagawa) BgLight() tcell.Color     { return tcell.NewHexColor(0x2a2a37) }
func (k *kanagawa) BgDark() tcell.Color      { return tcell.NewHexColor(0x16161d) }
func (k *kanagawa) Fg() tcell.Color          { return tcell.NewHexColor(0xdcd7ba) }
func (k *kanagawa) FgDim() tcell.Color       { return tcell.NewHexColor(0x727169) }
func (k *kanagawa) FgMuted() tcell.Color     { return tcell.NewHexColor(0x54546d) }
func (k *kanagawa) Accent() tcell.Color      { return tcell.NewHexColor(0x7e9cd8) }
func (k *kanagawa) AccentDim() tcell.Color   { return tcell.NewHexColor(0x957fb8) }
func (k *kanagawa) Highlight() tcell.Color   { return tcell.NewHexColor(0x2d4f67) }
func (k *kanagawa) Success() tcell.Color     { return tcell.NewHexColor(0x98bb6c) }
func (k *kanagawa) Warning() tcell.Color     { return tcell.NewHexColor(0xe6c384) }
func (k *kanagawa) Error() tcell.Color       { return tcell.NewHexColor(0xff5d62) }
func (k *kanagawa) Info() tcell.Color        { return tcell.NewHexColor(0x7fb4ca) }
func (k *kanagawa) Border() tcell.Color      { return tcell.NewHexColor(0x54546d) }
func (k *kanagawa) BorderFocus() tcell.Color { return tcell.NewHexColor(0x7e9cd8) }
func (k *kanagawa) Header() tcell.Color      { return tcell.NewHexColor(0x16161d) }
func (k *kanagawa) Menu() tcell.Color        { return tcell.NewHexColor(0x1f1f28) }
func (k *kanagawa) TableHeader() tcell.Color { return tcell.NewHexColor(0x7e9cd8) }
func (k *kanagawa) Key() tcell.Color         { return tcell.NewHexColor(0x957fb8) }
func (k *kanagawa) Crumb() tcell.Color       { return tcell.NewHexColor(0x7e9cd8) }
func (k *kanagawa) PanelBorder() tcell.Color { return tcell.NewHexColor(0x54546d) }
func (k *kanagawa) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x7e9cd8) }

// KanagawaDragon - Darker variant with warmer ink tones
var KanagawaDragon = &kanagawaDragon{}

type kanagawaDragon struct{}

func (k *kanagawaDragon) Bg() tcell.Color          { return tcell.NewHexColor(0x181616) }
func (k *kanagawaDragon) BgLight() tcell.Color     { return tcell.NewHexColor(0x282727) }
func (k *kanagawaDragon) BgDark() tcell.Color      { return tcell.NewHexColor(0x0d0c0c) }
func (k *kanagawaDragon) Fg() tcell.Color          { return tcell.NewHexColor(0xc5c9c5) }
func (k *kanagawaDragon) FgDim() tcell.Color       { return tcell.NewHexColor(0x737c73) }
func (k *kanagawaDragon) FgMuted() tcell.Color     { return tcell.NewHexColor(0x625e5a) }
func (k *kanagawaDragon) Accent() tcell.Color      { return tcell.NewHexColor(0x8ba4b0) }
func (k *kanagawaDragon) AccentDim() tcell.Color   { return tcell.NewHexColor(0xa292a3) }
func (k *kanagawaDragon) Highlight() tcell.Color   { return tcell.NewHexColor(0x2d4f67) }
func (k *kanagawaDragon) Success() tcell.Color     { return tcell.NewHexColor(0x87a987) }
func (k *kanagawaDragon) Warning() tcell.Color     { return tcell.NewHexColor(0xc4b28a) }
func (k *kanagawaDragon) Error() tcell.Color       { return tcell.NewHexColor(0xc4746e) }
func (k *kanagawaDragon) Info() tcell.Color        { return tcell.NewHexColor(0x8ea4a2) }
func (k *kanagawaDragon) Border() tcell.Color      { return tcell.NewHexColor(0x625e5a) }
func (k *kanagawaDragon) BorderFocus() tcell.Color { return tcell.NewHexColor(0x8ba4b0) }
func (k *kanagawaDragon) Header() tcell.Color      { return tcell.NewHexColor(0x0d0c0c) }
func (k *kanagawaDragon) Menu() tcell.Color        { return tcell.NewHexColor(0x181616) }
func (k *kanagawaDragon) TableHeader() tcell.Color { return tcell.NewHexColor(0x8ba4b0) }
func (k *kanagawaDragon) Key() tcell.Color         { return tcell.NewHexColor(0xa292a3) }
func (k *kanagawaDragon) Crumb() tcell.Color       { return tcell.NewHexColor(0x8ba4b0) }
func (k *kanagawaDragon) PanelBorder() tcell.Color { return tcell.NewHexColor(0x625e5a) }
func (k *kanagawaDragon) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x8ba4b0) }

// KanagawaLotus - Light variant with soft paper tones
var KanagawaLotus = &kanagawaLotus{}

type kanagawaLotus struct{}

func (k *kanagawaLotus) Bg() tcell.Color          { return tcell.NewHexColor(0xf2ecbc) }
func (k *kanagawaLotus) BgLight() tcell.Color     { return tcell.NewHexColor(0xf9f3c3) }
func (k *kanagawaLotus) BgDark() tcell.Color      { return tcell.NewHexColor(0xe5ddb0) }
func (k *kanagawaLotus) Fg() tcell.Color          { return tcell.NewHexColor(0x545464) }
func (k *kanagawaLotus) FgDim() tcell.Color       { return tcell.NewHexColor(0x8a8980) }
func (k *kanagawaLotus) FgMuted() tcell.Color     { return tcell.NewHexColor(0x9e9b93) }
func (k *kanagawaLotus) Accent() tcell.Color      { return tcell.NewHexColor(0x4d699b) }
func (k *kanagawaLotus) AccentDim() tcell.Color   { return tcell.NewHexColor(0x624c83) }
func (k *kanagawaLotus) Highlight() tcell.Color   { return tcell.NewHexColor(0xd9d5c3) }
func (k *kanagawaLotus) Success() tcell.Color     { return tcell.NewHexColor(0x6f894e) }
func (k *kanagawaLotus) Warning() tcell.Color     { return tcell.NewHexColor(0x77713f) }
func (k *kanagawaLotus) Error() tcell.Color       { return tcell.NewHexColor(0xc84053) }
func (k *kanagawaLotus) Info() tcell.Color        { return tcell.NewHexColor(0x597b75) }
func (k *kanagawaLotus) Border() tcell.Color      { return tcell.NewHexColor(0x9e9b93) }
func (k *kanagawaLotus) BorderFocus() tcell.Color { return tcell.NewHexColor(0x4d699b) }
func (k *kanagawaLotus) Header() tcell.Color      { return tcell.NewHexColor(0xe5ddb0) }
func (k *kanagawaLotus) Menu() tcell.Color        { return tcell.NewHexColor(0xf2ecbc) }
func (k *kanagawaLotus) TableHeader() tcell.Color { return tcell.NewHexColor(0x4d699b) }
func (k *kanagawaLotus) Key() tcell.Color         { return tcell.NewHexColor(0x624c83) }
func (k *kanagawaLotus) Crumb() tcell.Color       { return tcell.NewHexColor(0x4d699b) }
func (k *kanagawaLotus) PanelBorder() tcell.Color { return tcell.NewHexColor(0x9e9b93) }
func (k *kanagawaLotus) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x4d699b) }
