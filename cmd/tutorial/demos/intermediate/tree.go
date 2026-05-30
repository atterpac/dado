package intermediate

import (
	"github.com/rivo/tview"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/theme"
)

func init() {
	demos.Register(&TreeDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "Tree",
			DemoDescription: "Collapsible tree view",
			DemoCategory:    demos.Intermediate,
			DemoCode:        treeCode,
		},
	})
}

// TreeDemo demonstrates the Tree component.
type TreeDemo struct {
	demos.DemoBase
	tree      *components.Tree
	showLines bool
	showIcons bool
}

// Component returns the demo component.
func (d *TreeDemo) Component() tview.Primitive {
	d.showLines = true
	d.showIcons = true

	d.tree = components.NewTree().
		SetShowLines(d.showLines).
		SetShowIcons(d.showIcons)

	// Build a sample tree structure
	root := &components.TreeNode{
		ID:       "root",
		Label:    "Project",
		Icon:     theme.IconFolder,
		Expanded: true,
	}

	src := &components.TreeNode{ID: "src", Label: "src", Icon: theme.IconFolder, Expanded: true}
	src.AddChild(&components.TreeNode{ID: "main", Label: "main.go", Icon: theme.IconFile})
	src.AddChild(&components.TreeNode{ID: "utils", Label: "utils.go", Icon: theme.IconFile})

	pkg := &components.TreeNode{ID: "pkg", Label: "pkg", Icon: theme.IconFolder}
	pkg.AddChild(&components.TreeNode{ID: "api", Label: "api.go", Icon: theme.IconFile})
	pkg.AddChild(&components.TreeNode{ID: "models", Label: "models.go", Icon: theme.IconFile})

	root.AddChild(src)
	root.AddChild(pkg)
	root.AddChild(&components.TreeNode{ID: "readme", Label: "README.md", Icon: theme.IconFile})
	root.AddChild(&components.TreeNode{ID: "go.mod", Label: "go.mod", Icon: theme.IconFile})

	d.tree.SetRoot(root)

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("showLines", "Show connecting lines",
			func() bool { return d.showLines },
			func(v bool) { d.showLines = v; d.tree.SetShowLines(v) },
			true,
		),
		demos.BoolProp("showIcons", "Show node icons",
			func() bool { return d.showIcons },
			func(v bool) { d.showIcons = v; d.tree.SetShowIcons(v) },
			true,
		),
	}

	return d.tree
}

const treeCode = `package main

import (
    "github.com/atterpac/dado/components"
    "github.com/atterpac/dado/theme"
)

// Create tree
tree := components.NewTree().
    SetShowLines(true).
    SetShowIcons(true).
    SetIndentSize(2)

// Build nodes
root := &components.TreeNode{
    ID:       "root",
    Label:    "Project",
    Icon:     theme.IconFolder,
    Expanded: true,
}

child := &components.TreeNode{
    ID:    "child",
    Label: "main.go",
    Icon:  theme.IconFile,
}
root.AddChild(child)

tree.SetRoot(root)

// Callbacks
tree.SetOnSelect(func(node *components.TreeNode) {
    fmt.Printf("Selected: %s\n", node.Label)
})

tree.SetOnExpand(func(node *components.TreeNode) {
    fmt.Printf("Expanded: %s\n", node.Label)
})

// Lazy loading for large trees
tree.SetLazyLoader(func(node *components.TreeNode) []*components.TreeNode {
    return loadChildrenFromDisk(node.ID)
})
`
