package graphlib

import (
	"errors"
	"testing"
)

func TestAdjList_RemoveArcAndEdge(t *testing.T) {
	l, err := InitializeAdjList(3)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range []Edge{{Vertex{0}, Vertex{1}}, {Vertex{1}, Vertex{2}}} {
		if err := l.AddEdge(e); err != nil {
			t.Fatal(err)
		}
	}

	// RemoveArc removes exactly one direction ...
	if err := l.RemoveArc(Arc{Vertex{0}, Vertex{1}}); err != nil {
		t.Fatalf("RemoveArc: %v", err)
	}
	if l.HasArc(Arc{Vertex{0}, Vertex{1}}) || !l.HasArc(Arc{Vertex{1}, Vertex{0}}) {
		t.Error("RemoveArc(0,1) must remove 0->1 and keep 1->0")
	}
	if !l.HasEdge(Edge{Vertex{1}, Vertex{2}}) {
		t.Error("RemoveArc(0,1) must not touch other edges")
	}
	// ... and reports missing arcs
	if err := l.RemoveArc(Arc{Vertex{0}, Vertex{1}}); !errors.Is(err, EdgeOrArcNotFoundError) {
		t.Errorf("RemoveArc on a missing arc: got %v, want EdgeOrArcNotFoundError", err)
	}

	if err := l.RemoveEdge(Edge{Vertex{2}, Vertex{1}}); err != nil {
		t.Fatalf("RemoveEdge: %v", err)
	}
	if l.HasArc(Arc{Vertex{1}, Vertex{2}}) || l.HasArc(Arc{Vertex{2}, Vertex{1}}) {
		t.Error("RemoveEdge must remove both directions")
	}
	if err := l.RemoveEdge(Edge{Vertex{2}, Vertex{1}}); !errors.Is(err, EdgeOrArcNotFoundError) {
		t.Errorf("RemoveEdge on a missing edge: got %v, want EdgeOrArcNotFoundError", err)
	}
	if err := l.RemoveEdge(Edge{Vertex{0}, Vertex{7}}); err == nil {
		t.Error("RemoveEdge with an unknown vertex must fail")
	}
}

func TestWeightedAdjList_RemoveEdge(t *testing.T) {
	g, err := InitializeWeightedAdjList(2)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AddEdge(Edge{Vertex{0}, Vertex{1}}, 2.5); err != nil {
		t.Fatal(err)
	}
	if err := g.RemoveEdge(Edge{Vertex{1}, Vertex{0}}); err != nil {
		t.Fatalf("RemoveEdge: %v", err)
	}
	if err := g.RemoveEdge(Edge{Vertex{1}, Vertex{0}}); !errors.Is(err, EdgeOrArcNotFoundError) {
		t.Errorf("RemoveEdge on a missing edge: got %v, want EdgeOrArcNotFoundError", err)
	}
}
