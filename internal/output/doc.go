// Package output renders a module import subgraph in various formats.
//
// Each function takes an io.Writer and the data produced by the modgraph
// package and writes the formatted result:
//
//   - Markdown / GlamourMarkdown — plain or terminal-rendered markdown table
//   - Dot — Graphviz dot source
//   - SVG / PNG — rendered graph images via embedded Graphviz WASM
//   - CSV — comma-separated table
//   - Mermaid — Mermaid graph LR diagram
package output
