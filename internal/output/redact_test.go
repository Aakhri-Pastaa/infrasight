package output

import (
	"testing"

	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
)

func TestRedact(t *testing.T) {
	doc := Document{
		Scan: ScanMeta{Hostname: "prod-web-01"},
		Nodes: []graph.Node{
			{ID: "package:apt:nginx", Type: graph.NodePackage, Label: "nginx", Version: "1.24.0", Metadata: map[string]any{"manager": "dpkg"}},
			{ID: "port:tcp:443", Type: graph.NodePort, Metadata: map[string]any{"port": "443", "bindAddress": "0.0.0.0"}},
			{ID: "os:ubuntu:24.04", Type: graph.NodeOS, Version: "24.04", Metadata: map[string]any{"hostname": "prod-web-01", "versionId": "24.04"}},
		},
	}

	r := Redact(doc)

	if r.Scan.Hostname != RedactedValue || !r.Redacted {
		t.Fatalf("scan not redacted: hostname=%q redacted=%v", r.Scan.Hostname, r.Redacted)
	}
	if r.Nodes[0].Version != RedactedValue {
		t.Errorf("package version not redacted: %q", r.Nodes[0].Version)
	}
	if r.Nodes[1].Metadata["bindAddress"] != RedactedValue {
		t.Errorf("bind address not redacted: %v", r.Nodes[1].Metadata["bindAddress"])
	}
	if r.Nodes[1].Metadata["port"] != "443" {
		t.Errorf("port should be preserved, got %v", r.Nodes[1].Metadata["port"])
	}
	if r.Nodes[2].Metadata["hostname"] != RedactedValue || r.Nodes[2].Metadata["versionId"] != RedactedValue {
		t.Errorf("os metadata not redacted: %v", r.Nodes[2].Metadata)
	}
	if r.Nodes[0].Label != "nginx" {
		t.Errorf("label should be preserved, got %q", r.Nodes[0].Label)
	}

	// The original document must be untouched (clone, not mutate).
	if doc.Scan.Hostname != "prod-web-01" || doc.Nodes[0].Version != "1.24.0" || doc.Nodes[1].Metadata["bindAddress"] != "0.0.0.0" {
		t.Error("Redact mutated the original document")
	}
}
