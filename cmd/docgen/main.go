// docgen extracts Go doc comments from the components package and writes one
// JSON file per component to the output directory.
//
// Usage:
//
//	go run ./cmd/docgen -out docs-site/src/content/api
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Categories for non-component types so Astro can filter/route them separately.
// Types not listed here are treated as "component".
var typeCategories = map[string]string{
	// Events — documented on the Events concept page, not as components.
	"ActivateEvent": "event",
	"BaseEvent":     "event",
	"CancelEvent":   "event",
	"ChangeEvent":   "event",
	"FocusEvent":    "event",
	"KeyEvent":      "event",
	"SelectEvent":   "event",
	"SubmitEvent":   "event",

	// Framework base types — documented on the Lifecycle concept page.
	"ComponentBase":        "base",
	"StatefulComponentBase": "base",
	"Layout":               "base",

	// Data helpers — documented alongside their parent component.
	"Changeset":     "data",
	"GitGraphData":  "data",
	"GraphTreeData": "data",
	"NodeGraphData": "data",
	"SliceSource":   "data",
}

// ComponentDoc is the JSON shape written for each component.
type ComponentDoc struct {
	Name        string      `json:"name"`
	Slug        string      `json:"slug"`
	Category    string      `json:"category"` // "component" | "event" | "base" | "data"
	Doc         string      `json:"doc"`
	Constructor *MethodDoc   `json:"constructor,omitempty"`
	Methods     []MethodDoc  `json:"methods"`
	Types       []TypeDoc    `json:"types,omitempty"`
	Examples    []ExampleDoc `json:"examples,omitempty"`
}

// ExampleDoc is a runnable Go example extracted from a _test.go file.
// Name is the optional label suffix (the part after "_" in the function name);
// it is empty for the canonical Example<Type> kitchen-sink example.
type ExampleDoc struct {
	Name string `json:"name"`
	Doc  string `json:"doc"`
	Code string `json:"code"`
}

type MethodDoc struct {
	Name      string `json:"name"`
	Signature string `json:"signature"`
	Doc       string `json:"doc"`
}

type TypeDoc struct {
	Name   string      `json:"name"`
	Doc    string      `json:"doc"`
	Kind   string      `json:"kind,omitempty"` // "struct" | "enum" | ""
	Fields []FieldDoc  `json:"fields,omitempty"`
	Values []ValueDoc  `json:"values,omitempty"`
}

type FieldDoc struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Doc  string `json:"doc"`
}

type ValueDoc struct {
	Name string `json:"name"`
	Doc  string `json:"doc"`
}

// tview plumbing methods that are not part of the dado API.
var excludedMethods = map[string]bool{
	"Draw":           true,
	"InputHandler":   true,
	"Focus":          true,
	"HasFocus":       true,
	"Blur":           true,
	"GetRect":        true,
	"SetRect":        true,
	"GetInnerRect":   true,
	"MouseHandler":   true,
	"PasteHandler":   true,
	"WrapInputHandler": true,
	"WrapMouseHandler": true,
	"Subs":           true,
	"SetTheme":       true,
}

