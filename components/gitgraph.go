package components

import (
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/atterpac/jig/theme"
)

// Box drawing characters for git graph
const (
	gitNode      = '●' // Regular commit
	gitMerge     = '◆' // Merge commit
	gitHead      = '◉' // HEAD commit
	gitStash     = '◇' // Stash entry (hollow diamond)
	gitVert      = '│' // Vertical line
	gitHoriz     = '─' // Horizontal line
	gitTopLeft   = '╭' // Corner down-right
	gitTopRight  = '╮' // Corner down-left
	gitVertRight = '├' // T-junction right
	gitVertLeft  = '┤' // T-junction left
	gitCross     = '┼' // Cross intersection
)

// GitCommit represents a git commit with graph positioning info
type GitCommit struct {
	Hash      string
	ShortHash string
	Message   string
	Author    string
	Date      time.Time
	Parents   []string // Parent commit hashes
	Children  []string // Child commit hashes
	Branch    string   // Branch name if this is a branch tip
	Tags      []string // Tags pointing to this commit
	IsMerge   bool     // True if len(Parents) > 1
	IsStash   bool     // True if this is a stash entry
	Column    int      // Assigned column in graph layout
	Row       int      // Row position in flat list
	Refs      []string // All refs (branches, tags) at this commit
	Ahead     int      // Commits ahead of upstream (for branch tips)
	Behind    int      // Commits behind upstream (for branch tips)
	Data      any      // Custom user data
}

// GitGraphData represents the commit graph with layout info
type GitGraphData struct {
	Commits    []*GitCommit          // Commits in topological order (newest first)
	CommitMap  map[string]*GitCommit // Hash -> Commit lookup
	Branches   []string              // All branch names
	MaxColumn  int                   // Maximum column used in layout
	ActiveCols map[int]string        // Column -> commit hash currently in that column
}

// NewGitGraphData creates an empty graph data structure
func NewGitGraphData() *GitGraphData {
	return &GitGraphData{
		CommitMap:  make(map[string]*GitCommit),
		ActiveCols: make(map[int]string),
	}
}

// AddCommit adds a commit to the graph
func (g *GitGraphData) AddCommit(c *GitCommit) {
	g.Commits = append(g.Commits, c)
	g.CommitMap[c.Hash] = c
}

// GetCommit retrieves a commit by hash
func (g *GitGraphData) GetCommit(hash string) *GitCommit {
	return g.CommitMap[hash]
}

// LayoutGraph assigns columns to commits for visualization
func (g *GitGraphData) LayoutGraph() {
	if len(g.Commits) == 0 {
		return
	}

	activeLines := make(map[int]string) // column -> commit hash that line leads to
	nextFreeCol := 0

	for i, commit := range g.Commits {
		commit.Row = i

		// Check if this commit was expected in an active line
		foundCol := -1
		for col, hash := range activeLines {
			if hash == commit.Hash {
				foundCol = col
				break
			}
		}

		if foundCol >= 0 {
			// This commit continues an existing line
			commit.Column = foundCol
		} else {
			// New branch/line - find first available column
			commit.Column = nextFreeCol
			for col := 0; col < nextFreeCol; col++ {
				if _, exists := activeLines[col]; !exists {
					commit.Column = col
					break
				}
			}
			if commit.Column == nextFreeCol {
				nextFreeCol++
			}
		}

		if commit.Column > g.MaxColumn {
			g.MaxColumn = commit.Column
		}

		// Remove this commit from active lines
		delete(activeLines, commit.Column)

		// Track parent commits - only if they exist in our commit list
		for idx, parentHash := range commit.Parents {
			// Only track parents that are in our loaded commits
			if g.CommitMap[parentHash] == nil {
				continue
			}

			// Check if parent is already tracked
			parentHasCol := false
			for _, h := range activeLines {
				if h == parentHash {
					parentHasCol = true
					break
				}
			}

			if !parentHasCol {
				if idx == 0 {
					// First parent continues in same column
					activeLines[commit.Column] = parentHash
				} else {
					// Secondary parents get a new column (for merge visualization)
					newCol := nextFreeCol
					for col := 0; col < nextFreeCol; col++ {
						if _, exists := activeLines[col]; !exists {
							newCol = col
							break
						}
					}
					if newCol == nextFreeCol {
						nextFreeCol++
					}
					activeLines[newCol] = parentHash
					if newCol > g.MaxColumn {
						g.MaxColumn = newCol
					}
				}
			}
		}
	}
}

