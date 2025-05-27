package graphlib

import (
	"math"
    "sort"
)

// ---------------------------
//  Degree Centrality
// ---------------------------
// --> for unweighted undirected graphs 

// DegreeCentralityOf 
// returns the (optionally normalised) degree centrality of v.
// Normalised value = deg(v)/(n-1); raw = deg(v).
func (g *AdjList) DegreeCentralityOf(v Vertex, normalised bool) (float64, error) {
    d, err := g.Degree(v)
    if err != nil {
        return 0, err
    }

    if normalised && g.Size > 1 {
        return float64(d) / float64(g.Size-1), nil
    }

    return float64(d), nil
}


// DegreeCentralityAll 
// returns the centrality for every vertex.
// Pass normalised=true/false.
func (g *AdjList) DegreeCentralityAll(normalised bool) ([]float64, error) {
    scores := make([]float64, g.Size)
    for i := 0; i < g.Size; i++ {
        d, err := g.Degree(Vertex{i})
        if err != nil { // should never fail for valid i
            return nil, err
        }
        if normalised && g.Size > 1 {
            scores[i] = float64(d) / float64(g.Size-1)
        } else {
            scores[i] = float64(d)
        }
    }
    return scores, nil
}


// HighestDegreeCentrality 
// returns the vertex with maximum degree centrality;
// ties are resolved by the smaller vertex id.
func (g *AdjList) HighestDegreeCentrality(normalised bool) (Vertex, float64, error) {
    maxV, maxC := Vertex{0}, -1.0
    for i := 0; i < g.Size; i++ {
        c, err := g.DegreeCentralityOf(Vertex{i}, normalised)
        if err != nil {
            return Vertex{}, 0, err
        }
        if c > maxC {
            maxC, maxV = c, Vertex{i}
        }
    }
    return maxV, maxC, nil
}


// TopKDegreeCentrality 
// returns the k vertices with the highest degree centrality
// ordered descending by centrality value.
func (g *AdjList) TopKDegreeCentrality(k int, normalised bool) ([]Vertex, []float64, error) {
    if k <= 0 {
        return nil, nil, nil
    }
    scores, err := g.DegreeCentralityAll(normalised)
    if err != nil {
        return nil, nil, err
    }
    idx := make([]int, g.Size)
    for i := range idx {
        idx[i] = i
    }
    sort.Slice(idx, func(i, j int) bool {
        return scores[idx[i]] > scores[idx[j]]
    })
    if k > len(idx) {
        k = len(idx)
    }
    vertices := make([]Vertex, k)
    values := make([]float64, k)
    for i := 0; i < k; i++ {
        vertices[i] = Vertex{idx[i]}
        values[i] = scores[idx[i]]
    }
    return vertices, values, nil
}


// ---------------------------
//  Closeness Centrality (unweighted graph)
// ---------------------------
// --> for unweighted graphs
// --> C(v) = (n-1) / Σ_u d(v,u)


// ClosenessCentralityOf 
// returns the closeness centrality of a single vertex v
// It performs a single‑source BFS which is optimal for one‑off queries.
func (g *AdjList) ClosenessCentralityOf(v Vertex, normalised bool, workers int) (float64, error) {
    scores, err := g.ClosenessCentralityAll(normalised, workers)
    if err != nil {
        return 0, err
    }
    if v.Id < 0 || v.Id >= len(scores) {
        return 0, nil
    }
    return scores[v.Id], nil
}

// ClosenessCentralityAll 
// returns closeness centrality for every vertex
// Internally it computes APSP once using the multithreaded BFS pool for speed.

func (g *AdjList) ClosenessCentralityAll(normalised bool, workers int) ([]float64, error)  {
    n := g.Size     
    distMatrix := g.MultiThreadedNaiveAllPairsShortestPath(workers)

    res := make([]float64, n)
    for i := 0; i < n; i++ {
        var sum float64
        reachable := 0

        for j := 0; j < n; j++ {
            if i == j {
                continue
            }
            d := distMatrix[i][j]
            if d < math.MaxInt32 { // erreichbar
                sum += float64(d)
                reachable++
            }
        }

        if sum == 0 {
            res[i] = 0
            continue
        }

        if normalised {
            // (Anzahl erreichte Knoten - 1) / Summe der Distanzen
            res[i] = float64(reachable) / sum
        } else {
            // unnormalisiert: 1 / Summe der Distanzen
            res[i] = 1.0 / sum
        }
    }
    return res, nil
}

// HighestClosenessCentrality 
// returns vertex with maximum closeness
func (g *AdjList) HighestClosenessCentrality(normalised bool, workers int) (Vertex, float64, error) {
    scores, err := g.ClosenessCentralityAll(normalised, workers)
    if err != nil {
        return Vertex{}, 0, err
    }
    maxV, maxC := Vertex{0}, -1.0
    for i, c := range scores {
        if c > maxC {
            maxC, maxV = c, Vertex{i} 
        }
    }
    return maxV, maxC, nil
}

// TopKClosenessCentrality 
// returns the k vertices with highest closeness scores
func (g *AdjList) TopKClosenessCentrality(k int, normalised bool, workers int) ([]Vertex, []float64, error) {
    if k <= 0 {
        return nil, nil, nil
    }
    scores, err := g.ClosenessCentralityAll(normalised, workers)
    if err != nil {
        return nil, nil, err
    }
    idx := make([]int, len(scores))
    for i := range idx {
        idx[i] = i
    }
    sort.Slice(idx, func(i, j int) bool {
        return scores[idx[i]] > scores[idx[j]]
    })
    if k > len(idx) {
        k = len(idx)
    }
    vertices := make([]Vertex, k)
    values := make([]float64, k)
    for i := 0; i < k; i++ {
        vertices[i] = Vertex{idx[i]}
        values[i] = scores[idx[i]]
    }
    return vertices, values, nil
}