package graphlib

import (
    "testing"
    "math"
)

// (Assuming buildTestFromList is defined as in the previous tests)

func TestLocalClustering(t *testing.T) {
    g := buildTestFromList(t)

    // Testfälle: Vertex → erwarteter lokaler Clustering-Koeffizient
    cases := []struct {
        v    Vertex
        want float64
    }{
        {Vertex{0}, 0.4},
        {Vertex{2}, 0.6190476190476191},
        {Vertex{6}, 0.6666666666666666},
        {Vertex{14}, 0.6666666666666666},
    }

    for _, c := range cases {
        got, err := g.LocalClustering(c.v)
        if err != nil {
            t.Errorf("LocalClustering(%d) returned error: %v", c.v.Id, err)
            continue
        }
        if math.Abs(got-c.want) > 1e-9 {
            t.Errorf("LocalClustering(%d) = %v; want %v", c.v.Id, got, c.want)
        }
    }
}

func TestGlobalClustering(t *testing.T) {
    g := buildTestFromList(t)

    got, err := g.GlobalClustering()
    if err != nil {
        t.Fatalf("GlobalClustering returned error: %v", err)
    }
    want := 0.4789156626506024
    if math.Abs(got-want) > 1e-9 {
        t.Errorf("GlobalClustering = %v; want %v", got, want)
    }
}
