package graphlib

import (
	"bufio"
	"errors"
	"math"
	"os"
	"reflect"
	"sort"
	"testing"
)

// buildTestFromList liest den Edge-List-Graph aus der Datei ein.
func buildTestFromList(t *testing.T) *AdjList {
	f, err := os.Open("../exampleGraphs/graph_for_test.graph")
	if err != nil {
		t.Fatalf("os.Open failed: %v", err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	g, err := readEdgeListToAdjList(r, false)
	if err != nil && !errors.Is(err, EdgeOrArcExistsError) {
		t.Fatalf("readEdgeListToAdjList returned error: %v", err)
	}
	return g
}

// Tests für Density, MaxDegree und HIndex
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
	wantMax := 11
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

func TestDegeneracy(t *testing.T) {
	g := buildTestFromList(t)
	d, err := g.Degeneracy()
	if err != nil {
		t.Fatalf("Degeneracy returned error: %v", err)
	}
	want := 5
	if d != want {
		t.Errorf("Degeneracy = %d; want %d", d, want)
	}
}

// TestDegeneracyOrder prüft die DegeneracyOrder-Funktion
func TestDegeneracyOrder(t *testing.T) {
	g := buildTestFromList(t)
	order, err := g.DegeneracyOrder()
	if err != nil {
		t.Fatalf("DegeneracyOrder returned error: %v", err)
	}
	// IDs aus der Rückgabe extrahieren
	got := make([]int, len(order))
	for i, v := range order {
		got[i] = v.Id
	}
	// Erwartete IDs 0..n-1
	want := make([]int, len(got))
	for i := range want {
		want[i] = i
	}
	// sortiere beide und vergleiche
	sort.Ints(got)
	sort.Ints(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DegeneracyOrder IDs = %v; want %v", got, want)
	}
}

func TestKCoreDecomposition(t *testing.T) {
	g := buildTestFromList(t)
	L, _, err := g.KCoreDecomp()

	if err != nil {
		t.Fatalf("KCoreDecomp returned error: %v", err)
	}

	// baue map[Vertex]coreNumber aus L
	coreMap := make(map[Vertex]int, g.Size)
	for k, bucket := range L {
		for _, v := range bucket {
			coreMap[v] = k
		}
	}

	want := map[Vertex]int{
		{Id: 0}:  5,
		{Id: 1}:  5,
		{Id: 2}:  5,
		{Id: 3}:  5,
		{Id: 4}:  5,
		{Id: 5}:  5,
		{Id: 6}:  5,
		{Id: 7}:  5,
		{Id: 8}:  4,
		{Id: 9}:  4,
		{Id: 10}: 4,
		{Id: 11}: 5,
		{Id: 12}: 5,
		{Id: 13}: 4,
		{Id: 14}: 4,
	}

	if !reflect.DeepEqual(coreMap, want) {
		t.Errorf("KCoreDecomp coreMap = %v; want %v", coreMap, want)
	}
}
