package graphlib

import (
	"runtime"
	"sort"
	"testing"
)

// pair ist eine Hilfsstruktur für sortierte (Vertex, Score)-Paare.
type pair struct {
	v   Vertex
	val float64
}

// topK wandelt den Score-Slice in Paare um, sortiert absteigend und gibt die Top-k zurück.
func topK(scores []float64, k int) []pair {
	pairs := make([]pair, len(scores))
	for i, sc := range scores {
		pairs[i] = pair{v: Vertex{i}, val: sc}
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].val > pairs[j].val
	})
	if k > len(pairs) {
		k = len(pairs)
	}
	return pairs[:k]
}

func TestTop5Centralities(t *testing.T) {
	g, err := ReadGraphAdjList("../exampleGraphs/graph_for_test.graph", false)
	if err != nil {
		t.Fatalf("could not load graph: %v", err)
	}

	measures := []struct {
		name    string
		compute func(*AdjList) []float64
	}{
		{
			name: "Degree Centrality (raw)",
			compute: func(gg *AdjList) []float64 {
				scores, err := gg.DegreeCentralityAll(false)
				if err != nil {
					panic(err)
				}
				return scores
			},
		},
		{
			name: "Closeness Centrality (normalised)",
			compute: func(gg *AdjList) []float64 {
				scores, err := gg.ClosenessCentralityAll(true, runtime.NumCPU())
				if err != nil {
					panic(err)
				}
				return scores
			},
		},
		{
			name: "Power Closeness Centrality (normalised)",
			compute: func(gg *AdjList) []float64 {
				scores, err := gg.PowerClosenessCentralityAll(true, 2, runtime.NumCPU())
				if err != nil {
					panic(err)
				}
				return scores
			},
		},
		{
			name: "Exponential Closeness Centrality (normalised)",
			compute: func(gg *AdjList) []float64 {
				scores, err := gg.ExponentialClosenessCentralityAll(true, 1.5, runtime.NumCPU())
				if err != nil {
					panic(err)
				}
				return scores
			},
		},
		// weitere Measures hier ergänzen …
	}

	t.Logf("Graph: %s", "../exampleGraphs/graph_for_test.graph")
	for _, m := range measures {
		scores := m.compute(g)
		top5 := topK(scores, 5)

		t.Logf("— %s —", m.name)
		for i, p := range top5 {
			t.Logf("  #%d: v=%d  centrality=%.6f", i+1, p.v.Id, p.val)
		}
	}
}
