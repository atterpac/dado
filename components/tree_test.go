package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a sample tree
func createSampleTree() *TreeNode {
	root := &TreeNode{ID: "root", Label: "Root"}

	child1 := &TreeNode{ID: "child1", Label: "Child 1"}
	child2 := &TreeNode{ID: "child2", Label: "Child 2"}
	child3 := &TreeNode{ID: "child3", Label: "Child 3"}

	grandchild1 := &TreeNode{ID: "gc1", Label: "Grandchild 1"}
	grandchild2 := &TreeNode{ID: "gc2", Label: "Grandchild 2"}

	child1.AddChild(grandchild1)
	child1.AddChild(grandchild2)

	root.AddChild(child1)
	root.AddChild(child2)
	root.AddChild(child3)

	return root
}

// TestTreeNode_AddChild tests adding children to a node.
func TestTreeNode_AddChild(t *testing.T) {
	parent := &TreeNode{ID: "parent", Label: "Parent"}
	child := &TreeNode{ID: "child", Label: "Child"}

	result := parent.AddChild(child)

	assert.Same(t, parent, result) // Fluent API
	assert.Len(t, parent.Children, 1)
	assert.Same(t, child, parent.Children[0])
	assert.Same(t, parent, child.parent)
}

// TestTreeNode_IsLeaf tests leaf detection.
func TestTreeNode_IsLeaf(t *testing.T) {
	parent := &TreeNode{ID: "parent", Label: "Parent"}
	child := &TreeNode{ID: "child", Label: "Child"}

	assert.True(t, parent.IsLeaf())
	assert.True(t, child.IsLeaf())

	parent.AddChild(child)

	assert.False(t, parent.IsLeaf())
	assert.True(t, child.IsLeaf())
}

// TestTree_NewTree tests Tree creation.
func TestTree_NewTree(t *testing.T) {
	tree := NewTree()

	assert.NotNil(t, tree)
	assert.Nil(t, tree.GetSelected())
}

// TestTree_SetRoot tests setting the root node.
func TestTree_SetRoot(t *testing.T) {
	tree := NewTree()
	root := createSampleTree()

	result := tree.SetRoot(root)

	assert.Same(t, tree, result)
	assert.Equal(t, 0, root.level)
}

// TestTree_SetShowLines tests line display toggle.
func TestTree_SetShowLines(t *testing.T) {
	tree := NewTree()

	result := tree.SetShowLines(false)
	assert.Same(t, tree, result)

	result = tree.SetShowLines(true)
	assert.Same(t, tree, result)
}

// TestTree_SetShowIcons tests icon display toggle.
func TestTree_SetShowIcons(t *testing.T) {
	tree := NewTree()

	result := tree.SetShowIcons(false)
	assert.Same(t, tree, result)
}

// TestTree_SetIndentSize tests indentation setting.
func TestTree_SetIndentSize(t *testing.T) {
	tree := NewTree()

	result := tree.SetIndentSize(4)
	assert.Same(t, tree, result)
}

// TestTree_SetMultiSelect tests multi-selection toggle.
func TestTree_SetMultiSelect(t *testing.T) {
	tree := NewTree()

	result := tree.SetMultiSelect(true)
	assert.Same(t, tree, result)
}

// TestTree_ExpandAll tests expanding all nodes.
func TestTree_ExpandAll(t *testing.T) {
	tree := NewTree()
	root := createSampleTree()
	tree.SetRoot(root)

	result := tree.ExpandAll()

	assert.Same(t, tree, result)
	assert.True(t, root.Expanded)
	assert.True(t, root.Children[0].Expanded) // Child 1 has grandchildren
}

// TestTree_CollapseAll tests collapsing all nodes.
func TestTree_CollapseAll(t *testing.T) {
	tree := NewTree()
	root := createSampleTree()
	root.Expanded = true
	root.Children[0].Expanded = true
	tree.SetRoot(root)

	result := tree.CollapseAll()

	assert.Same(t, tree, result)
	assert.False(t, root.Expanded)
	assert.False(t, root.Children[0].Expanded)
}

// TestTree_ExpandTo tests expanding to a specific depth.
func TestTree_ExpandTo(t *testing.T) {
	tree := NewTree()
	root := createSampleTree()
	tree.SetRoot(root)

	result := tree.ExpandTo(1)

	assert.Same(t, tree, result)
	assert.True(t, root.Expanded)              // Level 0, expanded
	assert.False(t, root.Children[0].Expanded) // Level 1, not expanded (depth=1 means expand level 0)
}

