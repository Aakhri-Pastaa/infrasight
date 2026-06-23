package hardware

import (
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

// Memory reports total/available RAM and swap, with a health classification
// based on utilisation.
type Memory struct{}

// NewMemory constructs the memory module.
func NewMemory() *Memory { return &Memory{} }

func (m *Memory) Name() string            { return "hardware.memory" }
func (m *Memory) Description() string     { return "RAM and swap totals with utilisation health" }
func (m *Memory) RequiredTools() []string { return nil }
func (m *Memory) RequiresRoot() bool      { return false }
func (m *Memory) RequiresNetwork() bool   { return false }
func (m *Memory) Timeout() time.Duration  { return 10 * time.Second }
func (m *Memory) Risk() discovery.Risk    { return discovery.RiskLow }
