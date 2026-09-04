package stategraph

import (
	"encoding/json"
	"os"
	"testing"
)

// real.tfstate is a genuine terraform.tfstate produced by `terraform apply` on a
// config exercising count, for_each, a child module, a transitive dependency and
// a sensitive attribute.
func loadRealGraph(t *testing.T) Graph {
	t.Helper()
	data, err := os.ReadFile("testdata/real.tfstate.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	out, err := BuildGraphFromState(data)
	if err != nil {
		t.Fatalf("BuildGraphFromState returned error: %v", err)
	}
	var g Graph
	if err := json.Unmarshal(out, &g); err != nil {
		t.Fatalf("output is not a valid Graph: %v", err)
	}
	return g
}

func nodeByAddr(g Graph, addr string) *Node {
	for i := range g.Nodes {
		if g.Nodes[i].Addr == addr {
			return &g.Nodes[i]
		}
	}
	return nil
}

func TestBuildGraphFromStateGroupsInstances(t *testing.T) {
	g := loadRealGraph(t)

	// count = 3 must collapse into a single node carrying 3 instances.
	workers := nodeByAddr(g, "null_resource.workers")
	if workers == nil {
		t.Fatalf("null_resource.workers node missing, got nodes: %+v", g.Nodes)
	}
	if workers.InstancesCount != 3 {
		t.Errorf("workers instances_count = %d, want 3", workers.InstancesCount)
	}

	// for_each with 2 keys must collapse into a single node with 2 instances.
	regions := nodeByAddr(g, "null_resource.regions")
	if regions == nil {
		t.Fatalf("null_resource.regions node missing")
	}
	if regions.InstancesCount != 2 {
		t.Errorf("regions instances_count = %d, want 2", regions.InstancesCount)
	}

	// A resource inside a child module must be present and tagged with its module.
	inner := nodeByAddr(g, "module.nested.null_resource.inner")
	if inner == nil {
		t.Fatalf("module.nested.null_resource.inner node missing")
	}
	if inner.Module == "" {
		t.Errorf("module resource has empty Module field")
	}
}

func TestBuildGraphFromStateEdgesResolveToNodes(t *testing.T) {
	g := loadRealGraph(t)

	if len(g.Edges) == 0 {
		t.Fatal("no edges built from a state that has dependencies")
	}

	ids := map[string]bool{}
	for _, n := range g.Nodes {
		ids[n.ID] = true
	}
	for _, e := range g.Edges {
		if !ids[e.From] {
			t.Errorf("edge %s->%s: From is not an existing node id", e.From, e.To)
		}
		if !ids[e.To] {
			t.Errorf("edge %s->%s: To is not an existing node id", e.From, e.To)
		}
		if e.From == e.To {
			t.Errorf("self edge on %s", e.From)
		}
	}

	// Dependencies must be de-duplicated: count=3 all depending on random_pet.base
	// must not yield 3 identical edges.
	seen := map[string]int{}
	for _, e := range g.Edges {
		seen[e.From+"->"+e.To]++
	}
	for k, n := range seen {
		if n > 1 {
			t.Errorf("duplicate edge %s appears %d times", k, n)
		}
	}
}

func TestBuildGraphFromStateFiltersSensitiveAttributes(t *testing.T) {
	g := loadRealGraph(t)

	secret := nodeByAddr(g, "random_password.secret")
	if secret == nil {
		t.Fatal("random_password.secret node missing")
	}
	// The generated password is flagged sensitive in state; it must not leak
	// into the graph served to the dashboard.
	raw, err := json.Marshal(secret)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, inst := range secret.Instances {
		if v, ok := inst.Attributes["result"]; ok {
			t.Errorf("sensitive attribute 'result' exposed in graph: %v (node json: %s)", v, raw)
		}
	}
}

func TestBuildGraphFromStateRejectsInvalidJSON(t *testing.T) {
	if _, err := BuildGraphFromState([]byte("not json")); err == nil {
		t.Fatal("expected an error on malformed state, got nil")
	}
}
