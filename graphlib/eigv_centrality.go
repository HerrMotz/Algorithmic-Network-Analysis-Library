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
func (m *AdjMatSparse) PowerIteration(x *Vector, iterations int) error {
	n := int(m.Dim())

	if len(*x) != n {
		return errors.New("vector length does not match matrix dimension")
	}

	var sum float64

	for i := 0; i < iterations; i++ {
		*x, _ = m.MulVec(x) // the error can be ignored because it is checked in this function as well
		// re-normalise
		for _, v := range *x {
			sum += v
		}
		// in place
		x.ScaleVector(1 / sum)
		sum = 0
	}

	return nil // this is not the vector, everything happens in place
}

// PowerIterationRayleigh executes a von Mises / power iteration and
// stops when two different start vectors have come closer than δ or
// when the maximum number of iterations is reached.
func (m *AdjMatSparse) PowerIterationRayleigh(
	x1, x2 *Vector, // start vectors (modified in-place)
	delta float64, // convergence threshold
	iterations int, // hard iteration cap
) error {

	n := int(m.Dim())
	if len(*x1) != n || len(*x2) != n {
		return errors.New("vector length does not match matrix dimension")
	}

	deltaSquared := delta * delta

	for k := 0; k < iterations; k++ {
		// power step
		*x1, _ = m.MulVec(x1)
		*x2, _ = m.MulVec(x2)

		// accumulate ℓ₁-norm
		var sum1, sum2 float64
		for i := 0; i < n; i++ {
			sum1 += (*x1)[i]
			sum2 += (*x2)[i]
		}
		x1.ScaleVector(1 / sum1)
		x2.ScaleVector(1 / sum2)

		// ‖x1 – x2‖₂ convergence test
		var diff float64
		for i := 0; i < n; i++ {
			d := (*x1)[i] - (*x2)[i]
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

func (m *AdjMatSparse) EigenvectorCentrality() {

}
