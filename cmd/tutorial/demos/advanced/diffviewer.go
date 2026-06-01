package advanced

import (

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

func init() {
	demos.Register(&DiffViewerDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "DiffViewer",
			DemoDescription: "Git diff display",
			DemoCategory:    demos.Advanced,
			DemoCode:        diffViewerCode,
		},
	})
}

// DiffViewerDemo demonstrates the DiffViewer component.
type DiffViewerDemo struct {
	demos.DemoBase
	viewer          *components.DiffViewer
	showLineNumbers bool
}

// Component returns the demo component.
func (d *DiffViewerDemo) Component() core.Widget {
	d.showLineNumbers = true

	d.viewer = components.NewDiffViewer().
		SetShowLineNumbers(d.showLineNumbers).
		SetTitle("example.go")

	// Sample diff
	diff := `diff --git a/example.go b/example.go
index 1234567..abcdefg 100644
--- a/example.go
+++ b/example.go
@@ -10,7 +10,9 @@ func main() {
     fmt.Println("Hello")
-    fmt.Println("World")
+    fmt.Println("Hello")
+    fmt.Println("World!")
+    fmt.Println("Welcome")
     return nil
 }`

	d.viewer.SetUnifiedDiff(diff)

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("showLineNumbers", "Show line numbers",
			func() bool { return d.showLineNumbers },
			func(v bool) { d.showLineNumbers = v; d.viewer.SetShowLineNumbers(v) },
			true,
		),
	}

	return d.viewer
}

const diffViewerCode = `package main

import (
    "github.com/atterpac/dado/components"
)

// Create diff viewer
viewer := components.NewDiffViewer().
    SetShowLineNumbers(true).
    SetTitle("file.go")

// Load diff from unified diff format
viewer.SetDiffText(diffString)

// Or load from DiffResult
result := &components.DiffResult{
    OldName: "old.go",
    NewName: "new.go",
    Hunks: []components.DiffHunk{
        {
            Header:   "@@ -1,5 +1,6 @@",
            OldStart: 1, OldCount: 5,
            NewStart: 1, NewCount: 6,
            Lines: []components.DiffLine{
                {Type: components.DiffLineContext, Content: "unchanged"},
                {Type: components.DiffLineRemoved, Content: "old line"},
                {Type: components.DiffLineAdded, Content: "new line"},
            },
        },
    },
}
viewer.SetDiff(result)

// Get statistics
stats := viewer.GetStats()
fmt.Printf("+%d -%d\n", stats.Additions, stats.Deletions)

// Navigation callback
viewer.SetOnLineSelect(func(line components.DiffLine) {
    fmt.Printf("Line: %s\n", line.Content)
})
`
