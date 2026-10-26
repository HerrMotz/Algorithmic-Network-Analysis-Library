package graphlib

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// ---------------------------------------
//  Tests reading from edgelist and METIS
// ---------------------------------------

var wantAdj = []VertexSet{
	{1: Naught{}, 2: Naught{}, 4: Naught{}},
	{0: Naught{}, 2: Naught{}, 3: Naught{}},
	{4: Naught{}, 3: Naught{}, 1: Naught{}, 0: Naught{}},
	{1: Naught{}, 2: Naught{}, 5: Naught{}, 6: Naught{}},
	{0: Naught{}, 2: Naught{}, 5: Naught{}},
	{4: Naught{}, 3: Naught{}, 6: Naught{}},
	{5: Naught{}, 3: Naught{}},
}

var wantAdjAsMETIS = []byte(`7 22
 2 3 5
 1 3 4
 1 2 4 5
 2 3 6 7
 1 3 6
 4 5 7
 4 6
`)

var wantAdjAsPlainEdgeList = []byte(`1 2
1 3
1 5
2 1
2 3
2 4
3 1
3 2
3 4
3 5
4 2
4 3
4 6
4 7
5 1
5 3
5 6
6 4
6 5
6 7
7 4
7 6
`)

// -------
// tests
// -------

func TestReadGraph_EdgeList(t *testing.T) {
	adj, err := ReadGraphAdjList("../exampleGraphs/tiny_01.graph", false)
	if err != nil && !errors.Is(err, EdgeOrArcExistsError) {
		t.Fatalf("ReadGraph_AdjList(edge list) returned error: %v", err)
	}
	checkAdjList(t, adj)
}

func TestReadGraph_METIS(t *testing.T) {
	adj, err := ReadGraphAdjList("../exampleGraphs/metis_tiny_01.mgraph", false)
	if err != nil && !errors.Is(err, EdgeOrArcExistsError) {
		t.Fatalf("ReadGraph_AdjList(METIS) returned error: %v", err)
	}
	checkAdjList(t, adj)
}

func TestReadGraph_MatrixMarket(t *testing.T) {
	g, err := ReadGraphAdjList("../exampleGraphs/soc-karate.mtx", false)
	if err != nil {
		t.Fatalf("ReadGraphAdjList(Matrix Market) returned error: %v", err)
	}
	if g.Size != 34 {
		t.Fatalf("size = %d, want 34", g.Size)
	}
	if m := g.TotalDegree() / 2; m != 78 {
		t.Errorf("edges = %d, want 78", m)
	}
	for v := 0; v < g.Size; v++ {
		if g.HasArc(Arc{Vertex{v}, Vertex{v}}) {
			t.Errorf("unexpected self-loop at vertex %d", v)
		}
	}
	// vertex 34 of Zachary's karate club (index 33) has 17 neighbours; the
	// size line "34 34 78" used to be read as an extra self-loop
	if d, _ := g.Degree(Vertex{33}); d != 17 {
		t.Errorf("degree of vertex 33 = %d, want 17", d)
	}
}

func TestReadGraph_MatrixMarketDirected(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	// a general matrix keeps the direction of its entries ...
	general := write("general.mtx", "%%MatrixMarket matrix coordinate pattern general\n% comment\n3 3 2\n1 2\n3 1\n")
	g, err := ReadGraphAdjList(general, true)
	if err != nil {
		t.Fatal(err)
	}
	want := []VertexSet{{1: Naught{}}, {}, {0: Naught{}}}
	if !reflect.DeepEqual(g.List, want) {
		t.Errorf("general matrix: got %v, want %v", g.List, want)
	}

	// ... while a symmetric one stores only one triangle and is mirrored
	symmetric := write("symmetric.mtx", "%%MatrixMarket matrix coordinate real symmetric\n3 3 2\n2 1 0.5\n3 1 1.5\n")
	g, err = ReadGraphAdjList(symmetric, true)
	if err != nil {
		t.Fatal(err)
	}
	want = []VertexSet{{1: Naught{}, 2: Naught{}}, {0: Naught{}}, {0: Naught{}}}
	if !reflect.DeepEqual(g.List, want) {
		t.Errorf("symmetric matrix: got %v, want %v", g.List, want)
	}

	// the size line must match the number of entries
	truncated := write("truncated.mtx", "%%MatrixMarket matrix coordinate pattern symmetric\n3 3 2\n2 1\n")
	if _, err := ReadGraphAdjList(truncated, false); err == nil {
		t.Error("expected an error for a file with fewer entries than announced")
	}
}

