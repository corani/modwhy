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

	// 2. Emit edges where both endpoints are on-path, with transitive reduction.
	var edges []Edge
	for from := range g.Adj {
		if !onPath[from] {
			continue
		}
		for to := range g.Adj[from] {
			if !onPath[to] {
				continue
			}
			// Root edges: keep if direct require or tool dependency.
			if from == root {
				if !indirect[to] || isToolMod(to) {
					edges = append(edges, Edge{from, to})
				}
				continue
			}
			// Non-root: suppress if any intermediate on-path node dominates.
			dominated := false
			for mid := range onPath {
				if mid == from || mid == to {
					continue
				}
				if g.Adj[from][mid] && g.Adj[mid][to] {
					dominated = true
					break
				}
			}
			if !dominated {
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
