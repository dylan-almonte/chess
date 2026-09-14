package info

import (
	"strconv"
	"strings"
)

// Telemetry holds the latest parsed engine info fields.
type Telemetry struct {
	Idle  bool
	Depth string
	Score string // e.g. "cp 12" or "mate 3"
	Nodes string
	NPS   string
	PV    string
	Raw   string
}

func Idle() Telemetry {
	return Telemetry{Idle: true}
}

// ParseInfo parses a UCI info line. Returns ok=false if not an info line.
func ParseInfo(line string) (Telemetry, bool) {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) == 0 || fields[0] != "info" {
		return Telemetry{}, false
	}
	t := Telemetry{Idle: false, Raw: strings.TrimSpace(line)}
	for i := 1; i < len(fields); i++ {
		switch fields[i] {
		case "depth":
			if i+1 < len(fields) {
				t.Depth = fields[i+1]
				i++
			}
		case "score":
			if i+1 < len(fields) {
				kind := fields[i+1]
				i++
				if i+1 < len(fields) && (kind == "cp" || kind == "mate") {
					t.Score = kind + " " + fields[i+1]
					i++
				} else {
					t.Score = kind
				}
			}
		case "nodes":
			if i+1 < len(fields) {
				t.Nodes = fields[i+1]
				i++
			}
		case "nps":
			if i+1 < len(fields) {
				t.NPS = fields[i+1]
				i++
			}
		case "pv":
			if i+1 < len(fields) {
				t.PV = strings.Join(fields[i+1:], " ")
				i = len(fields)
			}
		default:
			// skip unknown tokens; if looks like a bare number after unknown key, skip value
			if i+1 < len(fields) && looksNumeric(fields[i+1]) {
				i++
			}
		}
	}
	return t, true
}

func looksNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func (t Telemetry) Render() string {
	if t.Idle {
		return "(idle — no info yet)"
	}
	var parts []string
	if t.Depth != "" {
		parts = append(parts, "depth "+t.Depth)
	}
	if t.Score != "" {
		parts = append(parts, "score "+t.Score)
	}
	if t.Nodes != "" {
		parts = append(parts, "nodes "+t.Nodes)
	}
	if t.NPS != "" {
		parts = append(parts, "nps "+t.NPS)
	}
	if t.PV != "" {
		parts = append(parts, "pv "+t.PV)
	}
	if len(parts) == 0 {
		return t.Raw
	}
	return strings.Join(parts, "  |  ")
}
