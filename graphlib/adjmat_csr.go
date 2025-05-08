// Package graphlib provides basic graph data structures and initialisation helpers.
// It purposefully keeps the surface small so you can evolve the internals as you
// add algorithms (BFS/DFS, Dijkstra, Pagerank …) later on.

package graphlib

import (
	"errors"
	"sort"
)

// ---------------------------
//  CSR (Compressed Sparse Row)
// ---------------------------

// CSRAdjMatrix stores a sparse adjacency structure in Compressed Sparse Row
// format. Only column indices are stored because the graph is unweighted and
// undirected.
type CSRAdjMatrix struct {
	size   int
	rowPtr []int // len == size; rowPtr[i] is the start of row i in colInd
	colInd []int // column indices concatenated row‑wise
	deg    []int // degree of each vertex
}

// InitializeCSRAdjMat allocates empty CSR slices large enough to describe a
// graph with *size* vertices. The slices grow automatically when you add edges.
func InitializeCSRAdjMat(size int, edges []Edge, directed bool) (*CSRAdjMatrix, error) {
	if size <= 0 {
		return nil, errors.New("size must be > 0")
	}

	// ----- count degrees -----
	deg := make([]int, size)
	for _, e := range edges {
		deg[e.From]++
		if !directed {
			deg[e.To]++
		}
	}

	// prefix‑sum into IndPtr
	rowPtr := make([]int, size+1)
	for i := 0; i < size; i++ {
		rowPtr[i+1] = rowPtr[i] + deg[i]
	}

	// ----- fill Indices -----
	colIdx := make([]int, rowPtr[size]) // len = 2*m for undirected
	// work copy of row write positions
	indices := append([]int(nil), rowPtr...) // cheap O(size) copy

	for _, e := range edges {
		// forward arc
		pos := indices[e.From]
		colIdx[pos] = e.To
		indices[e.From]++

		if !directed && e.From != e.To {
			// symmetric arc
			pos = indices[e.To]
			colIdx[pos] = e.From
			indices[e.To]++
		}
	}

	// Optional: sort each row to make neighbors deterministic and enable
	// binary search.  Works in ~Σ deg_i log deg_i time.
	for v := 0; v < size; v++ {
		lo, hi := rowPtr[v], rowPtr[v+1]
		sort.Ints(colIdx[lo:hi])
		// If duplicates are possible, unique them in‑place (few lines of code).
	}

	return &CSRAdjMatrix{size: size, rowPtr: rowPtr, colInd: colIdx, deg: deg}, nil
}

// AddEdge inserts an undirected edge (u,v).
func (c *CSRAdjMatrix) AddEdge(u, v int) {
	// Append v to row u and u to row v. For real use you may want to keep the
	// colInd arrays sorted and deduplicated.
	c.colInd = append(c.colInd, v)
	c.rowPtr[u+1]++
	c.colInd = append(c.colInd, u)
	c.rowPtr[v+1]++
	// Fix the prefix sums after the row where we inserted. A full-featured
	// library would amortise this, but for clarity we keep it simple.
	for i := max(u, v) + 1; i < len(c.rowPtr); i++ {
		c.rowPtr[i] = c.rowPtr[i-1]
	}
}

// Neighbours returns the vertices adjacent to u.
func (c *CSRAdjMatrix) Neighbours(u int) []int {
	return c.colInd[c.rowPtr[u]:c.rowPtr[u+1]]
}
