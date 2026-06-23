//go:build !linux

package pkgmap

import "context"

// Owner is unsupported off Linux; package ownership maps to dpkg today.
func Owner(_ context.Context, _ string) (string, bool) { return "", false }
