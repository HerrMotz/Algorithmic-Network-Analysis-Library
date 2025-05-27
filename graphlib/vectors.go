package graphlib

import "fmt"

type Vector []float64

func (m *AdjMatSparse) MulVec(x *Vector) ([]float64, error) {
	n := int(m.Dim())
	if len(*x) != n {
		return nil, fmt.Errorf("vector length %d does not match matrix dimension %d", len(*x), n)
	}

	// Output vector
	r := make([]float64, n)

	for i := 0; i < n; i++ {
		start := m.IndPtr[i]
		end := m.IndPtr[i+1]
		sum := 0.0
		for _, j := range m.Indices[start:end] {
			sum += (*x)[j]
		}
		r[i] = sum
	}

	return r, nil
}

func (vec *Vector) ScaleVector(n float64) {
	if n == 0 {
		for i := range *vec {
			(*vec)[i] = 0
		}
		return
	}

	for i := range *vec {
		(*vec)[i] *= n
	}
}
