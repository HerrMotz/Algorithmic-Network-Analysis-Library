package graphlib

import "errors"

// ---------------------------
//  Adjacency List (slices)
// ---------------------------

type Naught struct{}

type AdjList struct {
	Size int              `json:"size,omitempty"`
	List []map[int]Naught `json:"list,omitempty"` // List[u] are neighbours of u
}

// InitializeAdjList creates an adjacency list with *size* isolated vertices.
func InitializeAdjList(size int) (*AdjList, error) {
	if size <= 0 {
		return nil, errors.New("size must be > 0")
	}
	l := &AdjList{
		Size: size,
		List: make([]map[int]Naught, size),
	}

	// Initialise the adjacency list for each vertex
	for i := 0; i < size; i++ {
		// Todo: If we have knowledge of the average vertex degree, we could preallocate the map with a size.
		l.List[i] = make(map[int]Naught, 4)
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

	if _, ok := l.List[arc.From.Id][arc.To.Id]; ok {
		return EdgeOrArcExistsError
	}

	l.List[arc.From.Id][arc.To.Id] = Naught{}

	return nil
}

func (l *AdjList) AddEdge(edge Edge) error {
	if l.HasArc(Arc{edge.U, edge.V}) || l.HasArc(Arc{edge.V, edge.U}) {
		return EdgeOrArcExistsError
	}

	err := l.AddArc(Arc{edge.U, edge.V})
	if err != nil {
		return err
	}

	err = l.AddArc(Arc{edge.V, edge.U})
	if errors.Is(err, EdgeOrArcExistsError) && edge.V == edge.U {
		// Ignore self-loops
		return nil
	}

	return err
}

func (l *AdjList) HasEdge(edge Edge) bool {
	return l.HasArc(Arc{From: edge.U, To: edge.V}) || l.HasArc(Arc{From: edge.V, To: edge.U})
}

func (l *AdjList) HasArc(arc Arc) bool {
	_, ok := l.List[arc.From.Id][arc.To.Id]
	return ok
}

func (l *AdjList) AddVertex() {
	l.Size++
	l.List = append(l.List, make(map[int]Naught))
}

func (l *AdjList) HasVertex(v Vertex) bool {
	return v.Id >= 0 && v.Id < l.Size
}

func (l *AdjList) RemoveVertex(v Vertex) error {
	if !l.HasVertex(v) {
		return errors.New("vertex does not exist")
	}

	l.List[v.Id] = nil
	l.Size--

	for i := 0; i < l.Size; i++ {
		delete(l.List[i], v.Id)
	}

	return nil
}

func (l *AdjList) RemoveArc() error {
	if l.Size == 0 {
		return errors.New("graph is empty")
	}

	for i := 0; i < l.Size; i++ {
		for j := 0; j < l.Size; j++ {
			if l.HasArc(Arc{From: Vertex{i}, To: Vertex{j}}) {
				delete(l.List[i], j)
			}
		}
	}

	return nil
}

func (l *AdjList) RemoveEdge(edge Edge) error {
	if l.Size == 0 {
		return errors.New("graph is empty")
	}

	if !l.HasVertex(edge.U) || !l.HasVertex(edge.V) {
		return errors.New("vertex does not exist")
	}

	if l.HasEdge(edge) {
		delete(l.List[edge.U.Id], edge.V.Id)
		delete(l.List[edge.V.Id], edge.U.Id)
	} else {
		return EdgeOrArcExistsError
	}

	return nil
}

func (l *AdjList) Neighbours(u Vertex) (map[int]Naught, error) {
	if !l.HasVertex(u) {
		return nil, errors.New("vertex does not exist")
	}
	return l.List[u.Id], nil
}

func (l *AdjList) Degree(u Vertex) (int, error) {
	n, err := l.Neighbours(u)
	return len(n), err
}
