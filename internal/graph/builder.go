package graph

// Builder accumulates nodes and edges from many modules, deduplicating by ID
// (nodes) and by source/relation/target (edges), then emits a sorted Graph.
type Builder struct {
	nodes map[string]Node
	edges map[string]Edge
}

// NewBuilder returns an empty Builder.
func NewBuilder() *Builder {
	return &Builder{
		nodes: make(map[string]Node),
		edges: make(map[string]Edge),
	}
}

// AddNode inserts a node, merging metadata if the ID was already seen.
func (b *Builder) AddNode(n Node) {
	if existing, ok := b.nodes[n.ID]; ok {
		b.nodes[n.ID] = mergeNode(existing, n)
		return
	}
	b.nodes[n.ID] = n
}

// AddEdge inserts an edge, collapsing duplicates.
func (b *Builder) AddEdge(e Edge) {
	b.edges[edgeKey(e)] = e
}

// Merge adds a batch of nodes and edges (a module's Result).
func (b *Builder) Merge(nodes []Node, edges []Edge) {
	for _, n := range nodes {
		b.AddNode(n)
	}
	for _, e := range edges {
		b.AddEdge(e)
	}
}

// Build materialises a deterministic Graph.
func (b *Builder) Build() Graph {
	g := Graph{
		Nodes: make([]Node, 0, len(b.nodes)),
		Edges: make([]Edge, 0, len(b.edges)),
	}
	for _, n := range b.nodes {
		g.Nodes = append(g.Nodes, n)
	}
	for _, e := range b.edges {
		g.Edges = append(g.Edges, e)
	}
	g.sortStable()
	return g
}

func edgeKey(e Edge) string {
	return e.Source + "|" + string(e.Relation) + "|" + e.Target
}

// mergeNode keeps the first node's identity but fills empty fields and unions
// metadata from the duplicate. Health escalates to the worst signal from any
// module, so e.g. a resources module reporting a full disk overrides an earlier
// "healthy" for the same node regardless of merge order.
func mergeNode(a, b Node) Node {
	if a.Metadata == nil {
		a.Metadata = make(map[string]any)
	}
	for k, v := range b.Metadata {
		if _, ok := a.Metadata[k]; !ok {
			a.Metadata[k] = v
		}
	}
	if a.Label == "" {
		a.Label = b.Label
	}
	if a.Version == "" {
		a.Version = b.Version
	}
	a.Health = worseHealth(a.Health, b.Health)
	if a.Resource == nil {
		a.Resource = b.Resource
	}
	return a
}

// worseHealth returns the more severe of two health values.
func worseHealth(x, y Health) Health {
	if healthRank(y) > healthRank(x) {
		return y
	}
	return x
}

func healthRank(h Health) int {
	switch h {
	case HealthCritical:
		return 3
	case HealthWarning:
		return 2
	case HealthHealthy:
		return 1
	default:
		return 0
	}
}
