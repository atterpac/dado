---
label: Tree
icon: list-unordered
order: 60
---

# Tree

Hierarchical tree view with expand/collapse support.

---

## Basic Usage

```go
import "github.com/atterpac/jig/components"

tree := components.NewTree().
    SetRoot(components.NewTreeNode("Root").
        AddChild(components.NewTreeNode("Child 1").
            AddChild(components.NewTreeNode("Grandchild 1")).
            AddChild(components.NewTreeNode("Grandchild 2"))).
        AddChild(components.NewTreeNode("Child 2")))
```

---

## TreeNode

```go
node := components.NewTreeNode("My Node").
    SetReference(myData).      // Attach custom data
    SetExpanded(true).         // Start expanded
    SetSelectable(true).       // Allow selection
    AddChild(childNode)
```

### Methods

| Method | Description |
|--------|-------------|
| `SetText(string)` | Set node text |
| `SetReference(any)` | Attach custom data |
| `SetExpanded(bool)` | Set expanded state |
| `SetSelectable(bool)` | Allow selection |
| `AddChild(*TreeNode)` | Add child node |
| `RemoveChild(*TreeNode)` | Remove child node |
| `GetChildren()` | Get child nodes |
| `GetReference()` | Get attached data |

---

## Events

```go
tree := components.NewTree().
    SetOnSelect(func(node *components.TreeNode) {
        log.Printf("Selected: %s", node.Text())
        if data := node.GetReference(); data != nil {
            // Handle attached data
        }
    }).
    SetOnExpand(func(node *components.TreeNode) {
        log.Printf("Expanded: %s", node.Text())
    }).
    SetOnCollapse(func(node *components.TreeNode) {
        log.Printf("Collapsed: %s", node.Text())
    })
```

---

## Navigation

```go
// Expand/collapse all
tree.ExpandAll()
tree.CollapseAll()

// Get current selection
selected := tree.GetCurrentNode()
```

---

## Example: File Browser

```go
type FileBrowser struct {
    *components.ComponentBase
    tree *components.Tree
    app  *layout.App
}

func NewFileBrowser(app *layout.App, rootPath string) *FileBrowser {
    tree := components.NewTree()

    // Build tree from filesystem
    root := buildFileTree(rootPath)
    tree.SetRoot(root)

    tree.SetOnSelect(func(node *components.TreeNode) {
        if path, ok := node.GetReference().(string); ok {
            info, _ := os.Stat(path)
            if !info.IsDir() {
                app.Pages().Push(NewFileViewer(app, path))
            }
        }
    })

    v := &FileBrowser{tree: tree, app: app}
    v.ComponentBase = components.NewComponentBase(tree).
        SetName("files").
        AddHint("Enter", "Open").
        AddHint("Space", "Expand/Collapse").
        AddHint("Esc", "Back")

    return v
}

func buildFileTree(path string) *components.TreeNode {
    info, _ := os.Stat(path)
    node := components.NewTreeNode(filepath.Base(path)).
        SetReference(path)

    if info.IsDir() {
        node.SetExpanded(false)
        entries, _ := os.ReadDir(path)
        for _, entry := range entries {
            child := buildFileTree(filepath.Join(path, entry.Name()))
            node.AddChild(child)
        }
    }

    return node
}
```

---

## Example: Lazy Loading

```go
func NewLazyTree() *components.Tree {
    tree := components.NewTree()

    // Only load children when expanded
    tree.SetOnExpand(func(node *components.TreeNode) {
        if node.GetChildren() == nil {
            // Load children on demand
            go func() {
                children := loadChildren(node.GetReference())
                app.QueueUpdateDraw(func() {
                    for _, c := range children {
                        child := components.NewTreeNode(c.Name).
                            SetReference(c)
                        node.AddChild(child)
                    }
                })
            }()
        }
    })

    return tree
}
```

---

## Keyboard Navigation

| Key | Action |
|-----|--------|
| `j` / `Down` | Move down |
| `k` / `Up` | Move up |
| `h` / `Left` | Collapse / Go to parent |
| `l` / `Right` | Expand / Enter |
| `Enter` | Select node |
| `Space` | Toggle expand/collapse |
| `g` | Go to root |
| `G` | Go to last visible |
