//go:build linux

package database

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/pkgbackend"
	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
	"github.com/Aakhri-Pastaa/infrasight/pkg/shell"
)

func (m *Module) Available() bool {
	for _, e := range engines() {
		if detectConfig(e) != "" || firstBinary(e) != "" {
			return true
		}
	}
	return false
}

func (m *Module) Probe(ctx context.Context) (*discovery.Result, error) {
	res := &discovery.Result{}
	now := time.Now().UTC()

	for _, e := range engines() {
		cfg := detectConfig(e)
		bin := firstBinary(e)
		if cfg == "" && bin == "" {
			continue // engine not present
		}

		port := e.defaultPort
		if cfg != "" {
			if data, err := os.ReadFile(cfg); err == nil {
				if p := e.parsePort(string(data)); p != "" {
					port = p
				}
			}
		}

		meta := map[string]any{"engine": e.key, "port": port}
		if cfg != "" {
			meta["configPath"] = cfg
		}
		if bin != "" {
			meta["binary"] = bin
		}
		// Link to the owning package via the config file (or binary). Setting
		// packageId lets graph.CrossLink draw DATABASE -DEPENDS_ON-> PACKAGE.
		ownerPath := cfg
		if ownerPath == "" {
			ownerPath = bin
		}
		if name, id, ok := pkgbackend.OwnerNode(ctx, ownerPath); ok {
			meta["package"] = name
			meta["packageId"] = id
		}

		version := ""
		if bin != "" {
			version = binaryVersion(ctx, bin)
		}

		dbID := "database:" + e.key + ":" + port
		res.Nodes = append(res.Nodes, graph.Node{
			ID:           dbID,
			Type:         graph.NodeDatabase,
			Label:        e.name + " :" + port,
			Version:      version,
			Status:       graph.StatusActive,
			Health:       graph.HealthHealthy,
			Metadata:     meta,
			DiscoveredAt: now,
			DiscoveredBy: m.Name(),
		})

		// DATABASE -LISTENS_ON-> PORT; the PORT node merges (by ID) with the live
		// listener from network.ports, tying the DB to its process and package.
		portID := "port:tcp:" + port
		res.Nodes = append(res.Nodes, graph.Node{
			ID:     portID,
			Type:   graph.NodePort,
			Label:  ":" + port,
			Status: graph.StatusActive,
			Health: graph.HealthHealthy,
			Metadata: map[string]any{
				"protocol": "tcp",
				"port":     port,
				"source":   "database",
			},
			DiscoveredAt: now,
			DiscoveredBy: m.Name(),
		})
		res.Edges = append(res.Edges, graph.Edge{Source: dbID, Target: portID, Relation: graph.RelListensOn})
	}
	return res, nil
}

// detectConfig returns the config file to use: the env override if set and
// present, otherwise the first existing default config.
func detectConfig(e engine) string {
	if e.configEnv != "" {
		if v := os.Getenv(e.configEnv); v != "" && fileExists(v) {
			return v
		}
	}
	for _, glob := range e.configGlobs {
		matches, _ := filepath.Glob(glob)
		for _, p := range matches {
			if fileExists(p) {
				return p
			}
		}
	}
	return ""
}

func firstBinary(e engine) string {
	for _, b := range e.binaries {
		if p, err := exec.LookPath(b); err == nil {
			return p
		}
	}
	return ""
}

// binaryVersion runs "<bin> --version" (read-only) and extracts a version.
func binaryVersion(ctx context.Context, bin string) string {
	out, err := shell.Run(ctx, 5*time.Second, bin, "--version")
	if err != nil && out == "" {
		return ""
	}
	return reVersion.FindString(out)
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}
