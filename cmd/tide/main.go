package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nickyhof/tide/internal/buffer"
	"github.com/nickyhof/tide/internal/command"
	"github.com/nickyhof/tide/internal/editor"
	"github.com/nickyhof/tide/internal/git"
	"github.com/nickyhof/tide/internal/term"
	"github.com/nickyhof/tide/internal/ui"
)

const sidebarWidth = 28
const gitPanelWidth = 50

// theme holds all styles for the application, defined once.
type theme struct {
	Bg           term.Style
	Fg           term.Style
	StatusBar    term.Style
	StatusAccent term.Style
	TabActive    term.Style
	TabInactive  term.Style
	LineNumber   term.Style
	Selection    term.Style
	Sidebar      term.Style
	SidebarBold  term.Style
	HelpKey      term.Style
	HelpDesc     term.Style
}

func newTheme() theme {
	bg := term.RGB(30, 30, 30)
	fg := term.RGB(212, 212, 212)
	statusBg := term.RGB(0, 122, 204)
	sidebarBg := term.RGB(37, 37, 38)
	tabActiveBg := term.RGB(30, 30, 30)
	tabInactiveBg := term.RGB(45, 45, 45)

	return theme{
		Bg:           term.NewStyle(fg, bg),
		Fg:           term.NewStyle(fg, bg),
		StatusBar:    term.NewStyle(term.RGB(255, 255, 255), statusBg),
		StatusAccent: term.NewStyle(term.RGB(255, 255, 255), statusBg).WithBold(true),
		TabActive:    term.NewStyle(fg, tabActiveBg),
		TabInactive:  term.NewStyle(term.RGB(150, 150, 150), tabInactiveBg),
		LineNumber:   term.NewStyle(term.RGB(133, 133, 133), bg),
		Selection:    term.NewStyle(fg, term.RGB(64, 64, 64)),
		Sidebar:      term.NewStyle(fg, sidebarBg),
		SidebarBold:  term.NewStyle(fg, sidebarBg).WithBold(true),
		HelpKey:      term.NewStyle(term.RGB(255, 255, 255), sidebarBg).WithBold(true),
		HelpDesc:     term.NewStyle(term.RGB(150, 150, 150), sidebarBg),
	}
}

// App is the top-level application state.
type App struct {
	screen   *term.Screen
	editors  []*editor.Editor
	active   int
	commands *command.Registry
	sidebar  *ui.Sidebar
	palette  *ui.Palette
	gitPanel *ui.GitPanel
	repo     *git.Repo
	theme    theme
	message  string
	quit     bool
	focus    focusArea
}

type focusArea int

const (
	focusEditor focusArea = iota
	focusSidebar
	focusGit
)

func main() {
	app, err := newApp(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "tide: %v\n", err)
		os.Exit(1)
	}
	defer app.screen.Fini()
	app.run()
}

func newApp(args []string) (*App, error) {
	scr, err := term.NewScreen()
	if err != nil {
		return nil, err
	}

	cwd, _ := os.Getwd()
	repo := git.Open(cwd)

	a := &App{
		screen:   scr,
		commands: command.NewRegistry(),
		sidebar:  ui.NewSidebar(cwd, sidebarWidth),
		palette:  ui.NewPalette(),
		gitPanel: ui.NewGitPanel(repo),
		repo:     repo,
		theme:    newTheme(),
		focus:    focusEditor,
	}

	a.sidebar.OnFileSelect = func(path string) {
		a.openFile(path)
		a.focus = focusEditor
	}
	a.palette.OnSelect = func(name string) {
		a.commands.Execute(name)
	}
	a.gitPanel.OnOpenFile = func(path string) {
		a.openFile(path)
		a.focus = focusEditor
	}

	if len(args) == 0 {
		a.editors = []*editor.Editor{editor.NewEditor(buffer.New())}
	} else {
		for _, path := range args {
			absPath, _ := filepath.Abs(path)
			buf, err := buffer.NewFromFile(absPath)
			if err != nil {
				buf = buffer.New()
				buf.FilePath = absPath
			}
			a.editors = append(a.editors, editor.NewEditor(buf))
		}
	}

	a.registerCommands()
	return a, nil
}

func (a *App) openFile(path string) {
	for i, ed := range a.editors {
		if ed.Buffer.FilePath == path {
			a.active = i
			return
		}
	}
	buf, err := buffer.NewFromFile(path)
	if err != nil {
		a.message = fmt.Sprintf("Error: %v", err)
		return
	}
	a.editors = append(a.editors, editor.NewEditor(buf))
	a.active = len(a.editors) - 1
}

