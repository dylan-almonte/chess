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

// Render returns an 8-line board string (rank 8 at top) with file labels.
func (b Board) Render() string {
	var sb strings.Builder
	for rank := 7; rank >= 0; rank-- {
		sb.WriteByte(byte('1' + rank))
		sb.WriteByte(' ')
		for file := 0; file < 8; file++ {
			p := b.sq[rank*8+file]
			if p == Empty {
				if (rank+file)%2 == 0 {
					sb.WriteByte('.')
				} else {
					sb.WriteByte(',')
				}
			} else {
				sb.WriteByte(byte(p))
			}
			sb.WriteByte(' ')
		}
		sb.WriteByte('\n')
	}
	sb.WriteString("  a b c d e f g h")
	return sb.String()
}
