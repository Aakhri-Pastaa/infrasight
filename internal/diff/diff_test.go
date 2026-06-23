package diff

import (
	"testing"

	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
	"github.com/Aakhri-Pastaa/infrasight/internal/output"
)

func doc(nodes ...graph.Node) output.Document {
	return output.Document{Nodes: nodes}
}

func TestCompare(t *testing.T) {
	base := doc(
		graph.Node{ID: "package:apt:nginx", Type: graph.NodePackage, Label: "nginx", Version: "1.22.1", Health: graph.HealthHealthy},
		graph.Node{ID: "hardware:disk:/", Type: graph.NodeHardware, Label: "/", Health: graph.HealthWarning},
		graph.Node{ID: "service:systemd:old.service", Type: graph.NodeService, Label: "old.service", Health: graph.HealthHealthy},
	)
	cur := doc(
		graph.Node{ID: "package:apt:nginx", Type: graph.NodePackage, Label: "nginx", Version: "1.24.0", Health: graph.HealthHealthy}, // upgraded
		graph.Node{ID: "hardware:disk:/", Type: graph.NodeHardware, Label: "/", Health: graph.HealthCritical},                        // regressed
		graph.Node{ID: "service:systemd:new.service", Type: graph.NodeService, Label: "new.service", Health: graph.HealthHealthy},    // added (old removed)
	)

	r := Compare(base, cur)

	if len(r.Added) != 1 || r.Added[0].ID != "service:systemd:new.service" {
		t.Errorf("added = %+v", r.Added)
	}
	if len(r.Removed) != 1 || r.Removed[0].ID != "service:systemd:old.service" {
		t.Errorf("removed = %+v", r.Removed)
	}
	if len(r.Versions) != 1 || r.Versions[0].From != "1.22.1" || r.Versions[0].To != "1.24.0" {
		t.Errorf("versions = %+v", r.Versions)
	}
	if len(r.Health) != 1 || r.Health[0].From != graph.HealthWarning || r.Health[0].To != graph.HealthCritical {
		t.Errorf("health = %+v", r.Health)
	}
	if r.Empty() {
		t.Error("expected drift, got Empty()")
	}
	if !r.HasCriticalRegression() {
		t.Error("expected a critical regression to be detected")
	}
}

func TestCompareIdentical(t *testing.T) {
	d := doc(graph.Node{ID: "os:x:1", Type: graph.NodeOS, Label: "x", Health: graph.HealthHealthy})
	if r := Compare(d, d); !r.Empty() {
		t.Errorf("identical scans should produce no drift: %+v", r)
	}
}
