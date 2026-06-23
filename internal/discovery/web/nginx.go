package web

import (
	"strings"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

// Nginx discovers nginx virtual hosts, their listen ports, TLS certificates and
// upstreams by parsing the config files (no nginx binary required).
type Nginx struct{}

// NewNginx constructs the nginx module.
func NewNginx() *Nginx { return &Nginx{} }

func (n *Nginx) Name() string            { return "web.nginx" }
func (n *Nginx) Description() string     { return "nginx virtual hosts, TLS certs and upstreams" }
func (n *Nginx) RequiredTools() []string { return nil }
func (n *Nginx) RequiresRoot() bool      { return false }
func (n *Nginx) RequiresNetwork() bool   { return false }
func (n *Nginx) Timeout() time.Duration  { return 15 * time.Second }
func (n *Nginx) Risk() discovery.Risk    { return discovery.RiskLow }

// --- pure config parsing (testable without nginx installed) ---

// parseNginx extracts virtual hosts from nginx config text. It understands
// nested server/location blocks but intentionally ignores include directives;
// callers feed each config file (sites-enabled/*, conf.d/*) directly.
func parseNginx(content string) []vhost {
	var out []vhost
	collectServers(statements(tokenize(content)), &out)
	return out
}

func collectServers(stmts []stmt, out *[]vhost) {
	for _, s := range stmts {
		switch s.name {
		case "server":
			*out = append(*out, parseServer(s.block))
		case "http", "stream":
			collectServers(statements(s.block), out)
		}
	}
}

func parseServer(block []token) vhost {
	vh := vhost{server: "nginx"}
	for _, s := range statements(block) {
		switch s.name {
		case "server_name":
			vh.names = append(vh.names, s.args...)
		case "listen":
			if len(s.args) > 0 {
				if p := portFromListen(s.args[0]); p != "" {
					vh.ports = appendUnique(vh.ports, p)
				}
				for _, a := range s.args {
					if a == "ssl" {
						vh.tls = true
					}
				}
			}
		case "ssl_certificate":
			if len(s.args) > 0 && vh.certPath == "" {
				vh.certPath = s.args[0]
			}
		case "root":
			if len(s.args) > 0 {
				vh.root = s.args[0]
			}
		case "location":
			for _, ls := range statements(s.block) {
				if ls.name == "proxy_pass" && len(ls.args) > 0 {
					vh.proxies = append(vh.proxies, parseUpstream(ls.args[0]))
				}
			}
		}
	}
	if vh.certPath != "" {
		vh.tls = true
	}
	return vh
}

// portFromListen extracts the numeric port from a listen value such as "443",
// "127.0.0.1:8080", "[::]:443" or "*:80".
func portFromListen(v string) string {
	if i := strings.LastIndex(v, ":"); i >= 0 {
		v = v[i+1:]
	}
	if isNumeric(v) {
		return v
	}
	return ""
}

// --- tokenizer ---

const (
	tWord = iota
	tLBrace
	tRBrace
	tSemi
)

type token struct {
	kind int
	val  string
}

type stmt struct {
	name  string
	args  []string
	block []token // inner tokens when the statement was a { } block
}

func tokenize(s string) []token {
	var toks []token
	i, n := 0, len(s)
	for i < n {
		c := s[i]
		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n':
			i++
		case c == '#':
			for i < n && s[i] != '\n' {
				i++
			}
		case c == '{':
			toks = append(toks, token{tLBrace, "{"})
			i++
		case c == '}':
			toks = append(toks, token{tRBrace, "}"})
			i++
		case c == ';':
			toks = append(toks, token{tSemi, ";"})
			i++
		case c == '"' || c == '\'':
			quote := c
			i++
			start := i
			for i < n && s[i] != quote {
				i++
			}
			toks = append(toks, token{tWord, s[start:i]})
			if i < n {
				i++ // closing quote
			}
		default:
			start := i
			for i < n {
				d := s[i]
				if d == ' ' || d == '\t' || d == '\r' || d == '\n' ||
					d == '{' || d == '}' || d == ';' || d == '#' {
					break
				}
				i++
			}
			toks = append(toks, token{tWord, s[start:i]})
		}
	}
	return toks
}

// statements splits a token slice into directives, capturing nested blocks.
func statements(toks []token) []stmt {
	var out []stmt
	i := 0
	for i < len(toks) {
		var words []string
		for i < len(toks) && toks[i].kind == tWord {
			words = append(words, toks[i].val)
			i++
		}
		if i >= len(toks) {
			break
		}
		switch toks[i].kind {
		case tSemi:
			i++
			if len(words) > 0 {
				out = append(out, stmt{name: words[0], args: words[1:]})
			}
		case tLBrace:
			depth := 1
			i++
			start := i
			for i < len(toks) && depth > 0 {
				switch toks[i].kind {
				case tLBrace:
					depth++
				case tRBrace:
					depth--
				}
				if depth == 0 {
					break
				}
				i++
			}
			block := toks[start:i]
			if i < len(toks) {
				i++ // consume closing brace
			}
			if len(words) > 0 {
				out = append(out, stmt{name: words[0], args: words[1:], block: block})
			}
		default: // stray '}' — skip
			i++
		}
	}
	return out
}
