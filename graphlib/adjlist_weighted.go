package graphlib

import (
	"errors"
	"sync"
)

// ---------------------------
//  Adjacency List for **weighted** graphs
// ---------------------------

// Weight represents the cost/length/capacity of an arc or edge.
type Weight float64

// WeightedAdjList stores |V| and, for every vertex *u*, a map that associates a
// neighbour *v* with the weight w(u,v).
//
//	List[u][v] = w(u,v)
//
// Using a map keeps look‑ups *O(1)* while allowing a variable vertex degree.
type WeightedAdjList struct {
	Size int              `json:"size,omitempty"`
	List []map[int]Weight `json:"list,omitempty"` // List[u][v] -> weight of (u,v)
}

// -----------------------------------------------------------------------------
//  Constructors & helper routines
// -----------------------------------------------------------------------------

// InitializeWeightedAdjList creates a graph with *size* isolated vertices and
// no arcs.
func InitializeWeightedAdjList(size int) (*WeightedAdjList, error) {
	if size <= 0 {
		return nil, errors.New("size must be > 0")
	}

	g := &WeightedAdjList{
		Size: size,
		List: make([]map[int]Weight, size),
	}

	// Optionally estimate the expected degree to pre‑allocate the maps.
	for i := 0; i < size; i++ {
		g.List[i] = make(map[int]Weight, 4)
	}

	return g, nil
}

// HasVertex reports whether *v* is a valid vertex identifier in the current
// graph. (Vertex is defined elsewhere in the package.)
func (g *WeightedAdjList) HasVertex(v Vertex) bool {
	return v.Id >= 0 && v.Id < g.Size
}

// -----------------------------------------------------------------------------
//  Arcs (directed)
// -----------------------------------------------------------------------------

// AddArc inserts (u,v) with weight *w*.  If the arc already exists we return
// *EdgeOrArcExistsError* so that callers can decide whether to overwrite or not.
func (g *WeightedAdjList) AddArc(arc Arc, w Weight) error {
	if g.Size == 0 {
		return errors.New("graph is empty")
	}
	if !g.HasVertex(arc.From) || !g.HasVertex(arc.To) {
		return errors.New("vertex does not exist")
	}

	if _, ok := g.List[arc.From.Id][arc.To.Id]; ok {
		return EdgeOrArcExistsError
	}

	g.List[arc.From.Id][arc.To.Id] = w
	return nil
}

// HasArc returns (weight, true) iff the arc exists.  The zero value of Weight
// is returned when the arc does not exist.
func (g *WeightedAdjList) HasArc(arc Arc) (Weight, bool) {
	w, ok := g.List[arc.From.Id][arc.To.Id]
	return w, ok
}

// RemoveArc removes (u,v) when it exists. Idempotent – removing a non‑existing
// arc is *not* an error so that callers can blindly try both directions.
func (g *WeightedAdjList) RemoveArc(arc Arc) {
	if g.HasVertex(arc.From) {
		delete(g.List[arc.From.Id], arc.To.Id)
	}
}

// -----------------------------------------------------------------------------
//  Edges (undirected = both directions)
// -----------------------------------------------------------------------------

// AddEdge inserts the undirected edge {u,v} with weight *w*.  Under the hood we
// store both directions. Self‑loops are allowed and stored once.
func (g *WeightedAdjList) AddEdge(edge Edge, w Weight) error {
	if _, ok := g.HasArc(Arc{edge.U, edge.V}); ok {
		// (u,v) already present – treat as duplicate edge.
		return EdgeOrArcExistsError
	}
	if _, ok := g.HasArc(Arc{edge.V, edge.U}); ok {
		return EdgeOrArcExistsError
	}

	if err := g.AddArc(Arc{edge.U, edge.V}, w); err != nil {
		return err
	}
	if edge.U != edge.V {
		if err := g.AddArc(Arc{edge.V, edge.U}, w); err != nil {
			return err
		}
	}
	return nil
}

// RemoveEdge removes *both* directions. It returns *EdgeOrArcNotFoundError*
// when the edge was absent.
func (g *WeightedAdjList) RemoveEdge(edge Edge) error {
	removed := false
	if _, ok := g.HasArc(Arc{edge.U, edge.V}); ok {
		g.RemoveArc(Arc{edge.U, edge.V})
		removed = true
	}
	if _, ok := g.HasArc(Arc{edge.V, edge.U}); ok {
		g.RemoveArc(Arc{edge.V, edge.U})
		removed = true
	}

	if !removed {
		return EdgeOrArcNotFoundError
	}
	return nil
}

