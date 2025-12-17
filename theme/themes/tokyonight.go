package themes

import "github.com/gdamore/tcell/v2"

// TokyoNight Night - The original dark variant
var TokyoNightNight = &tokyonightNight{}

type tokyonightNight struct{}

func (t *tokyonightNight) Bg() tcell.Color          { return tcell.NewHexColor(0x1a1b26) }
func (t *tokyonightNight) BgLight() tcell.Color     { return tcell.NewHexColor(0x24283b) }
func (t *tokyonightNight) BgDark() tcell.Color      { return tcell.NewHexColor(0x16161e) }
func (t *tokyonightNight) Fg() tcell.Color          { return tcell.NewHexColor(0xc0caf5) }
func (t *tokyonightNight) FgDim() tcell.Color       { return tcell.NewHexColor(0x565f89) }
func (t *tokyonightNight) FgMuted() tcell.Color     { return tcell.NewHexColor(0x414868) }
func (t *tokyonightNight) Accent() tcell.Color      { return tcell.NewHexColor(0x7aa2f7) }
func (t *tokyonightNight) AccentDim() tcell.Color   { return tcell.NewHexColor(0xbb9af7) }
func (t *tokyonightNight) Highlight() tcell.Color   { return tcell.NewHexColor(0x283457) }
func (t *tokyonightNight) Success() tcell.Color     { return tcell.NewHexColor(0x9ece6a) }
func (t *tokyonightNight) Warning() tcell.Color     { return tcell.NewHexColor(0xe0af68) }
func (t *tokyonightNight) Error() tcell.Color       { return tcell.NewHexColor(0xf7768e) }
func (t *tokyonightNight) Info() tcell.Color        { return tcell.NewHexColor(0x7dcfff) }
func (t *tokyonightNight) Border() tcell.Color      { return tcell.NewHexColor(0x15161e) }
func (t *tokyonightNight) BorderFocus() tcell.Color { return tcell.NewHexColor(0x7aa2f7) }
func (t *tokyonightNight) Header() tcell.Color      { return tcell.NewHexColor(0x16161e) }
func (t *tokyonightNight) Menu() tcell.Color        { return tcell.NewHexColor(0x1a1b26) }
func (t *tokyonightNight) TableHeader() tcell.Color { return tcell.NewHexColor(0x7aa2f7) }
func (t *tokyonightNight) Key() tcell.Color         { return tcell.NewHexColor(0xbb9af7) }
func (t *tokyonightNight) Crumb() tcell.Color       { return tcell.NewHexColor(0x7aa2f7) }
func (t *tokyonightNight) PanelBorder() tcell.Color { return tcell.NewHexColor(0x283457) }
func (t *tokyonightNight) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x7aa2f7) }

// TokyoNight Storm - Slightly lighter variant
var TokyoNightStorm = &tokyonightStorm{}

type tokyonightStorm struct{}

func (t *tokyonightStorm) Bg() tcell.Color          { return tcell.NewHexColor(0x24283b) }
func (t *tokyonightStorm) BgLight() tcell.Color     { return tcell.NewHexColor(0x292e42) }
func (t *tokyonightStorm) BgDark() tcell.Color      { return tcell.NewHexColor(0x1f2335) }
func (t *tokyonightStorm) Fg() tcell.Color          { return tcell.NewHexColor(0xc0caf5) }
func (t *tokyonightStorm) FgDim() tcell.Color       { return tcell.NewHexColor(0x565f89) }
func (t *tokyonightStorm) FgMuted() tcell.Color     { return tcell.NewHexColor(0x414868) }
func (t *tokyonightStorm) Accent() tcell.Color      { return tcell.NewHexColor(0x7aa2f7) }
func (t *tokyonightStorm) AccentDim() tcell.Color   { return tcell.NewHexColor(0xbb9af7) }
func (t *tokyonightStorm) Highlight() tcell.Color   { return tcell.NewHexColor(0x292e42) }
func (t *tokyonightStorm) Success() tcell.Color     { return tcell.NewHexColor(0x9ece6a) }
func (t *tokyonightStorm) Warning() tcell.Color     { return tcell.NewHexColor(0xe0af68) }
func (t *tokyonightStorm) Error() tcell.Color       { return tcell.NewHexColor(0xf7768e) }
func (t *tokyonightStorm) Info() tcell.Color        { return tcell.NewHexColor(0x7dcfff) }
func (t *tokyonightStorm) Border() tcell.Color      { return tcell.NewHexColor(0x1d202f) }
func (t *tokyonightStorm) BorderFocus() tcell.Color { return tcell.NewHexColor(0x7aa2f7) }
func (t *tokyonightStorm) Header() tcell.Color      { return tcell.NewHexColor(0x1f2335) }
func (t *tokyonightStorm) Menu() tcell.Color        { return tcell.NewHexColor(0x24283b) }
func (t *tokyonightStorm) TableHeader() tcell.Color { return tcell.NewHexColor(0x7aa2f7) }
func (t *tokyonightStorm) Key() tcell.Color         { return tcell.NewHexColor(0xbb9af7) }
func (t *tokyonightStorm) Crumb() tcell.Color       { return tcell.NewHexColor(0x7aa2f7) }
func (t *tokyonightStorm) PanelBorder() tcell.Color { return tcell.NewHexColor(0x292e42) }
func (t *tokyonightStorm) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x7aa2f7) }

