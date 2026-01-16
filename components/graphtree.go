package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// GraphNodeType defines the visual type of a node.
type GraphNodeType int

const (
	GraphNodePrimary   GraphNodeType = iota // Main/focused workflow
	GraphNodeSecondary                      // Child or related workflow
	GraphNodeLink                           // Signal or continue-as-new link
)

// GraphEdgeType defines the line style for edges.
type GraphEdgeType int

const (
	GraphEdgeSolid  GraphEdgeType = iota // Parent-child relationship
	GraphEdgeDashed                      // Signal relationship
	GraphEdgeDotted                      // Continue-as-new relationship
)

// GraphTreeNode represents a node in the workflow graph tree.
type GraphTreeNode struct {
	ID        string
	Label     string
	Sublabel  string        // Secondary text (e.g., workflow type)
	Status    string        // Status for coloring (e.g., "Running", "Completed")
	NodeType  GraphNodeType // Visual type
	Children  []string      // Child node IDs
	CanExpand bool          // Whether node can have children loaded
	Expanded  bool          // Whether children are visible
	Loading   bool          // Show loading indicator
	Depth     int           // Depth in tree
	Data      any           // Associated data
}

// GraphTreeEdge represents an edge between nodes.
type GraphTreeEdge struct {
	From  string
	To    string
	Type  GraphEdgeType
	Label string // Optional edge label
}

// GraphTreeData holds the data for a GraphTree.
type GraphTreeData struct {
	nodes  map[string]*GraphTreeNode
	edges  map[string]*GraphTreeEdge // keyed by "from:to"
	RootID string
}

// NewGraphTreeData creates a new GraphTreeData container.
func NewGraphTreeData() *GraphTreeData {
	return &GraphTreeData{
		nodes: make(map[string]*GraphTreeNode),
		edges: make(map[string]*GraphTreeEdge),
	}
}

// AddNode adds a node to the data.
func (d *GraphTreeData) AddNode(node *GraphTreeNode) {
	d.nodes[node.ID] = node
}

// GetNode retrieves a node by ID.
func (d *GraphTreeData) GetNode(id string) *GraphTreeNode {
	return d.nodes[id]
}

// AddEdge adds an edge to the data.
func (d *GraphTreeData) AddEdge(edge *GraphTreeEdge) {
	key := edge.From + ":" + edge.To
	d.edges[key] = edge
}

// GetEdge retrieves an edge between two nodes.
func (d *GraphTreeData) GetEdge(from, to string) *GraphTreeEdge {
	return d.edges[from+":"+to]
}

// GraphTree is a tree view specialized for displaying workflow relationships.
type GraphTree struct {
	*tview.Box

	data          *GraphTreeData
	flatNodes     []*GraphTreeNode // flattened visible nodes for rendering
	selectedIndex int
	offset        int // scroll offset

	showEdgeLabels bool
	indentSize     int

	// Callbacks
	onChange      func(node *GraphTreeNode)
	onSelect      func(node *GraphTreeNode)
	onLoadChildren func(nodeID string) ([]*GraphTreeNode, []*GraphTreeEdge)
}

// NewGraphTree creates a new GraphTree component.
func NewGraphTree() *GraphTree {
	return &GraphTree{
		Box:        tview.NewBox(),
		indentSize: 3,
	}
}

// SetData sets the tree data.
func (t *GraphTree) SetData(data *GraphTreeData) *GraphTree {
	t.data = data
	t.rebuildFlatList()
	return t
}

// SetShowEdgeLabels enables/disables edge label display.
func (t *GraphTree) SetShowEdgeLabels(show bool) *GraphTree {
	t.showEdgeLabels = show
	return t
}

// SetOnChange sets the callback for selection changes.
func (t *GraphTree) SetOnChange(fn func(node *GraphTreeNode)) *GraphTree {
	t.onChange = fn
	return t
}

