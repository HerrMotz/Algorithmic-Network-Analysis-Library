package graphlib

import (
	"reflect"
	"testing"
)

var adj1 = []VertexSet{
	{1: Naught{}, 2: Naught{}},
	{0: Naught{}, 3: Naught{}},
	{0: Naught{}, 3: Naught{}, 4: Naught{}},
	{1: Naught{}, 2: Naught{}, 5: Naught{}},
	{2: Naught{}},
	{3: Naught{}},
	{},
}
var wantPlex1 = VertexSet{0: {}, 1: {}, 2: {}, 3: {}}

var adj2 = []VertexSet{
	{},
	{},
	{},
	{},
	{},
	{},
	{},
}
var wantPlex2 = 2

func TestTwoPlexCorrectness(t *testing.T) {
	GoodVertices := VertexSet{0: {}, 1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}}

	// verification graph
	H1 := AdjList{List: adj1, Size: len(adj1)}
	iGraph1 := InducedGraph{AdjList: &H1, GoodVertices: GoodVertices}

	// empty graph
	H2 := AdjList{List: adj2, Size: len(adj2)}
	iGraph2 := InducedGraph{AdjList: &H2, GoodVertices: GoodVertices}

	// ================================
	// Test TwoPlexNaive (not parallel)
	// ================================
	plex, err := iGraph1.TwoPlexNaive()
	if err != nil {
		t.Fatalf("TwoPlexNaive: unexpected error %v", err)
	}

	if !reflect.DeepEqual(wantPlex1, plex) {
		t.Errorf("TwoPlexNaive: want %v got %v", wantPlex1, plex)
	}

	plex, err = iGraph2.TwoPlexNaive()
	if err != nil {
		t.Fatalf("TwoPlexNaive: unexpected error %v", err)
	}
	if plex.Size() != wantPlex2 {
		t.Errorf("TwoPlexNaive: want %d got %d", wantPlex2, plex.Size())
	}

	// =========================
	// Test TwoPlexNaiveParallel
	// =========================
	plex, err = iGraph1.TwoPlexNaiveParallel()
	if err != nil {
		t.Fatalf("TwoPlexNaiveParallel: unexpected error %v", err)
	}
	if !reflect.DeepEqual(wantPlex1, plex) {
		t.Errorf("TwoPlexNaive: want %v got %v", wantPlex1, plex)
	}

	// =========================
	// Test PermanentInducedGraph.TwoPlexPermanent
	// =========================
	pGraph := PermanentInducedGraph{iGraph1, VertexSet{}}
	plex, err = pGraph.TwoPlexPermanent()
	if err != nil {
		t.Fatalf("TwoPlexPermanentInduced: unexpected error %v", err)
	}
	if !reflect.DeepEqual(wantPlex1, plex) {
		t.Errorf("TwoPlexPermanentInduced: want %v got %v", wantPlex1, plex)
	}

	// =========================
	// Test PermanentInducedGraph.TwoPlexPermanentParallel
	// =========================
	plex, err = pGraph.TwoPlexPermanentParallel()
	if err != nil {
		t.Fatalf("TwoPlexPermanentParallel unexpected error %v", err)
	}
	if !reflect.DeepEqual(wantPlex1, plex) {
		t.Errorf("TwoPlexPermanentParallel: want %v got %v", wantPlex1, plex)
	}

	// Test PermanentInducedGraph.TwoPlexPermanentParallel on empty graph
	pGraph2 := PermanentInducedGraph{iGraph2, VertexSet{}}
	plex, err = pGraph2.TwoPlexPermanentParallel()
	if err != nil {
		t.Fatalf("TwoPlexPermanentParallel on empty graph: unexpected error %v", err)
	}
	if plex.Size() != wantPlex2 {
		t.Errorf("TwoPlexPermanentParallel on empty graph: want %d got %d", wantPlex2, plex.Size())
	}

	// =========================
	// Test PermanentInducedGraph.TwoPlexDistance
	// =========================
	/*plex, err = pGraph.TwoPlexDistance()
	if err != nil {
		t.Fatalf("TwoPlexDistance unexpected error %v", err)
	}
	if !reflect.DeepEqual(wantPlex1, plex) {
		t.Errorf("TwoPlexDistance: want %v got %v", wantPlex1, plex)
	}*/

}

