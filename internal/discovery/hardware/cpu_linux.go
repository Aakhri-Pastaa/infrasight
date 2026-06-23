//go:build linux

package hardware

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
)

func (c *CPU) Available() bool {
	_, err := os.Stat("/proc/cpuinfo")
	return err == nil
}

func (c *CPU) Probe(ctx context.Context) (*discovery.Result, error) {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return nil, err
	}

	var model string
	logical := 0
	coresPerSocket := 0
	sockets := map[string]struct{}{}

	for _, line := range strings.Split(string(data), "\n") {
		key, val, ok := splitKV(line)
		if !ok {
			continue
		}
		switch key {
		case "processor":
			logical++
		case "model name":
			if model == "" {
				model = val
			}
		case "cpu cores":
			if n, e := strconv.Atoi(val); e == nil {
				coresPerSocket = n
			}
		case "physical id":
			sockets[val] = struct{}{}
		}
	}

	if logical == 0 {
		logical = runtime.NumCPU()
	}
	if model == "" {
		model = "Unknown CPU"
	}
	socketCount := len(sockets)
	if socketCount == 0 {
		socketCount = 1
	}
	physical := coresPerSocket * socketCount
	if physical == 0 {
		physical = logical
	}

	node := graph.Node{
		ID:     "hardware:cpu",
		Type:   graph.NodeHardware,
		Label:  model,
		Status: graph.StatusActive,
		Health: graph.HealthHealthy,
		Metadata: map[string]any{
			"model":         model,
			"logicalCPUs":   logical,
			"physicalCores": physical,
			"sockets":       socketCount,
			"arch":          runtime.GOARCH,
		},
		DiscoveredAt: time.Now().UTC(),
		DiscoveredBy: c.Name(),
	}
	return &discovery.Result{Nodes: []graph.Node{node}}, nil
}
