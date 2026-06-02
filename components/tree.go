package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/theme"
)

// TreeNode is a single node in a Tree. Set Expanded to control initial
// open/closed state. Color 0 uses the default foreground. Data holds
// arbitrary caller data returned by onSelect and onHighlight callbacks.
type TreeNode struct {
	ID       string
	Label    string
	Icon     string
	Color    tcell.Color // optional label color (zero value = default)
	Children []*TreeNode
	Data     any
	Expanded bool
	parent   *TreeNode
	level    int
}

// AddChild adds a child node.
func (n *TreeNode) AddChild(child *TreeNode) *TreeNode {
	child.parent = n
	child.level = n.level + 1
	n.Children = append(n.Children, child)
	return n
}

// IsLeaf returns true if the node has no children.
func (n *TreeNode) IsLeaf() bool {
	return len(n.Children) == 0
}

// Tree is a collapsible tree view component.
type Tree struct {
	widgetBase

	root          *TreeNode
	flatNodes     []*TreeNode // flattened visible nodes for rendering
	selectedIndex int
	offset        int // scroll offset

	showLines  bool
	showIcons  bool
	indentSize int

	// Callbacks
	onSelect    func(node *TreeNode)
	onHighlight func(node *TreeNode)
	onExpand    func(node *TreeNode)
	onCollapse  func(node *TreeNode)
	lazyLoader  func(node *TreeNode) []*TreeNode

	// Multi-select
	multiSelect bool
	selected    map[*TreeNode]bool
}

// NewTree creates a new Tree component.
func NewTree() *Tree {
	t := &Tree{
		showLines:  true,
		showIcons:  true,
		indentSize: 2,
		selected:   make(map[*TreeNode]bool),
	}
	t.initWidget()
	return t
}

// SetRoot sets the root node of the tree.
func (t *Tree) SetRoot(root *TreeNode) *Tree {
	t.root = root
	t.selectedIndex = 0
	t.offset = 0
	if root != nil {
		root.level = 0
		t.setLevels(root, 0)
	}
	t.rebuildFlatList()
	return t
}

func (t *Tree) setLevels(node *TreeNode, level int) {
	node.level = level
	for _, child := range node.Children {
		child.parent = node
		t.setLevels(child, level+1)
	}
}

// SetShowLines enables/disables tree line drawing.
func (t *Tree) SetShowLines(show bool) *Tree {
	t.showLines = show
	return t
}

// SetShowIcons enables/disables node icons.
func (t *Tree) SetShowIcons(show bool) *Tree {
	t.showIcons = show
	return t
}

// SetIndentSize sets the indentation per level.
func (t *Tree) SetIndentSize(size int) *Tree {
	t.indentSize = size
	return t
}

// SetMultiSelect enables/disables multi-selection.
func (t *Tree) SetMultiSelect(enable bool) *Tree {
	t.multiSelect = enable
	return t
}

// SetOnSelect sets the callback for when a node is selected (Enter).
func (t *Tree) SetOnSelect(fn func(node *TreeNode)) *Tree {
	t.onSelect = fn
	return t
}

// SetOnHighlight sets the callback for when the highlighted node changes (j/k navigation).
func (t *Tree) SetOnHighlight(fn func(node *TreeNode)) *Tree {
	t.onHighlight = fn
	return t
}

// SetOnExpand sets the callback for when a node is expanded.
func (t *Tree) SetOnExpand(fn func(node *TreeNode)) *Tree {
	t.onExpand = fn
	return t
}

// SetOnCollapse sets the callback for when a node is collapsed.
func (t *Tree) SetOnCollapse(fn func(node *TreeNode)) *Tree {
	t.onCollapse = fn
	return t
}

// SetLazyLoader sets a function to load children on demand.
func (t *Tree) SetLazyLoader(fn func(node *TreeNode) []*TreeNode) *Tree {
	t.lazyLoader = fn
	return t
}

// rebuildFlatList flattens the tree for rendering.
func (t *Tree) rebuildFlatList() {
	t.flatNodes = nil
	if t.root != nil {
		t.flattenNode(t.root)
	}
}

func (t *Tree) flattenNode(node *TreeNode) {
	t.flatNodes = append(t.flatNodes, node)
	if node.Expanded {
		for _, child := range node.Children {
			t.flattenNode(child)
		}
	}
}

// ExpandAll expands all nodes.
func (t *Tree) ExpandAll() *Tree {
	if t.root != nil {
		t.expandAllRecursive(t.root)
	}
	t.rebuildFlatList()
	return t
}

