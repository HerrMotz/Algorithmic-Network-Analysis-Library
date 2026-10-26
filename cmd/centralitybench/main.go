package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"

	"github.com/HerrMotz/Algorithmic-Network-Analysis-Library/graphlib"
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
	graphPath := flag.String("graph", "exampleGraphs/jena_road_unfolded_50.metis", "path to the input graph (edge list, METIS or Matrix Market)")
	formatName := flag.String("format", "auto", "input format: auto (by file extension), edgelist, metis or mtx")
	outDir := flag.String("out", ".", "output directory for the *_ranking.csv files")
	flag.Parse()

	format, err := graphlib.ParseGraphFormat(*formatName)
	if err != nil {
		log.Fatal(err)
	}

	// 1) Graph laden
	g, err := graphlib.ReadGraphAdjListAs(*graphPath, format, false)
	if err != nil {
		log.Fatalf("could not load graph: %v", err)
	}
	m := g.ToAdjMatSparse()

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
	err = m.PowerIteration(&powerIterationScores, 1e-9, 10000)
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
	for i := range beta {
		beta[i] = 1.0
	}
	err = m.KatzCentrality(&katzCentralityScores, &alpha, &beta, 10000)
	if err != nil {
		log.Fatalf("KatzCentrality: %v", err)
	}

	// two random start vectors for the Rayleigh variant
	source := rand.New(rand.NewSource(42))
	powerIterationRayleighScores1 := graphlib.Vector(make([]float64, m.Dim()))
	for i := range powerIterationRayleighScores1 {
		powerIterationRayleighScores1[i] = source.Float64()
	}
	powerIterationRayleighScores2 := graphlib.Vector(make([]float64, m.Dim()))
	for i := range powerIterationRayleighScores2 {
		powerIterationRayleighScores2[i] = source.Float64()
	}
	err = m.PowerIterationRayleigh(&powerIterationRayleighScores1, &powerIterationRayleighScores2, 1e-9, 10000)
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
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("mkdir %s: %v", *outDir, err)
	}
	rankings := []struct {
		file    string
		ranking []pair
	}{
		{"degree_ranking.csv", degRank},
		{"closeness_ranking.csv", cloRank},
		{"power_closeness_ranking.csv", powRank},
		{"exponential_closeness_ranking.csv", expRank},
		{"power_iteration_ranking.csv", eigRank},
		{"katz_centrality_ranking.csv", katRank},
		{"power_iteration_rayleigh_ranking.csv", rayRank},
		// !! hier weitere CSVs hinzufügen für mehr Zentralitäten !!
	}
	for _, r := range rankings {
		if err := writeRankingCSV(filepath.Join(*outDir, r.file), r.ranking); err != nil {
			log.Fatalf("write %s: %v", r.file, err)
		}
	}

	// !! die CSVs dann in visuals/S3/<graph>/ packen, siehe visuals/S3/centrality_vis.py !!
}
