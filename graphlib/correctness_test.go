package graphlib

import (
	"testing"
)

func TestLookupTable(t *testing.T) {
	names := []string{"berlin", "hamburg", "munich", "cologne"}
	lut, _ := InitializeLookupTable(len(names), names)
	if idx := lut["munich"]; idx != 2 {
		t.Errorf("expect munich → 2, got %d", idx)
	}
}

func TestAdjMatrix_HasEdge(t *testing.T) {
	m, err := InitializeAdjMat(4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = m.AddEdge(Edge{Vertex{0}, Vertex{1}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = m.AddEdge(Edge{Vertex{1}, Vertex{2}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !m.HasEdge(Edge{Vertex{0}, Vertex{1}}) {
		t.Errorf("expected edge (0,1) to exist")
	}
	if m.HasEdge(Edge{Vertex{2}, Vertex{3}}) {
		t.Errorf("edge (2,3) should not exist")
	}
}

// ----------------------------
// Correctness checks For Adjacency List
// ----------------------------

func TestAdjList_InternalConsistency(t *testing.T) {
	adj, _ := InitializeAdjList(4)
	if err := adj.AddEdge(Edge{Vertex{1}, Vertex{3}}); err != nil {
		return
	}
	if len(adj.List[1]) != 1 || len(adj.List[3]) != 1 {
		t.Fatalf("edge (1,3) not inserted symmetrically")
	}
}

func TestAdjList_HasArc(t *testing.T) {
	adj, _ := InitializeAdjList(4)
	if err := adj.AddArc(Arc{Vertex{1}, Vertex{3}}); err != nil {
		return
	}
	if !adj.HasArc(Arc{Vertex{1}, Vertex{3}}) {
		t.Errorf("expected arc (1,3) to exist")
	}
	if adj.HasArc(Arc{Vertex{2}, Vertex{3}}) {
		t.Errorf("arc (2,3) should not exist")
	}
}

func TestAdjList_HasEdge(t *testing.T) {
	adj, _ := InitializeAdjList(4)
	if err := adj.AddArc(Arc{Vertex{1}, Vertex{3}}); err != nil {
		return
	}
	if err := adj.AddEdge(Edge{Vertex{1}, Vertex{3}}); err != nil {
		return
	}
	if !adj.HasEdge(Edge{Vertex{1}, Vertex{3}}) {
		t.Errorf("expected edge (1,3) to exist")
	}
}

func TestAdjList_Neighbours(t *testing.T) {
	adj, _ := InitializeAdjList(4)
	// add some edges and test
	if err := adj.AddEdge(Edge{Vertex{1}, Vertex{3}}); err != nil {
	}
	if err := adj.AddEdge(Edge{Vertex{1}, Vertex{2}}); err != nil {
	}
	if err := adj.AddEdge(Edge{Vertex{2}, Vertex{3}}); err != nil {
	}

	// test neighbours
	neighbours, err := adj.Neighbours(Vertex{1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(neighbours) != 2 {
		t.Fatalf("expected 2 neighbours, got %d", len(neighbours))
	}
	_, ok1 := neighbours[2]
	_, ok2 := neighbours[3]
	if !ok1 || !ok2 {
		t.Fatalf("expected neighbours to be [2,3], got %v", neighbours)
	}
}

func TestShortestPaths_DAG(t *testing.T) {
	adj, _ := InitializeAdjList(4)

	zero := Vertex{0}
	one := Vertex{1}
	two := Vertex{2}
	three := Vertex{3}

	if err := adj.AddEdge(Edge{zero, one}); err != nil {
		return
	}
	if err := adj.AddEdge(Edge{zero, two}); err != nil {
		return
	}
	if err := adj.AddEdge(Edge{one, three}); err != nil {
		return
	}
	if err := adj.AddEdge(Edge{two, three}); err != nil {
		return
	}

	parents, distances := adj.directedAcyclicGraph(Vertex{0})
	if len(parents) != 4 || len(distances) != 4 {
		t.Fatalf("expected 4 parents and distances, got %d and %d", len(parents), len(distances))
	}

	if parents[3] != one {
		t.Fatalf("expected parent of 3 to be 1, got %v", parents[3])
	}

	if distances[3] != 2 {
		t.Fatalf("expected distance to 3 to be 2, got %d", distances[3])
	}

	if parents[2] != zero {
		t.Fatalf("expected parent of 2 to be 0, got %v", parents[2])
	}

	if distances[2] != 1 {
		t.Fatalf("expected distance to 2 to be 1, got %d", distances[2])
	}
}