func (a *App) registerCommands() {
	a.commands.Register(command.Command{
		Name: "palette", Description: "Command Palette", Shortcut: "F1",
		Execute: func() { a.showPalette() },
	})
	a.commands.Register(command.Command{
		Name: "save", Description: "Save File", Shortcut: "F2",
		Execute: func() { a.save() },
	})
	a.commands.Register(command.Command{
		Name: "new", Description: "New File", Shortcut: "F3",
		Execute: func() {
			a.editors = append(a.editors, editor.NewEditor(buffer.New()))
			a.active = len(a.editors) - 1
		},
	})
	a.commands.Register(command.Command{
		Name: "close", Description: "Close Tab", Shortcut: "F4",
		Execute: func() { a.closeTab() },
	})
	a.commands.Register(command.Command{
		Name: "explorer", Description: "Toggle Explorer", Shortcut: "F5",
		Execute: func() {
			if a.sidebar.Open && a.focus == focusSidebar {
				a.sidebar.Toggle()
				a.focus = focusEditor
			} else if a.sidebar.Open {
				a.focus = focusSidebar
			} else {
				a.sidebar.Toggle()
				a.focus = focusSidebar
			}
		},
	})
	a.commands.Register(command.Command{
		Name: "git", Description: "Toggle Git Panel", Shortcut: "F6",
		Execute: func() {
			if a.gitPanel.Open && a.focus == focusGit {
				a.gitPanel.Toggle()
				a.focus = focusEditor
			} else if a.gitPanel.Open {
				a.focus = focusGit
			} else {
				a.gitPanel.Toggle()
				a.focus = focusGit
			}
		},
	})
	a.commands.Register(command.Command{
		Name: "prev-tab", Description: "Previous Tab", Shortcut: "F7",
		Execute: func() {
			if a.active > 0 {
				a.active--
			}
		},
	})
	a.commands.Register(command.Command{
		Name: "next-tab", Description: "Next Tab", Shortcut: "F8",
		Execute: func() {
			if a.active < len(a.editors)-1 {
				a.active++
			}
		},
	})
	a.commands.Register(command.Command{
		Name: "quit", Description: "Quit Tide", Shortcut: "F10",
		Execute: func() { a.quit = true },
	})
	a.commands.Register(command.Command{
		Name: "focus-editor", Description: "Focus Editor",
		Execute: func() { a.focus = focusEditor },
	})
	a.commands.Register(command.Command{
		Name: "focus-explorer", Description: "Focus Explorer",
		Execute: func() {
			if !a.sidebar.Open {
				a.sidebar.Toggle()
			}
			a.focus = focusSidebar
		},
	})
	a.commands.Register(command.Command{
		Name: "focus-git", Description: "Focus Git Panel",
		Execute: func() {
			if !a.gitPanel.Open {
				a.gitPanel.Toggle()
			}
			a.focus = focusGit
		},
	})
	a.commands.Register(command.Command{
		Name: "git-refresh", Description: "Git: Refresh Status",
		Execute: func() { a.gitPanel.Refresh() },
	})
}

func (a *App) showPalette() {
	cmds := a.commands.List()
	items := make([]ui.PaletteItem, len(cmds))
	for i, cmd := range cmds {
		items[i] = ui.PaletteItem{
			Name:     cmd.Name,
			Desc:     cmd.Description,
			Shortcut: cmd.Shortcut,
		}
	}
	a.palette.Show(items)
}

func (a *App) activeEditor() *editor.Editor {
	if a.active < len(a.editors) {
		return a.editors[a.active]
	}
	return nil
}

func (a *App) save() {
	ed := a.activeEditor()
	if ed == nil {
		return
	}
	if err := ed.Buffer.Save(); err != nil {
		a.message = fmt.Sprintf("Error saving: %v", err)
	} else {
		a.message = fmt.Sprintf("Saved %s", ed.Buffer.FilePath)
	}
}

func (a *App) closeTab() {
	if len(a.editors) <= 1 {
		a.quit = true
		return
	}
	a.editors = append(a.editors[:a.active], a.editors[a.active+1:]...)
	if a.active >= len(a.editors) {
		a.active = len(a.editors) - 1
	}
}

func (a *App) run() {
	for !a.quit {
		a.draw()
		a.screen.Show()

		ev := a.screen.PollEvent()
		switch ev := ev.(type) {
		case term.ResizeEvent:
			a.screen.Resize()
			a.screen.Sync()
		case term.KeyEvent:
			a.handleKey(ev)
		case term.MouseEvent:
			a.handleMouse(ev)
		}
	}
}

