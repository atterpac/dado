package advanced

import (
	"time"

	"github.com/rivo/tview"

	"github.com/atterpac/jig/cmd/tutorial/demos"
	"github.com/atterpac/jig/components"
)

func init() {
	demos.Register(&GitGraphDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "GitGraph",
			DemoDescription: "Git commit history visualization",
			DemoCategory:    demos.Advanced,
			DemoCode:        gitGraphCode,
		},
	})
}

// GitGraphDemo demonstrates the GitGraph component.
type GitGraphDemo struct {
	demos.DemoBase
	graph    *components.GitGraph
	showRefs bool
}

// Component returns the demo component.
func (d *GitGraphDemo) Component() tview.Primitive {
	d.showRefs = true

	d.graph = components.NewGitGraph().
		SetShowRefs(d.showRefs).
		SetShowHash(true).
		SetShowAuthor(true)

	// Create sample git history with multiple branches
	now := time.Now()
	data := components.NewGitGraphData()

	// Build a realistic multi-branch history (newest first)
	// Structure:
	//   main: a1 -- a2 -- merge1 -- a5 -- merge2 -- a8 (HEAD)
	//              \              /      \        /
	//   feature:    b1 -- b2 ----        hotfix: c1
	//
	commits := []*components.GitCommit{
		// HEAD on main
		{Hash: "a8", ShortHash: "a8f3d21", Message: "docs: update changelog", Author: "alice", Date: now, Parents: []string{"merge2"}, Refs: []string{"HEAD", "main"}},

		// Merge hotfix into main
		{Hash: "merge2", ShortHash: "m2e9c45", Message: "Merge branch 'hotfix'", Author: "alice", Date: now.Add(-1 * time.Hour), Parents: []string{"a5", "c1"}, IsMerge: true},

		// Hotfix branch
		{Hash: "c1", ShortHash: "c1b8a72", Message: "fix: critical security patch", Author: "charlie", Date: now.Add(-2 * time.Hour), Parents: []string{"a5"}, Refs: []string{"hotfix"}},

		// Main continues
		{Hash: "a5", ShortHash: "a5d4f89", Message: "chore: bump dependencies", Author: "alice", Date: now.Add(-3 * time.Hour), Parents: []string{"merge1"}},

		// Merge feature into main
		{Hash: "merge1", ShortHash: "m1c7b34", Message: "Merge branch 'feature'", Author: "bob", Date: now.Add(-4 * time.Hour), Parents: []string{"a2", "b2"}, IsMerge: true},

		// Feature branch commits
		{Hash: "b2", ShortHash: "b2e6a91", Message: "feat: add user dashboard", Author: "bob", Date: now.Add(-5 * time.Hour), Parents: []string{"b1"}, Refs: []string{"feature"}},
		{Hash: "b1", ShortHash: "b1f2c83", Message: "feat: implement auth flow", Author: "bob", Date: now.Add(-6 * time.Hour), Parents: []string{"a2"}},

		// Earlier main commits
		{Hash: "a2", ShortHash: "a2c5d76", Message: "refactor: clean up utils", Author: "alice", Date: now.Add(-7 * time.Hour), Parents: []string{"a1"}},
		{Hash: "a1", ShortHash: "a1b4e68", Message: "chore: initial commit", Author: "alice", Date: now.Add(-8 * time.Hour), Tags: []string{"v1.0.0"}},
	}

	for _, c := range commits {
		data.AddCommit(c)
	}

	// Layout computes column positions for branch visualization
	data.LayoutGraph()

	d.graph.SetGraph(data)

	d.Props = []demos.PropertyDescriptor{
		demos.BoolProp("showRefs", "Show branch/tag refs",
			func() bool { return d.showRefs },
			func(v bool) { d.showRefs = v; d.graph.SetShowRefs(v) },
			true,
		),
	}

	return d.graph
}

const gitGraphCode = `package main

import (
    "github.com/atterpac/jig/components"
)

// Create git graph
graph := components.NewGitGraph().
    SetShowRefs(true).
    SetShowHash(true).
    SetShowAuthor(true)

// Build graph data
data := components.NewGitGraphData()

// Add commits (newest first, with parent relationships)
// Main branch commits
data.AddCommit(&components.GitCommit{
    Hash:      "abc123",
    ShortHash: "abc123",
    Message:   "Merge feature branch",
    Parents:   []string{"def456", "feat2"}, // Two parents = merge
    IsMerge:   true,
    Refs:      []string{"HEAD", "main"},
})

// Feature branch (will show in separate column)
data.AddCommit(&components.GitCommit{
    Hash:      "feat2",
    ShortHash: "feat2",
    Message:   "feat: complete feature",
    Parents:   []string{"feat1"},
    Refs:      []string{"feature"},
})
data.AddCommit(&components.GitCommit{
    Hash:      "feat1",
    ShortHash: "feat1",
    Message:   "feat: start feature",
    Parents:   []string{"def456"},
})

// Main branch continues
data.AddCommit(&components.GitCommit{
    Hash:      "def456",
    ShortHash: "def456",
    Message:   "chore: update deps",
    Parents:   []string{"ghi789"},
})
data.AddCommit(&components.GitCommit{
    Hash:      "ghi789",
    ShortHash: "ghi789",
    Message:   "Initial commit",
    Tags:      []string{"v1.0.0"},
})

// IMPORTANT: Call LayoutGraph to compute branch columns
data.LayoutGraph()

// Load into graph
graph.SetGraph(data)

// Callbacks
graph.SetOnSelect(func(commit *components.GitCommit) {
    showCommitDetails(commit.Hash)
})
`
