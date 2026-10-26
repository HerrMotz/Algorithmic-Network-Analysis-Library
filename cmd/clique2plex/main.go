package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/HerrMotz/Algorithmic-Network-Analysis-Library/graphlib"
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

func benchmarkCliques(name string, run func() [][]int) {
	start := time.Now()
	cliques := run()
	dur := time.Since(start)
	num := len(cliques)
	largest := largestClique(cliques)

	fmt.Printf("%-32s ⇒ Cliques: %d  |  Größte Clique: %d  |  Zeit: %v\n",
		name, num, len(largest), dur)
}

func benchmarkPlex(name string, run func() (graphlib.VertexSet, error)) {
	start := time.Now()
	plex, err := run()
	dur := time.Since(start)
	if err != nil {
		fmt.Printf("%-32s ⇒ Fehler: %v\n", name, err)
		return
	}
	fmt.Printf("%-32s ⇒ 2-Plex Größe: %d  |  Zeit: %v\n",
		name, plex.Size(), dur)
}

func main() {
	// CLI
	graphPath := flag.String("graph", "", "path to the input graph (edge list, METIS or Matrix Market)")
	formatName := flag.String("format", "auto", "input format: auto (by file extension), edgelist, metis or mtx")
	flag.Parse()
	if *graphPath == "" {
		fmt.Println("Please specify the graph with -graph=...")
		os.Exit(1)
	}
	format, err := graphlib.ParseGraphFormat(*formatName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Einlesen
	fmt.Println("Einlesen des Graphen…")
	g, err := graphlib.ReadGraphAdjListAs(*graphPath, format, false)
	if err != nil {
		fmt.Printf("Fehler beim Einlesen: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Graph eingelesen")

	// ── Bron–Kerbosch Varianten ────────────────────────────────────────────
	fmt.Println("\n===== Bron–Kerbosch Varianten =====")
	bkVariants := []struct {
		name string
		run  func() [][]int
	}{
		{"BronKerboschSimpleIter", func() [][]int { return g.BronKerboschSimpleIter() }},
		{"BronKerboschPivot", func() [][]int { return g.BronKerboschPivot() }},
		{"BronKerboschDegeneracy", func() [][]int { return g.BronKerboschDegeneracy() }},
		{"BronKerboschOwnPivotDegeneracy", func() [][]int { return g.BronKerboschOwnPivotDegeneracy() }},
	}
	for _, v := range bkVariants {
		benchmarkCliques(v.name, v.run)
	}

	// Hilfsfunktion: Initialen PermanentInducedGraph erzeugen
	makePGraph := func() graphlib.PermanentInducedGraph {
		// GoodVertices = alle Knoten
		ig := graphlib.InducedGraph{
			AdjList:      g,
			GoodVertices: make(graphlib.VertexSet),
		}
		for i := 0; i < g.Size; i++ {
			ig.GoodVertices[i] = struct{}{}
		}
		return graphlib.PermanentInducedGraph{
			IGraph:            ig,
			PermanentVertices: make(graphlib.VertexSet),
		}
	}
	/*
		makeIGraph := func() *graphlib.InducedGraph {
			ig := &graphlib.InducedGraph{
				AdjList:      g,
				GoodVertices: make(graphlib.VertexSet),
			}
			for i := 0; i < g.Size; i++ {
				ig.GoodVertices[i] = struct{}{}
			}
			return ig
		}*/

	// ── 2-Plex Varianten ───────────────────────────────────────────────────
	fmt.Println("\n===== 2-Plex Varianten =====")
	plexVariants := []struct {
		name string
		run  func() (graphlib.VertexSet, error)
	}{ /*
			{"TwoPlexNaive", func() (graphlib.VertexSet, error) {
			      pg := makeIGraph()
			      return pg.TwoPlexNaive()
			}},
			{"TwoPlexNaiveMinDegree", func() (graphlib.VertexSet, error) {
				pg := makeIGraph()
				return pg.TwoPlexNaiveMinDegree()
			}},
			{"TwoPlexNaiveParallel", func() (graphlib.VertexSet, error) {
			    pg := makeIGraph()
			    return pg.TwoPlexNaiveParallel()
			}},
			{"TwoPlexPermanent", func() (graphlib.VertexSet, error) {
				pg := makePGraph()
				return pg.TwoPlexPermanent()
			}},
			{"TwoPlexPermanentParallelKaputt", func() (graphlib.VertexSet, error) {
				pg := makePGraph()
				return pg.TwoPlexPermanentParallelKaputt()
			}},
			{"TwoPlexDistance", func() (graphlib.VertexSet, error) {
				pg := makePGraph()
				return pg.TwoPlexDistance()
			}},*/
		{"TwoPlexNStar", func() (graphlib.VertexSet, error) {
			ig := makePGraph().IGraph
			return ig.TwoPlexNStar()
		}},
	}
	for _, v := range plexVariants {
		benchmarkPlex(v.name, v.run)
	}
}