func (a *App) handleKey(ev term.KeyEvent) {
	if a.palette.Open {
		a.palette.HandleKey(ev)
		return
	}

	switch ev.Key {
	case term.KeyF1:
		a.showPalette()
		return
	case term.KeyF2:
		a.commands.Execute("save")
		return
	case term.KeyF3:
		a.commands.Execute("new")
		return
	case term.KeyF4:
		a.commands.Execute("close")
		return
	case term.KeyF5:
		a.commands.Execute("explorer")
		return
	case term.KeyF6:
		a.commands.Execute("git")
		return
	case term.KeyF7:
		a.commands.Execute("prev-tab")
		return
	case term.KeyF8:
		a.commands.Execute("next-tab")
		return
	case term.KeyF10:
		a.commands.Execute("quit")
		return
	case term.KeyEscape:
		a.focus = focusEditor
		return
	case term.KeyCtrlS:
		a.save()
		return
	}

	switch a.focus {
	case focusSidebar:
		a.sidebar.HandleKey(ev)
	case focusGit:
		a.gitPanel.HandleKey(ev)
	default:
		if ed := a.activeEditor(); ed != nil {
			ed.HandleKey(ev)
			a.message = ""
		}
	}
}

func (a *App) handleMouse(ev term.MouseEvent) {
	x, y := ev.X, ev.Y

	if ev.Button != term.Button1 {
		return
	}

	if a.palette.Open {
		a.palette.Close()
		return
	}

	sideW := 0
	if a.sidebar.Open {
		sideW = a.sidebar.Width
	}

	gitW := 0
	if a.gitPanel.Open {
		gitW = gitPanelWidth
		if gitW > a.screen.Width/2 {
			gitW = a.screen.Width / 2
		}
	}
	gitX := a.screen.Width - gitW

	if a.sidebar.Open && x < sideW {
		a.focus = focusSidebar
		if y >= 1 && y < a.screen.Height-2 {
			a.sidebar.HandleClick(x, y, 1)
		}
		return
	}

	if a.gitPanel.Open && x >= gitX {
		a.focus = focusGit
		return
	}

	if y == 0 {
		a.clickTab(x - sideW)
		a.focus = focusEditor
		return
	}

	a.focus = focusEditor
	if ed := a.activeEditor(); ed != nil {
		editorY := 1
		editorH := a.screen.Height - 3
		if y >= editorY && y < editorY+editorH {
			row := ed.ScrollRow + (y - editorY)
			col := ed.ScrollCol + (x - sideW - 5)
			if col < 0 {
				col = 0
			}
			ed.SetCursor(row, col)
		}
	}
}

func (a *App) clickTab(x int) {
	pos := 0
	for i, ed := range a.editors {
		title := " " + ed.Title() + " "
		end := pos + len(title)
		if x >= pos && x < end {
			a.active = i
			return
		}
		pos = end
	}
}

func (a *App) draw() {
	a.screen.Clear()
	a.screen.FillRect(0, 0, a.screen.Width, a.screen.Height, a.theme.Bg, ' ')

	sideW := 0
	if a.sidebar.Open {
		sideW = a.sidebar.Width
	}
	gitW := 0
	if a.gitPanel.Open {
		gitW = gitPanelWidth
		if gitW > a.screen.Width/2 {
			gitW = a.screen.Width / 2
		}
	}

	tabBarY := 0
	editorY := 1
	editorH := a.screen.Height - 3
	helpY := a.screen.Height - 2
	statusY := a.screen.Height - 1
	editorW := a.screen.Width - sideW - gitW

	a.drawTabBar(tabBarY, sideW, sideW+editorW)

	if a.sidebar.Open {
		a.sidebar.Draw(a.screen, 0, tabBarY, editorH+1, ui.SidebarTheme{
			Normal:    a.theme.Sidebar,
			Bold:      a.theme.SidebarBold,
			Selection: a.theme.Selection,
		})
	}

	if ed := a.activeEditor(); ed != nil {
		ed.Draw(a.screen, sideW, editorY, editorW, editorH, editor.Theme{
			Foreground: a.theme.Fg,
			LineNumber: a.theme.LineNumber,
		})

		if a.focus == focusEditor && !a.palette.Open {
			cx, cy := ed.CursorScreenPos(sideW, editorY)
			a.screen.ShowCursor(cx, cy)
		} else {
			a.screen.HideCursor()
		}
	}

	if a.gitPanel.Open && gitW > 0 {
		gitX := a.screen.Width - gitW
		bg := term.RGB(30, 30, 30)
		a.gitPanel.Draw(a.screen, gitX, editorY, gitW, editorH, ui.GitPanelTheme{
			Normal:    a.theme.Sidebar,
			Bold:      a.theme.SidebarBold,
			Staged:    term.NewStyle(term.RGB(80, 200, 120), term.RGB(37, 37, 38)),
			Modified:  term.NewStyle(term.RGB(220, 180, 60), term.RGB(37, 37, 38)),
			Untracked: term.NewStyle(term.RGB(150, 150, 150), term.RGB(37, 37, 38)),
			Selected:  a.theme.Selection,
			Hint:      term.NewStyle(term.RGB(100, 100, 100), bg),
			Separator: a.theme.Sidebar.WithDim(true),
			DiffAdd:   term.NewStyle(term.RGB(80, 200, 120), term.RGB(30, 50, 30)),
			DiffRm:    term.NewStyle(term.RGB(220, 80, 80), term.RGB(50, 30, 30)),
			DiffHunk:  term.NewStyle(term.RGB(86, 156, 214), bg),
			DiffHead:  term.NewStyle(term.RGB(150, 150, 150), bg).WithBold(true),
			DiffCtx:   a.theme.Fg,
			TabBar:    a.theme.TabInactive,
		})
	}

	a.drawHelpBar(helpY)
	a.drawStatusBar(statusY)

	bg := term.RGB(37, 37, 38)
	a.palette.Draw(a.screen, ui.PaletteTheme{
		Border:      term.NewStyle(term.RGB(80, 80, 80), bg),
		Input:       term.NewStyle(term.RGB(230, 230, 230), term.RGB(60, 60, 60)),
		Item:        term.NewStyle(term.RGB(200, 200, 200), bg),
		Selected:    term.NewStyle(term.RGB(255, 255, 255), term.RGB(4, 57, 94)),
		Shortcut:    term.NewStyle(term.RGB(120, 120, 120), bg),
		ShortcutSel: term.NewStyle(term.RGB(180, 180, 180), term.RGB(4, 57, 94)),
	})
}

