package graphlib

import (
	"fmt"
)

// Copy Helper: Deep copy AdjList
func (l *AdjList) Copy() *AdjList {
	newList := make([]VertexSet, len(l.List))
	for i, m := range l.List {
		newMap := make(VertexSet, len(m))
		for k, v := range m {
			newMap[k] = v
		}
		newList[i] = newMap
	}
	return &AdjList{
		Size: l.Size,
		List: newList,
	}
}

// RemoveVertexCopy Helper: Remove a vertex and return a new graph
func (l *AdjList) RemoveVertexCopy(v Vertex) *AdjList {
	newL := l.Copy()
	err := newL.RemoveVertex(v)
	if err != nil {
		return nil
	}
	newL.List[v.Id] = nil
	return newL
}

// FirstBestDegree returns the first node with a degree lower than d or returns 0 and true
// if none of the nodes has degree <= d, then the Graph is a 2-plex
func (igraph *InducedGraph) FirstBestDegree(d int) (Vertex, bool) {
	for i := range igraph.GoodVertices {
		cmp, err := igraph.Degree(Vertex{i})

		if err != nil {
			continue
		}
		if cmp < d {
			return Vertex{i}, false
		}
	}
	return Vertex{0}, true // the graph is a 2-plex
}

// NotAdjacent returns two vertices not adjacent to u and not u
func (igraph *InducedGraph) NotAdjacent(u Vertex) (Vertex, Vertex, error) {
	neigh, err := igraph.Neighbours(u)
	if err != nil {
		return Vertex{-1}, Vertex{-1}, fmt.Errorf("failed to get neighbours of vertex %v: %w", u, err)
	}
	found := make([]int, 0, 2)
	for v := range igraph.GoodVertices {
		if v == u.Id {
			continue // skip the vertex itself
		}
		if _, ok := neigh[v]; !ok {
			found = append(found, v)
			if len(found) == 2 {
				return Vertex{found[0]}, Vertex{found[1]}, nil
			}
		}
	}
	return Vertex{-1}, Vertex{-1}, fmt.Errorf("not enough non-neighbors")
}

// TwoPlexNaive implements the naive algorithm
func (igraph *InducedGraph) TwoPlexNaive() (VertexSet, error) {
	n := igraph.Size()
	u, isTwoPlex := igraph.FirstBestDegree(n - 2)
	if isTwoPlex {
		return igraph.GoodVertices, nil
	}

	v, w, err := igraph.NotAdjacent(u)
	if err != nil {
		// this can usually not occur, else it would have been a 2-plex
		return nil, err
	}

	// we need three new copies of InducedGraph, but the
	// reference to AdjList must stay the same so it does not get copied

	H引くu, err1A := igraph.CreateReduced(u) // add no permanent vertex to set
	R1, err1B := H引くu.TwoPlexNaive()
	if err1A != nil || err1B != nil {
		return nil, fmt.Errorf("failed to make recursion step: CreateReduced(%v), TwoPlexPermanent(%v)", err1A, err1B)
	}

	H引くv, err2A := igraph.CreateReduced(v) // add u as permanent vertex
	R2, err2B := H引くv.TwoPlexNaive()
	if err2A != nil || err2B != nil {
		return nil, fmt.Errorf("failed to make recursion step: CreateReduced(%v), TwoPlexPermanent(%v)", err2A, err2B)
	}

	H引くw, err3A := igraph.CreateReduced(w) // add u and v as permanent vertices
	R3, err3B := H引くw.TwoPlexNaive()
	if err3A != nil || err3B != nil {
		return nil, fmt.Errorf("failed to make recursion step: CreateReduced(%v), TwoPlexPermanent(%v)", err3A, err3B)
	}

	// Find the largest node count
	results := []VertexSet{R1, R2, R3}
	maxIdx := 0
	maxNodes := len(results[0])
	for i := 1; i < 3; i++ {
		nodes := results[i].Size()
		if nodes > maxNodes {
			maxIdx = i
			maxNodes = nodes
		}
	}
	return results[maxIdx], nil
}

func runTwoPlexIf(pgraph *PermanentInducedGraph, testVertex Vertex, p1 Vertex, p2 Vertex) (VertexSet, error) {
	if !pgraph.PermanentVertices.HasVertex(testVertex) {
		reduced, errA := pgraph.CreateReduced(testVertex, p1, p2)
		result, errB := reduced.TwoPlexPermanent()
		if errA != nil || errB != nil {
			return nil, fmt.Errorf("failed to make recursion step: CreateReduced(%v), TwoPlexPermanent(%v)", errA, errB)
		}
		return result, nil
	}
	return nil, nil
}

var noV = Vertex{-1} // use -1 to indicate no vertex

// TODO
//  - Add more tests for the TwoPlexPermanent algorithm (bigger graphs for performance testing)
//  - Add the minimal degree optimisation
//  - Add distance improvement (remove all vertices that have distance >= 3 to a permanent vertex)
//  - Store the size of the currently found 2-Plex (stop recursion if it is not larger than the current best)
//  - Remove nodes with at most degree of n*-2

// TwoPlexPermanent implements the algorithm with permanent set
func (pgraph *PermanentInducedGraph) TwoPlexPermanent() (VertexSet, error) {
	n := pgraph.IGraph.Size()
	u, isTwoPlex := pgraph.IGraph.FirstBestDegree(n - 2)
	if isTwoPlex {
		return pgraph.IGraph.GoodVertices, nil
	}

	v, w, err := pgraph.IGraph.NotAdjacent(u)
	if err != nil {
		// this can usually not occur, else it would have been a 2-plex
		return nil, err
	}

	// special case: if u, v, w are all in the permanent vertex set,
	// then we can return a trivial 2-Plex. If there is a larger 2-Plex,
	// it will be found in the other recursive calls.
	if pgraph.PermanentVertices.HasVertex(u) &&
		pgraph.PermanentVertices.HasVertex(v) &&
		pgraph.PermanentVertices.HasVertex(w) {
		return VertexSet{0: Naught{}, 1: Naught{}}, nil
	}

	var results = make([]VertexSet, 0, 3)

	// only call to u if it is not in the permanent vertex set
	if r, err := runTwoPlexIf(pgraph, u, noV, noV); err != nil { // noV means "no vertex"
		return nil, fmt.Errorf("failed to make recursion step: runTwoPlexIf(%v, %v, %v): %w", u, noV, noV, err)
	} else if r != nil {
		results = append(results, r)
	}

	// only call to v if it is not in the permanent vertex set
	if r, err := runTwoPlexIf(pgraph, v, u, noV); err != nil {
		return nil, fmt.Errorf("failed to make recursion step: runTwoPlexIf(%v, %v, %v): %w", v, u, noV, err)
	} else if r != nil {
		results = append(results, r)
	}

	// same for w
	if r, err := runTwoPlexIf(pgraph, w, u, v); err != nil {
		return nil, fmt.Errorf("failed to make recursion step: runTwoPlexIf(%v, %v, %v): %w", w, u, v, err)
	} else if r != nil {
		results = append(results, r)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no 2-Plex found")
	}

	// Find the largest node count
	maxIdx := 0
	maxNodes := len(results[0])
	for i := 1; i < len(results); i++ {
		nodes := results[i].Size()
		if nodes > maxNodes {
			maxIdx = i
			maxNodes = nodes
		}
	}
	return results[maxIdx], nil
}
