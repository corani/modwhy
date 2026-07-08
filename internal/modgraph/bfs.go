package modgraph

import (
	"sort"
	"strings"
)

// Subgraph returns the minimal set of edges in the dependency graph that lie on
// any path from any module to target, with transitive reduction applied.
//
// An edge A->B is included only if B is not marked indirect in A's go.mod.
// Transitive reduction then suppresses A->B if a kept intermediate path A->M->B
// exists, provided doing so does not strand A or B.
func Subgraph(target string, g *Graph) []Edge {
	// 1. Reverse BFS from target to find all nodes that can reach it.
	onPath := map[string]bool{target: true}
	queue := []string{target}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for importer := range g.Radj[cur] {
			if !onPath[importer] {
				onPath[importer] = true
				queue = append(queue, importer)
			}
		}
	}

	root := g.Info.Module.Path

	// isToolEdge reports whether to is a tool dependency of from.
	isToolEdge := func(from, to string) bool {
		for t := range g.ToolDeps[from] {
			if t == to || strings.HasPrefix(t, to+"/") {
				return true
			}
		}
		return false
	}

	// isIndirect reports whether the edge from->to is marked indirect in from's go.mod.
	// Tool dependencies are never considered indirect.
	isIndirect := func(from, to string) bool {
		if isToolEdge(from, to) {
			return false
		}
		if deps, ok := g.Indirect[from]; ok {
			return deps[to]
		}
		return false
	}

	// 2. Collect candidate edges: both endpoints on-path, not indirect.
	type edge struct{ from, to string }
	var candidates []edge
	for from := range g.Adj {
		if !onPath[from] {
			continue
		}
		for to := range g.Adj[from] {
			if onPath[to] && !isIndirect(from, to) {
				candidates = append(candidates, edge{from, to})
			}
		}
	}

	// Build set of nodes reachable from root via kept root edges (for stranding check).
	rootReaches := map[string]bool{}
	for _, e := range candidates {
		if e.from == root {
			rootReaches[e.to] = true
		}
	}

	// 3. Transitive reduction over non-root candidates.
	// Suppress A->B if dominated by a kept A->M->B path, provided A retains
	// another outgoing kept edge and B retains another incoming kept edge.
	kept := make(map[edge]bool, len(candidates))
	for _, e := range candidates {
		kept[e] = true
	}

	for {
		changed := false
		for _, e := range candidates {
			if !kept[e] || e.from == root || e.to == target || isToolEdge(e.from, e.to) {
				continue
			}
			dominated := false
			for mid := range onPath {
				if mid == e.from || mid == e.to {
					continue
				}
				if kept[edge{e.from, mid}] && g.Adj[mid][e.to] {
					dominated = true
					break
				}
			}
			if !dominated {
				continue
			}
			otherOut := false
			for _, c := range candidates {
				if c.from == e.from && c != e && kept[c] {
					otherOut = true
					break
				}
			}
			if !otherOut {
				continue
			}
			otherIn := rootReaches[e.to]
			if !otherIn {
				for _, c := range candidates {
					if c.to == e.to && c != e && kept[c] {
						otherIn = true
						break
					}
				}
			}
			if !otherIn {
				continue
			}
			kept[e] = false
			changed = true
		}
		if !changed {
			break
		}
	}

	// 4. Emit kept edges with labels.
	var edges []Edge
	for _, e := range candidates {
		if kept[e] {
			edges = append(edges, Edge{From: e.from, To: e.to, Label: edgeLabel(e.from, e.to, g)})
		}
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})
	return edges
}

func edgeLabel(from, to string, g *Graph) string {
	for t := range g.ToolDeps[from] {
		if t == to || strings.HasPrefix(t, to+"/") {
			return "tool"
		}
	}
	if deps, ok := g.Indirect[from]; ok && !deps[to] {
		return "direct"
	}
	return ""
}

// DirectImporters returns all modules in subgraphNodes that have a direct edge
// to target in the raw graph.
func DirectImporters(target string, g *Graph, subgraphNodes map[string]bool) []string {
	seen := map[string]bool{}
	for from := range g.Radj[target] {
		if subgraphNodes[from] {
			seen[from] = true
		}
	}
	result := make([]string, 0, len(seen))
	for m := range seen {
		result = append(result, m)
	}
	sort.Strings(result)
	return result
}
