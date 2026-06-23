package web

import (
	"strings"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

// Apache discovers Apache httpd virtual hosts, their ports, TLS certs and
// reverse-proxy targets by parsing the config files.
type Apache struct{}

// NewApache constructs the apache module.
func NewApache() *Apache { return &Apache{} }

func (a *Apache) Name() string            { return "web.apache" }
func (a *Apache) Description() string     { return "Apache virtual hosts, TLS certs and proxies" }
func (a *Apache) RequiredTools() []string { return nil }
func (a *Apache) RequiresRoot() bool      { return false }
func (a *Apache) RequiresNetwork() bool   { return false }
func (a *Apache) Timeout() time.Duration  { return 15 * time.Second }
func (a *Apache) Risk() discovery.Risk    { return discovery.RiskLow }

// --- pure config parsing (testable without apache installed) ---

// parseApache extracts virtual hosts from Apache config text. It is line based
// and tracks <VirtualHost> ... </VirtualHost> sections.
func parseApache(content string) []vhost {
	var out []vhost
	var cur *vhost

	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(stripHash(raw))
		if line == "" {
			continue
		}
		low := strings.ToLower(line)

		switch {
		case strings.HasPrefix(low, "<virtualhost"):
			vh := vhost{server: "apache"}
			if p := apachePortFromTag(line); p != "" {
				vh.ports = appendUnique(vh.ports, p)
			}
			cur = &vh
		case strings.HasPrefix(low, "</virtualhost"):
			if cur != nil {
				if cur.certPath != "" {
					cur.tls = true
				}
				out = append(out, *cur)
				cur = nil
			}
		case cur != nil:
			applyApacheDirective(cur, strings.Fields(line))
		}
	}
	return out
}

func applyApacheDirective(vh *vhost, fields []string) {
	if len(fields) == 0 {
		return
	}
	switch strings.ToLower(fields[0]) {
	case "servername":
		if len(fields) > 1 {
			vh.names = append(vh.names, stripPortFromName(unquote(fields[1])))
		}
	case "serveralias":
		for _, f := range fields[1:] {
			vh.names = append(vh.names, unquote(f))
		}
	case "documentroot":
		if len(fields) > 1 {
			vh.root = unquote(fields[1])
		}
	case "sslcertificatefile":
		if len(fields) > 1 && vh.certPath == "" {
			vh.certPath = unquote(fields[1])
		}
	case "sslengine":
		if len(fields) > 1 && strings.EqualFold(fields[1], "on") {
			vh.tls = true
		}
	case "proxypass":
		for _, f := range fields[1:] {
			if strings.Contains(f, "://") {
				vh.proxies = append(vh.proxies, parseUpstream(f))
				break
			}
		}
	}
}

// apachePortFromTag pulls the port out of "<VirtualHost *:443>" or
// "<VirtualHost 127.0.0.1:8080>".
func apachePortFromTag(line string) string {
	line = strings.TrimSuffix(strings.TrimSpace(line), ">")
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return ""
	}
	addr := fields[1] // e.g. "*:443"
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		addr = addr[i+1:]
	}
	if isNumeric(addr) {
		return addr
	}
	return ""
}

func stripHash(line string) string {
	if i := strings.IndexByte(line, '#'); i >= 0 {
		return line[:i]
	}
	return line
}

func stripPortFromName(name string) string {
	if i := strings.IndexByte(name, ':'); i >= 0 {
		return name[:i]
	}
	return name
}

func unquote(s string) string {
	return strings.Trim(s, `"'`)
}
