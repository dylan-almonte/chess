package render_test

import (
	"testing"

	"github.com/dylanca/chess-tui/tui/internal/board"
	"github.com/dylanca/chess-tui/tui/internal/render"
)

func TestBoardImageSize(t *testing.T) {
	img := render.BoardImage(render.Options{
		Board:    board.StartPos(),
		TileSize: 60,
	})
	b := img.Bounds()
	if b.Dx() != 480 || b.Dy() != 480 {
		t.Fatalf("expected 480x480, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestBoardImageWithHighlights(t *testing.T) {
	img := render.BoardImage(render.Options{
		Board:          board.StartPos(),
		SelectedSquare: "e2",
		CursorSquare:   "e4",
		TileSize:       40,
	})
	b := img.Bounds()
	if b.Dx() != 320 || b.Dy() != 320 {
		t.Fatalf("expected 320x320, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestBoardImageDefaultTileSize(t *testing.T) {
	img := render.BoardImage(render.Options{
		Board: board.StartPos(),
	})
	b := img.Bounds()
	if b.Dx() != 480 {
		t.Fatalf("expected default 480px, got %d", b.Dx())
	}
}
