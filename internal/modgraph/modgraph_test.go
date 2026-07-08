package modgraph_test

import (
	"testing"

	"github.com/corani/modwhy/internal/modgraph"
	"github.com/stretchr/testify/require"
)

func TestKind(t *testing.T) {
	info := modgraph.ModInfo{
		Module: struct{ Path string }{Path: "example.com/root"},
		Require: []struct {
			Path     string
			Indirect bool
		}{
			{Path: "example.com/direct", Indirect: false},
			{Path: "example.com/indirect", Indirect: true},
		},
		Tool: []struct{ Path string }{
			{Path: "example.com/toolpkg/cmd/tool"},
		},
	}

	tests := []struct {
		mod  string
		want string
	}{
		{"example.com/root", "root"},
		{"example.com/direct", "direct"},
		{"example.com/indirect", "transitive"},
		{"example.com/toolpkg", "tool"},
		{"example.com/toolpkg/cmd/tool", "tool"},
		{"example.com/other", "transitive"},
	}

	for _, tt := range tests {
		t.Run(tt.mod, func(t *testing.T) {
			require.Equal(t, tt.want, modgraph.Kind(tt.mod, info))
		})
	}
}

// buildGraph constructs a Graph from a slice of "from@ver to@ver" edge strings
// and a ModInfo, mirroring what Load() produces.
func buildGraph(info modgraph.ModInfo, rawEdges []string) *modgraph.Graph {
	adj := make(map[string]map[string]bool)
	radj := make(map[string]map[string]bool)
	versions := make(map[string]string)
	edgeVersions := make(map[string]map[string]string)

	addEdge := func(from, fver, to, tver string) {
		if versions[from] == "" && fver != "" {
			versions[from] = fver
		}
		if versions[to] == "" && tver != "" {
			versions[to] = tver
		}
		if adj[from] == nil {
			adj[from] = make(map[string]bool)
		}
		adj[from][to] = true
		if radj[to] == nil {
			radj[to] = make(map[string]bool)
		}
		radj[to][from] = true
		if edgeVersions[from] == nil {
			edgeVersions[from] = make(map[string]string)
		}
		if edgeVersions[from][to] == "" {
			edgeVersions[from][to] = tver
		}
	}

	for _, e := range rawEdges {
		var fromMod, fromVer, toMod, toVer string
		// Parse "from@ver to@ver" or "from to"
		var a, b string
		_, _ = splitTwo(e, &a, &b)
		fromMod, fromVer = splitAt(a)
		toMod, toVer = splitAt(b)
		addEdge(fromMod, fromVer, toMod, toVer)
	}

	return &modgraph.Graph{
		Info:         info,
		Adj:          adj,
		Radj:         radj,
		Versions:     versions,
		EdgeVersions: edgeVersions,
		Indirect:     buildIndirect(info),
	}
}

func buildIndirect(info modgraph.ModInfo) map[string]map[string]bool {
	result := map[string]map[string]bool{}
	m := map[string]bool{}
	for _, r := range info.Require {
		m[r.Path] = r.Indirect
	}
	result[info.Module.Path] = m
	return result
}

func splitTwo(s string, a, b *string) (int, error) {
	for i, c := range s {
		if c == ' ' {
			*a = s[:i]
			*b = s[i+1:]
			return 2, nil
		}
	}
	*a = s
	return 1, nil
}

func splitAt(s string) (mod, ver string) {
	for i, c := range s {
		if c == '@' {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}

func TestSubgraph(t *testing.T) {
	root := "example.com/root"

	tests := []struct {
		name      string
		edges     []string
		requires  []struct{ Path string; Indirect bool }
		target    string
		wantEdges []modgraph.Edge
	}{
		{
			name: "direct only",
			edges: []string{
				"example.com/root@v0.0.0 example.com/a@v1.0.0",
				"example.com/root@v0.0.0 example.com/target@v1.0.0",
			},
			requires: []struct{ Path string; Indirect bool }{
				{Path: "example.com/a", Indirect: false},
				{Path: "example.com/target", Indirect: false},
			},
			target: "example.com/target",
			wantEdges: []modgraph.Edge{
				{From: "example.com/root", To: "example.com/target", Label: "direct"},
			},
		},
		{
			name: "transitive reduction: root->a->target suppresses root->target",
			edges: []string{
				"example.com/root@v0.0.0 example.com/a@v1.0.0",
				"example.com/root@v0.0.0 example.com/target@v1.0.0",
				"example.com/a@v1.0.0 example.com/target@v1.0.0",
			},
			requires: []struct{ Path string; Indirect bool }{
				{Path: "example.com/a", Indirect: false},
				{Path: "example.com/target", Indirect: true},
			},
			target: "example.com/target",
			wantEdges: []modgraph.Edge{
				{From: "example.com/a", To: "example.com/target", Label: ""},
				{From: "example.com/root", To: "example.com/a", Label: "direct"},
			},
		},
		{
			name: "unrelated modules excluded",
			edges: []string{
				"example.com/root@v0.0.0 example.com/a@v1.0.0",
				"example.com/root@v0.0.0 example.com/unrelated@v1.0.0",
				"example.com/a@v1.0.0 example.com/target@v1.0.0",
			},
			requires: []struct{ Path string; Indirect bool }{
				{Path: "example.com/a", Indirect: false},
				{Path: "example.com/unrelated", Indirect: false},
			},
			target: "example.com/target",
			wantEdges: []modgraph.Edge{
				{From: "example.com/a", To: "example.com/target", Label: ""},
				{From: "example.com/root", To: "example.com/a", Label: "direct"},
			},
		},
		{
			name: "target not in graph",
			edges: []string{
				"example.com/root@v0.0.0 example.com/a@v1.0.0",
			},
			requires:  nil,
			target:    "example.com/missing",
			wantEdges: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := modgraph.ModInfo{
				Module:  struct{ Path string }{Path: root},
				Require: tt.requires,
			}
			g := buildGraph(info, tt.edges)
			got := modgraph.Subgraph(tt.target, g)
			require.Equal(t, tt.wantEdges, got)
		})
	}
}

func TestDirectImporters(t *testing.T) {
	root := "example.com/root"
	info := modgraph.ModInfo{Module: struct{ Path string }{Path: root}}

	tests := []struct {
		name          string
		edges         []string
		target        string
		subgraphNodes map[string]bool
		want          []string
	}{
		{
			name: "two direct importers",
			edges: []string{
				"example.com/a@v1.0.0 example.com/target@v1.0.0",
				"example.com/b@v1.0.0 example.com/target@v1.0.0",
				"example.com/c@v1.0.0 example.com/other@v1.0.0",
			},
			target: "example.com/target",
			subgraphNodes: map[string]bool{
				"example.com/a":      true,
				"example.com/b":      true,
				"example.com/target": true,
			},
			want: []string{"example.com/a", "example.com/b"},
		},
		{
			name: "importer not in subgraph excluded",
			edges: []string{
				"example.com/a@v1.0.0 example.com/target@v1.0.0",
				"example.com/b@v1.0.0 example.com/target@v1.0.0",
			},
			target: "example.com/target",
			subgraphNodes: map[string]bool{
				"example.com/a":      true,
				"example.com/target": true,
			},
			want: []string{"example.com/a"},
		},
		{
			name:          "no importers",
			edges:         []string{},
			target:        "example.com/target",
			subgraphNodes: map[string]bool{"example.com/target": true},
			want:          []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := buildGraph(info, tt.edges)
			got := modgraph.DirectImporters(tt.target, g, tt.subgraphNodes)
			require.Equal(t, tt.want, got)
		})
	}
}
