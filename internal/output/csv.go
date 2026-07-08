package output

import (
	"encoding/csv"
	"io"

	"github.com/corani/modwhy/internal/modgraph"
)

func CSV(w io.Writer, target string, g *modgraph.Graph, edges []modgraph.Edge) {
	subgraphNodes := map[string]bool{}
	for _, e := range edges {
		subgraphNodes[e.From] = true
		subgraphNodes[e.To] = true
	}

	importers := modgraph.DirectImporters(target, g, subgraphNodes)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"Importer", "Importer Version", "Kind", "Imported Version"})

	for _, mod := range importers {
		k := modgraph.Kind(mod, g.Info)
		if k == "root" {
			k = modgraph.Kind(target, g.Info)
		}
		ver := g.Versions[mod]
		_ = cw.Write([]string{mod, ver, k, g.EdgeVersions[mod][target]})
	}

	cw.Flush()
}
