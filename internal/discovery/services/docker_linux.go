//go:build linux

package services

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
	"github.com/Aakhri-Pastaa/infrasight/pkg/shell"
)

// dockerPort extracts the host port from a published-ports entry such as
// "0.0.0.0:8080->80/tcp" (captures 8080 and tcp).
var dockerPort = regexp.MustCompile(`(\d+)->\d+/(tcp|udp)`)

// Real tab characters; docker emits them verbatim between fields.
const dockerFormat = "{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.State}}\t{{.Status}}\t{{.Ports}}"

func (d *Docker) Available() bool {
	return shell.Available("docker")
}

func (d *Docker) Probe(ctx context.Context) (*discovery.Result, error) {
	out, err := shell.Run(ctx, d.Timeout(), "docker", "ps", "--no-trunc", "--format", dockerFormat)
	res := &discovery.Result{}
	if err != nil {
		// Daemon not running or no permission — degrade gracefully.
		res.Warnings = append(res.Warnings,
			"services.docker: 'docker ps' failed (daemon not running or permission denied)")
		return res, nil
	}

	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		c := strings.Split(line, "\t")
		if len(c) < 5 {
			continue
		}
		id, name, image, state, status := c[0], c[1], c[2], c[3], c[4]
		ports := ""
		if len(c) >= 6 {
			ports = c[5]
		}
		short := id
		if len(short) > 12 {
			short = short[:12]
		}

		health := graph.HealthHealthy
		if state != "" && state != "running" {
			health = graph.HealthWarning
		}
		containerID := "container:docker:" + short
		res.Nodes = append(res.Nodes, graph.Node{
			ID:     containerID,
			Type:   graph.NodeContainer,
			Label:  name,
			Status: graph.StatusActive,
			Health: health,
			Metadata: map[string]any{
				"image":       image,
				"state":       state,
				"status":      status,
				"containerId": short,
			},
			DiscoveredAt: time.Now().UTC(),
			DiscoveredBy: d.Name(),
		})

		seen := map[string]bool{}
		for _, m := range dockerPort.FindAllStringSubmatch(ports, -1) {
			hostPort, proto := m[1], m[2]
			portID := "port:" + proto + ":" + hostPort
			if !seen[portID] {
				seen[portID] = true
				res.Nodes = append(res.Nodes, graph.Node{
					ID:     portID,
					Type:   graph.NodePort,
					Label:  ":" + hostPort,
					Status: graph.StatusActive,
					Health: graph.HealthHealthy,
					Metadata: map[string]any{
						"protocol": proto,
						"port":     hostPort,
						"source":   "docker",
					},
					DiscoveredAt: time.Now().UTC(),
					DiscoveredBy: d.Name(),
				})
			}
			res.Edges = append(res.Edges, graph.Edge{
				Source:   containerID,
				Target:   portID,
				Relation: graph.RelListensOn,
			})
		}
	}
	return res, nil
}
