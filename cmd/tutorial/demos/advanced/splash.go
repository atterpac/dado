package advanced

import (
	"github.com/rivo/tview"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&SplashDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Splash",
			DemoDescription: "Startup splash screen",
			DemoCategory:    demos.Advanced,
			DemoCode:        splashCode,
		},
	})
}

// SplashDemo demonstrates the Splash component.
type SplashDemo struct {
	demos.DemoBase
}

// Component returns the demo component.
func (d *SplashDemo) Component() tview.Primitive {
	// Create a preview of what a splash screen looks like
	// (The real Splash component is fullscreen, so we show a mock preview)

	// Logo with gradient
	logoView := tview.NewTextView()
	logoView.SetDynamicColors(true)
	logoView.SetTextAlign(tview.AlignCenter)
	logoView.SetBackgroundColor(theme.Bg())
	gradientLogo := theme.ApplyGradient(sampleLogo, theme.GradientDiagonal, theme.DefaultGradientColors())
	logoView.SetText(gradientLogo)

	// Status text
	statusView := tview.NewTextView()
	statusView.SetDynamicColors(true)
	statusView.SetTextAlign(tview.AlignCenter)
	statusView.SetBackgroundColor(theme.Bg())
	statusView.SetText("[" + theme.TagFgDim() + "]Press any key to continue...[-]")

	// Layout: center the logo and status
	content := tview.NewFlex().SetDirection(tview.FlexRow)
	content.SetBackgroundColor(theme.Bg())
	content.AddItem(nil, 0, 1, false) // Top spacer
	content.AddItem(logoView, 8, 0, false)
	content.AddItem(statusView, 2, 0, false)
	content.AddItem(nil, 0, 1, false) // Bottom spacer

	// Wrap in panel
	panel := components.NewPanel()
	panel.SetTitle("Splash Preview")
	panel.SetContent(content)

	return panel
}

const sampleLogo = `
     ██╗██╗ ██████╗
     ██║██║██╔════╝
     ██║██║██║  ███╗
██   ██║██║██║   ██║
╚█████╔╝██║╚██████╔╝
 ╚════╝ ╚═╝ ╚═════╝
`

const splashCode = `package main

import (
    "time"
    "github.com/atterpac/dado/components"
    "github.com/atterpac/dado/theme"
)

// Create splash screen
splash := components.NewSplash().
    SetLogo(asciiLogo).
    SetStatus("Loading...").
    SetGradient(theme.GradientDiagonal).
    Build() // Must call Build() after setting options

// Custom gradient colors
splash.SetColors([]string{"#FF6B6B", "#4ECDC4", "#45B7D1"})

// Auto-dismiss after duration
splash.SetAutoDismiss(3 * time.Second)

// Custom dismiss keys
splash.SetDismissKeys([]components.DismissKey{
    components.DismissEnter,
    components.DismissSpace,
})

// Callback when dismissed
splash.SetOnClose(func() {
    showMainApp()
})

// Dev mode: cycle themes with T/G keys
splash.SetDevMode(true)

// Show as initial page
pages.AddPage("splash", splash, true, true)
`
