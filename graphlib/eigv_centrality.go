package graphlib

import (
	"errors"
	"fmt"
	"math"
)

// ------------------------
// Eigenvector Centralities
// ------------------------

// Solve using vector iteration after von Mises / Power Iteration.
// 1. x(0) may not be orthogonal to v_1 -> else no convergence
// 2. re-normalise after each vec*mat to avoid float overflow

// PowerIteration executes a von Mises iteration for the start vector v, which
// is normalised to sum 1 after every step. It stops when two consecutive
// vectors differ by less than delta (Euclidean norm) or after the given number
// of iterations; with delta 0 it always performs all iterations.
func (m *AdjMatSparse) PowerIteration(v *Vector, delta float64, iterations int) error {
	n := int(m.Dim())

	if len(*v) != n {
		return errors.New("vector length does not match matrix dimension")
	}

	var sum float64
	// calculate the delta between iterations
	var vOld = make(Vector, n)
	for i := 0; i < n; i++ {
		vOld[i] = 0
	}

	deltaSquared := delta * delta

	for i := 0; i < iterations; i++ {
		*v, _ = m.MulVec(v) // the error can be ignored because it is checked in this function as well
		// re-normalise
		for _, v := range *v {
			sum += v
		}
		// in place
		v.ScaleVector(1 / sum)
		sum = 0

		// ‖v1 – v2‖₂ convergence test
		var diff float64
		for i := 0; i < n; i++ {
			d := vOld[i] - (*v)[i]
			diff += d * d
		}

		if diff < deltaSquared { // ‖v_k – v_{k-1}‖₂ < δ, sparing the sqrt call
			return nil
		}

		vOld = *v

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

		if diff < deltaSquared { // ‖v1 – v2‖₂ < δ, sparing the sqrt call
			return nil
		}
	}

	return nil // vectors were updated in-place
}

// katzTolerance is the relative change between two iterations (in the 1-norm)
// below which KatzCentrality considers the iteration converged.
const katzTolerance = 1e-12

// KatzCentrality calculates the Katz centrality x = Σ_k (αA)^k β, i.e. the
// solution of x = αAx + β, by iterating x ← αAx + β from the start vector v.
// alpha (the walk probability) and beta (the ground importance) hold one value
// per vertex. The iteration only converges if alpha is smaller than 1/λ_max,
// where λ_max is the largest eigenvalue of the adjacency matrix; an error is
// returned if it diverges or has not converged after the given number of
// iterations. The result is normalised to sum 1.
func (m *AdjMatSparse) KatzCentrality(v *Vector, alpha *Vector, beta *Vector, iterations int) error {
	n := int(m.Dim())

	if len(*v) != n || len(*alpha) != n || len(*beta) != n {
		return errors.New("vector length does not match matrix dimension")
	}

	for i := 0; i < iterations; i++ {
		// calculate x' := A*x
		next, err := m.MulVec(v)
		if err != nil {
			return err
		}
		// calculate alpha * x' + beta
		if err := next.FMA(alpha, beta); err != nil {
			return err
		}

		// relative change ‖x' – x‖₁ / ‖x'‖₁
		var change, size, sum float64
		for j := range next {
			change += math.Abs(next[j] - (*v)[j])
			size += math.Abs(next[j])
			sum += next[j]
		}
		*v = next
		if math.IsInf(size, 0) || math.IsNaN(size) {
			return errors.New("Katz centrality diverges: alpha must be smaller than 1/λ_max")
		}

		if change <= katzTolerance*size {
			if sum != 0 {
				v.ScaleVector(1 / sum) // in place
			}
			return nil
		}
	}

	return fmt.Errorf("Katz centrality did not converge in %d iterations: alpha must be smaller than 1/λ_max", iterations)
}

// TopKCentralities This function returns the top k centralities for a given vector.
func (v *Vector) TopKCentralities(k int) Vector {
	v.Sort() // they are sorted from smallest to largest with growing index
	n := len(*v)
	if k > n {
		k = n
	}
	return (*v)[n-k:]
}
