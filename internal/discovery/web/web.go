// Package web discovers web servers and reverse proxies (nginx, Apache) and the
// virtual hosts, TLS certificates, and upstreams they declare. It turns those
// into WEBSITE / CERTIFICATE / PORT nodes and the edges that connect them to the
// rest of the graph:
//
//	process --LISTENS_ON--> port --SERVES--> website --PROXIES_TO--> port(upstream)
//	                                          cert    --ENCRYPTS-->  website
package web

import (
	"strings"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/certinfo"
	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
)

// vhost is a normalized virtual host extracted from a server config, server
// being "nginx" or "apache".
type vhost struct {
	server   string
	names    []string
	ports    []string
	tls      bool
	certPath string
	root     string
	proxies  []upstream
	config   string
}

// upstream is a parsed proxy target (proxy_pass / ProxyPass).
type upstream struct {
	raw   string
	host  string
	port  string
	local bool
}

// Cert health thresholds, in days until expiry.
const (
	certWarnDays = 30
	certCritDays = 0
)

// buildVHost appends the nodes and edges for a single vhost to res. certCache is
// shared across vhosts so a cert referenced by several sites is parsed once.
func buildVHost(res *discovery.Result, vh vhost, by string, now time.Time, certCache map[string]*certinfo.Info) {
	primary := primaryName(vh.names)
	if primary == "" {
		return // catch-all / unnamed server block — nothing to anchor on
	}
	websiteID := "website:" + primary

	meta := map[string]any{
		"server": vh.server,
		"names":  vh.names,
		"ports":  vh.ports,
		"tls":    vh.tls,
	}
	if vh.root != "" {
		meta["root"] = vh.root
	}
	if vh.config != "" {
		meta["config"] = vh.config
	}
	res.Nodes = append(res.Nodes, graph.Node{
		ID:           websiteID,
		Type:         graph.NodeWebsite,
		Label:        primary,
		Status:       graph.StatusActive,
		Health:       graph.HealthHealthy,
		Metadata:     meta,
		DiscoveredAt: now,
		DiscoveredBy: by,
	})

	// Listen ports serve the website. These PORT nodes merge with the ones from
	// network.ports, so a listening server process reaches its vhosts.
	for _, p := range vh.ports {
		portID := "port:tcp:" + p
		res.Nodes = append(res.Nodes, portNode(portID, p, vh.server, now, by))
		res.Edges = append(res.Edges, graph.Edge{Source: portID, Target: websiteID, Relation: graph.RelServes})
	}

	// TLS certificate.
	if vh.certPath != "" {
		if info := loadCert(vh.certPath, certCache); info != nil {
			res.Nodes = append(res.Nodes, certNode(info, vh.certPath, now, by))
			res.Edges = append(res.Edges, graph.Edge{
				Source:   "cert:" + certID(info),
				Target:   websiteID,
				Relation: graph.RelEncrypts,
			})
		} else {
			res.Warnings = append(res.Warnings, by+": could not read certificate "+vh.certPath)
		}
	}

	// Upstreams: a local proxy target becomes a PORT node the website proxies to,
	// which then connects to whatever process listens there.
	for _, up := range vh.proxies {
		if up.local && up.port != "" {
			portID := "port:tcp:" + up.port
			res.Nodes = append(res.Nodes, portNode(portID, up.port, "upstream", now, by))
			res.Edges = append(res.Edges, graph.Edge{
				Source:   websiteID,
				Target:   portID,
				Relation: graph.RelProxiesTo,
				Metadata: map[string]any{"upstream": up.raw},
			})
		}
	}
}

func portNode(id, port, source string, now time.Time, by string) graph.Node {
	return graph.Node{
		ID:           id,
		Type:         graph.NodePort,
		Label:        ":" + port,
		Status:       graph.StatusActive,
		Health:       graph.HealthHealthy,
		Metadata:     map[string]any{"protocol": "tcp", "port": port, "source": source},
		DiscoveredAt: now,
		DiscoveredBy: by,
	}
}

func certNode(info *certinfo.Info, path string, now time.Time, by string) graph.Node {
	days := info.DaysUntilExpiry(now)
	health := graph.HealthHealthy
	switch {
	case days < certCritDays:
		health = graph.HealthCritical
	case days < certWarnDays:
		health = graph.HealthWarning
	}
	label := info.CommonName
	if label == "" && len(info.DNSNames) > 0 {
		label = info.DNSNames[0]
	}
	if label == "" {
		label = "certificate"
	}
	return graph.Node{
		ID:     "cert:" + certID(info),
		Type:   graph.NodeCertificate,
		Label:  label,
		Status: graph.StatusActive,
		Health: health,
		Metadata: map[string]any{
			"subject":           info.Subject,
			"issuer":            info.Issuer,
			"serial":            info.Serial,
			"notBefore":         info.NotBefore.Format(time.RFC3339),
			"notAfter":          info.NotAfter.Format(time.RFC3339),
			"daysUntilExpiry":   days,
			"sans":              info.DNSNames,
			"keyType":           info.KeyType,
			"keyBits":           info.KeyBits,
			"path":              path,
			"fingerprintSHA256": info.SHA256,
		},
		DiscoveredAt: now,
		DiscoveredBy: by,
	}
}

func certID(info *certinfo.Info) string {
	if len(info.SHA256) >= 16 {
		return info.SHA256[:16]
	}
	return info.SHA256
}

func loadCert(path string, cache map[string]*certinfo.Info) *certinfo.Info {
	if v, ok := cache[path]; ok {
		return v
	}
	info, err := certinfo.ParseFile(path)
	if err != nil {
		cache[path] = nil
		return nil
	}
	cache[path] = info
	return info
}

// primaryName returns the first usable server name, skipping nginx catch-alls.
func primaryName(names []string) string {
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" || n == "_" {
			continue
		}
		return n
	}
	return ""
}

// parseUpstream parses a proxy target such as "http://127.0.0.1:3000/api" into
// its host and port, marking loopback/unspecified hosts as local.
func parseUpstream(raw string) upstream {
	up := upstream{raw: raw}
	s := raw
	if i := strings.Index(s, "://"); i >= 0 {
		s = s[i+3:]
	}
	if i := strings.IndexAny(s, "/;"); i >= 0 {
		s = s[:i]
	}
	host, port := s, ""
	if i := strings.LastIndex(s, ":"); i >= 0 {
		host, port = s[:i], s[i+1:]
	}
	up.host, up.port = host, port
	switch host {
	case "127.0.0.1", "localhost", "::1", "[::1]", "0.0.0.0", "":
		up.local = true
	}
	if !isNumeric(port) {
		up.port = ""
	}
	return up
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func appendUnique(xs []string, s string) []string {
	for _, x := range xs {
		if x == s {
			return xs
		}
	}
	return append(xs, s)
}
