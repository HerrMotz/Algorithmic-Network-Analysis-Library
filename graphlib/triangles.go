package graphlib

import (
    "sort"
)

// Triangles ÜS2T3
// TrianglesNodeIterator finds all triangles in the graph using a node iterator.
func (l *AdjList) TrianglesNodeIterator() ([][3]Vertex, error) {
    n := l.Size
    if n < 3 {
        return nil, nil
    }

    // 1) count degrees
    deg := make([]int, n)
    for u := 0; u < n; u++ {
        deg[u] = len(l.List[u])
    }

    // 2) build forward adjacency list
    fwd := make([][]int, n)
    for u := 0; u < n; u++ {
        for v := range l.List[u] {
            if deg[u] < deg[v] || (deg[u] == deg[v] && u < v) {
                fwd[u] = append(fwd[u], v)
            }
        }
        sort.Ints(fwd[u])
    }

    // 3) merge 
    var tris [][3]Vertex
    for u := 0; u < n; u++ {
        for _, v := range fwd[u] {
            a, b := fwd[u], fwd[v]
            i, j := 0, 0
            for i < len(a) && j < len(b) {
                switch {
                case a[i] < b[j]:
                    i++
                case a[i] > b[j]:
                    j++
                default:
                    w := a[i]
                    tris = append(tris, [3]Vertex{
                        {Id: u}, {Id: v}, {Id: w},
                    })
                    i++; j++
                }
            }
        }
    }

    return tris, nil
}

// TrianglesCNI lists all triangles via Chiba&Nishizeki.
// returns triples (u,v,w) with r[u]<r[v]<r[w].
func (l *AdjList) TrianglesCNI() ([][3]Vertex, error) {
    n := l.Size
    if n < 3 {
        return nil, nil
    }

    // 1) get Degeneracy-Ordering
    ordering, err := l.DegeneracyOrder()
    if err != nil {
        return nil, err
    }

    // get rank r[v] = pos of v in the ordering
    rank := make([]int, n)
    for i, v := range ordering {
        rank[v.Id] = i
    }
    out := make([][]int, n)
    for u := 0; u < n; u++ {
        for v := range l.List[u] {
            if rank[u] < rank[v] {
                out[u] = append(out[u], v)
            }
        }
    }

    // 3) mark them and find triples
    tris := make([][3]Vertex, 0)
    mark := make([]bool, n)

	// for u = v1, . . . , vn−2 do {
    for u := 0; u < n; u++ {
        // for each v ∈ N (u) do mark v;
        for _, w := range out[u] {
            mark[w] = true
        }

        // for each v ∈ N (u) do {
        for _, v := range out[u] {
			// for each w ∈ N (v) do {
            for _, w := range out[v] {
				// if w is marked then {
                if mark[w] {
					// output triangle {u, v, w};
                    tris = append(tris, [3]Vertex{
                        {Id: u}, {Id: v}, {Id: w},
                    })
                }
            }
        }

        // unmark v;
        for _, w := range out[u] {
            mark[w] = false
        }
    }

    return tris, nil
}
