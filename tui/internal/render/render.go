// Package render draws the chess board as an image.
package render

import (
	"image"
	"image/color"
	"math"

	"github.com/dylanca/chess-tui/tui/internal/board"
)

var (
	lightTile    = color.RGBA{240, 217, 181, 255} // #f0d9b5
	darkTile     = color.RGBA{181, 136, 99, 255}  // #b58863
	cursorTile   = color.RGBA{127, 201, 127, 255} // #7fc97f
	selectedTile = color.RGBA{255, 255, 102, 255} // #ffff66

	whitePiece  = color.RGBA{255, 255, 255, 255}
	whiteBorder = color.RGBA{60, 60, 60, 255}
	blackPiece  = color.RGBA{40, 40, 40, 255}
	blackBorder = color.RGBA{200, 200, 200, 255}
)

// Options controls board image generation.
type Options struct {
	Board          board.Board
	SelectedSquare string
	CursorSquare   string
	TileSize       int // pixels per square (default 60)
}

// BoardImage generates an image of the chess board.
func BoardImage(opts Options) image.Image {
	ts := opts.TileSize
	if ts <= 0 {
		ts = 60
	}
	size := ts * 8
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			x0 := file * ts
			y0 := (7 - rank) * ts
			name := board.SquareName(rank*8 + file)

			// Tile color
			tc := darkTile
			if (rank+file)%2 == 1 {
				tc = lightTile
			}
			if name == opts.SelectedSquare {
				tc = selectedTile
			} else if name == opts.CursorSquare {
				tc = cursorTile
			}
			fillRect(img, x0, y0, ts, ts, tc)

			// Piece
			p := opts.Board.PieceAtName(name)
			if p != board.Empty {
				drawPiece(img, p, x0, y0, ts)
			}
		}
	}
	return img
}

func drawPiece(img *image.RGBA, p board.Piece, x0, y0, ts int) {
	fill, border := whitePiece, whiteBorder
	if p >= 'a' && p <= 'z' {
		fill, border = blackPiece, blackBorder
	}

	cx := float64(x0) + float64(ts)/2
	cy := float64(y0) + float64(ts)/2
	margin := float64(ts) * 0.1
	s := float64(ts) - margin*2 // usable area

	switch upper(p) {
	case 'P':
		drawPawn(img, cx, cy, s, fill, border)
	case 'R':
		drawRook(img, cx, cy, s, fill, border)
	case 'B':
		drawBishop(img, cx, cy, s, fill, border)
	case 'N':
		drawKnight(img, cx, cy, s, fill, border)
	case 'Q':
		drawQueen(img, cx, cy, s, fill, border)
	case 'K':
		drawKing(img, cx, cy, s, fill, border)
	}
}

func upper(p board.Piece) byte {
	if p >= 'a' && p <= 'z' {
		return byte(p) - 32
	}
	return byte(p)
}

// --- Piece drawing functions ---

func drawPawn(img *image.RGBA, cx, cy, s float64, fill, border color.RGBA) {
	r := s * 0.25
	fillCircle(img, cx, cy-s*0.05, r, fill)
	strokeCircle(img, cx, cy-s*0.05, r, border)
	// Base
	bw := s * 0.45
	bh := s * 0.15
	fillRect(img, int(cx-bw/2), int(cy+s*0.25), int(bw), int(bh), fill)
	strokeRect(img, int(cx-bw/2), int(cy+s*0.25), int(bw), int(bh), border)
}

func drawRook(img *image.RGBA, cx, cy, s float64, fill, border color.RGBA) {
	// Body
	bw := s * 0.5
	bh := s * 0.55
	bx := cx - bw/2
	by := cy - bh/2 + s*0.05
	fillRect(img, int(bx), int(by), int(bw), int(bh), fill)
	strokeRect(img, int(bx), int(by), int(bw), int(bh), border)
	// Crenellations: 3 small rectangles on top
	cw := bw / 5
	ch := s * 0.12
	for i := 0; i < 3; i++ {
		cx2 := bx + float64(i)*cw*2
		fillRect(img, int(cx2), int(by-ch), int(cw), int(ch), fill)
		strokeRect(img, int(cx2), int(by-ch), int(cw), int(ch), border)
	}
	// Base
	basew := s * 0.6
	baseh := s * 0.1
	fillRect(img, int(cx-basew/2), int(by+bh), int(basew), int(baseh), fill)
	strokeRect(img, int(cx-basew/2), int(by+bh), int(basew), int(baseh), border)
}

