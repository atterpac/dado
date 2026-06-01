---
title: Installation
description: Add dado to a Go module.
---

## Requirements

- Go 1.21 or later
- A terminal emulator that supports 256-colour or truecolour (most modern terminals do)

## Add the module

```sh
go get github.com/atterpac/dado@latest
```

That's it. dado pulls in tcell as its only runtime dependency.

## Verify

```go
package main

import "github.com/atterpac/dado"

func main() {
    app := dado.New()
    _ = app
}
```

```sh
go run .
```

No output means everything compiled correctly. Hit `Ctrl+C` to exit the blank
screen.

## Optional: vendor dependencies

```sh
go mod vendor
```

dado has no CGo dependencies, so cross-compilation and vendoring both work without
additional setup.

## Editor setup

dado ships with Go doc comments on every exported symbol. Standard
`gopls`-powered editors (VS Code, GoLand, Neovim) will show inline docs and
completions without any extra configuration.
