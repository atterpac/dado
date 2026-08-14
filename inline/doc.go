// Package inline provides normal-screen terminal output and a managed inline
// renderer. The renderer uses tcell for styled, grapheme-aware cell storage but
// owns its ANSI output lifecycle so it can preserve scrollback and coexist with
// ordinary command-line output.
//
// The package-level Print helpers remain available for one-shot CLI output.
package inline