func main() {
	pkgDir  := flag.String("pkg", "components", "path to components package")
	outDir  := flag.String("out", "docs-site/src/content/api", "output directory for JSON files")
	docsDir := flag.String("docs", "docs-site/src/content/docs/components", "output directory for MDX stubs")
	flag.Parse()

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, *pkgDir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse: %v\n", err)
		os.Exit(1)
	}

	pkg, ok := pkgs["components"]
	if !ok {
		fmt.Fprintln(os.Stderr, "package 'components' not found")
		os.Exit(1)
	}

	var files []*ast.File
	for _, f := range pkg.Files {
		files = append(files, f)
	}

	dpkg, err := doc.NewFromFiles(fset, files, "github.com/atterpac/dado/components")
	if err != nil {
		fmt.Fprintf(os.Stderr, "doc: %v\n", err)
		os.Exit(1)
	}

	// Build a map of constructor name → type name for quick lookup.
	constructorOf := map[string]string{} // "NewBadge" → "Badge"
	for _, t := range dpkg.Types {
		for _, f := range t.Funcs {
			if strings.HasPrefix(f.Name, "New") {
				constructorOf[f.Name] = t.Name
			}
		}
	}

	// Extract Go examples from _test.go files and group them by the component
	// they document. Naming convention follows go/doc: ExampleSelect documents
	// the Select type; ExampleNewSelect documents its constructor (mapped back
	// to Select); a trailing "_label" becomes the example's display label.
	// Examples are compiled by `go test`, so they cannot drift from the API.
	examplesByType := map[string][]ExampleDoc{}
	exFset := token.NewFileSet()
	exPkgs, err := parser.ParseDir(exFset, *pkgDir, func(fi os.FileInfo) bool {
		return strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse examples: %v\n", err)
		os.Exit(1)
	}
	var exFiles []*ast.File
	for _, p := range exPkgs {
		for _, f := range p.Files {
			exFiles = append(exFiles, f)
		}
	}
	// rank orders a component's examples: the type-named kitchen sink
	// (Example<Type>) first, then constructor-named, then labeled variants.
	type ranked struct {
		rank int
		ex   ExampleDoc
	}
	rankedByType := map[string][]ranked{}
	for _, ex := range doc.Examples(exFiles...) {
		key, label, _ := strings.Cut(ex.Name, "_")
		typeName := key
		rank := 0 // Example<Type> — the canonical kitchen-sink example
		if ctorType, ok := constructorOf[key]; ok {
			typeName = ctorType
			rank = 1 // ExampleNew<Type> — constructor-focused
		}
		if label != "" {
			rank += 2 // labeled variant — supplementary
		}
		rankedByType[typeName] = append(rankedByType[typeName], ranked{rank, ExampleDoc{
			Name: label,
			Doc:  strings.TrimSpace(ex.Doc),
			Code: exampleCode(exFset, ex),
		}})
	}
	for typeName, rs := range rankedByType {
		sort.SliceStable(rs, func(i, j int) bool { return rs[i].rank < rs[j].rank })
		exs := make([]ExampleDoc, len(rs))
		for i, r := range rs {
			exs[i] = r.ex
		}
		examplesByType[typeName] = exs
	}

	// Build a map of type name → exported struct fields.
	structFields := map[string][]FieldDoc{}
	for _, file := range files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok || st.Fields == nil {
					continue
				}
				var fields []FieldDoc
				for _, f := range st.Fields.List {
					if len(f.Names) == 0 {
						continue // embedded
					}
					name := f.Names[0].Name
					if name == "" || name[0] < 'A' || name[0] > 'Z' {
						continue // unexported
					}
					var typeBuf bytes.Buffer
					printer.Fprint(&typeBuf, fset, f.Type)
					doc := ""
					if f.Comment != nil {
						doc = strings.TrimSpace(f.Comment.Text())
					} else if f.Doc != nil {
						doc = strings.TrimSpace(f.Doc.Text())
					}
					fields = append(fields, FieldDoc{
						Name: name,
						Type: typeBuf.String(),
						Doc:  doc,
					})
				}
				if len(fields) > 0 {
					structFields[ts.Name.Name] = fields
				}
			}
		}
	}

	// Build a map of type name → associated enum values.
	enumValues := map[string][]ValueDoc{} // "BadgeVariant" → [...]
	for _, t := range dpkg.Types {
		if len(t.Consts) > 0 {
			var vals []ValueDoc
			for _, c := range t.Consts {
				for _, spec := range c.Decl.Specs {
					_ = spec
				}
				// Use the grouped doc on the block; individual names from c.Names.
				for _, name := range c.Names {
					vals = append(vals, ValueDoc{
						Name: name,
						Doc:  strings.TrimSpace(c.Doc),
					})
				}
			}
			if len(vals) > 0 {
				enumValues[t.Name] = vals
			}
		}
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(*docsDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}

	written := 0
	for _, t := range dpkg.Types {
		// Only emit types that have a New* constructor — these are components.
		var ctor *MethodDoc
		for _, f := range t.Funcs {
			if strings.HasPrefix(f.Name, "New") && strings.HasSuffix(f.Name, t.Name) {
				ctor = &MethodDoc{
					Name:      f.Name,
					Signature: funcSignature(fset, f.Decl),
					Doc:       strings.TrimSpace(f.Doc),
				}
				break
			}
		}
		if ctor == nil {
			continue
		}

		// Collect public API methods, skipping tview plumbing.
		var methods []MethodDoc
		for _, m := range t.Methods {
			if excludedMethods[m.Name] {
				continue
			}
			if strings.HasPrefix(m.Name, "wrap") || strings.HasPrefix(m.Name, "emit") {
				continue
			}
			methods = append(methods, MethodDoc{
				Name:      m.Name,
				Signature: funcSignature(fset, m.Decl),
				Doc:       strings.TrimSpace(m.Doc),
			})
		}

		// Build a lookup of all package types for transitive resolution.
		pkgTypeByName := map[string]*doc.Type{}
		for _, at := range dpkg.Types {
			pkgTypeByName[at.Name] = at
		}

		// collectType builds a TypeDoc and returns it.
		buildTypeDoc := func(name string) (TypeDoc, bool) {
			at, ok := pkgTypeByName[name]
			if !ok {
				return TypeDoc{}, false
			}
			vals := enumValues[name]
			fields := structFields[name]
			kind := ""
			if len(fields) > 0 {
				kind = "struct"
			} else if len(vals) > 0 {
				kind = "enum"
			}
			return TypeDoc{
				Name:   name,
				Doc:    strings.TrimSpace(at.Doc),
				Kind:   kind,
				Fields: fields,
				Values: vals,
			}, true
		}

		// Seed with types directly mentioned in method/constructor signatures.
		seen := map[string]bool{t.Name: true}
		queue := []string{}
		for _, at := range dpkg.Types {
			if at.Name == t.Name {
				continue
			}
			typeWord := regexp.MustCompile(`\b` + regexp.QuoteMeta(at.Name) + `\b`)
			mentioned := false
			for _, m := range methods {
				if typeWord.MatchString(m.Signature) {
					mentioned = true
					break
				}
			}
			if ctor != nil && typeWord.MatchString(ctor.Signature) {
				mentioned = true
			}
			if mentioned && !seen[at.Name] {
				seen[at.Name] = true
				queue = append(queue, at.Name)
			}
		}

		// Transitive closure: expand through struct field types.
		var assocTypes []TypeDoc
		for len(queue) > 0 {
			name := queue[0]
			queue = queue[1:]
			td, ok := buildTypeDoc(name)
			if !ok {
				continue
			}
			assocTypes = append(assocTypes, td)
			// Walk field types of structs to find further referenced package types.
			for _, f := range td.Fields {
				// Strip pointer/slice/map wrappers to get the bare type name.
				bare := strings.TrimLeft(f.Type, "[]*")
				bare = strings.SplitN(bare, ".", 2)[len(strings.SplitN(bare, ".", 2))-1]
				if _, exists := pkgTypeByName[bare]; exists && !seen[bare] {
					seen[bare] = true
					queue = append(queue, bare)
				}
			}
			// Also walk enum values — their names may reference other types indirectly,
			// but more usefully: check if any package type name appears in field type strings.
			for _, at := range dpkg.Types {
				if seen[at.Name] {
					continue
				}
				typeWord := regexp.MustCompile(`\b` + regexp.QuoteMeta(at.Name) + `\b`)
				for _, f := range td.Fields {
					if typeWord.MatchString(f.Type) {
						seen[at.Name] = true
						queue = append(queue, at.Name)
						break
					}
				}
			}
		}

		category := typeCategories[t.Name]
		if category == "" {
			category = "component"
		}

		comp := ComponentDoc{
			Name:        t.Name,
			Slug:        toSlug(t.Name),
			Category:    category,
			Doc:         strings.TrimSpace(t.Doc),
			Constructor: ctor,
			Methods:     methods,
			Types:       assocTypes,
			Examples:    examplesByType[t.Name],
		}

		out, err := json.MarshalIndent(comp, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal %s: %v\n", t.Name, err)
			continue
		}

		path := filepath.Join(*outDir, comp.Slug+".json")
		if err := os.WriteFile(path, out, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", path, err)
			continue
		}
		fmt.Printf("wrote %s\n", path)

		// Write MDX stub only for component-category types (not events/base/data).
		if comp.Category == "component" {
			mdx := buildMDX(comp)
			mdxPath := filepath.Join(*docsDir, comp.Slug+".mdx")
			// Only write if the file does not already exist — preserves hand-edited content.
			if _, err := os.Stat(mdxPath); os.IsNotExist(err) {
				if err := os.WriteFile(mdxPath, []byte(mdx), 0o644); err != nil {
					fmt.Fprintf(os.Stderr, "write MDX %s: %v\n", mdxPath, err)
				} else {
					fmt.Printf("wrote %s\n", mdxPath)
				}
			}
		}

		written++
	}

	fmt.Printf("\n%d component JSON files written to %s\n", written, *outDir)
}

