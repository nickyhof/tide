package term

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// Cell is a single character on screen with a style.
type Cell struct {
	Ch    rune
	Style Style
}

// Screen manages the terminal screen buffer and rendering.
type Screen struct {
	Width  int
	Height int
	front  []Cell // what's currently on the terminal
	back   []Cell // what we want on the terminal
	out    *os.File
	oldState *term.State
	input  *InputReader
	inputCh chan Event // single persistent channel fed by one goroutine
	curX   int // cursor position for Show
	curY   int
	sigCh  chan os.Signal
}

// NewScreen initializes raw mode and the alternate screen buffer.
func NewScreen() (*Screen, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("failed to set raw mode: %w", err)
	}

	out := os.Stdout
	w, h, err := term.GetSize(fd)
	if err != nil {
		term.Restore(fd, oldState)
		return nil, fmt.Errorf("failed to get terminal size: %w", err)
	}

	ir := NewInputReader(os.Stdin)
	inputCh := make(chan Event, 16)

	s := &Screen{
		Width:    w,
		Height:   h,
		front:    makeGrid(w, h),
		back:     makeGrid(w, h),
		out:      out,
		oldState: oldState,
		input:    ir,
		inputCh:  inputCh,
		curX:     -1,
		curY:     -1,
		sigCh:    make(chan os.Signal, 1),
	}

	// Switch to alternate screen, hide cursor, enable mouse
	s.write("\x1b[?1049h") // alternate screen
	s.write("\x1b[?25l")   // hide cursor
	s.write("\x1b[?1006h") // SGR mouse mode
	s.write("\x1b[?1003h") // all mouse events

	// Single persistent goroutine reads stdin and feeds inputCh
	go func() {
		for {
			ev := ir.Read()
			if ev != nil {
				inputCh <- ev
			}
		}
	}()

	// Listen for SIGWINCH
	signal.Notify(s.sigCh, syscall.SIGWINCH)

	return s, nil
}

// Fini restores the terminal to its original state.
func (s *Screen) Fini() {
	signal.Stop(s.sigCh)
	s.write("\x1b[?1003l") // disable mouse
	s.write("\x1b[?1006l") // disable SGR mouse
	s.write("\x1b[?25h")   // show cursor
	s.write("\x1b[?1049l") // restore main screen
	term.Restore(int(os.Stdin.Fd()), s.oldState)
}

// PollEvent blocks until an event is available.
func (s *Screen) PollEvent() Event {
	select {
	case ev := <-s.inputCh:
		return ev
	case <-s.sigCh:
		w, h, _ := term.GetSize(int(os.Stdin.Fd()))
		return ResizeEvent{Width: w, Height: h}
	}
}

// Resize updates the screen dimensions.
func (s *Screen) Resize() {
	w, h, _ := term.GetSize(int(os.Stdin.Fd()))
	if w == s.Width && h == s.Height {
		return
	}
	s.Width = w
	s.Height = h
	s.front = makeGrid(w, h)
	s.back = makeGrid(w, h)
}

// Clear fills the back buffer with spaces and default style.
func (s *Screen) Clear() {
	for i := range s.back {
		s.back[i] = Cell{Ch: ' ', Style: DefaultStyle}
	}
}

// SetContent sets a cell in the back buffer.
func (s *Screen) SetContent(x, y int, ch rune, style Style) {
	if x < 0 || x >= s.Width || y < 0 || y >= s.Height {
		return
	}
	s.back[y*s.Width+x] = Cell{Ch: ch, Style: style}
}

// ShowCursor positions the cursor. Pass -1,-1 to hide.
func (s *Screen) ShowCursor(x, y int) {
	s.curX = x
	s.curY = y
}

// HideCursor hides the cursor.
func (s *Screen) HideCursor() {
	s.curX = -1
	s.curY = -1
}

// Show flushes the back buffer to the terminal, only writing changed cells.
func (s *Screen) Show() {
	var buf strings.Builder
	buf.Grow(s.Width * s.Height * 4)

	var lastStyle Style
	styleSet := false

	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			i := y*s.Width + x
			bc := s.back[i]
			fc := s.front[i]

			if bc.Ch == fc.Ch && bc.Style == fc.Style {
				styleSet = false // force re-emit style after skipping
				continue
			}

			// Move cursor to position
			fmt.Fprintf(&buf, "\x1b[%d;%dH", y+1, x+1)

			// Emit style if different from last
			if !styleSet || bc.Style != lastStyle {
				buf.WriteString(bc.Style.ansiSeq())
				lastStyle = bc.Style
				styleSet = true
			}

			ch := bc.Ch
			if ch == 0 {
				ch = ' '
			}
			buf.WriteRune(ch)

			s.front[i] = bc
		}
	}

	// Position cursor
	if s.curX >= 0 && s.curY >= 0 {
		fmt.Fprintf(&buf, "\x1b[%d;%dH", s.curY+1, s.curX+1)
		buf.WriteString("\x1b[?25h") // show cursor
	} else {
		buf.WriteString("\x1b[?25l") // hide cursor
	}

	s.write(buf.String())
}

// Sync forces a full redraw (marks all front cells as dirty).
func (s *Screen) Sync() {
	for i := range s.front {
		s.front[i] = Cell{Ch: 0, Style: Style{Fg: RGB(255, 255, 255)}} // force mismatch
	}
}

// --- Drawing helpers ---

// DrawText draws a string at (x, y) clipped to maxWidth.
func (s *Screen) DrawText(x, y int, style Style, text string, maxWidth int) {
	col := 0
	for _, r := range text {
		if col >= maxWidth {
			break
		}
		s.SetContent(x+col, y, r, style)
		col++
	}
}

// FillRect fills a rectangle with a character and style.
func (s *Screen) FillRect(x, y, w, h int, style Style, ch rune) {
	for row := y; row < y+h && row < s.Height; row++ {
		for col := x; col < x+w && col < s.Width; col++ {
			s.SetContent(col, row, ch, style)
		}
	}
}

func (s *Screen) write(data string) {
	s.out.WriteString(data)
}

func makeGrid(w, h int) []Cell {
	grid := make([]Cell, w*h)
	for i := range grid {
		grid[i] = Cell{Ch: ' ', Style: DefaultStyle}
	}
	return grid
}
