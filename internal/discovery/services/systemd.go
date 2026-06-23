// Package services discovers service managers and the workloads they run:
// systemd units and Docker containers today.
package services

import (
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

// Systemd discovers running systemd service units, the main process of each,
// and the package that ships the unit.
type Systemd struct{}

// NewSystemd constructs the systemd module.
func NewSystemd() *Systemd { return &Systemd{} }

func (s *Systemd) Name() string { return "services.systemd" }
func (s *Systemd) Description() string {
	return "Running systemd services, their main process and package"
}
func (s *Systemd) RequiredTools() []string { return []string{"systemctl"} }
func (s *Systemd) RequiresRoot() bool      { return false }
func (s *Systemd) RequiresNetwork() bool   { return false }
func (s *Systemd) Timeout() time.Duration  { return 30 * time.Second }
func (s *Systemd) Risk() discovery.Risk    { return discovery.RiskLow }