func drawBishop(img *image.RGBA, cx, cy, s float64, fill, border color.RGBA) {
	// Tall triangle body
	top := cy - s*0.35
	bot := cy + s*0.3
	half := s * 0.22
	fillTriangle(img, cx, top, cx-half, bot, cx+half, bot, fill)
	strokeTriangle(img, cx, top, cx-half, bot, cx+half, bot, border)
	// Circle top
	fillCircle(img, cx, top-s*0.04, s*0.07, fill)
	strokeCircle(img, cx, top-s*0.04, s*0.07, border)
	// Base
	bw := s * 0.5
	bh := s * 0.1
	fillRect(img, int(cx-bw/2), int(bot), int(bw), int(bh), fill)
	strokeRect(img, int(cx-bw/2), int(bot), int(bw), int(bh), border)
}

func drawKnight(img *image.RGBA, cx, cy, s float64, fill, border color.RGBA) {
	// Simplified knight as an L-shaped polygon
	pts := [][2]float64{
		{cx - s*0.15, cy + s*0.35},  // bottom-left
		{cx - s*0.15, cy - s*0.1},   // neck
		{cx - s*0.25, cy - s*0.2},   // ear back
		{cx - s*0.1, cy - s*0.35},   // ear top
		{cx + s*0.05, cy - s*0.25},  // forehead
		{cx + s*0.25, cy - s*0.15},  // nose
		{cx + s*0.15, cy - s*0.05},  // chin
		{cx + s*0.15, cy + s*0.35},  // bottom-right
	}
	fillPolygon(img, pts, fill)
	strokePolygon(img, pts, border)
	// Base
	bw := s * 0.55
	bh := s * 0.1
	fillRect(img, int(cx-bw/2), int(cy+s*0.35), int(bw), int(bh), fill)
	strokeRect(img, int(cx-bw/2), int(cy+s*0.35), int(bw), int(bh), border)
}

func drawQueen(img *image.RGBA, cx, cy, s float64, fill, border color.RGBA) {
	// Body (trapezoid)
	topW := s * 0.3
	botW := s * 0.5
	top := cy - s*0.1
	bot := cy + s*0.3
	pts := [][2]float64{
		{cx - topW/2, top},
		{cx + topW/2, top},
		{cx + botW/2, bot},
		{cx - botW/2, bot},
	}
	fillPolygon(img, pts, fill)
	strokePolygon(img, pts, border)
	// Crown: 3 small circles on top
	for _, dx := range []float64{-0.15, 0, 0.15} {
		fillCircle(img, cx+s*dx, top-s*0.08, s*0.06, fill)
		strokeCircle(img, cx+s*dx, top-s*0.08, s*0.06, border)
	}
	// Big circle on center tip
	fillCircle(img, cx, top-s*0.16, s*0.05, fill)
	strokeCircle(img, cx, top-s*0.16, s*0.05, border)
	// Base
	bw := s * 0.55
	bh := s * 0.1
	fillRect(img, int(cx-bw/2), int(bot), int(bw), int(bh), fill)
	strokeRect(img, int(cx-bw/2), int(bot), int(bw), int(bh), border)
}

func drawKing(img *image.RGBA, cx, cy, s float64, fill, border color.RGBA) {
	// Body (trapezoid, same as queen but wider)
	topW := s * 0.35
	botW := s * 0.5
	top := cy - s*0.05
	bot := cy + s*0.3
	pts := [][2]float64{
		{cx - topW/2, top},
		{cx + topW/2, top},
		{cx + botW/2, bot},
		{cx - botW/2, bot},
	}
	fillPolygon(img, pts, fill)
	strokePolygon(img, pts, border)
	// Cross on top
	cw := s * 0.07
	ch := s * 0.22
	fillRect(img, int(cx-cw/2), int(top-ch), int(cw), int(ch), fill)
	strokeRect(img, int(cx-cw/2), int(top-ch), int(cw), int(ch), border)
	armw := s * 0.18
	armh := s * 0.07
	fillRect(img, int(cx-armw/2), int(top-ch*0.7), int(armw), int(armh), fill)
	strokeRect(img, int(cx-armw/2), int(top-ch*0.7), int(armw), int(armh), border)
	// Base
	bw := s * 0.55
	bh := s * 0.1
	fillRect(img, int(cx-bw/2), int(bot), int(bw), int(bh), fill)
	strokeRect(img, int(cx-bw/2), int(bot), int(bw), int(bh), border)
}

