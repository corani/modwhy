# modwhy

Answers "who imports this module and why?" for the current Go module.

Given a module path, `modwhy` shows which modules in your dependency graph import
it, at what version, and how (direct, tool, or transitive). It also supports
Graphviz dot, SVG, PNG, and Mermaid output for visualising the full import
subgraph.

## Install

```bash
go install github.com/corani/modwhy/cmd/modwhy@latest
```

## Usage

Run from the root of any Go module you want to query:

```bash
modwhy -m <module>
```

### Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-m` | (required) | Module path to query |
| `-f` | `txt` | Output format (see below) |
| `-o` | stdout | Write output to this file |

### Output formats

- **txt** - Plain markdown table, suitable for piping or saving.
- **md** - Rendered markdown in the terminal via
  [glamour](https://charm.land/glamour), respects terminal width and
  `GLAMOUR_STYLE`.
- **dot** - Graphviz dot source for the transitive-reduced import subgraph.
- **svg** - Graphviz SVG image, rendered via embedded Graphviz WASM (no
  external `dot` binary needed).
- **png** - Graphviz PNG image, produced by rasterizing the SVG via
  `tdewolff/canvas` (no external tools needed).
- **csv** - CSV table with the same columns as `txt`.
- **mermaid** - Mermaid `graph LR` diagram of the import subgraph.

### Examples

```bash
# Which modules import golang.org/x/net, and how?
modwhy -m golang.org/x/net

# Same, but rendered nicely in the terminal
modwhy -m golang.org/x/net -f md

# Save a Graphviz dot file and render manually
modwhy -m golang.org/x/net -f dot -o net.dot
dot -Tsvg net.dot -o net.svg

# SVG or PNG directly (no external dot binary needed)
modwhy -m golang.org/x/net -f svg -o net.svg
modwhy -m golang.org/x/net -f png -o net.png

# Mermaid diagram
modwhy -m golang.org/x/net -f mermaid
```

### Example output (txt)

```text
## Importers of `golang.org/x/net`

| Importer | Importer Version | Kind | Imported Version |
| --- | --- | --- | --- |
| `golang.org/x/crypto` | `v0.39.0` | transitive | `v0.39.0` |
| `github.com/some/tool` | `v1.2.3` | tool | `v0.38.0` |
```

### Kind column

| Value | Meaning |
| --- | --- |
| `direct` | Listed in `require` without `// indirect` |
| `tool` | Listed in `tool` directives in `go.mod` |
| `transitive` | Pulled in by another dependency |

## How it works

`modwhy` runs `go mod edit -json` and `go mod graph` in the current directory,
builds a reverse adjacency map, and performs a reverse BFS from the target module
to find all modules that can reach it. It then applies transitive reduction to
produce a minimal subgraph, keeping only the edges that are not already implied
by an intermediate node.
