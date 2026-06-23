// Package hardware discovers physical and virtual hardware: CPU, memory, and
// (in future) disks, NICs, and GPUs.
package hardware

import (
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

// CPU reports the host processor model and core/thread counts.
type CPU struct{}

// NewCPU constructs the CPU module.
func NewCPU() *CPU { return &CPU{} }

func (c *CPU) Name() string            { return "hardware.cpu" }
func (c *CPU) Description() string     { return "CPU model, core and thread counts" }
func (c *CPU) RequiredTools() []string { return nil }
func (c *CPU) RequiresRoot() bool      { return false }
func (c *CPU) RequiresNetwork() bool   { return false }
func (c *CPU) Timeout() time.Duration  { return 10 * time.Second }
func (c *CPU) Risk() discovery.Risk    { return discovery.RiskLow }
