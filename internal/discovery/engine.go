package discovery

import (
	"context"
	"sort"
	"sync"

	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
)

// ScanOptions tunes a scan run.
type ScanOptions struct {
	// Parallel caps concurrent probe workers. Zero means the default (50).
	Parallel int
}

// ScanReport is the full outcome of a scan: the merged graph plus bookkeeping.
type ScanReport struct {
	Graph      graph.Graph
	ModulesRun []string
	Skipped    []string
	Warnings   []string
	Errors     map[string]error
}

// Engine runs a fixed set of modules.
type Engine struct {
	modules []Module
}

// NewEngine constructs an Engine over the given modules.
func NewEngine(modules []Module) *Engine {
	return &Engine{modules: modules}
}

// Scan runs every available module concurrently (bounded by Parallel), honours
// each module's own timeout, and merges the results into a single graph. A
// module that errors does not abort the scan; its error is recorded and any
// partial result it returned is still merged.
func (e *Engine) Scan(ctx context.Context, opts ScanOptions) ScanReport {
	if opts.Parallel <= 0 {
		opts.Parallel = 50
	}

	type outcome struct {
		name string
		res  *Result
		err  error
	}

	report := ScanReport{Errors: make(map[string]error)}

	var runnable []Module
	for _, m := range e.modules {
		if m.Available() {
			runnable = append(runnable, m)
		} else {
			report.Skipped = append(report.Skipped, m.Name())
		}
	}

	sem := make(chan struct{}, opts.Parallel)
	out := make(chan outcome, len(runnable))
	var wg sync.WaitGroup

	for _, m := range runnable {
		wg.Add(1)
		sem <- struct{}{}
		go func(m Module) {
			defer wg.Done()
			defer func() { <-sem }()
			mctx, cancel := context.WithTimeout(ctx, m.Timeout())
			defer cancel()
			res, err := m.Probe(mctx)
			out <- outcome{name: m.Name(), res: res, err: err}
		}(m)
	}
	go func() { wg.Wait(); close(out) }()

	builder := graph.NewBuilder()
	for o := range out {
		if o.err != nil {
			report.Errors[o.name] = o.err
		}
		if o.res != nil {
			builder.Merge(o.res.Nodes, o.res.Edges)
			report.Warnings = append(report.Warnings, o.res.Warnings...)
		}
		report.ModulesRun = append(report.ModulesRun, o.name)
	}

	sort.Strings(report.ModulesRun)
	sort.Strings(report.Skipped)
	sort.Strings(report.Warnings)

	g := builder.Build()
	g.CrossLink()
	report.Graph = g
	return report
}
