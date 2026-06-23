// Package packages discovers installed software across package managers
// (dpkg, rpm, apk, pacman) via the pkgbackend abstraction. The backend is
// selected at runtime, so the same module works on Debian, RHEL, Alpine, and
// Arch family hosts.
package packages

import (
	"context"
	"errors"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/pkgbackend"
	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
)

// Module lists installed packages using whichever package manager is present.
type Module struct{}

// New constructs the packages module.
func New() *Module { return &Module{} }

func (m *Module) Name() string            { return "packages" }
func (m *Module) Description() string     { return "Installed packages (dpkg/rpm/apk/pacman)" }
func (m *Module) RequiredTools() []string { return nil } // depends on the detected backend
func (m *Module) RequiresRoot() bool      { return false }
func (m *Module) RequiresNetwork() bool   { return false }
func (m *Module) Timeout() time.Duration  { return 30 * time.Second }
func (m *Module) Risk() discovery.Risk    { return discovery.RiskLow }

func (m *Module) Available() bool { return pkgbackend.Detect() != nil }

func (m *Module) Probe(ctx context.Context) (*discovery.Result, error) {
	backend := pkgbackend.Detect()
	if backend == nil {
		return nil, errors.New("packages: no supported package manager found")
	}

	pkgs, err := backend.List(ctx, m.Timeout())
	if err != nil && len(pkgs) == 0 {
		return nil, err
	}

	res := &discovery.Result{}
	now := time.Now().UTC()
	for _, p := range pkgs {
		meta := map[string]any{"manager": backend.Key()}
		if p.Arch != "" {
			meta["arch"] = p.Arch
		}
		if p.SizeKB > 0 {
			meta["installedSizeKB"] = p.SizeKB
		}
		res.Nodes = append(res.Nodes, graph.Node{
			ID:           pkgbackend.NodeID(backend.Key(), p.Name),
			Type:         graph.NodePackage,
			Label:        p.Name,
			Version:      p.Version,
			Status:       graph.StatusActive,
			Health:       graph.HealthHealthy,
			Metadata:     meta,
			DiscoveredAt: now,
			DiscoveredBy: m.Name(),
		})
	}
	return res, nil
}
