package output

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"charm.land/glamour/v2"
	"github.com/corani/modwhy/internal/modgraph"
	"golang.org/x/term"
)

func markdownSource(target string, g *modgraph.Graph, edges []modgraph.Edge) string {
	subgraphNodes := map[string]bool{}

	for _, e := range edges {
		subgraphNodes[e.From] = true
		subgraphNodes[e.To] = true
	}

	importers := modgraph.DirectImporters(target, g, subgraphNodes)

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "## Importers of `%s`\n\n", target)
	fmt.Fprintln(&buf, "| Importer | Importer Version | Kind | Imported Version |")
	fmt.Fprintln(&buf, "| --- | --- | --- | --- |")

	for _, mod := range importers {
		k := modgraph.Kind(mod, g.Info)
		if k == "root" {
			continue
		}

		ver := g.Versions[mod]
		if ver == "" {
			ver = "-"
		}

		importedVer := g.EdgeVersions[mod][target]

		fmt.Fprintf(&buf, "| `%s` | `%s` | %s | `%s` |\n", mod, ver, k, importedVer)
	}

	return buf.String()
}

//nolint:errcheck
func Markdown(w io.Writer, target string, g *modgraph.Graph, edges []modgraph.Edge) {
	fmt.Fprint(w, markdownSource(target, g, edges))
}

//nolint:errcheck
func GlamourMarkdown(w io.Writer, target string, g *modgraph.Graph, edges []modgraph.Edge) {
	src := markdownSource(target, g, edges)

	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		width = 80
	}

	tr, err := glamour.NewTermRenderer(glamour.WithEnvironmentConfig(), glamour.WithWordWrap(width))
	if err != nil {
		fmt.Fprint(w, src)

		return
	}

	out, err := tr.Render(src)
	if err != nil {
		fmt.Fprint(w, src)

		return
	}

	fmt.Fprint(w, out)
}
