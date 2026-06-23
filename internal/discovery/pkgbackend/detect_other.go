//go:build !linux

package pkgbackend

// Detect returns nil off Linux; package managers here are Linux-only.
func Detect() Backend { return nil }
