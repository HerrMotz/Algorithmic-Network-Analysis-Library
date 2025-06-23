package graphlib

import (
	"math"
	"reflect"
	"testing"
)

var test_adj_two_cliques_bridge = []VertexSet{
	// 0
	{1: Naught{}, 2: Naught{}}, // 0 connects to 1 and 2 (clique 1)
	// 1
	{0: Naught{}, 2: Naught{}}, // 1 connects to 0 and 2 (clique 1)
	// 2
	{0: Naught{}, 1: Naught{}, 3: Naught{}}, // 2 connects to 0, 1 (clique 1), and 3 (bridge)
	// 3
	{2: Naught{}, 4: Naught{}, 5: Naught{}}, // 3 connects to 2 (bridge), 4, 5 (clique 2)
	// 4
	{3: Naught{}, 5: Naught{}}, // 4 connects to 3 and 5 (clique 2)
	// 5
	{3: Naught{}, 4: Naught{}}, // 5 connects to 3 and 4 (clique 2)
}

var test_adj_clique_isolates = []VertexSet{
	// 0
	{1: Naught{}, 2: Naught{}}, // 0 connects to 1 and 2 (clique)
	// 1
	{0: Naught{}, 2: Naught{}}, // 1 connects to 0 and 2 (clique)
	// 2
	{0: Naught{}, 1: Naught{}}, // 2 connects to 0 and 1 (clique)
	// 3
	{}, // 3 is isolated
	// 4
	{}, // 4 is isolated
}

func TestPartialModularity(t *testing.T) {
	// Create a graph with 4 nodes and 3 edges forming a triangle + 1 isolated node:
	// 0 -- 1
	//  \  /
	//   2
	// 3 (isolated)
	graph, err := InitializeAdjList(4)
	if err != nil {
		t.Fatalf("Failed to initialize graph: %v", err)
	}

	edges := []Edge{
		{U: Vertex{0}, V: Vertex{1}},
		{U: Vertex{1}, V: Vertex{2}},
		{U: Vertex{2}, V: Vertex{0}},
	}
	for _, e := range edges {
		if err := graph.AddEdge(e); err != nil {
			t.Fatalf("Failed to add edge %v: %v", e, err)
		}
	}

	// Define a cluster {0, 1, 2}
	C := VertexSet{
		0: Naught{},
		1: Naught{},
		2: Naught{},
	}

	modularity, err := partialModularityMetric(graph, C)
	if err != nil {
		t.Fatalf("partialModularityMetric calculation failed: %v", err)
	}

	// Manually compute expected modularity:
	// |E(C)| = 3, m = 3, deg(0)=2, deg(1)=2, deg(2)=2 => sum_deg = 6
	// q = 3/3 - (6^2)/(4*3^2) = 1 - 36 / 36 = 0

	expected := 0.0
	if math.Abs(modularity-expected) > 1e-9 {
		t.Errorf("partialModularityMetric = %f, expected %f", modularity, expected)
	}
}

func TestPartialModularityTemporary(t *testing.T) {
	adj := &AdjList{
		List: test_adj_two_cliques_bridge,
		Size: len(test_adj_two_cliques_bridge),
	}

	lol := VertexSet{
		0: Naught{}, 1: Naught{}, 2: Naught{}, 3: Naught{}, 4: Naught{}, 5: Naught{},
	}

	got, err := partialModularityMetric(adj, lol)

	if err != nil {
		t.Fatalf("MergeClustering failed: %v", err)
	}

	if got < 10000 {
		t.Errorf("Expected 2 clusters, got %f", got)
	}
}

func TestMergeClustering(t *testing.T) {
	tests := []struct {
		name string
		adj  []VertexSet
		want []VertexSet // expected clusters
	}{
		{
			name: "two cliques and bridge",
			adj:  test_adj_two_cliques_bridge,
			want: []VertexSet{
				{0: Naught{}, 1: Naught{}, 2: Naught{}, 3: Naught{}, 4: Naught{}, 5: Naught{}}, // all connected
			},
		},
		{
			name: "clique and isolates",
			adj:  test_adj_clique_isolates,
			want: []VertexSet{
				{0: Naught{}, 1: Naught{}, 2: Naught{}}, // clique
				{3: Naught{}},                           // isolate
				{4: Naught{}},                           // isolate
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			adj := &AdjList{
				List: make([]VertexSet, len(tc.adj)),
				Size: len(tc.adj),
			}
			for i, neighbors := range tc.adj {
				adj.List[i] = make(VertexSet)
				for n := range neighbors {
					adj.List[i][n] = Naught{}
				}
			}

			got, err := adj.MergeClustering(partialModularityMetric, partialModularityCompare)
			if err != nil {
				t.Fatalf("MergeClustering failed: %v", err)
			}

			// Check that every wanted cluster is present in the result, order does not matter
			for _, wantCluster := range tc.want {
				found := false
				for _, gotCluster := range got {
					if reflect.DeepEqual(gotCluster, wantCluster) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected cluster %v not found in result %v", wantCluster, got)
				}
			}
			// Optionally check size as well
			if len(got) != len(tc.want) {
				t.Errorf("Got %d clusters, want %d", len(got), len(tc.want))
			}
		})
	}
}
