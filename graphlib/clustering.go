package graphlib

import (
	"errors"
	"maps"
	"math"
	"math/rand"
	"slices"
	"sort"

	"gonum.org/v1/gonum/stat/combin"
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
		return 0, nil // this is a special case
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
func PartialDisagreementMetric(l *AdjList, C VertexSet) (float64, error) {
	if len(C) < 2 {
		// the graph contains 1 vertex, the binominal coefficient is then 0 and the edge count is also 0
		return 0, nil
	}

	// Count number of edges in the induced subgraph
	edgeCount := 0
	for u := range C {
		for v := range l.List[u] {
			if _, ok := C[v]; ok {
				edgeCount++
			}
		}
	}

	return float64(combin.Binomial(len(C), 2) - edgeCount), nil
}

func PartialDisagreementCompare(a, b float64) bool {
	if a > b {
		return false
	}
	return true
}

// strictlyBetter reports whether a is strictly better than b, given a compare
// function that reports whether its first argument is at least as good as its
// second (such as PartialModularityCompare or PartialDisagreementCompare).
// Because compare is a plain ordering of numbers, it ranks metric values and
// metric gains alike, whether the metric is maximised or minimised.
func strictlyBetter(compare func(a, b float64) bool, a, b float64) bool {
	return compare(a, b) && !compare(b, a)
}

// TotalDisagreementMetric calculates the full formula d(C) as given in the worksheet
func TotalDisagreementMetric(l *AdjList, C []VertexSet) (float64, error) {
	sum := 0.0
	for _, cluster := range C {
		q, err := PartialDisagreementMetric(l, cluster)
		if err != nil {
			return 0, err
		}
		sum += q
	}

	// calculate total edge count in the graph
	twoTimesM := l.TotalDegree()
	return float64(twoTimesM/2) + sum, nil
}

// --------------------------------------------------------------------
// Heuristic & Operation 1 - Merge
// --------------------------------------------------------------------

// MergeClusteringGreedyBFS starts with one cluster per vertex and merges every
// leaf into the cluster of its neighbour. Then, for growing cluster sizes k,
// every cluster of size k is merged with the neighbouring cluster that improves
// the metric the most. compare(a, b) must report whether a is at least as good
// as b.
//
// To bound the work per cluster, at most four neighbouring clusters are
// inspected. They are sampled at random from a generator with a fixed seed, so
// the result is reproducible.
func (l *AdjList) MergeClusteringGreedyBFS(
	metric func(*AdjList, VertexSet) (float64, error),
	compare func(newMetric, oldMetric float64) bool,
) ([]VertexSet, error) {

	// parameter for search width for each step
	branchingLimit := 4 - 1
	sampleRng := rand.New(rand.NewSource(1))

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
		// Only consider active clusters of size k (in index order, so that
		// the result does not depend on Go's random map iteration order)
		var clusterIndices []int
		for _, idx := range slices.Sorted(maps.Keys(activeClusters)) {
			if len(clusters[idx]) == k {
				clusterIndices = append(clusterIndices, idx)
			}
		}
		for _, i := range clusterIndices {
			cluster := clusters[i]
			neighbourClusters := make(VertexSet)
			for v := range cluster {
				neigh, _ := l.Neighbours(Vertex{v})
				for n := range neigh {
					j := vertexToCluster[n]
					if j != i {
						if _, ok := activeClusters[j]; ok {
							neighbourClusters[j] = Naught{}
						}
					}
				}
			}
			// Find best neighbouring cluster to merge
			bestJ := -1
			var bestGain float64
			foundImprovement := false
			// use the best improvement after at most k steps; the neighbouring
			// clusters are inspected in a random order from the seeded generator
			branchCount := 0
			candidates := slices.Sorted(maps.Keys(neighbourClusters))
			sampleRng.Shuffle(len(candidates), func(a, b int) {
				candidates[a], candidates[b] = candidates[b], candidates[a]
			})
			for _, j := range candidates {
				if branchCount > branchingLimit {
					break
				}
				branchCount++

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
					gain := newMetric - oldMetric
					if !foundImprovement || strictlyBetter(compare, gain, bestGain) {
						bestGain = gain
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
	for _, idx := range slices.Sorted(maps.Keys(activeClusters)) {
		if clusters[idx] != nil && len(clusters[idx]) > 0 {
			result = append(result, clusters[idx])
		}
	}
	return result, nil
}

// MergeClusteringMetricWithRandomPivot merges clusters by picking a random active cluster
// and merging it with the best possible neighbouring cluster according to the metric.
//
// rng determines the random order in which clusters are picked. Pass a seeded
// generator, e.g. rand.New(rand.NewSource(1)), to get reproducible results; if
// rng is nil, a randomly seeded generator is used.
func (l *AdjList) MergeClusteringMetricWithRandomPivot(
	rng *rand.Rand,
	metric func(*AdjList, VertexSet) (float64, error),
	compare func(newMetric, oldMetric float64) bool,
) ([]VertexSet, error) {
	if rng == nil {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}

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

	// STEP 4: Main merging procedure using metric
	for {
		// Gather all current active cluster indices (sorted first, so that the
		// shuffle below is the only source of randomness)
		clusterIndices := slices.Sorted(maps.Keys(activeClusters))
		if len(clusterIndices) <= 1 {
			break // Only one cluster left, or none
		}

		// Try clusters in random order to avoid bias
		rng.Shuffle(len(clusterIndices), func(i, j int) {
			clusterIndices[i], clusterIndices[j] = clusterIndices[j], clusterIndices[i]
		})
		merged := false

		for _, chosenIdx := range clusterIndices {
			cluster := clusters[chosenIdx]

			// Find neighbour clusters
			neighbourClusters := make(VertexSet)
			for v := range cluster {
				neigh, _ := l.Neighbours(Vertex{v})
				for n := range neigh {
					j := vertexToCluster[n]
					if j != chosenIdx {
						if _, ok := activeClusters[j]; ok {
							neighbourClusters[j] = Naught{}
						}
					}
				}
			}

			// Gather all neighbour indices
			neighbourIndices := slices.Sorted(maps.Keys(neighbourClusters))
			if len(neighbourIndices) == 0 {
				continue // No neighbours to merge with, try next cluster
			}

			// Find the best neighbour according to the metric
			bestNeighbour := -1
			bestGain := 0.0
			foundImprovement := false
			m1, err := metric(l, cluster)
			if err != nil {
				return nil, err
			}

			for _, j := range neighbourIndices {
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
					gain := newMetric - oldMetric
					if !foundImprovement || strictlyBetter(compare, gain, bestGain) {
						bestGain = gain
						bestNeighbour = j
						foundImprovement = true
					}
				}
			}

			if foundImprovement && bestNeighbour != -1 {
				// Merge clusters: chosenIdx <- bestNeighbour
				for u := range clusters[bestNeighbour] {
					clusters[chosenIdx][u] = Naught{}
					vertexToCluster[u] = chosenIdx
				}
				clusters[bestNeighbour] = nil
				delete(activeClusters, bestNeighbour)
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
	for _, idx := range slices.Sorted(maps.Keys(activeClusters)) {
		if clusters[idx] != nil && len(clusters[idx]) > 0 {
			result = append(result, clusters[idx])
		}
	}
	return result, nil
}

// ComponentClustering returns a clustering of the graph where each connected component is its own cluster.
// Each cluster is a set (VertexSet) of vertices. Clusters are ordered by their smallest vertex.
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

	// Convert map to slice of sets. Component IDs are numbered in order of
	// their smallest vertex, so sorting them gives a reproducible order.
	result := make([]VertexSet, 0, len(clusters))
	for _, compID := range slices.Sorted(maps.Keys(clusters)) {
		result = append(result, clusters[compID])
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
		for neighbour := range iGraph.AdjList.List[item.v] {
			if _, good := iGraph.GoodVertices[neighbour]; !good {
				continue // skip nodes not in the induced subgraph
			}
			if _, seen := visited[neighbour]; !seen {
				visited[neighbour] = struct{}{}
				queue = append(queue, struct{ v, d int }{neighbour, item.d + 1})
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

		// Find the pair (u, v) with maximal distance d in the current cluster using the precomputed distance matrix.
		// Vertices are visited in ascending order, so ties are always broken the same way.
		maxDist := -1
		var u, v int
		members := slices.Sorted(maps.Keys(cluster))
		for _, x := range members {
			for _, y := range members {
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

// iceTolerance is the smallest change of the summed metric that ICEClustering
// accepts as an improvement. Smaller changes are treated as ties, so that
// floating-point noise cannot move a vertex back and forth forever.
const iceTolerance = 1e-12

// ICEClustering starts from a clustering into neighbourhoods and improves it by
// moving single vertices between clusters:
//
//  1. Vertices are visited by ascending degree (ties: smaller id first). Every
//     vertex that is not clustered yet opens a new cluster together with its
//     neighbours that are not clustered yet.
//
//  2. The largest unfinished cluster S is examined. Its candidate is the vertex v
//     with the most neighbours outside S (ties: smaller degree, then smaller id).
//     v may move into a cluster T that contains one of its neighbours, or into a
//     new cluster of its own. The move that improves the summed metric the most,
//     i.e. that has the best value of
//
//     metric(S\{v}) + metric(T∪{v}) - metric(S) - metric(T)
//
//     (an empty or new cluster counts as 0), is applied. If no move improves the
//     sum by more than iceTolerance, S is marked as finished.
//
//  3. Step 2 is repeated until every cluster is finished.
//
// compare(a, b) must report whether a is at least as good as b. Under modularity
// and disagreement, moving v into a cluster without any of its neighbours is
// never better than giving v its own cluster, so such clusters are skipped.
// Every applied move strictly improves the summed metric, which guarantees
// termination, and every vertex ends up in exactly one cluster.
func (l *AdjList) ICEClustering(
	metric func(*AdjList, VertexSet) (float64, error),
	compare func(a, b float64) bool,
) ([]VertexSet, error) {
	n := l.Size

	// 1. Sort vertices by degree ascending (start with the smallest degree);
	//    the stable sort keeps vertices of equal degree in id order
	verticesSortedByDegree := make([]int, n)
	for k := range verticesSortedByDegree {
		verticesSortedByDegree[k] = k
	}
	sort.SliceStable(verticesSortedByDegree, func(i, j int) bool {
		return len(l.List[verticesSortedByDegree[i]]) < len(l.List[verticesSortedByDegree[j]])
	})

	// 2. Initial clustering: every unclustered vertex opens a cluster with its
	//    unclustered neighbours. Only unclustered neighbours are added, so the
	//    clusters are disjoint.
	clusterOf := make([]int, n) // clusterOf[v] is the index of v's cluster
	for v := range clusterOf {
		clusterOf[v] = -1
	}
	var clusters []VertexSet
	for _, v := range verticesSortedByDegree {
		if clusterOf[v] != -1 {
			continue
		}
		idx := len(clusters)
		cluster := VertexSet{v: Naught{}}
		clusterOf[v] = idx
		for u := range l.List[v] {
			if clusterOf[u] == -1 {
				cluster[u] = Naught{}
				clusterOf[u] = idx
			}
		}
		clusters = append(clusters, cluster)
	}

	// 3. Move vertices while that improves the summed metric. Clusters that
	//    become empty stay in the slice (so that indices remain valid) and are
	//    dropped at the end.
	done := make([]bool, len(clusters))
	for {
		// Find the largest unfinished cluster (ties: smaller index)
		s := -1
		for i, c := range clusters {
			if !done[i] && len(c) > 0 && (s == -1 || len(c) > len(clusters[s])) {
				s = i
			}
		}
		if s == -1 {
			break // all done
		}
		source := clusters[s]

		// Pick the candidate vertex to move out of the source cluster
		bestV, maxOutside, minDegree := -1, -1, 0
		for _, v := range slices.Sorted(maps.Keys(source)) {
			outside := 0
			for u := range l.List[v] {
				if clusterOf[u] != s {
					outside++
				}
			}
			deg := len(l.List[v])
			if outside > maxOutside || (outside == maxOutside && deg < minDegree) {
				bestV, maxOutside, minDegree = v, outside, deg
			}
		}

		// Change of the source cluster's metric when bestV leaves it
		metricSource, err := metric(l, source)
		if err != nil {
			return nil, err
		}
		source.RemoveVertex(Vertex{bestV})
		metricSourceWithout := 0.0 // an empty cluster contributes nothing
		if len(source) > 0 {
			if metricSourceWithout, err = metric(l, source); err != nil {
				return nil, err
			}
		}
		removalGain := metricSourceWithout - metricSource

		// Find the most improving move. target == -1 means "own cluster",
		// target == stay means that no move improves the metric.
		const stay = -2
		bestTarget, bestGain := stay, 0.0
		consider := func(target int, gain float64) {
			improves := compare(gain, 0) && math.Abs(gain) > iceTolerance
			if improves && (bestTarget == stay || strictlyBetter(compare, gain, bestGain)) {
				bestTarget, bestGain = target, gain
			}
		}

		// Option 1: bestV becomes a cluster of its own (pointless if it was alone)
		if len(source) > 0 {
			metricSolo, err := metric(l, VertexSet{bestV: Naught{}})
			if err != nil {
				return nil, err
			}
			consider(-1, removalGain+metricSolo)
		}

		// Option 2: bestV joins a cluster that contains one of its neighbours
		targets := make(VertexSet)
		for u := range l.List[bestV] {
			if t := clusterOf[u]; t != s {
				targets[t] = Naught{}
			}
		}
		for _, t := range slices.Sorted(maps.Keys(targets)) {
			target := clusters[t]
			metricTarget, err := metric(l, target)
			if err != nil {
				return nil, err
			}
			target[bestV] = Naught{} // temporarily add bestV
			metricTargetWith, err := metric(l, target)
			delete(target, bestV)
			if err != nil {
				return nil, err
			}
			consider(t, removalGain+metricTargetWith-metricTarget)
		}

		// Apply the best move
		switch bestTarget {
		case stay:
			source[bestV] = Naught{} // no improvement: restore bestV
			done[s] = true
		case -1:
			clusterOf[bestV] = len(clusters)
			clusters = append(clusters, VertexSet{bestV: Naught{}})
			done = append(done, false)
		default:
			clusters[bestTarget][bestV] = Naught{}
			clusterOf[bestV] = bestTarget
		}
	}

	// Drop the clusters that became empty
	var finalClusters []VertexSet
	for _, c := range clusters {
		if len(c) > 0 {
			finalClusters = append(finalClusters, c)
		}
	}
	return finalClusters, nil
}
