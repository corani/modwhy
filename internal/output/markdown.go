package output

import (
	"fmt"
	"io"

	"github.com/corani/modwhy/internal/modgraph"
)

func Markdown(w io.Writer, target string, g *modgraph.Graph, edges []modgraph.Edge) {
	// For the table, show all modules that directly import the target in the raw graph
	// (not just those in the reduced subgraph), but restrict to nodes that appear in
	// the subgraph (i.e. are on a path to target).
	subgraphNodes := map[string]bool{}
	for _, e := range edges {
		subgraphNodes[e.From] = true
		subgraphNodes[e.To] = true
	}

	importers := modgraph.DirectImporters(target, g, subgraphNodes)

	fmt.Fprintf(w, "## Importers of `%s`\n\n", target)
	fmt.Fprintln(w, "| Importer | Importer Version | Kind | Imported Version |")
	fmt.Fprintln(w, "| --- | --- | --- | --- |")

	for _, mod := range importers {
		k := modgraph.Kind(mod, g.Info)
		if k == "root" {
			continue
		}
		ver := g.Versions[mod]
		if ver == "" {
			ver = "—"
		}
		importedVer := g.EdgeVersions[mod][target]
		fmt.Fprintf(w, "| `%s` | `%s` | %s | `%s` |\n", mod, ver, k, importedVer)
	}
}
