package graphlib

import (
    "testing"
)

// ########################################################################
// read, single source shortest path, all pair, allgemeine Durchlaufzeit 
// helper to read once outside of b.N

func BenchmarkRead(b *testing.B) {
	b.ResetTimer()
	b.Run("Read Random Graph%", func(b *testing.B) {
		ReadGraphAdjList("random_graph_2000_0.5.graph", false)
	})
}

func BenchmarkSSSP(b *testing.B) {
    // lade Graph nur einmal vor dem Benchmark
    g, err := ReadGraphAdjList("random_graph_2000_0.5.graph", false)
    if err != nil {
		b.Fatalf("ReadGraphAdjLis failed: %v", err)
	}

	b.ResetTimer()
	b.Run("Single Source shortest Path%", func(b *testing.B) {
		g.SingleSourceShortestPath(&Vertex{0})
	})
}

func BenchmarkAllPairs(b *testing.B) {
    // lade Graph nur einmal vor dem Benchmark
    g, err := ReadGraphAdjList("random_graph_2000_0.5.graph", false)
    if err != nil {
		b.Fatalf("ReadGraphAdjLis failed: %v", err)
	}

	b.ResetTimer()
	b.Run("All-Pairs%", func(b *testing.B) {
		g.NaiveAllPairsShortestPath()
	})
}


// ########################################################################
// Hier Benchmarken wir die Graphlib allpairShortestPath-Funktion, diameter
/*
func BenchmarkGraphAlgorithms(b *testing.B) {
    const p = 0.25
    sizes := []int{}
    for n := 100; n <= 1000; n += 100 {
        sizes = append(sizes, n)
    }
    for n := 1250; n <= 3000; n += 250 {
        sizes = append(sizes, n)
    }

    for _, n := range sizes {
        g, _, _ := GenerateRandomGraph(n, p, false, false, 64)

        b.Run(
            "n="+string(n)+"/NaiveAllPairs", func(b *testing.B) {
                b.ResetTimer()
                for i := 0; i < b.N; i++ {
                    _ = g.NaiveAllPairsShortestPath()       // replace with your actual naive APSP name
                }
            },
        )
        b.Run(
            "n="+string(n)+"/ImprovedAllPairs", func(b *testing.B) {
                b.ResetTimer()
                for i := 0; i < b.N; i++ {
                    _ = g.ImprovedAllPairs()    // replace with your improved APSP name
                }
            },
        )
        b.Run(
            "n="+string(n)+"/NaiveDiameter", func(b *testing.B) {
                b.ResetTimer()
                for i := 0; i < b.N; i++ {
                    _ = g.NaiveDiameter()       // replace with your naive diameter name
                }
            },
        )
        b.Run(
            "n="+string(n)+"/ImprovedDiameter", func(b *testing.B) {
                b.ResetTimer()
                for i := 0; i < b.N; i++ {
                    _ = g.ImprovedDiameter()    // replace with your improved diameter name
                }
            },
        )
    }
}
*/

// ########################################################################
// benchmark zwischen triangles methoden
