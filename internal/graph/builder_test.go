package graph

import "testing"

func TestBuilderDedupAndSort(t *testing.T) {
	b := NewBuilder()
	b.AddNode(Node{ID: "b", Type: NodePackage, Label: "b", Metadata: map[string]any{"x": 1}})
	b.AddNode(Node{ID: "a", Type: NodePackage, Label: "a"})
	// Duplicate ID: keep first identity, union metadata.
	b.AddNode(Node{ID: "b", Type: NodePackage, Label: "b2", Metadata: map[string]any{"y": 2}})
	b.AddEdge(Edge{Source: "a", Target: "b", Relation: RelDependsOn})
	b.AddEdge(Edge{Source: "a", Target: "b", Relation: RelDependsOn}) // duplicate

	g := b.Build()

	if len(g.Nodes) != 2 {
		t.Fatalf("want 2 nodes after dedup, got %d", len(g.Nodes))
	}
	if g.Nodes[0].ID != "a" || g.Nodes[1].ID != "b" {
		t.Fatalf("nodes not sorted by ID: %+v", g.Nodes)
	}
	nb := g.Nodes[1]
	if nb.Label != "b" {
		t.Errorf("merge should keep first label 'b', got %q", nb.Label)
	}
	if nb.Metadata["x"] != 1 || nb.Metadata["y"] != 2 {
		t.Errorf("metadata not unioned: %+v", nb.Metadata)
	}
	if len(g.Edges) != 1 {
		t.Fatalf("want 1 edge after dedup, got %d", len(g.Edges))
	}
}

func TestCrossLinkAnchorsToOS(t *testing.T) {
	g := Graph{Nodes: []Node{
		{ID: "os:x:1", Type: NodeOS},
		{ID: "hardware:cpu", Type: NodeHardware},
		{ID: "process:1", Type: NodeProcess},
	}}
	g.CrossLink()

	var runsOn int
	for _, e := range g.Edges {
		if e.Relation == RelRunsOn {
			runsOn++
		}
	}
	if runsOn != 2 {
		t.Fatalf("want 2 RUNS_ON edges (OS->hw, process->OS), got %d: %+v", runsOn, g.Edges)
	}
}

func TestCrossLinkNoOSIsNoop(t *testing.T) {
	g := Graph{Nodes: []Node{{ID: "hardware:cpu", Type: NodeHardware}}}
	g.CrossLink()
	if len(g.Edges) != 0 {
		t.Fatalf("want no edges without an OS node, got %d", len(g.Edges))
	}
}
