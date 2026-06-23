// Package graph defines the InfraSight data model: typed nodes and edges that
// every discovery module contributes to, and the merged Graph they form.
package graph

import (
	"sort"
	"time"
)

// NodeType enumerates the kinds of infrastructure entities InfraSight models.
type NodeType string

const (
	NodeHardware    NodeType = "HARDWARE"
	NodeOS          NodeType = "OS"
	NodeService     NodeType = "SERVICE"
	NodeProcess     NodeType = "PROCESS"
	NodeContainer   NodeType = "CONTAINER"
	NodePackage     NodeType = "PACKAGE"
	NodeWebsite     NodeType = "WEBSITE"
	NodeDatabase    NodeType = "DATABASE"
	NodePort        NodeType = "PORT"
	NodeNetwork     NodeType = "NETWORK"
	NodeCertificate NodeType = "CERTIFICATE"
	NodeConfig      NodeType = "CONFIG"
	NodeCron        NodeType = "CRON"
	NodeTimer       NodeType = "TIMER"
	NodeCloud       NodeType = "CLOUD_RESOURCE"
)

// Status is the operational state of an entity.
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusError    Status = "error"
	StatusUnknown  Status = "unknown"
)

// Health is a coarse health classification derived from resource thresholds.
// The empty value means "not evaluated".
type Health string

const (
	HealthUnknown  Health = ""
	HealthHealthy  Health = "healthy"
	HealthWarning  Health = "warning"
	HealthCritical Health = "critical"
)

// ResourceUsage captures optional, point-in-time utilisation for a node.
type ResourceUsage struct {
	CPUPercent         *float64 `json:"cpuPercent,omitempty"`
	MemoryMB           *float64 `json:"memoryMB,omitempty"`
	DiskMB             *float64 `json:"diskMB,omitempty"`
	NetworkBytesPerSec *float64 `json:"networkBytesPerSec,omitempty"`
}

// Node is a single discovered entity. ID must be globally unique and stable
// across scans (format: "<module-domain>:<type>:<identifier>").
type Node struct {
	ID           string         `json:"id"`
	Type         NodeType       `json:"type"`
	Label        string         `json:"label"`
	Version      string         `json:"version,omitempty"`
	Status       Status         `json:"status"`
	Health       Health         `json:"health,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	Resource     *ResourceUsage `json:"resourceUsage,omitempty"`
	DiscoveredAt time.Time      `json:"discoveredAt"`
	DiscoveredBy string         `json:"discoveredBy"`
}

// RelationType enumerates the semantics of an edge (source -> target).
type RelationType string

const (
	RelRunsOn     RelationType = "RUNS_ON"
	RelDependsOn  RelationType = "DEPENDS_ON"
	RelServes     RelationType = "SERVES"
	RelListensOn  RelationType = "LISTENS_ON"
	RelConnectsTo RelationType = "CONNECTS_TO"
	RelProxiesTo  RelationType = "PROXIES_TO"
	RelMounts     RelationType = "MOUNTS"
	RelIncludes   RelationType = "INCLUDES"
	RelSchedules  RelationType = "SCHEDULES"
	RelContains   RelationType = "CONTAINS"
	RelParentOf   RelationType = "PARENT_OF"
	RelEncrypts   RelationType = "ENCRYPTS"
)

// Edge is a directed relationship between two nodes.
type Edge struct {
	Source   string         `json:"source"`
	Target   string         `json:"target"`
	Relation RelationType   `json:"relation"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Graph is the merged, deduplicated result of all modules.
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// sortStable orders nodes and edges deterministically so identical hosts
// produce byte-identical output (a stated non-functional requirement).
func (g *Graph) sortStable() {
	sort.Slice(g.Nodes, func(i, j int) bool { return g.Nodes[i].ID < g.Nodes[j].ID })
	sort.Slice(g.Edges, func(i, j int) bool {
		if g.Edges[i].Source != g.Edges[j].Source {
			return g.Edges[i].Source < g.Edges[j].Source
		}
		if g.Edges[i].Relation != g.Edges[j].Relation {
			return g.Edges[i].Relation < g.Edges[j].Relation
		}
		return g.Edges[i].Target < g.Edges[j].Target
	})
}

func (g *Graph) addEdgeUnique(e Edge) {
	for _, ex := range g.Edges {
		if ex.Source == e.Source && ex.Target == e.Target && ex.Relation == e.Relation {
			return
		}
	}
	g.Edges = append(g.Edges, e)
}

// CrossLink adds implicit edges between modules that probe independently.
// Today it anchors hardware and runtime entities to the OS node so the graph
// is connected rather than a set of islands. This is where richer inference
// (process -> port -> website -> cert) will grow.
func (g *Graph) CrossLink() {
	var osID string
	for _, n := range g.Nodes {
		if n.Type == NodeOS {
			osID = n.ID
			break
		}
	}
	if osID == "" {
		return
	}
	for _, n := range g.Nodes {
		switch n.Type {
		case NodeHardware:
			g.addEdgeUnique(Edge{Source: osID, Target: n.ID, Relation: RelRunsOn})
		case NodeProcess, NodeService, NodeContainer:
			g.addEdgeUnique(Edge{Source: n.ID, Target: osID, Relation: RelRunsOn})
		}
	}
	g.sortStable()
}