// GitGraph is a git commit graph visualization component
type GitGraph struct {
	*tview.Box

	graph         *GitGraphData
	selectedIndex int
	offset        int

	// Callbacks
	onSelect func(commit *GitCommit)
	onChange func(commit *GitCommit)

	// Style options
	showRefs   bool
	showHash   bool
	showAuthor bool
	showDate   bool
	dateFormat string
	laneColors []tcell.Color
}

// NewGitGraph creates a new git graph component
func NewGitGraph() *GitGraph {
	return &GitGraph{
		Box:        tview.NewBox(),
		showRefs:   true,
		showHash:   true,
		showAuthor: false,
		showDate:   false,
		dateFormat: "2006-01-02",
	}
}

// SetGraph sets the commit graph data to render
func (g *GitGraph) SetGraph(data *GitGraphData) *GitGraph {
	g.graph = data
	g.selectedIndex = 0
	g.offset = 0
	return g
}

// SetOnSelect sets the callback for when Enter is pressed on a commit
func (g *GitGraph) SetOnSelect(fn func(commit *GitCommit)) *GitGraph {
	g.onSelect = fn
	return g
}

// SetOnChange sets the callback for when the selection changes
func (g *GitGraph) SetOnChange(fn func(commit *GitCommit)) *GitGraph {
	g.onChange = fn
	return g
}

// SetShowRefs enables/disables showing branch/tag refs
func (g *GitGraph) SetShowRefs(show bool) *GitGraph {
	g.showRefs = show
	return g
}

// SetShowHash enables/disables showing commit hash
func (g *GitGraph) SetShowHash(show bool) *GitGraph {
	g.showHash = show
	return g
}

// SetShowAuthor enables/disables showing author
func (g *GitGraph) SetShowAuthor(show bool) *GitGraph {
	g.showAuthor = show
	return g
}

// SetShowDate enables/disables showing date
func (g *GitGraph) SetShowDate(show bool) *GitGraph {
	g.showDate = show
	return g
}

// SetDateFormat sets the date format string
func (g *GitGraph) SetDateFormat(format string) *GitGraph {
	g.dateFormat = format
	return g
}

// SetLaneColors sets custom colors for graph lanes
func (g *GitGraph) SetLaneColors(colors []tcell.Color) *GitGraph {
	g.laneColors = colors
	return g
}

// GetSelected returns the currently selected commit
func (g *GitGraph) GetSelected() *GitCommit {
	if g.graph == nil || g.selectedIndex >= len(g.graph.Commits) {
		return nil
	}
	return g.graph.Commits[g.selectedIndex]
}

// SetSelectedIndex sets the selected commit by index
func (g *GitGraph) SetSelectedIndex(index int) *GitGraph {
	if g.graph != nil && index >= 0 && index < len(g.graph.Commits) {
		g.selectedIndex = index
		g.triggerOnChange()
	}
	return g
}

// SelectByHash selects a commit by its hash
func (g *GitGraph) SelectByHash(hash string) *GitGraph {
	if g.graph == nil {
		return g
	}
	for i, commit := range g.graph.Commits {
		if commit.Hash == hash || commit.ShortHash == hash {
			g.selectedIndex = i
			g.triggerOnChange()
			break
		}
	}
	return g
}

func (g *GitGraph) triggerOnChange() {
	if g.onChange != nil {
		g.onChange(g.GetSelected())
	}
}

// laneState tracks rendering state for each lane
type gitLaneState struct {
	hasLine       bool
	hasNode       bool
	mergeFrom     int
	mergeTo       int
	branchFrom    int
	isStartOfLine bool
	commitHash    string
}

