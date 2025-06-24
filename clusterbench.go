package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

	ext := filepath.Ext(*graphPath)
	if ext != ".metis" && ext != ".edgelist" && ext != ".txt" {
		log.Fatalf("unsupported format %s – die Cluster-Heuristiken brauchen eine AdjList", ext)
	}

	// laden des Graphen
	g, err := graphlib.ReadGraphAdjList(*graphPath, *directed)
	if err != nil {
		log.Fatalf("Graph laden: %v", err)
	}
	fmt.Printf("Graph: |V|=%d  |E|≈%d\n", g.Size, g.TotalDegree()/2)

	// Heuristiken für Clustering definieren
	type run struct {
		name string
		algo func() ([]graphlib.VertexSet, error)
	}

	runs := []run{
		{
			name: "mergeBFS_modularity",
			algo: func() ([]graphlib.VertexSet, error) {
				return g.MergeClusteringGreedyBFS(
					graphlib.PartialModularityMetric,
					graphlib.PartialModularityCompare,
				)
			},
		},
		{
			name: "halve_diameter_modularity",
			algo: func() ([]graphlib.VertexSet, error) {
				return g.HalveClustering(
					graphlib.PartialModularityMetric,
					graphlib.PartialModularityCompare,
				)
			},
		},
		{
			name: "mergeRandom_disagreement",
			algo: func() ([]graphlib.VertexSet, error) {
				return g.MergeClusteringMetricWithRandomPivot(
					func(a *graphlib.AdjList, C graphlib.VertexSet) (float64, error) {
						d, err := graphlib.PartialDisagreementMetric(a, C)
						return float64(d), err
					},
					graphlib.PartialDisagreementCompare,
				)
			},
		},
	}

	// Directory für Ergebnisse erstellen
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("mkdir %s: %v", *outDir, err)
	}

	// Durchführen und Ergebnisse ausgeben
	for _, r := range runs {
		start := time.Now()
		cl, err := r.algo()
		if err != nil {
			log.Printf("%s ERROR: %v\n", r.name, err)
			continue
		}
		elapsed := time.Since(start)

		mod, _ := graphlib.TotalModularityMetric(g, cl)
		fmt.Printf("%-28s  k=%-4d  Q=%.6f  t=%v\n",
			r.name, len(cl), mod, elapsed)

		outFile := fmt.Sprintf("%s/%s.clusters.txt", *outDir, r.name)
		if err := WriteClustersTXT(outFile, cl); err != nil {
			log.Printf("clusters schreiben (%s): %v", r.name, err)
		}
	}
}