package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/corani/modwhy/internal/modgraph"
)

func Mermaid(w io.Writer, edges []modgraph.Edge) {
	fmt.Fprintln(w, "graph LR")
	for _, e := range edges {
		fmt.Fprintf(w, "  %s --> %s\n", mermaidID(e.From), mermaidID(e.To))
	}
}

func mermaidID(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `#quot;`) + `"`
}
