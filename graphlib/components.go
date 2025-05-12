package graphlib

func (l *AdjList) naiveConnectedComponents() []int {
	heuristicValue := 15 // todo chance to add a heuristic

	visited := make([]bool, l.Size)
	component := make([]int, l.Size) // list entry start stores the component number of node start
	componentCounter := 0

	for start := 0; start < l.Size; start++ {
		if visited[start] {
			continue
		}

		// start BFS from node start
		visited[start] = true
		component[start] = componentCounter

		queue := NewQueue[int](heuristicValue)
		queue.enqueue(start)

		for queue.size > 0 {
			v := queue.pop()
			neighbours, err := l.Neighbours(Vertex{v})
			if err != nil {
				return nil
			}

			for k := range neighbours {
				if !visited[k] {
					queue.enqueue(k)
					visited[k] = true
					component[k] = component[start]
				}
			}
		}

		// after bfs, increase component counter
		componentCounter++
	}

	return component
}
