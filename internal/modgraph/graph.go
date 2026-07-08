package modgraph

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

type ModInfo struct {
	Module  struct{ Path string }
	Require []struct {
		Path     string
		Indirect bool
	}
	Tool []struct{ Path string }
}

type Edge struct {
	From, To, Label string
}

type Graph struct {
	Info         ModInfo
	Adj          map[string]map[string]bool   // from -> to
	Radj         map[string]map[string]bool   // to -> from
	Versions     map[string]string            // module -> version (first seen)
	EdgeVersions map[string]map[string]string // from -> to -> to-version
	Indirect     map[string]map[string]bool   // from -> to -> is-indirect (from cached go.mod files)
	ToolDeps     map[string]map[string]bool   // from -> to -> is-tool-dep (from cached go.mod files)
}

func Load() (*Graph, error) {
	infoBytes, err := runCmd("go", "mod", "edit", "-json")
	if err != nil {
		return nil, fmt.Errorf("go mod edit -json: %w", err)
	}

	var info ModInfo
	if err := json.Unmarshal(infoBytes, &info); err != nil {
		return nil, fmt.Errorf("parse go mod edit output: %w", err)
	}

	graphBytes, err := runCmd("go", "mod", "graph")
	if err != nil {
		return nil, fmt.Errorf("go mod graph: %w", err)
	}

	adj := make(map[string]map[string]bool)
	radj := make(map[string]map[string]bool)
	versions := make(map[string]string)
	edgeVersions := make(map[string]map[string]string)

	addEdge := func(from, to, toVer string) {
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
			edgeVersions[from][to] = toVer
		}
	}

	scanner := bufio.NewScanner(bytes.NewReader(graphBytes))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		from, fver := splitModVer(parts[0])
		to, tver := splitModVer(parts[1])
		if fver != "" && versions[from] == "" {
			versions[from] = fver
		}
		if tver != "" && versions[to] == "" {
			versions[to] = tver
		}
		addEdge(from, to, tver)
	}

	indirect, toolDeps := loadIndirect(versions, info)

	return &Graph{Info: info, Adj: adj, Radj: radj, Versions: versions, EdgeVersions: edgeVersions, Indirect: indirect, ToolDeps: toolDeps}, nil
}

// loadIndirect reads cached go.mod files for all known modules and returns:
//   - indirect: from -> to -> is-indirect
//   - toolDeps: from -> to -> is-tool-dep
func loadIndirect(versions map[string]string, info ModInfo) (indirect, toolDeps map[string]map[string]bool) {
	cacheDir, err := goModCache()
	if err != nil {
		cacheDir = filepath.Join(os.Getenv("HOME"), "go", "pkg", "mod", "cache", "download")
	} else {
		cacheDir = filepath.Join(cacheDir, "cache", "download")
	}

	indirect = make(map[string]map[string]bool)
	toolDeps = make(map[string]map[string]bool)

	// Seed root module from parsed go.mod (no cached file needed).
	root := info.Module.Path
	indirect[root] = make(map[string]bool)
	toolDeps[root] = make(map[string]bool)
	for _, r := range info.Require {
		indirect[root][r.Path] = r.Indirect
	}
	for _, t := range info.Tool {
		toolDeps[root][t.Path] = true
	}

	for mod, ver := range versions {
		if mod == root || ver == "" {
			continue
		}
		data, err := readCachedMod(cacheDir, mod, ver)
		if err != nil {
			continue
		}
		f, err := modfile.Parse(mod+"@"+ver+"/go.mod", data, nil)
		if err != nil {
			continue
		}
		isToolDep := func(path string) bool {
			for _, t := range f.Tool {
				if t.Path == path || strings.HasPrefix(t.Path, path+"/") {
					return true
				}
			}
			return false
		}
		m := make(map[string]bool)
		for _, r := range f.Require {
			m[r.Mod.Path] = r.Indirect && !isToolDep(r.Mod.Path)
		}
		indirect[mod] = m
		td := make(map[string]bool)
		for _, t := range f.Tool {
			td[t.Path] = true
		}
		toolDeps[mod] = td
	}

	return indirect, toolDeps
}

func goModCache() (string, error) {
	out, err := runCmd("go", "env", "GOMODCACHE")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func readCachedMod(cacheDir, mod, ver string) ([]byte, error) {
	// Module paths with capital letters are escaped in the cache (A -> !a).
	escaped := escapeModPath(mod)
	path := filepath.Join(append([]string{cacheDir}, append(strings.Split(escaped, "/"), "@v", ver+".mod")...)...)
	return os.ReadFile(path)
}

// escapeModPath applies the Go module cache path escaping (uppercase -> !lowercase).
func escapeModPath(mod string) string {
	var b strings.Builder
	for _, c := range mod {
		if c >= 'A' && c <= 'Z' {
			b.WriteByte('!')
			b.WriteRune(c + 32)
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}

func splitModVer(s string) (mod, ver string) {
	if before, after, ok := strings.Cut(s, "@"); ok {
		return before, after
	}
	return s, ""
}

func runCmd(name string, args ...string) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w\n%s", err, stderr.String())
	}
	return stdout.Bytes(), nil
}
