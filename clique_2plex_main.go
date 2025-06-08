package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"graphlib/graphlib"
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

func setToSortedSlice(vs graphlib.VertexSet) []int {
	nodes := make([]int, 0, len(vs))
	for v := range vs {
		nodes = append(nodes, v)
	}
	sort.Ints(nodes)
	return nodes
}

func main() {
	graphPath := flag.String("graph", "", "Pfad zur Graph-Datei (Adjazenzlistenformat)")
	flag.Parse()

	if *graphPath == "" {
		fmt.Println("Bitte gib den Pfad zum Graphen an mit -graph=...")
		os.Exit(1)
	}

	// 1) Graph laden 
	g, err := graphlib.ReadGraphAdjList(*graphPath, false)
	if err != nil {
		fmt.Printf("Fehler beim Einlesen: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("graph eingelesen")

	// 2) Maximalen Cliquen finden
	cliques := g.BronKerboschDegeneracy()
	numCliques := len(cliques)
	largest := largestClique(cliques)
	sort.Ints(largest)

	// 3) Größten 2-Plex finden
	induced := graphlib.InducedGraph{
		AdjList:     g,
		GoodVertices: make(graphlib.VertexSet),
	}
	for i := 0; i < g.Size; i++ {
		induced.GoodVertices[i] = struct{}{}
	}
	plex, err := induced.TwoPlexNaive()
	if err != nil {
		fmt.Printf("Fehler bei TwoPlexNaive: %v\n", err)
		os.Exit(1)
	}
	plexNodes := setToSortedSlice(plex)

	// 5) "Menschenlesbare" Ausgabe
	fmt.Printf("==== Auswertung für %s ====\n", *graphPath)
	fmt.Printf("Anzahl (inclusion-)maximaler Cliquen: %d\n", numCliques)
	fmt.Printf("Größte Clique (Größe %d): %v\n", len(largest), largest)
	fmt.Printf("Größter 2-Plex (Größe %d): %v\n", len(plexNodes), plexNodes)
}
