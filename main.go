package main

import (
    "fmt"
    "log"

    "graphlib/graphlib"
)

func main() {
    // Define the two input graph files
    graphs := []struct {
        name string
        path string
    }{
        {"DIMACS10 netscie", "exampleGraphs/out.dimacs10-netscie.graph"},
        {"Pajek Erdos", "exampleGraphs/out.pajek-erdos.graph"},
    }

    for _, g := range graphs {
        // 1) Read the graph
        l, err := graphlib.ReadGraphAdjList(g.path, false)
        if err != nil {
            log.Fatalf("failed to read %s: %v", g.name, err)
        }

        // 2) Compute metrics
        density, err := l.Density()
        if err != nil {
            log.Fatalf("Density error for %s: %v", g.name, err)
        }

        maxDeg, err := l.MaxDegree()
        if err != nil {
            log.Fatalf("MaxDegree error for %s: %v", g.name, err)
        }

        hIndex, err := l.HIndex()
        if err != nil {
            log.Fatalf("HIndex error for %s: %v", g.name, err)
        }

        degeneracy, err := l.Degeneracy()
        if err != nil {
            log.Fatalf("Degeneracy error for %s: %v", g.name, err)
        }

        clustering, err := l.GlobalClustering()
        if err != nil {
            log.Fatalf("GlobalClustering error for %s: %v", g.name, err)
        }

        // 3) Print results
        fmt.Printf("=== Metrics for %s ===\n", g.name)
        fmt.Printf("Density:             %.6f\n", density)
        fmt.Printf("Maximum Degree:      %d\n", maxDeg)
        fmt.Printf("H-Index:             %d\n", hIndex)
        fmt.Printf("Degeneracy:          %d\n", degeneracy)
        fmt.Printf("Global Clustering:   %.6f\n\n", clustering)
    }
}
