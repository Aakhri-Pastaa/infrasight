package services

import (
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

// Docker discovers running containers and the host ports they publish.
type Docker struct{}

// NewDocker constructs the docker module.
func NewDocker() *Docker { return &Docker{} }

func (d *Docker) Name() string            { return "services.docker" }
func (d *Docker) Description() string     { return "Running Docker containers and published ports" }
func (d *Docker) RequiredTools() []string { return []string{"docker"} }
func (d *Docker) RequiresRoot() bool      { return false } // docker group is enough
func (d *Docker) RequiresNetwork() bool   { return false }
func (d *Docker) Timeout() time.Duration  { return 20 * time.Second }
func (d *Docker) Risk() discovery.Risk    { return discovery.RiskLow }
