package certinfo

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func makePEM(t *testing.T, cn string, notAfter time.Time) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(12345),
		Subject:      pkix.Name{CommonName: cn},
		DNSNames:     []string{cn, "www." + cn},
		NotBefore:    notAfter.Add(-365 * 24 * time.Hour),
		NotAfter:     notAfter,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func TestParse(t *testing.T) {
	exp := time.Now().Add(40 * 24 * time.Hour)
	info, err := Parse(makePEM(t, "api.example.com", exp))
	if err != nil {
		t.Fatal(err)
	}
	if info.CommonName != "api.example.com" {
		t.Errorf("CommonName = %q", info.CommonName)
	}
	if len(info.DNSNames) != 2 {
		t.Errorf("DNSNames = %v", info.DNSNames)
	}
	if info.KeyType != "ECDSA" || info.KeyBits != 256 {
		t.Errorf("key = %s/%d, want ECDSA/256", info.KeyType, info.KeyBits)
	}
	if len(info.SHA256) != 64 {
		t.Errorf("fingerprint length = %d, want 64 hex chars", len(info.SHA256))
	}
	if d := info.DaysUntilExpiry(time.Now()); d < 38 || d > 40 {
		t.Errorf("DaysUntilExpiry = %d, want ~39", d)
	}
}

func TestParseExpired(t *testing.T) {
	info, err := Parse(makePEM(t, "old.example.com", time.Now().Add(-10*24*time.Hour)))
	if err != nil {
		t.Fatal(err)
	}
	if d := info.DaysUntilExpiry(time.Now()); d >= 0 {
		t.Errorf("DaysUntilExpiry = %d, want negative for an expired cert", d)
	}
}

func TestParseGarbage(t *testing.T) {
	if _, err := Parse([]byte("not a certificate")); err == nil {
		t.Error("expected an error parsing non-certificate data")
	}
}
