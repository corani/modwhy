package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/corani/modwhy/internal/modgraph"
	"github.com/corani/modwhy/internal/output"
)

func main() {
	var module string
	var format string
	var outfile string

	flag.StringVar(&module, "m", "", "module path to query (required)")
	flag.StringVar(&format, "f", "txt", "output format: txt, md, dot, svg, png, csv, mermaid")
	flag.StringVar(&outfile, "o", "", "output file (default stdout)")
	flag.Parse()

	if module == "" {
		fmt.Fprintln(os.Stderr, "usage: modwhy -m <module> [-f txt|md|dot|csv] [-o file]")
		os.Exit(1)
	}

	var w io.Writer = os.Stdout
	if outfile != "" {
		f, err := os.Create(outfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "modwhy: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}

	g, err := modgraph.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "modwhy: %v\n", err)
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "modwhy: unknown format %q (use txt, md, dot, svg, png, csv, mermaid)\n", format)
		os.Exit(1)
	}
}
