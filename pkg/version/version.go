// Package version holds build-time version metadata, injected via -ldflags.
package version

var (
	// Version is the semantic version (overridden at build time).
	Version = "0.1.0-dev"
	// Commit is the short git SHA of the build.
	Commit = "unknown"
	// Date is the RFC3339 build timestamp.
	Date = "unknown"
)

// String returns a human-readable version line.
func String() string {
	return Version + " (commit " + Commit + ", built " + Date + ")"
}
