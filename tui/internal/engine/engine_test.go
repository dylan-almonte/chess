package engine_test

import (
	"testing"

	"github.com/dylanca/chess-tui/tui/internal/engine"
)

func TestHandshakeRecordsUciAndIsReady(t *testing.T) {
	c, err := engine.StartFake(engine.StubHandler)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Quit() }()

	if err := c.Handshake(); err != nil {
		t.Fatal(err)
	}

	log := c.Log()
	assertLogHas(t, log, engine.Outbound, "uci")
	assertLogHas(t, log, engine.Inbound, "uciok")
	assertLogHas(t, log, engine.Outbound, "isready")
	assertLogHas(t, log, engine.Inbound, "readyok")
}

func TestStatefulHandlerTracksPosition(t *testing.T) {
	h := engine.NewStatefulHandler()
	c, err := engine.StartFake(h.Handle)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Quit() }()

	if err := c.Handshake(); err != nil {
		t.Fatal(err)
	}

	// Send position startpos
	if err := c.Send("position startpos"); err != nil {
		t.Fatal(err)
	}

	// legalmoves should return the startpos set
	if err := c.Send("legalmoves"); err != nil {
		t.Fatal(err)
	}
	log := c.Log()
	found := false
	for _, e := range log {
		if e.Dir == engine.Inbound && e.Text != "" && len(e.Text) > len("legalmoves") {
			if e.Text[:10] == "legalmoves" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("expected legalmoves response with moves")
	}
}

func TestStatefulHandlerConfigurableBestmove(t *testing.T) {
	h := engine.NewStatefulHandler()
	h.SetBestmove("e2e5") // intentionally illegal
	c, err := engine.StartFake(h.Handle)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Quit() }()

	if err := c.Handshake(); err != nil {
		t.Fatal(err)
	}

	if err := c.Send("position startpos"); err != nil {
		t.Fatal(err)
	}
	if err := c.Send("go"); err != nil {
		t.Fatal(err)
	}
	log := c.Log()
	assertLogHas(t, log, engine.Inbound, "bestmove e2e5")
}

func TestStatefulHandlerPositionWithMoves(t *testing.T) {
	h := engine.NewStatefulHandler()
	// Register legal moves for the position after e2e4
	h.SetLegalMoves("e2e4", []string{"e7e5", "d7d5", "c7c5"})
	c, err := engine.StartFake(h.Handle)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.Quit() }()

	if err := c.Handshake(); err != nil {
		t.Fatal(err)
	}

	if err := c.Send("position startpos moves e2e4"); err != nil {
		t.Fatal(err)
	}
	if err := c.Send("legalmoves"); err != nil {
		t.Fatal(err)
	}
	log := c.Log()
	assertLogHas(t, log, engine.Inbound, "legalmoves e7e5 d7d5 c7c5")
}

func assertLogHas(t *testing.T, log []engine.LogEntry, dir engine.Direction, text string) {
	t.Helper()
	for _, e := range log {
		if e.Dir == dir && e.Text == text {
			return
		}
	}
	t.Fatalf("log missing %s %q; got %#v", dir, text, log)
}