// buildRowStates computes lane states for rendering
func (g *GitGraph) buildRowStates() []map[int]*gitLaneState {
	if g.graph == nil || len(g.graph.Commits) == 0 {
		return nil
	}

	numRows := len(g.graph.Commits)
	states := make([]map[int]*gitLaneState, numRows)
	activeLanes := make(map[int]string)

	for row := 0; row < numRows; row++ {
		commit := g.graph.Commits[row]
		states[row] = make(map[int]*gitLaneState)

		maxLane := g.graph.MaxColumn
		for lane := 0; lane <= maxLane; lane++ {
			states[row][lane] = &gitLaneState{
				mergeFrom:  -1,
				mergeTo:    -1,
				branchFrom: -1,
			}
		}

		for lane, hash := range activeLanes {
			states[row][lane].hasLine = true
			states[row][lane].commitHash = hash
		}

		states[row][commit.Column].hasNode = true
		states[row][commit.Column].commitHash = commit.Hash

		if commit.IsMerge && len(commit.Parents) > 1 {
			for i := 1; i < len(commit.Parents); i++ {
				parentHash := commit.Parents[i]
				parent := g.graph.GetCommit(parentHash)
				if parent != nil {
					fromLane := parent.Column
					toLane := commit.Column
					states[row][fromLane].mergeTo = toLane
					states[row][toLane].mergeFrom = fromLane
				}
			}
		}

		delete(activeLanes, commit.Column)

		for i, parentHash := range commit.Parents {
			parent := g.graph.GetCommit(parentHash)
			if parent == nil {
				continue
			}

			if i == 0 {
				activeLanes[commit.Column] = parentHash
			} else {
				if _, exists := activeLanes[parent.Column]; !exists {
					activeLanes[parent.Column] = parentHash
					states[row][parent.Column].isStartOfLine = true
					states[row][parent.Column].branchFrom = commit.Column
				}
			}
		}
	}

	return states
}

