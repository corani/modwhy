# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## What this tool does

`modwhy` answers "who imports this module and why?" for the current Go module.
Given a module path, it shows which modules in the dependency graph import it,
at what version, and how (direct, tool, transitive). It also supports Graphviz
dot, SVG, and PNG output for visualising the full import subgraph.

## Commands

```bash
go build ./...
go test ./...
go install ./cmd/modwhy
```

Run from the root of any Go module you want to query:

```bash
modwhy -m <module>              # plain markdown table (stdout)
modwhy -m <module> -f md        # rendered markdown via glamour
modwhy -m <module> -f dot       # Graphviz dot graph
modwhy -m <module> -f svg       # Graphviz SVG image
modwhy -m <module> -f png       # Graphviz PNG image
modwhy -m <module> -f csv       # CSV table
modwhy -m <module> -f mermaid   # Mermaid graph LR
modwhy -m <module> -o out.md    # write to file
modwhy -C /path/to/mod -m <module>  # query a different module
```

## Architecture

```text
cmd/modwhy/main.go         flag parsing, output dispatch
internal/modgraph/
  graph.go                 run go mod edit/graph, build adjacency maps
  bfs.go                   reverse BFS + transitive reduction → subgraph edges
  kind.go                  classify a module as root/tool/direct/transitive
internal/output/
  markdown.go              render plain markdown table (txt) or glamour (md)
  dot.go                   render Graphviz dot graph
  graphviz.go              render SVG and PNG via go-graphviz + tdewolff/canvas
  csv.go                   render CSV table
  mermaid.go               render Mermaid graph LR
```

### modgraph

`Load()` runs `go mod edit -json` and `go mod graph` in the caller's working
directory and returns a `Graph` with:

- `Adj` / `Radj` — forward and reverse adjacency maps (module path → set of
  module paths, versions stripped)
- `Versions` — first-seen version per module
- `EdgeVersions` — per-edge `from → to → to-version` (needed for the
  "imported version" column)
- `Info` — parsed `go.mod` metadata (module path, require list, tool list)

`Subgraph(target, g)` performs a reverse BFS from `target` to collect all
modules that can reach it (`onPath`), then applies transitive reduction:

- Root module edges: kept if the dependency is direct (not `// indirect`) or
  is a tool dependency. This preserves `root -> tool` edges even when the tool
  is also reachable transitively.
- All other edges: suppressed if any intermediate `onPath` node dominates them
  (`A -> M -> B` makes `A -> B` redundant).

`DirectImporters(target, g, subgraphNodes)` returns all modules in
`subgraphNodes` that have a raw graph edge to `target` — used by the markdown
renderer to build the table rows independently of the transitive-reduced
subgraph.

### kind classification

Priority order: root > tool (prefix match against `Tool[]` paths) > direct >
transitive.

Tool paths in `go.mod` are package paths (e.g. `.../cmd/golangci-lint`), so
the match checks whether any tool path has the module path as a prefix.

### Output

Markdown renders one row per direct importer of `target` (from the raw graph,
filtered to subgraph nodes), skipping the root module itself.

Dot renders all edges from the transitive-reduced subgraph, which shows the
minimal set of paths leading to `target`.

SVG and PNG use the same transitive-reduced subgraph. SVG is rendered directly
by the embedded Graphviz WASM binary. PNG is produced by rasterizing that SVG
via `tdewolff/canvas` (no external `dot` binary required for either).

## Post-change review checklist

- Can any public symbols be made private?
- Is any code unnecessary or duplicated?
- Could it be simpler or more idiomatic?
- Is the code easily testable?
