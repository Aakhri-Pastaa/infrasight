package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Aakhri-Pastaa/infrasight/pkg/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "infrasight "+version.String())
			return nil
		},
	}
}
