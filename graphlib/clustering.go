package graphlib

import (
	"errors"
	"math"
)

// Dummy metric: just size of cluster (for easy testing)
func clusterSizeMetric(_ *AdjList, C VertexSet) float64 {
	return float64(len(C))
}

// Simple compare: merge if merged cluster is not smaller (just for testing)
func clusterSizeCompare(newMetric, oldMetric float64) bool {
	return newMetric >= oldMetric
}

// partialModularityMetric calculates the modularity q(C) of a community C in the graph.
// It essentially gets an induced subgraph, but for simplicity the special type was avoided.
func partialModularityMetric(l *AdjList, C VertexSet) (float64, error) {
	if l.Size == 0 || len(C) == 0 {
		return 0, errors.New("graph or community is empty")
	}

	var (
		eC        int                   // number of edges inside community C
		degreeSum int                   // sum of degrees of all vertices in C
		m         = l.TotalDegree() / 2 // total number of edges in the graph
	)

	if m == 0 {
		return 0, errors.New("graph has no edges")
	}

	// Compute E(C) and degree sum
	for v := range C {
		vDeg, err := l.Degree(Vertex{v})
		if err != nil {
			return 0, err
		}
		degreeSum += vDeg

		for u := range l.List[v] {
			if _, ok := C[u]; ok {
				eC++
			}
		}
	}
	eC /= 2 // each edge was counted twice

	// partialModularityMetric formula:
	// q(C) = |E(C)| / m - (sum_deg(v) for v in C)^2 / (4 * m^2)
	q := float64(eC)/float64(m) - math.Pow(float64(degreeSum), 2)/(4.0*math.Pow(float64(m), 2))
	return q, nil
}

func partialModularityCompare(a, b float64) bool {
	// Compare two modularity values
	if a < b {
		return false
	}
	return true
}

// partialDisagreementMetric calculates d(C) = (|C| choose 2) - 2 * |E(C)|
// It measures how far a subgraph is from being a complete graph.
// It essentially gets an induced subgraph, but for simplicity the special type was avoided.
func partialDisagreementMetric(l *AdjList, C VertexSet) (int, error) {
	if len(C) < 2 {
		return 0, nil
	}

	// Count number of edges in the induced subgraph
	edgeCount := 0
	for u := range C {
		for v := range l.List[u] {
			if _, ok := C[v]; ok && u < v {
				edgeCount++
			}
		}
	}

	numPairs := (len(C) * (len(C) - 1)) / 2
	return numPairs - 2*edgeCount, nil
}

func PartialDisagreementCompare(a, b float64) bool {
	if a > b {
		return false
	}
	return true
}

// --------------------------------------------------------------------
// Heuristic & Operation 1 - Merge
// --------------------------------------------------------------------

type ClusterIndex [2]int // [bucketKey, sliceIdx]

func (l *AdjList) FindAdjacentClusters(
	k, i int,
	clusters *map[int][]VertexSet,
) []ClusterIndex {
	result := make(map[ClusterIndex]struct{})
	clusterBuckets := *clusters
	currCluster := clusterBuckets[k][i]

	// For each vertex in the current cluster
	for v := range currCluster {
		// For each neighbor of v
		neighbors, err := l.Neighbours(Vertex{v})
		if err != nil {
			continue
		}
		for u := range neighbors {
			// Search all clusters except this one
			for k2, clustersOfSize := range clusterBuckets {
				for i2, otherCluster := range clustersOfSize {
					// skip itself
					if k2 == k && i2 == i {
						continue
					}
					// if u is in otherCluster, add its [bucket,slice] to result
					if _, found := otherCluster[u]; found {
						idx := ClusterIndex{k2, i2}
						result[idx] = struct{}{}
					}
				}
			}
		}
	}

	// convert result map to a slice
	out := make([]ClusterIndex, 0, len(result))
	for idx := range result {
		out = append(out, idx)
	}
	return out
}
