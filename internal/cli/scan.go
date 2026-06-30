package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/output"
	htmlout "github.com/Aakhri-Pastaa/infrasight/internal/output/html"
	jsonout "github.com/Aakhri-Pastaa/infrasight/internal/output/json"
	"github.com/Aakhri-Pastaa/infrasight/internal/registry"
	"github.com/Aakhri-Pastaa/infrasight/internal/security"
	"github.com/Aakhri-Pastaa/infrasight/internal/store"
	"github.com/Aakhri-Pastaa/infrasight/pkg/version"
)

type scanFlags struct {
	format   string
	output   string
	modules  []string
	exclude  []string
	parallel int
	timeout  time.Duration
	quiet    bool
	noColor  bool
	open     bool
	security bool
	deep     bool
	save     string
	redact   bool
}

func newScanCmd() *cobra.Command {
	f := &scanFlags{}
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Discover host resources and render a dependency graph",
		Long: "Scan runs every applicable discovery module concurrently and writes a\n" +
			"graph of the host as JSON and/or a self-contained HTML report.\n\n" +
			"Exit code: 0 clean, 1 warnings, 2 critical findings.",
		RunE: func(cmd *cobra.Command, _ []string) error { return runScan(cmd, f) },
	}

	fl := cmd.Flags()
	fl.StringVar(&f.format, "format", "all", "output format: json, html, all")
	fl.StringVarP(&f.output, "output", "o", "infrasight", "output path or basename (extension added per format)")
	fl.StringSliceVar(&f.modules, "modules", nil, "only run these modules (full name or domain, e.g. hardware)")
	fl.StringSliceVar(&f.exclude, "exclude-modules", nil, "skip these modules")
	fl.IntVar(&f.parallel, "parallel", 50, "max concurrent probe workers")
	fl.DurationVar(&f.timeout, "timeout", 5*time.Minute, "global scan timeout")
	fl.BoolVar(&f.quiet, "quiet", false, "suppress the terminal summary")
	fl.BoolVar(&f.noColor, "no-color", false, "disable coloured output")
	fl.BoolVar(&f.open, "open", false, "open the HTML report in a browser (not yet implemented)")
	fl.BoolVar(&f.security, "security", false, "enable the security audit module over the graph")
	fl.BoolVar(&f.deep, "deep", false, "deep inspection (current modules already probe fully)")
	fl.StringVar(&f.save, "save", "", "also save this scan as a named baseline for 'infrasight diff'")
	fl.BoolVar(&f.redact, "redact", false, "redact hostname, versions and bind addresses (safe to share)")
	return cmd
}

func runScan(cmd *cobra.Command, f *scanFlags) error {
	color := !f.noColor && isTTY(os.Stdout)
	if f.open {
		fmt.Fprintln(cmd.ErrOrStderr(), "note: --open is not yet implemented in v0.1")
	}

	mods := filterModules(registry.All(), f.modules, f.exclude)
	if len(mods) == 0 {
		return fmt.Errorf("no modules selected (check --modules / --exclude-modules)")
	}

	hostname, _ := os.Hostname()
	started := time.Now()

	ctx, cancel := context.WithTimeout(cmd.Context(), f.timeout)
	defer cancel()

	report := discovery.NewEngine(mods).Scan(ctx, discovery.ScanOptions{Parallel: f.parallel})
	completed := time.Now()

	meta := output.ScanMeta{
		ID:          "scan_" + started.UTC().Format("20060102_150405"),
		Hostname:    hostname,
		StartedAt:   started.UTC().Format(time.RFC3339),
		CompletedAt: completed.UTC().Format(time.RFC3339),
		DurationMs:  completed.Sub(started).Milliseconds(),
		Modules:     report.ModulesRun,
		Version:     version.Version,
	}
	doc := output.BuildDocument(report.Graph, meta, report.Warnings)
	if f.security {
		// Audit the unredacted graph; findings reference node IDs, not secrets.
		doc.Security = security.Audit(report.Graph.Nodes)
	}
	if f.redact {
		doc = output.Redact(doc)
	}

	if err := writeOutputs(cmd.OutOrStdout(), f, doc); err != nil {
		return err
	}
	if f.save != "" {
		data, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return err
		}
		path, err := store.Save(f.save, data)
		if err != nil {
			return fmt.Errorf("saving baseline %q: %w", f.save, err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "saved baseline "+f.save+" -> "+path)
	}
	// The report is a host inventory — surface that it is sensitive.
	if f.redact {
		fmt.Fprintln(cmd.ErrOrStderr(), "note: written with --redact (hostname, versions and bind addresses removed)")
	} else {
		fmt.Fprintln(cmd.ErrOrStderr(), "note: output contains a host inventory (hostname, ports, services, packages, versions) — treat as sensitive; use --redact to share")
	}

	if !f.quiet {
		output.PrintSummary(cmd.OutOrStdout(), doc, report.Skipped, report.Errors, color)
		if f.security {
			security.PrintFindings(cmd.OutOrStdout(), doc.Security, color)
		}
	}

	exitCode = doc.ExitCode()
	if f.security {
		switch security.WorstSeverity(doc.Security) {
		case security.SevCritical, security.SevHigh:
			if exitCode < 2 {
				exitCode = 2
			}
		case security.SevMedium, security.SevLow:
			if exitCode < 1 {
				exitCode = 1
			}
		}
	}
	return nil
}

// filterModules applies --modules / --exclude-modules. A selector matches a
// module's full name ("hardware.cpu") or its domain prefix ("hardware").
func filterModules(mods []discovery.Module, include, exclude []string) []discovery.Module {
	matches := func(name string, list []string) bool {
		for _, want := range list {
			if name == want {
				return true
			}
			if domain, _, ok := strings.Cut(name, "."); ok && domain == want {
				return true
			}
		}
		return false
	}
	var out []discovery.Module
	for _, m := range mods {
		if len(include) > 0 && !matches(m.Name(), include) {
			continue
		}
		if len(exclude) > 0 && matches(m.Name(), exclude) {
			continue
		}
		out = append(out, m)
	}
	return out
}

func writeOutputs(w io.Writer, f *scanFlags, doc output.Document) error {
	base := f.output
	for _, ext := range []string{".json", ".html", ".svg"} {
		base = strings.TrimSuffix(base, ext)
	}

	formats := strings.Split(strings.ToLower(f.format), ",")
	if len(formats) == 1 && strings.TrimSpace(formats[0]) == "all" {
		formats = []string{"json", "html"}
	}

	for _, fm := range formats {
		switch strings.TrimSpace(fm) {
		case "json":
			path := base + ".json"
			if err := writeFile(path, func(out io.Writer) error { return jsonout.Render(out, doc) }); err != nil {
				return err
			}
			fmt.Fprintln(w, "wrote "+path)
		case "html":
			path := base + ".html"
			if err := writeFile(path, func(out io.Writer) error { return htmlout.Render(out, doc) }); err != nil {
				return err
			}
			fmt.Fprintln(w, "wrote "+path)
		default:
			return fmt.Errorf("unknown format %q (want json, html, or all)", fm)
		}
	}
	return nil
}

func writeFile(path string, render func(io.Writer) error) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return render(file)
}

func isTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
