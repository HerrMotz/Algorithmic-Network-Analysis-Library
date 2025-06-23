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

func (l *AdjList) MergeClustering(
	metric func(*AdjList, VertexSet) (float64, error),
	compare func(newMetric, oldMetric float64) bool,
) ([]VertexSet, error) {

	// STEP 1: Initialize each vertex as its own cluster
	clusters := make([]VertexSet, l.Size)
	vertexToCluster := make(map[int]int)
	for i := 0; i < l.Size; i++ {
		clusters[i] = VertexSet{i: Naught{}}
		vertexToCluster[i] = i
	}
	activeClusters := make(map[int]Naught) // this is the same type as a VertexSet, but this was confusing to read.
	for i := 0; i < l.Size; i++ {
		activeClusters[i] = Naught{}
	}

	// STEP 2: Identify isolated vertices
	isolated := make(map[int]Naught)
	for i := 0; i < l.Size; i++ {
		neigh, _ := l.Neighbours(Vertex{i})
		if len(neigh) == 0 {
			isolated[i] = Naught{}
		}
	}
	// STEP 3: Merge all leaves with their single neighbour
	for i := 0; i < l.Size; i++ {
		if _, ok := isolated[i]; ok {
			continue
		}
		neigh, _ := l.Neighbours(Vertex{i})
		if len(neigh) == 1 {
			var n int
			for u := range neigh {
				n = u
			}
			idxV := vertexToCluster[i]
			idxN := vertexToCluster[n]
			_, okV := activeClusters[idxV]
			_, okN := activeClusters[idxN]
			if idxV != idxN && okV && okN {
				// Merge V into N
				for u := range clusters[idxV] {
					clusters[idxN][u] = Naught{}
					vertexToCluster[u] = idxN
				}
				clusters[idxV] = nil
				delete(activeClusters, idxV)
			}
		}
	}

	// STEP 4: Main merging procedure
	n := l.Size
	for k := 1; k <= n; k++ {
		var toMerge [][2]int
		// Only consider active clusters of size k
		var clusterIndices []int
		for idx := range activeClusters {
			if len(clusters[idx]) == k {
				clusterIndices = append(clusterIndices, idx)
			}
		}
		for _, i := range clusterIndices {
			cluster := clusters[i]
			neighborClusters := make(VertexSet)
			for v := range cluster {
				neigh, _ := l.Neighbours(Vertex{v})
				for n := range neigh {
					j := vertexToCluster[n]
					if j != i {
						if _, ok := activeClusters[j]; ok {
							neighborClusters[j] = Naught{}
						}
					}
				}
			}
			// Find best neighbouring cluster to merge
			bestJ := -1
			var bestImprovement float64
			foundImprovement := false
			for j := range neighborClusters {
				other := clusters[j]
				mergedCluster := cluster.Copy()
				for u := range other {
					mergedCluster[u] = Naught{}
				}
				// Calculate metrics
				m1, err := metric(l, cluster)
				if err != nil {
					return nil, err
				}
				m2, err := metric(l, other)
				if err != nil {
					return nil, err
				}
				newMetric, err := metric(l, mergedCluster)
				if err != nil {
					return nil, err
				}
				oldMetric := m1 + m2
				if compare(newMetric, oldMetric) {
					improvement := oldMetric - newMetric
					if !foundImprovement || improvement > bestImprovement {
						bestImprovement = improvement
						bestJ = j
						foundImprovement = true
					}
				}
			}
			if bestJ != -1 {
				toMerge = append(toMerge, [2]int{i, bestJ})
			}
		}
		// Apply merges (each cluster at most once)
		mergedThisRound := make(map[int]Naught)
		for _, pair := range toMerge {
			i, j := pair[0], pair[1]
			_, okI := activeClusters[i]
			_, okJ := activeClusters[j]
			_, mergedI := mergedThisRound[i]
			_, mergedJ := mergedThisRound[j]
			if okI && okJ && !mergedI && !mergedJ {
				for u := range clusters[j] {
					clusters[i][u] = Naught{}
					vertexToCluster[u] = i
				}
				clusters[j] = nil
				delete(activeClusters, j)
				mergedThisRound[i] = Naught{}
				mergedThisRound[j] = Naught{}
			}
		}
	}

	// Return active clusters as VertexSets
	var result []VertexSet
	for idx := range activeClusters {
		if clusters[idx] != nil && len(clusters[idx]) > 0 {
			result = append(result, clusters[idx])
		}
	}
	return result, nil
}
