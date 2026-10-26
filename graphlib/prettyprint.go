package graphlib

import (
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"
)

// PrintClusters prints the clusters in a human-readable form to standard
// output, biggest clusters first.
func PrintClusters(clusters []VertexSet) {
	FprintClusters(os.Stdout, clusters)
}

// FprintClusters writes the clusters in a human-readable form to w, biggest
// clusters first. The order of the clusters slice itself is not changed.
func FprintClusters(w io.Writer, clusters []VertexSet) {
	if len(clusters) == 0 {
		fmt.Fprintln(w, "(no clusters)")
		return
	}

	// sort biggest clusters first (on a copy, so the caller's order is kept)
	clusters = slices.Clone(clusters)
	sort.SliceStable(clusters, func(i, j int) bool {
		return len(clusters[i]) > len(clusters[j])
	})

	// output
	longestLabel := len(fmt.Sprintf("%d", len(clusters)-1))
	fmt.Fprintln(w, "CLUSTERING")
	fmt.Fprintln(w, strings.Repeat("=", 40))

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
		fmt.Fprintf(w, "%s : %v\n", label, nodes)

		sizeHist[len(c)]++
	}

	// short summary of clusters and stats
	fmt.Fprintln(w, strings.Repeat("-", 40))
	fmt.Fprintf(w, "Total: %d clusters\n", len(clusters))
	fmt.Fprintln(w, "Size distribution:")
	var sizes []int
	for s := range sizeHist {
		sizes = append(sizes, s)
	}
	sort.Ints(sizes)
	for _, s := range sizes {
		fmt.Fprintf(w, "  %2d× cluster of size %d\n", sizeHist[s], s)
	}
	fmt.Fprintln(w, strings.Repeat("=", 40))
}
