package graphlib

import (
	"fmt"
	"math"
	"math/rand"
)

func GenerateRandomGraph(n int, rho float64, loops, directed bool, seed int64) ([][]float64, []Edge, error) {
	if n <= 0 {
		return nil, nil, fmt.Errorf("n must be > 0")
	}
	if rho < 0 || rho > 1 {
		return nil, nil, fmt.Errorf("rho must be in [0,1]")
	}

	// ------------------------------------------------------------------ //
	// 1.  How many edges do we need for the requested density?
	mFloat := rho * float64(n) * float64(n-1)
	if !directed {
		mFloat /= 2
	}
	if loops {
		mFloat += float64(n)
	}

	r := rand.New(rand.NewSource(seed))
	var m int
	if r.Intn(2) == 0 {
		m = int(math.Ceil(mFloat))
	} else {
		m = int(math.Floor(mFloat))
	}

	// ------------------------------------------------------------------ //
	// 2.  Enumerate every legal edge (or arc).
	type pair struct{ x, y int }
	possible := make([]pair, 0)

	for x := 0; x < n; x++ {
		for y := 0; y < n; y++ {
			switch {
			case directed && loops:
				possible = append(possible, pair{x, y})
			case directed && !loops:
				if x != y {
					possible = append(possible, pair{x, y})
				}
			case !directed && loops:
				if x <= y { // store upper‑triangle incl. diagonal
					possible = append(possible, pair{x, y})
				}
			case !directed && !loops:
				if x < y { // strict upper‑triangle
					possible = append(possible, pair{x, y})
				}
			}
		}
	}

	if m > len(possible) {
		// This can only happen with an impossible density request.
		m = len(possible)
	}

	// ------------------------------------------------------------------ //
	// 3.  Sample m distinct edges without replacement.
	rand.Shuffle(len(possible), func(i, j int) { possible[i], possible[j] = possible[j], possible[i] })
	chosen := possible[:m]

	// ------------------------------------------------------------------ //
	// 4.  Build adjacency matrix.
	A := make([][]float64, n)
	for i := range A {
		A[i] = make([]float64, n)
	}

	for _, e := range chosen {
		A[e.x][e.y] = 1
		if !directed {
			A[e.y][e.x] = 1
		}
	}

	// Convert to public Edge slice so callers don’t depend on internal pair.
	edges := make([]Edge, len(chosen))
	for i, p := range chosen {
		edges[i] = Edge{Vertex{p.x}, Vertex{p.y}}
	}

	return A, edges, nil
}
