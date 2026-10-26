package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"maps"
	"math/rand"
	"os"
	"slices"
	"time"

	"github.com/HerrMotz/Algorithmic-Network-Analysis-Library/graphlib"
)

type metricFunc = func(*graphlib.AdjList, graphlib.VertexSet) (float64, error)
type compareFunc = func(float64, float64) bool

// WriteClustersTXT speichert Cluster zeilenweise – kompatibel zu plot_clusters.py.
// Die Knoten jeder Zeile sind aufsteigend sortiert, damit Läufe vergleichbar bleiben.
func WriteClustersTXT(path string, clusters []graphlib.VertexSet) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, c := range clusters {
		for i, v := range slices.Sorted(maps.Keys(c)) {
			if i > 0 {
				fmt.Fprint(w, " ")
			}
			fmt.Fprint(w, v)
		}
		fmt.Fprintln(w)
	}
	return w.Flush()
}

func main() {

	// cli flags definieren
	graphPath := flag.String("graph", "", "path to the input graph (edge list, METIS or Matrix Market)")
	outDir := flag.String("out", "results", "output directory for the *.clusters.txt files")
	directed := flag.Bool("directed", false, "read the graph as directed")
	seed := flag.Int64("seed", 1, "seed for the random pivot order of the mergeRandom heuristic")
	flag.Parse()

	if *graphPath == "" {
		log.Fatal("-graph is required")
	}

	// laden des Graphen
	g, err := graphlib.ReadGraphAdjList(*graphPath, *directed)
	if err != nil {
		log.Fatalf("loading graph: %v", err)
	}
	fmt.Printf("Graph: |V|=%d  |E|≈%d\n", g.Size, g.TotalDegree()/2)

	// ---- Run-Definition ---------------------------------------------
	type run struct {
		name       string
		algo       func() ([]graphlib.VertexSet, error)
		metricType string // "M" (Modularity) oder "D" (Disagreement)
	}
	add := func(tag string, build func(metric metricFunc, compare compareFunc) ([]graphlib.VertexSet, error)) []run {

		return []run{
			{
				name:       tag + "_M",
				metricType: "M",
				algo: func() ([]graphlib.VertexSet, error) {
					return build(graphlib.PartialModularityMetric,
						graphlib.PartialModularityCompare)
				},
			},
			{
				name:       tag + "_D",
				metricType: "D",
				algo: func() ([]graphlib.VertexSet, error) {
					conv := func(a *graphlib.AdjList, C graphlib.VertexSet) (float64, error) {
						d, err := graphlib.PartialDisagreementMetric(a, C)
						return float64(d), err
					}
					return build(conv, graphlib.PartialDisagreementCompare)
				},
			},
		}
	}

	runs := []run{}
	// 1) Merge-BFS
	runs = append(runs, add("mergeBFS", g.MergeClusteringGreedyBFS)...)
	// 2) Halve
	runs = append(runs, add("halve", g.HalveClustering)...)
	// 3) Merge-Random: every run gets a fresh generator with the same seed, so
	// the M and D results are reproducible independently of each other
	runs = append(runs, add("mergeRandom", func(metric metricFunc, compare compareFunc) ([]graphlib.VertexSet, error) {
		return g.MergeClusteringMetricWithRandomPivot(rand.New(rand.NewSource(*seed)), metric, compare)
	})...)
	// 4) ICE
	runs = append(runs, add("ice", g.ICEClustering)...)

	// Directory für Ergebnisse erstellen
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("mkdir %s: %v", *outDir, err)
	}

	// Durchführen und Ergebnisse ausgeben
	for _, r := range runs {
		start := time.Now()
		C, err := r.algo()
		if err != nil {
			log.Printf("%s ERROR: %v\n", r.name, err)
			continue
		}
		elapsed := time.Since(start)

		var score float64
		if r.metricType == "M" {
			score, _ = graphlib.TotalModularityMetric(g, C)
		} else {
			score, _ = graphlib.TotalDisagreementMetric(g, C)
		}

		fmt.Printf("%-28s  k=%-4d  Q=%.6f  t=%v\n",
			r.name, len(C), score, elapsed)

		outFile := fmt.Sprintf("%s/%s.clusters.txt", *outDir, r.name)
		if err := WriteClustersTXT(outFile, C); err != nil {
			log.Printf("writing clusters (%s): %v", r.name, err)
		}
	}
}
