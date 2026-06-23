package security

import (
	"testing"

	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
)

func TestAudit(t *testing.T) {
	nodes := []graph.Node{
		{ID: "cert:a", Type: graph.NodeCertificate, Label: "expired.example.com",
			Metadata: map[string]any{"daysUntilExpiry": -5, "keyType": "RSA", "keyBits": 4096}},
		{ID: "cert:b", Type: graph.NodeCertificate, Label: "weak.example.com",
			Metadata: map[string]any{"daysUntilExpiry": 200, "keyType": "RSA", "keyBits": 1024}},
		{ID: "cert:c", Type: graph.NodeCertificate, Label: "soon.example.com",
			Metadata: map[string]any{"daysUntilExpiry": 10, "keyType": "ECDSA", "keyBits": 256}},
		{ID: "port:tcp:22", Type: graph.NodePort,
			Metadata: map[string]any{"port": "22", "bindAddress": "0.0.0.0"}},
		{ID: "port:tcp:5432", Type: graph.NodePort,
			Metadata: map[string]any{"port": "5432", "bindAddress": "127.0.0.1"}}, // not world-listening
		{ID: "cert:ok", Type: graph.NodeCertificate, Label: "good.example.com",
			Metadata: map[string]any{"daysUntilExpiry": 300, "keyType": "ECDSA", "keyBits": 256}}, // clean
	}

	fs := Audit(nodes)

	// expected: expired (critical), weak key (high), expiring-soon (high), world-listening (low)
	if len(fs) != 4 {
		t.Fatalf("want 4 findings, got %d: %+v", len(fs), fs)
	}
	if fs[0].Severity != SevCritical || fs[0].ID != "cert-expired" {
		t.Errorf("most severe should be the expired cert, got %+v", fs[0])
	}
	if last := fs[len(fs)-1]; last.Severity != SevLow || last.ID != "port-world-listening" {
		t.Errorf("least severe should be the world-listening port, got %+v", last)
	}
	if WorstSeverity(fs) != SevCritical {
		t.Errorf("WorstSeverity = %q, want critical", WorstSeverity(fs))
	}
}

func TestAuditClean(t *testing.T) {
	nodes := []graph.Node{
		{ID: "cert:ok", Type: graph.NodeCertificate, Metadata: map[string]any{"daysUntilExpiry": 365, "keyType": "ECDSA", "keyBits": 256}},
		{ID: "port:tcp:5432", Type: graph.NodePort, Metadata: map[string]any{"port": "5432", "bindAddress": "127.0.0.1"}},
	}
	if fs := Audit(nodes); len(fs) != 0 {
		t.Errorf("clean host should have no findings, got %+v", fs)
	}
	if WorstSeverity(nil) != "" {
		t.Error("WorstSeverity(nil) should be empty")
	}
}
