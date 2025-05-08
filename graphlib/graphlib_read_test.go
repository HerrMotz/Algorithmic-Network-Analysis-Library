package graphlib

import (
	"reflect"
	"testing"
)

var (
	wantRowPtr = []int{0, 2, 4, 6}
	wantColInd = []int{1, 2, 0, 2, 0, 1}
)

func TestReadGraph_EdgeList(t *testing.T) {
	csr, err := ReadGraph("../exampleGraphs/edge_01.graph", false)
	if err != nil {
		t.Fatalf("ReadGraph(edge list) returned error: %v", err)
	}
	checkCSR(t, csr)
}

func TestReadGraph_METIS(t *testing.T) {
	csr, err := ReadGraph("../exampleGraphs/metis_edge_01_out.graph", false)
	if err != nil {
		t.Fatalf("ReadGraph(METIS) returned error: %v", err)
	}
	checkCSR(t, csr)
}

// -------------------------------------------------------------------------
// helpers
// -------------------------------------------------------------------------

func checkCSR(t *testing.T, got *CSRAdjMatrix) {
	t.Helper()

	if !reflect.DeepEqual(got.rowPtr, wantRowPtr) {
		t.Errorf("rowPtr mismatch\n  got : %v\n  want: %v", got.rowPtr, wantRowPtr)
	}
	if !reflect.DeepEqual(got.colInd, wantColInd) {
		t.Errorf("colInd mismatch\n  got : %v\n  want: %v", got.colInd, wantColInd)
	}
	if got.size != 8 {
		t.Errorf("size = %d, want 3", got.size)
	}
}
