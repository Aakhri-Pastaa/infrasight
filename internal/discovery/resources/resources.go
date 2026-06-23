// Package resources profiles live host utilisation — CPU load and filesystem
// usage — and attaches health to the affected nodes. CPU data enriches the
// existing hardware:cpu node (merged by ID); each mounted filesystem becomes a
// HARDWARE disk node.
package resources

import (
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

// System reports CPU utilisation/load and per-filesystem disk usage.
type System struct{}

// NewSystem constructs the resources module.
func NewSystem() *System { return &System{} }

func (s *System) Name() string            { return "resources.system" }
func (s *System) Description() string     { return "CPU utilisation/load and filesystem usage with health" }
func (s *System) RequiredTools() []string { return nil }
func (s *System) RequiresRoot() bool      { return false }
func (s *System) RequiresNetwork() bool   { return false }
func (s *System) Timeout() time.Duration  { return 15 * time.Second }
func (s *System) Risk() discovery.Risk    { return discovery.RiskLow }
