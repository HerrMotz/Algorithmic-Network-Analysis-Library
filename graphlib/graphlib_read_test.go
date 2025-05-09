package graphlib

import (
	"reflect"
	"testing"
)

var wantAdj = []map[int]bool{
	{1: true, 2: true, 4: true},
	{0: true, 2: true, 3: true},
	{4: true, 3: true, 1: true, 0: true},
	{1: true, 2: true, 5: true, 6: true},
	{0: true, 2: true, 5: true},
	{4: true, 3: true, 6: true},
	{5: true, 3: true},
}

// -------------------------------------------------------------------------
// tests
// -------------------------------------------------------------------------

func TestReadGraph_EdgeList(t *testing.T) {
	adj, err := ReadGraph("../exampleGraphs/tiny_01.graph", false)
	if err != nil {
		t.Fatalf("ReadGraph(edge list) returned error: %v", err)
	}
	checkAdjList(t, adj)
}

func TestReadGraph_METIS(t *testing.T) {
	adj, err := ReadGraph("../exampleGraphs/metis_tiny_01.mgraph", false)
	if err != nil {
		t.Fatalf("ReadGraph(METIS) returned error: %v", err)
	}
	checkAdjList(t, adj)
}

// -------------------------------------------------------------------------
// helpers
// -------------------------------------------------------------------------

func checkAdjList(t *testing.T, got *AdjList) {
	t.Helper()

	if got.size != len(wantAdj) {
		t.Fatalf("size = %d, want %d", got.size, len(wantAdj))
	}

	for v, gotNbrs := range got.list {
		wantNbrs := wantAdj[v]

		if !reflect.DeepEqual(gotNbrs, wantNbrs) {
			t.Errorf("neighbours of %d mismatch\n  got : %v\n  want: %v",
				v, gotNbrs, wantNbrs)
		}
	}
}
