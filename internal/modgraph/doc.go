// Package modgraph loads and queries the Go module dependency graph.
//
// Load runs "go mod edit -json" and "go mod graph" in the current working
// directory and returns a Graph containing forward and reverse adjacency maps,
// per-module versions, per-edge imported versions, and the parsed go.mod
// metadata.
//
// Subgraph performs a reverse BFS from a target module to collect all modules
// that can reach it, then applies transitive reduction to produce the minimal
// set of edges that explain every path to the target.
//
// DirectImporters returns the subset of those modules that have a raw graph
// edge directly to the target, used by renderers to build per-importer rows.
//
// Kind classifies a module as root, tool, direct, or transitive based on its
// relationship to the root module's go.mod.
package modgraph
