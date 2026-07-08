// Package main implements the modwhy command-line tool.
//
// modwhy answers "who imports this module and why?" for the current Go module.
// Run it from the root of any Go module:
//
//	modwhy -m <module>
//
// Use -f to select an output format (txt, md, dot, svg, png, csv, mermaid),
// -o to write to a file instead of stdout, and -C to query a different module
// without changing the working directory.
package main
