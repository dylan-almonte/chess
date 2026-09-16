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

func TestKingsideCastlingDisplay(t *testing.T) {
	b := board.StartPos()
	// Clear path: remove knight and bishop
	b.ClearSquare("f1")
	b.ClearSquare("g1")
	if err := b.ApplyUCI("e1g1"); err != nil {
		t.Fatalf("kingside castling failed: %v", err)
	}
	if b.PieceAtName("g1") != 'K' {
		t.Fatalf("expected king on g1, got %c", b.PieceAtName("g1"))
	}
	if b.PieceAtName("f1") != 'R' {
		t.Fatalf("expected rook on f1, got %c", b.PieceAtName("f1"))
	}
	if b.PieceAtName("e1") != board.Empty {
		t.Fatal("e1 should be empty after castling")
	}
	if b.PieceAtName("h1") != board.Empty {
		t.Fatal("h1 should be empty after castling")
	}
}

func TestQueensideCastlingDisplay(t *testing.T) {
	b := board.StartPos()
	// Clear path: remove queen, bishop, knight
	b.ClearSquare("d1")
	b.ClearSquare("c1")
	b.ClearSquare("b1")
	if err := b.ApplyUCI("e1c1"); err != nil {
		t.Fatalf("queenside castling failed: %v", err)
	}
	if b.PieceAtName("c1") != 'K' {
		t.Fatalf("expected king on c1, got %c", b.PieceAtName("c1"))
	}
	if b.PieceAtName("d1") != 'R' {
		t.Fatalf("expected rook on d1, got %c", b.PieceAtName("d1"))
	}
	if b.PieceAtName("e1") != board.Empty {
		t.Fatal("e1 should be empty")
	}
	if b.PieceAtName("a1") != board.Empty {
		t.Fatal("a1 should be empty")
	}
}

func TestEnPassantDisplay(t *testing.T) {
	b := board.StartPos()
	// Setup: white pawn on e5, black pawn on d5 (just arrived via d7d5)
	_ = b.ApplyUCI("e2e4")
	_ = b.ApplyUCI("e4e5") // manually advance
	_ = b.ApplyUCI("d7d5") // black double pawn push

	// En passant: e5d6
	if err := b.ApplyUCI("e5d6"); err != nil {
		t.Fatalf("en passant failed: %v", err)
	}
	if b.PieceAtName("d6") != 'P' {
		t.Fatalf("expected white pawn on d6, got %c", b.PieceAtName("d6"))
	}
	if b.PieceAtName("e5") != board.Empty {
		t.Fatal("e5 should be empty")
	}
	if b.PieceAtName("d5") != board.Empty {
		t.Fatal("d5 should be empty (captured pawn removed)")
	}
}

func TestPromotionDisplay(t *testing.T) {
	b := board.StartPos()
	// Setup: white pawn on e7, target square e8 empty
	b.ClearSquare("e2")
	b.ClearSquare("e7")    // remove black pawn
	b.SetSquare("e7", 'P') // white pawn on 7th rank

	// Promote to queen
	if err := b.ApplyUCI("e7e8q"); err != nil {
		t.Fatalf("promotion failed: %v", err)
	}
	if b.PieceAtName("e8") != 'Q' {
		t.Fatalf("expected white queen on e8, got %c", b.PieceAtName("e8"))
	}
	if b.PieceAtName("e7") != board.Empty {
		t.Fatal("e7 should be empty after promotion")
	}
}

func TestBlackPromotionDisplay(t *testing.T) {
	b := board.StartPos()
	// Setup: black pawn on e2, target square e1 empty
	b.ClearSquare("e7")
	b.ClearSquare("e2")    // remove white pawn
	b.SetSquare("e2", 'p') // black pawn on 2nd rank

	if err := b.ApplyUCI("e2e1q"); err != nil {
		t.Fatalf("black promotion failed: %v", err)
	}
	if b.PieceAtName("e1") != 'q' {
		t.Fatalf("expected black queen on e1, got %c", b.PieceAtName("e1"))
	}
}

func TestPromotionToKnightDisplay(t *testing.T) {
	b := board.StartPos()
	b.ClearSquare("e7")
	b.SetSquare("e7", 'P')

	if err := b.ApplyUCI("e7e8n"); err != nil {
		t.Fatalf("promotion to knight failed: %v", err)
	}
	if b.PieceAtName("e8") != 'N' {
		t.Fatalf("expected white knight on e8, got %c", b.PieceAtName("e8"))
	}
}

func TestApplyUCIRejectsEmptyOrigin(t *testing.T) {
	b := board.StartPos()
	b.ClearSquare("e4")
	err := b.ApplyUCI("e4e5")
	if err == nil {
		t.Fatal("expected error for move from empty square")
	}
}
