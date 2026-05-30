package basic

import (
	"github.com/rivo/tview"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
)

func init() {
	demos.Register(&SkeletonDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Skeleton",
			DemoDescription: "Loading placeholder animations",
			DemoCategory:    demos.Basic,
			DemoCode:        skeletonCode,
		},
	})
}

// SkeletonDemo demonstrates the Skeleton component.
type SkeletonDemo struct {
	demos.DemoBase
	container *tview.Flex
	skeletons []*components.Skeleton
	animated  bool
}

// Component returns the demo component.
func (d *SkeletonDemo) Component() tview.Primitive {
	d.animated = true

	textSkeleton := components.NewSkeleton().
		SetVariant(components.SkeletonText).
		SetLines(3).
		SetAnimated(true)
	textSkeleton.Start()

	blockSkeleton := components.NewSkeleton().
		SetVariant(components.SkeletonBlock).
		SetAnimated(true)
	blockSkeleton.Start()

	d.skeletons = []*components.Skeleton{textSkeleton, blockSkeleton}

	d.container = tview.NewFlex().SetDirection(tview.FlexRow)

	textLabel := tview.NewTextView().SetText("Text skeleton (3 lines):")
	blockLabel := tview.NewTextView().SetText("Block skeleton:")

	d.container.AddItem(textLabel, 1, 0, false)
	d.container.AddItem(textSkeleton, 3, 0, false)
	d.container.AddItem(nil, 1, 0, false)
	d.container.AddItem(blockLabel, 1, 0, false)
	d.container.AddItem(blockSkeleton, 3, 0, false)

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("animated", "Enable animation",
			func() bool { return d.animated },
			func(v bool) {
				d.animated = v
				for _, skel := range d.skeletons {
					skel.SetAnimated(v)
					if v {
						skel.Start()
					} else {
						skel.Stop()
					}
				}
			},
			true,
		),
	}

	return d.container
}

const skeletonCode = `package main

import "github.com/atterpac/dado/components"

// Text skeleton (loading text placeholder)
textSkel := components.NewSkeleton().
    SetVariant(components.SkeletonText).
    SetLines(3)  // Number of text lines

// Block skeleton (loading card/image)
blockSkel := components.NewSkeleton().
    SetVariant(components.SkeletonBlock)

// Circle skeleton (loading avatar)
circleSkel := components.NewSkeleton().
    SetVariant(components.SkeletonCircle)

// Start animation
skel.Start()

// Stop when content loads
skel.Stop()

// Disable animation
skel.SetAnimated(false)
`
