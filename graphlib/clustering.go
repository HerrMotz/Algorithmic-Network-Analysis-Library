package graphlib

import (
	"errors"
	"math"
)

// Dummy metric: just size of cluster (for easy testing)
func clusterSizeMetric(_ *AdjList, C VertexSet) (float64, error) {
	return float64(len(C)), nil
}

// Simple compare: merge if merged cluster is not smaller (just for testing)
func clusterSizeCompare(newMetric, oldMetric float64) (bool, error) {
	return newMetric >= oldMetric, nil
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

	// partialModularityMetric formula (same as in the worksheet, but without the sum over all clusters):
	// q(C) = |E(C)| / m - (sum_deg(v) for v in C)^2 / (4 * m^2)
	q := float64(eC)/float64(m) - math.Pow(float64(degreeSum), 2)/(4.0*math.Pow(float64(m), 2))
	return q, nil
}

// totalModularityMetric calculates the full formula as given in the worksheet
func totalModularityMetric(l *AdjList, C []VertexSet) (float64, error) {
	var totalModularity float64
	for _, cluster := range C {
		q, err := partialModularityMetric(l, cluster)
		if err != nil {
			return 0, err
		}
		totalModularity += q
	}
	return totalModularity, nil
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

// ComponentClustering returns a clustering of the graph where each connected component is its own cluster.
// Each cluster is a set (VertexSet) of vertices.
func (l *AdjList) ComponentClustering() []VertexSet {
	componentIDs := l.naiveConnectedComponents()
	if componentIDs == nil {
		return nil // graph might be invalid or an error occurred
	}

	clusters := make(map[int]VertexSet)

	for vertex, compID := range componentIDs {
		if _, exists := clusters[compID]; !exists {
			clusters[compID] = make(VertexSet)
		}
		clusters[compID][vertex] = struct{}{}
	}

	// Convert map to slice of sets
	result := make([]VertexSet, 0, len(clusters))
	for _, cluster := range clusters {
		result = append(result, cluster)
	}

	return result
}

// BFSWithDepth returns all vertices in an induced subgraph that have a certain max.
// distance to a given start node using BFS
func (iGraph *InducedGraph) BFSWithDepth(start int, depth int) VertexSet {
	visited := make(VertexSet)
	queue := []struct {
		v, d int
	}{{start, 0}}
	visited[start] = struct{}{}
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		if item.d >= depth {
			continue
		}
		for neighbor := range iGraph.AdjList.List[item.v] {
			if _, good := iGraph.GoodVertices[neighbor]; !good {
				continue // skip nodes not in the induced subgraph
			}
			if _, seen := visited[neighbor]; !seen {
				visited[neighbor] = struct{}{}
				queue = append(queue, struct{ v, d int }{neighbor, item.d + 1})
			}
		}
	}
	return visited
}

func (l *AdjList) HalveClustering(metric func(*AdjList, VertexSet) (float64, error), compare func(a, b float64) bool) ([]VertexSet, error) {
	var clustering []VertexSet
	clusteringToDo := l.ComponentClustering() // Initial clusters: each connected component

	for len(clusteringToDo) > 0 {
		// Pop a cluster
		cluster := clusteringToDo[len(clusteringToDo)-1]
		clusteringToDo = clusteringToDo[:len(clusteringToDo)-1]

		// If cluster is too small to split, just keep it
		if len(cluster) <= 1 {
			clustering = append(clustering, cluster)
			continue
		}

		// Build induced subgraph for this cluster
		iGraph := &InducedGraph{
			AdjList:      l,
			GoodVertices: cluster,
		}

		// Get distance matrix
		distMatrix := iGraph.NaiveAllPairsShortestPath()

		// Get slice of vertices in this cluster
		vertexIDs := make([]int, 0, len(cluster))
		vertexIndex := make(map[int]int)
		idx := 0
		for v := range cluster {
			vertexIDs = append(vertexIDs, v)
			vertexIndex[v] = idx
			idx++
		}

		// Find the pair (u, v) with maximal distance d
		maxDist := -1
		var uIdx, vIdx int
		for i := 0; i < len(vertexIDs); i++ {
			for j := 0; j < len(vertexIDs); j++ {
				if distMatrix[i][j] > maxDist {
					maxDist = distMatrix[i][j]
					uIdx, vIdx = i, j
				}
			}
		}

		// If cannot split further, add cluster to clustering
		// For maxDist (=diameter) 1 the heuristic will give us only 1 cluster
		if maxDist <= 1 {
			clustering = append(clustering, cluster)
			continue
		}

		// BFS from u and v (in induced subgraph) for ceil(maxDist/2) steps
		depth := (maxDist + 1) / 2 // Integer division in Go floors by default
		u := vertexIDs[uIdx]
		v := vertexIDs[vIdx]

		cluster1 := iGraph.BFSWithDepth(u, depth)
		cluster2 := iGraph.BFSWithDepth(v, depth)

		// Delete overlapping nodes from cluster2 because they are already in cluster1
		for node := range cluster2 {
			if _, in1 := cluster1[node]; in1 {
				delete(cluster2, node)
			}
		}

		// Remove empty clusters
		// This should not happen since we already processed clusters with diameter <= 1
		if len(cluster1) == 0 || len(cluster2) == 0 {
			clustering = append(clustering, cluster)
			continue
		}

		// Compute metrics for the different clusters
		m, err := metric(l, cluster)
		if err != nil {
			return nil, err
		}
		m1, err1 := metric(l, cluster1)
		if err1 != nil {
			return nil, err1
		}
		m2, err2 := metric(l, cluster2)
		if err2 != nil {
			return nil, err2
		}
		// Compare the metrics
		if compare(m, m1+m2) {
			clustering = append(clustering, cluster) // Add single cluster to clustering because it has a better metric
		} else {
			clusteringToDo = append(clusteringToDo, cluster1)
			clusteringToDo = append(clusteringToDo, cluster2)
		}
	}

	return clustering, nil
}
