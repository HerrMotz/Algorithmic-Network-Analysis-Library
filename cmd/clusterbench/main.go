package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"graphlib/graphlib"
)

// WriteClustersTXT speichert Cluster zeilenweise – kompatibel zu plot_clusters.py
func WriteClustersTXT(path string, clusters []graphlib.VertexSet) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, c := range clusters {
		first := true
		for v := range c {
			if !first {
				fmt.Fprint(w, " ")
			}
			fmt.Fprint(w, v)
			first = false
		}
		fmt.Fprintln(w)
	}
	return w.Flush()
}

func main() {

	// cli flags definieren
	graphPath := flag.String("graph", "", "Pfad zu *.metis oder *.edgelist (AdjList-kompatibel)")
	outDir := flag.String("out", "results", "Ausgabeverzeichnis")
	directed := flag.Bool("directed", false, "Graph als gerichtet einlesen")
	flag.Parse()

	if *graphPath == "" {
		log.Fatal("-graph erforderlich")
	}

	// laden des Graphen
	g, err := graphlib.ReadGraphAdjList(*graphPath, *directed)
	if err != nil {
		log.Fatalf("Graph laden: %v", err)
	}
	fmt.Printf("Graph: |V|=%d  |E|≈%d\n", g.Size, g.TotalDegree()/2)

	// ---- Run-Definition ---------------------------------------------
	type run struct {
		name       string
		algo       func() ([]graphlib.VertexSet, error)
		metricType string // "M" (Modularity) oder "D" (Disagreement)
	}
	add := func(tag string,
		build func(metric func(*graphlib.AdjList, graphlib.VertexSet) (float64, error),
			compare func(float64, float64) bool) ([]graphlib.VertexSet, error)) []run {

		return []run{
			{
				name: tag + "_M",
				metricType: "M",
				algo: func() ([]graphlib.VertexSet, error) {
					return build(graphlib.PartialModularityMetric,
						graphlib.PartialModularityCompare)
				},
			},
			{
				name: tag + "_D",
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
	// 3) Merge-Random
	runs = append(runs, add("mergeRandom", g.MergeClusteringMetricWithRandomPivot)...)
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
			log.Printf("clusters schreiben (%s): %v", r.name, err)
		}
	}
}