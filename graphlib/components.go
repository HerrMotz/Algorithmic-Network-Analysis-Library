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

// CutVertices returns the set of all articulation points of an undirected
// graph stored in adj.  The caller must pass the number of vertices that were
// originally given to InitializeAdjList; vertices are assumed to be numbered
// 0 … size‑1.
//
// The function works for disconnected graphs as well.
func (l *AdjList) CutVertices(size int) (VertexSet, error) {
	disc := make([]int, size) // time the node was discovered in the DFS
	low := make([]int, size)  // the lowest possible discovery time of a node or his descendants possibly using back nodes
	visited := make([]bool, size)
	parent := make([]int, size) // parent in DFS tree (‑1 == root)

	for i := range parent {
		parent[i] = -1
		disc[i] = -1
		low[i] = -1
	}

	//  “set” that will hold the result
	cut := make(VertexSet)

	//  DFS
	time := 0
	var dfs func(u int) error
	dfs = func(u int) error {
		visited[u] = true
		disc[u] = time
		low[u] = time
		time++

		childCnt := 0

		neigh, err := l.Neighbours(Vertex{u})
		if err != nil {
			return err
		}

		for v := range neigh {
			if !visited[v] {
				parent[v] = u
				childCnt++
				if err := dfs(v); err != nil {
					return err
				}

				// Update low value of u.
				if low[v] < low[u] {
					low[u] = low[v]
				}

				// Articulation‑point tests.
				if parent[u] == -1 && childCnt > 1 { // root with >1 child
					cut[u] = Naught{}
				}
				if parent[u] != -1 && low[v] >= disc[u] { // bridge‑like subtree
					cut[u] = Naught{}
				}
			} else if v != parent[u] { // back edge
				if disc[v] < low[u] {
					low[u] = disc[v]
				}
			}
		}
		return nil
	}

	// run DFS from every component root
	for i := 0; i < size; i++ {
		if l.HasVertex(Vertex{i}) && !visited[i] {
			if err := dfs(i); err != nil {
				return nil, err
			}
		}
	}

	return cut, nil
}
