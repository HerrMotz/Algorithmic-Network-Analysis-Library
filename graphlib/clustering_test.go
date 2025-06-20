package graphlib

import (
	"math"
	"testing"
)

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

	modularity, err := graph.PartialModularity(C)
	if err != nil {
		t.Fatalf("PartialModularity calculation failed: %v", err)
	}

	// Manually compute expected modularity:
	// |E(C)| = 3, m = 3, deg(0)=2, deg(1)=2, deg(2)=2 => sum_deg = 6
	// q = 3/3 - (6^2)/(4*3^2) = 1 - 36 / 36 = 0

	expected := 0.0
	if math.Abs(modularity-expected) > 1e-9 {
		t.Errorf("PartialModularity = %f, expected %f", modularity, expected)
	}
}
