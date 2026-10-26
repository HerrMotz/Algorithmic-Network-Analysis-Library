package graphlib

import (
	"sort"
	"testing"
)

func largestClique(cliques [][]int) []int {
	var max []int
	for _, c := range cliques {
		if len(c) > len(max) {
			max = c
		}
	}
	return max
}

func setToSortedSlice(vs VertexSet) []int {
	nodes := make([]int, 0, len(vs))
	for v := range vs {
		nodes = append(nodes, v)
	}
	sort.Ints(nodes)
	return nodes
}

func TestCliqueAnd2PlexSummary(t *testing.T) {
	adj := []VertexSet{
		{1: Naught{}, 2: Naught{}},
		{0: Naught{}, 3: Naught{}},
		{0: Naught{}, 3: Naught{}, 4: Naught{}},
		{1: Naught{}, 2: Naught{}, 5: Naught{}},
		{2: Naught{}},
		{3: Naught{}},
		{},
	}
	H := AdjList{List: adj, Size: len(adj)}

	// 1) Maximalen Cliquen mit Bron-Kerbosch (Degeneracy-Variante)
	cliques := H.BronKerboschDegeneracy()
	numCliques := len(cliques)
	largest := largestClique(cliques)
	sort.Ints(largest)

	// 2) Größten 2-Plex mit bestehenden Naive-Implementierung
	induced := InducedGraph{
		AdjList:      &H,
		GoodVertices: make(VertexSet),
	}
	for i := 0; i < H.Size; i++ {
		induced.GoodVertices[i] = struct{}{}
	}
	plex, err := induced.TwoPlexNaive()
	if err != nil {
		t.Fatalf("Fehler bei TwoPlexNaive: %v", err)
	}
	plexNodes := setToSortedSlice(plex)

	// 3) Menschenlesbare Ausgabe ins Test-Log
	t.Logf("Number of maximal cliques: %d", numCliques)
	t.Logf("Largest clique (size %d): %v", len(largest), largest)
	t.Logf("Largest 2-plex (size %d): %v", len(plexNodes), plexNodes)

	// Erwartete maximale Clique-Größe
	expectedCliqueSize := 2
	if len(largest) != expectedCliqueSize {
		t.Errorf("Größte Clique: erwartet Größe %d, erhalten %d (Knoten: %v)", expectedCliqueSize, len(largest), largest)
	}
	// Erwartete 2-Plex Größe
	expected2Plex := []int{0, 1, 2, 3}
	sort.Ints(expected2Plex)
	if len(plexNodes) != len(expected2Plex) {
		t.Errorf("Größter 2-Plex: erwartet Größe %d, erhalten %d (Knoten: %v)", len(expected2Plex), len(plexNodes), plexNodes)
	}
	for i, v := range expected2Plex {
		if plexNodes[i] != v {
			t.Errorf("Größter 2-Plex: erwartet Knoten %v, erhalten %v", expected2Plex, plexNodes)
			break
		}
	}
}
