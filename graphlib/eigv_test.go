package graphlib

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func TestPowerIteration_Correctness_Against_Gonum(t *testing.T) {
	const (
		n          = 10000 // number of nodes
		avgDegree  = 10    // sparse connectivity
		iterations = 100   // power iteration steps
		tol        = 1e-5  // error tolerance
	)

	// Generate a random sparse graph
	source := rand.New(rand.NewSource(42))
	var s, d []uint32
	for i := 0; i < n; i++ {
		for j := 0; j < avgDegree; j++ {
			target := source.Intn(n)
			s = append(s, uint32(i))
			d = append(d, uint32(target))
		}
	}

	if err := SortIJ(&s, &d); err != nil {
		t.Fatalf("failed to sort input: %v", err)
	}

	adj, err := NewFromSortedIJ(s, d)
	if err != nil {
		t.Fatalf("failed to build sparse matrix: %v", err)
	}

	// Initial uniform vector
	start := make(Vector, n)
	for i := range start {
		start[i] = 1.0 / float64(n)
	}

	// Run PowerIteration on custom implementation
	x := make(Vector, n)
	copy(x, start)

	err = adj.PowerIteration(&x, 0.1, iterations)
	if err != nil {
		t.Fatalf("PowerIteration failed: %v", err)
	}

	// Build Gonum Dense Matrix
	gonumMat := mat.NewDense(n, n, nil)
	for row := 0; row < n; row++ {
		cols, _ := adj.GetRow(uint32(row))
		for _, col := range cols {
			gonumMat.Set(row, int(col), 1.0)
		}
	}

	//  Run Gonum power iteration
	v := mat.NewVecDense(n, nil)
	for i := 0; i < n; i++ {
		v.SetVec(i, 1.0/float64(n))
	}
	tmp := mat.NewVecDense(n, nil)
	for i := 0; i < iterations; i++ {
		tmp.MulVec(gonumMat, v)
		norm := mat.Norm(tmp, 1)
		tmp.ScaleVec(1/norm, tmp)
		v.CopyVec(tmp)
	}

	// Compare result vectors
	for i := 0; i < n; i++ {
		got := x[i]
		want := v.AtVec(i)
		if math.Abs(got-want) > tol {
			t.Errorf("Mismatch at %d: got %v, want %v", i, got, want)
		}
	}
}

// BenchmarkAdjMatSparse_PowerIteration measures run-time and allocations
// of (*AdjMatSparse).PowerIteration on a synthetic sparse graph whose
// adjacency matrix is constructed via SortIJ + NewFromSortedIJ.
func BenchmarkAdjMatSparse_PowerIteration(b *testing.B) {
	source := rand.New(rand.NewSource(42))

	runBenchmark := func(b *testing.B, matrixSize int, avgDegree int, iterations int) {
		// 1. Generate random edge list (src, dst)
		nEdges := matrixSize * avgDegree
		src := make([]uint32, nEdges)
		dst := make([]uint32, nEdges)
		for i := 0; i < nEdges; i++ {
			src[i] = uint32(source.Intn(matrixSize))
			dst[i] = uint32(source.Intn(matrixSize))
		}

		// 2. Sort and deduplicate
		if err := SortIJ(&src, &dst); err != nil {
			b.Fatalf("SortIJ failed: %v", err)
		}

		// 3. Build sparse matrix
		A, err := NewFromSortedIJ(src, dst)
		if err != nil {
			b.Fatalf("NewFromSortedIJ failed: %v", err)
		}

		// 4. Prepare base vector
		baseVec := make(Vector, matrixSize)
		val := 1.0 / float64(matrixSize)
		for i := range baseVec {
			baseVec[i] = val
		}

		// 5. Run benchmark
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			v := make(Vector, len(baseVec))
			copy(v, baseVec)

			if err := (&A).PowerIteration(&v, .1, iterations); err != nil {
				b.Fatalf("PowerIteration failed: %v", err)
			}
		}
	}

	h := func(b *testing.B, matrixSize int, avgDegree int, iterations int) {
		b.Run(fmt.Sprintf("n=%d_deg=%d_i=%d", matrixSize, avgDegree, iterations), func(b *testing.B) {
			runBenchmark(b, 1000, 10, 20)
		})
	}
	h(b, 1000, 4, 200)
	h(b, 1000, 10, 200)
	h(b, 10000, 4, 200)
	h(b, 10000, 10, 200)
	h(b, 100000, 10, 200)
}

func TestTopKCentralities(t *testing.T) {
	v := Vector{0.12, 2.34, 5.67, 1.23, 3.45}
	top3 := v.TopKCentralities(3)

	expected := Vector{2.34, 3.45, 5.67}
	if !reflect.DeepEqual(top3, expected) {
		t.Errorf("TopKCentralities got %v; want %v", top3, expected)
	}
}
