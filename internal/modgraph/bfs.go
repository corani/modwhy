package modgraph

import (
	"sort"
	"strings"
)

// Subgraph returns the minimal set of edges in the dependency graph that lie on
// any path from any module to target, with transitive reduction applied.
//
// Transitive reduction: suppress edge A->B if there exists an intermediate node
// M (on the path) such that A->M and M->B both exist — except when A is the root
// module and B is a direct or tool dependency (those edges are always kept).
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

	// Build sets for fast lookup.
	indirect := make(map[string]bool)
	for _, r := range g.Info.Require {
		if r.Indirect {
			indirect[r.Path] = true
		}
	}
	toolMod := make(map[string]bool)
	for _, t := range g.Info.Tool {
		// Tool paths are package paths; check if the node (module path) is a prefix.
		toolMod[t.Path] = true
	}
	isToolMod := func(mod string) bool {
		for t := range toolMod {
			if t == mod || strings.HasPrefix(t, mod+"/") {
				return true
			}
		}
		return false
	}

	// 2. Collect candidate edges where both endpoints are on-path.
	type edge struct{ from, to string }
	var candidates []edge
	for from := range g.Adj {
		if !onPath[from] || from == root {
			continue
		}
		for to := range g.Adj[from] {
			if onPath[to] {
				candidates = append(candidates, edge{from, to})
			}
		}
	}

	// Build a set of kept non-root edges via transitive reduction.
	// An edge A->B is suppressed only if there exists a kept edge A->M and a
	// raw edge M->B (for some on-path M). Edges directly into target are never
	// suppressed — they are the reason the node is in the subgraph at all.
	kept := make(map[edge]bool, len(candidates))
	for _, e := range candidates {
		kept[e] = true
	}
	for {
		changed := false
		for _, e := range candidates {
			if !kept[e] || e.to == target {
				continue
			}
			for mid := range onPath {
				if mid == e.from || mid == e.to {
					continue
				}
				if kept[edge{e.from, mid}] && g.Adj[mid][e.to] {
					kept[e] = false
					changed = true
					break
				}
			}
		}
		if !changed {
			break
		}
	}

	// Post-pass: restore suppressed edges for any stranded node — one that has
	// no kept outgoing edge (except target) or no kept incoming edge (except root).
	for {
		changed := false
		hasOut := map[string]bool{target: true}
		hasIn := map[string]bool{root: true}
		for _, e := range candidates {
			if kept[e] {
				hasOut[e.from] = true
				hasIn[e.to] = true
			}
		}
		// Also count root edges toward hasIn.
		for from := range g.Adj {
			if from != root || !onPath[from] {
				continue
			}
			for to := range g.Adj[from] {
				if onPath[to] && (!indirect[to] || isToolMod(to)) {
					hasIn[to] = true
				}
			}
		}
		for _, e := range candidates {
			if kept[e] {
				continue
			}
			if (!hasOut[e.from] && hasOut[e.to]) || (!hasIn[e.to] && hasIn[e.from]) {
				kept[e] = true
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	// 3. Emit root edges and kept non-root edges.
	var edges []Edge
	for from := range g.Adj {
		if !onPath[from] {
			continue
		}
		for to := range g.Adj[from] {
			if !onPath[to] {
				continue
			}
			if from == root {
				if !indirect[to] || isToolMod(to) {
					edges = append(edges, Edge{from, to})
				}
				continue
			}
			if kept[edge{from, to}] {
				edges = append(edges, Edge{from, to})
			}
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
