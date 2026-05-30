package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/theme"
)

// Sidebar is the navigation sidebar showing component categories.
type Sidebar struct {
	*components.Panel
	tree        *components.Tree
	onSelect    func(demo demos.Demo)
	onHighlight func(demo demos.Demo)
	demoMap     map[string]demos.Demo
}

// NewSidebar creates a new sidebar component.
func NewSidebar() *Sidebar {
	s := &Sidebar{
		Panel:   components.NewPanel(),
		tree:    components.NewTree(),
		demoMap: make(map[string]demos.Demo),
	}

	s.Panel.SetTitle("Components")
	s.Panel.SetTitleAlign(components.AlignLeft)
	s.Panel.SetContent(s.tree)

	// Configure tree
	s.tree.SetShowLines(true)
	s.tree.SetShowIcons(true)
	s.tree.SetIndentSize(2)

	// Handle node selection (Enter key)
	s.tree.SetOnSelect(func(node *components.TreeNode) {
		if node.IsLeaf() {
			if demo, ok := s.demoMap[node.ID]; ok && s.onSelect != nil {
				s.onSelect(demo)
			}
		}
	})

	// Handle node highlighting (navigation)
	s.tree.SetOnHighlight(func(node *components.TreeNode) {
		if node.IsLeaf() {
			if demo, ok := s.demoMap[node.ID]; ok && s.onHighlight != nil {
				s.onHighlight(demo)
			}
		}
	})

	return s
}

// SetOnSelect sets the callback for when a demo is selected (Enter key).
func (s *Sidebar) SetOnSelect(fn func(demo demos.Demo)) {
	s.onSelect = fn
}

// SetOnHighlight sets the callback for when a demo is highlighted (navigation).
func (s *Sidebar) SetOnHighlight(fn func(demo demos.Demo)) {
	s.onHighlight = fn
}

// PopulateFromRegistry populates the sidebar from the demo registry.
func (s *Sidebar) PopulateFromRegistry(registry *demos.Registry) {
	root := &components.TreeNode{
		ID:       "root",
		Label:    "Components",
		Icon:     theme.IconFolder,
		Expanded: true,
	}

	// Create category nodes
	categories := map[demos.Category]*components.TreeNode{
		demos.Basic: {
			ID:       "basic",
			Label:    "Basic",
			Icon:     theme.IconFolder,
			Expanded: true,
		},
		demos.Intermediate: {
			ID:       "intermediate",
			Label:    "Intermediate",
			Icon:     theme.IconFolder,
			Expanded: true,
		},
		demos.Advanced: {
			ID:       "advanced",
			Label:    "Advanced",
			Icon:     theme.IconFolder,
			Expanded: true,
		},
	}

	// Add category nodes to root
	root.AddChild(categories[demos.Basic])
	root.AddChild(categories[demos.Intermediate])
	root.AddChild(categories[demos.Advanced])

	// Add demos to categories
	for _, demo := range registry.All() {
		catNode := categories[demo.Category()]
		if catNode != nil {
			demoNode := &components.TreeNode{
				ID:    demo.Name(),
				Label: demo.Name(),
				Icon:  theme.IconFile,
			}
			catNode.AddChild(demoNode)
			s.demoMap[demo.Name()] = demo
		}
	}

	s.tree.SetRoot(root)
	s.tree.ExpandAll()
}

// Draw renders the sidebar.
func (s *Sidebar) Draw(screen tcell.Screen) {
	s.Panel.Draw(screen)
}

// InputHandler returns the input handler for the sidebar.
func (s *Sidebar) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return s.tree.InputHandler()
}

// Focus handles focus for the sidebar.
func (s *Sidebar) Focus(delegate func(tview.Primitive)) {
	delegate(s.tree)
}

// HasFocus returns whether the sidebar has focus.
func (s *Sidebar) HasFocus() bool {
	return s.tree.HasFocus()
}

// MouseHandler returns the mouse handler for the sidebar.
func (s *Sidebar) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return s.tree.MouseHandler()
}
