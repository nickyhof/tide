# Tide

A lightweight terminal IDE built from scratch in Go — VSCode in the terminal.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-Apache%202.0-blue)

## Features

- **Tabbed editor** with piece table buffer (same data structure as VSCode)
- **Syntax highlighting** for 15+ languages (Go, Python, JS/TS, Rust, C/C++, Java, Ruby, SQL, HTML, CSS, YAML, JSON, Markdown, Shell)
- **File explorer** with tree navigation and lazy directory loading
- **Command palette** (F1) with fuzzy search
- **Git integration** — branch indicator, staged/modified/untracked counts, inline diff viewer, stage/unstage from the editor
- **Mouse support** — click to place cursor, select tabs, navigate panels
- **Custom terminal library** — no tcell or external TUI frameworks, just ANSI escape codes and `golang.org/x/term`

## Install

```bash
go install github.com/nickyhof/tide/cmd/tide@latest
```

Or build from source:

```bash
git clone https://github.com/nickyhof/tide.git
cd tide
go build -o tide ./cmd/tide
```

## Usage

```bash
tide                        # empty buffer
tide file.go                # open a file
tide main.go go.mod README  # open multiple files in tabs
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| **F1** | Command palette |
| **F2** | Save file |
| **F3** | New file |
| **F4** | Close tab |
| **F5** | Toggle file explorer |
| **F6** | Toggle git panel |
| **F7 / F8** | Previous / next tab |
| **F10** | Quit |
| **Ctrl+S** | Save file |
| **Esc** | Return focus to editor |

### Git panel (when focused)

| Key | Action |
|-----|--------|
| **s** | Stage / unstage file |
| **r** | Refresh status |
| **Enter** | Open file in editor |
| **PgUp / PgDn** | Scroll diff |

## Architecture

```
cmd/tide/           Entry point, event loop, layout, theme
internal/term/      Custom terminal library (screen buffer, input parser, ANSI styles)
internal/buffer/    Piece table text buffer
internal/editor/    Editor core (cursor, viewport, key handling, syntax rendering)
internal/highlight/ Regex-based syntax highlighting engine + language definitions
internal/command/   Command registry with fuzzy search
internal/git/       Git CLI wrapper (status, diff, stage/unstage)
internal/ui/        UI components (sidebar, command palette, git panel)
```

Only two external dependencies: `golang.org/x/term` and `golang.org/x/sys`.

## License

[Apache 2.0](LICENSE)