// SetOnSelect sets the callback for node selection (Enter).
func (t *GraphTree) SetOnSelect(fn func(node *GraphTreeNode)) *GraphTree {
	t.onSelect = fn
	return t
}

// SetOnLoadChildren sets the lazy loading callback.
func (t *GraphTree) SetOnLoadChildren(fn func(nodeID string) ([]*GraphTreeNode, []*GraphTreeEdge)) *GraphTree {
	t.onLoadChildren = fn
	return t
}

// GetSelected returns the currently selected node.
func (t *GraphTree) GetSelected() *GraphTreeNode {
	if t.selectedIndex >= 0 && t.selectedIndex < len(t.flatNodes) {
		return t.flatNodes[t.selectedIndex]
	}
	return nil
}

// rebuildFlatList flattens the tree for rendering.
func (t *GraphTree) rebuildFlatList() {
	t.flatNodes = nil
	if t.data == nil || t.data.RootID == "" {
		return
	}
	t.flattenNode(t.data.RootID, 0)
}

func (t *GraphTree) flattenNode(nodeID string, depth int) {
	node := t.data.GetNode(nodeID)
	if node == nil {
		return
	}
	node.Depth = depth
	t.flatNodes = append(t.flatNodes, node)
	if node.Expanded {
		for _, childID := range node.Children {
			t.flattenNode(childID, depth+1)
		}
	}
}

// Draw renders the tree.
func (t *GraphTree) Draw(screen tcell.Screen) {
	t.Box.DrawForSubclass(screen, t)
	x, y, width, height := t.GetInnerRect()

	if width <= 0 || height <= 0 || len(t.flatNodes) == 0 {
		return
	}

	// Get colors at draw time
	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
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
		isSelected := t.offset+i == t.selectedIndex

		// Determine base style
		style := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		if isSelected {
			style = style.Background(selectionBg).Foreground(selectionFg).Bold(true)
		}

		// Clear row
		for col := x; col < x+width; col++ {
			screen.SetContent(col, rowY, ' ', nil, style)
		}

		col := x

		// Draw indent with tree lines
		prefix := t.buildLinePrefix(node, t.offset+i)
		prefixStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
		if isSelected {
			prefixStyle = style
		}
		for _, r := range prefix {
			if col < x+width {
				screen.SetContent(col, rowY, r, nil, prefixStyle)
				col++
			}
		}

		// Draw node type indicator
		var indicator string
		switch node.NodeType {
		case GraphNodePrimary:
			indicator = "◉ "
		case GraphNodeSecondary:
			indicator = "◆ "
		case GraphNodeLink:
			indicator = "┆ "
		}

		indicatorStyle := prefixStyle
		if !isSelected {
			indicatorStyle = tcell.StyleDefault.Background(bgColor).Foreground(theme.Accent())
		}
		for _, r := range indicator {
			if col < x+width {
				screen.SetContent(col, rowY, r, nil, indicatorStyle)
				col++
			}
		}

		// Draw expand/collapse indicator if node can expand
		if node.CanExpand && len(node.Children) > 0 {
			var expInd string
			if node.Expanded {
				expInd = theme.IconChevronD + " "
			} else {
				expInd = theme.IconChevronR + " "
			}
			for _, r := range expInd {
				if col < x+width {
					screen.SetContent(col, rowY, r, nil, prefixStyle)
					col++
				}
			}
		} else if node.Loading {
			for _, r := range "⏳ " {
				if col < x+width {
					screen.SetContent(col, rowY, r, nil, prefixStyle)
					col++
				}
			}
		} else if node.CanExpand {
			for _, r := range theme.IconChevronR + " " {
				if col < x+width {
					screen.SetContent(col, rowY, r, nil, prefixStyle)
					col++
				}
			}
		}

		// Draw label with status color
		statusColor := t.getStatusColor(node.Status)
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(statusColor)
		if isSelected {
			labelStyle = style.Foreground(statusColor)
		}
		for _, r := range node.Label {
			if col < x+width {
				screen.SetContent(col, rowY, r, nil, labelStyle)
				col++
			}
		}

		// Draw status indicator at end of row
		if node.Status != "" {
			statusStyle := tcell.StyleDefault.Background(bgColor).Foreground(statusColor)
			if isSelected {
				statusStyle = style.Foreground(statusColor)
			}
			statusText := " " + t.getStatusIcon(node.Status)
			startCol := x + width - len([]rune(statusText))
			if startCol > col {
				col = startCol
			}
			for _, r := range statusText {
				if col < x+width {
					screen.SetContent(col, rowY, r, nil, statusStyle)
					col++
				}
			}
		}

		_ = fgDimColor // Suppress unused warning
	}
}

