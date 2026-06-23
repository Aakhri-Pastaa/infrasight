//go:build linux

package osinfo

import (
	"context"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
)

func (d *Distro) Available() bool {
	_, err := os.Stat("/etc/os-release")
	return err == nil
}

func (d *Distro) Probe(ctx context.Context) (*discovery.Result, error) {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return nil, err
	}
	fields := parseOSRelease(data)

	id := orDefault(fields["ID"], "linux")
	versionID := fields["VERSION_ID"]
	pretty := orDefault(fields["PRETTY_NAME"], id)

	hostname, _ := os.Hostname()
	kernel := readTrim("/proc/sys/kernel/osrelease")

	node := graph.Node{
		ID:      "os:" + id + ":" + orDefault(versionID, "unknown"),
		Type:    graph.NodeOS,
		Label:   pretty,
		Version: versionID,
		Status:  graph.StatusActive,
		Health:  graph.HealthHealthy,
		Metadata: map[string]any{
			"id":         id,
			"versionId":  versionID,
			"prettyName": pretty,
			"kernel":     kernel,
			"hostname":   hostname,
			"arch":       runtime.GOARCH,
		},
		DiscoveredAt: time.Now().UTC(),
		DiscoveredBy: d.Name(),
	}
	return &discovery.Result{Nodes: []graph.Node{node}}, nil
}

// parseOSRelease parses KEY=VALUE lines, stripping surrounding quotes.
func parseOSRelease(data []byte) map[string]string {
	out := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"'`)
	}
	return out
}

func readTrim(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
