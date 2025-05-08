package graphlib

import "errors"

// ---------------------------
//  Adjacency List (slices)
// ---------------------------

type AdjList struct {
	size int
	list [][]int // list[u] are neighbours of u
}

// InitializeAdjList creates an adjacency list with *size* isolated vertices.
func InitializeAdjList(size int) (*AdjList, error) {
	if size <= 0 {
		return nil, errors.New("size must be > 0")
	}
	l := &AdjList{
		size: size,
		list: make([][]int, size),
	}
	return l, nil
}

// AddEdge appends v to u and u to v.
func (l *AdjList) AddEdge(u, v int) {
	l.list[u] = append(l.list[u], v)
	l.list[v] = append(l.list[v], u)
}
