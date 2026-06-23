// Package discovery defines the module contract and the orchestration engine
// that runs every module concurrently and merges their findings.
package discovery

import (
	"context"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
)

// Risk classifies how intrusive a module is, for future consent gating.
type Risk int

const (
	RiskLow Risk = iota
	RiskMedium
	RiskHigh
)

func (r Risk) String() string {
	switch r {
	case RiskMedium:
		return "MEDIUM"
	case RiskHigh:
		return "HIGH"
	default:
		return "LOW"
	}
}

// Module is implemented by every discovery module. It mirrors the
// DiscoveryModule contract from the design spec. (Clean() is intentionally
// omitted: no current module produces side effects that need cleanup. It can
// be reintroduced as an optional interface when a module needs it.)
type Module interface {
	// Name is a stable identifier, e.g. "hardware.cpu" or "network.ports".
	Name() string
	// Description is a one-line human summary.
	Description() string
	// RequiredTools lists external binaries the module shells out to.
	RequiredTools() []string
	// RequiresRoot reports whether full data needs elevated privileges.
	RequiresRoot() bool
	// RequiresNetwork reports whether the module makes network calls.
	RequiresNetwork() bool
	// Timeout bounds a single Probe invocation.
	Timeout() time.Duration
	// Risk gates consent for intrusive modules.
	Risk() Risk
	// Available reports whether the module can run on this host (right OS and
	// required tools present). It must be cheap and must not probe.
	Available() bool
	// Probe performs the discovery. It may return a partial Result alongside an
	// error when it gathered some data before failing.
	Probe(ctx context.Context) (*Result, error)
}

// Result is a module's contribution to the global graph.
type Result struct {
	Nodes    []graph.Node
	Edges    []graph.Edge
	Warnings []string
}
