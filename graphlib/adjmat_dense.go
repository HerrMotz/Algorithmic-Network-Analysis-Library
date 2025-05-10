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
	}
	return m, nil
}

// AddEdge adds an (undirected) edge (u,v).
func (m *AdjMatDense) AddEdge(edge Edge) {
	var u, v = edge.U, edge.V
	idx := u*m.Size + v
	jdx := v*m.Size + u
	m.Data[jdx], m.Data[idx] = true, true
	m.Data[idx], m.Data[jdx] = true, true
}

// AddArc adds a(n) (directed) arc (u,v).
func (m *AdjMatDense) AddArc(arc Arc) {
	var u, v = arc.From, arc.To
	idx := u*m.Size + v
	jdx := v*m.Size + u
	m.Data[jdx], m.Data[idx] = true, true
	m.Data[idx], m.Data[jdx] = true, true
}

// HasEdge reports whether (u,v) exists.
func (m *AdjMatDense) HasEdge(edge Edge) bool {
	return m.Data[edge.U*m.Size+edge.V]
}
