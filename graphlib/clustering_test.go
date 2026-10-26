package graphlib

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

var testAdjTwoCliquesBridge = []VertexSet{
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

var testAdjCliqueIsolates = []VertexSet{
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

//	    5
//	    |
//	    2
//	   / \
//	  /   \
//	 0-----1
//	/       \
//
// 3         4
var testBeyblade = []VertexSet{
	// 0
	{1: Naught{}, 2: Naught{}, 3: Naught{}},
	// 1
	{0: Naught{}, 2: Naught{}, 4: Naught{}},
	// 2
	{0: Naught{}, 1: Naught{}, 5: Naught{}},
	// 3
	{0: Naught{}},
	// 4
	{1: Naught{}},
	// 5
	{2: Naught{}},
}

// compareVertexSets is a helper function to compare clusters ignoring order
func compareVertexSets(a, b VertexSet) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
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

	modularity, err := PartialModularityMetric(graph, C)
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

func TestMergeClustering(t *testing.T) {
	tests := []struct {
		name string
		adj  []VertexSet
		want []VertexSet // expected clusters
	}{
		{
			name: "two cliques and bridge",
			adj:  testAdjTwoCliquesBridge,
			want: []VertexSet{
				{0: Naught{}, 1: Naught{}, 2: Naught{}},
				{3: Naught{}, 4: Naught{}, 5: Naught{}},
			},
		},
		{
			name: "clique and isolates",
			adj:  testAdjCliqueIsolates,
			want: []VertexSet{
				{0: Naught{}, 1: Naught{}, 2: Naught{}}, // clique
				{3: Naught{}},                           // isolate
				{4: Naught{}},                           // isolate
			},
		},
		{
			name: "weird circle", // test the special case where leafs are put in a cluster with their neighbours
			adj:  testBeyblade,
			want: []VertexSet{
				{0: Naught{}, 3: Naught{}}, // leaf 3 is in a cluster with 0
				{1: Naught{}, 4: Naught{}}, // leaf 4 is in a cluster with 1
				{2: Naught{}, 5: Naught{}}, // leaf 5 is in a cluster with 2
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

			got, err := adj.MergeClusteringGreedyBFS(PartialModularityMetric, PartialModularityCompare)
			if err != nil {
				t.Fatalf("MergeClusteringGreedyBFS failed: %v", err)
			}

			// Check that every wanted cluster is present in the result, order does not matter
			if len(got) != len(tc.want) {
				t.Errorf("MergeClusteringGreedyBFS failed: got %d clusters, want %d", len(got), len(tc.want))
			}
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

func TestComponentClustering(t *testing.T) {
	// Create a graph:
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

	clusters := graph.ComponentClustering()

	// Build a set representation for expected clusters
	wantClusters := []VertexSet{
		{0: struct{}{}, 1: struct{}{}, 2: struct{}{}}, // Triangle cluster
		{3: struct{}{}}, // Isolated node
	}

	// For each wantCluster, check that it's found in clusters
	for _, want := range wantClusters {
		found := false
		for _, got := range clusters {
			// Convert got to VertexSet if necessary
			gotSet := VertexSet{}
			for v := range got {
				gotSet[v] = struct{}{}
			}
			if compareVertexSets(want, gotSet) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected cluster %v not found in result %v", want, clusters)
		}
	}

	// Also check that the number of clusters is correct
	if len(clusters) != len(wantClusters) {
		t.Errorf("Expected %d clusters, got %d", len(wantClusters), len(clusters))
	}
}

func TestHalveClustering(t *testing.T) {
	tests := []struct {
		name string
		adj  []VertexSet
		want []VertexSet // expected clusters
	}{
		{
			name: "two cliques and bridge",
			adj:  testAdjTwoCliquesBridge,
			want: []VertexSet{
				{0: Naught{}, 1: Naught{}, 2: Naught{}}, // clique
				{3: Naught{}, 4: Naught{}, 5: Naught{}}, // clique
			},
		},
		{
			name: "clique and isolates",
			adj:  testAdjCliqueIsolates,
			want: []VertexSet{
				{0: Naught{}, 1: Naught{}, 2: Naught{}}, // clique
				{3: Naught{}},                           // isolate
				{4: Naught{}},                           // isolate
			},
		},
		{
			name: "weird circle", // test the special case where leafs are put in a cluster with their neighbours
			adj:  testBeyblade,
			want: []VertexSet{
				{0: Naught{}, 3: Naught{}}, // leaf 3 is in a cluster with 0
				{1: Naught{}, 4: Naught{}}, // leaf 4 is in a cluster with 1
				{2: Naught{}, 5: Naught{}}, // leaf 5 is in a cluster with 2
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

			got, err := adj.HalveClustering(PartialModularityMetric, PartialModularityCompare)
			if err != nil {
				t.Fatalf("HalveClustering failed: %v", err)
			}

			// Check that every wanted cluster is present in the result, order does not matter
			if len(got) != len(tc.want) {
				t.Errorf("HalveClustering failed: got %d clusters, want %d", len(got), len(tc.want))
			}
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

func TestICEClustering(t *testing.T) {
	tests := []struct {
		name string
		adj  []VertexSet
		want []VertexSet // expected clusters
	}{
		{
			name: "two cliques and bridge",
			adj:  testAdjTwoCliquesBridge,
			want: []VertexSet{
				{0: Naught{}, 1: Naught{}, 2: Naught{}},
				{3: Naught{}, 4: Naught{}, 5: Naught{}},
			},
		},
		{
			name: "clique and isolates",
			adj:  testAdjCliqueIsolates,
			want: []VertexSet{
				{0: Naught{}, 1: Naught{}, 2: Naught{}}, // clique
				{3: Naught{}},                           // isolate
				{4: Naught{}},                           // isolate
			},
		},
		{
			name: "weird circle", // test the special case where leafs are put in a cluster with their neighbours
			adj:  testBeyblade,
			want: []VertexSet{
				{0: Naught{}, 3: Naught{}}, // leaf 3 is in a cluster with 0
				{1: Naught{}, 4: Naught{}}, // leaf 4 is in a cluster with 1
				{2: Naught{}, 5: Naught{}}, // leaf 5 is in a cluster with 2
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

			got, err := adj.ICEClustering(PartialModularityMetric, PartialModularityCompare)
			if err != nil {
				t.Fatalf("ICEClustering failed: %v", err)
			}

			// Check that every wanted cluster is present in the result, order does not matter
			if len(got) != len(tc.want) {
				t.Errorf("ICEClustering failed: got %d clusters, want %d", len(got), len(tc.want))
			}
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

// TestClusteringIsReproducible runs every heuristic twice on the same graph and
// expects identical clusterings in identical order. Before the heuristics
// iterated in sorted order, the results depended on Go's random map order.
func TestClusteringIsReproducible(t *testing.T) {
	g, err := ReadGraphAdjList("../exampleGraphs/abi.edgelist", false)
	if err != nil {
		t.Fatalf("reading graph: %v", err)
	}

	heuristics := map[string]func() ([]VertexSet, error){
		"MergeClusteringGreedyBFS": func() ([]VertexSet, error) {
			return g.MergeClusteringGreedyBFS(PartialModularityMetric, PartialModularityCompare)
		},
		"MergeClusteringMetricWithRandomPivot": func() ([]VertexSet, error) {
			return g.MergeClusteringMetricWithRandomPivot(rand.New(rand.NewSource(7)), PartialModularityMetric, PartialModularityCompare)
		},
		"HalveClustering": func() ([]VertexSet, error) {
			return g.HalveClustering(PartialDisagreementMetric, PartialDisagreementCompare)
		},
	}

	for name, run := range heuristics {
		t.Run(name, func(t *testing.T) {
			first, err := run()
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 5; i++ {
				again, err := run()
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(first, again) {
					t.Fatalf("run %d differs from the first run:\n  first: %v\n  again: %v", i+2, first, again)
				}
			}
		})
	}
}
