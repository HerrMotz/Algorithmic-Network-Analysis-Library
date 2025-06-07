package graphlib

import (
	"errors"
	"fmt"
	"sync"
)

// ---------------------------
//  Adjacency List (slices of maps)
// ---------------------------

type Naught struct{}

type AdjList struct {
	Size int         `json:"size,omitempty"`
	List []VertexSet `json:"list,omitempty"` // List[u] are neighbours of u
}

// InitializeAdjList creates an adjacency list with *size* isolated vertices.
func InitializeAdjList(size int) (*AdjList, error) {
	if size <= 0 {
		return nil, errors.New("size must be > 0")
	}
	l := &AdjList{
		Size: size,
		List: make([]VertexSet, size),
	}

	// Initialise the adjacency list for each vertex
	for i := 0; i < size; i++ {
		// Todo: If we have knowledge of the average vertex degree, we could preallocate the map with a size.
		l.List[i] = make(VertexSet, 4)
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
	l.List = append(l.List, make(VertexSet))
}

func (l *AdjList) HasVertex(v Vertex) bool {
	return v.Id >= 0 && v.Id < l.Size
}

func (l *AdjList) RemoveVertex(v Vertex) error {
	// 1) check existence
	if !l.HasVertex(v) {
		return fmt.Errorf("vertex %d does not exist", v.Id)
	}

	// 2) delete any edges pointing to v
	for u := range l.List {
		delete(l.List[u], v.Id)
	}

	// 3) remove the v-th row from the slice
	//    this shifts all later vertices “up” by one
	l.List = append(l.List[:v.Id], l.List[v.Id+1:]...)
	l.Size--

	// 4) re-build each adjacency‐map so that any neighbor ≥ v.Id
	//    gets renumbered down by one
	for u, nbrs := range l.List {
		newMap := make(VertexSet, len(nbrs))
		for w, mark := range nbrs {
			if w > v.Id {
				newMap[w-1] = mark
			} else {
				newMap[w] = mark
			}
		}
		l.List[u] = newMap
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

func (l *AdjList) Neighbours(u Vertex) (VertexSet, error) {
	if !l.HasVertex(u) {
		return nil, errors.New("vertex does not exist")
	}
	return l.List[u.Id], nil
}

func (l *AdjList) Degree(u Vertex) (int, error) {
	n, err := l.Neighbours(u)
	return len(n), err
}

func (l *AdjList) TotalDegree() int {
	total := 0
	for _, m := range l.List {
		total += len(m)
	}
	return total
}

func (l *AdjList) TotalDegreeParallel(numWorkers int) int {
	size := len(l.List)
	chunk := (size + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup
	sums := make(chan int, numWorkers)

	for i := 0; i < size; i += chunk {
		wg.Add(1)
		end := i + chunk
		if end > size {
			end = size
		}
		go func(sub []VertexSet) {
			defer wg.Done()
			s := 0
			for _, m := range sub {
				s += len(m) // to make life easier, this does not use the getters and setters
			}
			sums <- s
		}(l.List[i:end])
	}

	wg.Wait()
	close(sums)

	total := 0
	for s := range sums {
		total += s
	}
	return total
}
