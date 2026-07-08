package modgraph

import "strings"

// Kind classifies a module as "root", "tool", "direct", or "transitive".
func Kind(mod string, info ModInfo) string {
	if mod == info.Module.Path {
		return "root"
	}
	for _, t := range info.Tool {
		if t.Path == mod || strings.HasPrefix(t.Path, mod+"/") {
			return "tool"
		}
	}
	for _, r := range info.Require {
		if r.Path == mod && !r.Indirect {
			return "direct"
		}
	}
	return "transitive"
}
