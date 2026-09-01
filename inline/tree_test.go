package inline

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTreeRendersHierarchyAndDetails(t *testing.T) {
	t.Parallel()

	tree := NewTree("Files",
		TreeNode{Label: "cmd", Children: []TreeNode{{Label: "dado", Detail: "binary"}}},
		TreeNode{Label: "go.mod", Detail: "module"},
	)
	lines := plainFrameLines(tree.Frame(40))
	assert.Equal(t, []string{
		"Files",
		"├─ cmd",
		"│ └─ dado binary",
		"└─ go.mod module",
	}, lines)
}

func TestTreeASCIIThemeAndNarrowFrames(t *testing.T) {
	t.Parallel()

	tree := NewTree("Tree", TreeNode{Label: "parent", Children: []TreeNode{{Label: "a very long child"}}}).SetTheme(ASCIIInlineTheme())
	assert.Contains(t, strings.Join(plainFrameLines(tree.Frame(32)), "\n"), "`- a very long child")
	for width := range 16 {
		for _, line := range plainFrameLines(tree.Frame(width)) {
			assert.LessOrEqual(t, displayWidth(line), width)
		}
	}
}

func TestTreeConcurrentUpdatesAndFrames(t *testing.T) {
	t.Parallel()

	tree := NewTree("Concurrent")
	var workers sync.WaitGroup
	for worker := range 4 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range 50 {
				tree.Add(TreeNode{Label: string(rune('a' + worker)), Detail: string(rune('0' + index%10))})
				_ = tree.Frame(24)
			}
		}()
	}
	workers.Wait()
	assert.Len(t, plainFrameLines(tree.Frame(24)), 201)
}