func (t *GraphTree) buildLinePrefix(node *GraphTreeNode, flatIndex int) string {
	if node.Depth == 0 {
		return ""
	}

	prefix := strings.Repeat(" ", node.Depth*t.indentSize)

	// Determine edge type from parent
	var edgeChars string
	if flatIndex > 0 {
		parentNode := t.findParentNode(node)
		if parentNode != nil {
			edge := t.data.GetEdge(parentNode.ID, node.ID)
			if edge != nil {
				isLast := t.isLastChild(parentNode, node.ID)
				switch edge.Type {
				case GraphEdgeSolid:
					if isLast {
						edgeChars = "└─"
					} else {
						edgeChars = "├─"
					}
				case GraphEdgeDashed:
					if isLast {
						edgeChars = "└╌"
					} else {
						edgeChars = "├╌"
					}
				case GraphEdgeDotted:
					if isLast {
						edgeChars = "└·"
					} else {
						edgeChars = "├·"
					}
				}
			}
		}
	}

	if edgeChars != "" {
		// Replace last part of prefix with edge chars
		prefixRunes := []rune(prefix)
		edgeRunes := []rune(edgeChars)
		if len(prefixRunes) >= len(edgeRunes) {
			copy(prefixRunes[len(prefixRunes)-len(edgeRunes):], edgeRunes)
			prefix = string(prefixRunes)
		}
	}

	return prefix
}

func (t *GraphTree) findParentNode(child *GraphTreeNode) *GraphTreeNode {
	if t.data == nil {
		return nil
	}
	for _, node := range t.data.nodes {
		for _, childID := range node.Children {
			if childID == child.ID {
				return node
			}
		}
	}
	return nil
}

func (t *GraphTree) isLastChild(parent *GraphTreeNode, childID string) bool {
	if len(parent.Children) == 0 {
		return true
	}
	return parent.Children[len(parent.Children)-1] == childID
}

func (t *GraphTree) getStatusColor(status string) tcell.Color {
	switch strings.ToLower(status) {
	case "running":
		return theme.Info()
	case "completed":
		return theme.Success()
	case "failed":
		return theme.Error()
	case "canceled", "cancelled":
		return theme.Warning()
	case "terminated":
		return theme.Error()
	case "timedout", "timed_out":
		return theme.Warning()
	default:
		return theme.FgDim()
	}
}

func (t *GraphTree) getStatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "running":
		return "●"
	case "completed":
		return "✓"
	case "failed":
		return "✗"
	case "canceled", "cancelled":
		return "⊘"
	case "terminated":
		return "⊗"
	case "timedout", "timed_out":
		return "⏱"
	default:
		return "○"
	}
}

// InputHandler handles keyboard input.
func (t *GraphTree) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return t.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if len(t.flatNodes) == 0 {
			return
		}

		prevIndex := t.selectedIndex

		switch event.Key() {
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
			switch event.Rune() {
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
			case ' ', 'o':
				t.toggleExpanded()
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
		}

		// Call onChange if the selected index changed
		if t.selectedIndex != prevIndex && t.onChange != nil {
			if node := t.GetSelected(); node != nil {
				t.onChange(node)
			}
		}
	})
}