// Draw renders the git graph
func (g *GitGraph) Draw(screen tcell.Screen) {
	g.Box.DrawForSubclass(screen, g)
	x, y, width, height := g.GetInnerRect()

	if width <= 0 || height <= 0 || g.graph == nil || len(g.graph.Commits) == 0 {
		return
	}

	bgColor := theme.Bg()
	fgColor := theme.Fg()
	fgDimColor := theme.FgDim()
	accentColor := theme.Accent()

	laneColors := g.laneColors
	if len(laneColors) == 0 {
		laneColors = []tcell.Color{
			theme.Accent(),
			theme.Success(),
			theme.Warning(),
			theme.Info(),
			theme.Error(),
			tcell.ColorMediumPurple,
			tcell.ColorDarkCyan,
		}
	}

	if g.selectedIndex < g.offset {
		g.offset = g.selectedIndex
	}
	if g.selectedIndex >= g.offset+height {
		g.offset = g.selectedIndex - height + 1
	}

	rowStates := g.buildRowStates()

	for i := 0; i < height && g.offset+i < len(g.graph.Commits); i++ {
		commit := g.graph.Commits[g.offset+i]
		rowY := y + i
		rowIndex := g.offset + i

		isSelected := rowIndex == g.selectedIndex
		rowStyle := tcell.StyleDefault.Background(bgColor).Foreground(fgColor)
		if isSelected {
			rowStyle = rowStyle.Background(accentColor).Foreground(bgColor)
		}

		for col := x; col < x+width; col++ {
			screen.SetContent(col, rowY, ' ', nil, rowStyle)
		}

		col := x
		states := rowStates[rowIndex]

		for lane := 0; lane <= g.graph.MaxColumn; lane++ {
			state := states[lane]
			laneColor := laneColors[lane%len(laneColors)]

			cellStyle := rowStyle
			if !isSelected {
				cellStyle = tcell.StyleDefault.Background(bgColor).Foreground(laneColor)
			}

			chars := g.getLaneChars(commit, lane, state, states, rowIndex, rowStates)

			for ci, ch := range chars {
				if col < x+width {
					charStyle := cellStyle
					if ci == 1 && state.hasNode && !isSelected {
						charStyle = tcell.StyleDefault.Background(bgColor).Foreground(laneColors[commit.Column%len(laneColors)])
					}
					screen.SetContent(col, rowY, ch, nil, charStyle)
					col++
				}
			}
		}

		if col < x+width {
			screen.SetContent(col, rowY, ' ', nil, rowStyle)
			col++
		}

		// Build ref text FIRST to know exact space needed
		var refText string
		if g.showRefs && len(commit.Refs) > 0 {
			// Build a set of remote refs for quick lookup
			remoteRefs := make(map[string]bool)
			for _, ref := range commit.Refs {
				if strings.HasPrefix(ref, "origin/") {
					remoteRefs[strings.TrimPrefix(ref, "origin/")] = true
				}
			}

			// Build ahead/behind indicator
			aheadBehind := ""
			if commit.Ahead > 0 || commit.Behind > 0 {
				if commit.Ahead > 0 && commit.Behind > 0 {
					aheadBehind = " ↑" + strconv.Itoa(commit.Ahead) + "↓" + strconv.Itoa(commit.Behind)
				} else if commit.Ahead > 0 {
					aheadBehind = " ↑" + strconv.Itoa(commit.Ahead)
				} else if commit.Behind > 0 {
					aheadBehind = " ↓" + strconv.Itoa(commit.Behind)
				}
			}

			// Build ref display parts
			var refParts []string
			for _, ref := range commit.Refs {
				if strings.HasPrefix(ref, "origin/") {
					continue
				}
				if ref == "HEAD" {
					continue
				}
				if remoteRefs[ref] {
					refParts = append(refParts, "↕ "+ref+aheadBehind)
				} else {
					refParts = append(refParts, "● "+ref)
				}
			}

			// Remote-only refs
			for _, ref := range commit.Refs {
				if strings.HasPrefix(ref, "origin/") {
					localName := strings.TrimPrefix(ref, "origin/")
					hasLocal := false
					for _, r := range commit.Refs {
						if r == localName {
							hasLocal = true
							break
						}
					}
					if !hasLocal {
						refParts = append(refParts, "○ "+localName)
					}
				}
			}

			if len(refParts) > 0 {
				refText = " " + strings.Join(refParts, "  ") + " "
			}
		}

		// Calculate space needed for right-side elements
		rightReserved := len([]rune(refText))
		if g.showHash {
			rightReserved += len(commit.ShortHash) + 1
		}

		// Calculate available space for message
		msgSpace := (x + width) - col - rightReserved - 2
		msgRunes := []rune(commit.Message)

		msgStyle := rowStyle
		if msgSpace > 0 {
			if len(msgRunes) > msgSpace {
				// Truncate with ellipsis
				for i := 0; i < msgSpace-1 && col < x+width; i++ {
					screen.SetContent(col, rowY, msgRunes[i], nil, msgStyle)
					col++
				}
				if col < x+width {
					screen.SetContent(col, rowY, '…', nil, msgStyle)
					col++
				}
			} else {
				// Full message fits
				for _, ch := range commit.Message {
					if col < x+width {
						screen.SetContent(col, rowY, ch, nil, msgStyle)
						col++
					}
				}
			}
		}

		if g.showAuthor {
			authorStyle := rowStyle
			if !isSelected {
				authorStyle = tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
			}
			authorText := " - " + commit.Author
			for _, ch := range authorText {
				if col < x+width {
					screen.SetContent(col, rowY, ch, nil, authorStyle)
					col++
				}
			}
		}

		if g.showDate {
			dateStyle := rowStyle
			if !isSelected {
				dateStyle = tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
			}
			dateText := " " + commit.Date.Format(g.dateFormat)
			for _, ch := range dateText {
				if col < x+width {
					screen.SetContent(col, rowY, ch, nil, dateStyle)
					col++
				}
			}
		}

		if g.showHash {
			hashStyle := rowStyle
			if !isSelected {
				hashStyle = tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
			}
			hashText := " " + commit.ShortHash
			for _, ch := range hashText {
				if col < x+width {
					screen.SetContent(col, rowY, ch, nil, hashStyle)
					col++
				}
			}
		}

		// Right-aligned refs with lane color matching (refText already built above)
		if refText != "" {
			refLen := len([]rune(refText))
			refStartCol := x + width - refLen

			// Always render refs - they take priority
			laneColor := laneColors[commit.Column%len(laneColors)]
			refStyle := rowStyle
			if !isSelected {
				refStyle = tcell.StyleDefault.Background(bgColor).Foreground(laneColor).Bold(true)
			}

			refCol := refStartCol
			for _, ch := range refText {
				if refCol < x+width {
					screen.SetContent(refCol, rowY, ch, nil, refStyle)
					refCol++
				}
			}
		}
	}
}

