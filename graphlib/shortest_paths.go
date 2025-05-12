package graphlib

/*
// directedAcyclicGraph returns the vertices of a directed acyclic graph (DAG) for all nodes in the connected component
// of the given node.
func (l *AdjList) directedAcyclicGraph(start *Vertex) ([]Vertex, []int) {
	// Initialise a queue. Here would be the chance to use a heuristic, which tells
	// us something about the maximum queue size. todo
	// For now, we just use a slice and append to it.
	heuristicValue := 15 // e.g. the maximum vertex degree

	queue := make([]int, 0, heuristicValue)
	visited := make([]bool, heuristicValue)
	parents := make([]Vertex, l.Size)
	distances := make([]int, l.Size)

	queue = append(queue, start.Id)
	visited[start.Id] = true
	parents[start.Id] = Vertex{}
	distances[start.Id] = 0

	// BFS
	for len(queue) > 0 {

		// Iterate over all neighbours of v
		neighbours, err := l.Neighbours(Vertex{v})
		if err != nil {
			return nil, nil
		}

		for k := range neighbours {
			if !visited[k] {
				queue = append(queue, k)
				visited[k] = true
				parents[k] = Vertex{v}
				distances[k] = distances[v] + 1
			}
		}
	}

	return parents, distances
}

func (l *AdjList) singleSourceShortestPath(start *Vertex) []int {
	_, parents := l.directedAcyclicGraph(start)
	return parents
}*/
