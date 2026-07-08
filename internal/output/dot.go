package output

import (
	"fmt"
	"io"

	"github.com/corani/modwhy/internal/modgraph"
)

//nolint:errcheck
func Dot(w io.Writer, edges []modgraph.Edge) {
	fmt.Fprintln(w, "digraph {")
	fmt.Fprintln(w, "  rankdir=LR")

	for _, e := range edges {
		fmt.Fprintf(w, "  %q -> %q\n", e.From, e.To)
	}

	fmt.Fprintln(w, "}")
}
