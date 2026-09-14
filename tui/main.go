package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dylanca/chess-tui/tui/internal/path"
	"github.com/dylanca/chess-tui/tui/internal/ui"
)

func main() {
	base, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cwd: %v\n", err)
		os.Exit(1)
	}
	enginePath := path.Resolve(base)
	// Also try resolving relative to this source file's module when cwd is repo root.
	if err := path.Validate(enginePath); err != nil {
		alt := path.Resolve(filepath.Join(base, "tui"))
		if path.Validate(alt) == nil {
			enginePath = alt
		}
	}

	m := ui.New(ui.Config{EnginePath: enginePath})
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tui: %v\n", err)
		os.Exit(1)
	}
}
