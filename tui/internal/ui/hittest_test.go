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

func TestSquareAtCellUsesScaledMetrics(t *testing.T) {
	scaled := ui.BoardMetrics{OriginX: 2, OriginY: 1, LabelW: 2, CellW: 3, CellH: 2}
	x, y := cellForMetrics("e2", scaled)
	if got, ok := ui.SquareAtCellWith(x, y, scaled); !ok || got != "e2" {
		t.Fatalf("3x2 e2: got %q ok=%v at %d,%d", got, ok, x, y)
	}
	x, y = cellForMetrics("e4", scaled)
	if got, ok := ui.SquareAtCellWith(x, y, scaled); !ok || got != "e4" {
		t.Fatalf("3x2 e4: got %q ok=%v at %d,%d", got, ok, x, y)
	}
	compact := ui.DefaultBoardMetrics()
	x, y = cellForMetrics("e2", compact)
	if got, ok := ui.SquareAtCellWith(x, y, compact); !ok || got != "e2" {
		t.Fatalf("2x1 e2: got %q ok=%v", got, ok)
	}
	x, y = cellForMetrics("e4", compact)
	if got, ok := ui.SquareAtCellWith(x, y, compact); !ok || got != "e4" {
		t.Fatalf("2x1 e4: got %q ok=%v", got, ok)
	}
}

func cellForMetrics(name string, m ui.BoardMetrics) (int, int) {
	file := int(name[0] - 'a')
	rank := int(name[1] - '1')
	return m.OriginX + m.LabelW + file*m.CellW, m.OriginY + (7-rank)*m.CellH
}

func cellFor(name string) (int, int) {
	file := int(name[0] - 'a')
	rank := int(name[1] - '1')
	return 4 + file*2, 1 + (7 - rank)
}
