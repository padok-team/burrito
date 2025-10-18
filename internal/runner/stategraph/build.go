package stategraph

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Node struct {
	ID             string         `json:"id"`
	Addr           string         `json:"addr"`
	Mode           string         `json:"mode"`
	Type           string         `json:"type"`
	Name           string         `json:"name"`
	Module         string         `json:"module,omitempty"`
	Provider       string         `json:"provider,omitempty"`
	InstancesCount int            `json:"instances_count,omitempty"`
	Instances      []InstanceInfo `json:"instances,omitempty"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type InstanceInfo struct {
	Addr         string         `json:"addr"`
	Index        string         `json:"index,omitempty"`
	Dependencies []string       `json:"dependencies,omitempty"`
	Attributes   map[string]any `json:"attributes,omitempty"`
	CreatedAt    string         `json:"created_at,omitempty"`
}

// BuildGraphFromState parses a terraform.tfstate JSON and builds a graph of
// resource instances and their dependencies.
func BuildGraphFromState(data []byte) ([]byte, error) {
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("decode state: %w", err)
	}

	nodes := make([]Node, 0)
	edges := make([]Edge, 0)

	// Indexes to resolve dependencies (grouped)
	// baseAddr -> node ID (same as base)
	byBase := map[string]string{}
	// full addr -> node ID (its base)
	byFull := map[string]string{}

	// Accumulator for grouped nodes
	groups := map[string]*Node{}

	// Build grouped nodes per base address
	// e.g., module.foo.aws_instance.bar
	// with instances like [0], [1], ["key"], ...
	// All instances of the same base are grouped in a single Node.
	for _, rsrc := range st.Resources {
		base := resourceBaseAddr(rsrc)
		g := groups[base]
		if g == nil {
			g = &Node{
				ID:       base,
				Addr:     base,
				Mode:     rsrc.Mode,
				Type:     rsrc.Type,
				Name:     rsrc.Name,
				Module:   rsrc.Module,
				Provider: cleanProvider(rsrc.Provider),
			}
			groups[base] = g
			byBase[base] = g.ID
		}
		if len(rsrc.Instances) == 0 {
			// No instances recorded: treat base as a single instance
			byFull[base] = g.ID
			continue
		}
		for _, inst := range rsrc.Instances {
			idx := instanceIndexSuffix(inst)
			full := base + idx
			attrs := filterSensitive(inst.Attributes, inst.SensitiveAttributes)
			createdAt := guessCreatedAt(attrs)
			g.Instances = append(g.Instances, InstanceInfo{
				Addr:         full,
				Index:        idx,
				Dependencies: append([]string{}, inst.Dependencies...),
				Attributes:   attrs,
				CreatedAt:    createdAt,
			})
			byFull[full] = g.ID
		}
	}
	// Add instances count property and collect nodes
	for _, n := range groups {
		n.InstancesCount = len(n.Instances)
		nodes = append(nodes, *n)
	}

	// Add edges between grouped nodes
	for _, rsrc := range st.Resources {
		base := resourceBaseAddr(rsrc)
		targetID := byBase[base]
		for _, inst := range rsrc.Instances {
			if len(inst.Dependencies) == 0 {
				continue
			}
			for _, dep := range inst.Dependencies {
				fromIDs := resolveDependencyGrouped(dep, byFull, byBase)
				for _, from := range fromIDs {
					if from == "" || from == targetID {
						continue
					}
					edges = append(edges, Edge{From: from, To: targetID})
				}
			}
		}
	}

	// Stable order for deterministic output
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From == edges[j].From {
			return edges[i].To < edges[j].To
		}
		return edges[i].From < edges[j].From
	})
	graph := Graph{Nodes: nodes, Edges: uniqueEdges(edges)}
	graph = *onlyManaged(&graph)
	graph = *reduceTransitive(&graph)

	jsonGraph, err := json.Marshal(graph)
	if err != nil {
		return nil, fmt.Errorf("encode graph: %w", err)
	}
	return jsonGraph, nil
}

// OnlyManaged returns a new graph containing only managed resources and edges
// where both endpoints are managed.
// It basically filters out data resources which are noisy in the graph and not
// relevant for plan / apply operations.
func onlyManaged(g *Graph) *Graph {
	keep := make(map[string]bool, len(g.Nodes))
	out := &Graph{}
	for _, n := range g.Nodes {
		if n.Mode == "managed" {
			keep[n.ID] = true
			out.Nodes = append(out.Nodes, n)
		}
	}
	for _, e := range g.Edges {
		if keep[e.From] && keep[e.To] {
			out.Edges = append(out.Edges, e)
		}
	}
	return out
}

// resourceBaseAddr builds the base address of a resource, without instance index.
// E.g., module.foo.aws_instance.bar
// For data resources, includes the "data." prefix.
// E.g., module.foo.data.aws_ami.bar
func resourceBaseAddr(r Resource) string {
	var b strings.Builder
	if r.Module != "" {
		b.WriteString(r.Module)
		b.WriteString(".")
	}
	if r.Mode == "data" {
		b.WriteString("data.")
	}
	b.WriteString(r.Type)
	b.WriteString(".")
	b.WriteString(r.Name)
	return b.String()
}

// instanceIndexSuffix returns the index suffix for an instance, e.g., [0] or ["key"].
// Returns empty string if no index (single instance).
func instanceIndexSuffix(inst Instance) string {
	if inst.IndexKey == nil {
		return ""
	}
	// Terraform addresses render indices like [0] or ["key"]
	switch v := inst.IndexKey.(type) {
	case float64:
		return fmt.Sprintf("[%d]", int(v))
	case string:
		return fmt.Sprintf("[\"%s\"]", v)
	default:
		return fmt.Sprintf("[%v]", v)
	}
}

var bracketRe = regexp.MustCompile(`\[[^\]]*\]$`)

// resolveDependencyGrouped resolves a dependency to the grouped node ID.
// Tries full address first, then normalized index, then base address.
// Returns nil if not found.
// The goal here is to resolve to a single node ID if possible, to avoid
// creating edges to multiple nodes in the same group, which would be noisy in the graph.
func resolveDependencyGrouped(dep string, byFull map[string]string, byBase map[string]string) []string {
	if id, ok := byFull[dep]; ok {
		return []string{id}
	}
	normalized := normalizeIndex(dep)
	if id, ok := byFull[normalized]; ok {
		return []string{id}
	}
	base := bracketRe.ReplaceAllString(dep, "")
	if id, ok := byBase[base]; ok {
		return []string{id}
	}
	return nil
}

func normalizeIndex(addr string) string {
	// Convert [key] to ["key"] if needed
	i := strings.LastIndex(addr, "[")
	if i < 0 || !strings.HasSuffix(addr, "]") {
		return addr
	}
	inner := addr[i+1 : len(addr)-1]
	if inner == "" {
		return addr
	}
	if inner[0] == '"' || inner[0] == '\'' {
		return addr
	}
	// If it's a number, keep as-is
	if _, err := fmt.Sscanf(inner, "%d", new(int)); err == nil {
		return addr
	}
	return addr[:i] + "[\"" + inner + "\"]"
}

// cleanProvider extracts a short provider name from Terraform provider
// address strings which usually look like:
//
//	provider["registry.terraform.io/hashicorp/random"]
//
// It can also be other formats, e.g.:
//
//	provider["random"]
//	provider['random']
//	provider[registry.terraform.io/hashicorp/random]
//
//	...
func cleanProvider(in string) string {
	if in == "" {
		return ""
	}
	// Look for content inside the brackets: provider["..."]
	// We'll accept both single and double quotes.
	// First, find the first '[' and the last ']'.
	i := strings.Index(in, "[")
	j := strings.LastIndex(in, "]")
	var inside string
	if i >= 0 && j > i {
		inside = in[i+1 : j]
		inside = strings.Trim(inside, "'\" ")
	}
	// If we found something like registry.terraform.io/hashicorp/random
	// return the full path
	if inside != "" {
		return inside
	}
	// Fallback: return the trimmed original string (remove surrounding quotes/space)
	return strings.Trim(in, "'\" ")
}

// uniqueEdges removes duplicate edges while preserving order.
func uniqueEdges(in []Edge) []Edge {
	seen := make(map[string]struct{}, len(in))
	out := make([]Edge, 0, len(in))
	for _, e := range in {
		k := e.From + "->" + e.To
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, e)
	}
	return out
}

// filterSensitive removes attributes that are marked sensitive in the state's
// SensitiveAttributes structure. We best-effort extract leaf keys and drop
// those from the attributes map. If parsing fails, returns attrs unchanged.
// Reason why this function looks like this: Terraform encodes sensitive attributes
// in a few different ways depending on the version and resource type, e.g.:
//
//	"sensitive_attributes": [ ["value", 3], ["nested", 1, "password"] ]
//	"sensitive_attributes": [ ["value"], ["nested", "password"] ]
//	"sensitive_attributes": { "type": "list", "value": ["password"] }
//	"sensitive_attributes": { "type": "map", "value": ["password"] }
//	"sensitive_attributes": { "type": "object", "value": ["password"] }
func filterSensitive(attrs map[string]any, sensitive any) map[string]any {
	if attrs == nil {
		return nil
	}
	if sensitive == nil {
		// nothing to filter
		return attrs
	}
	// Collect leaf keys (top-level attribute names) from various encodings
	keys := map[string]struct{}{}
	switch s := sensitive.(type) {
	case []any:
		for _, v := range s {
			switch vv := v.(type) {
			case []any:
				// Expect [[ { "type": "get_attr", "value": "field" } ]]
				for _, inner := range vv {
					if m, ok := inner.(map[string]any); ok {
						if val, ok2 := m["value"].(string); ok2 && val != "" {
							keys[val] = struct{}{}
						}
					}
				}
			case map[string]any:
				if val, ok2 := vv["value"].(string); ok2 && val != "" {
					keys[val] = struct{}{}
				}
			}
		}
	case map[string]any:
		if val, ok := s["value"].(string); ok && val != "" {
			keys[val] = struct{}{}
		}
	}
	if len(keys) == 0 {
		return attrs
	}
	out := make(map[string]any, len(attrs))
	for k, v := range attrs {
		if _, drop := keys[k]; drop {
			continue
		}
		out[k] = v
	}
	return out
}

// guessCreatedAt pulls a likely RFC3339 timestamp from attributes when available
// (e.g., time_static resources expose rfc3339). Returns empty string if none.
// CreatedAt is not always available in state, so this is a best-effort guess.
func guessCreatedAt(attrs map[string]any) string {
	if attrs == nil {
		return ""
	}
	if s, ok := attrs["rfc3339"].(string); ok {
		if _, err := time.Parse(time.RFC3339, s); err == nil {
			return s
		}
	}
	// Fallback: if id looks like RFC3339
	if s, ok := attrs["id"].(string); ok {
		if _, err := time.Parse(time.RFC3339, s); err == nil {
			return s
		}
	}
	return ""
}

// ReduceTransitive removes edges (u->v) for which there exists an alternate
// path u => ... => v of length >= 2. Assumes a DAG, but works conservatively
// even with cycles (keeps the original edge if unsure).
func reduceTransitive(g *Graph) *Graph {
	// Build adjacency
	adj := map[string][]string{}
	for _, e := range g.Edges {
		adj[e.From] = append(adj[e.From], e.To)
	}
	// For each edge u->v, check if v is reachable from u without using that
	// direct edge on the first step.
	keep := make([]bool, len(g.Edges))
	for i, e := range g.Edges {
		if hasAlternatePath(adj, e.From, e.To) {
			keep[i] = false
		} else {
			keep[i] = true
		}
	}
	out := make([]Edge, 0, len(g.Edges))
	for i, e := range g.Edges {
		if keep[i] {
			out = append(out, e)
		}
	}
	return &Graph{Nodes: g.Nodes, Edges: out}
}

// hasAlternatePath checks if there is a path from 'from' to 'to' in the graph
// represented by adj, excluding the direct edge from->to.
// Uses BFS. Returns true if such a path exists.
func hasAlternatePath(adj map[string][]string, from, to string) bool {
	q := make([]string, 0, len(adj[from]))
	for _, n := range adj[from] {
		if n != to { // skip direct edge
			q = append(q, n)
		}
	}
	if len(q) == 0 {
		return false
	}
	visited := map[string]bool{from: true}
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		if visited[cur] {
			continue
		}
		visited[cur] = true
		if cur == to {
			return true
		}
		for _, n := range adj[cur] {
			if !visited[n] {
				q = append(q, n)
			}
		}
	}
	return false
}
