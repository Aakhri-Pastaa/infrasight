package web

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/certinfo"
)

const nginxSample = `
http {
  server {
    listen 80;
    server_name api.example.com www.api.example.com;
    return 301 https://$host$request_uri;
  }
  server {
    listen 443 ssl;  # TLS vhost
    server_name api.example.com;
    ssl_certificate /etc/ssl/api.pem;
    root /var/www/api;
    location / {
      proxy_pass http://127.0.0.1:3000;
    }
  }
}
`

func TestParseNginx(t *testing.T) {
	vhosts := parseNginx(nginxSample)
	if len(vhosts) != 2 {
		t.Fatalf("want 2 vhosts, got %d", len(vhosts))
	}
	var ssl *vhost
	for i := range vhosts {
		if vhosts[i].tls {
			ssl = &vhosts[i]
		}
	}
	if ssl == nil {
		t.Fatal("no TLS vhost parsed")
	}
	if primaryName(ssl.names) != "api.example.com" {
		t.Errorf("server_name = %v", ssl.names)
	}
	if len(ssl.ports) != 1 || ssl.ports[0] != "443" {
		t.Errorf("ports = %v, want [443]", ssl.ports)
	}
	if ssl.certPath != "/etc/ssl/api.pem" {
		t.Errorf("certPath = %q", ssl.certPath)
	}
	if len(ssl.proxies) != 1 || ssl.proxies[0].port != "3000" || !ssl.proxies[0].local {
		t.Errorf("proxies = %+v", ssl.proxies)
	}
}

const apacheSample = `
<VirtualHost *:443>
    ServerName shop.example.com
    ServerAlias www.shop.example.com   # alias
    DocumentRoot "/var/www/shop"
    SSLEngine on
    SSLCertificateFile /etc/ssl/shop.pem
    ProxyPass / http://127.0.0.1:8080/
</VirtualHost>
`

func TestParseApache(t *testing.T) {
	vhosts := parseApache(apacheSample)
	if len(vhosts) != 1 {
		t.Fatalf("want 1 vhost, got %d", len(vhosts))
	}
	v := vhosts[0]
	if primaryName(v.names) != "shop.example.com" {
		t.Errorf("ServerName = %v", v.names)
	}
	if !v.tls {
		t.Error("expected tls=true")
	}
	if len(v.ports) != 1 || v.ports[0] != "443" {
		t.Errorf("ports = %v, want [443]", v.ports)
	}
	if v.certPath != "/etc/ssl/shop.pem" {
		t.Errorf("certPath = %q", v.certPath)
	}
	if len(v.proxies) != 1 || v.proxies[0].port != "8080" {
		t.Errorf("proxies = %+v", v.proxies)
	}
}

func TestBuildVHostEmitsChain(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "api.pem")
	// Expires in ~10 days -> should be classified "warning".
	if err := os.WriteFile(certPath, testCertPEM(t, "api.example.com", 10*24*time.Hour), 0o600); err != nil {
		t.Fatal(err)
	}

	vh := vhost{
		server:   "nginx",
		names:    []string{"api.example.com"},
		ports:    []string{"443"},
		tls:      true,
		certPath: certPath,
		proxies:  []upstream{{raw: "http://127.0.0.1:3000", host: "127.0.0.1", port: "3000", local: true}},
	}
	res := &discovery.Result{}
	buildVHost(res, vh, "web.nginx", time.Now().UTC(), map[string]*certinfo.Info{})

	types := map[string]int{}
	var certHealth string
	for _, n := range res.Nodes {
		types[string(n.Type)]++
		if n.Type == "CERTIFICATE" {
			certHealth = string(n.Health)
		}
	}
	// website + cert + listen port (443) + upstream port (3000)
	if types["WEBSITE"] != 1 || types["CERTIFICATE"] != 1 || types["PORT"] != 2 {
		t.Fatalf("node types = %v", types)
	}
	if certHealth != "warning" {
		t.Errorf("cert health = %q, want warning (expires in ~10 days)", certHealth)
	}

	rels := map[string]int{}
	for _, e := range res.Edges {
		rels[string(e.Relation)]++
	}
	if rels["SERVES"] != 1 || rels["ENCRYPTS"] != 1 || rels["PROXIES_TO"] != 1 {
		t.Fatalf("edge relations = %v, want one each of SERVES/ENCRYPTS/PROXIES_TO", rels)
	}
}

func testCertPEM(t *testing.T, cn string, validFor time.Duration) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: cn},
		DNSNames:     []string{cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(validFor),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}
