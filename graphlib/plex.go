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

// TwoPlexRec implements the algorithm
func (igraph *InducedGraph) TwoPlexRec() (VertexSet, error) {
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
	H引くu, err1 := igraph.CreateReducedInducedGraph(u)
	H引くv, err2 := igraph.CreateReducedInducedGraph(v)
	H引くw, err3 := igraph.CreateReducedInducedGraph(w)
	if err1 != nil || err2 != nil || err3 != nil {
		// this can usually not occur, else it would have been a 2-plex
		return nil, fmt.Errorf("failed to create two Plex ReducedInducedGraph: %v %v %v", err1, err2, err3)
	}

	// Recursively call TwoPlexRec on the reduced graphs
	R1, err1 := H引くu.TwoPlexRec()
	R2, err2 := H引くv.TwoPlexRec()
	R3, err3 := H引くw.TwoPlexRec()
	if err1 != nil || err2 != nil || err3 != nil {
		// this can usually not occur, else it would have been a 2-plex
		return nil, fmt.Errorf("failed to create two Plex ReducedInducedGraph: %v %v %v", err1, err2, err3)
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
