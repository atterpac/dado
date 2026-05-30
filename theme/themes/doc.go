// Package themes provides the built-in color themes.
//
// Themes are authored as YAML in defs/*.yaml (the source of truth) and compiled
// into the type-safe registry in themes_gen.go. After editing a def, regenerate:
//
//	go generate ./...
//
//go:generate go run ./gen
package themes
