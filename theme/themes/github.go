package themes

import "github.com/gdamore/tcell/v2"

// GitHubDark - GitHub's dark theme
var GitHubDark = &githubDark{}

type githubDark struct{}

func (g *githubDark) Bg() tcell.Color          { return tcell.NewHexColor(0x0d1117) }
func (g *githubDark) BgLight() tcell.Color     { return tcell.NewHexColor(0x161b22) }
func (g *githubDark) BgDark() tcell.Color      { return tcell.NewHexColor(0x010409) }
func (g *githubDark) Fg() tcell.Color          { return tcell.NewHexColor(0xc9d1d9) }
func (g *githubDark) FgDim() tcell.Color       { return tcell.NewHexColor(0x8b949e) }
func (g *githubDark) FgMuted() tcell.Color     { return tcell.NewHexColor(0x6e7681) }
func (g *githubDark) Accent() tcell.Color      { return tcell.NewHexColor(0x58a6ff) }
func (g *githubDark) AccentDim() tcell.Color   { return tcell.NewHexColor(0xa371f7) }
func (g *githubDark) Highlight() tcell.Color   { return tcell.NewHexColor(0x21262d) }
func (g *githubDark) Success() tcell.Color     { return tcell.NewHexColor(0x3fb950) }
func (g *githubDark) Warning() tcell.Color     { return tcell.NewHexColor(0xd29922) }
func (g *githubDark) Error() tcell.Color       { return tcell.NewHexColor(0xf85149) }
func (g *githubDark) Info() tcell.Color        { return tcell.NewHexColor(0x58a6ff) }
func (g *githubDark) Border() tcell.Color      { return tcell.NewHexColor(0x30363d) }
func (g *githubDark) BorderFocus() tcell.Color { return tcell.NewHexColor(0x58a6ff) }
func (g *githubDark) Header() tcell.Color      { return tcell.NewHexColor(0x010409) }
func (g *githubDark) Menu() tcell.Color        { return tcell.NewHexColor(0x0d1117) }
func (g *githubDark) TableHeader() tcell.Color { return tcell.NewHexColor(0x58a6ff) }
func (g *githubDark) Key() tcell.Color         { return tcell.NewHexColor(0xa371f7) }
func (g *githubDark) Crumb() tcell.Color       { return tcell.NewHexColor(0x58a6ff) }
func (g *githubDark) PanelBorder() tcell.Color { return tcell.NewHexColor(0x30363d) }
func (g *githubDark) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x58a6ff) }

// GitHubLight - GitHub's light theme
var GitHubLight = &githubLight{}

type githubLight struct{}

func (g *githubLight) Bg() tcell.Color          { return tcell.NewHexColor(0xffffff) }
func (g *githubLight) BgLight() tcell.Color     { return tcell.NewHexColor(0xf6f8fa) }
func (g *githubLight) BgDark() tcell.Color      { return tcell.NewHexColor(0xf0f0f0) }
func (g *githubLight) Fg() tcell.Color          { return tcell.NewHexColor(0x24292f) }
func (g *githubLight) FgDim() tcell.Color       { return tcell.NewHexColor(0x57606a) }
func (g *githubLight) FgMuted() tcell.Color     { return tcell.NewHexColor(0x8b949e) }
func (g *githubLight) Accent() tcell.Color      { return tcell.NewHexColor(0x0969da) }
func (g *githubLight) AccentDim() tcell.Color   { return tcell.NewHexColor(0x8250df) }
func (g *githubLight) Highlight() tcell.Color   { return tcell.NewHexColor(0xf6f8fa) }
func (g *githubLight) Success() tcell.Color     { return tcell.NewHexColor(0x1a7f37) }
func (g *githubLight) Warning() tcell.Color     { return tcell.NewHexColor(0x9a6700) }
func (g *githubLight) Error() tcell.Color       { return tcell.NewHexColor(0xcf222e) }
func (g *githubLight) Info() tcell.Color        { return tcell.NewHexColor(0x0969da) }
func (g *githubLight) Border() tcell.Color      { return tcell.NewHexColor(0xd0d7de) }
func (g *githubLight) BorderFocus() tcell.Color { return tcell.NewHexColor(0x0969da) }
func (g *githubLight) Header() tcell.Color      { return tcell.NewHexColor(0xf0f0f0) }
func (g *githubLight) Menu() tcell.Color        { return tcell.NewHexColor(0xffffff) }
func (g *githubLight) TableHeader() tcell.Color { return tcell.NewHexColor(0x0969da) }
func (g *githubLight) Key() tcell.Color         { return tcell.NewHexColor(0x8250df) }
func (g *githubLight) Crumb() tcell.Color       { return tcell.NewHexColor(0x0969da) }
func (g *githubLight) PanelBorder() tcell.Color { return tcell.NewHexColor(0xd0d7de) }
func (g *githubLight) PanelTitle() tcell.Color  { return tcell.NewHexColor(0x0969da) }
