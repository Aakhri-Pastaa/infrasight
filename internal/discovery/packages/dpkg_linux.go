//go:build linux

package packages

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
	"github.com/Aakhri-Pastaa/infrasight/pkg/shell"
)

// Tab/newline are real characters here; dpkg-query emits them verbatim.
const dpkgFormat = "-f=${Package}\t${Version}\t${Architecture}\t${Installed-Size}\n"

func (d *Dpkg) Available() bool {
	return shell.Available("dpkg-query")
}

func (d *Dpkg) Probe(ctx context.Context) (*discovery.Result, error) {
	out, err := shell.Run(ctx, d.Timeout(), "dpkg-query", "-W", dpkgFormat)
	if err != nil && out == "" {
		return nil, err
	}

	res := &discovery.Result{}
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) < 2 || cols[0] == "" {
			continue
		}

		name, ver := cols[0], cols[1]
		meta := map[string]any{"manager": "dpkg"}
		if len(cols) >= 3 && cols[2] != "" {
			meta["arch"] = cols[2]
		}
		if len(cols) >= 4 {
			if kb, e := strconv.ParseInt(strings.TrimSpace(cols[3]), 10, 64); e == nil {
				meta["installedSizeKB"] = kb
			}
		}

		res.Nodes = append(res.Nodes, graph.Node{
			ID:           "package:apt:" + name,
			Type:         graph.NodePackage,
			Label:        name,
			Version:      ver,
			Status:       graph.StatusActive,
			Health:       graph.HealthHealthy,
			Metadata:     meta,
			DiscoveredAt: time.Now().UTC(),
			DiscoveredBy: d.Name(),
		})
	}
	return res, nil
}
