package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Aakhri-Pastaa/infrasight/internal/diff"
	"github.com/Aakhri-Pastaa/infrasight/internal/output"
	"github.com/Aakhri-Pastaa/infrasight/internal/store"
)

func newDiffCmd() *cobra.Command {
	var noColor bool
	cmd := &cobra.Command{
		Use:   "diff <baseline> <current>",
		Short: "Show what changed between two scans",
		Long: "Compare two scan JSON files (or saved scan names) and report the drift:\n" +
			"nodes added/removed, version changes, and health changes.\n\n" +
			"Each argument is a path to a scan JSON, or the name of a scan saved with\n" +
			"`infrasight scan --save <name>`.\n\n" +
			"Exit code: 0 no drift, 1 drift, 2 a regression to critical health.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			baseline, err := loadScan(args[0])
			if err != nil {
				return err
			}
			current, err := loadScan(args[1])
			if err != nil {
				return err
			}

			result := diff.Compare(baseline, current)
			color := !noColor && isTTY(os.Stdout)
			diff.RenderTerminal(cmd.OutOrStdout(), baseline, current, result, color)

			switch {
			case result.Empty():
				exitCode = 0
			case result.HasCriticalRegression():
				exitCode = 2
			default:
				exitCode = 1
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&noColor, "no-color", false, "disable coloured output")
	return cmd
}

func loadScan(arg string) (output.Document, error) {
	path := store.Resolve(arg)
	data, err := os.ReadFile(path)
	if err != nil {
		return output.Document{}, fmt.Errorf("reading scan %q: %w", arg, err)
	}
	var doc output.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return output.Document{}, fmt.Errorf("parsing scan %q: %w", path, err)
	}
	return doc, nil
}
