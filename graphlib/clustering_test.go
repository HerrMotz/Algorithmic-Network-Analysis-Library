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

	modularity, err := partialModularityMetric(graph, C)
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

	// Helper function to compare clusters ignoring order
	match := func(a, b VertexSet) bool {
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

	// For each wantCluster, check that it's found in clusters
	for _, want := range wantClusters {
		found := false
		for _, got := range clusters {
			// Convert got to VertexSet if necessary
			gotSet := VertexSet{}
			for v := range got {
				gotSet[v] = struct{}{}
			}
			if match(want, gotSet) {
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

// Test function for HalveClustering using the modularity Metric
func TestHalveClustering(t *testing.T) {
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

	clusters, err := graph.HalveClustering(partialModularityMetric, partialModularityCompare)
	if err != nil {
		t.Fatalf("HalveClustering returned error: %v", err)
	}

	// The optimal clustering (with minimal cut) is the two clusters:
	// {0,1,2} (the triangle) and {3} (isolated)
	wantClusters := []VertexSet{
		{0: struct{}{}, 1: struct{}{}, 2: struct{}{}}, // Triangle
		{3: struct{}{}}, // Isolated node
	}

	// Helper function to compare clusters ignoring order
	match := func(a, b VertexSet) bool {
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

	// For each wantCluster, check that it's found in clusters
	for _, want := range wantClusters {
		found := false
		for _, got := range clusters {
			gotSet := VertexSet{}
			for v := range got {
				gotSet[v] = struct{}{}
			}
			if match(want, gotSet) {
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
