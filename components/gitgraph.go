package components

import (
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Box drawing characters for git graph
const (
	gitNode      = '●' // Regular commit
	gitMerge     = '◆' // Merge commit
	gitHead      = '◉' // HEAD commit
	gitStash     = '◇' // Stash entry (hollow diamond)
	gitUnstaged  = '○' // Unstaged changes (empty circle)
	gitVert      = '│' // Vertical line
	gitHoriz     = '─' // Horizontal line
	gitTopLeft   = '╭' // Corner down-right
	gitTopRight  = '╮' // Corner down-left
	gitBotLeft   = '╰' // Corner up-right
	gitBotRight  = '╯' // Corner up-left
	gitVertRight = '├' // T-junction right
	gitVertLeft  = '┤' // T-junction left
	gitCross     = '┼' // Cross intersection
)

// GitCommit represents a git commit with graph positioning info
type GitCommit struct {
	Hash         string
	ShortHash    string
	Message      string
	Author       string
	Date         time.Time
	Parents      []string // Parent commit hashes
	Children     []string // Child commit hashes
	Branch       string   // Branch name if this is a branch tip
	Tags         []string // Tags pointing to this commit
	IsMerge      bool     // True if len(Parents) > 1
	IsStash      bool     // True if this is a stash entry
	IsPseudoNode bool     // True for virtual nodes (unstaged changes, staged changes)
	PseudoType   string   // "unstaged", "staged", or future types
	Column       int      // Assigned column in graph layout
	Row          int      // Row position in flat list
	Refs         []string // All refs (branches, tags) at this commit
	Ahead        int      // Commits ahead of upstream (for branch tips)
	Behind       int      // Commits behind upstream (for branch tips)
	Data         any      // Custom user data
}

// GitGraphData represents the commit graph with layout info
type GitGraphData struct {
	Commits       []*GitCommit          // Commits in topological order (newest first)
	CommitMap     map[string]*GitCommit // Hash -> Commit lookup
	Branches      []string              // All branch names
	CurrentBranch string                // Current/active branch name (gets column 0)
	MaxColumn     int                   // Maximum column used in layout
	ColumnCap     int                   // Maximum columns to display (0 = unlimited)
	ActiveCols    map[int]string        // Column -> commit hash currently in that column
	ColumnBranch  map[int]string        // Column -> branch name that owns this lane
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
// Each branch lineage gets its own column that persists until the branch merges
// The current branch (if set) is always assigned to column 0
func (g *GitGraphData) LayoutGraph() {
	if len(g.Commits) == 0 {
		return
	}

	// Default column cap if not set
	columnCap := g.ColumnCap
	if columnCap <= 0 {
		columnCap = 12 // Default to 12 columns max (36 chars for graph)
	}

	// activeLines: column -> commit hash that line is expecting next
	activeLines := make(map[int]string)
	// commitColumn: hash -> assigned column (persists for the branch)
	commitColumn := make(map[string]int)
	// columnBranch: column -> branch name that owns this lane
	g.ColumnBranch = make(map[int]string)
	nextFreeCol := 0

	// Find the current branch tip commit and reserve column 0 for it
	var currentBranchTip *GitCommit
	if g.CurrentBranch != "" {
		for _, commit := range g.Commits {
			for _, ref := range commit.Refs {
				if ref == g.CurrentBranch || ref == "HEAD" {
					currentBranchTip = commit
					break
				}
			}
			if currentBranchTip != nil {
				break
			}
		}
	}

	// If we found the current branch tip, reserve column 0 for it
	if currentBranchTip != nil {
		commitColumn[currentBranchTip.Hash] = 0
		g.ColumnBranch[0] = g.CurrentBranch
		nextFreeCol = 1
	}

	for i, commit := range g.Commits {
		commit.Row = i

		// Check if this commit already has a column assigned (e.g., current branch tip)
		if col, exists := commitColumn[commit.Hash]; exists {
			commit.Column = col
			// Check if this commit was expected in an active line and update
			for activeCol, hash := range activeLines {
				if hash == commit.Hash {
					delete(activeLines, activeCol)
					break
				}
			}
		} else {
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
				commitColumn[commit.Hash] = foundCol
			} else {
				// New branch head - allocate a new column
				if nextFreeCol < columnCap {
					commit.Column = nextFreeCol
					nextFreeCol++
				} else {
					// At cap - reuse the last column
					commit.Column = columnCap - 1
				}
				commitColumn[commit.Hash] = commit.Column
			}
		}

		// Record branch name for this column if commit has refs and column doesn't have a branch yet
		if g.ColumnBranch[commit.Column] == "" {
			for _, ref := range commit.Refs {
				if !strings.HasPrefix(ref, "origin/") && ref != "HEAD" {
					g.ColumnBranch[commit.Column] = ref
					break
				}
			}
		}

		if commit.Column > g.MaxColumn && commit.Column < columnCap {
			g.MaxColumn = commit.Column
		}
		if g.MaxColumn >= columnCap {
			g.MaxColumn = columnCap - 1
		}

		// Remove this commit from active lines (we've processed it)
		delete(activeLines, commit.Column)

		// Track parent commits - only if they exist in our commit list
		for idx, parentHash := range commit.Parents {
			// Only track parents that are in our loaded commits
			if g.CommitMap[parentHash] == nil {
				continue
			}

			// Check if parent is already tracked in an active line
			parentHasCol := false
			for _, h := range activeLines {
				if h == parentHash {
					parentHasCol = true
					break
				}
			}

			if !parentHasCol {
				if idx == 0 {
					// First parent continues in same column as this commit
					activeLines[commit.Column] = parentHash
				} else {
					// Secondary parent (merge source) - needs its own column
					// Check if this parent already has a column assigned from another path
					if existingCol, exists := commitColumn[parentHash]; exists {
						// Parent already has a column, mark it active
						activeLines[existingCol] = parentHash
					} else {
						// New line for merge source - allocate column if under cap
						var newCol int
						if nextFreeCol < columnCap {
							newCol = nextFreeCol
							nextFreeCol++
						} else {
							newCol = columnCap - 1
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

	// Parse merge commits to find branch names for merged branches
	// Standard merge messages are like "Merge branch 'feature-x' into main"
	for _, commit := range g.Commits {
		if commit.IsMerge && len(commit.Parents) > 1 {
			branchName := parseMergeBranch(commit.Message)
			if branchName != "" {
				// Find the second parent and associate its column with this branch
				secondParent := g.CommitMap[commit.Parents[1]]
				if secondParent != nil && g.ColumnBranch[secondParent.Column] == "" {
					g.ColumnBranch[secondParent.Column] = branchName
				}
			}
		}
	}

	// Assign branch names to commits based on their column
	// This preserves the original branch even after merges
	for _, commit := range g.Commits {
		// First check if commit has its own branch ref
		if commit.Branch == "" {
			for _, ref := range commit.Refs {
				if !strings.HasPrefix(ref, "origin/") && ref != "HEAD" {
					commit.Branch = ref
					break
				}
			}
		}
		// If still no branch, use the column's branch
		if commit.Branch == "" {
			commit.Branch = g.ColumnBranch[commit.Column]
		}
	}
}

// parseMergeBranch extracts the branch name from a merge commit message
// Handles formats like:
// - "Merge branch 'feature-x'"
// - "Merge branch 'feature-x' into main"
// - "Merge pull request #123 from user/feature-x"
func parseMergeBranch(message string) string {
	// Try "Merge branch 'xyz'" format
	if strings.HasPrefix(message, "Merge branch '") {
		start := len("Merge branch '")
		end := strings.Index(message[start:], "'")
		if end > 0 {
			return message[start : start+end]
		}
	}

	// Try "Merge pull request #N from user/branch" format
	if strings.HasPrefix(message, "Merge pull request") {
		if idx := strings.Index(message, " from "); idx > 0 {
			rest := message[idx+6:]
			// Extract branch part after "user/"
			if slashIdx := strings.Index(rest, "/"); slashIdx > 0 {
				branch := rest[slashIdx+1:]
				// Trim any trailing whitespace or newlines
				if spaceIdx := strings.IndexAny(branch, " \n\t"); spaceIdx > 0 {
					branch = branch[:spaceIdx]
				}
				return branch
			}
		}
	}

	return ""
}

// GitGraph is a git commit graph visualization component
type GitGraph struct {
	widgetBase

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
	columnCap  int // Maximum columns to display (0 = default of 12)
}

// NewGitGraph creates a new git graph component
func NewGitGraph() *GitGraph {
	g := &GitGraph{
		showRefs:   true,
		showHash:   true,
		showAuthor: false,
		showDate:   false,
		dateFormat: "2006-01-02",
	}
	g.initWidget()
	return g
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

// SetColumnCap sets the maximum number of columns/lanes to display
// Default is 12 columns. Set to 0 to use default.
func (g *GitGraph) SetColumnCap(cap int) *GitGraph {
	g.columnCap = cap
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

// GetGraph returns the underlying graph data
func (g *GitGraph) GetGraph() *GitGraphData {
	return g.graph
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
	branchInto    int // if != -1, this lane curves down into the given column (branch point)
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
				branchInto: -1,
			}
		}

		for lane, hash := range activeLanes {
			// A lane whose expected commit lives in a different column is a
			// branch point: this lane curves down into the commit's column.
			if hash == commit.Hash && lane != commit.Column {
				states[row][lane].branchInto = commit.Column
				states[row][lane].commitHash = hash
				delete(activeLanes, lane)
				continue
			}
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
	g.Box.DrawForSubclass(screen)
	x, y, width, height := g.GetInnerRect()

	if width <= 0 || height <= 0 {
		return
	}

	th := g.th()
	bgColor := th.Bg()

	// Fill entire area with background color first
	bgStyle := tcell.StyleDefault.Background(bgColor)
	fillRect(screen, x, y, width, height, bgStyle)

	if g.graph == nil || len(g.graph.Commits) == 0 {
		return
	}

	fgColor := th.Fg()
	fgDimColor := th.FgDim()
	accentColor := th.Accent()

	laneColors := g.laneColors
	if len(laneColors) == 0 {
		laneColors = []tcell.Color{
			th.Accent(),
			th.Success(),
			th.Warning(),
			th.Info(),
			th.Error(),
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

		fillLine(screen, x, rowY, width, rowStyle)

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

		// Calculate space available for content after graph lanes
		availableSpace := (x + width) - col - 2

		// Build branch display text for selected rows without refs
		branchDisplay := ""
		if refText == "" && isSelected && commit.Branch != "" {
			branchDisplay = commit.Branch
			if len(branchDisplay) > 25 {
				branchDisplay = branchDisplay[:22] + "..."
			}
		}

		// Calculate how much space each element needs
		refSpace := len([]rune(refText))
		if refSpace == 0 && branchDisplay != "" {
			refSpace = len(" " + branchDisplay + " ")
		}
		hashSpace := 0
		if g.showHash {
			hashSpace = len(commit.ShortHash) + 1
		}
		authorSpace := 0
		if g.showAuthor {
			authorSpace = len(" - ") + len(commit.Author)
			if authorSpace > 20 {
				authorSpace = 20 // Cap author display
			}
		}

		// Message gets remaining space (minimum 15 chars if possible)
		rightReserved := refSpace + hashSpace
		msgSpace := availableSpace - rightReserved - authorSpace
		if msgSpace < 15 {
			// Not enough space - reduce author/hash to fit message
			authorSpace = 0
			hashSpace = 0
			rightReserved = refSpace
			msgSpace = availableSpace - rightReserved
		}
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

		if g.showAuthor && authorSpace > 0 {
			authorStyle := rowStyle
			if !isSelected {
				authorStyle = tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
			}
			authorText := " - " + commit.Author
			authorRunes := []rune(authorText)
			// Truncate if needed
			if len(authorRunes) > authorSpace {
				authorRunes = append(authorRunes[:authorSpace-1], '…')
			}
			for _, ch := range authorRunes {
				if col < x+width-rightReserved {
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

		if g.showHash && hashSpace > 0 {
			hashStyle := rowStyle
			if !isSelected {
				hashStyle = tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
			}
			hashText := " " + commit.ShortHash
			for _, ch := range hashText {
				if col < x+width-refSpace {
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
		} else if isSelected && branchDisplay != "" {
			// For selected rows without refs, show the branch name dimmed
			branchText := " " + branchDisplay + " "
			branchLen := len([]rune(branchText))
			branchStartCol := x + width - branchLen

			branchStyle := tcell.StyleDefault.Background(accentColor).Foreground(bgColor).Dim(true)

			branchCol := branchStartCol
			for _, ch := range branchText {
				if branchCol < x+width {
					screen.SetContent(branchCol, rowY, ch, nil, branchStyle)
					branchCol++
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

	// Branch point: this lane curves down into another column toward its parent.
	if state != nil && state.branchInto != -1 {
		if state.branchInto < lane {
			return []rune{gitHoriz, gitBotRight, ' '} // ─╯
		}
		return []rune{' ', gitBotLeft, gitHoriz} // ╰─
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
		left, right := ' ', ' '
		// Connect horizontally to any lanes curving into this commit's column.
		for l, s := range allStates {
			if s != nil && s.branchInto == lane {
				if l < lane {
					left = gitHoriz
				} else if l > lane {
					right = gitHoriz
				}
			}
		}
		return []rune{left, node, right}
	}

	// Span horizontally across columns between a branch point and its target.
	for l, s := range allStates {
		if s == nil || s.branchInto == -1 {
			continue
		}
		lo, hi := l, s.branchInto
		if lo > hi {
			lo, hi = hi, lo
		}
		if lane > lo && lane < hi {
			if state != nil && state.hasLine {
				return []rune{gitHoriz, gitCross, gitHoriz}
			}
			return []rune{gitHoriz, gitHoriz, gitHoriz}
		}
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
	if commit.IsPseudoNode {
		return gitUnstaged
	}
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

// HandleKey handles keyboard input
func (g *GitGraph) HandleKey(ev *tcell.EventKey) bool {
	if g.graph == nil || len(g.graph.Commits) == 0 {
		return false
	}

	prevIndex := g.selectedIndex

	switch ev.Key() {
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
		switch ev.Rune() {
		case 'j':
			g.moveDown()
		case 'k':
			g.moveUp()
		case 'g':
			if g.selectedIndex != 0 {
				g.selectedIndex = 0
				g.triggerOnChange()
			}
		case 'G':
			lastIdx := len(g.graph.Commits) - 1
			if g.selectedIndex != lastIdx {
				g.selectedIndex = lastIdx
				g.triggerOnChange()
			}
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
	// Return true if the key matched a known action (even if selection didn't change).
	switch ev.Key() {
	case tcell.KeyDown, tcell.KeyUp, tcell.KeyHome, tcell.KeyEnd, tcell.KeyPgDn, tcell.KeyPgUp, tcell.KeyEnter,
		tcell.KeyCtrlD, tcell.KeyCtrlU:
		return true
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'j', 'k', 'g', 'G', 'p', 'c':
			return true
		}
	}
	return false
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

// Focus handles focus
// HasFocus returns whether the component has focus
func (g *GitGraph) HasFocus() bool {
	return g.Box.HasFocus()
}