// buildMDX returns a Starlight MDX stub for a component.
// Only written if the file does not already exist so hand-edits are preserved.
func buildMDX(c ComponentDoc) string {
	return fmt.Sprintf(`---
title: %s
description: %s
---

import ComponentPage from '../../../components/ComponentPage.astro';
import data from '../../api/%s.json';
import { getAvailableThemes } from '../../../lib/themes';

export const themes = await getAvailableThemes();

<ComponentPage data={data} themes={themes} defaultTheme="nord" />
`, c.Name, firstSentence(c.Doc), c.Slug)
}

// firstSentence returns the text up to and including the first sentence-ending
// punctuation, or the full string if none is found. Newlines are stripped.
func firstSentence(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	for i, r := range s {
		if r == '.' || r == '!' || r == '?' {
			return strings.TrimSpace(s[:i+1])
		}
	}
	return strings.TrimSpace(s)
}

// exampleCode renders an example's body to source text, stripping the outer
// braces and dedenting one level so it reads as a standalone snippet.
func exampleCode(fset *token.FileSet, ex *doc.Example) string {
	if ex.Code == nil {
		return ""
	}
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, ex.Code)
	code := strings.TrimSpace(buf.String())
	if strings.HasPrefix(code, "{") && strings.HasSuffix(code, "}") {
		inner := code[1 : len(code)-1]
		lines := strings.Split(inner, "\n")
		for i, l := range lines {
			lines[i] = strings.TrimPrefix(l, "\t")
		}
		code = strings.Trim(strings.Join(lines, "\n"), "\n")
	}
	return code
}

// funcSignature prints just the signature (no body) of a func declaration.
func funcSignature(fset *token.FileSet, decl *ast.FuncDecl) string {
	// Temporarily nil the body so printer only emits the signature line.
	body := decl.Body
	decl.Body = nil
	defer func() { decl.Body = body }()

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, decl)
	return strings.TrimSpace(buf.String())
}

// toSlug converts "AutocompleteInput" → "autocomplete-input".
func toSlug(name string) string {
	var b strings.Builder
	for i, r := range name {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r + 32)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
