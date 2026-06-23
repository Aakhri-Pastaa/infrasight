// Package pkgbackend abstracts the host package manager (dpkg, rpm, apk, pacman)
// behind one interface, so package discovery and file->package cross-linking work
// across Linux distributions. The backend present on the host is selected at
// runtime via Detect().
//
// Parsing is kept in pure functions (parse.go) separate from command execution,
// so every backend's output handling is unit-tested with recorded fixtures —
// the rpm/apk/pacman paths are validated even on a host that only has dpkg.
package pkgbackend

import (
	"context"
	"time"
)

// Package is one installed package.
type Package struct {
	Name    string
	Version string
	Arch    string
	SizeKB  int64
}

// Backend is one package manager. Implementations live in build-tagged files;
// available() lets Detect() pick the one present on this host.
type Backend interface {
	// Key namespaces this manager's package node IDs (e.g. "apt", "rpm").
	Key() string
	available() bool
	// List returns the installed packages.
	List(ctx context.Context, timeout time.Duration) ([]Package, error)
	// Owner returns the package that owns the file at path, if any.
	Owner(ctx context.Context, path string) (name string, ok bool)
}

// NodeID is the graph node ID for a package, namespaced by manager key so the
// inventory module and the cross-link owner-lookup produce matching IDs.
func NodeID(key, name string) string { return "package:" + key + ":" + name }

// OwnerNode resolves the package owning a file and returns its name and graph
// node ID. ok is false when there is no backend or no owning package.
func OwnerNode(ctx context.Context, path string) (name, id string, ok bool) {
	b := Detect()
	if b == nil {
		return "", "", false
	}
	n, ok := b.Owner(ctx, path)
	if !ok {
		return "", "", false
	}
	return n, NodeID(b.Key(), n), true
}
