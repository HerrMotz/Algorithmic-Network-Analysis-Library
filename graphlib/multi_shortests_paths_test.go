package graphlib

import (
    "runtime"
    "testing"
    "time"
)


// TestAPSPImplementationsPerformance tests the performance of different APSP implementations
func TestAPSPImplementationsPerformance(t *testing.T) {
    g, err := ReadGraphAdjList("../exampleGraphs/out.pajek-erdos.graph", false)
    if err != nil {
        t.Fatalf("could not load graph: %v", err)
    }

    workers := runtime.NumCPU()

    start := time.Now()
    _ = g.NaiveAllPairsShortestPath()
    naiveDur := time.Since(start)

    start = time.Now()
    _ = g.MultiThreadedNaiveAllPairsShortestPath(workers)
    mtNaiveDur := time.Since(start)

    start = time.Now()
    _, err = g.ImprovedAllPairsShortestPaths()
    if err != nil {
        t.Fatalf("improved failed: %v", err)
    }
    improvedDur := time.Since(start)

    start = time.Now()
    _, err = g.MultiThreadedImprovedAllPairsShortestPaths(workers)
    if err != nil {
        t.Fatalf("multi improved failed: %v", err)
    }
    mtImprovedDur := time.Since(start)

    t.Logf("Graph: %s", "../exampleGraphs/out.pajek-erdos.graph")
    t.Logf("CPU cores used: %d", workers)
    t.Logf("NaiveAllPairsShortestPath:                        %v", naiveDur)
    t.Logf("MultiThreadedNaiveAllPairsShortestPath:            %v", mtNaiveDur)
    t.Logf("ImprovedAllPairsShortestPaths:                     %v", improvedDur)
    t.Logf("MultiThreadedImprovedAllPairsShortestPaths:        %v", mtImprovedDur)
}
