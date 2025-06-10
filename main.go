package main

/*func main() {
    const rho = 0.15
    sizes := []int{}
    for n := 100; n <= 1000; n += 100 {
        sizes = append(sizes, n)
    }
    for n := 1250; n <= 3000; n += 250 {
        sizes = append(sizes, n)
    }

    for _, n := range sizes {
        g, _, _ := graphlib.GenerateRandomGraph(n, rho, false, false, 42)

        // Degree
        start := time.Now()
        if _, err := g.DegreeCentralityAll(false); err != nil {
            panic(err)
        }
        fmt.Printf("n=%4d DegreeCentralityAll: %v\n", n, time.Since(start))

        // Closeness
        start = time.Now()
        if _, err := g.ClosenessCentralityAll(true, runtime.NumCPU()); err != nil {
            panic(err)
        }
        fmt.Printf("n=%4d ClosenessCentralityAll: %v\n\n", n, time.Since(start))

        // Power Closeness
        start = time.Now()
        if _, err := g.PowerClosenessCentralityAll(true, 2, runtime.NumCPU()); err != nil {
            panic(err)
        }
        fmt.Printf("n=%4d PowerClosenessCentralityAll: %v\n", n, time.Since(start))

        // Exponential Closeness
        start = time.Now()
        if _, err := g.ExponentialClosenessCentralityAll(true, 1.5, runtime.NumCPU()); err != nil {
            panic(err)
        }
        fmt.Printf("n=%4d ExponentialClosenessCentralityAll: %v\n", n, time.Since(start))

        // Power Iteration
        g.WriteToEdgeList("exampleGraphs/buffer.edgelist", 0)
        m, _ := graphlib.ReadGraphSparseMatrix("exampleGraphs/buffer.edgelist", false)

        powerIterationScores := graphlib.Vector(make([]float64, m.Dim()))
        for i := range powerIterationScores {
            powerIterationScores[i] = 1.0
        }
        start = time.Now()
        if err := m.PowerIteration(&powerIterationScores, 10000); err != nil {
            panic(err)
        }
        fmt.Printf("n=%4d PowerIteration: %v\n", n, time.Since(start))
    }
}
*/
