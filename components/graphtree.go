package components

import (
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"

	"github.com/atterpac/dado/theme"
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
	widgetBase

	data          *GraphTreeData
	flatNodes     []*GraphTreeNode // flattened visible nodes for rendering
	selectedIndex int
	offset        int // scroll offset

	showEdgeLabels bool
	indentSize     int
	prefixBuf      []rune // reused line-prefix scratch (Draw only)

	// Callbacks
	onChange       func(node *GraphTreeNode)
	onSelect       func(node *GraphTreeNode)
	onLoadChildren func(nodeID string) ([]*GraphTreeNode, []*GraphTreeEdge)
}

// NewGraphTree creates a new GraphTree component.
func NewGraphTree() *GraphTree {
	t := &GraphTree{
		indentSize: 3,
	}
	t.initWidget()
	return t
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
		isSelected := t.offset+i == t.selectedIndex

		// Determine base style
		style := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		if isSelected {
			style = style.Background(selectionBg).Foreground(selectionFg).Bold(true)
		}

		// Clear row
		fillLine(screen, x, rowY, width, style)

		col := x

		// Draw indent with tree lines
		prefix := t.linePrefix(node, t.offset+i)
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
			indicatorStyle = tcell.StyleDefault.Background(bgColor).Foreground(th.Accent())
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

// linePrefix fills the reused buffer with node's indentation (spaces) plus the
// two edge connector runes at its tail, returning it. The reused []rune avoids
// the per-node Repeat + []rune conversions + string() this used to allocate.
func (t *GraphTree) linePrefix(node *GraphTreeNode, flatIndex int) []rune {
	n := node.Depth * t.indentSize
	if cap(t.prefixBuf) < n {
		t.prefixBuf = make([]rune, n)
	}
	buf := t.prefixBuf[:n]
	for i := range buf {
		buf[i] = ' '
	}
	if n == 0 {
		return buf
	}

	// Determine edge connector (always two runes) from the parent edge type.
	var edgeChars string
	if flatIndex > 0 {
		parentNode := t.findParentNode(node)
		if parentNode != nil {
			if edge := t.data.GetEdge(parentNode.ID, node.ID); edge != nil {
				isLast := t.isLastChild(parentNode, node.ID)
				switch edge.Type {
				case GraphEdgeSolid:
					edgeChars = pick(isLast, "└─", "├─")
				case GraphEdgeDashed:
					edgeChars = pick(isLast, "└╌", "├╌")
				case GraphEdgeDotted:
					edgeChars = pick(isLast, "└·", "├·")
				}
			}
		}
	}

	if edgeChars != "" && n >= 2 {
		r0, sz := utf8.DecodeRuneInString(edgeChars)
		r1, _ := utf8.DecodeRuneInString(edgeChars[sz:])
		buf[n-2] = r0
		buf[n-1] = r1
	}
	return buf
}

func pick(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
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

// statusEqual compares status case-insensitively without the strings.ToLower
// allocation (this runs per node per frame).
func statusEqual(status string, names ...string) bool {
	for _, n := range names {
		if strings.EqualFold(status, n) {
			return true
		}
	}
	return false
}

func (t *GraphTree) getStatusColor(status string) tcell.Color {
	th := t.th()
	switch {
	case statusEqual(status, "running"):
		return th.Info()
	case statusEqual(status, "completed"):
		return th.Success()
	case statusEqual(status, "failed", "terminated"):
		return th.Error()
	case statusEqual(status, "canceled", "cancelled", "timedout", "timed_out"):
		return th.Warning()
	default:
		return th.FgDim()
	}
}

func (t *GraphTree) getStatusIcon(status string) string {
	switch {
	case statusEqual(status, "running"):
		return "●"
	case statusEqual(status, "completed"):
		return "✓"
	case statusEqual(status, "failed"):
		return "✗"
	case statusEqual(status, "canceled", "cancelled"):
		return "⊘"
	case statusEqual(status, "terminated"):
		return "⊗"
	case statusEqual(status, "timedout", "timed_out"):
		return "⏱"
	default:
		return "○"
	}
}

// HandleKey handles keyboard input.
func (t *GraphTree) HandleKey(ev *tcell.EventKey) bool {
	if len(t.flatNodes) == 0 {
		return false
	}

	prevIndex := t.selectedIndex

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
	return false
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

// Focus handles focus.
// HasFocus returns whether the tree has focus.
func (t *GraphTree) HasFocus() bool {
	return t.Box.HasFocus()
}
