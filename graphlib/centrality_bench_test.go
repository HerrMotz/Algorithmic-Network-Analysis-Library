package graphlib

import (
    "fmt"
    "runtime"
    "testing"
)

func BenchmarkCentralityAll(b *testing.B) {
    const rho = 0.15

    // prepare the list of graph sizes
    var sizes []int
    for n := 100; n <= 1000; n += 100 {
        sizes = append(sizes, n)
    }
    for n := 1250; n <= 3000; n += 250 {
        sizes = append(sizes, n)
    }

    for _, n := range sizes {
        // generate once per size (AdjList graph)
        g, _, _ := GenerateRandomGraph(n, rho, false, false, 42)

        // DegreeCentralityAll benchmark
        b.Run(fmt.Sprintf("n=%d/DegreeCentralityAll", n), func(b *testing.B) {
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                if _, err := g.DegreeCentralityAll(false); err != nil {
                    b.Fatal(err)
                }
            }
            b.ReportAllocs()
        })

        // ClosenessCentralityAll benchmark (normalised, max workers)
        b.Run(fmt.Sprintf("n=%d/ClosenessCentralityAll", n), func(b *testing.B) {
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                if _, err := g.ClosenessCentralityAll(true, runtime.NumCPU()); err != nil {
                    b.Fatal(err)
                }
            }
            b.ReportAllocs()
        })
    }
}