func TestPermanentInducedGraph_TwoPlexPermanent(t *testing.T) {
	H := AdjList{List: adj1, Size: len(adj1)}
	permanentInducedGraph := PermanentInducedGraph{InducedGraph{AdjList: &H, GoodVertices: VertexSet{0: {}, 1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}}}, VertexSet{}}
	plex, err := permanentInducedGraph.TwoPlexPermanent()
	if err != nil {
		t.Fatalf("TestPermanentInducedGraph_TwoPlexPermanent: unexpected error %v", err)
	}
	if !reflect.DeepEqual(wantPlex1, plex) {
		t.Errorf("TestPermanentInducedGraph_TwoPlexPermanent: want %v got %v", wantPlex1, plex)
	}
}

func BenchmarkInducedGraph_TwoPlexNaive(b *testing.B) {
	H, err := ReadGraphAdjList("../exampleGraphs/metis_4elt.mgraph", false)
	if err != nil {
		b.Fatalf("failed to read erdos graph: %v", err)
	}
	for i := 0; i < b.N; i++ {
		// create vertex set from H
		GoodVertices := make(VertexSet, H.Size)
		for v := 0; v < H.Size; v++ {
			GoodVertices[v] = Naught{}
		}
		inducedGraph := InducedGraph{AdjList: H, GoodVertices: GoodVertices}
		plex, err := inducedGraph.TwoPlexNaive()
		if plex != nil {
			b.Logf("PermanentInducedGraph_TwoPlexPermanent: %v", plex.Size())
		} else {
			b.Fatalf("PermanentInducedGraph_TwoPlexPermanent: No plex returned!")
		}
		if err != nil {
			b.Fatalf("BenchmarkInducedGraph_TwoPlexNaive: unexpected error %v", err)
		}
	}
}

func BenchmarkInducedGraph_TwoPlexNaiveParallel(b *testing.B) {
	H, err := ReadGraphAdjList("../exampleGraphs/metis_4elt.mgraph", false)
	if err != nil {
		b.Fatalf("failed to read erdos graph: %v", err)
	}
	// create vertex set from H
	GoodVertices := make(VertexSet, H.Size)
	for v := 0; v < H.Size; v++ {
		GoodVertices[v] = Naught{}
	}
	inducedGraph := InducedGraph{AdjList: H, GoodVertices: GoodVertices}
	plex, err := inducedGraph.TwoPlexNaiveParallel()
	if plex != nil {
		b.Logf("PermanentInducedGraph_TwoPlexPermanent: %v", plex.Size())
	} else {
		b.Fatalf("PermanentInducedGraph_TwoPlexPermanent: No plex returned!")
	}
	if err != nil {
		b.Fatalf("BenchmarkInducedGraph_TwoPlexNaive: unexpected error %v", err)
	}
}

func BenchmarkPermanentInducedGraph_TwoPlexPermanent(b *testing.B) {
	H, err := ReadGraphAdjList("../exampleGraphs/metis_4elt.mgraph", false)
	if err != nil {
		b.Fatalf("failed to read erdos graph: %v", err)
	}
	// create vertex set from H
	GoodVertices := make(VertexSet, H.Size)
	for v := 0; v < H.Size; v++ {
		GoodVertices[v] = Naught{}
	}
	permanentInducedGraph := PermanentInducedGraph{InducedGraph{AdjList: H, GoodVertices: GoodVertices}, VertexSet{}}
	plex, err := permanentInducedGraph.TwoPlexPermanent()
	// write the size of the plex to stdout
	if plex != nil {
		b.Logf("PermanentInducedGraph_TwoPlexPermanent: %v", plex.Size())
	} else {
		b.Fatalf("PermanentInducedGraph_TwoPlexPermanent: No plex returned!")
	}
	if err != nil {
		b.Fatalf("BenchmarkPermanentInducedGraph_TwoPlexPermanent: unexpected error %v", err)
	}
}

func BenchmarkInducedGraph_TwoPlexPermanentParallel(b *testing.B) {
	H, err := ReadGraphAdjList("../exampleGraphs/metis_4elt.mgraph", false)
	if err != nil {
		b.Fatalf("failed to read erdos graph: %v", err)
	}
	// create vertex set from H
	GoodVertices := make(VertexSet, H.Size)
	for v := 0; v < H.Size; v++ {
		GoodVertices[v] = Naught{}
	}
	pGraph := PermanentInducedGraph{InducedGraph{AdjList: H, GoodVertices: GoodVertices}, VertexSet{}}
	plex, err := pGraph.TwoPlexPermanentParallel()
	if plex != nil {
		b.Logf("PermanentInducedGraph_TwoPlexPermanent: %v", plex.Size())
	} else {
		b.Fatalf("PermanentInducedGraph_TwoPlexPermanent: No plex returned!")
	}
	if err != nil {
		b.Fatalf("BenchmarkInducedGraph_TwoPlexNaive: unexpected error %v", err)
	}
}
