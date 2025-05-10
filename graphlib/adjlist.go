package graphlib

import "errors"

// ---------------------------
//  Adjacency List (slices)
// ---------------------------

type AdjList struct {
	Size int            `json:"size,omitempty"`
	List []map[int]bool `json:"list,omitempty"` // List[u] are neighbours of u
	Deg  []int          `json:"deg,omitempty"`
}

// InitializeAdjList creates an adjacency list with *size* isolated vertices.
func InitializeAdjList(size int) (*AdjList, error) {
	if size <= 0 {
		return nil, errors.New("size must be > 0")
	}
	l := &AdjList{
		Size: size,
		List: make([]map[int]bool, size),
	}

	for i := 0; i < size; i++ {
		l.List[i] = make(map[int]bool)
	}

	return l, nil
}

// AddArc adds an arc (u,v) to the graph.
func (l *AdjList) AddArc(arc Arc) {
	if arc.From != arc.To {
		l.List[arc.From][arc.To] = true
	}
}

func (l *AdjList) AddEdge(edge Edge) {
	if edge.U != edge.V {
		l.List[edge.U][edge.V] = true
		l.List[edge.V][edge.U] = true
	}
}
