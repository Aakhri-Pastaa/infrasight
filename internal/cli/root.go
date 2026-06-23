// Package cli wires the Cobra command tree for the infrasight binary.
package cli

import "github.com/spf13/cobra"

// exitCode is set by leaf commands (e.g. scan) and returned by Execute so the
// process can signal warnings (1) and critical findings (2) to callers/CI.
var exitCode int

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "infrasight",
		Short: "Read-only host discovery agent",
		Long: "InfraSight auto-discovers compute resources, services, packages and\n" +
			"network endpoints on a host and renders them as a dependency graph.\n" +
			"Every probe is non-destructive and read-only.",
		SilenceUsage: true,
	}
	root.AddCommand(newScanCmd(), newDiffCmd(), newVersionCmd())
	return root
}

// Execute runs the root command and returns a process exit code.
func Execute() int {
	if err := newRootCmd().Execute(); err != nil {
		if exitCode == 0 {
			exitCode = 1
		}
	}
	return exitCode
}
