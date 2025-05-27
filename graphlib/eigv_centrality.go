package graphlib

import (
	"errors"
	"fmt"
)

// ------------------------
// Eigenvector Centralities
// ------------------------

// Solve using vector iteration after von Mises / Power Iteration.
// 1. x(0) may not be orthogonal to v_1 -> else no convergence
// 2. re-normalise after each vec*mat to avoid float overflow

// PowerIteration Executes a von Mises iteration for start vector x
// for a fixed number of iterations.
func (m *AdjMatSparse) PowerIteration(v *Vector, iterations int) error {
	n := int(m.Dim())

	if len(*v) != n {
		return errors.New("vector length does not match matrix dimension")
	}

	var sum float64

	for i := 0; i < iterations; i++ {
		*v, _ = m.MulVec(v) // the error can be ignored because it is checked in this function as well
		// re-normalise
		for _, v := range *v {
			sum += v
		}
		// in place
		v.ScaleVector(1 / sum)
		sum = 0
	}

	return nil // this is not the vector, everything happens in place
}

// PowerIterationRayleigh executes a von Mises / power iteration and
// stops when two different start vectors have come closer than δ or
// when the maximum number of iterations is reached.
func (m *AdjMatSparse) PowerIterationRayleigh(
	v1, v2 *Vector, // start vectors (modified in-place)
	delta float64, // convergence threshold
	iterations int, // hard iteration cap
) error {

	n := int(m.Dim())
	if len(*v1) != n || len(*v2) != n {
		return errors.New("vector length does not match matrix dimension")
	}

	deltaSquared := delta * delta

	for k := 0; k < iterations; k++ {
		// power step
		*v1, _ = m.MulVec(v1)
		*v2, _ = m.MulVec(v2)

		// accumulate ℓ₁-norm
		var sum1, sum2 float64
		for i := 0; i < n; i++ {
			sum1 += (*v1)[i]
			sum2 += (*v2)[i]
		}
		v1.ScaleVector(1 / sum1)
		v2.ScaleVector(1 / sum2) // todo can this be done with only one sum?

		// ‖v1 – v2‖₂ convergence test
		var diff float64
		for i := 0; i < n; i++ {
			d := (*v1)[i] - (*v2)[i]
			diff += d * d
		}

		if deltaSquared < delta { // spare the sqrt call
			fmt.Println("[PowItRayleigh] delta criterion reached")
			return nil
		}
	}

	fmt.Println("[PowItRayleigh] max number of iterations reached")
	return nil // vectors were updated in-place
}

/*
// KatzCentrality calculates the katz centrality using power iteration
func (m *AdjMatSparse) KatzCentrality(beta *Vector, iterations int) error {
	n := int(m.Dim())
}*/

// TopKCentralities This function returns the top k centralities for a given vector.
func (v *Vector) TopKCentralities(k int) Vector {
	v.Sort() // they are sorted from smallest to largest with growing index
	n := len(*v)
	if k > n {
		k = n
	}
	return (*v)[n-k:]
}
