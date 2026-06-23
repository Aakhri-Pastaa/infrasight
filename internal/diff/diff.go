// Package diff compares two scans and reports the drift between them: nodes that
// appeared or disappeared, version changes, and health changes.
package diff

import (
	"sort"

	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
	"github.com/Aakhri-Pastaa/infrasight/internal/output"
)

// VersionChange is a node whose version differs between baseline and current.
type VersionChange struct {
	ID    string
	Label string
	Type  graph.NodeType
	From  string
	To    string
}

// HealthChange is a node whose health differs between baseline and current.
type HealthChange struct {
	ID    string
	Label string
	Type  graph.NodeType
	From  graph.Health
	To    graph.Health
}

// Result is the computed drift between two scans.
type Result struct {
	Added    []graph.Node
	Removed  []graph.Node
	Versions []VersionChange
	Health   []HealthChange
}

// Empty reports whether there is no drift at all.
func (r Result) Empty() bool {
	return len(r.Added) == 0 && len(r.Removed) == 0 &&
		len(r.Versions) == 0 && len(r.Health) == 0
}

// Compare diffs current against baseline by node ID.
func Compare(baseline, current output.Document) Result {
	base := index(baseline.Nodes)
	cur := index(current.Nodes)

	var r Result
	for id, n := range cur {
		if _, ok := base[id]; !ok {
			r.Added = append(r.Added, n)
		}
	}
	for id, n := range base {
		if _, ok := cur[id]; !ok {
			r.Removed = append(r.Removed, n)
		}
	}
	for id, cn := range cur {
		bn, ok := base[id]
		if !ok {
			continue
		}
		if bn.Version != cn.Version {
			r.Versions = append(r.Versions, VersionChange{
				ID: id, Label: cn.Label, Type: cn.Type, From: bn.Version, To: cn.Version,
			})
		}
		if bn.Health != cn.Health {
			r.Health = append(r.Health, HealthChange{
				ID: id, Label: cn.Label, Type: cn.Type, From: bn.Health, To: cn.Health,
			})
		}
	}

	sortNodes(r.Added)
	sortNodes(r.Removed)
	sort.Slice(r.Versions, func(i, j int) bool {
		return less(r.Versions[i].Label, r.Versions[i].ID, r.Versions[j].Label, r.Versions[j].ID)
	})
	sort.Slice(r.Health, func(i, j int) bool { return less(r.Health[i].Label, r.Health[i].ID, r.Health[j].Label, r.Health[j].ID) })
	return r
}

func index(nodes []graph.Node) map[string]graph.Node {
	m := make(map[string]graph.Node, len(nodes))
	for _, n := range nodes {
		m[n.ID] = n
	}
	return m
}

func sortNodes(ns []graph.Node) {
	sort.Slice(ns, func(i, j int) bool { return less(string(ns[i].Type), ns[i].ID, string(ns[j].Type), ns[j].ID) })
}

func less(a1, a2, b1, b2 string) bool {
	if a1 != b1 {
		return a1 < b1
	}
	return a2 < b2
}

// Severity maps health to an ordinal so callers can detect regressions.
func Severity(h graph.Health) int {
	switch h {
	case graph.HealthCritical:
		return 3
	case graph.HealthWarning:
		return 2
	case graph.HealthHealthy:
		return 1
	default:
		return 0
	}
}

// WorstNewHealth returns the most severe "to" health among health changes that
// got worse — used to set the process exit code.
func (r Result) WorstNewHealth() graph.Health {
	worst := graph.HealthUnknown
	for _, h := range r.Health {
		if Severity(h.To) > Severity(h.From) && Severity(h.To) > Severity(worst) {
			worst = h.To
		}
	}
	return worst
}

// HasCriticalRegression reports whether anything regressed to critical health.
func (r Result) HasCriticalRegression() bool {
	return r.WorstNewHealth() == graph.HealthCritical
}
