package graphlib

import (
	"reflect"
	"sort"
	"testing"
)

func TestCliquesCorrectness(t *testing.T) {
	adj := []VertexSet{
		{1: Naught{}, 2: Naught{}},
		{0: Naught{}, 3: Naught{}},
		{0: Naught{}, 3: Naught{}, 4: Naught{}},
		{1: Naught{}, 2: Naught{}, 5: Naught{}},
		{2: Naught{}},
		{3: Naught{}},
	}
	H := AdjList{List: adj, Size: len(adj)}

	cliques := H.BronKerboschDegeneracy()
	var largest []int
	for _, c := range cliques {
		if len(c) > len(largest) {
			largest = c
		}
	}
	sort.Ints(largest)

	// Erwartete maximale Clique-Größe
	expectedSize := 2

	if len(largest) != expectedSize {
		t.Errorf("Größte Clique hat unerwartete Größe: erwartet %d, erhalten %d (Knoten: %v)", expectedSize, len(largest), largest)
	}

	validCliques := [][]int{
		{0, 1},
		{0, 2},
		{1, 3},
		{2, 3},
		{2, 4},
		{3, 5},
	}
	found := false
	for _, want := range validCliques {
		wantCopy := append([]int{}, want...)
		sort.Ints(wantCopy)
		if reflect.DeepEqual(wantCopy, largest) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Größte gefundene Clique (%v) ist keine gültige Kante im Graphen!", largest)
	}
}
