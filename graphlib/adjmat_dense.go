package graphlib

import "errors"

// ---------------------------
//  Adjacency Matrix (dense)
// ---------------------------

// AdjMatrix stores a *symmetric* adjacency matrix for an undirected graph. The
// matrix is flattened into a single slice to optimise cache‑locality and avoid
// the [][]bool allocation overhead.  The zero‑value represents *no* edge.
//
// Index layout (row‑major):
//   idx = i*size + j
// The matrix is kept symmetric by mirroring writes.
//

type AdjMatrix struct {
	size int
	data []bool
}

// InitializeAdjMat returns a zero‑filled symmetric adjacency matrix able to
// hold *size* vertices.
func InitializeAdjMat(size int) (*AdjMatrix, error) {
	if size <= 0 {
		return nil, errors.New("size must be > 0")
	}
	m := &AdjMatrix{
		size: size,
		data: make([]bool, size*size),
	}
	return m, nil
}

// SetEdge adds an undirected edge (u,v).
func (m *AdjMatrix) SetEdge(u, v int) {
	idx := u*m.size + v
	jdx := v*m.size + u
	m.data[idx], m.data[jdx] = true, true
}

// HasEdge reports whether (u,v) exists.
func (m *AdjMatrix) HasEdge(u, v int) bool {
	return m.data[u*m.size+v]
}
