package graphlib

import (
	"slices"
)

func (l *AdjList) NaiveDiameter() int {
	diameter := 0

	for i := 0; i < l.Size; i++ {
		distances, _ := l.SingleSourceShortestPath(Vertex{i})
		_n := slices.Max(distances)
		if _n == INF { // if there is more than one connected component
			return INF
		}
		diameter = max(diameter, _n)

	}

	return diameter
}

// NeighbourhoodSubsets returns a set of vertices u such that there exists another
// vertex v in the graph whose (open or closed) neighborhood is a subset of the
// (open or closed) neighborhood of u.
//
// – Open neighbourhood   N(u)  := { w | {u,w} ∈ E }
// – Closed neighbourhood N[u] := N(u) ∪ {u}
func (l *AdjList) NeighbourhoodSubsets() (VertexSet, error) {
	// Helper for the subset test
	isSubset := func(sub, sup VertexSet) bool {
		for x := range sub {
			if _, ok := sup[x]; !ok {
				return false
			}
		}
		return true
	}

	// Main search
	result := make(VertexSet)

	for u := 0; u < l.Size; u++ {
		// --- collect candidates ( neighbours of u OR of any neighbour of u ) --
		candidates := make(VertexSet)

		openU := make(VertexSet, 8) // todo: chance to add a heuristic
		openU, _ = l.Neighbours(Vertex{u})

		// we only need closedU, because the comparison works with only closedU
		closedU := make(VertexSet, len(openU)+1)
		closedU = openU
		closedU[u] = Naught{}

		for v := range openU {
			// only reference passing
			openV := make(VertexSet, 8) // todo: chance to add a heuristic
			openV, _ = l.Neighbours(Vertex{v})

			candidates[v] = Naught{}
			for w := range openV {
				candidates[w] = Naught{}
			}
		}
		delete(candidates, u) // we only want "another" vertex

		// --- check each candidate --------------------------------------------
		for v := range candidates {
			// only reference passing
			openV := make(VertexSet, 8) // todo: chance to add a heuristic
			openV, _ = l.Neighbours(Vertex{v})

			if _, already := result[v]; already {
				// mimic the Python early‑out: if v is already in the result set
				// (because N(v)=N(u) etc.) do not put both u and v into result.
				continue
			}

			if isSubset(openV, openU) || isSubset(openV, closedU) {
				// if isSubset(openV, openU) ||  {
				result[u] = Naught{}
				break // proceed to next u
			}
		}
	}

	return result, nil
}

// ImprovedDiameter returns the diameter.
//
// ‑ If the graph is disconnected the function returns math.MaxInt.
// ‑ The algorithm follows the Python reference: it skips BFS roots that are
//
//	articulation points or have duplicate neighbourhoods, and tightens lower /
//	upper bounds on‑the‑fly (  d ≤ diam ≤ 2·d  ).
func (l *AdjList) ImprovedDiameter() int {
	// 1.  Determine vertices we do *not* start a BFS from
	cuts, _ := l.CutVertices(l.Size)
	dupNbhd, _ := l.NeighbourhoodSubsets()

	// 2.  BFS scans and on‑the‑fly bounds update
	lower := 0
	upper := INF // “∞”
	for v := range l.List {
		// do it in exactly this order,
		// because probably dupNbhd is larger than cuts
		_, bad1 := dupNbhd[v]
		_, bad2 := cuts[v]
		// start := all - skip
		if !bad1 && !bad2 {
			maxDist, err := l.bfsMaxDistance(v)
			if err != nil {
				return 0
			}
			if maxDist == INF { // disconnected
				return INF
			}

			if maxDist > lower {
				lower = maxDist
			}
			if d2 := maxDist * 2; d2 < upper {
				upper = d2
			}
			if upper == lower {
				return lower // tight bound reached
			}
		}
	}

	return lower // lower ≤ diameter ≤ upper, but both equal => diameter
}

// bfsMaxDistance returns the greatest shortest‑path distance from start to any
// vertex reachable in the component. If some vertex exists in the graph but
// is unreachable from start, math.MaxInt is returned to signal “∞”.
func (l *AdjList) bfsMaxDistance(start int) (int, error) {
	visited := make([]bool, l.Size)
	dist := make([]int, l.Size)
	for i := range dist {
		dist[i] = -1
	}

	q := []int{start}
	visited[start] = true
	dist[start] = 0
	maxD := 0

	for head := 0; head < len(q); head++ {
		u := q[head]

		neigh, err := l.Neighbours(Vertex{u})
		if err != nil {
			return 0, err
		}

		for v := range neigh {
			if !visited[v] {
				visited[v] = true
				dist[v] = dist[u] + 1
				if dist[v] > maxD {
					maxD = dist[v]
				}
				q = append(q, v)
			}
		}
	}

	// any existing vertex not reached → disconnected component
	for v := 0; v < l.Size; v++ {
		if l.HasVertex(Vertex{v}) && !visited[v] {
			return INF, nil
		}
	}
	return maxD, nil
}
