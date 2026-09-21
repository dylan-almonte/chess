package board

import (
	"fmt"
	"strings"
	"unicode"
)

// Piece letters: uppercase white, lowercase black. Empty is 0.
type Piece byte

const (
	Empty Piece = 0
)

// Board is an 8×8 display board. Index 0 = a1 … 63 = h8 (LERF).
type Board struct {
	sq [64]Piece
}

func StartPos() Board {
	b := Board{}
	place := func(rank int, chars string) {
		for file, ch := range chars {
			b.sq[rank*8+file] = Piece(ch)
		}
	}
	place(0, "RNBQKBNR")
	place(1, "PPPPPPPP")
	place(6, "pppppppp")
	place(7, "rnbqkbnr")
	return b
}

func (b Board) At(sq int) Piece {
	if sq < 0 || sq >= 64 {
		return Empty
	}
	return b.sq[sq]
}

// ClearSquare sets the named square to empty.
func (b *Board) ClearSquare(name string) {
	sq, err := ParseSquare(name)
	if err != nil {
		return
	}
	b.sq[sq] = Empty
}

// SetSquare places a piece (given as a byte, e.g. 'P', 'p') on the named square.
func (b *Board) SetSquare(name string, p byte) {
	sq, err := ParseSquare(name)
	if err != nil {
		return
	}
	b.sq[sq] = Piece(p)
}

func (b Board) PieceAtName(name string) Piece {
	sq, err := ParseSquare(name)
	if err != nil {
		return Empty
	}
	return b.At(sq)
}

func ParseSquare(name string) (int, error) {
	if len(name) != 2 {
		return 0, fmt.Errorf("bad square %q", name)
	}
	file := name[0] - 'a'
	rank := name[1] - '1'
	if file > 7 || rank > 7 {
		return 0, fmt.Errorf("bad square %q", name)
	}
	return int(rank)*8 + int(file), nil
}

func SquareName(sq int) string {
	return string([]byte{byte('a' + sq%8), byte('1' + sq/8)})
}

// ApplyUCI applies a UCI long-algebraic move for display (move piece, capture,
// basic castling and promotion). Does not validate legality.
func (b *Board) ApplyUCI(uci string) error {
	uci = strings.TrimSpace(uci)
	if len(uci) < 4 {
		return fmt.Errorf("bad move %q", uci)
	}
	from, err := ParseSquare(uci[0:2])
	if err != nil {
		return err
	}
	to, err := ParseSquare(uci[2:4])
	if err != nil {
		return err
	}
	p := b.sq[from]
	if p == Empty {
		return fmt.Errorf("empty origin %s", uci[0:2])
	}

	// Castling: king moves two files.
	if (p == 'K' || p == 'k') && abs(from%8-to%8) == 2 {
		b.sq[to] = p
		b.sq[from] = Empty
		if to%8 == 6 { // kingside
			rookFrom, rookTo := from+3, from+1
			b.sq[rookTo] = b.sq[rookFrom]
			b.sq[rookFrom] = Empty
		} else if to%8 == 2 { // queenside
			rookFrom, rookTo := from-4, from-1
			b.sq[rookTo] = b.sq[rookFrom]
			b.sq[rookFrom] = Empty
		}
		return nil
	}

	// En passant: pawn moves diagonally onto empty square.
	if (p == 'P' || p == 'p') && from%8 != to%8 && b.sq[to] == Empty {
		capSq := to - 8
		if p == 'p' {
			capSq = to + 8
		}
		b.sq[capSq] = Empty
	}

	b.sq[to] = p
	b.sq[from] = Empty

	if len(uci) >= 5 {
		promo := unicode.ToLower(rune(uci[4]))
		if p == 'P' {
			promo = unicode.ToUpper(promo)
		}
		b.sq[to] = Piece(promo)
	}
	return nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

var unicodeGlyphs = map[Piece]string{
	'K': "♔", 'Q': "♕", 'R': "♖", 'B': "♗", 'N': "♘", 'P': "♙",
	'k': "♚", 'q': "♛", 'r': "♜", 'b': "♝", 'n': "♞", 'p': "♟",
}

// pieceArt3 holds 3-row ASCII art for each piece type (case-insensitive).
// Each entry is [top, middle, bottom]. Width is kept to 3 characters.
var pieceArt3 = map[byte][3]string{
	'K': {" + ", ".|.", "/_\\"},
	'Q': {".^.", ")|(", "/_\\"},
	'R': {"[_]", "|#|", "/_\\"},
	'B': {" o ", "(>)", "/_\\"},
	'N': {" ~ ", "(> ", "/_\\"},
	'P': {" _ ", "( )", "/_\\"},
}

// PieceArt returns a multi-line representation of a piece sized to fit
// the given cell dimensions. For small cells it returns the Unicode glyph
// centered. For CellH >= 3, it returns 3-line ASCII art.
func PieceArt(p Piece, cellW, cellH int, useUnicode bool) []string {
	if p == Empty {
		blank := strings.Repeat(" ", cellW)
		out := make([]string, cellH)
		for i := range out {
			out[i] = blank
		}
		return out
	}

	// For tall cells, use ASCII art
	if cellH >= 3 && cellW >= 3 {
		key := byte(unicode.ToUpper(rune(p)))
		art, ok := pieceArt3[key]
		if ok {
			out := make([]string, cellH)
			blank := strings.Repeat(" ", cellW)
			startRow := (cellH - 3) / 2
			for i := range out {
				out[i] = blank
			}
			for j := 0; j < 3; j++ {
				out[startRow+j] = padCenterStr(art[j], cellW)
			}
			return out
		}
	}

	// Fallback: single glyph centered
	glyph := string(p)
	if useUnicode {
		if g, ok := unicodeGlyphs[p]; ok {
			glyph = g
		}
	}
	out := make([]string, cellH)
	blank := strings.Repeat(" ", cellW)
	for i := range out {
		if i == cellH/2 {
			out[i] = padCenterStr(glyph, cellW)
		} else {
			out[i] = blank
		}
	}
	return out
}

func padCenterStr(s string, w int) string {
	n := len([]rune(s))
	if n >= w {
		return s
	}
	left := (w - n) / 2
	right := w - n - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// Render returns an 8-line Unicode board (rank 8 at top) with file labels.
func (b Board) Render() string {
	return b.RenderMode(true)
}

// RenderMode returns the board using Unicode glyphs when unicode is true,
// otherwise ASCII letters. Each square occupies two columns (glyph + pad).
func (b Board) RenderMode(unicode bool) string {
	var sb strings.Builder
	for rank := 7; rank >= 0; rank-- {
		sb.WriteByte(byte('1' + rank))
		sb.WriteByte(' ')
		for file := 0; file < 8; file++ {
			sb.WriteString(b.cellGlyph(rank, file, unicode))
			sb.WriteByte(' ')
		}
		sb.WriteByte('\n')
	}
	sb.WriteString("  a b c d e f g h")
	return sb.String()
}

// GlyphAt returns the display glyph for a named square.
func (b Board) GlyphAt(name string, unicode bool) string {
	sq, err := ParseSquare(name)
	if err != nil {
		return "?"
	}
	return b.cellGlyph(sq/8, sq%8, unicode)
}

func (b Board) cellGlyph(rank, file int, unicode bool) string {
	p := b.sq[rank*8+file]
	if p == Empty {
		return " "
	}
	if unicode {
		if g, ok := unicodeGlyphs[p]; ok {
			return g
		}
	}
	return string(p)
}