// --- Drawing primitives ---

func fillRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			img.SetRGBA(x+dx, y+dy, c)
		}
	}
}

func strokeRect(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	for dx := 0; dx < w; dx++ {
		img.SetRGBA(x+dx, y, c)
		img.SetRGBA(x+dx, y+h-1, c)
	}
	for dy := 0; dy < h; dy++ {
		img.SetRGBA(x, y+dy, c)
		img.SetRGBA(x+w-1, y+dy, c)
	}
}

func fillCircle(img *image.RGBA, cx, cy, r float64, c color.RGBA) {
	r2 := r * r
	for y := int(cy - r); y <= int(cy+r); y++ {
		for x := int(cx - r); x <= int(cx+r); x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			if dx*dx+dy*dy <= r2 {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func strokeCircle(img *image.RGBA, cx, cy, r float64, c color.RGBA) {
	steps := int(2 * math.Pi * r)
	if steps < 60 {
		steps = 60
	}
	for i := 0; i < steps; i++ {
		t := 2 * math.Pi * float64(i) / float64(steps)
		x := int(cx + r*math.Cos(t))
		y := int(cy + r*math.Sin(t))
		img.SetRGBA(x, y, c)
	}
}

func fillTriangle(img *image.RGBA, x1, y1, x2, y2, x3, y3 float64, c color.RGBA) {
	minY := int(math.Min(y1, math.Min(y2, y3)))
	maxY := int(math.Max(y1, math.Max(y2, y3)))
	minX := int(math.Min(x1, math.Min(x2, x3)))
	maxX := int(math.Max(x1, math.Max(x2, x3)))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if pointInTriangle(float64(x), float64(y), x1, y1, x2, y2, x3, y3) {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func strokeTriangle(img *image.RGBA, x1, y1, x2, y2, x3, y3 float64, c color.RGBA) {
	drawLine(img, x1, y1, x2, y2, c)
	drawLine(img, x2, y2, x3, y3, c)
	drawLine(img, x3, y3, x1, y1, c)
}

func pointInTriangle(px, py, x1, y1, x2, y2, x3, y3 float64) bool {
	d1 := sign(px, py, x1, y1, x2, y2)
	d2 := sign(px, py, x2, y2, x3, y3)
	d3 := sign(px, py, x3, y3, x1, y1)
	hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)
	return !(hasNeg && hasPos)
}

func sign(px, py, x1, y1, x2, y2 float64) float64 {
	return (px-x2)*(y1-y2) - (x1-x2)*(py-y2)
}

func drawLine(img *image.RGBA, x1, y1, x2, y2 float64, c color.RGBA) {
	dist := math.Hypot(x2-x1, y2-y1)
	steps := int(dist) * 2
	if steps < 2 {
		steps = 2
	}
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := x1 + t*(x2-x1)
		y := y1 + t*(y2-y1)
		img.SetRGBA(int(x), int(y), c)
	}
}

func fillPolygon(img *image.RGBA, pts [][2]float64, c color.RGBA) {
	if len(pts) < 3 {
		return
	}
	minX, minY := pts[0][0], pts[0][1]
	maxX, maxY := pts[0][0], pts[0][1]
	for _, p := range pts {
		if p[0] < minX {
			minX = p[0]
		}
		if p[0] > maxX {
			maxX = p[0]
		}
		if p[1] < minY {
			minY = p[1]
		}
		if p[1] > maxY {
			maxY = p[1]
		}
	}
	for y := int(minY); y <= int(maxY); y++ {
		for x := int(minX); x <= int(maxX); x++ {
			if pointInPolygon(float64(x), float64(y), pts) {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func pointInPolygon(x, y float64, pts [][2]float64) bool {
	n := len(pts)
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		yi, xi := pts[i][1], pts[i][0]
		yj, xj := pts[j][1], pts[j][0]
		if ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}
	return inside
}

func strokePolygon(img *image.RGBA, pts [][2]float64, c color.RGBA) {
	n := len(pts)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		drawLine(img, pts[i][0], pts[i][1], pts[j][0], pts[j][1], c)
	}
}
