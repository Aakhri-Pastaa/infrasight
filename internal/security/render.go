package security

import (
	"fmt"
	"io"
)

type palette struct{ reset, dim, yellow, red, bold string }

func newPalette(color bool) palette {
	if !color {
		return palette{}
	}
	return palette{reset: "\033[0m", dim: "\033[2m", yellow: "\033[33m", red: "\033[31m", bold: "\033[1m"}
}

func sevColor(s Severity, p palette) string {
	switch s {
	case SevCritical, SevHigh:
		return p.red
	case SevMedium:
		return p.yellow
	default:
		return p.dim
	}
}

// PrintFindings writes the security findings to w.
func PrintFindings(w io.Writer, findings []Finding, color bool) {
	p := newPalette(color)
	if len(findings) == 0 {
		fmt.Fprintf(w, "%ssecurity:%s no findings\n", p.bold, p.reset)
		return
	}
	fmt.Fprintf(w, "%ssecurity findings (%d):%s\n", p.bold, len(findings), p.reset)
	for _, f := range findings {
		fmt.Fprintf(w, "  %s%-8s%s %s — %s\n",
			sevColor(f.Severity, p), f.Severity, p.reset, f.Title, f.Detail)
	}
}
