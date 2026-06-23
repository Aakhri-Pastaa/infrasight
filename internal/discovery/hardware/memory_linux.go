//go:build linux

package hardware

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
)

// Utilisation thresholds for health classification.
const (
	memWarnPercent = 75.0
	memCritPercent = 90.0
)

func (m *Memory) Available() bool {
	_, err := os.Stat("/proc/meminfo")
	return err == nil
}

func (m *Memory) Probe(ctx context.Context) (*discovery.Result, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	kb := meminfoKB(data)

	total := kb["MemTotal"]
	if total == 0 {
		return nil, errors.New("hardware.memory: MemTotal missing from /proc/meminfo")
	}
	avail := kb["MemAvailable"]
	used := total - avail
	usedPct := float64(used) / float64(total) * 100

	health := graph.HealthHealthy
	switch {
	case usedPct >= memCritPercent:
		health = graph.HealthCritical
	case usedPct >= memWarnPercent:
		health = graph.HealthWarning
	}

	usedMB := float64(used) / 1024
	totalGiB := float64(total) / (1024 * 1024)

	node := graph.Node{
		ID:     "hardware:memory",
		Type:   graph.NodeHardware,
		Label:  fmt.Sprintf("RAM %.1f GiB", totalGiB),
		Status: graph.StatusActive,
		Health: health,
		Metadata: map[string]any{
			"totalKB":     total,
			"availableKB": avail,
			"usedKB":      used,
			"usedPercent": math.Round(usedPct*10) / 10,
			"swapTotalKB": kb["SwapTotal"],
			"swapFreeKB":  kb["SwapFree"],
		},
		Resource:     &graph.ResourceUsage{MemoryMB: &usedMB},
		DiscoveredAt: time.Now().UTC(),
		DiscoveredBy: m.Name(),
	}
	return &discovery.Result{Nodes: []graph.Node{node}}, nil
}
