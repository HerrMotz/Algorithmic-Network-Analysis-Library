// file: centrality_benchmark.go
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

// pair verbindet Vertex und Score
type pair struct {
    v   graphlib.Vertex
    val float64
}

// topKSorted liefert die komplette Rangliste (best→worst)
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

// writeRankingCSV schreibt eine CSV mit Spalten: Rank,Vertex,Score
func writeRankingCSV(filename string, ranking []pair) error {
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()

    w := csv.NewWriter(f)
    defer w.Flush()

    // Header
    if err := w.Write([]string{"Rank", "Vertex", "Score"}); err != nil {
        return err
    }
    // Datenzeilen
    for i, p := range ranking {
        rec := []string{
            strconv.Itoa(i + 1),
            strconv.Itoa(p.v.Id),
            fmt.Sprintf("%.6f", p.val),
        }
        if err := w.Write(rec); err != nil {
            return err
        }
    }
    return nil
}

// kendallTau berechnet Kendall's Tau für zwei Ranglisten gleicher Länge.
// rankX und rankY geben für jede Position die Vertex-IDs.
func kendallTau(rankX, rankY []pair) float64 {
    n := len(rankX)
    // position maps
    posX := make(map[int]int, n)
    posY := make(map[int]int, n)
    for i, p := range rankX {
        posX[p.v.Id] = i
    }
    for i, p := range rankY {
        posY[p.v.Id] = i
    }

    var concordant, discordant int
    for i := 0; i < n-1; i++ {
        for j := i + 1; j < n; j++ {
            vi, vj := rankX[i].v.Id, rankX[j].v.Id
            // in X: vi vor vj; in Y prüfen
            if (posY[vi] < posY[vj]) {
                concordant++
            } else {
                discordant++
            }
        }
    }
    totalPairs := n * (n - 1) / 2
    return float64(concordant-discordant) / float64(totalPairs)
}

func main() {

    // 1) Graph laden
    g, err := graphlib.ReadGraphAdjList("exampleGraphs/out.pajek-erdos.graph", false)
    if err != nil {
        log.Fatalf("could not load graph: %v", err)
    }

    // 2) Zentralitäten berechnen
    degreeScores, err := g.DegreeCentralityAll(false)
    if err != nil {
        log.Fatalf("DegreeCentralityAll: %v", err)
    }
    closenessScores, err := g.ClosenessCentralityAll(true, runtime.NumCPU())
    if err != nil {
        log.Fatalf("ClosenessCentralityAll: %v", err)
    }

    // !! hier weitere hinzufügen für mehr Zentralitäten !!


    // 3) Ranglisten erstellen
    degRank := topKSorted(degreeScores)
    cloRank := topKSorted(closenessScores)

    // !! hier weitere Ranglisten hinzufügen für mehr Zentralitäten !!

    // 4) CSVs schreiben
    if err := writeRankingCSV("degree_ranking.csv", degRank); err != nil {
        log.Fatalf("write degree CSV: %v", err)
    }
    if err := writeRankingCSV("closeness_ranking.csv", cloRank); err != nil {
        log.Fatalf("write closeness CSV: %v", err)
    }
    // !! hier weitere CSVs hinzufügen für mehr Zentralitäten !!

    // !! die CSVs dann in visuals/S3/centrality_ranking/ packen und die centrality_vis.py anpassen !! 
}
