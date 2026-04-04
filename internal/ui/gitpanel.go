package ui

import (
	"fmt"
	"strings"

	"github.com/nickyhof/tide/internal/git"
	"github.com/nickyhof/tide/internal/term"
)

// GitPanel shows the git stage explorer and diff viewer.
type GitPanel struct {
	Open       bool
	Repo       *git.Repo
	Files      []git.FileStatus
	Cursor     int
	Scroll     int
	DiffLines  []DiffLine
	DiffScroll int
	splitRatio float64
	OnOpenFile func(path string)
}

// DiffLine is a single line of a diff with its type.
type DiffLine struct {
	Text string
	Kind DiffLineKind
}

// DiffLineKind classifies diff lines.
type DiffLineKind int

const (
	DiffContext DiffLineKind = iota
	DiffAdd
	DiffRemove
	DiffHeader
	DiffHunk
)

// GitPanelTheme holds styles for the git panel.
type GitPanelTheme struct {
	Normal    term.Style
	Bold      term.Style
	Staged    term.Style
	Modified  term.Style
	Untracked term.Style
	Selected  term.Style
	Hint      term.Style
	Separator term.Style
	DiffAdd   term.Style
	DiffRm    term.Style
	DiffHunk  term.Style
	DiffHead  term.Style
	DiffCtx   term.Style
	TabBar    term.Style
}

// NewGitPanel creates a git panel.
func NewGitPanel(repo *git.Repo) *GitPanel {
	return &GitPanel{
		Repo:       repo,
		splitRatio: 0.4,
	}
}

// Refresh reloads git status.
func (g *GitPanel) Refresh() {
	if g.Repo == nil {
		return
	}
	g.Files = g.Repo.Status()
	if g.Cursor >= len(g.Files) {
		g.Cursor = len(g.Files) - 1
	}
	if g.Cursor < 0 {
		g.Cursor = 0
	}
	g.loadDiff()
}

// Toggle opens/closes the panel.
func (g *GitPanel) Toggle() {
	g.Open = !g.Open
	if g.Open {
		g.Refresh()
	}
}

// HandleKey processes input.
func (g *GitPanel) HandleKey(ev term.KeyEvent) bool {
	if !g.Open {
		return false
	}
	switch ev.Key {
	case term.KeyUp:
		if g.Cursor > 0 {
			g.Cursor--
			g.loadDiff()
		}
	case term.KeyDown:
		if g.Cursor < len(g.Files)-1 {
			g.Cursor++
			g.loadDiff()
		}
	case term.KeyEnter:
		if g.Cursor < len(g.Files) && g.OnOpenFile != nil {
			g.OnOpenFile(g.Files[g.Cursor].AbsPath)
		}
	case term.KeyPgDn:
		g.DiffScroll += 10
	case term.KeyPgUp:
		g.DiffScroll -= 10
		if g.DiffScroll < 0 {
			g.DiffScroll = 0
		}
	case term.KeyRune:
		switch ev.Rune {
		case 's', 'S':
			g.stageToggle()
		case 'r', 'R':
			g.Refresh()
		case 'j':
			if g.Cursor < len(g.Files)-1 {
				g.Cursor++
				g.loadDiff()
			}
		case 'k':
			if g.Cursor > 0 {
				g.Cursor--
				g.loadDiff()
			}
		default:
			return false
		}
	default:
		return false
	}
	return true
}

func (g *GitPanel) stageToggle() {
	if g.Cursor >= len(g.Files) || g.Repo == nil {
		return
	}
	f := g.Files[g.Cursor]
	if f.IsStaged() {
		g.Repo.Unstage(f.Path)
	} else {
		g.Repo.Stage(f.Path)
	}
	g.Refresh()
}

func (g *GitPanel) loadDiff() {
	g.DiffLines = nil
	g.DiffScroll = 0
	if g.Cursor >= len(g.Files) || g.Repo == nil {
		return
	}
	f := g.Files[g.Cursor]
	var raw string
	if f.IsUntracked() {
		raw = g.Repo.DiffUntracked(f.AbsPath)
	} else if f.IsStaged() {
		raw = g.Repo.DiffFile(f.Path, true)
	} else {
		raw = g.Repo.DiffFile(f.Path, false)
	}
	if raw == "" {
		return
	}
	for _, line := range strings.Split(raw, "\n") {
		g.DiffLines = append(g.DiffLines, parseDiffLine(line))
	}
}

func parseDiffLine(line string) DiffLine {
	if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") ||
		strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "index ") ||
		strings.HasPrefix(line, "new file") || strings.HasPrefix(line, "deleted file") {
		return DiffLine{Text: line, Kind: DiffHeader}
	}
	if strings.HasPrefix(line, "@@") {
		return DiffLine{Text: line, Kind: DiffHunk}
	}
	if strings.HasPrefix(line, "+") {
		return DiffLine{Text: line, Kind: DiffAdd}
	}
	if strings.HasPrefix(line, "-") {
		return DiffLine{Text: line, Kind: DiffRemove}
	}
	return DiffLine{Text: line, Kind: DiffContext}
}