// TokyoNight Moon - Warmer, more purple tinted variant
var TokyoNightMoon = &tokyonightMoon{}

type tokyonightMoon struct{}

func (t *tokyonightMoon) Bg() tcell.Color          { return tcell.NewHexColor(0x222436) }
func (t *tokyonightMoon) BgLight() tcell.Color     { return tcell.NewHexColor(0x2f334d) }
func (t *tokyonightMoon) BgDark() tcell.Color      { return tcell.NewHexColor(0x1e2030) }
func (t *tokyonightMoon) Fg() tcell.Color          { return tcell.NewHexColor(0xc8d3f5) }
func (t *tokyonightMoon) FgDim() tcell.Color       { return tcell.NewHexColor(0x636da6) }
func (t *tokyonightMoon) FgMuted() tcell.Color     { return tcell.NewHexColor(0x545c7e) }
func (t *tokyonightMoon) Accent() tcell.Color      { return tcell.NewHexColor(0x82aaff) }
func (t *tokyonightMoon) AccentDim() tcell.Color   { return tcell.NewHexColor(0xc099ff) }
func (t *tokyonightMoon) Highlight() tcell.Color   { return tcell.NewHexColor(0x2f334d) }
func (t *tokyonightMoon) Success() tcell.Color     { return tcell.NewHexColor(0xc3e88d) }
func (t *tokyonightMoon) Warning() tcell.Color     { return tcell.NewHexColor(0xffc777) }
func (t *tokyonightMoon) Error() tcell.Color       { return tcell.NewHexColor(0xff757f) }
func (t *tokyonightMoon) Info() tcell.Color        { return tcell.NewHexColor(0x86e1fc) }
func (t *tokyonightMoon) Border() tcell.Color      { return tcell.NewHexColor(0x1b1d2b) }
func (t *tokyonightMoon) BorderFocus() tcell.Color { return tcell.NewHexColor(0x82aaff) }
func (t *tokyonightMoon) Header() tcell.Color      { return tcell.NewHexColor(0x1e2030) }
func (t *tokyonightMoon) Menu() tcell.Color        { return tcell.NewHexColor(0x222436) }
func (t *tokyonightMoon) TableHeader() tcell.Color { return tcell.NewHexColor(0x82aaff) }
func (t *tokyonightMoon) Key() tcell.Color         { return tcell.NewHexColor(0xc099ff) }
func (t *tokyonightMoon) Crumb() tcell.Color       { return tcell.NewHexColor(0x82aaff) }
func (t *tokyonightMoon) PanelBorder() tcell.Color { return tcell.NewHexColor(0x2f334d) }
func (t *tokyonightMoon) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x82aaff) }

// TokyoNight Day - Light variant
var TokyoNightDay = &tokyonightDay{}

type tokyonightDay struct{}

func (t *tokyonightDay) Bg() tcell.Color          { return tcell.NewHexColor(0xe1e2e7) }
func (t *tokyonightDay) BgLight() tcell.Color     { return tcell.NewHexColor(0xd0d5e3) }
func (t *tokyonightDay) BgDark() tcell.Color      { return tcell.NewHexColor(0xb4b5b9) }
func (t *tokyonightDay) Fg() tcell.Color          { return tcell.NewHexColor(0x3760bf) }
func (t *tokyonightDay) FgDim() tcell.Color       { return tcell.NewHexColor(0x848cb5) }
func (t *tokyonightDay) FgMuted() tcell.Color     { return tcell.NewHexColor(0x8990b3) }
func (t *tokyonightDay) Accent() tcell.Color      { return tcell.NewHexColor(0x2e7de9) }
func (t *tokyonightDay) AccentDim() tcell.Color   { return tcell.NewHexColor(0x9854f1) }
func (t *tokyonightDay) Highlight() tcell.Color   { return tcell.NewHexColor(0xb7c1e3) }
func (t *tokyonightDay) Success() tcell.Color     { return tcell.NewHexColor(0x587539) }
func (t *tokyonightDay) Warning() tcell.Color     { return tcell.NewHexColor(0x8c6c3e) }
func (t *tokyonightDay) Error() tcell.Color       { return tcell.NewHexColor(0xf52a65) }
func (t *tokyonightDay) Info() tcell.Color        { return tcell.NewHexColor(0x188092) }
func (t *tokyonightDay) Border() tcell.Color      { return tcell.NewHexColor(0xb4b5b9) }
func (t *tokyonightDay) BorderFocus() tcell.Color { return tcell.NewHexColor(0x2e7de9) }
func (t *tokyonightDay) Header() tcell.Color      { return tcell.NewHexColor(0xd0d5e3) }
func (t *tokyonightDay) Menu() tcell.Color        { return tcell.NewHexColor(0xe1e2e7) }
func (t *tokyonightDay) TableHeader() tcell.Color { return tcell.NewHexColor(0x2e7de9) }
func (t *tokyonightDay) Key() tcell.Color         { return tcell.NewHexColor(0x9854f1) }
func (t *tokyonightDay) Crumb() tcell.Color       { return tcell.NewHexColor(0x2e7de9) }
func (t *tokyonightDay) PanelBorder() tcell.Color { return tcell.NewHexColor(0xb4b5b9) }
func (t *tokyonightDay) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x2e7de9) }
