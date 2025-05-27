// centrality_benchmark.go
package main

import (
    "encoding/csv"
    "fmt"
    "log"
    "os"
    "runtime"
    "sort"
    "strconv"

    "graphlib/graphlib"
)

type pair struct {
    v   graphlib.Vertex
    val float64
}

// topKSorted given a list of scores, returns a sorted list of pairs (Vertex, Score).
func topKSorted(scores []float64) []pair {
    n := len(scores)
    pairs := make([]pair, n)
    for i, sc := range scores {
        pairs[i] = pair{v: graphlib.Vertex{i}, val: sc}
    }
    sort.Slice(pairs, func(i, j int) bool {
        return pairs[i].val > pairs[j].val
    })
    return pairs
}

// kendallTau berechnet Kendall's Tau für zwei Ranglisten gleicher Länge.
func kendallTau(rankX, rankY []pair) float64 {
    n := len(rankX)
    posY := make(map[int]int, n)
    for i, p := range rankY {
        posY[p.v.Id] = i
    }

    var concordant, discordant int
    for i := 0; i < n-1; i++ {
        for j := i + 1; j < n; j++ {
            xi, xj := rankX[i].v.Id, rankX[j].v.Id
            if posY[xi] < posY[xj] {
                concordant++
            } else {
                discordant++
            }
        }
    }
    total := n * (n - 1) / 2
    return float64(concordant-discordant) / float64(total)
}

func main() {
    // load graph
    g, err := graphlib.ReadGraphAdjList("exampleGraphs/graph_for_test.graph", false)
    if err != nil {
        log.Fatalf("could not load graph: %v", err)
    }

    // calculate centrality measures
    degScores, err := g.DegreeCentralityAll(false)
    if err != nil {
        log.Fatalf("DegreeCentralityAll: %v", err)
    }
    cloScores, err := g.ClosenessCentralityAll(true, runtime.NumCPU())
    if err != nil {
        log.Fatalf("ClosenessCentralityAll: %v", err)
    }

    // generate rankings
    degRank := topKSorted(degScores)
    cloRank := topKSorted(cloScores)

    f, err := os.Create("centrality_rankings.csv")
    if err != nil {
        log.Fatalf("could not create CSV: %v", err)
    }
    defer f.Close()

    w := csv.NewWriter(f)
    defer w.Flush()

    if err := w.Write([]string{
        "DegreeVertex", "DegreeScore",
        "ClosenessVertex", "ClosenessScore",
    }); err != nil {
        log.Fatalf("could not write header: %v", err)
    }

    n := len(degRank)
    if len(cloRank) < n {
        n = len(cloRank)
    }
    for i := 0; i < n; i++ {
        row := []string{
            strconv.Itoa(degRank[i].v.Id),
            fmt.Sprintf("%.6f", degRank[i].val),
            strconv.Itoa(cloRank[i].v.Id),
            fmt.Sprintf("%.6f", cloRank[i].val),
        }
        if err := w.Write(row); err != nil {
            log.Fatalf("could not write row %d: %v", i+1, err)
        }
    }

    // 5) calc Kendall’s Tau 
    tau := kendallTau(degRank, cloRank)
    fmt.Printf("Kendall's Tau between Degree and Closeness rankings: %.4f\n", tau)
}
