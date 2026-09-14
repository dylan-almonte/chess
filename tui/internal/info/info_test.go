package info_test

import (
	"testing"

	"github.com/dylanca/chess-tui/tui/internal/info"
)

func TestParseInfo(t *testing.T) {
	tel, ok := info.ParseInfo("info depth 1 score cp 12 pv e2e4")
	if !ok {
		t.Fatal("expected ok")
	}
	if tel.Depth != "1" {
		t.Fatalf("depth=%q", tel.Depth)
	}
	if tel.Score != "cp 12" {
		t.Fatalf("score=%q", tel.Score)
	}
	if tel.PV != "e2e4" {
		t.Fatalf("pv=%q", tel.PV)
	}
	r := tel.Render()
	if !contains(r, "depth 1") || !contains(r, "score cp 12") {
		t.Fatalf("render=%q", r)
	}
}

func TestIdle(t *testing.T) {
	tel := info.Idle()
	if !tel.Idle {
		t.Fatal("expected idle")
	}
	if tel.Render() == "" {
		t.Fatal("idle render should be non-empty")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(len(s) > 0 && (indexOf(s, sub) >= 0)))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
