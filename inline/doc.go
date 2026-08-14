// Package inline provides renderer-native terminal components and a managed
// normal-screen renderer. It uses tcell for styled, grapheme-aware cell storage
// but owns its ANSI output lifecycle so it can preserve scrollback and coexist
// with ordinary command-line output.
package inline
