package term

import (
	"os"
	"unicode/utf8"
)

// Event is the interface for all terminal events.
type Event interface {
	eventTag()
}

// KeyEvent represents a key press.
type KeyEvent struct {
	Key  Key
	Rune rune
}

func (KeyEvent) eventTag() {}

// MouseEvent represents a mouse action.
type MouseEvent struct {
	X, Y   int
	Button MouseButton
}

func (MouseEvent) eventTag() {}

// ResizeEvent indicates the terminal was resized.
type ResizeEvent struct {
	Width, Height int
}

func (ResizeEvent) eventTag() {}

// Key identifies special keys.
type Key int

const (
	KeyNone Key = iota
	KeyRune     // regular character — check Rune field
	KeyEnter
	KeyTab
	KeyBackspace
	KeyEscape
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPgUp
	KeyPgDn
	KeyDelete
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyCtrlA
	KeyCtrlB
	KeyCtrlC
	KeyCtrlD
	KeyCtrlE
	KeyCtrlK
	KeyCtrlL
	KeyCtrlN
	KeyCtrlR
	KeyCtrlS
	KeyCtrlT
	KeyCtrlU
	KeyCtrlW
	KeyCtrlZ
)

// MouseButton identifies mouse buttons.
type MouseButton int

const (
	ButtonNone MouseButton = iota
	Button1                // left click
	Button2                // middle click
	Button3                // right click
)

// InputReader reads and parses terminal input.
type InputReader struct {
	file *os.File
	buf  [256]byte
}

// NewInputReader creates an input reader on the given file (usually os.Stdin).
func NewInputReader(f *os.File) *InputReader {
	return &InputReader{file: f}
}

// Read blocks until an event is available.
func (ir *InputReader) Read() Event {
	for {
		n, err := ir.file.Read(ir.buf[:])
		if err != nil || n == 0 {
			continue
		}
		data := ir.buf[:n]
		ev := ir.parse(data)
		if ev != nil {
			return ev
		}
	}
}

func (ir *InputReader) parse(data []byte) Event {
	if len(data) == 0 {
		return nil
	}

	// Escape sequences
	if data[0] == 0x1b {
		return ir.parseEscape(data)
	}

	// Ctrl keys (0x01-0x1a)
	if data[0] < 0x20 {
		return ir.parseControl(data[0])
	}

	// DEL (backspace on most terminals)
	if data[0] == 0x7f {
		return KeyEvent{Key: KeyBackspace}
	}

	// Regular UTF-8 character
	r, _ := utf8.DecodeRune(data)
	if r != utf8.RuneError {
		return KeyEvent{Key: KeyRune, Rune: r}
	}

	return nil
}

func (ir *InputReader) parseControl(b byte) Event {
	switch b {
	case 0x01:
		return KeyEvent{Key: KeyCtrlA}
	case 0x02:
		return KeyEvent{Key: KeyCtrlB}
	case 0x03:
		return KeyEvent{Key: KeyCtrlC}
	case 0x04:
		return KeyEvent{Key: KeyCtrlD}
	case 0x05:
		return KeyEvent{Key: KeyCtrlE}
	case 0x08:
		return KeyEvent{Key: KeyBackspace}
	case 0x09:
		return KeyEvent{Key: KeyTab}
	case 0x0a, 0x0d:
		return KeyEvent{Key: KeyEnter}
	case 0x0b:
		return KeyEvent{Key: KeyCtrlK}
	case 0x0c:
		return KeyEvent{Key: KeyCtrlL}
	case 0x0e:
		return KeyEvent{Key: KeyCtrlN}
	case 0x12:
		return KeyEvent{Key: KeyCtrlR}
	case 0x13:
		return KeyEvent{Key: KeyCtrlS}
	case 0x14:
		return KeyEvent{Key: KeyCtrlT}
	case 0x15:
		return KeyEvent{Key: KeyCtrlU}
	case 0x17:
		return KeyEvent{Key: KeyCtrlW}
	case 0x1a:
		return KeyEvent{Key: KeyCtrlZ}
	}
	return nil
}

func (ir *InputReader) parseEscape(data []byte) Event {
	if len(data) == 1 {
		return KeyEvent{Key: KeyEscape}
	}

	// CSI sequences: ESC [
	if data[1] == '[' {
		return ir.parseCSI(data[2:])
	}

	// SS3 sequences: ESC O (some terminals use this for F1-F4)
	if data[1] == 'O' && len(data) >= 3 {
		switch data[2] {
		case 'P':
			return KeyEvent{Key: KeyF1}
		case 'Q':
			return KeyEvent{Key: KeyF2}
		case 'R':
			return KeyEvent{Key: KeyF3}
		case 'S':
			return KeyEvent{Key: KeyF4}
		}
	}

	return KeyEvent{Key: KeyEscape}
}

