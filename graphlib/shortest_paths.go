package graphlib

import (
	"slices"
)

// DirectedAcyclicGraph returns the vertices of a directed acyclic graph (DAG) for all nodes in the connected component
// of the given node.
func (l *AdjList) DirectedAcyclicGraph(start *Vertex) ([]Vertex, []int) {
	// Initialise a queue. Here would be the chance to use a heuristic, which tells
	// us something about the maximum queue size. todo
	// For now, we just use a slice and append to it.
	heuristicValue := 15 // e.g. the maximum vertex degree

	queue := NewQueue[int](heuristicValue)
	visited := make([]bool, l.Size)
	parents := make([]Vertex, l.Size)
	distances := slices.Repeat([]int{INF}, l.Size)

	queue.enqueue(start.Id)
	visited[start.Id] = true
	parents[start.Id] = Vertex{}
	distances[start.Id] = 0

	// BFS
	for queue.size > 0 {
		v := queue.pop()
		// Iterate over all neighbours of v
		neighbours, err := l.Neighbours(Vertex{v})
		if err != nil {
			return nil, nil
		}

		for k := range neighbours {
			if !visited[k] {
				queue.enqueue(k)
				visited[k] = true
				parents[k] = Vertex{v}
				distances[k] = distances[v] + 1
			}
		}
	}

	return parents, distances
}

// SingleSourceShortestPath starts BFS from the given start node and returns the distance to each node in the connected
// component.
func (l *AdjList) SingleSourceShortestPath(source Vertex) ([]int, error) {
	size := l.Size
	dist := make([]int, size)
	for i := range dist {
		dist[i] = INF
	}
	dist[source.Id] = 0
	queue := NewQueue[int](10) // todo heuristic chance
	visited := make([]bool, l.Size)
	visited[source.Id] = true

	queue.enqueue(source.Id)
	for queue.size > 0 {
		u := queue.pop()

		neighbours, err := l.Neighbours(Vertex{u})
		if err != nil {
			return nil, err
		}
		for v := range neighbours {
			if !visited[v] {
				visited[v] = true
				dist[v] = dist[u] + 1
				queue.enqueue(v)
			}
		}
	}
	return dist, nil
}

func (iGraph *InducedGraph) SingleSourceShortestPath(source Vertex) ([]int, error) {
	size := iGraph.Size()
	dist := make([]int, size)
	for i := range dist {
		dist[i] = INF
	}
	dist[source.Id] = 0
	queue := NewQueue[int](10) // todo heuristic chance
	visited := make([]bool, iGraph.Size())
	visited[source.Id] = true

	queue.enqueue(source.Id)
	for queue.size > 0 {
		u := queue.pop()

		neighbours, err := iGraph.Neighbours(Vertex{u})
		if err != nil {
			return nil, err
		}
		for v := range neighbours {
			if !visited[v] {
				visited[v] = true
				dist[v] = dist[u] + 1
				queue.enqueue(v)
			}
		}
	}
	return dist, nil
}

// NaiveAllPairsShortestPath return the full distance matrix for a given adjacency list
func (l *AdjList) NaiveAllPairsShortestPath() [][]int {
	// we return a full matrix, because the distance from each node to each other is well-defined,
	// even if they are not in the same connected component

	distanceMatrix := make([][]int, l.Size)
	for i := 0; i < l.Size; i++ {
		distanceMatrix[i], _ = l.SingleSourceShortestPath(Vertex{i})
	}

	return distanceMatrix
}

func (iGraph *InducedGraph) NaiveAllPairsShortestPath() [][]int {
	// we return a full matrix, because the distance from each node to each other is well-defined,
	// even if they are not in the same connected component

	distanceMatrix := make([][]int, iGraph.Size())
	for i := 0; i < iGraph.Size(); i++ {
		distanceMatrix[i], _ = iGraph.SingleSourceShortestPath(Vertex{i})
	}

	return distanceMatrix
}

