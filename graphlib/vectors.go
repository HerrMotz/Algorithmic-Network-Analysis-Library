package graphlib

import (
	"fmt"
	"sort"
)

type Vector []float64

func (m *AdjMatSparse) MulVec(v *Vector) ([]float64, error) {
	n := int(m.Dim())
	if len(*v) != n {
		return nil, fmt.Errorf("vector length %d does not match matrix dimension %d", len(*v), n)
	}

	// Output vector
	r := make([]float64, n)

	for i := 0; i < n; i++ {
		start := m.IndPtr[i]
		end := m.IndPtr[i+1]
		sum := 0.0
		for _, j := range m.Indices[start:end] {
			sum += (*v)[j]
		}
		r[i] = sum
	}

	return r, nil
}

func (v *Vector) ScaleVector(n float64) {
	if n == 0 {
		for i := range *v {
			(*v)[i] = 0
		}
		return
	}

	for i := range *v {
		(*v)[i] *= n
	}
}

func (v *Vector) Sort() {
	sort.Float64s(*v) // this sorts ascending
	// you can also use:
	/*
		sort.Slice(*v, func(i, j int) bool {
			return (*v)[i] > (*v)[j] // descending
		})
	*/
}
