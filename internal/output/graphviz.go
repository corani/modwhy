package output

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"io"

	"github.com/goccy/go-graphviz"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
	"github.com/corani/modwhy/internal/modgraph"
)

func svgBytes(edges []modgraph.Edge) ([]byte, error) {
	var dot bytes.Buffer
	Dot(&dot, edges)

	ctx := context.Background()
	g, err := graphviz.New(ctx)
	if err != nil {
		return nil, err
	}
	defer g.Close()

	graph, err := graphviz.ParseBytes(dot.Bytes())
	if err != nil {
		return nil, err
	}
	defer graph.Close()

	var buf bytes.Buffer
	if err := g.Render(ctx, graph, graphviz.SVG, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func SVG(w io.Writer, edges []modgraph.Edge) {
	b, err := svgBytes(edges)
	if err != nil {
		fmt.Fprintf(w, "modwhy: graphviz svg: %v\n", err)
		return
	}
	w.Write(b)
}

func PNG(w io.Writer, edges []modgraph.Edge) {
	b, err := svgBytes(edges)
	if err != nil {
		fmt.Fprintf(w, "modwhy: graphviz svg: %v\n", err)
		return
	}

	c, err := canvas.ParseSVG(bytes.NewReader(b))
	if err != nil {
		fmt.Fprintf(w, "modwhy: svg parse: %v\n", err)
		return
	}

	img := rasterizer.Draw(c, canvas.DPMM(3.78), canvas.DefaultColorSpace)
	if err := png.Encode(w, img); err != nil {
		fmt.Fprintf(w, "modwhy: png encode: %v\n", err)
	}
}
