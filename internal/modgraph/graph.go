package modgraph

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
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
	From, To string
}

type Graph struct {
	Info         ModInfo
	Adj          map[string]map[string]bool   // from -> to
	Radj         map[string]map[string]bool   // to -> from
	Versions     map[string]string            // module -> version (first seen)
	EdgeVersions map[string]map[string]string // from -> to -> to-version
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

	return &Graph{Info: info, Adj: adj, Radj: radj, Versions: versions, EdgeVersions: edgeVersions}, nil
}

func splitModVer(s string) (mod, ver string) {
	if i := strings.Index(s, "@"); i >= 0 {
		return s[:i], s[i+1:]
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
