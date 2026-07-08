package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/corani/modwhy/internal/modgraph"
	"github.com/corani/modwhy/internal/output"
)

func main() {
	var module string
	var dot bool

	flag.StringVar(&module, "m", "", "module path to query (required)")
	flag.BoolVar(&dot, "dot", false, "output Graphviz dot format instead of markdown")
	flag.Parse()

	if module == "" {
		fmt.Fprintln(os.Stderr, "usage: modwhy -m <module> [-dot]")
		os.Exit(1)
	}

	g, err := modgraph.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "modwhy: %v\n", err)
		os.Exit(1)
	}

	edges := modgraph.Subgraph(module, g)

	if dot {
		output.Dot(os.Stdout, edges)
	} else {
		output.Markdown(os.Stdout, module, g, edges)
	}
}
