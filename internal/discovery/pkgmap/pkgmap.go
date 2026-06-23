// Package pkgmap resolves which installed package owns a file on disk, so that
// modules can link the entities they discover (a running process, a systemd
// unit) back to the PACKAGE node that provides them.
package pkgmap

// NodeID returns the graph node ID for a package name. It must match the ID
// scheme used by the packages.dpkg module so cross-linking can find the node.
func NodeID(pkg string) string { return "package:apt:" + pkg }
