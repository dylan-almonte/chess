package board_test

import (
	"testing"

	"github.com/dylanca/chess-tui/tui/internal/board"
)

func TestFENStartpos(t *testing.T) {
	if got := board.StartPos().FEN(true); got != board.StartFEN {
		t.Fatalf("startpos FEN: got %q", got)
	}
}

func TestFENKingsAndQueen(t *testing.T) {
	var b board.Board
	b.SetSquare("e1", 'K')
	b.SetSquare("d1", 'Q')
	b.SetSquare("e8", 'k')
	want := "4k3/8/8/8/8/8/8/3QK3 w - - 0 1"
	if got := b.FEN(true); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestCountKings(t *testing.T) {
	w, b := board.StartPos().CountKings()
	if w != 1 || b != 1 {
		t.Fatalf("startpos kings %d/%d", w, b)
	}
}
