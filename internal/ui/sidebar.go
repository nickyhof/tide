package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nickyhof/tide/internal/term"
)

// FileEntry represents a file or directory in the explorer.
type FileEntry struct {
	Name     string
	Path     string
	IsDir    bool
	Depth    int
	Expanded bool
	Children []*FileEntry
}

// SidebarTheme holds the styles for the sidebar.
type SidebarTheme struct {
	Normal    term.Style
	Bold      term.Style
	Selection term.Style
}

// Sidebar is a file explorer panel.
type Sidebar struct {
	RootPath string
	Root     *FileEntry
	Flat     []*FileEntry
	Cursor   int
	Scroll   int
	Width    int
	Open     bool
	OnFileSelect func(path string)
}

// NewSidebar creates a sidebar rooted at the given directory.
func NewSidebar(root string, width int) *Sidebar {
	s := &Sidebar{
		RootPath: root,
		Width:    width,
		Open:     true,
	}
	s.Reload()
	return s
}

// Reload rescans the root directory.
func (s *Sidebar) Reload() {
	s.Root = scanDir(s.RootPath, 0, 1)
	if s.Root != nil {
		s.Root.Expanded = true
	}
	s.rebuildFlat()
}

// Toggle opens/closes the sidebar.
func (s *Sidebar) Toggle() {
	s.Open = !s.Open
}

// HandleKey processes keyboard input.
func (s *Sidebar) HandleKey(ev term.KeyEvent) bool {
	if !s.Open {
		return false
	}
	switch ev.Key {
	case term.KeyUp:
		if s.Cursor > 0 {
			s.Cursor--
		}
	case term.KeyDown:
		if s.Cursor < len(s.Flat)-1 {
			s.Cursor++
		}
	case term.KeyEnter:
		s.activate()
	case term.KeyRight:
		if s.Cursor < len(s.Flat) {
			entry := s.Flat[s.Cursor]
			if entry.IsDir && !entry.Expanded {
				s.expand(entry)
			}
		}
	case term.KeyLeft:
		if s.Cursor < len(s.Flat) {
			entry := s.Flat[s.Cursor]
			if entry.IsDir && entry.Expanded {
				entry.Expanded = false
				s.rebuildFlat()
			}
		}
	default:
		return false
	}
	s.scrollIntoView()
	return true
}

// HandleClick processes a mouse click.
func (s *Sidebar) HandleClick(x, y, offsetY int) bool {
	if !s.Open || x >= s.Width {
		return false
	}
	idx := s.Scroll + (y - offsetY)
	if idx >= 0 && idx < len(s.Flat) {
		s.Cursor = idx
		s.activate()
		return true
	}
	return false
}

// Draw renders the sidebar.
func (s *Sidebar) Draw(scr *term.Screen, x, y, h int, theme SidebarTheme) {
	if !s.Open {
		return
	}

	scr.FillRect(x, y, s.Width, h, theme.Normal, ' ')
	scr.DrawText(x, y, theme.Bold, " EXPLORER", s.Width)

	for row := 0; row < h-1; row++ {
		idx := s.Scroll + row
		if idx >= len(s.Flat) {
			break
		}
		entry := s.Flat[idx]
		screenY := y + 1 + row

		style := theme.Normal
		if idx == s.Cursor {
			style = theme.Selection
		}

		for col := x; col < x+s.Width; col++ {
			scr.SetContent(col, screenY, ' ', style)
		}

		indent := strings.Repeat("  ", entry.Depth)
		icon := fileIcon(entry)
		text := indent + icon + entry.Name
		scr.DrawText(x+1, screenY, style, text, s.Width-2)
	}
}

func (s *Sidebar) activate() {
	if s.Cursor >= len(s.Flat) {
		return
	}
	entry := s.Flat[s.Cursor]
	if entry.IsDir {
		if entry.Expanded {
			entry.Expanded = false
		} else {
			s.expand(entry)
		}
		s.rebuildFlat()
	} else if s.OnFileSelect != nil {
		s.OnFileSelect(entry.Path)
	}
}

func (s *Sidebar) expand(entry *FileEntry) {
	if len(entry.Children) == 0 {
		entry.Children = scanChildren(entry.Path, entry.Depth+1)
	}
	entry.Expanded = true
	s.rebuildFlat()
}

func (s *Sidebar) rebuildFlat() {
	s.Flat = nil
	if s.Root != nil {
		s.flattenEntry(s.Root)
	}
}

func (s *Sidebar) flattenEntry(entry *FileEntry) {
	if entry == s.Root {
		if entry.Expanded {
			for _, child := range entry.Children {
				s.flattenChildren(child)
			}
		}
		return
	}
	s.flattenChildren(entry)
}

func (s *Sidebar) flattenChildren(entry *FileEntry) {
	s.Flat = append(s.Flat, entry)
	if entry.IsDir && entry.Expanded {
		for _, child := range entry.Children {
			s.flattenChildren(child)
		}
	}
}

func (s *Sidebar) scrollIntoView() {
	visible := 20
	if s.Cursor < s.Scroll {
		s.Scroll = s.Cursor
	}
	if s.Cursor >= s.Scroll+visible {
		s.Scroll = s.Cursor - visible + 1
	}
}

func fileIcon(entry *FileEntry) string {
	if entry.IsDir {
		if entry.Expanded {
			return "▾ "
		}
		return "▸ "
	}
	ext := strings.ToLower(filepath.Ext(entry.Name))
	switch ext {
	case ".go":
		return "  "
	case ".mod", ".sum":
		return "  "
	case ".md":
		return "  "
	case ".json", ".yaml", ".yml", ".toml":
		return "  "
	default:
		return "  "
	}
}

func scanDir(path string, depth, maxDepth int) *FileEntry {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	entry := &FileEntry{
		Name:  filepath.Base(path),
		Path:  path,
		IsDir: info.IsDir(),
		Depth: depth,
	}
	if info.IsDir() && depth < maxDepth {
		entry.Children = scanChildren(path, depth+1)
	}
	return entry
}

func scanChildren(path string, depth int) []*FileEntry {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}
	var dirs, files []*FileEntry
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		child := &FileEntry{
			Name:  name,
			Path:  filepath.Join(path, name),
			IsDir: e.IsDir(),
			Depth: depth,
		}
		if e.IsDir() {
			dirs = append(dirs, child)
		} else {
			files = append(files, child)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name < dirs[j].Name })
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	return append(dirs, files...)
}
