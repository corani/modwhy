package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/corani/modwhy/internal/modgraph"
	"github.com/corani/modwhy/internal/output"
	"github.com/stretchr/testify/require"
)

var testInfo = modgraph.ModInfo{
	Module: struct{ Path string }{Path: "example.com/root"},
	Require: []struct {
		Path     string
		Indirect bool
	}{
		{Path: "example.com/direct", Indirect: false},
		{Path: "example.com/indirect", Indirect: true},
	},
	Tool: []struct{ Path string }{
		{Path: "example.com/tool/cmd/tool"},
	},
}

var testGraph = &modgraph.Graph{
	Info: testInfo,
	Adj: map[string]map[string]bool{
		"example.com/root":   {"example.com/direct": true, "example.com/target": true},
		"example.com/direct": {"example.com/target": true},
	},
	Radj: map[string]map[string]bool{
		"example.com/direct": {"example.com/root": true},
		"example.com/target": {"example.com/root": true, "example.com/direct": true},
	},
	Versions: map[string]string{
		"example.com/root":   "v0.0.0",
		"example.com/direct": "v1.2.3",
		"example.com/target": "v2.0.0",
	},
	EdgeVersions: map[string]map[string]string{
		"example.com/root":   {"example.com/target": "v2.0.0"},
		"example.com/direct": {"example.com/target": "v2.0.0"},
	},
}

var testEdges = []modgraph.Edge{
	{From: "example.com/direct", To: "example.com/target"},
	{From: "example.com/root", To: "example.com/direct"},
}

func TestDot(t *testing.T) {
	var buf bytes.Buffer
	output.Dot(&buf, testEdges)
	got := buf.String()

	require.Contains(t, got, "digraph {")
	require.Contains(t, got, "rankdir=LR")
	require.Contains(t, got, `"example.com/direct" -> "example.com/target"`)
	require.Contains(t, got, `"example.com/root" -> "example.com/direct"`)
	require.Contains(t, got, "}")
}

func TestMermaid(t *testing.T) {
	var buf bytes.Buffer
	output.Mermaid(&buf, testEdges)
	got := buf.String()

	require.Contains(t, got, "graph LR")
	require.Contains(t, got, `"example.com/direct" --> "example.com/target"`)
	require.Contains(t, got, `"example.com/root" --> "example.com/direct"`)
}

func TestMarkdown(t *testing.T) {
	var buf bytes.Buffer
	output.Markdown(&buf, "example.com/target", testGraph, testEdges)
	got := buf.String()

	require.Contains(t, got, "example.com/target")
	require.Contains(t, got, "example.com/direct")
	require.Contains(t, got, "v1.2.3")
	require.Contains(t, got, "v2.0.0")
	require.Contains(t, got, "direct")
	// root module should not appear as a row
	lines := strings.Split(got, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "|") {
			require.NotContains(t, line, "example.com/root")
		}
	}
}

func TestCSV(t *testing.T) {
	var buf bytes.Buffer
	output.CSV(&buf, "example.com/target", testGraph, testEdges)
	got := buf.String()

	lines := strings.Split(strings.TrimSpace(got), "\n")
	require.Equal(t, "Importer,Importer Version,Kind,Imported Version", lines[0])
	require.Len(t, lines, 2)
	require.Contains(t, lines[1], "example.com/direct")
	require.Contains(t, lines[1], "v1.2.3")
	require.Contains(t, lines[1], "direct")
	require.Contains(t, lines[1], "v2.0.0")
}

func TestDotMermaidNoEdges(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*bytes.Buffer, []modgraph.Edge)
	}{
		{"dot", func(b *bytes.Buffer, e []modgraph.Edge) { output.Dot(b, e) }},
		{"mermaid", func(b *bytes.Buffer, e []modgraph.Edge) { output.Mermaid(b, e) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.fn(&buf, nil)
			require.NotEmpty(t, buf.String())
		})
	}
}
