package graphlib

import (
	"reflect"
	"testing"
)

// TestAllPairsShortestPaths_Disconnected compares the improved all-pairs
// implementations with the BFS-based naive one on a disconnected graph. The
// improved variants derive the row of a low-degree vertex from its neighbours'
// rows; unreachable vertices used to get INF+1 there instead of INF.
func TestAllPairsShortestPaths_Disconnected(t *testing.T) {
	// component A: K4 on {0,1,2,3} plus the pendant vertex 4 at 0,
	// component B: K4 on {5,6,7,8}
	l, err := InitializeAdjList(9)
	if err != nil {
		t.Fatal(err)
	}
	for _, k4 := range [][]int{{0, 1, 2, 3}, {5, 6, 7, 8}} {
		for i := 0; i < len(k4); i++ {
			for j := i + 1; j < len(k4); j++ {
				if err := l.AddEdge(Edge{Vertex{k4[i]}, Vertex{k4[j]}}); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if err := l.AddEdge(Edge{Vertex{0}, Vertex{4}}); err != nil {
		t.Fatal(err)
	}

	want := l.NaiveAllPairsShortestPath()
	if want[4][5] != INF {
		t.Fatalf("naive APSP: distance 4->5 = %d, want INF", want[4][5])
	}

	improved, err := l.ImprovedAllPairsShortestPaths()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(improved, want) {
		t.Errorf("ImprovedAllPairsShortestPaths differs from the naive result\n  got : %v\n  want: %v", improved, want)
	}

	multi, err := l.MultiThreadedImprovedAllPairsShortestPaths(4)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(multi, want) {
		t.Errorf("MultiThreadedImprovedAllPairsShortestPaths differs from the naive result\n  got : %v\n  want: %v", multi, want)
	}
}

// TestDiameterLeavesGraphUnchanged checks that the diameter functions do not
// modify their input. NeighbourhoodSubsets, which ImprovedDiameter uses, used
// to add a self-loop to every vertex.
func TestDiameterLeavesGraphUnchanged(t *testing.T) {
	l, err := ReadGraphAdjList("../exampleGraphs/out.dimacs10-netscie.graph", false)
	if err != nil {
		t.Fatal(err)
	}
	before := l.Copy()

	if _, err := l.NeighbourhoodSubsets(); err != nil {
		t.Fatal(err)
	}
	l.ImprovedDiameter()

	if !reflect.DeepEqual(l.List, before.List) {
		for v := range l.List {
			if !reflect.DeepEqual(l.List[v], before.List[v]) {
				t.Fatalf("the neighbours of vertex %d changed from %v to %v", v, before.List[v], l.List[v])
			}
		}
	}
}
