# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build ./cmd/tide        # build binary
go run ./cmd/tide          # run from source (empty buffer)
go run ./cmd/tide file.go  # open files in tabs
go test ./...              # run all tests
go vet ./...               # lint
```

## Architecture

Tide is a terminal IDE ("VSCode in the terminal") built from scratch in Go with minimal dependencies. The only external dep is `golang.org/x/term` for raw mode.

### Package Dependency Flow

```
cmd/tide/main.go  (app shell: event loop, layout, theme, keybindings)
  ├── internal/term      (custom terminal library — no tcell)
  ├── internal/buffer    (piece table text buffer)
  ├── internal/editor    (cursor, viewport, key handling, syntax rendering)
  ├── internal/highlight (regex tokenizer + 15 language defs)
  ├── internal/command   (command registry with fuzzy search)
  ├── internal/git       (shells out to git CLI for status/diff/stage)
  └── internal/ui        (sidebar, command palette, git panel)
```

### Key Design Decisions

- **`internal/term`**: Custom terminal library using raw ANSI escape sequences. Double-buffered screen (front/back grids) with dirty-cell diffing — only changed cells are written on `Show()`. A single persistent goroutine reads stdin into a buffered channel; `PollEvent` selects on that channel and SIGWINCH signals. Supports SGR mouse, F-keys, Ctrl keys, and UTF-8.

- **`internal/buffer`**: Piece table (same data structure as VSCode). Original text is immutable; edits append to an add buffer. Pieces are `(source, offset, length)` spans. This means inserts never copy the whole file.

- **`internal/highlight`**: Regex rules per language with a `TrimRight` field on rules to emulate lookaheads (Go's regexp/RE2 has no lookahead support). Block comment state is tracked line-by-line. The highlighter walks from line 0 to the scroll position to establish correct block comment state before rendering visible lines.

- **`internal/git`**: Wraps the `git` CLI (no libgit2/go-git). Parses `--porcelain` output. Diff viewing supports staged, unstaged, and untracked files.

- **Theme/styling**: All styles are `term.Style` structs with `term.Color` (RGB). The app's theme is defined once in `cmd/tide/main.go` (`newTheme()`) and passed into components as theme structs — components never hardcode colors.

### Event Flow

`App.run()` is a synchronous loop: `draw() → Show() → PollEvent() → handleKey/handleMouse`. Focus state (`focusEditor`, `focusSidebar`, `focusGit`) routes key events to the correct panel. Global F-key bindings are checked before focus-based routing.

### Adding a New Command

Register in `App.registerCommands()` with a `command.Command{Name, Description, Shortcut, Execute}`. It automatically appears in the F1 command palette. Map an F-key in `handleKey` if it needs a global shortcut.

### Adding a New Language

Add a `langFoo()` function in `internal/highlight/languages.go` returning a `*Language` with extensions, comment syntax, and regex `Rule` slices. Register it in `init()`. Use `funcCall()` helper for function-name highlighting (matches `word(` with `TrimRight: 1`).
