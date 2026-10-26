package graphlib

import (
	"fmt"
	"math"
	"math/rand"
)

// GenerateRandomGraph builds a random graph of n vertices with density rho.
// loops controls whether self‐edges are allowed, directed controls if the graph is directed.
// Returns an initialized AdjList and the list of edges/arcs added.
func GenerateRandomGraph(n int, rho float64, loops, directed bool, seed int64) (*AdjList, []Edge, error) {
	// 1. Validate
	if n <= 0 {
		return nil, nil, fmt.Errorf("n must be > 0")
	}
	if rho < 0 || rho > 1 {
		return nil, nil, fmt.Errorf("rho must be in [0,1]")
	}

	// 2. Compute target number of edges/arcs
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

	// 3. Enumerate all legal pairs
	type pair struct{ x, y int }
	possible := make([]pair, 0, n*n)
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
				if x <= y {
					possible = append(possible, pair{x, y})
				}
			case !directed && !loops:
				if x < y {
					possible = append(possible, pair{x, y})
				}
			}
		}
	}
	if m > len(possible) {
		m = len(possible)
	}

	r.Shuffle(len(possible), func(i, j int) { possible[i], possible[j] = possible[j], possible[i] })
	chosen := possible[:m]

	// 4. Initialize the adjacency list
	adjList, err := InitializeAdjList(n)
	if err != nil {
		return nil, nil, err
	}

	edges := make([]Edge, 0, m)
	for _, p := range chosen {
		u := Vertex{p.x}
		v := Vertex{p.y}
		if directed {
			arc := Arc{u, v}
			if !adjList.HasArc(arc) {
				if err := adjList.AddArc(arc); err != nil {
					return nil, nil, err
				}
			}
			edges = append(edges, Edge{u, v})
		} else {
			edge := Edge{u, v}
			if !adjList.HasEdge(edge) {
				if err := adjList.AddEdge(edge); err != nil {
					return nil, nil, err
				}
			}
			edges = append(edges, edge)
		}
	}

	return adjList, edges, nil
}
