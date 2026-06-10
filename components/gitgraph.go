package components

import (
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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
	// Per-row lane state stored as dense fixed-width (columnCap) slices carved
	// from one backing array, instead of a map per row. Lane indices are always
	// < columnCap, so indexing is safe and this avoids numRows map allocations
	// plus per-cell hashing.
	statesBacking := make([]*gitLaneState, numRows*columnCap)
	states := make([][]*gitLaneState, numRows)
	for i := range states {
		states[i] = statesBacking[i*columnCap : (i+1)*columnCap : (i+1)*columnCap]
	}
	// active: column -> parent hash the lane is currently flowing toward.
	active := make(map[int]string)
	// incoming is reused across rows to avoid a per-row slice allocation.
	var incoming []int

	// Arena for gitLaneState. Layout allocates one state per live lane per row
	// plus one per empty cell when materializing the dense grid; for a 500-commit
	// graph that is thousands of tiny allocations. Handing out pointers into
	// reusable blocks collapses them to one allocation per ~512 states. The
	// blocks are never grown after a pointer is taken, so the pointers stay valid.
	var arena []gitLaneState
	newState := func() *gitLaneState {
		if len(arena) == cap(arena) {
			arena = make([]gitLaneState, 0, 512)
		}
		arena = append(arena, gitLaneState{mergeFrom: -1, mergeTo: -1, branchFrom: -1, branchInto: -1})
		return &arena[len(arena)-1]
	}

	// Line-identity coloring: lineID maps an active column to the color id of
	// the line currently flowing through it. A new id is minted whenever a
	// fresh line starts in a column (a branch tip or a merge source) and is
	// carried along while the line continues, so a column reused by a different
	// branch gets a different color. Ids start at 0 so the first line (usually
	// the current branch / HEAD) keeps the leading palette color.
	lineSeq := 0
	lineID := make(map[int]int)
	nextID := func() int { id := lineSeq; lineSeq++; return id }

	for i, commit := range g.Commits {
		commit.Row = i
		row := states[i]
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
		incoming = incoming[:0]
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

		// Determine the color id of the line this commit sits on. When lanes
		// were already flowing into this column, the commit inherits that
		// line's id; otherwise it starts a brand-new line.
		if _, ok := lineID[col]; !ok {
			lineID[col] = nextID()
		}
		colorID := lineID[col]

		// Pass-through lines and branch-point convergence.
		for lane, h := range active {
			if h == commit.Hash {
				if lane != col {
					// Extra lane expecting this commit curves down into its column.
					s := st(lane)
					s.branchInto = col
					s.commitHash = h
					s.colorID = lineID[lane]
					delete(active, lane)
					delete(lineID, lane)
				}
				continue
			}
			s := st(lane)
			s.hasLine = true
			s.commitHash = h
			s.colorID = lineID[lane]
		}

		// Place the node and free its lane (the first parent repopulates it).
		node := st(col)
		node.hasNode = true
		node.commitHash = commit.Hash
		node.colorID = colorID
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
		firstParentContinues := false
		for idx, parentHash := range commit.Parents {
			if g.CommitMap[parentHash] == nil {
				continue
			}
			if idx == 0 {
				// First parent continues this commit's column and line.
				active[col] = parentHash
				lineID[col] = colorID
				firstParentContinues = true
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
				lineID[plane] = nextID() // merge source starts its own line
				s := st(plane)
				s.isStartOfLine = true
				s.branchFrom = col
				s.colorID = lineID[plane]
			}
			// Merge edge: from the parent's lane into the commit's column.
			st(plane).mergeTo = col
			st(plane).colorID = lineID[plane]
			st(col).mergeFrom = plane
		}
		// If no first parent kept this column alive, the line ends here; drop
		// its id so a later branch reusing the column mints a fresh color.
		if !firstParentContinues {
			delete(lineID, col)
		}

		// Track the widest column seen, in either this row or the live lanes.
		for lane := 0; lane < len(row); lane++ {
			if row[lane] != nil && lane > g.MaxColumn {
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

	// Fill empty cells in place and trim each row to MaxColumn+1. The backing
	// array is already dense (columnCap wide), so no second allocation/copy is
	// needed; the renderer reads rowStates[i][0..MaxColumn] with no nils.
	dense := make([][]*gitLaneState, numRows)
	for i := range states {
		row := states[i]
		for lane := 0; lane <= g.MaxColumn; lane++ {
			if row[lane] == nil {
				row[lane] = newState()
			}
		}
		dense[i] = row[:g.MaxColumn+1]
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

	// dateBuf is a reusable scratch buffer for formatting commit dates during
	// Draw, so date.AppendFormat reuses storage instead of allocating a fresh
	// string per visible row each frame.
	dateBuf []byte

	// laneColorBuf is a reusable buffer for the default lane palette, rebuilt
	// (in place) from the theme each frame only when no custom colors are set,
	// so the fallback path no longer allocates a slice every Draw.
	laneColorBuf []tcell.Color

	// Style options
	showRefs   bool
	showHash   bool
	showAuthor bool
	showDate   bool
	dateFormat string
	laneColors []tcell.Color
	columnCap  int // Maximum columns to display (0 = default of 12)
	maxRefLen  int // Max displayed length of a branch/tag name (0 = default of 20)
}

// defaultMaxRefLen caps how many runes of a branch/tag name are shown in the
// ref decoration before it is truncated with an ellipsis.
const defaultMaxRefLen = 20

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

// SetMaxRefLen sets the maximum displayed length (in runes) of a branch/tag
// name in the ref decoration; longer names are truncated with an ellipsis.
// Default is 20. Set to 0 to use the default; set negative to disable truncation.
func (g *GitGraph) SetMaxRefLen(n int) *GitGraph {
	g.maxRefLen = n
	g.refTextCache = nil
	return g
}

// truncateRef shortens a ref name to the configured max length, appending an
// ellipsis when it overflows.
func (g *GitGraph) truncateRef(name string) string {
	max := g.maxRefLen
	if max == 0 {
		max = defaultMaxRefLen
	}
	if max < 0 {
		return name
	}
	r := []rune(name)
	if len(r) <= max {
		return name
	}
	if max == 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
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
// When debouncing, the fired callback is marshaled back onto the draw goroutine
// via theme.QueueUpdateDraw, so onChange (and the selection read it sees) runs
// with the same threading guarantees as the synchronous path — no extra locking
// is required in the callback.
func (g *GitGraph) SetChangeDebounce(d time.Duration) *GitGraph {
	g.changeDebounce = d
	// Register a one-time teardown that stops any pending timer when the
	// component is stopped, so a debounce in flight can't fire into a
	// torn-down widget or leak its goroutine.
	if d > 0 && !g.debounceCleanup {
		g.debounceCleanup = true
		g.Subs().Add(func() {
			g.mu.Lock()
			if g.changeTimer != nil {
				g.changeTimer.Stop()
				g.changeTimer = nil
			}
			g.mu.Unlock()
		})
	}
	return g
}

// triggerOnChange notifies the onChange callback of a selection change. It is
// only ever called from HandleKey/SetSelectedIndex on the draw goroutine, so
// changeTimer is single-goroutine state needing no lock. When debouncing, the
// timer fires on its own goroutine but immediately hands work back to the draw
// goroutine via theme.QueueUpdateDraw, where GetSelected/onChange run safely.
func (g *GitGraph) triggerOnChange() {
	if g.onChange == nil {
		return
	}
	if g.changeDebounce <= 0 {
		g.onChange(g.GetSelected())
		return
	}
	g.mu.Lock()
	if g.changeTimer != nil {
		g.changeTimer.Stop()
	}
	g.changeTimer = time.AfterFunc(g.changeDebounce, func() {
		theme.QueueUpdateDraw(func() {
			g.onChange(g.GetSelected())
		})
	})
	g.mu.Unlock()
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
	// colorID identifies the *line* occupying this lane, not the lane index.
	// A fresh id is assigned each time a new line starts in a column, so a
	// column reused by a later branch draws in a different color — making it
	// clear where one branch ends and another begins in the same column.
	colorID int
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
			refParts = append(refParts, "↕ "+g.truncateRef(ref)+aheadBehind)
		} else {
			refParts = append(refParts, "● "+g.truncateRef(ref))
		}
	}

	// Remote-only refs (no matching local branch).
	for _, ref := range commit.Refs {
		if localName, ok := strings.CutPrefix(ref, "origin/"); ok && !slices.Contains(commit.Refs, localName) {
			refParts = append(refParts, "○ "+g.truncateRef(localName))
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
		// Rebuild the default palette into the reused buffer rather than
		// allocating a new slice every frame (keeps Draw allocation-free while
		// still tracking theme changes).
		g.laneColorBuf = append(g.laneColorBuf[:0],
			th.Accent(),
			th.Success(),
			th.Warning(),
			th.Info(),
			th.Error(),
			tcell.ColorMediumPurple,
			tcell.ColorDarkCyan,
		)
		laneColors = g.laneColorBuf
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
			// Color follows the line occupying the lane, not the lane index, so
			// a column reused by a different branch is visibly a different color.
			laneColor := laneColors[state.colorID%len(laneColors)]

			cellStyle := rowStyle
			if !isSelected {
				cellStyle = tcell.StyleDefault.Background(bgColor).Foreground(laneColor)
			}

			chars := g.getLaneChars(commit, lane, state, states, rowIndex, rowStates)

			for _, ch := range chars {
				if col < x+width {
					screen.SetContent(col, rowY, ch, nil, cellStyle)
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
			branchDisplay = g.truncateRef(commit.Branch)
		}

		// Calculate how much space each element needs
		refSpace := utf8.RuneCountInString(refText)
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
		msgStyle := rowStyle
		if msgSpace > 0 {
			if utf8.RuneCountInString(commit.Message) > msgSpace {
				// Truncate with ellipsis. Range over the string (alloc-free)
				// rather than converting to []rune.
				n := 0
				for _, ch := range commit.Message {
					if n >= msgSpace-1 || col >= x+width {
						break
					}
					screen.SetContent(col, rowY, ch, nil, msgStyle)
					col++
					n++
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
			// Stream " - " then the author name, truncating to authorSpace runes
			// with a trailing ellipsis. Done inline (no []rune, no concat) to
			// keep Draw allocation-free.
			trunc := 3+utf8.RuneCountInString(commit.Author) > authorSpace // " - " is 3 runes
			n := 0
			done := false
			for _, ch := range " - " {
				if col >= x+width-rightReserved {
					done = true
					break
				}
				if trunc && n == authorSpace-1 {
					screen.SetContent(col, rowY, '…', nil, authorStyle)
					col++
					done = true
					break
				}
				screen.SetContent(col, rowY, ch, nil, authorStyle)
				col++
				n++
			}
			if !done {
				for _, ch := range commit.Author {
					if col >= x+width-rightReserved {
						break
					}
					if trunc && n == authorSpace-1 {
						screen.SetContent(col, rowY, '…', nil, authorStyle)
						col++
						break
					}
					screen.SetContent(col, rowY, ch, nil, authorStyle)
					col++
					n++
				}
			}
		}

		if g.showDate {
			dateStyle := rowStyle
			if !isSelected {
				dateStyle = tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
			}
			if col < x+width {
				screen.SetContent(col, rowY, ' ', nil, dateStyle)
				col++
			}
			// AppendFormat reuses dateBuf; decode runes from the bytes without
			// allocating an intermediate string.
			g.dateBuf = commit.Date.AppendFormat(g.dateBuf[:0], g.dateFormat)
			for bi := 0; bi < len(g.dateBuf); {
				r, sz := utf8.DecodeRune(g.dateBuf[bi:])
				bi += sz
				if col < x+width {
					screen.SetContent(col, rowY, r, nil, dateStyle)
					col++
				}
			}
		}

		if g.showHash && hashSpace > 0 {
			hashStyle := rowStyle
			if !isSelected {
				hashStyle = tcell.StyleDefault.Background(bgColor).Foreground(fgDimColor)
			}
			if col < x+width-refSpace {
				screen.SetContent(col, rowY, ' ', nil, hashStyle)
				col++
			}
			for _, ch := range commit.ShortHash {
				if col < x+width-refSpace {
					screen.SetContent(col, rowY, ch, nil, hashStyle)
					col++
				}
			}
		}

		// Right-aligned refs with lane color matching (refText already built above)
		if refText != "" {
			refLen := utf8.RuneCountInString(refText)
			refStartCol := x + width - refLen

			// Always render refs - they take priority. Match the commit's line
			// color (by colorID, same as the node) rather than its raw column.
			laneColor := laneColors[states[commit.Column].colorID%len(laneColors)]
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
			// For selected rows without refs, show the branch name dimmed,
			// padded with a space on each side. Written piecewise to avoid the
			// " "+branch+" " concat and []rune conversion.
			branchLen := utf8.RuneCountInString(branchDisplay) + 2
			branchStartCol := x + width - branchLen

			branchStyle := tcell.StyleDefault.Background(accentColor).Foreground(bgColor).Dim(true)

			branchCol := branchStartCol
			if branchCol < x+width {
				screen.SetContent(branchCol, rowY, ' ', nil, branchStyle)
				branchCol++
			}
			for _, ch := range branchDisplay {
				if branchCol < x+width {
					screen.SetContent(branchCol, rowY, ch, nil, branchStyle)
					branchCol++
				}
			}
			if branchCol < x+width {
				screen.SetContent(branchCol, rowY, ' ', nil, branchStyle)
				branchCol++
			}
		}
	}
}

// getLaneChars returns the 3 characters to draw for a lane
func (g *GitGraph) getLaneChars(commit *GitCommit, lane int, state *gitLaneState, allStates []*gitLaneState, rowIndex int, allRowStates [][]*gitLaneState) [3]rune {
	if state != nil && (state.mergeFrom != -1 || state.mergeTo != -1) {
		if state.hasNode {
			node := g.getNodeChar(commit)
			if state.mergeFrom != -1 {
				if state.mergeFrom < lane {
					return [3]rune{gitHoriz, node, ' '}
				} else if state.mergeFrom > lane {
					return [3]rune{' ', node, gitHoriz}
				}
			}
			if state.mergeTo != -1 {
				if state.mergeTo < lane {
					return [3]rune{gitHoriz, node, ' '}
				} else if state.mergeTo > lane {
					return [3]rune{' ', node, gitHoriz}
				}
			}
			return [3]rune{' ', node, ' '}
		}

		if state.mergeTo != -1 {
			if state.mergeTo < lane {
				if state.hasLine {
					return [3]rune{gitHoriz, gitVertLeft, ' '}
				}
				return [3]rune{gitHoriz, gitTopRight, ' '}
			} else if state.mergeTo > lane {
				if state.hasLine {
					return [3]rune{' ', gitVertRight, gitHoriz}
				}
				return [3]rune{' ', gitTopLeft, gitHoriz}
			}
		}

		if state.hasLine {
			return [3]rune{' ', gitVert, ' '}
		}
		return [3]rune{' ', ' ', ' '}
	}

	// Branch point: this lane curves down into another column toward its parent.
	if state != nil && state.branchInto != -1 {
		if state.branchInto < lane {
			return [3]rune{gitHoriz, gitBotRight, ' '} // ─╯
		}
		return [3]rune{' ', gitBotLeft, gitHoriz} // ╰─
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
		return [3]rune{left, node, right}
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
		return [3]rune{left, node, right}
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
				return [3]rune{gitHoriz, gitCross, gitHoriz}
			}
			return [3]rune{gitHoriz, gitHoriz, gitHoriz}
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
						return [3]rune{' ', gitVertRight, gitHoriz}
					}
					return [3]rune{' ', gitTopLeft, gitHoriz}
				}
				if state.hasLine {
					return [3]rune{gitHoriz, gitVertLeft, ' '}
				}
				return [3]rune{gitHoriz, gitTopRight, ' '}
			}

			if lane > minLane && lane < maxLane {
				if state.hasLine {
					return [3]rune{gitHoriz, gitCross, gitHoriz}
				}
				return [3]rune{gitHoriz, gitHoriz, gitHoriz}
			}
		}
	}

	if state != nil && state.hasLine {
		return [3]rune{' ', gitVert, ' '}
	}

	if state != nil && state.isStartOfLine {
		if state.branchFrom != -1 {
			if state.branchFrom < lane {
				return [3]rune{gitHoriz, gitTopRight, ' '}
			} else {
				return [3]rune{' ', gitTopLeft, gitHoriz}
			}
		}
	}

	return [3]rune{' ', ' ', ' '}
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
