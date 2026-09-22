package board

import "strings"

const StartFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// CountKings returns the number of white and black kings on the board.
func (b Board) CountKings() (white, black int) {
	for _, p := range b.sq {
		if p == 'K' {
			white++
		}
		if p == 'k' {
			black++
		}
	}
	return white, black
}

// FEN returns a six-field FEN for the display board. sideWhite is the side
// to move. Castling is inferred from kings and rooks still on their home
// squares. En passant is always "-"; clocks are 0 1.
func (b Board) FEN(sideWhite bool) string {
	var parts []string
	for rank := 7; rank >= 0; rank-- {
		empty := 0
		var sb strings.Builder
		for file := 0; file < 8; file++ {
			p := b.sq[rank*8+file]
			if p == Empty {
				empty++
				continue
			}
			if empty > 0 {
				sb.WriteByte(byte('0' + empty))
				empty = 0
			}
			sb.WriteByte(byte(p))
		}
		if empty > 0 {
			sb.WriteByte(byte('0' + empty))
		}
		parts = append(parts, sb.String())
	}
	side := "b"
	if sideWhite {
		side = "w"
	}
	return strings.Join(parts, "/") + " " + side + " " + b.castleRights() + " - 0 1"
}

func (b Board) castleRights() string {
	var s strings.Builder
	if b.PieceAtName("e1") == 'K' && b.PieceAtName("h1") == 'R' {
		s.WriteByte('K')
	}
	if b.PieceAtName("e1") == 'K' && b.PieceAtName("a1") == 'R' {
		s.WriteByte('Q')
	}
	if b.PieceAtName("e8") == 'k' && b.PieceAtName("h8") == 'r' {
		s.WriteByte('k')
	}
	if b.PieceAtName("e8") == 'k' && b.PieceAtName("a8") == 'r' {
		s.WriteByte('q')
	}
	if s.Len() == 0 {
		return "-"
	}
	return s.String()
}

// Clear empties every square.
func (b *Board) Clear() {
	*b = Board{}
}
