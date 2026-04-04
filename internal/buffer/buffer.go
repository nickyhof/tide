package buffer

import (
	"os"
	"strings"
)

// Piece represents a span of text referencing either the original or add buffer.
type Piece struct {
	Source Source
	Start  int
	Length int
}

// Source indicates which buffer a piece references.
type Source int

const (
	Original Source = iota
	Add
)

// Buffer implements a piece table for efficient text editing.
// This is the same data structure used by VSCode.
type Buffer struct {
	original  string
	add       strings.Builder
	pieces    []Piece
	FilePath  string
	Modified  bool
	LineCount int
}

// New creates an empty buffer.
func New() *Buffer {
	b := &Buffer{
		pieces:    []Piece{},
		LineCount: 1,
	}
	return b
}

// NewFromString creates a buffer initialized with the given text.
func NewFromString(text string) *Buffer {
	b := &Buffer{
		original:  text,
		LineCount: countLines(text),
	}
	if len(text) > 0 {
		b.pieces = []Piece{{Source: Original, Start: 0, Length: len(text)}}
	} else {
		b.pieces = []Piece{}
	}
	return b
}

// NewFromFile reads a file and creates a buffer from its contents.
func NewFromFile(path string) (*Buffer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	b := NewFromString(string(data))
	b.FilePath = path
	return b, nil
}

// Text returns the full content of the buffer.
func (b *Buffer) Text() string {
	var sb strings.Builder
	for _, p := range b.pieces {
		sb.WriteString(b.source(p))
	}
	return sb.String()
}

// Length returns the total number of bytes in the buffer.
func (b *Buffer) Length() int {
	n := 0
	for _, p := range b.pieces {
		n += p.Length
	}
	return n
}

// Insert inserts text at the given byte offset.
func (b *Buffer) Insert(offset int, text string) {
	if len(text) == 0 {
		return
	}

	addStart := b.add.Len()
	b.add.WriteString(text)
	newPiece := Piece{Source: Add, Start: addStart, Length: len(text)}

	if len(b.pieces) == 0 {
		b.pieces = []Piece{newPiece}
		b.Modified = true
		b.LineCount = countLines(b.Text())
		return
	}

	idx, within := b.findPiece(offset)

	if within == 0 {
		// Insert before piece at idx
		b.pieces = insertPiece(b.pieces, idx, newPiece)
	} else if within == b.pieces[idx].Length {
		// Insert after piece at idx
		b.pieces = insertPiece(b.pieces, idx+1, newPiece)
	} else {
		// Split piece at idx
		orig := b.pieces[idx]
		left := Piece{Source: orig.Source, Start: orig.Start, Length: within}
		right := Piece{Source: orig.Source, Start: orig.Start + within, Length: orig.Length - within}
		replacement := []Piece{left, newPiece, right}
		b.pieces = replacePieces(b.pieces, idx, 1, replacement)
	}

	b.Modified = true
	b.LineCount = countLines(b.Text())
}

// Delete removes length bytes starting at offset.
func (b *Buffer) Delete(offset, length int) {
	if length == 0 {
		return
	}

	remaining := length
	pos := offset

	for remaining > 0 && len(b.pieces) > 0 {
		idx, within := b.findPiece(pos)
		if idx >= len(b.pieces) {
			break
		}

		p := b.pieces[idx]
		canDelete := p.Length - within
		if canDelete > remaining {
			canDelete = remaining
		}

		if within == 0 && canDelete == p.Length {
			// Remove entire piece
			b.pieces = append(b.pieces[:idx], b.pieces[idx+1:]...)
		} else if within == 0 {
			// Trim from start
			b.pieces[idx] = Piece{Source: p.Source, Start: p.Start + canDelete, Length: p.Length - canDelete}
		} else if within+canDelete == p.Length {
			// Trim from end
			b.pieces[idx] = Piece{Source: p.Source, Start: p.Start, Length: within}
		} else {
			// Split and remove middle
			left := Piece{Source: p.Source, Start: p.Start, Length: within}
			right := Piece{Source: p.Source, Start: p.Start + within + canDelete, Length: p.Length - within - canDelete}
			b.pieces = replacePieces(b.pieces, idx, 1, []Piece{left, right})
		}

		remaining -= canDelete
	}

	b.Modified = true
	b.LineCount = countLines(b.Text())
}

// Line returns the content of the given line (0-indexed), including the newline if present.
func (b *Buffer) Line(line int) string {
	text := b.Text()
	current := 0
	start := 0
	for i, ch := range text {
		if current == line {
			// Find end of this line
			for j := i; j < len(text); j++ {
				if text[j] == '\n' {
					return text[i : j+1]
				}
			}
			return text[i:]
		}
		if ch == '\n' {
			current++
			start = i + 1
		}
	}
	_ = start
	return ""
}

// LineStart returns the byte offset of the start of the given line (0-indexed).
func (b *Buffer) LineStart(line int) int {
	if line == 0 {
		return 0
	}
	text := b.Text()
	current := 0
	for i, ch := range text {
		if ch == '\n' {
			current++
			if current == line {
				return i + 1
			}
		}
	}
	return len(text)
}

// LineLength returns the length of the given line excluding the newline.
func (b *Buffer) LineLength(line int) int {
	l := b.Line(line)
	l = strings.TrimRight(l, "\n")
	return len(l)
}

// Save writes the buffer to the file it was loaded from.
func (b *Buffer) Save() error {
	if b.FilePath == "" {
		return nil
	}
	return b.SaveAs(b.FilePath)
}

// SaveAs writes the buffer to the given path.
func (b *Buffer) SaveAs(path string) error {
	err := os.WriteFile(path, []byte(b.Text()), 0644)
	if err != nil {
		return err
	}
	b.FilePath = path
	b.Modified = false
	return nil
}

// source returns the text of a piece.
func (b *Buffer) source(p Piece) string {
	var src string
	if p.Source == Original {
		src = b.original
	} else {
		src = b.add.String()
	}
	return src[p.Start : p.Start+p.Length]
}

// findPiece returns the piece index and offset within that piece for the given byte offset.
func (b *Buffer) findPiece(offset int) (int, int) {
	pos := 0
	for i, p := range b.pieces {
		if offset <= pos+p.Length {
			return i, offset - pos
		}
		pos += p.Length
	}
	return len(b.pieces), 0
}

func insertPiece(pieces []Piece, idx int, p Piece) []Piece {
	pieces = append(pieces, Piece{})
	copy(pieces[idx+1:], pieces[idx:])
	pieces[idx] = p
	return pieces
}

func replacePieces(pieces []Piece, idx, count int, replacement []Piece) []Piece {
	tail := append([]Piece{}, pieces[idx+count:]...)
	pieces = append(pieces[:idx], replacement...)
	pieces = append(pieces, tail...)
	return pieces
}

func countLines(text string) int {
	if len(text) == 0 {
		return 1
	}
	n := 1
	for _, ch := range text {
		if ch == '\n' {
			n++
		}
	}
	return n
}
