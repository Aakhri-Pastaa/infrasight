// Package osinfo discovers the operating system and kernel. It is named
// "osinfo" rather than "os" to avoid shadowing the standard library.
package osinfo

import (
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

// Distro identifies the Linux distribution and kernel. The node it emits acts
// as the anchor that hardware and runtime entities cross-link to.
type Distro struct{}

// NewDistro constructs the distro module.
func NewDistro() *Distro { return &Distro{} }

func (d *Distro) Name() string            { return "os.distro" }
func (d *Distro) Description() string     { return "Distribution, version, kernel and hostname" }
func (d *Distro) RequiredTools() []string { return nil }
func (d *Distro) RequiresRoot() bool      { return false }
func (d *Distro) RequiresNetwork() bool   { return false }
func (d *Distro) Timeout() time.Duration  { return 10 * time.Second }
func (d *Distro) Risk() discovery.Risk    { return discovery.RiskLow }
