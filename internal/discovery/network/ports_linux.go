//go:build linux

package network

import (
	"context"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/pkgbackend"
	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
	"github.com/Aakhri-Pastaa/infrasight/pkg/shell"
)

// ssProcess matches the leading process entry in ss's users:(("name",pid=N,...)) column.
var ssProcess = regexp.MustCompile(`\("([^"]+)",pid=(\d+)`)

func (p *Ports) Available() bool {
	return shell.Available("ss")
}

func (p *Ports) Probe(ctx context.Context) (*discovery.Result, error) {
	// -H no header, -t TCP, -l listening, -n numeric, -p show process.
	out, err := shell.Run(ctx, p.Timeout(), "ss", "-H", "-t", "-l", "-n", "-p")
	if err != nil && out == "" {
		return nil, err
	}

	res := &discovery.Result{}
	if err != nil {
		res.Warnings = append(res.Warnings, "network.ports: ss reported an error; results may be partial")
	}

	// Resolve a process binary to its owning package once per unique path.
	pkgCache := map[string][2]string{} // path -> {name, nodeID}
	owner := func(path string) (name, id string, ok bool) {
		if v, cached := pkgCache[path]; cached {
			return v[0], v[1], v[0] != ""
		}
		n, nodeID, found := pkgbackend.OwnerNode(ctx, path)
		if !found {
			pkgCache[path] = [2]string{}
			return "", "", false
		}
		pkgCache[path] = [2]string{n, nodeID}
		return n, nodeID, true
	}

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		// State Recv-Q Send-Q Local:Port Peer:Port [users:((...))]
		if len(fields) < 4 {
			continue
		}
		state := fields[0]
		bindAddr, port := splitHostPort(fields[3])
		if port == "" {
			continue
		}

		portID := "port:tcp:" + port
		res.Nodes = append(res.Nodes, graph.Node{
			ID:     portID,
			Type:   graph.NodePort,
			Label:  ":" + port,
			Status: graph.StatusActive,
			Health: graph.HealthHealthy,
			Metadata: map[string]any{
				"protocol":    "tcp",
				"port":        port,
				"bindAddress": bindAddr,
				"state":       state,
			},
			DiscoveredAt: time.Now().UTC(),
			DiscoveredBy: p.Name(),
		})

		if len(fields) >= 6 {
			if name, pid, ok := parseSSProcess(fields[5]); ok {
				procID := "process:" + pid
				meta := map[string]any{"pid": pid, "name": name}
				// Resolve the binary and the package that ships it so the
				// graph builder can link this process to its PACKAGE node.
				if exe, e := os.Readlink("/proc/" + pid + "/exe"); e == nil && exe != "" {
					meta["exe"] = exe
					if name, id, ok := owner(exe); ok {
						meta["package"] = name
						meta["packageId"] = id
					}
				}
				res.Nodes = append(res.Nodes, graph.Node{
					ID:           procID,
					Type:         graph.NodeProcess,
					Label:        name,
					Status:       graph.StatusActive,
					Health:       graph.HealthHealthy,
					Metadata:     meta,
					DiscoveredAt: time.Now().UTC(),
					DiscoveredBy: p.Name(),
				})
				res.Edges = append(res.Edges, graph.Edge{
					Source:   procID,
					Target:   portID,
					Relation: graph.RelListensOn,
				})
			}
		}
	}
	return res, nil
}

// splitHostPort splits "addr:port" on the final colon, handling IPv6 forms
// such as "[::]:22" and "*:8080".
func splitHostPort(s string) (host, port string) {
	i := strings.LastIndexByte(s, ':')
	if i < 0 {
		return s, ""
	}
	return s[:i], s[i+1:]
}

func parseSSProcess(field string) (name, pid string, ok bool) {
	m := ssProcess.FindStringSubmatch(field)
	if m == nil {
		return "", "", false
	}
	return m[1], m[2], true
}
