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
	H := AdjList{List: adj1, Size: len(adj1)}
	inducedGraph := InducedGraph{AdjList: &H, GoodVertices: VertexSet{0: {}, 1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}}}
	plex, err := inducedGraph.TwoPlexNaive()
	if err != nil {
		t.Fatalf("TestTwoPlexCorrectness: unexpected error %v", err)
	}

	if !reflect.DeepEqual(wantPlex1, plex) {
		t.Errorf("TestTwoPlexCorrectness: want %v got %v", wantPlex1, plex)
	}

	// test what happens if the induced graph is empty
	H = AdjList{List: adj2, Size: len(adj2)}
	inducedGraph = InducedGraph{AdjList: &H, GoodVertices: VertexSet{0: {}, 1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}}}
	plex, err = inducedGraph.TwoPlexNaive()
	if err != nil {
		t.Fatalf("TestTwoPlexCorrectness: unexpected error %v", err)
	}
	if plex.Size() != wantPlex2 {
		t.Errorf("TestTwoPlexCorrectness: want %d got %d", wantPlex2, plex.Size())
	}
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
