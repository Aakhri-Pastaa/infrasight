// Package security runs an audit over already-discovered graph nodes and reports
// security findings. It performs no host I/O — every rule reads data the probes
// already collected (certificate expiry/keys, port bind addresses, ...), which
// is what makes `--security` cheap and safe.
package security

import (
	"fmt"
	"sort"

	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
)

// Severity ranks a finding.
type Severity string

const (
	SevLow      Severity = "low"
	SevMedium   Severity = "medium"
	SevHigh     Severity = "high"
	SevCritical Severity = "critical"
)

func sevRank(s Severity) int {
	switch s {
	case SevCritical:
		return 4
	case SevHigh:
		return 3
	case SevMedium:
		return 2
	case SevLow:
		return 1
	default:
		return 0
	}
}

// Finding is a single security observation about a node.
type Finding struct {
	ID       string   `json:"id"`
	Severity Severity `json:"severity"`
	Title    string   `json:"title"`
	NodeID   string   `json:"nodeId"`
	Detail   string   `json:"detail,omitempty"`
}

// Audit applies the rule set over the nodes and returns findings, most severe
// first (stable within a severity by node ID).
func Audit(nodes []graph.Node) []Finding {
	var findings []Finding
	for _, n := range nodes {
		switch n.Type {
		case graph.NodeCertificate:
			findings = append(findings, auditCert(n)...)
		case graph.NodePort:
			findings = append(findings, auditPort(n)...)
		}
	}
	sort.SliceStable(findings, func(i, j int) bool {
		if a, b := sevRank(findings[i].Severity), sevRank(findings[j].Severity); a != b {
			return a > b
		}
		return findings[i].NodeID < findings[j].NodeID
	})
	return findings
}

func auditCert(n graph.Node) []Finding {
	var fs []Finding
	if days, ok := metaInt(n, "daysUntilExpiry"); ok {
		switch {
		case days < 0:
			fs = append(fs, Finding{"cert-expired", SevCritical, "TLS certificate expired", n.ID,
				fmt.Sprintf("%s expired %d day(s) ago", n.Label, -days)})
		case days < 14:
			fs = append(fs, Finding{"cert-expiring", SevHigh, "TLS certificate expiring very soon", n.ID,
				fmt.Sprintf("%s expires in %d day(s)", n.Label, days)})
		case days < 30:
			fs = append(fs, Finding{"cert-expiring", SevMedium, "TLS certificate expiring soon", n.ID,
				fmt.Sprintf("%s expires in %d day(s)", n.Label, days)})
		}
	}
	if metaString(n, "keyType") == "RSA" {
		if bits, ok := metaInt(n, "keyBits"); ok && bits > 0 && bits < 2048 {
			fs = append(fs, Finding{"cert-weak-key", SevHigh, "Weak TLS key", n.ID,
				fmt.Sprintf("%s uses a %d-bit RSA key (< 2048)", n.Label, bits)})
		}
	}
	return fs
}

func auditPort(n graph.Node) []Finding {
	switch metaString(n, "bindAddress") {
	case "0.0.0.0", "::", "[::]", "*":
		return []Finding{{"port-world-listening", SevLow, "Port reachable from any interface", n.ID,
			fmt.Sprintf("port %s is bound to %s", metaString(n, "port"), metaString(n, "bindAddress"))}}
	}
	return nil
}

// WorstSeverity returns the most severe finding severity, or "" if there are none.
func WorstSeverity(findings []Finding) Severity {
	worst := Severity("")
	for _, f := range findings {
		if sevRank(f.Severity) > sevRank(worst) {
			worst = f.Severity
		}
	}
	return worst
}

func metaInt(n graph.Node, key string) (int, bool) {
	switch x := n.Metadata[key].(type) {
	case int:
		return x, true
	case int64:
		return int(x), true
	case float64:
		return int(x), true
	}
	return 0, false
}

func metaString(n graph.Node, key string) string {
	if s, ok := n.Metadata[key].(string); ok {
		return s
	}
	return ""
}
