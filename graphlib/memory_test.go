package graphlib

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
)

// ----------------------------
// Allocation‑aware benchmarks
// ----------------------------

// numberOfEdges controls approximately how many undirected edges each vertex will
// receive in the benchmark graphs. A density of 4 means every vertex obtains
// ~4 undirected edges in total (8 *directed* insertions).
const numberOfEdges = 36

// populateGraph adds a repeatable pseudo‑random set of edges with the requested
// density to the graph described by addEdge(u, v).
func populateGraph(size int, addEdge func(Edge) error) {
	r := rand.New(rand.NewSource(42)) // deterministic
	edgeInsertions := size * numberOfEdges
	for i := 0; i < edgeInsertions; i++ {
		u := r.Intn(size)
		v := r.Intn(size)
		if u != v {
			err := addEdge(Edge{Vertex{u}, Vertex{v}})
			if err != nil {
				return
			}
		}
	}
}

func benchAdjMatrixAlloc(b *testing.B, size int) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m, _ := InitializeAdjMat(size)
		populateGraph(size, m.AddEdge)
	}
}

func benchAdjListAlloc(b *testing.B, size int) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l, _ := InitializeAdjList(size)
		populateGraph(size, l.AddEdge)
	}
}

func BenchmarkGraphAllocations(b *testing.B) {
	sizes := []int{128, 512, 2048, 8192}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("AdjMatDense/%d", n), func(b *testing.B) {
			benchAdjMatrixAlloc(b, n)
		})
		b.Run(fmt.Sprintf("AdjList/%d", n), func(b *testing.B) {
			benchAdjListAlloc(b, n)
		})
	}
}

// ----------------------------
// Peak heap usage snapshots
// ----------------------------

// The helpers below demonstrate one possible pattern for recording the peak
// heap consumption of different graph representations without leaving the tight
// inner loops of the actual benchmarks.
//
// They are *not* intended to replace the official ‑memprofile flag but act as
// quick‑turnaround diagnostics when you iterate on the internal data layouts.

func capturePeakHeap(f func()) uint64 {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	f()
	runtime.GC()
	runtime.ReadMemStats(&after)
	return after.HeapAlloc - before.HeapAlloc
}

func TestPeakHeapUsage(t *testing.T) {
	const size = 10_000

	peakAdjMat := capturePeakHeap(func() {
		m, _ := InitializeAdjMat(size)
		populateGraph(size, m.AddEdge)
	})

	peakAdjList := capturePeakHeap(func() {
		l, _ := InitializeAdjList(size)
		populateGraph(size, l.AddEdge)
	})

	t.Logf("Peak heap usage @%d vertices —  AdjMatDense: %d KB | AdjList: %d KB",
		size,
		peakAdjMat/1024, peakAdjList/1024,
	)
}

// ----------------------------
// File read benchmark
// ----------------------------

func BenchmarkReadGraph_EDGELIST_undirected(b *testing.B) {
	files := []string{
		"tiny_01.graph",
		"Wiki-Vote.txt",
	}

	for _, file := range files {
		b.Run(fmt.Sprintf("EDGELIST/%s", file), func(b *testing.B) {
			benchReadGraph(b, file)
		})
	}
}

func BenchmarkReadGraph_METIS_undirected(b *testing.B) {
	files := []string{
		"metis_tiny_01.mgraph",
		"metis_test.mgraph",
	}

	for _, file := range files {
		b.Run(fmt.Sprintf("METIS/%s", file), func(b *testing.B) {
			benchReadGraph(b, file)
		})
	}
}

func benchReadGraph(b *testing.B, file string) {
	b.ReportAllocs()
	_, _ = ReadGraphAdjList("../exampleGraphs/"+file, false)
}
