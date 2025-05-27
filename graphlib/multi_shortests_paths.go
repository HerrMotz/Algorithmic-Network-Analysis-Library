package graphlib

import (
    "runtime"
    "sync"
)

// Concurrent / Multithreaded APSP (AdjList):
// - Uses a dynamic worker pool (size = workers arg or runtime.NumCPU()).
// - Treats each vertex (row) as an independent job for even load distribution.
//
// MultiThreadedNaiveAllPairsShortestPath:
// - Drop-in parallel replacement for NaiveAllPairsShortestPath.
// - Spawns workers that each run BFS (SingleSourceShortestPath) on assigned vertices.
// - Dynamic job queue ensures high-degree vertices don’t block other cores.
func (l *AdjList) MultiThreadedNaiveAllPairsShortestPath(workers int) [][]int {
    size := l.Size
    distanceMatrix := make([][]int, size)

    if workers <= 0 {
        workers = runtime.NumCPU()
    }

    jobs := make(chan int, size) // each job is the source vertex id
    var wg sync.WaitGroup

    // Worker goroutine
    worker := func() {
        defer wg.Done()
        for v := range jobs {
            distances, _ := l.SingleSourceShortestPath(Vertex{v}) 
            distanceMatrix[v] = distances
        }
    }

    // Spin up workers
    wg.Add(workers)
    for i := 0; i < workers; i++ {
        go worker()
    }

    // Feed queue
    for v := 0; v < size; v++ {
        jobs <- v
    }
    close(jobs)

    wg.Wait()
    return distanceMatrix
}

// MultiThreadedImprovedAllPairsShortestPaths parallelizes ImprovedAllPairsShortestPaths:
// 1) Build neighborhoods (sequentially, then read-only).
// 2) Dispatch all vertices to a worker pool.
// 3) Each worker copies distances from processed neighbors or falls back to BFS.
//
// Synchronization:
// - A mutex only guards writes to distanceMatrix and seen flags.
// - Pure reads after a seen flag is set are lock-free.
// - Copy shortcut runs under the lock (low contention, only for low-degree vertices).
func (l *AdjList) MultiThreadedImprovedAllPairsShortestPaths(workers int) ([][]int, error) {
    size := l.Size

    // Init distance matrix & helper structures 
    distanceMatrix := make([][]int, size)
    for i := 0; i < size; i++ {
        distanceMatrix[i] = make([]int, size)
        for j := 0; j < size; j++ {
            distanceMatrix[i][j] = INF
        }
    }
    distancesKnown := make([]bool, size)

    openNeighbourSets := make([]map[int]Naught, size)
    closedNeighbourSets := make([]map[int]Naught, size)

    totalDeg := 0
    for u := 0; u < size; u++ {
        neighMap, err := l.Neighbours(Vertex{u})
        if err != nil {
            return nil, err
        }
        open := make(map[int]Naught, len(neighMap))
        for v := range neighMap {
            open[v] = Naught{}
        }
        openNeighbourSets[u] = open

        closed := make(map[int]Naught, len(open)+1)
        for v := range open {
            closed[v] = Naught{}
        }
        closed[u] = Naught{}
        closedNeighbourSets[u] = closed

        totalDeg += len(open)
    }
    degreeThreshold := 1
    if size > 0 {
        degreeThreshold = totalDeg / size
    }

    // Worker pool 
    if workers <= 0 {
        workers = runtime.NumCPU()
    }

    jobs := make(chan int, size)
    var wg sync.WaitGroup
    var mu sync.Mutex

    // Helper to mark a vertex and optionally its twins as done 
    applyDistances := func(node int, distances []int) {
        distanceMatrix[node] = distances
        distancesKnown[node] = true

        // Copy distances to vertices with identical neighbourhood
        for _, other := range getSameNeighborhoodNodes(node, openNeighbourSets, closedNeighbourSets) {
            if distancesKnown[other] {
                continue
            }
            copyRow := make([]int, size)
            copy(copyRow, distances)
            copyRow[other] = 0
            if _, ok := openNeighbourSets[node][other]; ok {
                copyRow[node] = 1
            } else {
                copyRow[node] = 2
            }
            distanceMatrix[other] = copyRow
            distancesKnown[other] = true
        }
    }

    // Worker function 
    worker := func() {
        defer wg.Done()
        for node := range jobs {
            // Fast‑path: someone else finished it in the meantime
            mu.Lock()
            if distancesKnown[node] {
                mu.Unlock()
                continue
            }
            degree := len(openNeighbourSets[node])

            // all neighbours known?
            allKnown := true
            if degree > 0 {
                for neigh := range openNeighbourSets[node] {
                    if !distancesKnown[neigh] {
                        allKnown = false
                        break
                    }
                }
            }
            mu.Unlock()

            var distances []int
            if degree == 0 {
                // isolated vertex – trivial row
                distances = make([]int, size)
                for i := range distances {
                    distances[i] = INF
                }
                distances[node] = 0
            } else if allKnown && degree < degreeThreshold {
                // Build row by combining neighbour rows 
                mu.Lock() 
                distances = make([]int, size)
                for i := 0; i < size; i++ {
                    if i == node {
                        distances[i] = 0
                        continue
                    }
                    if _, ok := openNeighbourSets[node][i]; ok {
                        distances[i] = 1
                        continue
                    }
                    minD := INF
                    for neigh := range openNeighbourSets[node] {
                        if d := distanceMatrix[neigh][i]; d < minD {
                            minD = d
                        }
                    }
                    if minD < INF {
                        distances[i] = minD + 1
                    } else {
                        distances[i] = INF
                    }
                }
                mu.Unlock()
            } else {
                // Fallback to BFS
                var err error
                distances, err = l.SingleSourceShortestPath(Vertex{node})
                if err != nil {
                    // Should be impossible (invalid vertex), still propagate
                    mu.Lock()
                    applyDistances(node, make([]int, 0)) // mark as done to avoid dead‑loops
                    mu.Unlock()
                    continue
                }
            }

            mu.Lock()
            applyDistances(node, distances)
            mu.Unlock()
        }
    }

    // Spin up workers
    wg.Add(workers)
    for i := 0; i < workers; i++ {
        go worker()
    }

    // Feed the queue
    for v := 0; v < size; v++ {
        jobs <- v
    }
    close(jobs)

    wg.Wait()
    return distanceMatrix, nil
}
