package ui_test

import (
	"testing"

	"github.com/dylanca/chess-tui/tui/internal/ui"
)

func TestSquareAtCellMapsBoardOrigin(t *testing.T) {
	if got, ok := ui.SquareAtCell(cellFor("e2")); !ok || got != "e2" {
		t.Fatalf("e2: got %q ok=%v", got, ok)
	}
	if got, ok := ui.SquareAtCell(cellFor("e4")); !ok || got != "e4" {
		t.Fatalf("e4: got %q ok=%v", got, ok)
	}
	if got, ok := ui.SquareAtCell(0, 0); ok {
		t.Fatalf("border/padding should miss, got %q", got)
	}
	if got, ok := ui.SquareAtCell(2, 1); ok {
		t.Fatalf("rank label should miss, got %q", got)
	}
	if got, ok := ui.SquareAtCell(12, 9); ok {
		t.Fatalf("file labels should miss, got %q", got)
	}
	if got, ok := ui.SquareAtCell(100, 100); ok {
		t.Fatalf("off-board should miss, got %q", got)
	}
}

func cellFor(name string) (int, int) {
	file := int(name[0] - 'a')
	rank := int(name[1] - '1')
	return 4 + file*2, 1 + (7 - rank)
}
