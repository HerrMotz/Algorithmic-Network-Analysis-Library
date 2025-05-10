package graphlib

import "errors"

// ---------------------------
//  Adjacency Matrix (dense)
// ---------------------------

// AdjMatDense stores a *symmetric* adjacency matrix for an undirected graph. The
// matrix is flattened into a single slice to optimise cache‑locality and avoid
// the [][]bool allocation overhead.  The zero‑value represents *no* edge.
//
// Index layout (row‑major):
//   idx = i*size + j
// The matrix is kept symmetric by mirroring writes.
//

type AdjMatDense struct {
	Size int    `json:"size,omitempty"`
	Data []bool `json:"data,omitempty"`
	Deg  []int  `json:"deg,omitempty"`
}

// InitializeAdjMat returns a zero‑filled symmetric adjacency matrix able to
// hold *size* vertices.
func InitializeAdjMat(size int) (*AdjMatDense, error) {
	if size <= 0 {
		return nil, errors.New("size must be > 0")
	}
	m := &AdjMatDense{
		Size: size,
		Data: make([]bool, size*size),
		Deg:  make([]int, size),
	}
	return m, nil
}

// AddEdge adds an (undirected) edge (u,v).
func (m *AdjMatDense) AddEdge(edge Edge) error {
	var u, v = edge.U, edge.V

	if m.Size == 0 {
		return errors.New("graph is empty")
	}

	if !m.HasVertex(u) || !m.HasVertex(v) {
		return errors.New("vertex does not exist")
	}

	idx := u.Id*m.Size + v.Id // from u to v
	jdx := v.Id*m.Size + u.Id // from v to u

	if m.Data[jdx] || m.Data[idx] {
		return errors.New("edge already exists in at least one direction")
	}

	m.Data[idx],
		m.Data[jdx] = true, true

	m.Deg[u.Id]++
	m.Deg[v.Id]++

	return nil
}

// AddArc adds a(n) (directed) arc (u,v).
func (m *AdjMatDense) AddArc(arc Arc) error {
	var u, v = arc.From, arc.To

	if m.Size == 0 {
		return errors.New("graph is empty")
	}

	if !m.HasVertex(u) || !m.HasVertex(v) {
		return errors.New("vertex does not exist")
	}

	idx := u.Id*m.Size + v.Id // from u to v

	if m.Data[idx] {
		return EdgeOrArcExistsError
	}

	m.Data[idx] = true
	return nil
}

// HasEdge reports whether (u,v) exists.
func (m *AdjMatDense) HasEdge(edge Edge) bool {
	return m.Data[edge.U.Id*m.Size+edge.V.Id]
}

func (m *AdjMatDense) HasArc(arc Arc) bool {
	return m.Data[arc.From.Id*m.Size+arc.To.Id]
}

func (m *AdjMatDense) HasVertex(v Vertex) bool {
	return v.Id >= 0 && v.Id < m.Size
}

func (m *AdjMatDense) Neighbours(u Vertex) ([]int, error) {
	if u.Id < 0 || u.Id >= m.Size {
		return nil, errors.New("vertex does not exist")
	}
	var neighbours []int
	for i := 0; i < m.Size; i++ {
		if m.Data[u.Id*m.Size+i] {
			neighbours = append(neighbours, i)
		}
	}

	return neighbours, nil
}

func (m *AdjMatDense) Degree(u Vertex) (int, error) {
	if u.Id < 0 || u.Id >= m.Size {
		return 0, errors.New("vertex does not exist")
	}
	var degree int
	for i := 0; i < m.Size; i++ {
		if m.Data[u.Id*m.Size+i] {
			degree++
		}
	}

	return degree, nil
}
