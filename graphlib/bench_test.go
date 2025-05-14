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

var algos = []struct {
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

func BenchmarkGraphAlgorithmsWithRandomGraphs(b *testing.B) {
	const rho = 0.15

	// prepare the list of graph sizes
	var sizes []int
	for n := 100; n <= 1000; n += 100 {
		sizes = append(sizes, n)
	}
	//for n := 1250; n <= 3000; n += 250 {
	//	sizes = append(sizes, n)
	//}

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
                b.ReportAllocs()
			})
		}
	}
}

func BenchmarkGraphAlgorithmsWithErdos(b *testing.B) {
	b.ReportAllocs()
	// read erdos
	l, err := ReadGraphAdjList("../exampleGraphs/out.pajek-erdos.graph", false)
	if err != nil {
		b.Fatalf("ReadGraphAdjLis failed: %v", err)
	}

	for _, algo := range algos {
		benchName := fmt.Sprintf("%s", algo.name)
		b.Run(benchName, func(b *testing.B) {
			// reset the timer to exclude graph construction
			for i := 0; i < b.N; i++ {
				algo.run(l)
			}
		})
	}

	l, err = ReadGraphAdjList("../exampleGraphs/out.dimacs10-netscie.graph", false)
	if err != nil {
		b.Fatalf("ReadGraphAdjLis failed: %v", err)
	}

	for _, algo := range algos {
		benchName := fmt.Sprintf("%s", algo.name)
		b.Run(benchName, func(b *testing.B) {
			// reset the timer to exclude graph construction
			for i := 0; i < b.N; i++ {
				algo.run(l)
			}
		})
	}
}

// ########################################################################
// benchmark zwischen triangles methoden

// BenchmarkTrianglesOnErdos benchmarks both triangle‐finding methods on the out.pajek-erdos.graph dataset.
func BenchmarkTrianglesOnErdos(b *testing.B) {
    b.ReportAllocs()

    // 1) load the Erdos graph (exclude load time from benchmarks)
    l, err := ReadGraphAdjList("../exampleGraphs/out.pajek-erdos.graph", false)
    if err != nil {
        b.Fatalf("failed to read erdos graph: %v", err)
    }

    // 2) define the two triangle‐finding methods
    methods := []struct {
        name string
        fn   func(*AdjList) ([][3]Vertex, error)
    }{
        {"TrianglesNodeIterator", (*AdjList).TrianglesNodeIterator},
        {"TrianglesCNI",          (*AdjList).TrianglesCNI},
    }

    // 3) run benchmarks
    for _, m := range methods {
        b.Run(m.name, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                tris, err := m.fn(l)
                if err != nil {
                    b.Fatalf("%s returned error: %v", m.name, err)
                }
                // optionally inspect tris to prevent compiler optimizations,
                // e.g. b.Log(len(tris))
                _ = len(tris)
            }
        })
    }
}

// BenchmarkTrianglesOnErdos benchmarks both triangle‐finding methods on the out.pajek-erdos.graph dataset.
func BenchmarkTrianglesOnDimacs(b *testing.B) {
    b.ReportAllocs()

    // 1) load the Erdos graph (exclude load time from benchmarks)
    l, err := ReadGraphAdjList("../exampleGraphs/out.dimacs10-netscie.graph", false)
    if err != nil {
        b.Fatalf("failed to read erdos graph: %v", err)
    }

    // 2) define the two triangle‐finding methods
    methods := []struct {
        name string
        fn   func(*AdjList) ([][3]Vertex, error)
    }{
        {"TrianglesNodeIterator", (*AdjList).TrianglesNodeIterator},
        {"TrianglesCNI",          (*AdjList).TrianglesCNI},
    }

    // 3) run benchmarks
    for _, m := range methods {
        b.Run(m.name, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                tris, err := m.fn(l)
                if err != nil {
                    b.Fatalf("%s returned error: %v", m.name, err)
                }
                // optionally inspect tris to prevent compiler optimizations,
                // e.g. b.Log(len(tris))
                _ = len(tris)
            }
        })
    }
}
