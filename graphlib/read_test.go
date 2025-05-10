package graphlib

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// ---------------------------------------
//  Tests reading from edgelist and METIS
// ---------------------------------------

var wantAdj = []map[int]bool{
	{1: true, 2: true, 4: true},
	{0: true, 2: true, 3: true},
	{4: true, 3: true, 1: true, 0: true},
	{1: true, 2: true, 5: true, 6: true},
	{0: true, 2: true, 5: true},
	{4: true, 3: true, 6: true},
	{5: true, 3: true},
}

// -------
// tests
// -------

func TestReadGraph_EdgeList(t *testing.T) {
	adj, err := ReadGraphAdjList("../exampleGraphs/tiny_01.graph", false)
	if err != nil && !errors.Is(err, EdgeOrArcExistsError) {
		t.Fatalf("ReadGraph_AdjList(edge list) returned error: %v", err)
	}
	checkAdjList(t, adj)
}

func TestReadGraph_METIS(t *testing.T) {
	adj, err := ReadGraphAdjList("../exampleGraphs/metis_tiny_01.mgraph", false)
	if err != nil && !errors.Is(err, EdgeOrArcExistsError) {
		t.Fatalf("ReadGraph_AdjList(METIS) returned error: %v", err)
	}
	checkAdjList(t, adj)
}

// ---------
// helpers
// ---------

func checkAdjList(t *testing.T, got *AdjList) {
	t.Helper()

	if got == nil {
		t.Fatalf("got nil")
	}

	if got.Size != len(wantAdj) {
		t.Fatalf("size = %d, want %d", got.Size, len(wantAdj))
	}

	for v, gotNbrs := range got.List {
		wantNbrs := wantAdj[v]

		if !reflect.DeepEqual(gotNbrs, wantNbrs) {
			t.Errorf("neighbours of %d mismatch\n  got : %v\n  want: %v",
				v, gotNbrs, wantNbrs)
		}
	}
}

// ----------------------------
// File read benchmark
// ----------------------------

func BenchmarkReadGraph_EDGELIST_undirected(b *testing.B) {
	files := []string{
		"tiny_01.graph",
		"Wiki-Vote.txt",
	}

	for _, file := range files {
		b.Run(fmt.Sprintf("EDGELIST/%s", file), func(b *testing.B) {
			benchReadGraph(b, file)
		})
	}
}

func BenchmarkReadGraph_METIS_undirected(b *testing.B) {
	files := []string{
		"metis_tiny_01.mgraph",
		"4elt.mgraph",
	}

	for _, file := range files {
		b.Run(fmt.Sprintf("METIS/%s", file), func(b *testing.B) {
			benchReadGraph(b, file)
		})
	}
}

func benchReadGraph(b *testing.B, file string) {
	b.ReportAllocs()
	l, err := ReadGraphAdjList("../exampleGraphs/"+file, false)

	if err != nil {
		b.Fatalf("ReadGraph_AdjList returned error: %v", err)
	}

	if l == nil {
		b.Fatalf("ReadGraph_AdjList returned nil")
	}

	if l.Size == 0 {
		b.Fatalf("ReadGraph_AdjList returned empty graph")
	}

	ok := l.HasVertex(Vertex{0})
	if !ok {
		b.Fatalf("ReadGraph_AdjList returned graph without vertex 0")
	}
}
