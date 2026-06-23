package output

import (
	"fmt"
	"io"
	"sort"
)

// ANSI colour helpers (disabled when noColor is set).
type palette struct{ reset, dim, green, yellow, red, bold string }

func newPalette(color bool) palette {
	if !color {
		return palette{}
	}
	return palette{
		reset:  "\033[0m",
		dim:    "\033[2m",
		green:  "\033[32m",
		yellow: "\033[33m",
		red:    "\033[31m",
		bold:   "\033[1m",
	}
}

// PrintSummary writes a concise human summary of a scan to w.
func PrintSummary(w io.Writer, doc Document, skipped []string, errs map[string]error, color bool) {
	p := newPalette(color)
	h := doc.Summary.Health

	fmt.Fprintf(w, "%s%sInfraSight%s  host=%s  modules=%d  %dms\n",
		p.bold, "", p.reset, doc.Scan.Hostname, len(doc.Scan.Modules), doc.Scan.DurationMs)

	fmt.Fprintf(w, "  nodes=%d edges=%d  [%s%d healthy%s  %s%d warning%s  %s%d critical%s]\n",
		doc.Summary.TotalNodes, doc.Summary.TotalEdges,
		p.green, h["healthy"], p.reset,
		p.yellow, h["warning"], p.reset,
		p.red, h["critical"], p.reset)

	// nodes by type, sorted
	types := make([]string, 0, len(doc.Summary.NodesByType))
	for t := range doc.Summary.NodesByType {
		types = append(types, t)
	}
	sort.Strings(types)
	for _, t := range types {
		fmt.Fprintf(w, "    %s%-14s%s %d\n", p.dim, t, p.reset, doc.Summary.NodesByType[t])
	}

	if len(skipped) > 0 {
		fmt.Fprintf(w, "  %sskipped (unavailable): %v%s\n", p.dim, skipped, p.reset)
	}
	for name, err := range errs {
		fmt.Fprintf(w, "  %s! %s: %v%s\n", p.yellow, name, err, p.reset)
	}

	for _, c := range doc.Summary.Criticals {
		fmt.Fprintf(w, "  %scritical:%s %s\n", p.red, p.reset, c)
	}
}
