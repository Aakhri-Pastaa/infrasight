//go:build linux

package pkgmap

import (
	"context"
	"strings"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/pkg/shell"
)

// Owner returns the dpkg package that owns the file at path, if any. It is a
// best-effort lookup: on non-dpkg systems, unreadable paths, or files not owned
// by any package it simply returns ok=false.
func Owner(ctx context.Context, path string) (string, bool) {
	if path == "" || !shell.Available("dpkg-query") {
		return "", false
	}
	out, err := shell.Run(ctx, 5*time.Second, "dpkg-query", "-S", path)
	if err != nil || out == "" {
		return "", false
	}

	// First match looks like "pkg1, pkg2: /the/path"; take the first package.
	line := out
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	i := strings.LastIndex(line, ": ")
	if i < 0 {
		return "", false
	}
	pkg := line[:i]
	if j := strings.IndexByte(pkg, ','); j >= 0 {
		pkg = pkg[:j]
	}
	pkg = strings.TrimSpace(pkg)
	// Drop a ":arch" multiarch qualifier (e.g. "libc6:amd64" -> "libc6") so it
	// matches the ${Package} name recorded by the packages.dpkg module.
	if c := strings.IndexByte(pkg, ':'); c >= 0 {
		pkg = pkg[:c]
	}
	if pkg == "" {
		return "", false
	}
	return pkg, true
}