// HasEdge reports (weight, true) when at least one direction of {u,v} exists.
// If *both* directions exist and weights differ the weight of (u,v) is
// returned. Use *HasArc* if you need the directed variant.
func (g *WeightedAdjList) HasEdge(edge Edge) (Weight, bool) {
	if w, ok := g.HasArc(Arc{edge.U, edge.V}); ok {
		return w, true
	}
	if w, ok := g.HasArc(Arc{edge.V, edge.U}); ok {
		return w, true
	}
	return 0, false
}

// -----------------------------------------------------------------------------
//  Vertex management
// -----------------------------------------------------------------------------

// AddVertex appends an isolated vertex (with id == Size) to the graph.
func (g *WeightedAdjList) AddVertex() {
	g.Size++
	g.List = append(g.List, make(map[int]Weight))
}

// RemoveVertex removes *v* and every incident arc.  Note: existing vertex IDs
// remain valid; the removed entry becomes *nil* to keep O(1) access.
func (g *WeightedAdjList) RemoveVertex(v Vertex) error {
	if !g.HasVertex(v) {
		return errors.New("vertex does not exist")
	}

	g.List[v.Id] = nil
	// Do *not* shrink Size or reindex vertices – that would change their IDs.

	for i := range g.List {
		delete(g.List[i], v.Id)
	}
	return nil
}

// -----------------------------------------------------------------------------
//  Statistics & helpers
// -----------------------------------------------------------------------------

// Neighbours returns the map "v -> weight" of vertex *u*.
func (g *WeightedAdjList) Neighbours(u Vertex) (map[int]Weight, error) {
	if !g.HasVertex(u) {
		return nil, errors.New("vertex does not exist")
	}
	return g.List[u.Id], nil
}

// Degree is the out‑degree of *u*. For undirected graphs use this with either
// direction or maintain symmetry yourself.
func (g *WeightedAdjList) Degree(u Vertex) (int, error) {
	n, err := g.Neighbours(u)
	return len(n), err
}

// TotalDegree counts the number of stored arcs (directed). For an undirected
// graph this is 2*|E| except for self‑loops which contribute 1.
func (g *WeightedAdjList) TotalDegree() int {
	total := 0
	for _, m := range g.List {
		total += len(m)
	}
	return total
}

// TotalDegreeParallel works like *TotalDegree* but processes the vertex set in
// *numWorkers* goroutines. Useful for very large graphs.
func (g *WeightedAdjList) TotalDegreeParallel(numWorkers int) int {
	size := len(g.List)
	chunk := (size + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup
	sums := make(chan int, numWorkers)

	for i := 0; i < size; i += chunk {
		wg.Add(1)
		end := i + chunk
		if end > size {
			end = size
		}
		go func(sub []map[int]Weight) {
			defer wg.Done()
			s := 0
			for _, m := range sub {
				s += len(m)
			}
			sums <- s
		}(g.List[i:end])
	}

	wg.Wait()
	close(sums)

	total := 0
	for s := range sums {
		total += s
	}
	return total
}

// TotalWeight returns \u03A3 w(u,v) over *all* stored arcs.
func (g *WeightedAdjList) TotalWeight() Weight {
	var sum Weight
	for _, m := range g.List {
		for _, w := range m {
			sum += w
		}
	}
	return sum
}

// TotalWeightParallel parallelises the accumulation of the graph weight across
// *numWorkers* goroutines.
func (g *WeightedAdjList) TotalWeightParallel(numWorkers int) Weight {
	size := len(g.List)
	chunk := (size + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup
	sums := make(chan Weight, numWorkers)

	for i := 0; i < size; i += chunk {
		wg.Add(1)
		end := i + chunk
		if end > size {
			end = size
		}
		go func(sub []map[int]Weight) {
			defer wg.Done()
			var s Weight
			for _, m := range sub {
				for _, w := range m {
					s += w
				}
			}
			sums <- s
		}(g.List[i:end])
	}

	wg.Wait()
	close(sums)

	var total Weight
	for s := range sums {
		total += s
	}
	return total
}