func (ir *InputReader) parseCSI(data []byte) Event {
	if len(data) == 0 {
		return KeyEvent{Key: KeyEscape}
	}

	// Simple arrow keys: ESC [ A/B/C/D
	switch {
	case len(data) >= 1 && data[0] == 'A':
		return KeyEvent{Key: KeyUp}
	case len(data) >= 1 && data[0] == 'B':
		return KeyEvent{Key: KeyDown}
	case len(data) >= 1 && data[0] == 'C':
		return KeyEvent{Key: KeyRight}
	case len(data) >= 1 && data[0] == 'D':
		return KeyEvent{Key: KeyLeft}
	case len(data) >= 1 && data[0] == 'H':
		return KeyEvent{Key: KeyHome}
	case len(data) >= 1 && data[0] == 'F':
		return KeyEvent{Key: KeyEnd}
	}

	// Numbered sequences: ESC [ N ~ or ESC [ N ; M ~
	// Collect digits and semicolons
	num := 0
	i := 0
	for i < len(data) && data[i] >= '0' && data[i] <= '9' {
		num = num*10 + int(data[i]-'0')
		i++
	}

	// Skip modifier parameter (e.g., ;2 for shift)
	if i < len(data) && data[i] == ';' {
		i++ // skip semicolon
		for i < len(data) && data[i] >= '0' && data[i] <= '9' {
			i++
		}
	}

	if i < len(data) && data[i] == '~' {
		switch num {
		case 1:
			return KeyEvent{Key: KeyHome}
		case 2:
			return nil // Insert — ignore
		case 3:
			return KeyEvent{Key: KeyDelete}
		case 4:
			return KeyEvent{Key: KeyEnd}
		case 5:
			return KeyEvent{Key: KeyPgUp}
		case 6:
			return KeyEvent{Key: KeyPgDn}
		case 11:
			return KeyEvent{Key: KeyF1}
		case 12:
			return KeyEvent{Key: KeyF2}
		case 13:
			return KeyEvent{Key: KeyF3}
		case 14:
			return KeyEvent{Key: KeyF4}
		case 15:
			return KeyEvent{Key: KeyF5}
		case 17:
			return KeyEvent{Key: KeyF6}
		case 18:
			return KeyEvent{Key: KeyF7}
		case 19:
			return KeyEvent{Key: KeyF8}
		case 20:
			return KeyEvent{Key: KeyF9}
		case 21:
			return KeyEvent{Key: KeyF10}
		case 23:
			return KeyEvent{Key: KeyF11}
		case 24:
			return KeyEvent{Key: KeyF12}
		}
	}

	// Mouse: ESC [ < Cb ; Cx ; Cy M/m (SGR mouse)
	if len(data) >= 1 && data[0] == '<' {
		return ir.parseSGRMouse(data[1:])
	}

	// Mouse: ESC [ M Cb Cx Cy (X10 mouse)
	if len(data) >= 4 && data[0] == 'M' {
		btn := data[1] - 32
		x := int(data[2]) - 33
		y := int(data[3]) - 33
		var button MouseButton
		switch btn & 0x03 {
		case 0:
			button = Button1
		case 1:
			button = Button2
		case 2:
			button = Button3
		default:
			button = ButtonNone
		}
		return MouseEvent{X: x, Y: y, Button: button}
	}

	return nil
}

func (ir *InputReader) parseSGRMouse(data []byte) Event {
	// Format: Cb;Cx;CyM or Cb;Cx;Cym
	// Cb=button, Cx=column(1-based), Cy=row(1-based), M=press, m=release
	params := [3]int{}
	pi := 0
	i := 0
	for i < len(data) && pi < 3 {
		if data[i] == ';' {
			pi++
			i++
			continue
		}
		if data[i] >= '0' && data[i] <= '9' {
			params[pi] = params[pi]*10 + int(data[i]-'0')
			i++
			continue
		}
		if data[i] == 'M' || data[i] == 'm' {
			isPress := data[i] == 'M'
			if !isPress {
				return nil // ignore release
			}
			btn := params[0]
			x := params[1] - 1
			y := params[2] - 1
			var button MouseButton
			switch btn & 0x03 {
			case 0:
				button = Button1
			case 1:
				button = Button2
			case 2:
				button = Button3
			default:
				button = ButtonNone
			}
			return MouseEvent{X: x, Y: y, Button: button}
		}
		break
	}
	return nil
}
