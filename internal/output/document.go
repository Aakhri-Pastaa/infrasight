// Package output turns a discovery graph into the user-facing artifacts:
// machine-readable JSON, a self-contained HTML report, and a terminal summary.
package output

import (
	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
)

// Document is the top-level serialised scan result. Its JSON form matches the
// infrasight schema (scan metadata + nodes + edges + summary).
type Document struct {
	Schema   string       `json:"$schema"`
	Scan     ScanMeta     `json:"scan"`
	Nodes    []graph.Node `json:"nodes"`
	Edges    []graph.Edge `json:"edges"`
	Summary  Summary      `json:"summary"`
	Redacted bool         `json:"redacted,omitempty"`
}

// ScanMeta describes a single scan run.
type ScanMeta struct {
	ID          string   `json:"id"`
	Hostname    string   `json:"hostname"`
	StartedAt   string   `json:"startedAt"`
	CompletedAt string   `json:"completedAt"`
	DurationMs  int64    `json:"durationMs"`
	Modules     []string `json:"modules"`
	Version     string   `json:"version"`
}

// Summary is the at-a-glance rollup.
type Summary struct {
	TotalNodes  int            `json:"totalNodes"`
	NodesByType map[string]int `json:"nodesByType"`
	TotalEdges  int            `json:"totalEdges"`
	Health      map[string]int `json:"healthSummary"`
	Warnings    []string       `json:"warnings"`
	Criticals   []string       `json:"criticals"`
}

const schemaURL = "https://infrasight.dev/schema/v1"

// BuildDocument assembles a Document from a graph, scan metadata, and any
// module-level warnings, computing the summary rollup from node health.
func BuildDocument(g graph.Graph, meta ScanMeta, moduleWarnings []string) Document {
	byType := make(map[string]int)
	health := map[string]int{"healthy": 0, "warning": 0, "critical": 0}
	var criticals, warnings []string

	for _, n := range g.Nodes {
		byType[string(n.Type)]++
		switch n.Health {
		case graph.HealthCritical:
			health["critical"]++
			criticals = append(criticals, string(n.Type)+" "+n.Label)
		case graph.HealthWarning:
			health["warning"]++
			warnings = append(warnings, string(n.Type)+" "+n.Label)
		case graph.HealthHealthy:
			health["healthy"]++
		}
	}
	warnings = append(warnings, moduleWarnings...)

	return Document{
		Schema: schemaURL,
		Scan:   meta,
		Nodes:  g.Nodes,
		Edges:  g.Edges,
		Summary: Summary{
			TotalNodes:  len(g.Nodes),
			NodesByType: byType,
			TotalEdges:  len(g.Edges),
			Health:      health,
			Warnings:    warnings,
			Criticals:   criticals,
		},
	}
}

// ExitCode maps findings to a process exit status: 2 for criticals, 1 for
// warnings, 0 for a clean scan.
func (d Document) ExitCode() int {
	switch {
	case len(d.Summary.Criticals) > 0:
		return 2
	case len(d.Summary.Warnings) > 0:
		return 1
	default:
		return 0
	}
}
