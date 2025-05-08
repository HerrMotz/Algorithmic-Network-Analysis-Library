// Package graphlib provides basic graph data structures and initialisation helpers.
// It purposefully keeps the surface small so you can evolve the internals as you
// add algorithms (BFS/DFS, Dijkstra, Pagerank …) later on.

package graphlib

import (
	"errors"
	"fmt"
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
	rowPtr []int // len == size+1; rowPtr[i] is the start of row i in colInd
	colInd []int // column indices concatenated row‑wise
	deg    []int // degree of each vertex
}

// InitializeCSRAdjMat allocates empty CSR slices large enough to describe a
// graph with *size* vertices. The slices grow automatically when you add edges.
func InitializeCSRAdjMat(size int, edges []Edge) (*CSRAdjMatrix, error) {
	if size <= 0 {
		return nil, errors.New("size must be > 0")
	}
	if edges == nil {
		edges = []Edge{}
	}

	// Überprüfe Gültigkeit der Kanten
	for _, e := range edges {
		if e.From < 0 || e.From >= size || e.To < 0 || e.To >= size {
			return nil, fmt.Errorf("ungültige Kante (%d,%d) für Größe %d", e.From, e.To, size)
		}
	}

	// Zähle Grade
	deg := make([]int, size)
	for _, e := range edges {
		deg[e.From]++
		if e.From != e.To { // Selbst-Schleifen nur einmal zählen
			deg[e.To]++
		}
	}

	// Berechne Zeilen-Zeiger (Präfix-Summen)
	rowPtr := make([]int, size+1)
	for i := 0; i < size; i++ {
		rowPtr[i+1] = rowPtr[i] + deg[i]
	}

	// Initialisiere Spalten-Indizes
	colIdx := make([]int, rowPtr[size])
	nextIdx := make([]int, size) // Temporäres Array für nächste freie Position pro Zeile

	// Fülle Spalten-Indizes
	for _, e := range edges {
		pos := rowPtr[e.From] + nextIdx[e.From]
		colIdx[pos] = e.To
		nextIdx[e.From]++

		if e.From != e.To {
			pos = rowPtr[e.To] + nextIdx[e.To]
			colIdx[pos] = e.From
			nextIdx[e.To]++
		}
	}

	// Sortiere jede Zeile für deterministische Nachbarn und binäre Suche
	for v := 0; v < size; v++ {
		lo, hi := rowPtr[v], rowPtr[v+1]
		sort.Ints(colIdx[lo:hi])
	}

	return &CSRAdjMatrix{
		size:   size,
		rowPtr: rowPtr,
		colInd: colIdx,
		deg:    deg,
	}, nil
}

// AddEdge inserts an undirected edge (u,v).
func (c *CSRAdjMatrix) AddEdge(u, v int) {
	// Speichere die ursprünglichen rowPtr-Werte
	oldPtr := make([]int, len(c.rowPtr))
	copy(oldPtr, c.rowPtr)

	// Erzeuge ein neues colInd-Array
	newColInd := make([]int, len(c.colInd)+2)

	// Kopiere die Werte vor der Einfügung
	copied := 0
	for i := 0; i < u; i++ {
		n := oldPtr[i+1] - oldPtr[i]
		copy(newColInd[copied:], c.colInd[oldPtr[i]:oldPtr[i+1]])
		copied += n
	}

	// Füge v in Zeile u ein
	copy(newColInd[copied:], c.colInd[oldPtr[u]:oldPtr[u+1]])
	copied += oldPtr[u+1] - oldPtr[u]
	newColInd[copied] = v
	copied++
	c.rowPtr[u+1]++

	// Kopiere die Werte zwischen u und v
	for i := u + 1; i < v; i++ {
		n := oldPtr[i+1] - oldPtr[i]
		copy(newColInd[copied:], c.colInd[oldPtr[i]:oldPtr[i+1]])
		copied += n
		c.rowPtr[i+1] = c.rowPtr[i]
	}

	// Füge u in Zeile v ein (wenn u != v)
	if u != v {
		copy(newColInd[copied:], c.colInd[oldPtr[v]:oldPtr[v+1]])
		copied += oldPtr[v+1] - oldPtr[v]
		newColInd[copied] = u
		copied++
		c.rowPtr[v+1]++
	}

	// Kopiere die restlichen Werte
	for i := v + 1; i < c.size; i++ {
		n := oldPtr[i+1] - oldPtr[i]
		copy(newColInd[copied:], c.colInd[oldPtr[i]:oldPtr[i+1]])
		copied += n
		c.rowPtr[i+1] = c.rowPtr[i]
	}

	// Aktualisiere colInd
	c.colInd = newColInd[:copied]

	// Sortiere die betroffenen Zeilen
	sort.Ints(c.colInd[c.rowPtr[u]:c.rowPtr[u+1]])
	if u != v {
		sort.Ints(c.colInd[c.rowPtr[v]:c.rowPtr[v+1]])
	}
}

// Neighbours returns the vertices adjacent to u.
func (c *CSRAdjMatrix) Neighbours(u int) []int {
	if u < 0 || u >= c.size {
		return []int{}
	}
	return c.colInd[c.rowPtr[u]:c.rowPtr[u+1]]
}