// Draw renders the git panel.
func (g *GitPanel) Draw(scr *term.Screen, x, y, w, h int, theme GitPanelTheme) {
	if !g.Open || w < 10 || h < 5 {
		return
	}

	scr.FillRect(x, y, w, h, theme.Normal, ' ')

	// Separator
	for row := y; row < y+h; row++ {
		scr.SetContent(x, row, '│', theme.Separator)
	}

	innerX := x + 1
	innerW := w - 1

	listH := int(float64(h) * g.splitRatio)
	if listH < 3 {
		listH = 3
	}
	diffY := y + listH
	diffH := h - listH

	// Header
	branch := ""
	if g.Repo != nil {
		branch = g.Repo.Branch()
	}
	headerText := " GIT"
	if branch != "" {
		headerText += "  " + branch
	}
	scr.FillRect(innerX, y, innerW, 1, theme.TabBar, ' ')
	scr.DrawText(innerX, y, theme.Bold, headerText, innerW)

	// File list
	fileAreaY := y + 1
	fileAreaH := listH - 1

	if g.Cursor < g.Scroll {
		g.Scroll = g.Cursor
	}
	if g.Cursor >= g.Scroll+fileAreaH {
		g.Scroll = g.Cursor - fileAreaH + 1
	}

	for row := 0; row < fileAreaH; row++ {
		idx := g.Scroll + row
		if idx >= len(g.Files) {
			break
		}
		f := g.Files[idx]
		screenY := fileAreaY + row

		var style term.Style
		switch {
		case f.IsStaged():
			style = theme.Staged
		case f.IsModified():
			style = theme.Modified
		case f.IsUntracked():
			style = theme.Untracked
		default:
			style = theme.Normal
		}

		if idx == g.Cursor {
			// Merge fg with selection bg
			style = theme.Selected.WithFg(style.Fg)
		}

		scr.FillRect(innerX, screenY, innerW, 1, style, ' ')

		icon := statusIcon(f)
		scr.DrawText(innerX+1, screenY, style, icon, 2)
		scr.DrawText(innerX+3, screenY, style, f.Path, innerW-4)
	}

	// Divider with hint
	scr.FillRect(innerX, diffY, innerW, 1, theme.Hint, '─')
	hint := " s:stage/unstage  r:refresh  Enter:open  PgUp/Dn:scroll"
	scr.DrawText(innerX, diffY, theme.Hint, hint, innerW)

	// Diff viewer
	diffAreaY := diffY + 1
	diffAreaH := diffH - 1
	if diffAreaH < 1 {
		return
	}

	maxScroll := len(g.DiffLines) - diffAreaH
	if maxScroll < 0 {
		maxScroll = 0
	}
	if g.DiffScroll > maxScroll {
		g.DiffScroll = maxScroll
	}

	for row := 0; row < diffAreaH; row++ {
		idx := g.DiffScroll + row
		if idx >= len(g.DiffLines) {
			break
		}
		dl := g.DiffLines[idx]
		screenY := diffAreaY + row

		var style term.Style
		switch dl.Kind {
		case DiffAdd:
			style = theme.DiffAdd
		case DiffRemove:
			style = theme.DiffRm
		case DiffHunk:
			style = theme.DiffHunk
		case DiffHeader:
			style = theme.DiffHead
		default:
			style = theme.DiffCtx
		}

		scr.FillRect(innerX, screenY, innerW, 1, style, ' ')
		text := dl.Text
		if len(text) > innerW {
			text = text[:innerW]
		}
		scr.DrawText(innerX, screenY, style, text, innerW)
	}
}

func statusIcon(f git.FileStatus) string {
	switch {
	case f.IsStaged():
		return "● "
	case f.IsModified():
		return "○ "
	case f.IsUntracked():
		return "? "
	default:
		return "  "
	}
}

// StatusSummary returns a string like "+2 ~3 ?1" for the status bar.
func StatusSummary(repo *git.Repo) string {
	if repo == nil {
		return ""
	}
	branch := repo.Branch()
	if branch == "" {
		return ""
	}
	status := repo.Status()
	staged, modified, untracked := 0, 0, 0
	for _, f := range status {
		if f.IsStaged() {
			staged++
		}
		if f.IsModified() {
			modified++
		}
		if f.IsUntracked() {
			untracked++
		}
	}
	info := branch
	if staged > 0 || modified > 0 || untracked > 0 {
		info += fmt.Sprintf("  +%d ~%d ?%d", staged, modified, untracked)
	}
	return info
}
