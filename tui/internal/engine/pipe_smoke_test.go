package engine_test

import (
	"testing"
	"time"

	"github.com/dylanca/chess-tui/tui/internal/engine"
)

func TestReadLoopAfterSend(t *testing.T) {
	c, err := engine.StartFake(engine.StubHandler)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan any, 8)
	c.ReadLoop(func(v any) { ch <- v })
	if err := c.Send("uci"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		select {
		case m := <-ch:
			t.Logf("got %#v", m)
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout waiting for line %d", i)
		}
	}
}