func (t *GraphTree) moveDown() {
	if t.selectedIndex < len(t.flatNodes)-1 {
		t.selectedIndex++
	}
}

func (t *GraphTree) moveUp() {
	if t.selectedIndex > 0 {
		t.selectedIndex--
	}
}

func (t *GraphTree) expandOrMoveIn() {
	node := t.GetSelected()
	if node == nil {
		return
	}

	if !node.CanExpand {
		return
	}

	if !node.Expanded {
		// Lazy load if needed
		if t.onLoadChildren != nil && len(node.Children) == 0 {
			nodes, edges := t.onLoadChildren(node.ID)
			for _, n := range nodes {
				t.data.AddNode(n)
				node.Children = append(node.Children, n.ID)
			}
			for _, e := range edges {
				t.data.AddEdge(e)
			}
		}
		node.Expanded = true
		t.rebuildFlatList()
	} else if len(node.Children) > 0 {
		// Move to first child
		t.selectedIndex++
	}
}

func (t *GraphTree) collapseOrMoveOut() {
	node := t.GetSelected()
	if node == nil {
		return
	}

	if node.Expanded && node.CanExpand {
		node.Expanded = false
		t.rebuildFlatList()
	} else {
		// Move to parent
		parent := t.findParentNode(node)
		if parent != nil {
			for i, n := range t.flatNodes {
				if n == parent {
					t.selectedIndex = i
					break
				}
			}
		}
	}
}

func (t *GraphTree) toggleExpanded() {
	node := t.GetSelected()
	if node == nil || !node.CanExpand {
		return
	}

	if node.Expanded {
		node.Expanded = false
	} else {
		// Lazy load if needed
		if t.onLoadChildren != nil && len(node.Children) == 0 {
			nodes, edges := t.onLoadChildren(node.ID)
			for _, n := range nodes {
				t.data.AddNode(n)
				node.Children = append(node.Children, n.ID)
			}
			for _, e := range edges {
				t.data.AddEdge(e)
			}
		}
		node.Expanded = true
	}
	t.rebuildFlatList()
}

// MouseHandler handles mouse input.
func (t *GraphTree) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return t.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		_, y, _, _ := t.GetInnerRect()
		mx, my := event.Position()

		if !t.InRect(mx, my) {
			return false, nil
		}

		switch action {
		case tview.MouseLeftClick:
			setFocus(t)
			clickedIndex := t.offset + (my - y)
			if clickedIndex >= 0 && clickedIndex < len(t.flatNodes) {
				prevIndex := t.selectedIndex
				t.selectedIndex = clickedIndex
				if t.selectedIndex != prevIndex && t.onChange != nil {
					if node := t.GetSelected(); node != nil {
						t.onChange(node)
					}
				}
				return true, t
			}
		case tview.MouseLeftDoubleClick:
			clickedIndex := t.offset + (my - y)
			if clickedIndex >= 0 && clickedIndex < len(t.flatNodes) {
				t.selectedIndex = clickedIndex
				node := t.flatNodes[clickedIndex]
				if node.CanExpand {
					t.toggleExpanded()
				} else if t.onSelect != nil {
					t.onSelect(node)
				}
				return true, t
			}
		case tview.MouseScrollUp:
			if t.offset > 0 {
				t.offset--
			}
			return true, t
		case tview.MouseScrollDown:
			_, _, _, height := t.GetInnerRect()
			if t.offset < len(t.flatNodes)-height {
				t.offset++
			}
			return true, t
		}

		return false, nil
	})
}

// Focus handles focus.
func (t *GraphTree) Focus(delegate func(tview.Primitive)) {
	t.Box.Focus(delegate)
}

// HasFocus returns whether the tree has focus.
func (t *GraphTree) HasFocus() bool {
	return t.Box.HasFocus()
}