// ImprovedAllPairsShortestPaths computes the full distance matrix using a mix of BFS and neighborhood copying.
func (l *AdjList) ImprovedAllPairsShortestPaths() ([][]int, error) {
	// Assume AdjList was initialized with a fixed size field
	size := l.Size
	// Initialize distance matrix to Infinity
	distanceMatrix := make([][]int, size)
	for i := range distanceMatrix {
		distanceMatrix[i] = make([]int, size)
		for j := range distanceMatrix[i] {
			distanceMatrix[i][j] = INF
		}
	}
	// Track which rows are known
	distancesKnown := make([]bool, size)

	// Build open and closed neighborhood sets
	openNeighbourSets := make([]VertexSet, size)
	closedNeighborSets := make([]VertexSet, size)
	for u := 0; u < size; u++ {
		neighMap, err := l.Neighbours(Vertex{u})
		if err != nil {
			return nil, err
		}
		open := make(VertexSet)
		for v := range neighMap {
			open[v] = Naught{}
		}
		openNeighbourSets[u] = open

		closed := make(VertexSet)
		for v := range open {
			closed[v] = Naught{}
		}
		closed[u] = Naught{}
		closedNeighborSets[u] = closed
	}

	// Compute degree threshold m/n
	totalDeg := l.TotalDegree()
	degreeThreshold := totalDeg / size

	for node := 0; node < size; node++ {
		if distancesKnown[node] {
			continue
		}
		degree := len(openNeighbourSets[node])
		if degree == 0 {
			// isolated vertex
			distancesKnown[node] = true
			distanceMatrix[node][node] = 0
			continue
		}

		// Check if we can use neighbor distances
		allKnown := true
		for neigh := range openNeighbourSets[node] {
			if !distancesKnown[neigh] {
				allKnown = false
				break
			}
		}

		var distances []int
		if allKnown && degree < degreeThreshold {
			// Combine neighbor distances
			distances = make([]int, size)
			for u := 0; u < size; u++ {
				if u == node {
					distances[u] = 0
				} else if _, ok := openNeighbourSets[node][u]; ok {
					distances[u] = 1
				} else {
					minD := INF
					for neigh := range openNeighbourSets[node] {
						// todo: could replace with slices.Min for improved readability
						if d := distanceMatrix[neigh][u]; d < minD {
							minD = d
						}
					}
					distances[u] = minD + 1
				}
			}
		} else {
			// Fallback to BFS
			var err error
			distances, err = l.SingleSourceShortestPath(Vertex{node})
			if err != nil {
				return nil, err
			}
		}

		// Save and mark known
		distanceMatrix[node] = distances
		distancesKnown[node] = true

		// Copy to other nodes with identical neighbourhoods
		for _, other := range getSameNeighborhoodNodes(node, openNeighbourSets, closedNeighborSets) {
			if distancesKnown[other] {
				continue
			}
			copyRow := make([]int, size)
			copy(copyRow, distances)
			copyRow[other] = 0
			if _, ok := openNeighbourSets[node][other]; ok {
				copyRow[node] = 1
			} else {
				copyRow[node] = 2
			}
			distanceMatrix[other] = copyRow
			distancesKnown[other] = true
		}
	}

	return distanceMatrix, nil
}

// getSameNeighborhoodNodes finds other vertices with identical open or closed neighborhoods
func getSameNeighborhoodNodes(node int, openNeighborSets, closedNeighborSets []VertexSet) []int {
	size := len(openNeighborSets)
	var same []int
	for u := 0; u < size; u++ {
		if u == node {
			continue
		}
		if equalSets(openNeighborSets[node], openNeighborSets[u]) || equalSets(closedNeighborSets[node], closedNeighborSets[u]) {
			same = append(same, u)
		}
	}
	return same
}

// equalSets checks if two vertex sets are identical
func equalSets(a, b VertexSet) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}
