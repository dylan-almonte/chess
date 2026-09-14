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

func assertLogHas(t *testing.T, log []engine.LogEntry, dir engine.Direction, text string) {
	t.Helper()
	for _, e := range log {
		if e.Dir == dir && e.Text == text {
			return
		}
	}
	t.Fatalf("log missing %s %q; got %#v", dir, text, log)
}