// ---------
// helpers
// ---------

func checkAdjList(t *testing.T, got *AdjList) {
	t.Helper()

	if got == nil {
		t.Fatalf("got nil")
	}

	if got.Size != len(wantAdj) {
		t.Fatalf("size = %d, want %d", got.Size, len(wantAdj))
	}

	for v, gotNbrs := range got.List {
		wantNbrs := wantAdj[v]

		if !reflect.DeepEqual(gotNbrs, wantNbrs) {
			t.Errorf("neighbours of %d mismatch\n  got : %v\n  want: %v",
				v, gotNbrs, wantNbrs)
		}
	}
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
		"4elt.mgraph",
	}

	for _, file := range files {
		b.Run(fmt.Sprintf("METIS/%s", file), func(b *testing.B) {
			benchReadGraph(b, file)
		})
	}
}

func benchReadGraph(b *testing.B, file string) {
	b.ReportAllocs()
	l, err := ReadGraphAdjList("../exampleGraphs/"+file, false)

	if err != nil {
		b.Fatalf("ReadGraph_AdjList returned error: %v", err)
	}

	if l == nil {
		b.Fatalf("ReadGraph_AdjList returned nil")
	}

	if l.Size == 0 {
		b.Fatalf("ReadGraph_AdjList returned empty graph")
	}

	ok := l.HasVertex(Vertex{0})
	if !ok {
		b.Fatalf("ReadGraph_AdjList returned graph without vertex 0")
	}
}

func TestWriteSizeInFile(t *testing.T) {
	file := "write_test.graph"

	adj := []VertexSet{
		{1: Naught{}},
		{2: Naught{}},
		{3: Naught{}},
	}

	l := AdjList{List: adj, Size: len(adj)}

	heurPlain := l.Size * 4 * 1     // |vertices| × bytes per neighbour × average node degree
	heurMETIS := (l.Size+1)*3*1 + 1 // todo: Think of a good heuristic for METIS

	written, err := l.WriteToEdgeList(file, heurPlain)
	checkWriteHeuristic(t, "WriteToEdgeList", file, heurPlain, written, err)

	written, err = l.WriteToMETIS(file, heurMETIS)
	checkWriteHeuristic(t, "WriteToMETIS", file, heurMETIS, written, err)
}

func checkWriteHeuristic(t *testing.T, method string, file string, heur int, written int, err error) {
	if err != nil {
		t.Fatalf("%v(%v) failed: %v", method, file, err)
	}

	if heur != written { // todo maybe it is sufficient to check <=
		t.Fatalf("%v(%v) bad heuristic. expected: %v, actually wrote %v", method, file, heur, written)
	}
}

func TestWriteAdjListToEdgeList(t *testing.T) {
	file := "write_test.graph"
	l := AdjList{List: wantAdj, Size: len(wantAdj)}
	heur := l.Size * 10
	written, err := l.WriteToEdgeList(file, heur)

	if err != nil {
		t.Fatalf("WriteToEdgeList(%v) failed: %v", file, err)
	}

	t.Logf("WriteToEdgeList(%v) wrote %v bytes, heuristic was %v bytes", file, written, heur)

	// read the file and validate its contents
	wantSum := sha256.Sum256(wantAdjAsPlainEdgeList)

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("os.Open(%v) failed: %v", file, err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatalf("hash error: %v", err)
	}
	gotSum := h.Sum(nil)

	if !bytes.Equal(gotSum, wantSum[:]) {
		t.Fatalf("contents differ (by hash).\nexpected %v,\ngot      %v", wantSum, gotSum)
	}
}

func TestWriteAdjListToMETIS(t *testing.T) {
	file := "write_test.mgraph"
	l := AdjList{List: wantAdj, Size: len(wantAdj)}
	heur := l.Size * 10
	written, err := l.WriteToMETIS(file, heur)
	if err != nil {
		t.Fatalf("WriteToMETIS(%v) failed: %v", file, err)
	}

	t.Logf("WriteToMetis(%v) wrote %v bytes, heuristic was %v bytes", file, written, heur)

	// read the file and validate its contents
	wantSum := sha256.Sum256(wantAdjAsMETIS)

	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("os.Open(%v) failed: %v", file, err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatalf("hash error: %v", err)
	}
	gotSum := h.Sum(nil)

	if !bytes.Equal(gotSum, wantSum[:]) {
		t.Fatalf("contents differ (by hash).\nexpected %v,\ngot      %v", wantSum, gotSum)
	}
}
