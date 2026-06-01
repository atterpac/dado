package components

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// GraphNode represents a node in the 2D node graph.
type GraphNode struct {
	ID      string
	Label   string
	Status  string // For status-based coloring
	Focused bool   // Whether this is the focused/primary node
	Data    any    // Associated data

	// Layout (computed)
	x, y          int
	width, height int
}

// GraphEdge represents an edge in the 2D node graph.
type GraphEdge struct {
	From  string
	To    string
	Type  GraphEdgeType
	Label string
}

// NodeGraphData holds the data for a NodeGraph.
type NodeGraphData struct {
	nodes   map[string]*GraphNode
	edges   []*GraphEdge
	FocusID string // ID of the focused node
}

// NewNodeGraphData creates a new NodeGraphData container.
func NewNodeGraphData() *NodeGraphData {
	return &NodeGraphData{
		nodes: make(map[string]*GraphNode),
	}
}

// AddNode adds a node to the data.
func (d *NodeGraphData) AddNode(node *GraphNode) {
	d.nodes[node.ID] = node
}

// GetNode retrieves a node by ID.
func (d *NodeGraphData) GetNode(id string) *GraphNode {
	return d.nodes[id]
}

// AddEdge adds an edge to the data.
func (d *NodeGraphData) AddEdge(edge *GraphEdge) {
	d.edges = append(d.edges, edge)
}

// NodeGraph is a 2D node graph visualization component.
type NodeGraph struct {
	widgetBase

	data           *NodeGraphData
	nodeWidth      int
	nodeHeight     int
	showEdgeLabels bool

	// Viewport offset for panning
	offsetX, offsetY int

	// Auto-center flag (centers on first draw when dimensions are available)
	needsCenter bool
	fitAll      bool // center bounding box of all nodes instead of focused node

	// Callbacks
	onSelect func(node *GraphNode)
}

// NewNodeGraph creates a new NodeGraph component.
func NewNodeGraph() *NodeGraph {
	g := &NodeGraph{
		nodeWidth:  18,
		nodeHeight: 3,
	}
	g.initWidget()
	return g
}

// SetData sets the graph data and triggers layout computation. The viewport
// centers on the first draw after this call — on FocusID if set, or on the
// bounding box of all nodes if SetFit(true) was called.
func (g *NodeGraph) SetData(data *NodeGraphData) *NodeGraph {
	g.data = data
	g.computeLayout()
	g.needsCenter = true // Auto-center on next draw
	return g
}

// SetNodeWidth sets the width of nodes.
func (g *NodeGraph) SetNodeWidth(width int) *NodeGraph {
	g.nodeWidth = width
	return g
}

// SetFit controls the initial viewport centering strategy. When true, the
// viewport centers on the bounding box of all nodes. When false (default),
// it centers on NodeGraphData.FocusID instead.
func (g *NodeGraph) SetFit(fit bool) *NodeGraph {
	g.fitAll = fit
	return g
}

// SetShowEdgeLabels enables/disables edge label display.
func (g *NodeGraph) SetShowEdgeLabels(show bool) *NodeGraph {
	g.showEdgeLabels = show
	return g
}

// SetOnSelect sets the callback for node selection.
func (g *NodeGraph) SetOnSelect(fn func(node *GraphNode)) *NodeGraph {
	g.onSelect = fn
	return g
}

// SetFocus sets the focused node.
func (g *NodeGraph) SetFocus(nodeID string) {
	if g.data != nil {
		// Update focused flags
		for _, node := range g.data.nodes {
			node.Focused = node.ID == nodeID
		}
		g.data.FocusID = nodeID
		g.centerOnNode(nodeID)
	}
}

// centerOnAll centers the viewport on the bounding box of all nodes.
func (g *NodeGraph) centerOnAll() {
	if g.data == nil || len(g.data.nodes) == 0 {
		return
	}
	_, _, width, height := g.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}
	minX, minY := 1<<30, 1<<30
	maxX, maxY := -1<<30, -1<<30
	for _, node := range g.data.nodes {
		if node.x < minX {
			minX = node.x
		}
		if node.y < minY {
			minY = node.y
		}
		if node.x+node.width > maxX {
			maxX = node.x + node.width
		}
		if node.y+node.height > maxY {
			maxY = node.y + node.height
		}
	}
	g.offsetX = minX + (maxX-minX)/2 - width/2
	g.offsetY = minY + (maxY-minY)/2 - height/2
}

