package graphlib

import (
	"fmt"
	"sort"
	"testing"
)

// buildSimpleGraph creates a small undirected graph with one triangle {0,1,2} and an extra edge 2-3.
func buildSimpleGraph(t *testing.T) *AdjList {
	g, err := InitializeAdjList(4)
	if err != nil {
		t.Fatalf("InitializeAdjList failed: %v", err)
	}
	edges := []Edge{
		{U: Vertex{0}, V: Vertex{1}},
		{U: Vertex{1}, V: Vertex{2}},
		{U: Vertex{2}, V: Vertex{0}},
		{U: Vertex{2}, V: Vertex{3}},
	}
	for _, e := range edges {
		if err := g.AddEdge(e); err != nil {
			t.Fatalf("AddEdge %v failed: %v", e, err)
		}
	}
	return g
}

// helper to normalize and compare triangle sets
func compareTriangles(t *testing.T, got [][3]Vertex, want [][]int) {
	format := func(tri [3]Vertex) string {
		ids := []int{tri[0].Id, tri[1].Id, tri[2].Id}
		sort.Ints(ids)
		return fmt.Sprintf("%d,%d,%d", ids[0], ids[1], ids[2])
	}
	gotSet := make(map[string]bool)
	for _, tri := range got {
		gotSet[format(tri)] = true
	}
	wantSet := make(map[string]bool)
	for _, tri := range want {
		key := fmt.Sprintf("%d,%d,%d", tri[0], tri[1], tri[2])
		wantSet[key] = true
	}
	if len(gotSet) != len(wantSet) {
		t.Fatalf("number of triangles = %d; want %d", len(gotSet), len(wantSet))
	}
	for key := range wantSet {
		if !gotSet[key] {
			t.Errorf("missing triangle %s", key)
		}
	}
}

func TestTrianglesNodeIterator(t *testing.T) {
	g := buildSimpleGraph(t)
	tris, err := g.TrianglesNodeIterator()
	if err != nil {
		t.Fatalf("TrianglesNodeIterator returned error: %v", err)
	}
	want := [][]int{{0, 1, 2}}
	compareTriangles(t, tris, want)
}

func TestTrianglesCNI(t *testing.T) {
	g := buildSimpleGraph(t)
	tris, err := g.TrianglesCNI()
	if err != nil {
		t.Fatalf("TrianglesCNI returned error: %v", err)
	}
	want := [][]int{{0, 1, 2}}
	compareTriangles(t, tris, want)
}