// getLaneChars returns the 3 characters to draw for a lane
func (g *GitGraph) getLaneChars(commit *GitCommit, lane int, state *gitLaneState, allStates map[int]*gitLaneState, rowIndex int, allRowStates []map[int]*gitLaneState) []rune {
	if state != nil && (state.mergeFrom != -1 || state.mergeTo != -1) {
		if state.hasNode {
			node := g.getNodeChar(commit)
			if state.mergeFrom != -1 {
				if state.mergeFrom < lane {
					return []rune{gitHoriz, node, ' '}
				} else if state.mergeFrom > lane {
					return []rune{' ', node, gitHoriz}
				}
			}
			if state.mergeTo != -1 {
				if state.mergeTo < lane {
					return []rune{gitHoriz, node, ' '}
				} else if state.mergeTo > lane {
					return []rune{' ', node, gitHoriz}
				}
			}
			return []rune{' ', node, ' '}
		}

		if state.mergeTo != -1 {
			if state.mergeTo < lane {
				if state.hasLine {
					return []rune{gitHoriz, gitVertLeft, ' '}
				}
				return []rune{gitHoriz, gitTopRight, ' '}
			} else if state.mergeTo > lane {
				if state.hasLine {
					return []rune{' ', gitVertRight, gitHoriz}
				}
				return []rune{' ', gitTopLeft, gitHoriz}
			}
		}

		if state.hasLine {
			return []rune{' ', gitVert, ' '}
		}
		return []rune{' ', ' ', ' '}
	}

	if commit.IsMerge && lane == commit.Column {
		node := g.getNodeChar(commit)
		if len(commit.Parents) > 1 {
			secondParent := g.graph.GetCommit(commit.Parents[1])
			if secondParent != nil {
				if secondParent.Column < lane {
					return []rune{gitHoriz, node, ' '}
				} else if secondParent.Column > lane {
					return []rune{' ', node, gitHoriz}
				}
			}
		}
		return []rune{' ', node, ' '}
	}

	if state.hasNode {
		node := g.getNodeChar(commit)
		return []rune{' ', node, ' '}
	}

	if commit.IsMerge && len(commit.Parents) > 1 {
		secondParent := g.graph.GetCommit(commit.Parents[1])
		if secondParent != nil {
			minLane := commit.Column
			maxLane := secondParent.Column
			if minLane > maxLane {
				minLane, maxLane = maxLane, minLane
			}

			if lane > minLane && lane < maxLane {
				if state.hasLine {
					return []rune{gitHoriz, gitCross, gitHoriz}
				}
				return []rune{gitHoriz, gitHoriz, gitHoriz}
			}

			if lane == secondParent.Column {
				if lane < commit.Column {
					if state.hasLine {
						return []rune{' ', gitVertRight, gitHoriz}
					}
					return []rune{' ', gitTopLeft, gitHoriz}
				} else {
					if state.hasLine {
						return []rune{gitHoriz, gitVertLeft, ' '}
					}
					return []rune{gitHoriz, gitTopRight, ' '}
				}
			}
		}
	}

	if state != nil && state.hasLine {
		return []rune{' ', gitVert, ' '}
	}

	if state != nil && state.isStartOfLine {
		if state.branchFrom != -1 {
			if state.branchFrom < lane {
				return []rune{gitHoriz, gitTopRight, ' '}
			} else {
				return []rune{' ', gitTopLeft, gitHoriz}
			}
		}
	}

	return []rune{' ', ' ', ' '}
}

func (g *GitGraph) getNodeChar(commit *GitCommit) rune {
	if commit.IsStash {
		return gitStash
	}
	for _, ref := range commit.Refs {
		if ref == "HEAD" {
			return gitHead
		}
	}
	if commit.IsMerge {
		return gitMerge
	}
	return gitNode
}

