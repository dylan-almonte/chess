package ui_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dylanca/chess-tui/tui/internal/engine"
	"github.com/dylanca/chess-tui/tui/internal/ui"
)

// TestRealEngineE2E4SyncAndLegalmoves builds the Rust engine and runs a
// real-process integration test covering human move e2e4, one legal engine
// reply, full-history synchronization, and the next legalmoves response.
func TestRealEngineE2E4SyncAndLegalmoves(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real-engine integration test in short mode")
	}

	// Find repo root (two levels up from tui/internal/ui)
	repoRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	// Build the engine
	buildCmd := exec.Command("cargo", "build", "--release")
	buildCmd.Dir = repoRoot
	buildCmd.Env = append(os.Environ(), "DEVELOPER_DIR=/Library/Developer/CommandLineTools")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("cargo build failed: %v\n%s", err, out)
	}

	enginePath := filepath.Join(repoRoot, "target", "release", "chess")
	if _, err := os.Stat(enginePath); err != nil {
		t.Fatalf("engine binary not found: %v", err)
	}

	// Create model with real engine
	m := ui.New(ui.Config{EnginePath: enginePath})
	cmd := m.Init()

	// Pump to ready state
	um := pump(t, m, cmd, func(u ui.Model) bool {
		return u.PhaseReady()
	}, 100)

	// Submit e2e4
	um.SetInputValue("e2e4")
	model, cmd2 := um.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Pump until we have at least 2 moves (human + engine reply)
	um = pump(t, model, cmd2, func(u ui.Model) bool {
		return len(u.Moves()) >= 2
	}, 200)

	// Verify human move was e2e4
	if um.Moves()[0] != "e2e4" {
		t.Fatalf("expected first move e2e4, got %s", um.Moves()[0])
	}

	// Engine replied with some move
	engineMove := um.Moves()[1]
	if len(engineMove) < 4 {
		t.Fatalf("expected valid engine move, got %q", engineMove)
	}

	// Verify engine was synchronized with full history
	log := um.ClientLog()
	expectedSync := "position startpos moves e2e4 " + engineMove
	foundSync := false
	for _, e := range log {
		if e.Dir == engine.Outbound && e.Text == expectedSync {
			foundSync = true
		}
	}
	if !foundSync {
		t.Fatalf("expected outbound %q in log", expectedSync)
	}

	// Verify legalmoves was requested
	foundLegalmoves := false
	for _, e := range log {
		if e.Dir == engine.Outbound && e.Text == "legalmoves" {
			foundLegalmoves = true
		}
	}
	if !foundLegalmoves {
		var logTexts []string
		for _, e := range log {
			logTexts = append(logTexts, e.Dir.String()+" "+e.Text)
		}
		t.Fatalf("expected outbound 'legalmoves' in log, got:\n%s", strings.Join(logTexts, "\n"))
	}

	// Verify legalmoves response was received
	foundLegalmovesReply := false
	for _, e := range log {
		if e.Dir == engine.Inbound && strings.HasPrefix(e.Text, "legalmoves") {
			foundLegalmovesReply = true
		}
	}
	if !foundLegalmovesReply {
		t.Fatal("expected inbound legalmoves response")
	}
}
