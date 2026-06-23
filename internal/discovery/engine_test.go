package discovery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
)

type fakeModule struct {
	name   string
	avail  bool
	result *Result
	err    error
}

func (f *fakeModule) Name() string                           { return f.name }
func (f *fakeModule) Description() string                    { return "" }
func (f *fakeModule) RequiredTools() []string                { return nil }
func (f *fakeModule) RequiresRoot() bool                     { return false }
func (f *fakeModule) RequiresNetwork() bool                  { return false }
func (f *fakeModule) Timeout() time.Duration                 { return time.Second }
func (f *fakeModule) Risk() Risk                             { return RiskLow }
func (f *fakeModule) Available() bool                        { return f.avail }
func (f *fakeModule) Probe(context.Context) (*Result, error) { return f.result, f.err }

func TestEngineScan(t *testing.T) {
	good := &fakeModule{name: "good", avail: true,
		result: &Result{Nodes: []graph.Node{{ID: "os:x", Type: graph.NodeOS}}}}
	// returns a partial result AND an error — both must be honoured
	partial := &fakeModule{name: "partial", avail: true,
		result: &Result{Nodes: []graph.Node{{ID: "hw:y", Type: graph.NodeHardware}}, Warnings: []string{"heads up"}},
		err:    errors.New("partial failure")}
	unavail := &fakeModule{name: "skipme", avail: false}
	errOnly := &fakeModule{name: "boom", avail: true, err: errors.New("kaput")}

	rep := NewEngine([]Module{good, partial, unavail, errOnly}).
		Scan(context.Background(), ScanOptions{Parallel: 2})

	if !contains(rep.Skipped, "skipme") {
		t.Errorf("unavailable module should be skipped: %v", rep.Skipped)
	}
	if len(rep.ModulesRun) != 3 {
		t.Errorf("ModulesRun = %v, want the 3 available modules", rep.ModulesRun)
	}
	// good + partial nodes both merged (partial returned data alongside its error)
	if len(rep.Graph.Nodes) != 2 {
		t.Errorf("merged nodes = %d, want 2", len(rep.Graph.Nodes))
	}
	if rep.Errors["partial"] == nil || rep.Errors["boom"] == nil {
		t.Errorf("errors not recorded: %v", rep.Errors)
	}
	if !contains(rep.Warnings, "heads up") {
		t.Errorf("warning not collected: %v", rep.Warnings)
	}
	// cross-linking ran: the OS node anchors the hardware node
	var runsOn int
	for _, e := range rep.Graph.Edges {
		if e.Relation == graph.RelRunsOn {
			runsOn++
		}
	}
	if runsOn == 0 {
		t.Error("expected RUNS_ON edges from the OS anchor after the scan")
	}
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
