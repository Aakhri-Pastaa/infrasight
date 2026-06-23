//go:build linux

package web

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/certinfo"
)

// nginxConfigEnv lets users point at a non-standard config location (also used
// by tests). It may be a single file or a directory of config files.
const nginxConfigEnv = "INFRASIGHT_NGINX_CONF"

func (n *Nginx) configFiles() []string {
	var paths []string
	if v := os.Getenv(nginxConfigEnv); v != "" {
		paths = append(paths, expandConfig(v)...)
	}
	paths = append(paths, "/etc/nginx/nginx.conf")
	paths = append(paths, globFiles("/etc/nginx/sites-enabled/*")...)
	paths = append(paths, globFiles("/etc/nginx/conf.d/*.conf")...)
	return existingFiles(paths)
}

func (n *Nginx) Available() bool {
	return len(n.configFiles()) > 0
}

func (n *Nginx) Probe(ctx context.Context) (*discovery.Result, error) {
	res := &discovery.Result{}
	now := time.Now().UTC()
	certCache := map[string]*certinfo.Info{}

	for _, path := range n.configFiles() {
		data, err := os.ReadFile(path)
		if err != nil {
			res.Warnings = append(res.Warnings, "web.nginx: cannot read "+path)
			continue
		}
		for _, vh := range parseNginx(string(data)) {
			vh.config = path
			buildVHost(res, vh, n.Name(), now, certCache)
		}
	}
	return res, nil
}

// expandConfig turns a file or directory path into a list of config files.
func expandConfig(path string) []string {
	fi, err := os.Stat(path)
	if err != nil {
		return nil
	}
	if fi.IsDir() {
		return globFiles(filepath.Join(path, "*"))
	}
	return []string{path}
}

func globFiles(pattern string) []string {
	m, err := filepath.Glob(pattern)
	if err != nil {
		return nil
	}
	return m
}

func existingFiles(paths []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range paths {
		if p == "" || seen[p] {
			continue
		}
		fi, err := os.Stat(p)
		if err != nil || fi.IsDir() {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}
