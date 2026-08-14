package inline

import (
	"sync"
)

// TreeNode is one item in a renderer-native Tree. Detail is drawn after Label
// with the theme's muted style.
type TreeNode struct {
	Label    string
	Detail   string
	Children []TreeNode
}

// Tree renders a hierarchy into a Frame. Its mutation and rendering methods
// are safe to call from separate goroutines.
type Tree struct {
	mu    sync.RWMutex
	title string
	nodes []TreeNode
	theme InlineTheme
}

// NewTree creates a tree with an optional title and root nodes.
func NewTree(title string, nodes ...TreeNode) *Tree {
	return &Tree{title: title, nodes: cloneTreeNodes(nodes), theme: RoundedInlineTheme()}
}

// SetTheme changes the tree's visual preset.
func (t *Tree) SetTheme(theme InlineTheme) *Tree {
	t.mu.Lock()
	t.theme = normalizedInlineTheme(theme)
	t.mu.Unlock()
	return t
}

// SetNodes replaces all root nodes.
func (t *Tree) SetNodes(nodes ...TreeNode) *Tree {
	t.mu.Lock()
	t.nodes = cloneTreeNodes(nodes)
	t.mu.Unlock()
	return t
}

// Add appends a root node.
func (t *Tree) Add(node TreeNode) *Tree {
	t.mu.Lock()
	t.nodes = append(t.nodes, cloneTreeNode(node))
	t.mu.Unlock()
	return t
}

// Frame returns the tree's current renderer frame.
func (t *Tree) Frame(width int) *Frame {
	width = max(width, 0)
	t.mu.RLock()
	title := t.title
	nodes := cloneTreeNodes(t.nodes)
	theme := t.theme
	t.mu.RUnlock()

	height := treeNodeCount(nodes)
	if title != "" {
		height++
	}
	frame := NewFrame(width, height)
	y := 0
	if title != "" {
		drawClipped(frame, 0, y, title, width, theme.Accent.Bold(true))
		y++
	}
	drawTreeNodes(frame, nodes, nil, y, width, theme)
	return frame
}

func drawTreeNodes(frame *Frame, nodes []TreeNode, ancestors []bool, y, width int, theme InlineTheme) int {
	for index, node := range nodes {
		last := index == len(nodes)-1
		x := 0
		for _, ancestorLast := range ancestors {
			guide := theme.Tree.Vertical
			if ancestorLast {
				guide = theme.Tree.Space
			}
			drawClipped(frame, x, y, guide, max(width-x, 0), theme.Border)
			x += displayWidth(guide)
		}
		branch := theme.Tree.Branch
		if last {
			branch = theme.Tree.Last
		}
		drawClipped(frame, x, y, branch, max(width-x, 0), theme.Border)
		x += displayWidth(branch) + 1
		drawClipped(frame, x, y, node.Label, max(width-x, 0), theme.Text)
		x += displayWidth(node.Label)
		if node.Detail != "" && x+1 < width {
			drawClipped(frame, x+1, y, node.Detail, width-x-1, theme.Muted)
		}
		y++
		y = drawTreeNodes(frame, node.Children, append(ancestors, last), y, width, theme)
	}
	return y
}

func treeNodeCount(nodes []TreeNode) int {
	count := len(nodes)
	for _, node := range nodes {
		count += treeNodeCount(node.Children)
	}
	return count
}

func cloneTreeNodes(nodes []TreeNode) []TreeNode {
	cloned := make([]TreeNode, len(nodes))
	for index, node := range nodes {
		cloned[index] = cloneTreeNode(node)
	}
	return cloned
}

func cloneTreeNode(node TreeNode) TreeNode {
	node.Children = cloneTreeNodes(node.Children)
	return node
}