// InputHandler handles keyboard input
func (g *GitGraph) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return g.WrapInputHandler(func(event *tcell.EventKey, setFocus func(tview.Primitive)) {
		if g.graph == nil || len(g.graph.Commits) == 0 {
			return
		}

		prevIndex := g.selectedIndex

		switch event.Key() {
		case tcell.KeyDown:
			g.moveDown()
		case tcell.KeyUp:
			g.moveUp()
		case tcell.KeyHome:
			g.selectedIndex = 0
		case tcell.KeyEnd:
			g.selectedIndex = len(g.graph.Commits) - 1
		case tcell.KeyPgDn:
			_, _, _, height := g.GetInnerRect()
			g.selectedIndex += height
			if g.selectedIndex >= len(g.graph.Commits) {
				g.selectedIndex = len(g.graph.Commits) - 1
			}
		case tcell.KeyPgUp:
			_, _, _, height := g.GetInnerRect()
			g.selectedIndex -= height
			if g.selectedIndex < 0 {
				g.selectedIndex = 0
			}
		case tcell.KeyEnter:
			if commit := g.GetSelected(); commit != nil && g.onSelect != nil {
				g.onSelect(commit)
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				g.moveDown()
			case 'k':
				g.moveUp()
			case 'g':
				g.selectedIndex = 0
			case 'G':
				g.selectedIndex = len(g.graph.Commits) - 1
			case 'p':
				// Jump to first parent
				if commit := g.GetSelected(); commit != nil && len(commit.Parents) > 0 {
					g.SelectByHash(commit.Parents[0])
				}
			case 'P':
				// Jump to second parent (merge commits)
				if commit := g.GetSelected(); commit != nil && len(commit.Parents) > 1 {
					g.SelectByHash(commit.Parents[1])
				}
			case 'c':
				// Jump to first child
				if commit := g.GetSelected(); commit != nil && len(commit.Children) > 0 {
					g.SelectByHash(commit.Children[0])
				}
			}
		case tcell.KeyCtrlD:
			_, _, _, height := g.GetInnerRect()
			g.selectedIndex += height / 2
			if g.selectedIndex >= len(g.graph.Commits) {
				g.selectedIndex = len(g.graph.Commits) - 1
			}
		case tcell.KeyCtrlU:
			_, _, _, height := g.GetInnerRect()
			g.selectedIndex -= height / 2
			if g.selectedIndex < 0 {
				g.selectedIndex = 0
			}
		}

		if g.selectedIndex != prevIndex {
			g.triggerOnChange()
		}
	})
}

func (g *GitGraph) moveDown() {
	if g.selectedIndex < len(g.graph.Commits)-1 {
		g.selectedIndex++
	}
}

func (g *GitGraph) moveUp() {
	if g.selectedIndex > 0 {
		g.selectedIndex--
	}
}

// MouseHandler handles mouse input
func (g *GitGraph) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return g.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(tview.Primitive)) (bool, tview.Primitive) {
		_, y, _, _ := g.GetInnerRect()
		mx, my := event.Position()

		if !g.InRect(mx, my) {
			return false, nil
		}

		switch action {
		case tview.MouseLeftClick:
			setFocus(g)
			clickedIndex := g.offset + (my - y)
			if clickedIndex >= 0 && clickedIndex < len(g.graph.Commits) {
				prevIndex := g.selectedIndex
				g.selectedIndex = clickedIndex
				if g.selectedIndex != prevIndex {
					g.triggerOnChange()
				}
				return true, g
			}
		case tview.MouseLeftDoubleClick:
			clickedIndex := g.offset + (my - y)
			if clickedIndex >= 0 && clickedIndex < len(g.graph.Commits) {
				g.selectedIndex = clickedIndex
				if g.onSelect != nil {
					g.onSelect(g.graph.Commits[clickedIndex])
				}
				return true, g
			}
		case tview.MouseScrollUp:
			if g.offset > 0 {
				g.offset--
			}
			return true, g
		case tview.MouseScrollDown:
			_, _, _, height := g.GetInnerRect()
			if g.offset < len(g.graph.Commits)-height {
				g.offset++
			}
			return true, g
		}

		return false, nil
	})
}

// Focus handles focus
func (g *GitGraph) Focus(delegate func(tview.Primitive)) {
	g.Box.Focus(delegate)
}

// HasFocus returns whether the component has focus
func (g *GitGraph) HasFocus() bool {
	return g.Box.HasFocus()
}
