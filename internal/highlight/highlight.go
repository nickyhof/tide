package highlight

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nickyhof/tide/internal/term"
)

// TokenKind classifies a syntax token.
type TokenKind int

const (
	Normal TokenKind = iota
	Keyword
	Type
	String
	Number
	Comment
	Function
	Operator
	Builtin
	Punctuation
)

// Token is a colored span within a line.
type Token struct {
	Start int
	End   int
	Kind  TokenKind
}

// Palette maps token kinds to styles.
type Palette struct {
	styles map[TokenKind]term.Style
}

// DefaultPalette returns a VSCode-dark-inspired color palette.
func DefaultPalette(bg term.Color) *Palette {
	p := &Palette{styles: map[TokenKind]term.Style{
		Normal:      term.NewStyle(term.RGB(212, 212, 212), bg),
		Keyword:     term.NewStyle(term.RGB(197, 134, 192), bg),
		Type:        term.NewStyle(term.RGB(78, 201, 176), bg),
		String:      term.NewStyle(term.RGB(206, 145, 120), bg),
		Number:      term.NewStyle(term.RGB(181, 206, 168), bg),
		Comment:     term.NewStyle(term.RGB(106, 153, 85), bg),
		Function:    term.NewStyle(term.RGB(220, 220, 170), bg),
		Operator:    term.NewStyle(term.RGB(212, 212, 212), bg),
		Builtin:     term.NewStyle(term.RGB(86, 156, 214), bg),
		Punctuation: term.NewStyle(term.RGB(212, 212, 212), bg),
	}}
	return p
}

// Style returns the style for a token kind.
func (p *Palette) Style(kind TokenKind) term.Style {
	if s, ok := p.styles[kind]; ok {
		return s
	}
	return p.styles[Normal]
}

// Rule defines a regex pattern and its token kind.
type Rule struct {
	Pattern   *regexp.Regexp
	Kind      TokenKind
	TrimRight int
}

// Language holds the highlighting rules for a language.
type Language struct {
	Name          string
	Extensions    []string
	Rules         []Rule
	LineComment   string
	BlockComStart string
	BlockComEnd   string
}

// Highlighter produces tokens for lines of code.
type Highlighter struct {
	lang    *Language
	Palette *Palette
}

var languages []*Language

// Register adds a language to the global registry.
func Register(lang *Language) {
	languages = append(languages, lang)
}

// ForFile returns a highlighter for the given filename, or nil if unknown.
func ForFile(filename string, bg term.Color) *Highlighter {
	ext := strings.ToLower(filepath.Ext(filename))
	base := strings.ToLower(filepath.Base(filename))
	for _, lang := range languages {
		for _, e := range lang.Extensions {
			if e == ext || e == base {
				return &Highlighter{lang: lang, Palette: DefaultPalette(bg)}
			}
		}
	}
	return nil
}

// Highlight tokenizes a single line.
func (h *Highlighter) Highlight(line string, inBlockComment bool) ([]Token, bool) {
	tokens := make([]Token, 0, 8)
	covered := make([]bool, len(line))

	// Block comments spanning lines
	if inBlockComment && h.lang.BlockComEnd != "" {
		idx := strings.Index(line, h.lang.BlockComEnd)
		if idx == -1 {
			return []Token{{Start: 0, End: len(line), Kind: Comment}}, true
		}
		end := idx + len(h.lang.BlockComEnd)
		tokens = append(tokens, Token{Start: 0, End: end, Kind: Comment})
		markCovered(covered, 0, end)
		inBlockComment = false
	}

	// Block comment starts
	if h.lang.BlockComStart != "" {
		remaining := line
		offset := 0
		for {
			idx := strings.Index(remaining, h.lang.BlockComStart)
			if idx == -1 {
				break
			}
			start := offset + idx
			if isCovered(covered, start) {
				offset = start + 1
				remaining = line[offset:]
				continue
			}
			endIdx := strings.Index(remaining[idx+len(h.lang.BlockComStart):], h.lang.BlockComEnd)
			if endIdx == -1 {
				tokens = append(tokens, Token{Start: start, End: len(line), Kind: Comment})
				markCovered(covered, start, len(line))
				return tokens, true
			}
			end := start + len(h.lang.BlockComStart) + endIdx + len(h.lang.BlockComEnd)
			tokens = append(tokens, Token{Start: start, End: end, Kind: Comment})
			markCovered(covered, start, end)
			offset = end
			remaining = line[offset:]
		}
	}

	// Line comments
	if h.lang.LineComment != "" {
		idx := findUnquoted(line, h.lang.LineComment, covered)
		if idx >= 0 {
			tokens = append(tokens, Token{Start: idx, End: len(line), Kind: Comment})
			markCovered(covered, idx, len(line))
		}
	}

	// Strings
	for _, q := range []byte{'"', '\'', '`'} {
		i := 0
		for i < len(line) {
			if line[i] == q && !isCovered(covered, i) {
				start := i
				i++
				for i < len(line) {
					if line[i] == '\\' && q != '`' {
						i += 2
						continue
					}
					if line[i] == q {
						i++
						break
					}
					i++
				}
				tokens = append(tokens, Token{Start: start, End: i, Kind: String})
				markCovered(covered, start, i)
			} else {
				i++
			}
		}
	}

	// Regex rules
	for _, rule := range h.lang.Rules {
		matches := rule.Pattern.FindAllStringIndex(line, -1)
		for _, m := range matches {
			if !isCovered(covered, m[0]) {
				end := m[1] - rule.TrimRight
				if end <= m[0] {
					end = m[1]
				}
				tokens = append(tokens, Token{Start: m[0], End: end, Kind: rule.Kind})
				markCovered(covered, m[0], end)
			}
		}
	}

	sortTokens(tokens)
	return tokens, inBlockComment
}

func findUnquoted(line, substr string, covered []bool) int {
	idx := 0
	for {
		pos := strings.Index(line[idx:], substr)
		if pos == -1 {
			return -1
		}
		abs := idx + pos
		if !isCovered(covered, abs) {
			return abs
		}
		idx = abs + 1
		if idx >= len(line) {
			return -1
		}
	}
}

func markCovered(covered []bool, start, end int) {
	for i := start; i < end && i < len(covered); i++ {
		covered[i] = true
	}
}

func isCovered(covered []bool, pos int) bool {
	return pos < len(covered) && covered[pos]
}

func sortTokens(tokens []Token) {
	for i := 1; i < len(tokens); i++ {
		key := tokens[i]
		j := i - 1
		for j >= 0 && tokens[j].Start > key.Start {
			tokens[j+1] = tokens[j]
			j--
		}
		tokens[j+1] = key
	}
}