func (t *Tree) expandAllRecursive(node *TreeNode) {
	if !node.IsLeaf() {
		node.Expanded = true
		for _, child := range node.Children {
			t.expandAllRecursive(child)
		}
	}
}

// CollapseAll collapses all nodes.
func (t *Tree) CollapseAll() *Tree {
	if t.root != nil {
		t.collapseAllRecursive(t.root)
	}
	t.rebuildFlatList()
	return t
}

func (t *Tree) collapseAllRecursive(node *TreeNode) {
	node.Expanded = false
	for _, child := range node.Children {
		t.collapseAllRecursive(child)
	}
}

// ExpandTo expands nodes to the specified depth.
func (t *Tree) ExpandTo(depth int) *Tree {
	if t.root != nil {
		t.expandToDepth(t.root, depth)
	}
	t.rebuildFlatList()
	return t
}

func (t *Tree) expandToDepth(node *TreeNode, depth int) {
	if node.level < depth && !node.IsLeaf() {
		node.Expanded = true
		for _, child := range node.Children {
			t.expandToDepth(child, depth)
		}
	}
}

// GetSelected returns the currently highlighted node.
func (t *Tree) GetSelected() *TreeNode {
	if t.selectedIndex >= 0 && t.selectedIndex < len(t.flatNodes) {
		return t.flatNodes[t.selectedIndex]
	}
	return nil
}

