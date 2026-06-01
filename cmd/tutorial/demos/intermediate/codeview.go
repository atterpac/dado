package intermediate

import (

	"github.com/atterpac/dado/cmd/tutorial/demos"
	"github.com/atterpac/dado/components"
	"github.com/atterpac/dado/core"
)

func init() {
	demos.Register(&CodeViewDemo{
		DemoBase: demos.DemoBase{
			DemoName:        "CodeView",
			DemoDescription: "Syntax-highlighted code display",
			DemoCategory:    demos.Intermediate,
			DemoCode:        codeViewCode,
		},
	})
}

// CodeViewDemo demonstrates the CodeView component.
type CodeViewDemo struct {
	demos.DemoBase
	codeview    *components.CodeView
	language    string
	showNumbers bool
}

// Component returns the demo component.
func (d *CodeViewDemo) Component() core.Widget {
	d.language = "go"
	d.showNumbers = true

	sampleCode := `package main

import (
    "fmt"
    "strings"
)

// Greet returns a greeting message
func Greet(name string) string {
    if name == "" {
        return "Hello, World!"
    }
    return fmt.Sprintf("Hello, %s!", strings.Title(name))
}

func main() {
    message := Greet("dado")
    fmt.Println(message)  // Hello, Dado!
}
`

	d.codeview = components.NewCodeView().
		SetCode(sampleCode).
		SetLanguage(components.LangGo).
		SetShowLineNumbers(true).
		SetTabWidth(4).
		SetHighlightLine(10)

	d.Props = []demos.PropertyDescriptor{
		demos.SelectProp("language", "Syntax highlighting language",
			[]string{"go", "python", "javascript", "json", "bash"},
			func() string { return d.language },
			func(v string) {
				d.language = v
				switch v {
				case "go":
					d.codeview.SetLanguage(components.LangGo)
				case "python":
					d.codeview.SetLanguage(components.LangPython)
				case "javascript":
					d.codeview.SetLanguage(components.LangJavaScript)
				case "json":
					d.codeview.SetLanguage(components.LangJSON)
				case "bash":
					d.codeview.SetLanguage(components.LangBash)
				}
			},
			"go",
		),
		demos.BoolProp("showNumbers", "Show line numbers",
			func() bool { return d.showNumbers },
			func(v bool) { d.showNumbers = v; d.codeview.SetShowLineNumbers(v) },
			true,
		),
	}

	return d.codeview
}

const codeViewCode = `package main


// Create a code viewer
code := components.NewCodeView().
    SetCode(sourceCode).
    SetLanguage(components.LangGo)

// Available languages:
// - LangGo
// - LangPython
// - LangJavaScript
// - LangTypeScript
// - LangJSON
// - LangBash
// - LangSQL
// - LangYAML

// Display options
code.SetShowLineNumbers(true)
code.SetTabWidth(4)
code.SetWrapLines(false)

// Highlight a specific line
code.SetHighlightLine(10)  // 1-based

// Scroll to line
code.ScrollTo(50)

// Navigation (vim-style):
// j/k - scroll up/down
// g/G - top/bottom
// h/l - scroll left/right
`