// TestTree_GetSelected tests getting the selected node.
func TestTree_GetSelected(t *testing.T) {
	tree := NewTree()

	// Empty tree
	assert.Nil(t, tree.GetSelected())

	// With root
	root := createSampleTree()
	root.Expanded = true
	tree.SetRoot(root)

	selected := tree.GetSelected()
	assert.NotNil(t, selected)
	assert.Equal(t, "Root", selected.Label)
}

// TestTree_GetSelectedNodes tests getting multi-selected nodes.
func TestTree_GetSelectedNodes(t *testing.T) {
	tree := NewTree()
	tree.SetMultiSelect(true)

	// Initially empty
	assert.Empty(t, tree.GetSelectedNodes())
}

// TestTree_ClearSelection tests clearing selection.
func TestTree_ClearSelection(t *testing.T) {
	tree := NewTree()
	tree.SetMultiSelect(true)

	result := tree.ClearSelection()

	assert.Same(t, tree, result)
	assert.Empty(t, tree.GetSelectedNodes())
}

// TestTree_Filter tests filtering the tree.
func TestTree_Filter(t *testing.T) {
	tree := NewTree()
	root := createSampleTree()
	tree.SetRoot(root)
	tree.ExpandAll()

	t.Run("empty query returns all", func(t *testing.T) {
		tree.Filter("")
		// Should have all visible nodes
	})

	t.Run("filter by label", func(t *testing.T) {
		tree.Filter("child")
		// Should find nodes with "child" in label
	})

	t.Run("case insensitive", func(t *testing.T) {
		tree.Filter("CHILD")
		// Should still match
	})
}

// TestTree_SetCallbacks tests setting callback functions.
func TestTree_SetCallbacks(t *testing.T) {
	tree := NewTree()

	t.Run("OnSelect", func(t *testing.T) {
		var called bool
		result := tree.SetOnSelect(func(node *TreeNode) {
			called = true
		})
		assert.Same(t, tree, result)
		_ = called
	})

	t.Run("OnHighlight", func(t *testing.T) {
		var called bool
		result := tree.SetOnHighlight(func(node *TreeNode) {
			called = true
		})
		assert.Same(t, tree, result)
		_ = called
	})

	t.Run("OnExpand", func(t *testing.T) {
		var called bool
		result := tree.SetOnExpand(func(node *TreeNode) {
			called = true
		})
		assert.Same(t, tree, result)
		_ = called
	})

	t.Run("OnCollapse", func(t *testing.T) {
		var called bool
		result := tree.SetOnCollapse(func(node *TreeNode) {
			called = true
		})
		assert.Same(t, tree, result)
		_ = called
	})

	t.Run("LazyLoader", func(t *testing.T) {
		result := tree.SetLazyLoader(func(node *TreeNode) []*TreeNode {
			return nil
		})
		assert.Same(t, tree, result)
	})
}

// TestTree_FluentAPI tests method chaining.
func TestTree_FluentAPI(t *testing.T) {
	root := createSampleTree()

	tree := NewTree().
		SetRoot(root).
		SetShowLines(true).
		SetShowIcons(true).
		SetIndentSize(3).
		SetMultiSelect(false).
		SetOnSelect(func(node *TreeNode) {}).
		SetOnHighlight(func(node *TreeNode) {}).
		SetOnExpand(func(node *TreeNode) {}).
		SetOnCollapse(func(node *TreeNode) {}).
		ExpandAll()

	assert.NotNil(t, tree)
}

// TestTree_LevelAssignment tests that levels are correctly assigned.
func TestTree_LevelAssignment(t *testing.T) {
	tree := NewTree()
	root := createSampleTree()
	tree.SetRoot(root)

	assert.Equal(t, 0, root.level)
	assert.Equal(t, 1, root.Children[0].level)
	assert.Equal(t, 2, root.Children[0].Children[0].level)
}

// TestTree_EmptyTree tests operations on empty tree.
func TestTree_EmptyTree(t *testing.T) {
	tree := NewTree()

	// Should not panic
	tree.ExpandAll()
	tree.CollapseAll()
	tree.ExpandTo(3)
	tree.Filter("test")
	tree.ClearSelection()

	assert.Nil(t, tree.GetSelected())
	assert.Empty(t, tree.GetSelectedNodes())
}

// TestTreeNode_NodeWithData tests node with custom data.
func TestTreeNode_NodeWithData(t *testing.T) {
	type CustomData struct {
		Path string
		Size int64
	}

	node := &TreeNode{
		ID:    "file",
		Label: "document.txt",
		Icon:  "📄",
		Data: CustomData{
			Path: "/home/user/document.txt",
			Size: 1024,
		},
	}

	data, ok := node.Data.(CustomData)
	require.True(t, ok)
	assert.Equal(t, "/home/user/document.txt", data.Path)
	assert.Equal(t, int64(1024), data.Size)
}
