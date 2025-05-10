// graphlib_memory_test.go
//
// Unit and benchmark tests for the graphlib package.  Run
//
//     go test -run ^$ -bench . -benchmem -memprofile mem.out -cpuprofile cpu.out
//
// to obtain detailed allocation, heap and CPU profiles.

package graphlib

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
)

// ----------------------------
// Correctness checks
// ----------------------------

func TestAdjMatrix_HasEdge(t *testing.T) {
	m, err := InitializeAdjMat(4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m.AddEdge(Edge{0, 1})
	m.AddEdge(Edge{1, 2})

	if !m.HasEdge(Edge{0, 1}) {
		t.Errorf("expected edge (0,1) to exist")
	}
	if m.HasEdge(Edge{2, 3}) {
		t.Errorf("edge (2,3) should not exist")
	}
}

func TestAdjList_InternalConsistency(t *testing.T) {
	adj, _ := InitializeAdjList(4)
	adj.AddEdge(Edge{1, 3})
	if len(adj.List[1]) != 1 || len(adj.List[3]) != 1 {
		t.Fatalf("edge (1,3) not inserted symmetrically")
	}
}

func TestLookupTable(t *testing.T) {
	names := []string{"berlin", "hamburg", "munich", "cologne"}
	lut, _ := InitializeLookupTable(len(names), names)
	if idx := lut["munich"]; idx != 2 {
		t.Errorf("expect munich → 2, got %d", idx)
	}
}

// ----------------------------
// Allocation‑aware benchmarks
// ----------------------------

// numberOfEdges controls approximately how many undirected edges each vertex will
// receive in the benchmark graphs. A density of 4 means every vertex obtains
// ~4 undirected edges in total (8 *directed* insertions).
const numberOfEdges = 36

// populateGraph adds a repeatable pseudo‑random set of edges with the requested
// density to the graph described by addEdge(u, v).
func populateGraph(size int, addEdge func(Edge)) {
	r := rand.New(rand.NewSource(42)) // deterministic
	edgeInsertions := size * numberOfEdges
	for i := 0; i < edgeInsertions; i++ {
		u := r.Intn(size)
		v := r.Intn(size)
		if u != v {
			addEdge(Edge{u, v})
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