// centerOnNode centers the view on the given node.
func (g *NodeGraph) centerOnNode(nodeID string) {
	if g.data == nil {
		return
	}
	node := g.data.GetNode(nodeID)
	if node == nil {
		return
	}

	_, _, width, height := g.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	// Center the node in the viewport
	g.offsetX = node.x + node.width/2 - width/2
	g.offsetY = node.y + node.height/2 - height/2
}

// computeLayout calculates node positions using a hierarchical layout.
func (g *NodeGraph) computeLayout() {
	if g.data == nil || len(g.data.nodes) == 0 {
		return
	}

	// Build parent-child relationships from edges
	children := make(map[string][]string)
	parents := make(map[string]string)
	for _, edge := range g.data.edges {
		children[edge.From] = append(children[edge.From], edge.To)
		parents[edge.To] = edge.From
	}

	// Find root nodes (nodes with no parents)
	var roots []string
	for id := range g.data.nodes {
		if _, hasParent := parents[id]; !hasParent {
			roots = append(roots, id)
		}
	}

	// If focus node exists, start from it or its root
	if g.data.FocusID != "" {
		// Find the root of the focused node's tree
		focusRoot := g.data.FocusID
		for {
			if parent, ok := parents[focusRoot]; ok {
				focusRoot = parent
			} else {
				break
			}
		}
		// Put focus root first
		for i, r := range roots {
			if r == focusRoot {
				roots[0], roots[i] = roots[i], roots[0]
				break
			}
		}
	}

	// Assign levels to each node
	levels := make(map[string]int)
	var assignLevels func(nodeID string, level int)
	assignLevels = func(nodeID string, level int) {
		levels[nodeID] = level
		for _, childID := range children[nodeID] {
			assignLevels(childID, level+1)
		}
	}

	for _, root := range roots {
		assignLevels(root, 0)
	}

	// Group nodes by level
	levelNodes := make(map[int][]string)
	maxLevel := 0
	for id, level := range levels {
		levelNodes[level] = append(levelNodes[level], id)
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Position nodes
	vertSpacing := g.nodeHeight + 2
	horizSpacing := g.nodeWidth + 4

	for level := 0; level <= maxLevel; level++ {
		nodes := levelNodes[level]
		startX := -(len(nodes)*horizSpacing)/2 + horizSpacing/2

		for i, nodeID := range nodes {
			node := g.data.GetNode(nodeID)
			if node != nil {
				node.x = startX + i*horizSpacing
				node.y = level * vertSpacing
				node.width = g.nodeWidth
				node.height = g.nodeHeight
			}
		}
	}
}

// Draw renders the node graph.
func (g *NodeGraph) Draw(screen tcell.Screen) {
	g.Box.DrawForSubclass(screen)
	x, y, width, height := g.GetInnerRect()

	if width <= 0 || height <= 0 || g.data == nil {
		return
	}

	// Auto-center on first draw after data is set
	if g.needsCenter {
		g.needsCenter = false
		if g.fitAll {
			g.centerOnAll()
		} else if g.data.FocusID != "" {
			g.centerOnNode(g.data.FocusID)
		} else if len(g.data.nodes) > 0 {
			for id := range g.data.nodes {
				g.centerOnNode(id)
				break
			}
		}
	}

	// Get colors at draw time
	th := g.th()
	bgColor := th.Bg()
	fgColor := th.Fg()
	fgDimColor := th.FgDim()

	// Clear the area
	fillRect(screen, x, y, width, height, tcell.StyleDefault.Background(bgColor))

	// Draw edges first (behind nodes)
	for _, edge := range g.data.edges {
		g.drawEdge(screen, x, y, width, height, edge)
	}

	// Draw nodes
	for _, node := range g.data.nodes {
		g.drawNode(screen, x, y, width, height, node, bgColor, fgColor, fgDimColor)
	}
}

func (g *NodeGraph) drawNode(screen tcell.Screen, viewX, viewY, viewWidth, viewHeight int, node *GraphNode, bgColor, fgColor, fgDimColor tcell.Color) {
	// Calculate screen position (offset already includes centering adjustment)
	screenX := viewX + node.x - g.offsetX
	screenY := viewY + node.y - g.offsetY

	// Check if node is visible
	if screenX+node.width < viewX || screenX >= viewX+viewWidth {
		return
	}
	if screenY+node.height < viewY || screenY >= viewY+viewHeight {
		return
	}

	// Determine node colors based on status
	statusColor := g.getStatusColor(node.Status)
	borderColor := statusColor
	labelColor := fgColor
	if node.Focused {
		labelColor = statusColor
	}

	// Draw box border
	//   ╭──────────────╮
	//   │ Label        │
	//   │   ● status   │
	//   ╰──────────────╯

	borderStyle := tcell.StyleDefault.Background(bgColor).Foreground(borderColor)

	// Top border
	if screenY >= viewY && screenY < viewY+viewHeight {
		col := screenX
		if col >= viewX && col < viewX+viewWidth {
			screen.SetContent(col, screenY, '╭', nil, borderStyle)
		}
		col++
		for i := 0; i < node.width-2 && col < viewX+viewWidth; i++ {
			if col >= viewX {
				screen.SetContent(col, screenY, '─', nil, borderStyle)
			}
			col++
		}
		if col >= viewX && col < viewX+viewWidth {
			screen.SetContent(col, screenY, '╮', nil, borderStyle)
		}
	}

	// Middle rows
	for row := 1; row < node.height-1; row++ {
		rowY := screenY + row
		if rowY < viewY || rowY >= viewY+viewHeight {
			continue
		}

		// Left border
		if screenX >= viewX && screenX < viewX+viewWidth {
			screen.SetContent(screenX, rowY, '│', nil, borderStyle)
		}

		// Content
		labelStyle := tcell.StyleDefault.Background(bgColor).Foreground(labelColor)
		if row == 1 {
			// Draw label centered
			label := g.truncateLabel(node.Label, node.width-2)
			padding := (node.width - 2 - len([]rune(label))) / 2
			col := screenX + 1
			for i := 0; i < padding && col < viewX+viewWidth; i++ {
				if col >= viewX {
					screen.SetContent(col, rowY, ' ', nil, labelStyle)
				}
				col++
			}
			for _, r := range label {
				if col >= viewX && col < viewX+viewWidth {
					screen.SetContent(col, rowY, r, nil, labelStyle)
				}
				col++
			}
			for col < screenX+node.width-1 && col < viewX+viewWidth {
				if col >= viewX {
					screen.SetContent(col, rowY, ' ', nil, labelStyle)
				}
				col++
			}
		} else {
			// Draw status centered
			statusStyle := tcell.StyleDefault.Background(bgColor).Foreground(statusColor)
			statusIcon := g.getStatusIcon(node.Status)
			statusText := statusIcon + " " + node.Status

			if len(statusText) > node.width-2 {
				statusText = statusText[:node.width-2]
			}
			padding := (node.width - 2 - len([]rune(statusText))) / 2
			col := screenX + 1
			for i := 0; i < padding && col < viewX+viewWidth; i++ {
				if col >= viewX {
					screen.SetContent(col, rowY, ' ', nil, statusStyle)
				}
				col++
			}
			for _, r := range statusText {
				if col >= viewX && col < viewX+viewWidth {
					screen.SetContent(col, rowY, r, nil, statusStyle)
				}
				col++
			}
			for col < screenX+node.width-1 && col < viewX+viewWidth {
				if col >= viewX {
					screen.SetContent(col, rowY, ' ', nil, statusStyle)
				}
				col++
			}
		}

		// Right border
		rightX := screenX + node.width - 1
		if rightX >= viewX && rightX < viewX+viewWidth {
			screen.SetContent(rightX, rowY, '│', nil, borderStyle)
		}
	}

	// Bottom border
	bottomY := screenY + node.height - 1
	if bottomY >= viewY && bottomY < viewY+viewHeight {
		col := screenX
		if col >= viewX && col < viewX+viewWidth {
			screen.SetContent(col, bottomY, '╰', nil, borderStyle)
		}
		col++
		for i := 0; i < node.width-2 && col < viewX+viewWidth; i++ {
			if col >= viewX {
				screen.SetContent(col, bottomY, '─', nil, borderStyle)
			}
			col++
		}
		if col >= viewX && col < viewX+viewWidth {
			screen.SetContent(col, bottomY, '╯', nil, borderStyle)
		}
	}
}

func (g *NodeGraph) drawEdge(screen tcell.Screen, viewX, viewY, viewWidth, viewHeight int, edge *GraphEdge) {
	fromNode := g.data.GetNode(edge.From)
	toNode := g.data.GetNode(edge.To)
	if fromNode == nil || toNode == nil {
		return
	}

	// Calculate connection points (offset already includes centering adjustment)
	fromX := viewX + fromNode.x + fromNode.width/2 - g.offsetX
	fromY := viewY + fromNode.y + fromNode.height - g.offsetY
	toX := viewX + toNode.x + toNode.width/2 - g.offsetX
	toY := viewY + toNode.y - g.offsetY

	// Determine edge style
	th := g.th()
	edgeColor := th.FgDim()
	var vertChar, horizChar, downChar rune
	switch edge.Type {
	case GraphEdgeSolid:
		vertChar = '│'
		horizChar = '─'
		downChar = '▼'
	case GraphEdgeDashed:
		vertChar = '┆'
		horizChar = '╌'
		downChar = '▼'
		edgeColor = th.Warning()
	case GraphEdgeDotted:
		vertChar = '┊'
		horizChar = '·'
		downChar = '▼'
	}

	edgeStyle := tcell.StyleDefault.Background(th.Bg()).Foreground(edgeColor)

	// Draw simple vertical line for parent-child (assuming hierarchical layout)
	if fromX == toX {
		// Straight vertical line
		for y := fromY; y < toY; y++ {
			if y >= viewY && y < viewY+viewHeight && fromX >= viewX && fromX < viewX+viewWidth {
				if y == toY-1 {
					screen.SetContent(fromX, y, downChar, nil, edgeStyle)
				} else {
					screen.SetContent(fromX, y, vertChar, nil, edgeStyle)
				}
			}
		}
	} else {
		// L-shaped or stepped connection
		midY := (fromY + toY) / 2

		// Vertical down from source
		for y := fromY; y <= midY; y++ {
			if y >= viewY && y < viewY+viewHeight && fromX >= viewX && fromX < viewX+viewWidth {
				screen.SetContent(fromX, y, vertChar, nil, edgeStyle)
			}
		}

		// Horizontal across
		startX, endX := fromX, toX
		if startX > endX {
			startX, endX = endX, startX
		}
		for x := startX; x <= endX; x++ {
			if midY >= viewY && midY < viewY+viewHeight && x >= viewX && x < viewX+viewWidth {
				screen.SetContent(x, midY, horizChar, nil, edgeStyle)
			}
		}

		// Draw corners
		if midY >= viewY && midY < viewY+viewHeight {
			if fromX >= viewX && fromX < viewX+viewWidth {
				if toX > fromX {
					screen.SetContent(fromX, midY, '└', nil, edgeStyle)
				} else {
					screen.SetContent(fromX, midY, '┘', nil, edgeStyle)
				}
			}
			if toX >= viewX && toX < viewX+viewWidth {
				if toX > fromX {
					screen.SetContent(toX, midY, '┐', nil, edgeStyle)
				} else {
					screen.SetContent(toX, midY, '┌', nil, edgeStyle)
				}
			}
		}

		// Vertical down to target (between corner and arrow)
		for y := midY + 1; y < toY-1; y++ {
			if y >= viewY && y < viewY+viewHeight && toX >= viewX && toX < viewX+viewWidth {
				screen.SetContent(toX, y, vertChar, nil, edgeStyle)
			}
		}

		// Always draw arrow pointing into target (if there's space)
		arrowY := toY - 1
		if arrowY > midY && arrowY >= viewY && arrowY < viewY+viewHeight && toX >= viewX && toX < viewX+viewWidth {
			screen.SetContent(toX, arrowY, downChar, nil, edgeStyle)
		}
	}
}

func (g *NodeGraph) truncateLabel(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

func (g *NodeGraph) getStatusColor(status string) tcell.Color {
	th := g.th()
	switch strings.ToLower(status) {
	case "running":
		return th.Info()
	case "completed":
		return th.Success()
	case "failed":
		return th.Error()
	case "canceled", "cancelled":
		return th.Warning()
	case "terminated":
		return th.Error()
	case "timedout", "timed_out":
		return th.Warning()
	default:
		return th.FgDim()
	}
}

func (g *NodeGraph) getStatusIcon(status string) string {
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

// HandleKey handles keyboard input.
func (g *NodeGraph) HandleKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyUp:
		g.offsetY -= 2
	case tcell.KeyDown:
		g.offsetY += 2
	case tcell.KeyLeft:
		g.offsetX -= 4
	case tcell.KeyRight:
		g.offsetX += 4
	case tcell.KeyEnter:
		if g.data != nil && g.data.FocusID != "" {
			if node := g.data.GetNode(g.data.FocusID); node != nil && g.onSelect != nil {
				g.onSelect(node)
			}
		}
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'h':
			g.offsetX -= 4
		case 'l':
			g.offsetX += 4
		case 'k':
			g.offsetY -= 2
		case 'j':
			g.offsetY += 2
		case 'c':
			// Center on focused node
			if g.data != nil && g.data.FocusID != "" {
				g.centerOnNode(g.data.FocusID)
			}
		}
	}
	return false
}

// Focus handles focus.
// HasFocus returns whether the graph has focus.
func (g *NodeGraph) HasFocus() bool {
	return g.Box.HasFocus()
}
