package graphlib

import (
	"fmt"
	"sort"
	"strings"
)

// Prettyprint the clusters in a human-readable form in console output.
func PrintClusters(clusters []VertexSet) {
	if len(clusters) == 0 {
		fmt.Println("(keine Cluster)")
		return
	}

	// sort biggest clusters first
	sort.SliceStable(clusters, func(i, j int) bool {
		return len(clusters[i]) > len(clusters[j])
	})

	// output
	longestLabel := len(fmt.Sprintf("%d", len(clusters)-1))
	fmt.Println("CLUSTERING")
	fmt.Println(strings.Repeat("=", 40))

	sizeHist := make(map[int]int)
	for idx, c := range clusters {
		// sort nodes in cluster
		nodes := make([]int, 0, len(c))
		for v := range c {
			nodes = append(nodes, v)
		}
		sort.Ints(nodes)

		// print
		label := fmt.Sprintf("C%-*d (|C|=%d)", longestLabel, idx, len(c))
		fmt.Printf("%s : %v\n", label, nodes)

		sizeHist[len(c)]++
	}

	// short summary of clusters and stats
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Gesamt: %d Cluster\n", len(clusters))
	fmt.Println("Größenverteilung:")
	var sizes []int
	for s := range sizeHist {
		sizes = append(sizes, s)
	}
	sort.Ints(sizes)
	for _, s := range sizes {
		fmt.Printf("  %2d× Cluster der Größe %d\n", sizeHist[s], s)
	}
	fmt.Println(strings.Repeat("=", 40))
}
