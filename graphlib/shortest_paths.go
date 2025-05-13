package graphlib

import (
	"math"
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
	distances := slices.Repeat([]int{math.MaxUint32}, l.Size)

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
func (l *AdjList) SingleSourceShortestPath(start *Vertex) []int {
	// Note: The parents array could be spared.
	_, d := l.DirectedAcyclicGraph(start)
	return d
}

// NaiveAllPairsShortestPath return the full distance matrix for a given adjacency list
func (l *AdjList) NaiveAllPairsShortestPath() [][]int {
	// we return a full matrix, because the distance from each node to each other is well-defined,
	// even if they are not in the same connected component

	distanceMatrix := make([][]int, l.Size)
	for i := 0; i < l.Size; i++ {
		distanceMatrix[i] = l.SingleSourceShortestPath(&Vertex{i})
	}

	return distanceMatrix
}
