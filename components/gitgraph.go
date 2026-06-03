package components

import (
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/atterpac/dado/theme"
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

	// rowStates caches the per-row lane render state produced by LayoutGraph as
	// dense slices indexed 0..MaxColumn (no nil entries). Computed once per
	// layout and reused on every Draw, so scrolling never recomputes the graph
	// and the render hot path uses slice indexing instead of map lookups.
	rowStates [][]*gitLaneState
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

// lowestFreeColumn returns the smallest column index in [0, cap) that is not
// currently occupied by an active line. When every column under the cap is
// taken it falls back to the last column so lanes collapse instead of growing
// without bound.
func lowestFreeColumn(active map[int]string, cap int) int {
	for c := 0; c < cap; c++ {
		if _, used := active[c]; !used {
			return c
		}
	}
	return cap - 1
}

// LayoutGraph assigns columns to commits and builds the per-row render state in
// a single deterministic pass. Lanes are reclaimed (lowest free column reused)
// as branches terminate, so short-lived branches no longer exhaust the column
// cap. The resulting lane states are cached on the data so Draw never recomputes
// them while scrolling.
func (g *GitGraphData) LayoutGraph() {
	g.rowStates = nil
	g.MaxColumn = 0
	if len(g.Commits) == 0 {
		return
	}

	// Default column cap if not set
	columnCap := g.ColumnCap
	if columnCap <= 0 {
		columnCap = 12 // Default to 12 columns max (36 chars for graph)
	}

	g.ColumnBranch = make(map[int]string)

	// Find the current branch tip so we can prefer column 0 for it.
	tipHash := ""
	if g.CurrentBranch != "" {
		for _, commit := range g.Commits {
			for _, ref := range commit.Refs {
				if ref == g.CurrentBranch || ref == "HEAD" {
					tipHash = commit.Hash
					break
				}
			}
			if tipHash != "" {
				break
			}
		}
	}

	numRows := len(g.Commits)
	states := make([]map[int]*gitLaneState, numRows)
	// active: column -> parent hash the lane is currently flowing toward.
	active := make(map[int]string)

	newState := func() *gitLaneState {
		return &gitLaneState{mergeFrom: -1, mergeTo: -1, branchFrom: -1, branchInto: -1}
	}

	for i, commit := range g.Commits {
		commit.Row = i
		row := make(map[int]*gitLaneState)
		states[i] = row
		st := func(lane int) *gitLaneState {
			s := row[lane]
			if s == nil {
				s = newState()
				row[lane] = s
			}
			return s
		}

		// Collect every lane currently expecting this commit, sorted so the
		// choice of column is deterministic regardless of map iteration order.
		var incoming []int
		for lane, h := range active {
			if h == commit.Hash {
				incoming = append(incoming, lane)
			}
		}
		sort.Ints(incoming)

		// Choose the commit's column.
		var col int
		switch {
		case len(incoming) > 0:
			col = incoming[0]
		case commit.Hash == tipHash:
			if _, used := active[0]; !used {
				col = 0
			} else {
				col = lowestFreeColumn(active, columnCap)
			}
		default:
			col = lowestFreeColumn(active, columnCap)
		}
		commit.Column = col

		// Pass-through lines and branch-point convergence.
		for lane, h := range active {
			if h == commit.Hash {
				if lane != col {
					// Extra lane expecting this commit curves down into its column.
					s := st(lane)
					s.branchInto = col
					s.commitHash = h
					delete(active, lane)
				}
				continue
			}
			s := st(lane)
			s.hasLine = true
			s.commitHash = h
		}

		// Place the node and free its lane (the first parent repopulates it).
		node := st(col)
		node.hasNode = true
		node.commitHash = commit.Hash
		delete(active, col)

		// Record this column's branch name from refs if not already set.
		if g.ColumnBranch[col] == "" {
			for _, ref := range commit.Refs {
				if !strings.HasPrefix(ref, "origin/") && ref != "HEAD" {
					g.ColumnBranch[col] = ref
					break
				}
			}
		}

		// Route parents into lanes.
		for idx, parentHash := range commit.Parents {
			if g.CommitMap[parentHash] == nil {
				continue
			}
			if idx == 0 {
				// First parent continues this commit's column.
				active[col] = parentHash
				continue
			}
			// Secondary parent (merge source). Reuse a lane already flowing to
			// it (lowest), otherwise allocate the lowest free column.
			plane := -1
			for lane, h := range active {
				if h == parentHash && (plane == -1 || lane < plane) {
					plane = lane
				}
			}
			if plane == -1 {
				plane = lowestFreeColumn(active, columnCap)
				active[plane] = parentHash
				s := st(plane)
				s.isStartOfLine = true
				s.branchFrom = col
			}
			// Merge edge: from the parent's lane into the commit's column.
			st(plane).mergeTo = col
			st(col).mergeFrom = plane
		}

		// Track the widest column seen, in either this row or the live lanes.
		for lane := range row {
			if lane > g.MaxColumn {
				g.MaxColumn = lane
			}
		}
		for lane := range active {
			if lane > g.MaxColumn {
				g.MaxColumn = lane
			}
		}
	}

	if g.MaxColumn > columnCap-1 {
		g.MaxColumn = columnCap - 1
	}

	// Materialize each row's sparse lane map into a dense slice indexed
	// 0..MaxColumn (no nil entries). The renderer indexes lanes directly and
	// iterates them every frame, so slices avoid per-cell map hashing.
	dense := make([][]*gitLaneState, numRows)
	for i, row := range states {
		laneSlice := make([]*gitLaneState, g.MaxColumn+1)
		for lane := 0; lane <= g.MaxColumn; lane++ {
			if s := row[lane]; s != nil {
				laneSlice[lane] = s
			} else {
				laneSlice[lane] = newState()
			}
		}
		dense[i] = laneSlice
	}
	g.rowStates = dense

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

	// changeDebounce, when > 0, coalesces rapid selection changes (e.g. holding
	// j/k) so onChange fires only once movement settles, instead of once per
	// step. Zero (the default) fires onChange synchronously on every change.
	changeDebounce time.Duration
	changeTimer    *time.Timer
	// debounceCleanup records that a Subs() cleanup stopping the pending timer
	// has been registered, so we register it exactly once.
	debounceCleanup bool

	// refTextCache holds the pre-rendered right-aligned ref string for each
	// commit row. It is width-independent, so it is built once and reused across
	// frames; nil means it needs rebuilding (after SetGraph or a showRefs change).
	refTextCache []string

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
	g.refTextCache = nil
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
	g.refTextCache = nil
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

// SetChangeDebounce coalesces selection-change callbacks during rapid
// navigation. With d > 0, onChange fires once the selection stops changing for
// d (useful when onChange does expensive work like loading a diff). With d == 0
// (the default) onChange fires synchronously on every change.
//
// When debouncing, onChange is invoked from a timer goroutine, so the callback
// must marshal any UI work back onto the draw goroutine (e.g. App.QueueUpdateDraw).
func (g *GitGraph) SetChangeDebounce(d time.Duration) *GitGraph {
	g.changeDebounce = d
	return g
}

func (g *GitGraph) triggerOnChange() {
	if g.onChange == nil {
		return
	}
	if g.changeDebounce <= 0 {
		g.onChange(g.GetSelected())
		return
	}
	if g.changeTimer != nil {
		g.changeTimer.Stop()
	}
	g.changeTimer = time.AfterFunc(g.changeDebounce, func() {
		g.onChange(g.GetSelected())
	})
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

// buildRowStates returns the per-row lane states computed by LayoutGraph.
// The states are cached on the graph data, so this is cheap to call every
// frame. If the data was never laid out, it is laid out on demand.
func (g *GitGraph) buildRowStates() [][]*gitLaneState {
	if g.graph == nil || len(g.graph.Commits) == 0 {
		return nil
	}
	if g.graph.rowStates == nil {
		g.graph.LayoutGraph()
	}
	return g.graph.rowStates
}

// buildRefTexts returns the cached per-row ref strings, building them on first
// use. The strings depend only on commit refs and the showRefs flag, not on
// width or scroll, so they are reused across frames. Invalidated (set nil) by
// SetGraph and SetShowRefs.
func (g *GitGraph) buildRefTexts() []string {
	if g.refTextCache != nil {
		return g.refTextCache
	}
	cache := make([]string, len(g.graph.Commits))
	if g.showRefs {
		for i, commit := range g.graph.Commits {
			cache[i] = g.formatRefText(commit)
		}
	}
	g.refTextCache = cache
	return cache
}

// formatRefText renders the right-aligned ref decoration for a single commit
// (branches, tags, remote-tracking arrows, ahead/behind counts).
func (g *GitGraph) formatRefText(commit *GitCommit) string {
	if len(commit.Refs) == 0 {
		return ""
	}

	// Build a set of remote refs for quick lookup.
	remoteRefs := make(map[string]bool)
	for _, ref := range commit.Refs {
		if name, ok := strings.CutPrefix(ref, "origin/"); ok {
			remoteRefs[name] = true
		}
	}

	// Build ahead/behind indicator.
	aheadBehind := ""
	switch {
	case commit.Ahead > 0 && commit.Behind > 0:
		aheadBehind = " ↑" + strconv.Itoa(commit.Ahead) + "↓" + strconv.Itoa(commit.Behind)
	case commit.Ahead > 0:
		aheadBehind = " ↑" + strconv.Itoa(commit.Ahead)
	case commit.Behind > 0:
		aheadBehind = " ↓" + strconv.Itoa(commit.Behind)
	}

	var refParts []string
	for _, ref := range commit.Refs {
		if strings.HasPrefix(ref, "origin/") || ref == "HEAD" {
			continue
		}
		if remoteRefs[ref] {
			refParts = append(refParts, "↕ "+ref+aheadBehind)
		} else {
			refParts = append(refParts, "● "+ref)
		}
	}

	// Remote-only refs (no matching local branch).
	for _, ref := range commit.Refs {
		if localName, ok := strings.CutPrefix(ref, "origin/"); ok && !slices.Contains(commit.Refs, localName) {
			refParts = append(refParts, "○ "+localName)
		}
	}

	if len(refParts) == 0 {
		return ""
	}
	return " " + strings.Join(refParts, "  ") + " "
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
	refTexts := g.buildRefTexts()

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

		// Ref text is independent of width/scroll, so it is built once and
		// cached rather than rebuilt for every visible row each frame.
		refText := refTexts[rowIndex]

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
func (g *GitGraph) getLaneChars(commit *GitCommit, lane int, state *gitLaneState, allStates []*gitLaneState, rowIndex int, allRowStates [][]*gitLaneState) []rune {
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
		// Span toward every merge source, not just the first, so octopus
		// merges connect on both sides of the node.
		secCols := g.secondaryParentCols(commit)
		left, right := ' ', ' '
		for _, c := range secCols {
			if c < lane {
				left = gitHoriz
			} else if c > lane {
				right = gitHoriz
			}
		}
		return []rune{left, node, right}
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
		secCols := g.secondaryParentCols(commit)
		if len(secCols) > 0 {
			minLane, maxLane := commit.Column, commit.Column
			isSec := false
			for _, c := range secCols {
				if c < minLane {
					minLane = c
				}
				if c > maxLane {
					maxLane = c
				}
				if c == lane {
					isSec = true
				}
			}

			if isSec {
				if lane < commit.Column {
					if state.hasLine {
						return []rune{' ', gitVertRight, gitHoriz}
					}
					return []rune{' ', gitTopLeft, gitHoriz}
				}
				if state.hasLine {
					return []rune{gitHoriz, gitVertLeft, ' '}
				}
				return []rune{gitHoriz, gitTopRight, ' '}
			}

			if lane > minLane && lane < maxLane {
				if state.hasLine {
					return []rune{gitHoriz, gitCross, gitHoriz}
				}
				return []rune{gitHoriz, gitHoriz, gitHoriz}
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

// secondaryParentCols returns the assigned columns of all merge-source parents
// (every parent after the first) that exist in the loaded commit set.
func (g *GitGraph) secondaryParentCols(commit *GitCommit) []int {
	if !commit.IsMerge || len(commit.Parents) < 2 {
		return nil
	}
	cols := make([]int, 0, len(commit.Parents)-1)
	for i := 1; i < len(commit.Parents); i++ {
		if p := g.graph.GetCommit(commit.Parents[i]); p != nil {
			cols = append(cols, p.Column)
		}
	}
	return cols
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
