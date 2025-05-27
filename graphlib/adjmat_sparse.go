package graphlib

// IMPORTANT NOTICE by Daniel
// This file is an exact copy of https://github.com/sbromberger/graphmatrix/blob/master/graphmatrix.go
// The licence is found below.

import (
	"errors"
	"fmt"
	"sort"
)

/*
MIT License

Copyright (c) 2019 Seth Bromberger
Portions Copyright (c) 2017 James Bowman

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// Graphmatrix provides an implementation of sparse matrices used to describe unweighted graphs.
// Matrices are represented by two vectors representing a CSR sparse matrix index, with no
// `nzval` vector. Methods for setting and testing at a given (row, col) index are provided, as
// well as an iterator over all set points.

// AdjMatSparse holds a row index and vector of column pointers.
// If a point is defined at a particular row i and column j, an
// edge exists between vertex i and vertex j.
// AdjMatSparse thus represent directed graphs; undirected graphs
// must explicitly set the reverse edge from j to i.
type AdjMatSparse struct {
	IndPtr  []uint64 // indexes into Indices - must be twice the width of Indices.
	Indices []uint32 // contains the row values for each column. A stride represents the outneighbors of a vertex at col j.
}

// String is used for pretty printing.
func (m *AdjMatSparse) String() string {
	return fmt.Sprintf("AdjMatSparse %v, %v, size %d", m.IndPtr, m.Indices, m.Dim())
}

// GetRow returns the 'n'th row slice, or an empty slice if empty.
func (m *AdjMatSparse) GetRow(r uint32) ([]uint32, error) {
	if r > uint32(len(m.IndPtr))-1 {
		return []uint32{}, fmt.Errorf("row %d out of bounds (max %d)", r, len(m.IndPtr))
	}
	rowStart := m.IndPtr[r]
	rowEnd := m.IndPtr[r+1]
	return m.Indices[rowStart:rowEnd], nil
}

// cumsum was taken from github.com/james-bowman/sparse.
func cumsum(p []uint64, c []uint64, n uint64) uint64 {
	nz := uint64(0)
	for i := nz; i < n; i++ {
		p[i] = nz
		nz += uint64(c[i])
		c[i] = p[i]
	}
	p[n] = nz
	return nz
}

// compress was modified from github.com/james-bowman/sparse.
func compress(row []uint32, col []uint32, n uint64) (ia []uint64, ja []uint32) {
	w := make([]uint64, n+1)
	ia = make([]uint64, n+1)
	ja = make([]uint32, len(col))

	for _, v := range row {
		w[v]++
	}
	cumsum(ia, w, n)

	for j, v := range col {
		p := w[row[j]]
		ja[p] = v
		w[row[j]]++
	}
	return
}

// inRange returns true if (r, c) is a valid index into v.
func (m *AdjMatSparse) inRange(r, c uint32) bool {
	n := m.Dim()
	return (c < n) && (r < n)
}

// Dim returns the (single-axis) dimension of the AdjMatSparse.
func (m *AdjMatSparse) Dim() uint32 {
	return uint32(len(m.IndPtr) - 1)
}

// N returns the number of defined values in the AdjMatSparse.
func (m *AdjMatSparse) N() uint64 {
	return uint64(len(m.Indices))
}

// GetIndex returns true if the value at (r, c) is defined.
func (m *AdjMatSparse) GetIndex(r, c uint32) bool {
	if uint32(len(m.IndPtr)) <= c+1 {
		return false
	}

	r1 := m.IndPtr[r]
	r2 := m.IndPtr[r+1]
	if r1 >= r2 {
		return false
	}
	_, found := SearchSorted32(m.Indices, c, r1, r2)
	return found
}

// SetIndex sets the value at (r, c) to true.
// This can be a relatively expensive operation as it can force
// reallocation as the vectors increase in size.
func (m *AdjMatSparse) SetIndex(r, c uint32) error {
	if !m.inRange(r, c) {
		return errors.New("index out of range")
	}
	rowStartIdx := m.IndPtr[r] // this is the pointer into the Indices for column c
	rowEndIdx := m.IndPtr[r+1]

	i, found := SearchSorted32(m.Indices, c, rowStartIdx, rowEndIdx)
	if found { // already set
		return nil
	}
	m.Indices = append(m.Indices, 0)
	copy(m.Indices[i+1:], m.Indices[i:])
	m.Indices[i] = c

	for i := int(r + 1); i < len(m.IndPtr); i++ {
		m.IndPtr[i]++
	}
	return nil
}

// returns the maximum uint and its position in the vector.
// returns -1 as position if the vector is empty.
func maxUint32(v []uint32) (max uint32, maxPos int) {
	if len(v) == 0 {
		return 0, -1
	}
	max = 0
	maxPos = 0
	for i, n := range v {
		if n > max {
			max = n
			maxPos = i
		}
	}
	return max, maxPos
}

// NewZero creates an m x m sparse matrix.
func NewZero(m int) (AdjMatSparse, error) {
	if m < 0 {
		return AdjMatSparse{}, errors.New("dimensions must be non-negative")
	}
	i := make([]uint32, 0)
	ip := make([]uint64, m+1)
	return AdjMatSparse{IndPtr: ip, Indices: i}, nil
}

func NewFromSortedIJ(s []uint32, d []uint32) (AdjMatSparse, error) {
	if len(s) != len(d) {
		return AdjMatSparse{}, fmt.Errorf("graph inputs must be of the same length (got %d, %d)", len(s), len(d))
	}
	m1 := s[len(s)-1]     // max s - this is O(1)
	m2, _ := maxUint32(d) // max d - this is O(n)
	m := m1
	if m2 > m1 {
		m = m2
	}
	m++ // m is the number of rows/cols for this matrix.

	ia, ja := compress(s, d, uint64(m))
	// ja, data = dedupe(ia, ja, data, c.r, c.c)
	return AdjMatSparse{ia, ja}, nil
}

// SortIJ sorts two vectors s and d by s, then by d, and eliminates any duplicate pairs.
// Modifies s and d.
func SortIJ(s, d *[]uint32) error {
	if len(*s) != len(*d) {
		return errors.New("inputs must be of the same length")
	}
	sd := make([]uint64, len(*s))
	for i := 0; i < len(*s); i++ {
		sd[i] = uint64((*s)[i])<<32 + uint64((*d)[i])
	}
	sort.Slice(sd, func(i, j int) bool { return sd[i] < sd[j] })
	UniqSorted(&sd)

	for i := 0; i < len(sd); i++ {
		(*s)[i] = uint32(sd[i] >> 32)
		(*d)[i] = uint32(sd[i] & 0x00000000ffffffff)
	}

	(*s) = (*s)[:len(sd)]
	(*d) = (*d)[:len(sd)]
	return nil
}

// UniqSorted deduplicates a sorted vector in place.
func UniqSorted(a *[]uint64) {
	j := 0
	for i := 1; i < len(*a); i++ {
		if (*a)[j] == (*a)[i] {
			continue
		}
		j++
		// preserve the original data
		// in[i], in[j] = in[j], in[i]
		// only set what is required
		(*a)[j] = (*a)[i]
	}
	*a = (*a)[:j+1]
}

// SearchSorted32 finds a value x in sorted vector v.
// Returns index and true/false indicating found.
// lo and hi constrains search to these Indices.
// If lo/hi are out of bounds, return -1 and false unless
// the vector is empty, in which case return 0 and false.
func SearchSorted32(v []uint32, x uint32, lo, hi uint64) (int, bool) {
	ulen := uint64(len(v))
	if ulen == 0 {
		return 0, false
	}
	if lo == hi {
		return int(lo), false
	}
	if ulen < lo || ulen < hi || lo > hi {
		return -1, false
	}
	s := sort.Search(int(hi-lo), func(i int) bool { return v[int(lo)+i] >= x }) + int(lo)
	found := (s < len(v)) && (v[s] == x)

	return s, found
}
