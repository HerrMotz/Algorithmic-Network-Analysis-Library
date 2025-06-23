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

func (l *AdjList) HalveClustering(metric func(VertexSet) float64, partialModularityCompare func(a, b float64) bool) []VertexSet {
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

		// If cannot split further (isolated node, or all unreachable), add as is
		if maxDist <= 0 {
			clustering = append(clustering, cluster)
			continue
		}

		// BFS from u and v (in induced subgraph) for ceil(maxDist/2) steps
		depth := (maxDist + 1) / 2
		u := vertexIDs[uIdx]
		v := vertexIDs[vIdx]

		cluster1 := iGraph.BFSWithDepth(u, depth)
		cluster2 := iGraph.BFSWithDepth(v, depth)

		// Assign overlapping nodes arbitrarily (here: to cluster1)
		for node := range cluster2 {
			if _, in1 := cluster1[node]; in1 {
				delete(cluster2, node)
			}
		}

		// Remove empty clusters
		if len(cluster1) == 0 || len(cluster2) == 0 {
			clustering = append(clustering, cluster)
			continue
		}

		// Metric comparison
		if partialModularityCompare(metric(cluster), metric(cluster1)+metric(cluster2)) {
			clustering = append(clustering, cluster)
		} else {
			clusteringToDo = append(clusteringToDo, cluster1)
			clusteringToDo = append(clusteringToDo, cluster2)
		}
	}

	return clustering
}