// GetSelectedNodes returns all multi-selected nodes.
func (t *Tree) GetSelectedNodes() []*TreeNode {
	var nodes []*TreeNode
	for node := range t.selected {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetSelectedIndex returns the current cursor index in the flat node list.
func (t *Tree) GetSelectedIndex() int {
	return t.selectedIndex
}

// SetSelectedIndex sets the cursor to the given index in the flat node list.
// The index is clamped to valid bounds. onHighlight fires only when the index
// actually changes. The scroll offset is adjusted so the new selection is
// visible on the next draw.
func (t *Tree) SetSelectedIndex(index int) *Tree {
	if len(t.flatNodes) == 0 {
		t.selectedIndex = 0
		t.offset = 0
		return t
	}
	if index < 0 {
		index = 0
	}
	if index >= len(t.flatNodes) {
		index = len(t.flatNodes) - 1
	}

	prev := t.selectedIndex
	t.selectedIndex = index

	// Scroll into view using current inner rect height when available.
	if _, _, _, height := t.GetInnerRect(); height > 0 {
		if t.selectedIndex < t.offset {
			t.offset = t.selectedIndex
		} else if t.selectedIndex >= t.offset+height {
			t.offset = t.selectedIndex - height + 1
		}
	}

	if t.selectedIndex != prev && t.onHighlight != nil {
		t.onHighlight(t.flatNodes[t.selectedIndex])
	}
	return t
}

// ClearSelection clears multi-selection.
func (t *Tree) ClearSelection() *Tree {
	t.selected = make(map[*TreeNode]bool)
	return t
}

// Filter filters the tree by a query string.
func (t *Tree) Filter(query string) *Tree {
	if query == "" {
		t.rebuildFlatList()
		return t
	}

	query = strings.ToLower(query)
	t.flatNodes = nil
	if t.root != nil {
		t.filterNode(t.root, query)
	}
	return t
}

func (t *Tree) filterNode(node *TreeNode, query string) bool {
	matches := strings.Contains(strings.ToLower(node.Label), query)

	// Check children first to determine if any match
	var hasMatchingChild bool
	childStart := len(t.flatNodes)
	for _, child := range node.Children {
		if t.filterNode(child, query) {
			hasMatchingChild = true
		}
	}

	if matches || hasMatchingChild {
		// Insert parent before its children (pre-order)
		children := make([]*TreeNode, len(t.flatNodes)-childStart)
		copy(children, t.flatNodes[childStart:])
		t.flatNodes = append(t.flatNodes[:childStart], node)
		t.flatNodes = append(t.flatNodes, children...)
		return true
	}
	return false
}

// Draw renders the tree.
func (t *Tree) Draw(screen tcell.Screen) {
	t.Box.DrawForSubclass(screen)
	x, y, width, height := t.GetInnerRect()

	if width <= 0 || height <= 0 || len(t.flatNodes) == 0 {
		return
	}

	// Get colors at draw time
	th := t.th()
	bgColor := th.Bg()
	fgColor := th.Fg()
	fgDimColor := th.FgDim()
	selectionBg := theme.SelectionBg()
	selectionFg := theme.SelectionFg()

	// Ensure selected index is valid
	if t.selectedIndex >= len(t.flatNodes) {
		t.selectedIndex = len(t.flatNodes) - 1
	}
	if t.selectedIndex < 0 {
		t.selectedIndex = 0
	}

	// Adjust scroll offset
	if t.selectedIndex < t.offset {
		t.offset = t.selectedIndex
	}
	if t.selectedIndex >= t.offset+height {
		t.offset = t.selectedIndex - height + 1
	}

	// Draw visible nodes
	for i := 0; i < height && t.offset+i < len(t.flatNodes); i++ {
		node := t.flatNodes[t.offset+i]
		rowY := y + i

		// Determine style
		style := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		if t.offset+i == t.selectedIndex {
			style = style.Background(selectionBg).Foreground(selectionFg).Bold(true)
		} else if t.selected[node] {
			style = style.Background(selectionBg).Foreground(selectionFg)
		}

		// Clear row
		fillLine(screen, x, rowY, width, style)

		// Build line prefix
		var prefix string
		if t.showLines && node.level > 0 {
			prefix = t.buildLinePrefix(node)
		} else {
			prefix = strings.Repeat(" ", node.level*t.indentSize)
		}

		// Expand/collapse indicator
		var indicator string
		if !node.IsLeaf() {
			if node.Expanded {
				indicator = theme.IconChevronD + " "
			} else {
				indicator = theme.IconChevronR + " "
			}
		} else {
			indicator = "  "
		}

		// Icon
		var icon string
		if t.showIcons && node.Icon != "" {
			icon = node.Icon + " "
		}

		// Draw line
		col := x
		lineStyle := style
		indicatorStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
		if t.offset+i == t.selectedIndex {
			indicatorStyle = style
		}

		for _, r := range prefix {
			if col < x+width {
				screen.SetContent(col, rowY, r, nil, indicatorStyle)
				col++
			}
		}
		for _, r := range indicator {
			if col < x+width {
				screen.SetContent(col, rowY, r, nil, indicatorStyle)
				col++
			}
		}
		for _, r := range icon {
			if col < x+width {
				iconStyle := lineStyle
				if t.offset+i != t.selectedIndex && !t.selected[node] {
					iconStyle = tcell.StyleDefault.Background(bgColor).Foreground(th.Accent())
				}
				screen.SetContent(col, rowY, r, nil, iconStyle)
				col++
			}
		}
		labelStyle := lineStyle
		if node.Color != 0 && t.offset+i != t.selectedIndex {
			labelStyle = tcell.StyleDefault.Background(bgColor).Foreground(node.Color)
		}
		for _, r := range node.Label {
			if col < x+width {
				screen.SetContent(col, rowY, r, nil, labelStyle)
				col++
			}
		}
	}
}

func (t *Tree) buildLinePrefix(node *TreeNode) string {
	if node.level == 0 {
		return ""
	}

	// Build from current node up to root
	parts := make([]string, node.level)
	current := node

	for i := node.level - 1; i >= 0; i-- {
		parent := current.parent
		if parent == nil {
			parts[i] = strings.Repeat(" ", t.indentSize)
		} else {
			isLast := current == parent.Children[len(parent.Children)-1]
			if i == node.level-1 {
				// Direct connection to this node
				if isLast {
					parts[i] = theme.IconTreeLast + strings.Repeat(" ", t.indentSize-1)
				} else {
					parts[i] = theme.IconTreeBranch + strings.Repeat(" ", t.indentSize-1)
				}
			} else {
				// Vertical line or space
				if isLast {
					parts[i] = strings.Repeat(" ", t.indentSize)
				} else {
					parts[i] = theme.IconTreeVert + strings.Repeat(" ", t.indentSize-1)
				}
			}
		}
		current = parent
	}

	return strings.Join(parts, "")
}

// HandleKey processes a key event for the Tree.
func (t *Tree) HandleKey(ev *tcell.EventKey) bool {
	if len(t.flatNodes) == 0 {
		return false
	}

	prevIndex := t.selectedIndex
	handled := true

	switch ev.Key() {
	case tcell.KeyDown:
		t.moveDown()
	case tcell.KeyUp:
		t.moveUp()
	case tcell.KeyRight:
		t.expandOrMoveIn()
	case tcell.KeyLeft:
		t.collapseOrMoveOut()
	case tcell.KeyHome:
		t.selectedIndex = 0
	case tcell.KeyEnd:
		t.selectedIndex = len(t.flatNodes) - 1
	case tcell.KeyPgDn:
		_, _, _, height := t.GetInnerRect()
		t.selectedIndex += height
		if t.selectedIndex >= len(t.flatNodes) {
			t.selectedIndex = len(t.flatNodes) - 1
		}
	case tcell.KeyPgUp:
		_, _, _, height := t.GetInnerRect()
		t.selectedIndex -= height
		if t.selectedIndex < 0 {
			t.selectedIndex = 0
		}
	case tcell.KeyEnter:
		if node := t.GetSelected(); node != nil && t.onSelect != nil {
			t.onSelect(node)
		}
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'j':
			t.moveDown()
		case 'k':
			t.moveUp()
		case 'l':
			t.expandOrMoveIn()
		case 'h':
			t.collapseOrMoveOut()
		case 'g':
			t.selectedIndex = 0
		case 'G':
			t.selectedIndex = len(t.flatNodes) - 1
		case ' ':
			if t.multiSelect {
				if node := t.GetSelected(); node != nil {
					if t.selected[node] {
						delete(t.selected, node)
					} else {
						t.selected[node] = true
					}
				}
			} else {
				t.toggleExpanded()
			}
		case 'o':
			t.toggleExpanded()
		case 'O':
			t.ExpandAll()
		case 'C':
			t.CollapseAll()
		default:
			handled = false
		}
	case tcell.KeyCtrlD:
		_, _, _, height := t.GetInnerRect()
		t.selectedIndex += height / 2
		if t.selectedIndex >= len(t.flatNodes) {
			t.selectedIndex = len(t.flatNodes) - 1
		}
	case tcell.KeyCtrlU:
		_, _, _, height := t.GetInnerRect()
		t.selectedIndex -= height / 2
		if t.selectedIndex < 0 {
			t.selectedIndex = 0
		}
	default:
		handled = false
	}

	// Call onHighlight if the selected index changed
	if t.selectedIndex != prevIndex && t.onHighlight != nil {
		if node := t.GetSelected(); node != nil {
			t.onHighlight(node)
		}
	}
	return handled
}

func (t *Tree) moveDown() {
	if t.selectedIndex < len(t.flatNodes)-1 {
		t.selectedIndex++
	}
}

func (t *Tree) moveUp() {
	if t.selectedIndex > 0 {
		t.selectedIndex--
	}
}

func (t *Tree) expandOrMoveIn() {
	node := t.GetSelected()
	if node == nil {
		return
	}

	if node.IsLeaf() {
		return
	}

	if !node.Expanded {
		// Lazy load if needed
		if t.lazyLoader != nil && len(node.Children) == 0 {
			children := t.lazyLoader(node)
			for _, child := range children {
				node.AddChild(child)
			}
		}
		node.Expanded = true
		t.rebuildFlatList()
		if t.onExpand != nil {
			t.onExpand(node)
		}
	} else if len(node.Children) > 0 {
		// Move to first child
		t.selectedIndex++
	}
}

func (t *Tree) collapseOrMoveOut() {
	node := t.GetSelected()
	if node == nil {
		return
	}

	if node.Expanded && !node.IsLeaf() {
		node.Expanded = false
		t.rebuildFlatList()
		if t.onCollapse != nil {
			t.onCollapse(node)
		}
	} else if node.parent != nil {
		// Move to parent
		for i, n := range t.flatNodes {
			if n == node.parent {
				t.selectedIndex = i
				break
			}
		}
	}
}

func (t *Tree) toggleExpanded() {
	node := t.GetSelected()
	if node == nil || node.IsLeaf() {
		return
	}

	if node.Expanded {
		node.Expanded = false
		if t.onCollapse != nil {
			t.onCollapse(node)
		}
	} else {
		// Lazy load if needed
		if t.lazyLoader != nil && len(node.Children) == 0 {
			children := t.lazyLoader(node)
			for _, child := range children {
				node.AddChild(child)
			}
		}
		node.Expanded = true
		if t.onExpand != nil {
			t.onExpand(node)
		}
	}
	t.rebuildFlatList()
}

// Focus handles focus.
// HasFocus returns whether the tree has focus.
func (t *Tree) HasFocus() bool {
	return t.Box.HasFocus()
}
