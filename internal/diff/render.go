package diff

import (
	"fmt"
	"io"

	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
	"github.com/Aakhri-Pastaa/infrasight/internal/output"
)

type palette struct{ reset, dim, green, yellow, red, bold string }

func newPalette(color bool) palette {
	if !color {
		return palette{}
	}
	return palette{
		reset: "\033[0m", dim: "\033[2m", green: "\033[32m",
		yellow: "\033[33m", red: "\033[31m", bold: "\033[1m",
	}
}

// RenderTerminal writes a human-readable drift report.
func RenderTerminal(w io.Writer, baseline, current output.Document, r Result, color bool) {
	p := newPalette(color)

	fmt.Fprintf(w, "%sInfraSight drift%s\n", p.bold, p.reset)
	fmt.Fprintf(w, "  baseline: %s  (%s)\n", baseline.Scan.ID, baseline.Scan.StartedAt)
	fmt.Fprintf(w, "  current : %s  (%s)\n\n", current.Scan.ID, current.Scan.StartedAt)

	if r.Empty() {
		fmt.Fprintf(w, "  %sno drift — the two scans are identical%s\n", p.green, p.reset)
		return
	}

	if len(r.Added) > 0 {
		fmt.Fprintf(w, "%s+ ADDED (%d)%s\n", p.green, len(r.Added), p.reset)
		for _, n := range r.Added {
			fmt.Fprintf(w, "  %s+%s %s%-11s%s %s%s\n", p.green, p.reset, p.dim, n.Type, p.reset, n.Label, version(n))
		}
		fmt.Fprintln(w)
	}
	if len(r.Removed) > 0 {
		fmt.Fprintf(w, "%s- REMOVED (%d)%s\n", p.red, len(r.Removed), p.reset)
		for _, n := range r.Removed {
			fmt.Fprintf(w, "  %s-%s %s%-11s%s %s%s\n", p.red, p.reset, p.dim, n.Type, p.reset, n.Label, version(n))
		}
		fmt.Fprintln(w)
	}
	if len(r.Versions) > 0 {
		fmt.Fprintf(w, "%s~ VERSION CHANGES (%d)%s\n", p.yellow, len(r.Versions), p.reset)
		for _, v := range r.Versions {
			fmt.Fprintf(w, "  %s~%s %s%-11s%s %s  %s%s → %s%s\n",
				p.yellow, p.reset, p.dim, v.Type, p.reset, v.Label, p.dim, orNone(v.From), orNone(v.To), p.reset)
		}
		fmt.Fprintln(w)
	}
	if len(r.Health) > 0 {
		fmt.Fprintf(w, "%s~ HEALTH CHANGES (%d)%s\n", p.yellow, len(r.Health), p.reset)
		for _, h := range r.Health {
			arrow := healthColor(h.To, p)
			regressed := ""
			if Severity(h.To) > Severity(h.From) {
				regressed = p.red + "  (regression)" + p.reset
			} else if Severity(h.To) < Severity(h.From) {
				regressed = p.green + "  (recovered)" + p.reset
			}
			fmt.Fprintf(w, "  %s~%s %s%-11s%s %s  %s → %s%s%s%s\n",
				p.yellow, p.reset, p.dim, h.Type, p.reset, h.Label,
				orHealth(h.From), arrow, string(orHealth(h.To)), p.reset, regressed)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "%ssummary:%s +%d added, -%d removed, %d version, %d health\n",
		p.bold, p.reset, len(r.Added), len(r.Removed), len(r.Versions), len(r.Health))
}

func version(n graph.Node) string {
	if n.Version == "" {
		return ""
	}
	return " " + n.Version
}

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

func orHealth(h graph.Health) graph.Health {
	if h == graph.HealthUnknown {
		return "unknown"
	}
	return h
}

func healthColor(h graph.Health, p palette) string {
	switch h {
	case graph.HealthCritical:
		return p.red
	case graph.HealthWarning:
		return p.yellow
	case graph.HealthHealthy:
		return p.green
	default:
		return ""
	}
}
