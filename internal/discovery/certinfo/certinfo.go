// Package certinfo parses X.509 certificates from PEM (or raw DER) using the
// standard library, so InfraSight can describe TLS certs and flag expiry
// without shelling out to openssl.
package certinfo

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"os"
	"time"
)

// Info is the subset of certificate fields InfraSight surfaces.
type Info struct {
	Subject    string
	CommonName string
	Issuer     string
	Serial     string
	NotBefore  time.Time
	NotAfter   time.Time
	DNSNames   []string
	KeyType    string
	KeyBits    int
	SHA256     string // hex fingerprint of the DER
}

// DaysUntilExpiry returns whole days from now until NotAfter (negative if the
// certificate has already expired).
func (i *Info) DaysUntilExpiry(now time.Time) int {
	return int(i.NotAfter.Sub(now).Hours() / 24)
}

// ParseFile reads and parses the leaf certificate in a PEM file.
func ParseFile(path string) (*Info, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}

// Parse extracts the first certificate from PEM data (falling back to raw DER).
func Parse(data []byte) (*Info, error) {
	der := firstCertDER(data)
	if der == nil {
		der = data // maybe it is already DER
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, err
	}

	info := &Info{
		Subject:    cert.Subject.String(),
		CommonName: cert.Subject.CommonName,
		Issuer:     cert.Issuer.String(),
		Serial:     cert.SerialNumber.String(),
		NotBefore:  cert.NotBefore,
		NotAfter:   cert.NotAfter,
		DNSNames:   cert.DNSNames,
	}
	sum := sha256.Sum256(cert.Raw)
	info.SHA256 = hex.EncodeToString(sum[:])
	info.KeyType, info.KeyBits = keyInfo(cert)
	return info, nil
}

func firstCertDER(data []byte) []byte {
	rest := data
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			return nil
		}
		if block.Type == "CERTIFICATE" {
			return block.Bytes
		}
	}
}

func keyInfo(cert *x509.Certificate) (string, int) {
	switch k := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return "RSA", k.N.BitLen()
	case *ecdsa.PublicKey:
		return "ECDSA", k.Curve.Params().BitSize
	case ed25519.PublicKey:
		return "Ed25519", 256
	default:
		return "unknown", 0
	}
}

// ErrNoCertificate is returned when no certificate could be decoded.
var ErrNoCertificate = errors.New("certinfo: no certificate found")
