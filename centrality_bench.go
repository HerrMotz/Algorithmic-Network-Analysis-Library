// file: centrality_benchmark.go
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
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
		pairs[i] = pair{v: graphlib.Vertex{Id: i}, val: sc}
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
			if posY[vi] < posY[vj] {
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
	g, err := graphlib.ReadGraphAdjList("exampleGraphs/jena_road_unfolded_50.metis", false)
    g.WriteToEdgeList("exampleGraphs/jena_road_unfolded_50.edgelist", 0)
	m, err2 := graphlib.ReadGraphSparseMatrix("exampleGraphs/jena_road_unfolded_50.edgelist", false)

	if err != nil {
		log.Fatalf("could not load graph: %v", err)
	}
	if err2 != nil {
		log.Fatalf("could not load graph: %v", err2)
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

    powerClosenessScores, err := g.PowerClosenessCentralityAll(true, 2, runtime.NumCPU())
    if err != nil {
        log.Fatalf("PowerClosenessCentralityAll: %v", err)
    }
    exponentialClosenessScores, err := g.ExponentialClosenessCentralityAll(true, 1.5, runtime.NumCPU())
    if err != nil {
        log.Fatalf("ExponentialClosenessCentralityAll: %v", err)
    }

    // others
	powerIterationScores := graphlib.Vector(make([]float64, m.Dim()))
	for i := range powerIterationScores {
		powerIterationScores[i] = 1.0
	}
	err = m.PowerIteration(&powerIterationScores, 0.001, 10000)
	if err != nil {
		log.Fatalf("PowerIteration: %v", err)
	}

	katzCentralityScores := graphlib.Vector(make([]float64, m.Dim()))
	for i := range katzCentralityScores {
		katzCentralityScores[i] = 1.0
	}
	alpha := graphlib.Vector(make([]float64, m.Dim()))
	for i := range alpha {
		alpha[i] = .02
	}
	beta := graphlib.Vector(make([]float64, m.Dim()))
	for i := range alpha {
		beta[i] = 1.0
	}
	err = m.KatzCentrality(&katzCentralityScores, &alpha, &beta, 10000)
	if err != nil {
		log.Fatalf("KatzCentrality: %v", err)
	}

	powerIterationRayleighScores1 := graphlib.Vector(make([]float64, m.Dim()))
	source := rand.New(rand.NewSource(42))
	for i := range powerIterationRayleighScores1 {
		katzCentralityScores[i] = source.Float64()
	}
	powerIterationRayleighScores2 := graphlib.Vector(make([]float64, m.Dim()))
	for i := range powerIterationRayleighScores2 {
		katzCentralityScores[i] = source.Float64()
	}
	err = m.PowerIterationRayleigh(&powerIterationRayleighScores1, &powerIterationRayleighScores2, 0.2, 10000)
	if err != nil {
		log.Fatalf("PowerIterationRayleigh: %v", err)
	}

	// !! hier weitere hinzufügen für mehr Zentralitäten !!

	// 3) Ranglisten erstellen
	degRank := topKSorted(degreeScores)
	cloRank := topKSorted(closenessScores)
    powRank := topKSorted(powerClosenessScores)
    expRank := topKSorted(exponentialClosenessScores)
	eigRank := topKSorted(powerIterationScores)
	katRank := topKSorted(katzCentralityScores)
	rayRank := topKSorted(powerIterationRayleighScores2)

	// !! hier weitere Ranglisten hinzufügen für mehr Zentralitäten !!

	// 4) CSVs schreiben
	if err := writeRankingCSV("degree_ranking.csv", degRank); err != nil {
		log.Fatalf("write degree CSV: %v", err)
	}
	if err := writeRankingCSV("closeness_ranking.csv", cloRank); err != nil {
		log.Fatalf("write closeness CSV: %v", err)
	}
    if err := writeRankingCSV("power_closeness_ranking.csv", powRank); err != nil {
        log.Fatalf("write power closeness CSV: %v", err)
    }
    if err := writeRankingCSV("exponential_closeness_ranking.csv", expRank); err != nil {
        log.Fatalf("write exponential closeness CSV: %v", err)
    }
    if err := writeRankingCSV("power_iteration_ranking.csv", eigRank); err != nil {
        log.Fatalf("write power iteration CSV: %v", err)
    }
    if err := writeRankingCSV("katz_centrality_ranking.csv", katRank); err != nil {
        log.Fatalf("write katz centrality CSV: %v", err)
    }
    if err := writeRankingCSV("power_iteration_rayleigh_ranking.csv", rayRank); err != nil {
        log.Fatalf("write power iteration rayleigh CSV: %v", err)
    }
	// !! hier weitere CSVs hinzufügen für mehr Zentralitäten !!

	// !! die CSVs dann in visuals/S3/centrality_ranking/ packen und die centrality_vis.py anpassen !!
}
