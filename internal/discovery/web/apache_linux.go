//go:build linux

package web

import (
	"context"
	"os"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/certinfo"
)

const apacheConfigEnv = "INFRASIGHT_APACHE_CONF"

func (a *Apache) configFiles() []string {
	var paths []string
	if v := os.Getenv(apacheConfigEnv); v != "" {
		paths = append(paths, expandConfig(v)...)
	}
	paths = append(paths, globFiles("/etc/apache2/sites-enabled/*.conf")...)
	paths = append(paths, globFiles("/etc/httpd/conf.d/*.conf")...)
	return existingFiles(paths)
}

func (a *Apache) Available() bool {
	return len(a.configFiles()) > 0
}

func (a *Apache) Probe(ctx context.Context) (*discovery.Result, error) {
	res := &discovery.Result{}
	now := time.Now().UTC()
	certCache := map[string]*certinfo.Info{}

	for _, path := range a.configFiles() {
		data, err := os.ReadFile(path)
		if err != nil {
			res.Warnings = append(res.Warnings, "web.apache: cannot read "+path)
			continue
		}
		for _, vh := range parseApache(string(data)) {
			vh.config = path
			buildVHost(res, vh, a.Name(), now, certCache)
		}
	}
	return res, nil
}
