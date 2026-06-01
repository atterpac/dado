package intermediate

import (

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&SplitDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Split",
			DemoDescription: "Resizable split pane layout",
			DemoCategory:    demos.Intermediate,
			DemoCode:        splitCode,
		},
	})
}

// SplitDemo demonstrates the Split component.
type SplitDemo struct {
	demos.DemoBase
	split       *components.Split
	direction   string
	showDivider bool
}

// Component returns the demo component.
func (d *SplitDemo) Component() core.Widget {
	d.direction = "horizontal"
	d.showDivider = true

	d.split = components.NewSplit().
		SetDirection(components.SplitHorizontal).
		SetRatio(0.5).
		SetMinSize(10).
		SetShowDivider(d.showDivider).
		SetResizable(true)

	left := d.createPane("Left Pane", "Use Ctrl+Left/Right to resize")
	right := d.createPane("Right Pane", "Tab to switch focus")

	d.split.SetLeft(left)
	d.split.SetRight(right)

	// Setup Props and ResetFunc
	d.Props = []demos.PropertyDescriptor{
		demos.SelectProp("direction", "Split direction",
			[]string{"horizontal", "vertical"},
			func() string { return d.direction },
			func(v string) {
				d.direction = v
				if v == "vertical" {
					d.split.SetDirection(components.SplitVertical)
				} else {
					d.split.SetDirection(components.SplitHorizontal)
				}
			},
			"horizontal",
		),
		demos.BoolProp("showDivider", "Show divider line",
			func() bool { return d.showDivider },
			func(v bool) { d.showDivider = v; d.split.SetShowDivider(v) },
			true,
		),
	}
	d.ResetFunc = func() {
		// Also reset the ratio which isn't a property
		d.split.SetRatio(0.5)
	}

	return d.split
}

func (d *SplitDemo) createPane(title, content string) *components.Panel {
	tv := core.NewTextView()
	tv.SetText(content)
	tv.SetBackgroundColor(theme.Bg())

	panel := components.NewPanel().
		SetTitle(title).
		SetContent(tv)

	return panel
}

const splitCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create horizontal split (left/right)
split := components.NewSplit().
    SetDirection(components.SplitHorizontal).
    SetRatio(0.3).      // 30% for first pane
    SetMinSize(10).     // Minimum pane size
    SetResizable(true). // Allow Ctrl+Arrow resizing
    SetShowDivider(true)

split.SetLeft(leftPane)
split.SetRight(rightPane)

// Or vertical split (top/bottom)
split.SetDirection(components.SplitVertical)
split.SetTop(topPane)
split.SetBottom(bottomPane)

// Focus management
split.FocusFirst()  // Focus left/top pane
split.FocusSecond() // Focus right/bottom pane
split.ToggleFocus() // Switch between panes

// Resize callback
split.SetOnResize(func(ratio float64) {
    fmt.Printf("New ratio: %.2f\n", ratio)
})

// Get current state
ratio := split.GetRatio()
focused := split.FocusedPane() // 0 or 1
`
