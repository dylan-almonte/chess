package board_test

import (
	"testing"

	"github.com/dylanca/chess-tui/tui/internal/board"
)

func TestStartPosAndE2E4(t *testing.T) {
	b := board.StartPos()
	if b.PieceAtName("e2") != 'P' {
		t.Fatalf("e2 want P got %c", b.PieceAtName("e2"))
	}
	if err := b.ApplyUCI("e2e4"); err != nil {
		t.Fatal(err)
	}
	if b.PieceAtName("e4") != 'P' {
		t.Fatalf("e4 want P got %c", b.PieceAtName("e4"))
	}
	if b.PieceAtName("e2") != board.Empty {
		t.Fatal("e2 should be empty")
	}
}