func (a *App) drawTabBar(y, offsetX, maxX int) {
	a.screen.FillRect(offsetX, y, maxX-offsetX, 1, a.theme.TabInactive, ' ')
	x := offsetX
	for i, ed := range a.editors {
		title := " " + ed.Title() + " "
		style := a.theme.TabInactive
		if i == a.active {
			style = a.theme.TabActive
		}
		a.screen.DrawText(x, y, style, title, len(title))
		x += len(title)
	}
}

func (a *App) drawHelpBar(y int) {
	a.screen.FillRect(0, y, a.screen.Width, 1, a.theme.HelpDesc, ' ')

	type shortcut struct {
		key  string
		desc string
	}
	shortcuts := []shortcut{
		{"F1", "Palette"},
		{"F2", "Save"},
		{"F3", "New"},
		{"F4", "Close"},
		{"F5", "Explorer"},
		{"F6", "Git"},
		{"F7/F8", "Tabs"},
		{"F10", "Quit"},
		{"Esc", "Editor"},
	}

	x := 1
	for _, sc := range shortcuts {
		a.screen.DrawText(x, y, a.theme.HelpKey, sc.key, len(sc.key))
		x += len(sc.key)
		desc := " " + sc.desc + "  "
		a.screen.DrawText(x, y, a.theme.HelpDesc, desc, len(desc))
		x += len(desc)
		if x >= a.screen.Width-10 {
			break
		}
	}
}

func (a *App) drawStatusBar(y int) {
	a.screen.FillRect(0, y, a.screen.Width, 1, a.theme.StatusBar, ' ')

	ed := a.activeEditor()
	if ed == nil {
		return
	}

	var left string
	if a.message != "" {
		left = " " + a.message
	} else {
		path := ed.Buffer.FilePath
		if path == "" {
			path = "Untitled"
		}
		mod := ""
		if ed.Buffer.Modified {
			mod = " [Modified]"
		}
		left = fmt.Sprintf(" %s%s", filepath.Base(path), mod)
	}
	a.screen.DrawText(0, y, a.theme.StatusBar, left, a.screen.Width/3)

	cursor := fmt.Sprintf("Ln %d, Col %d  │  %d lines", ed.CursorRow+1, ed.CursorCol+1, ed.Buffer.LineCount)
	gitInfo := ""
	if summary := ui.StatusSummary(a.repo); summary != "" {
		gitInfo = "  │  " + summary
	}
	right := cursor + gitInfo + "  │  UTF-8 "
	rightX := a.screen.Width - len(right)
	if rightX < 0 {
		rightX = 0
	}
	a.screen.DrawText(rightX, y, a.theme.StatusBar, right, len(right))

	mode := " TIDE "
	modeX := (a.screen.Width - len(mode)) / 2
	a.screen.DrawText(modeX, y, a.theme.StatusAccent, mode, len(mode))
}
