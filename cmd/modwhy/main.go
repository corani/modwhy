package main

import (
	"fmt"
	"io"
	"os"

	"github.com/corani/modwhy/internal/modgraph"
	"github.com/corani/modwhy/internal/output"
	"github.com/spf13/cobra"
)

func main() {
	var format, outfile, chdir string

	cmd := &cobra.Command{
		Use:   "modwhy -m <module>",
		Short: "Show who imports a Go module and why",
		Long: `modwhy answers "who imports this module and why?" for the current Go module.

Given a module path, it shows which modules in the dependency graph import it,
at what version, and how (direct, tool, transitive). It also supports Graphviz
dot, SVG, and PNG output for visualising the full import subgraph.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			module, _ := cmd.Flags().GetString("module")
			if module == "" {
				return fmt.Errorf("-m <module> is required")
			}

			if chdir != "" {
				if err := os.Chdir(chdir); err != nil {
					return err
				}
			}

			var w io.Writer = os.Stdout

			if outfile != "" {
				f, err := os.Create(outfile)
				if err != nil {
					return err
				}
				defer func() {
					if err := f.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "modwhy: %v\n", err)
					}
				}()
				w = f
			}

			g, err := modgraph.Load()
			if err != nil {
				return err
			}

			edges := modgraph.Subgraph(module, g)

			switch format {
			case "txt":
				output.Markdown(w, module, g, edges)
			case "md":
				output.GlamourMarkdown(w, module, g, edges)
			case "dot":
				output.Dot(w, edges)
			case "svg":
				output.SVG(w, edges)
			case "png":
				output.PNG(w, edges)
			case "mermaid":
				output.Mermaid(w, edges)
			case "csv":
				output.CSV(w, module, g, edges)
			default:
				return fmt.Errorf("unknown format %q (use txt, md, dot, svg, png, csv, mermaid)", format)
			}

			return nil
		},
	}

	cmd.Flags().StringP("module", "m", "", "module path to query (required)")
	cmd.Flags().StringVarP(&format, "format", "f", "txt", "output format: txt, md, dot, svg, png, csv, mermaid")
	cmd.Flags().StringVarP(&outfile, "output", "o", "", "write output to file instead of stdout")
	cmd.Flags().StringVarP(&chdir, "chdir", "C", "", "change to this directory before running")
	cmd.MarkFlagRequired("module") //nolint:errcheck

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
