package graphlib

import (
	"errors"
	"gonum.org/v1/gonum/stat/combin"
	"math"
	"math/rand"
	"sort"
)

// Dummy metric: just size of cluster (for easy testing)
func clusterSizeMetric(_ *AdjList, C VertexSet) (float64, error) {
	return float64(len(C)), nil
}

// Simple compare: merge if merged cluster is not smaller (just for testing)
func clusterSizeCompare(newMetric, oldMetric float64) (bool, error) {
	return newMetric >= oldMetric, nil
}

// PartialModularityMetric calculates the modularity q_hat(C) of a community C in the graph.
// It essentially gets an induced subgraph, but for simplicity the special type was avoided.
func PartialModularityMetric(l *AdjList, C VertexSet) (float64, error) {
	if l.Size == 0 || len(C) == 0 {
		return 0, errors.New("partialModularityMetric: cluster is empty")
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
	// q_hat(C) = |E(C)| / m - (sum_deg(v) for v in C)^2 / (4 * m^2)
	q := float64(eC)/float64(m) - math.Pow(float64(degreeSum), 2)/(4.0*math.Pow(float64(m), 2))
	return q, nil
}

// TotalModularityMetric calculates the full formula as given in the worksheet
func TotalModularityMetric(l *AdjList, C []VertexSet) (float64, error) {
	var totalModularity float64
	for _, cluster := range C {
		q, err := PartialModularityMetric(l, cluster)
		if err != nil {
			return 0, err
		}
		totalModularity += q
	}
	return totalModularity, nil
}

func PartialModularityCompare(a, b float64) bool {
	// Compare two modularity values
	if a < b {
		return false
	}
	return true
}

// PartialDisagreementMetric calculates d_hat(C) = (|C| choose 2) - 2 * |E(C)|
// It measures how far a subgraph is from being a complete graph.
// It essentially gets an induced subgraph, but for simplicity the special type was avoided.
func PartialDisagreementMetric(l *AdjList, C VertexSet) (int, error) {
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

	return combin.Binomial(len(C), 2) - 2*edgeCount, nil
}

func PartialDisagreementCompare(a, b float64) bool {
	if a > b {
		return false
	}
	return true
}

// TotalDisagreementMetric calculates the full formula d(C) as given in the worksheet
func TotalDisagreementMetric(l *AdjList, C []VertexSet) (int, error) {
	sum := 0
	for _, cluster := range C {
		q, err := PartialDisagreementMetric(l, cluster)
		if err != nil {
			return 0, err
		}
		sum += q
	}

	// calculate total edge count in the graph
	m := l.TotalDegree()
	return (m / 2) + sum, nil
}

// --------------------------------------------------------------------
// Heuristic & Operation 1 - Merge
// --------------------------------------------------------------------

func (l *AdjList) MergeClusteringGreedyBFS(
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

// MergeClusteringMetricWithRandomPivot merges clusters by picking a random active cluster
// and merging it with the best possible neighboring cluster according to the metric.
func (l *AdjList) MergeClusteringMetricWithRandomPivot(
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
	activeClusters := make(map[int]Naught)
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
	// STEP 3: Merge all leaves with their single neighbor
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

	// STEP 4: Main merging procedure using metric
	for {
		// Gather all current active cluster indices
		var clusterIndices []int
		for idx := range activeClusters {
			clusterIndices = append(clusterIndices, idx)
		}
		if len(clusterIndices) <= 1 {
			break // Only one cluster left, or none
		}

		// Try clusters in random order to avoid bias
		rand.Shuffle(len(clusterIndices), func(i, j int) {
			clusterIndices[i], clusterIndices[j] = clusterIndices[j], clusterIndices[i]
		})
		merged := false

		for _, chosenIdx := range clusterIndices {
			cluster := clusters[chosenIdx]

			// Find neighbor clusters
			neighborClusters := make(VertexSet)
			for v := range cluster {
				neigh, _ := l.Neighbours(Vertex{v})
				for n := range neigh {
					j := vertexToCluster[n]
					if j != chosenIdx {
						if _, ok := activeClusters[j]; ok {
							neighborClusters[j] = Naught{}
						}
					}
				}
			}

			// Gather all neighbor indices
			var neighborIndices []int
			for j := range neighborClusters {
				neighborIndices = append(neighborIndices, j)
			}
			if len(neighborIndices) == 0 {
				continue // No neighbors to merge with, try next cluster
			}

			// Find the best neighbor according to the metric
			bestNeighbor := -1
			bestImprovement := 0.0
			foundImprovement := false
			m1, err := metric(l, cluster)
			if err != nil {
				return nil, err
			}

			for _, j := range neighborIndices {
				other := clusters[j]

				mergedCluster := cluster.Copy()
				for u := range other {
					mergedCluster[u] = Naught{}
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
						bestNeighbor = j
						foundImprovement = true
					}
				}
			}

			if foundImprovement && bestNeighbor != -1 {
				// Merge clusters: chosenIdx <- bestNeighbor
				for u := range clusters[bestNeighbor] {
					clusters[chosenIdx][u] = Naught{}
					vertexToCluster[u] = chosenIdx
				}
				clusters[bestNeighbor] = nil
				delete(activeClusters, bestNeighbor)
				merged = true
				break // Merge only one pair per iteration
			}
			// else, try next cluster
		}

		if !merged {
			// No possible merges remain
			break
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
// With our updated approach this function is not used anymore
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
	distMatrix, err := l.ImprovedAllPairsShortestPaths()
	if err != nil {
		return nil, err
	}

	for len(clusteringToDo) > 0 {
		// Pop a cluster
		cluster := clusteringToDo[len(clusteringToDo)-1]
		clusteringToDo = clusteringToDo[:len(clusteringToDo)-1]

		// If cluster is too small to split, just keep it
		if len(cluster) <= 1 {
			clustering = append(clustering, cluster.Copy())
			continue
		}

		// Build induced subgraph for this cluster
		iGraph := &InducedGraph{
			AdjList:      l,
			GoodVertices: cluster,
		}

		// Find the pair (u, v) with maximal distance d in the current cluster using the precomputed distance matrix
		maxDist := -1
		var u, v int
		for x := range cluster {
			for y := range cluster {
				if x == y {
					continue
				}
				if distMatrix[x][y] > maxDist {
					maxDist = distMatrix[x][y]
					u, v = x, y
				}
			}
		}

		// If cannot split further, add cluster to clustering
		// For maxDist (=diameter) 1 the heuristic will give us only 1 cluster
		if maxDist <= 1 {
			clustering = append(clustering, cluster.Copy())
			continue
		}

		cluster1 := make(VertexSet)
		cluster2 := make(VertexSet)

		for w := range iGraph.GoodVertices {
			// Compare distances from u and v
			if distMatrix[u][w] <= distMatrix[v][w] {
				cluster1[w] = struct{}{}
			} else {
				cluster2[w] = struct{}{}
			}
		}

		// Remove empty clusters
		// This should not happen since we already processed clusters with diameter <= 1
		if len(cluster1) == 0 || len(cluster2) == 0 {
			clustering = append(clustering, cluster.Copy())
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
			clustering = append(clustering, cluster.Copy()) // Add single cluster to clustering because it has a better metric
		} else {
			clusteringToDo = append(clusteringToDo, cluster1)
			clusteringToDo = append(clusteringToDo, cluster2)
		}
	}

	return clustering, nil
}

func (l *AdjList) ICEClustering(
	metric func(*AdjList, VertexSet) (float64, error),
	compare func(a, b float64) bool,
) ([]VertexSet, error) {
	n := l.Size
	assigned := make(map[int]bool)
	var clusters []VertexSet

	// 1. Sort vertices by degree ascending (start with the smallest degree)
	verticesSortedByDegree := make([]int, n)
	for k := range l.List {
		verticesSortedByDegree[k] = k
	}
	sort.Slice(verticesSortedByDegree, func(i, j int) bool {
		n1, _ := l.Neighbours(Vertex{verticesSortedByDegree[i]})
		n2, _ := l.Neighbours(Vertex{verticesSortedByDegree[j]})
		return len(n1) < len(n2)
	})

	// 2. Initial clustering: group each vertex and its neighbors if not already clustered
	for _, v := range verticesSortedByDegree {
		// this also puts isolated vertices into their own cluster
		if assigned[v] {
			continue
		}
		cluster := make(VertexSet)
		cluster[v] = Naught{}
		neigh, err := l.Neighbours(Vertex{v})
		if err != nil {
			return nil, err
		}
		for u := range neigh {
			cluster[u] = Naught{}
		}
		// Skip if all already assigned
		skip := true
		for u := range cluster {
			if !assigned[u] {
				skip = false
				break
			}
		}
		if skip {
			continue
		}
		for u := range cluster {
			assigned[u] = true
		}
		clusters = append(clusters, cluster)
	}

	// For marking finished clusters
	done := make(map[int]bool)

	for {
		// Find the largest unfinished cluster
		maxIdx, maxSize := -1, -1
		for i, c := range clusters {
			if done[i] {
				continue
			}
			if len(c) > maxSize {
				maxSize = len(c)
				maxIdx = i
			}
		}
		if maxIdx == -1 {
			break // all done
		}

		// Pick candidate vertex to move from the biggest cluster
		c := clusters[maxIdx]
		bestV, maxOutside, minDegree := -1, -1, INF
		for v := range c {
			outside := 0
			neigh, err := l.Neighbours(Vertex{v})
			if err != nil {
				return nil, err
			}
			for u := range neigh {
				inSame := false
				for w := range c {
					if u == w {
						inSame = true
						break
					}
				}
				if !inSame {
					outside++
				}
			}
			deg := len(neigh)
			if outside > maxOutside || (outside == maxOutside && deg < minDegree) {
				bestV = v
				maxOutside = outside
				minDegree = deg
			}
		}
		if bestV == -1 {
			done[maxIdx] = true
			continue
		}

		// Prepare affected clusters
		oldCluster := clusters[maxIdx]
		// Remove bestV from current cluster
		newCluster := make(VertexSet)
		for u := range oldCluster {
			if u != bestV {
				newCluster[u] = Naught{}
			}
		}
		// Add bestV as its own cluster
		soloCluster := make(VertexSet)
		soloCluster[bestV] = Naught{}

		// Compute metric sums for affected clusters before and after
		oldMetric, err := metric(l, oldCluster)
		if err != nil {
			return nil, err
		}
		newMetric1, err := metric(l, newCluster)
		if err != nil {
			return nil, err
		}
		newMetric2, err := metric(l, soloCluster)
		if err != nil {
			return nil, err
		}

		// todo It should be EITHER newmetric1 >= oldMetric OR newMetric2 >= oldMetric and
		//  then the appropriate action is taken (replacing the extended cluster and removing from the old cluster)
		//  OR removing from the old cluster and adding the new cluster
		if compare(newMetric1+newMetric2, oldMetric) {
			// Accept the move
			clusters[maxIdx] = newCluster
			clusters = append(clusters, soloCluster)
			// Remove any empty clusters
			pruned := clusters[:0]
			for _, c := range clusters {
				if len(c) > 0 {
					pruned = append(pruned, c)
				}
			}
			clusters = pruned
		} else {
			done[maxIdx] = true
		}
	}

	return clusters, nil
}
