package graphlib

import (
    "testing"
    "math"
    "reflect"
	"errors"
)


func buildTestFromList(t *testing.T) *AdjList {
    g, err := ReadGraphAdjList("../exampleGraphs/graph_for_test.graph", false)
    if err != nil && !errors.Is(err, EdgeOrArcExistsError) {
		t.Fatalf("ReadGraph_AdjList(edge list) returned error: %v", err)
	}
    return g
}

// Tests for the graph properties: Density, MaxDegree, and HIndex.
func TestDensity_MaxDegree_HIndex(t *testing.T) {
    g := buildTestFromList(t)

    // Density
    dens, err := g.Density()
    if err != nil {
        t.Fatalf("Density returned error: %v", err)
    }
    wantD := 0.47619047619047616
    if math.Abs(dens-wantD) > 1e-9 {
        t.Errorf("Density = %v; want %v", dens, wantD)
    }

    // MaxDegree
    maxDeg, err := g.MaxDegree()
    if err != nil {
        t.Fatalf("MaxDegree returned error: %v", err)
    }
    wantMax := 10
    if maxDeg != wantMax {
        t.Errorf("MaxDegree = %d; want %d", maxDeg, wantMax)
    }

    // HIndex
    h, err := g.HIndex()
    if err != nil {
        t.Fatalf("HIndex returned error: %v", err)
    }
    wantH := 7
    if h != wantH {
        t.Errorf("HIndex = %d; want %d", h, wantH)
    }
}

// TestDegeneracyOrdering tests the degeneracy ordering of the graph.
func TestDegeneracyOrdering(t *testing.T) {
    g := buildTestFromList(t)
    order, err := g.DegeneracyOrder()
    if err != nil {
        t.Fatalf("DegeneracyOrdering returned error: %v", err)
    }
    wantIDs := []int{6, 7, 8, 2, 9, 4, 5, 12, 11, 3, 0, 1, 10, 13, 14}
    if len(order) != len(wantIDs) {
        t.Fatalf("DegeneracyOrdering length = %d; want %d", len(order), len(wantIDs))
    }
    got := make([]int, len(order))
    for i, v := range order {
        got[i] = v.Id
    }
    if !reflect.DeepEqual(got, wantIDs) {
        t.Errorf("DegeneracyOrdering = %v; want %v", got, wantIDs)
    }
}

// TestKCoreDecomposition tests the k-core decomposition of the graph.
func TestKCoreDecomposition(t *testing.T) {
    g := buildTestFromList(t)
    cores, _, err := g.KCoreDecomp()
    if err != nil {
        t.Fatalf("KCoreDecomposition returned error: %v", err)
    }
    want := map[Vertex]int{
        {Id: 0}: 5,
        {Id: 1}: 5,
        {Id: 2}: 4,
        {Id: 3}: 5,
        {Id: 4}: 5,
        {Id: 5}: 5,
        {Id: 6}: 3,
        {Id: 7}: 4,
        {Id: 8}: 4,
        {Id: 9}: 5,
        {Id: 10}: 5,
        {Id: 11}: 5,
        {Id: 12}: 5,
        {Id: 13}: 5,
        {Id: 14}: 5,
    }
    if !reflect.DeepEqual(cores, want) {
        t.Errorf("KCoreDecomposition = %v; want %v", cores, want)
    }
}