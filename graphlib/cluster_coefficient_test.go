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
		// possible = deg*(deg-1)/2
		// connected = number of edges between neighbours
		// formula = connected / possible
		
        {Vertex{0}, 0.38181818181818183},   // deg=11, connected=20, possible=55
        {Vertex{2}, 0.6071428571428571},    // deg=8,  connected=17, possible=28
        {Vertex{6}, 0.6666666666666666},    // deg=6,  connected=10, possible=15
        {Vertex{14}, 0.5},                  // deg=4,  connected=3,  possible=6
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
    // Insgesamt T=49 Dreiecke; sum_v deg(v)*(deg(v)-1) = 627
    // c(G) = 6*49/627 ≈ 0.4696485623003195
    want := 0.4696485623003195
    if math.Abs(got-want) > 1e-9 {
        t.Errorf("GlobalClustering = %v; want %v", got, want)
    }
}
