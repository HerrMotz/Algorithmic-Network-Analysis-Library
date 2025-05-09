package graphlib

import "errors"

// ---------------------------
//  Adjacency List (slices)
// ---------------------------

type AdjList struct {
	size int
	list []map[int]bool // list[u] are neighbours of u
}

// InitializeAdjList creates an adjacency list with *size* isolated vertices.
func InitializeAdjList(size int) (*AdjList, error) {
	if size <= 0 {
		return nil, errors.New("size must be > 0")
	}
	l := &AdjList{
		size: size,
		list: make([]map[int]bool, size),
	}

	for i := 0; i < size; i++ {
		l.list[i] = make(map[int]bool)
	}

	return l, nil
}

// AddArc adds an arc (u,v) to the graph.
func (l *AdjList) AddArc(arc Arc) {
	if arc.From != arc.To {
		l.list[arc.From][arc.To] = true
	}
}

func (l *AdjList) AddEdge(edge Edge) {
	if edge.u != edge.v {
		l.list[edge.u][edge.v] = true
		l.list[edge.v][edge.u] = true
	}
}
