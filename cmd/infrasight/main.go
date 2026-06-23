// Command infrasight is a zero-config, read-only host discovery agent.
//
// It auto-discovers compute resources, services, packages, and network
// endpoints on the local host and renders them as a dependency graph.
// Every probe is non-destructive: files are opened O_RDONLY and only
// well-known read-only commands are executed.
package main

import (
	"os"

	"github.com/Aakhri-Pastaa/infrasight/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
