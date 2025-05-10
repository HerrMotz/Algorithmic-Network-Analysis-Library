package graphlib

import "errors"

// ---------------------------
//  Adjacency List (slices)
// ---------------------------

type AdjList struct {
	Size int            `json:"size,omitempty"`
	List []map[int]bool `json:"list,omitempty"` // List[u] are neighbours of u
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

	// Initialise the adjacency list for each vertex
	for i := 0; i < size; i++ {
		l.List[i] = make(map[int]bool)
	}

	return l, nil
}

// AddArc adds an arc (u,v) to the graph.
func (l *AdjList) AddArc(arc Arc) error {
	if l.Size == 0 {
		return errors.New("graph is empty")
	}

	if !l.HasVertex(arc.From) || !l.HasVertex(arc.To) {
		return errors.New("vertex does not exist")
	}

	if l.List[arc.From.Id][arc.To.Id] {
		return EdgeOrArcExistsError
	}

	if arc.From != arc.To {
		l.List[arc.From.Id][arc.To.Id] = true
	}

	return nil
}

func (l *AdjList) AddEdge(edge Edge) error {
	if l.Size == 0 {
		return errors.New("graph is empty")
	}

	if !l.HasVertex(edge.U) || !l.HasVertex(edge.V) {
		return errors.New("vertex does not exist")
	}

	if l.List[edge.U.Id][edge.V.Id] || l.List[edge.V.Id][edge.U.Id] {
		return EdgeOrArcExistsError
	}

	if edge.U != edge.V {
		l.List[edge.U.Id][edge.V.Id],
			l.List[edge.V.Id][edge.U.Id] = true, true
	}

	return nil
}

func (l *AdjList) HasEdge(edge Edge) bool {
	return l.List[edge.U.Id][edge.V.Id] || l.List[edge.V.Id][edge.U.Id]
}

func (l *AdjList) HasArc(arc Arc) bool {
	return l.List[arc.From.Id][arc.To.Id]
}

func (l *AdjList) AddVertex() {
	l.Size++
	l.List = append(l.List, make(map[int]bool))
}

func (l *AdjList) HasVertex(v Vertex) bool {
	return v.Id >= 0 && v.Id < l.Size
}

func (l *AdjList) Neighbours(u Vertex) ([]int, error) {
	if !l.HasVertex(u) {
		return nil, errors.New("vertex does not exist")
	}
	var neighbours []int
	for v := range l.List[u.Id] {
		neighbours = append(neighbours, v)
	}
	return neighbours, nil
}
