package output

import (
	"testing"

	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
)

func TestBuildDocumentSummary(t *testing.T) {
	g := graph.Graph{Nodes: []graph.Node{
		{ID: "1", Type: graph.NodeHardware, Label: "cpu", Health: graph.HealthHealthy},
		{ID: "2", Type: graph.NodeHardware, Label: "ram", Health: graph.HealthCritical},
		{ID: "3", Type: graph.NodePackage, Label: "pkg", Health: graph.HealthWarning},
	}}

	doc := BuildDocument(g, ScanMeta{Hostname: "h"}, []string{"a module warning"})

	if doc.Summary.TotalNodes != 3 {
		t.Fatalf("want 3 nodes, got %d", doc.Summary.TotalNodes)
	}
	if doc.Summary.NodesByType["HARDWARE"] != 2 {
		t.Errorf("want 2 HARDWARE, got %d", doc.Summary.NodesByType["HARDWARE"])
	}
	if doc.Summary.Health["critical"] != 1 || doc.Summary.Health["warning"] != 1 {
		t.Errorf("health rollup wrong: %+v", doc.Summary.Health)
	}
	if len(doc.Summary.Criticals) != 1 {
		t.Errorf("want 1 critical finding, got %v", doc.Summary.Criticals)
	}
	// One node warning + one module warning.
	if len(doc.Summary.Warnings) != 2 {
		t.Errorf("want 2 warnings, got %v", doc.Summary.Warnings)
	}
	if got := doc.ExitCode(); got != 2 {
		t.Errorf("want exit code 2 (critical present), got %d", got)
	}
}

func TestExitCodeWarningsOnly(t *testing.T) {
	g := graph.Graph{Nodes: []graph.Node{
		{ID: "1", Type: graph.NodePort, Label: "p", Health: graph.HealthWarning},
	}}
	doc := BuildDocument(g, ScanMeta{}, nil)
	if got := doc.ExitCode(); got != 1 {
		t.Errorf("want exit code 1 (warnings only), got %d", got)
	}
}
