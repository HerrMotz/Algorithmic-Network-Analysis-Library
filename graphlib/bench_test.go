package graphlib

import (
	"fmt"
	"testing"
	"time"
)

// ########################################################################
// read, single source shortest path, all pair, allgemeine Durchlaufzeit
// helper to read once outside of b.N

func BenchmarkRead(b *testing.B) {
	b.ResetTimer()
	start := time.Now()

	b.Run("Read Random Graph", func(b *testing.B) {
		ReadGraphAdjList("random_graph_2000_0_5.mgraph", false)
	})

	b.Logf("Finished %d runs in %s", b.N, time.Since(start))

	b.ReportAllocs()

}

func BenchmarkSSSP(b *testing.B) {
	// lade Graph nur einmal vor dem Benchmark
	g, err := ReadGraphAdjList("random_graph_2000_0_5.mgraph", false)
	if err != nil {
		b.Fatalf("ReadGraphAdjLis failed: %v", err)
	}

	b.ResetTimer()

	start := time.Now()
	b.Run("Single Source shortest Path%", func(b *testing.B) {
		g.SingleSourceShortestPath(Vertex{0})
	})

	b.Logf("Finished %d runs in %s", b.N, time.Since(start))
	b.ReportAllocs()
}

func BenchmarkAllPairs(b *testing.B) {
	// lade Graph nur einmal vor dem Benchmark
	g, err := ReadGraphAdjList("random_graph_2000_0_5.mgraph", false)
	if err != nil {
		b.Fatalf("ReadGraphAdjLis failed: %v", err)
	}

	b.ResetTimer()

	start := time.Now()
	b.Run("All-Pairs%", func(b *testing.B) {
		g.NaiveAllPairsShortestPath()
	})

	b.Logf("Finished %d runs in %s", b.N, time.Since(start))
	b.ReportAllocs()
}

// ########################################################################
// Hier Benchmarken wir die Graphlib allpairShortestPath-Funktion, diameter

func BenchmarkGraphAlgorithms(b *testing.B) {
	const rho = 0.15

	// prepare the list of graph sizes
	var sizes []int
	for n := 100; n <= 1000; n += 100 {
		sizes = append(sizes, n)
	}
	/*for n := 1250; n <= 3000; n += 250 {
		sizes = append(sizes, n)
	}*/

	// define each algorithm as a named function
	algos := []struct {
		name string
		run  func(g *AdjList)
	}{
		{
			name: "NaiveAllPairs",
			run: func(g *AdjList) {
				_ = g.NaiveAllPairsShortestPath()
			},
		},
		{
			name: "ImprovedAllPairs",
			run: func(g *AdjList) {
				_, _ = g.ImprovedAllPairsShortestPaths()
			},
		},
		{
			name: "NaiveDiameter",
			run: func(g *AdjList) {
				_ = g.NaiveDiameter()
			},
		},
		{
			name: "ImprovedDiameter",
			run: func(g *AdjList) {
				_ = g.ImprovedDiameter()
			},
		},
	}

	// for each size and each algorithm, run a sub‐benchmark
	for _, n := range sizes {
		// generate once per size
		g, _, _ := GenerateRandomGraph(n, rho, false, false, 42)

		for _, algo := range algos {
			benchName := fmt.Sprintf("n=%d/%s", n, algo.name)
			b.Run(benchName, func(b *testing.B) {
				// reset the timer to exclude graph construction
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					algo.run(g)
				}
			})
		}
	}
}

// ########################################################################
// benchmark zwischen triangles methoden
