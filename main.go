package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"graphlib/graphlib" // <-- an dein go module anpassen
)

// ----------------------------------------------------------------------------
// 1.  Parameter 
// ----------------------------------------------------------------------------
type cliqueInfo struct {
	cliqueCount    int
	maxCliqueSize  int
}

// Liste der zu verarbeitenden Dateien – Pfade ggf. anpassen
var graphFiles = []string{
	"exampleGraphs/soc-karate.mtx",
	"exampleGraphs/soc-dolphins.mtx",
	"exampleGraphs/abi.edgelist",
	"exampleGraphs/out.dimacs10-netscie.graph",
	"exampleGraphs/out.pajek-erdos.graph",
	"exampleGraphs/fb-pages-public-figure.edges",
	"exampleGraphs/fb-pages-media.edges",
}

// ----------------------------------------------------------------------------
// 2.  Mathematische Schranken
// ----------------------------------------------------------------------------
func worstCaseBound(n, d int) float64 {
    return float64(n) * math.Pow(3, float64(d+1)/3)
}

// ----------------------------------------------------------------------------
// 3.  Degeneracy = core-Zahl d(G)
// ----------------------------------------------------------------------------
func degeneracy(adj *graphlib.AdjList) int {
	n := adj.Size
	deg := make([]int, n)
	maxDeg := 0
	for v := 0; v < n; v++ {
		d := len(adj.List[v])
		deg[v] = d
		if d > maxDeg {
			maxDeg = d
		}
	}

	// Buckets für O(n)-Algorithmus
	buckets := make([][]int, maxDeg+1)
	for v, d := range deg {
		buckets[d] = append(buckets[d], v)
	}

	removed := make([]bool, n)
	dValue := 0

	for k := 0; k <= maxDeg; k++ {
		for len(buckets[k]) > 0 {
			v := buckets[k][len(buckets[k])-1]
			buckets[k] = buckets[k][:len(buckets[k])-1]

			if removed[v] || deg[v] != k {
				continue
			}
			if k > dValue {
				dValue = k
			}
			removed[v] = true

			for u := range adj.List[v] {
				if removed[u] {
					continue
				}
				du := deg[u]
				deg[u]--
				buckets[du-1] = append(buckets[du-1], u)
			}
		}
	}
	return dValue
}

// ------------------------------------------------------------
// Clique-Zahl und Max-Größe dynamisch ermitteln
// ------------------------------------------------------------
func cliqueStats(g *graphlib.AdjList) (count, max int) {
    cliques := g.BronKerboschDegeneracy()
    count = len(cliques)
    for _, c := range cliques {
        if len(c) > max {
            max = len(c)
        }
    }
    return
}

// ----------------------------------------------------------------------------
// 4.  Analyse eines Graphen
// ----------------------------------------------------------------------------
type stats struct {
	graphName               string
	nodes, edges            int
	density                 float64
	degen                   int
	maxCliqueSize           int
	cliqueCount             int
    worstCaseBound          float64
}

func analyze(path string) (*stats, error) {
	fmt.Printf("→  Verarbeite %s\n", filepath.Base(path))

	g, err := graphlib.ReadGraphAdjList(path, false) // undirected
	if err != nil {
		return nil, err
	}

	n := g.Size
	m := g.TotalDegree() / 2 // jede Kante in beiden Richtungen gespeichert

	var dens float64
	if n > 1 {
		dens = float64(2*m) / float64(n*(n-1))
	}

	d := degeneracy(g)

	name := filepath.Base(path)

	cliqueCount, maxClique := cliqueStats(g)

    wc := worstCaseBound(n, d)

	return &stats{
		graphName:     name,
		nodes:         n,
		edges:         m,
		density:       dens,
		degen:         d,
		maxCliqueSize: maxClique,
		cliqueCount:   cliqueCount,
		worstCaseBound: wc,
	}, nil
}

func ratio(x, y float64) float64 {
	if y == 0 {
		return 0
	}
	return x / y
}

// ----------------------------------------------------------------------------
// 5.  Hauptprogramm
// ----------------------------------------------------------------------------
func main() {
	var rows [][]string

	for _, file := range graphFiles {
		s, err := analyze(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Fehler bei %s: %v\n", file, err)
			continue
		}
		rows = append(rows, []string{
			s.graphName,
			fmt.Sprintf("%d", s.nodes),
			fmt.Sprintf("%d", s.edges),
			fmt.Sprintf("%.6f", s.density),
			fmt.Sprintf("%d", s.degen),
			fmt.Sprintf("%d", s.maxCliqueSize),
			fmt.Sprintf("%d", s.cliqueCount),
            fmt.Sprintf("%.3e", s.worstCaseBound),
		})
	}

	out, err := os.Create("graph_clique_analysis.csv")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	w := csv.NewWriter(out)
	header := []string{
		"graph_name", "nodes", "edges", "density",
		"degeneracy", "max_clique_size", "clique_count",
		"worst_case_bound",
	}
	if err := w.Write(header); err != nil {
		panic(err)
	}
	if err := w.WriteAll(rows); err != nil {
		panic(err)
	}
	w.Flush()
	fmt.Println("\nAnalyse abgeschlossen – Ergebnisse in graph_clique_analysis.csv.")
}
