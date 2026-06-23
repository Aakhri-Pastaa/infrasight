//go:build linux

package services

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/pkgbackend"
	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
	"github.com/Aakhri-Pastaa/infrasight/pkg/shell"
)

func (s *Systemd) Available() bool {
	if !shell.Available("systemctl") {
		return false
	}
	// The canonical "is systemd actually running as PID 1" check.
	fi, err := os.Stat("/run/systemd/system")
	return err == nil && fi.IsDir()
}

func (s *Systemd) Probe(ctx context.Context) (*discovery.Result, error) {
	list, err := shell.Run(ctx, s.Timeout(), "systemctl",
		"list-units", "--type=service", "--state=running",
		"--no-legend", "--plain", "--no-pager")
	if err != nil && list == "" {
		return nil, err
	}

	var units []string
	for _, line := range strings.Split(list, "\n") {
		f := strings.Fields(line)
		if len(f) > 0 && strings.HasSuffix(f[0], ".service") {
			units = append(units, f[0])
		}
	}

	res := &discovery.Result{}
	if len(units) == 0 {
		return res, nil
	}

	// One batched `systemctl show` for every unit; blocks are blank-separated.
	args := append([]string{"show", "--no-pager",
		"--property=Id,MainPID,FragmentPath,ActiveState,SubState,MemoryCurrent,Description"},
		units...)
	show, err := shell.Run(ctx, s.Timeout(), "systemctl", args...)
	if err != nil && show == "" {
		return nil, err
	}

	pkgCache := map[string][2]string{} // path -> {name, nodeID}
	owner := func(path string) (name, id string, ok bool) {
		if path == "" {
			return "", "", false
		}
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

	for _, block := range strings.Split(show, "\n\n") {
		kv := parseShow(block)
		id := kv["Id"]
		if id == "" {
			continue
		}

		serviceID := "service:systemd:" + id
		meta := map[string]any{
			"activeState": kv["ActiveState"],
			"subState":    kv["SubState"],
		}
		if kv["FragmentPath"] != "" {
			meta["fragmentPath"] = kv["FragmentPath"]
		}
		if kv["Description"] != "" {
			meta["description"] = kv["Description"]
		}
		if mem, ok := parseMemory(kv["MemoryCurrent"]); ok {
			meta["memoryBytes"] = mem
		}

		mainPid := strings.TrimSpace(kv["MainPID"])
		hasPid := mainPid != "" && mainPid != "0"
		if hasPid {
			meta["mainPid"] = mainPid
		}

		// The package that ships the unit file (links service -> package).
		svcPkg, svcPkgID, svcHasPkg := owner(kv["FragmentPath"])
		if svcHasPkg {
			meta["package"] = svcPkg
			meta["packageId"] = svcPkgID
		}

		res.Nodes = append(res.Nodes, graph.Node{
			ID:           serviceID,
			Type:         graph.NodeService,
			Label:        id,
			Status:       graph.StatusActive,
			Health:       graph.HealthHealthy,
			Metadata:     meta,
			DiscoveredAt: time.Now().UTC(),
			DiscoveredBy: s.Name(),
		})

		if !hasPid {
			continue
		}

		// Emit the service's main process and a SERVICE -> PROCESS edge. If the
		// process also listens on a port, this node merges with the one from
		// network.ports (same ID), connecting services to live ports.
		procID := "process:" + mainPid
		pmeta := map[string]any{"pid": mainPid}
		label := id
		if comm := readComm(mainPid); comm != "" {
			pmeta["name"] = comm
			label = comm
		}
		if exe, e := os.Readlink("/proc/" + mainPid + "/exe"); e == nil && exe != "" {
			pmeta["exe"] = exe
			name, id, ok := owner(exe)
			if !ok && svcHasPkg {
				name, id, ok = svcPkg, svcPkgID, true
			}
			if ok {
				pmeta["package"] = name
				pmeta["packageId"] = id
			}
		} else if svcHasPkg {
			pmeta["package"] = svcPkg
			pmeta["packageId"] = svcPkgID
		}

		res.Nodes = append(res.Nodes, graph.Node{
			ID:           procID,
			Type:         graph.NodeProcess,
			Label:        label,
			Status:       graph.StatusActive,
			Health:       graph.HealthHealthy,
			Metadata:     pmeta,
			DiscoveredAt: time.Now().UTC(),
			DiscoveredBy: s.Name(),
		})
		res.Edges = append(res.Edges, graph.Edge{
			Source:   serviceID,
			Target:   procID,
			Relation: graph.RelContains,
		})
	}
	return res, nil
}

// parseShow parses a block of `systemctl show` KEY=VALUE lines.
func parseShow(block string) map[string]string {
	m := make(map[string]string)
	for _, line := range strings.Split(block, "\n") {
		k, v, ok := strings.Cut(line, "=")
		if ok {
			m[strings.TrimSpace(k)] = v
		}
	}
	return m
}

// parseMemory parses MemoryCurrent; systemd reports uint64-max for "not set",
// which overflows int64 and is treated as absent.
func parseMemory(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" || s == "[not set]" {
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

func readComm(pid string) string {
	b, err := os.ReadFile("/proc/" + pid + "/comm")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
