// Package network discovers network topology: listening ports, connections,
// and (in future) firewall rules, routes, and tunnels.
package network

import (
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

// Ports discovers TCP listening sockets and the processes that own them.
type Ports struct{}

// NewPorts constructs the ports module.
func NewPorts() *Ports { return &Ports{} }

func (p *Ports) Name() string            { return "network.ports" }
func (p *Ports) Description() string     { return "TCP listening ports and owning processes" }
func (p *Ports) RequiredTools() []string { return []string{"ss"} }
func (p *Ports) RequiresRoot() bool      { return false } // root reveals more process owners
func (p *Ports) RequiresNetwork() bool   { return false }
func (p *Ports) Timeout() time.Duration  { return 15 * time.Second }
func (p *Ports) Risk() discovery.Risk    { return discovery.RiskMedium }
