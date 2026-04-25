// Package input provides helpers for handling keyboard input in TUI applications.
//
// This package offers three main facilities:
//
// # KeyBindings Builder
//
// KeyBindings provides a fluent API for building input handlers:
//
//	bindings := input.NewKeyBindings().
//	    On(tcell.KeyEnter, func(e *tcell.EventKey) bool {
//	        submit()
//	        return true
//	    }).
//	    OnRune('q', func(e *tcell.EventKey) bool {
//	        quit()
//	        return true
//	    }).
//	    OnCtrlRune('s', func(e *tcell.EventKey) bool {
//	        save()
//	        return true
//	    })
//
//	component.SetInputHandler(bindings.Build())
//
// For components.ComponentBase (bool-return convention) use BuildBool:
//
//	componentBase.SetInputHandler(bindings.BuildBool())
//
// # Vim Navigation
//
// The package provides vim-style navigation helpers:
//
//	bindings := input.NewKeyBindings().AddVimNavigation(input.VimNavigator{
//	    Up:     list.MoveUp,
//	    Down:   list.MoveDown,
//	    Select: list.SelectCurrent,
//	    Back:   list.Close,
//	})
//
// Or for list-like components:
//
//	bindings := input.VimListBindings(myList)
//
// # Action Registry
//
// For complex applications, ActionRegistry provides named actions with key bindings:
//
//	actions := input.NewActionRegistry()
//	actions.Register(&input.Action{
//	    Name:        "save",
//	    Description: "Save changes",
//	    Key:         "Ctrl+S",
//	    Handler:     func() { save() },
//	})
//	actions.BindRune('s', "save")
//
//	// Get hints for menu display
//	hints := actions.KeyHints()
//
//	// Execute programmatically
//	actions.Execute("save")
package input
