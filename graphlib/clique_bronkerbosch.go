package graphlib

// Bron-Kerbosch ohne Pivot
func (g *AdjList) BronKerboschSimple() [][]int {
    var res [][]int
    var bk func(R, P, X map[int]struct{})
    bk = func(R, P, X map[int]struct{}) {
        if len(P) == 0 && len(X) == 0 {
            clique := make([]int, 0, len(R))
            for v := range R {
                clique = append(clique, v)
            }
            res = append(res, clique)
            return
        }
        // Iteriere über Kopie von P!
        Pcopy := copySet(P)
        for v := range Pcopy {
            R2 := copySet(R)
            R2[v] = struct{}{}
            N_v := g.neighSet(v)
            P2 := intersect(P, N_v)
            X2 := intersect(X, N_v)
            bk(R2, P2, X2)
        }
    }
    P := make(map[int]struct{}, g.Size)
    for v := 0; v < g.Size; v++ {
        P[v] = struct{}{}
    }
    bk(make(map[int]struct{}), P, make(map[int]struct{}))
    return res
}


// Bron-Kerbosch mit Pivot (Tomita-Pivotregel)
func (g *AdjList) BronKerboschPivot() [][]int {
    var res [][]int
    var bk func(R, P, X map[int]struct{})
    bk = func(R, P, X map[int]struct{}) {
        if len(P) == 0 && len(X) == 0 {
            clique := make([]int, 0, len(R))
            for v := range R {
                clique = append(clique, v)
            }
            res = append(res, clique)
            return
        }

        U := union(P, X)
        var pivot, maxNeighbors int
        for u := range U {
            deg := len(intersect(P, g.neighSet(u)))
            if deg > maxNeighbors {
                maxNeighbors, pivot = deg, u
            }
        }

        PwoN := difference(P, g.neighSet(pivot))
        PwoNcopy := copySet(PwoN)
        for v := range PwoNcopy {
            R2 := copySet(R)
            R2[v] = struct{}{}

            N_v := g.neighSet(v)
            P2 := intersect(P, N_v)
            X2 := intersect(X, N_v)

            bk(R2, P2, X2)
        }
    }

    P := make(map[int]struct{}, g.Size)
    for v := 0; v < g.Size; v++ {
        P[v] = struct{}{}
    }
    bk(make(map[int]struct{}), P, make(map[int]struct{}))
    return res
}


// Bron-Kerbosch mit Pivot und Degeneracy Ordering
func (g *AdjList) BronKerboschDegeneracy() [][]int {
    order := g.degeneracyOrder() // Liefert []int
    var res [][]int
    seen := make(map[int]struct{})
    for _, v := range order {
        R := map[int]struct{}{v: {}}
        Nv := g.neighSet(v)
        P := difference(Nv, seen)    // Nur Nachbarn, die noch nicht gesehen wurden
        X := intersect(Nv, seen)     // Bereits gesehene Nachbarn
        subCliques := g.bkPivotDegeneracy(R, P, X)
        res = append(res, subCliques...)
        seen[v] = struct{}{}
    }
    return res
}

func (g *AdjList) bkPivotDegeneracy(R, P, X map[int]struct{}) [][]int {
    var res [][]int
    var bk func(R, P, X map[int]struct{})
    bk = func(R, P, X map[int]struct{}) {
        if len(P) == 0 && len(X) == 0 {
            clique := make([]int, 0, len(R))
            for v := range R {
                clique = append(clique, v)
            }
            res = append(res, clique)
            return
        }
        U := union(P, X)
        var pivot, maxNeighbors int
        for u := range U {
            deg := len(intersect(P, g.neighSet(u)))
            if deg > maxNeighbors {
                maxNeighbors, pivot = deg, u
            }
        }
        PwoN := difference(P, g.neighSet(pivot))
        PwoNcopy := copySet(PwoN)
        for v := range PwoNcopy {
            R2 := copySet(R)
            R2[v] = struct{}{}
            N_v := g.neighSet(v)
            P2 := intersect(P, N_v)
            X2 := intersect(X, N_v)
            bk(R2, P2, X2)
        }
    }
    bk(copySet(R), copySet(P), copySet(X))
    return res
}



// Hilfsfunktionen für Mengenoperationen
func copySet(s map[int]struct{}) map[int]struct{} {
    r := make(map[int]struct{}, len(s))
    for k := range s {
        r[k] = struct{}{}
    }
    return r
}
func intersect(a, b map[int]struct{}) map[int]struct{} {
    r := make(map[int]struct{})
    for k := range a {
        if _, ok := b[k]; ok {
            r[k] = struct{}{}
        }
    }
    return r
}
func difference(a, b map[int]struct{}) map[int]struct{} {
    r := make(map[int]struct{})
    for k := range a {
        if _, ok := b[k]; !ok {
            r[k] = struct{}{}
        }
    }
    return r
}
func union(a, b map[int]struct{}) map[int]struct{} {
    r := make(map[int]struct{}, len(a)+len(b))
    for k := range a {
        r[k] = struct{}{}
    }
    for k := range b {
        r[k] = struct{}{}
    }
    return r
}


// Nachbarn als Menge
func (g *AdjList) neighSet(u int) map[int]struct{} {
	n, _ := g.Neighbours(Vertex{Id: u})
	res := make(map[int]struct{}, len(n))
	for v := range n {
		res[v] = struct{}{}
	}
	return res
}

// Adjazenzliste symmetrieren (falls nötig)
func (l *AdjList) MakeUndirected() {
	for u, neighbours := range l.List {
		for v := range neighbours {
			if l.List[v] == nil {
				l.List[v] = make(VertexSet)
			}
			l.List[v][u] = Naught{}
		}
	}
}

// Degeneracy Order berechnen
func (g *AdjList) degeneracyOrder() []int {
	// degeneracy order ist eine Reihenfolge der Knoten, die aufsteigend nach Grad sortiert ist
	order := make([]int, 0, g.Size)
	degree := make([]int, g.Size)
	for i := 0; i < g.Size; i++ {
		degree[i] = len(g.List[i])
	}
	type pair struct{ v, d int }
	nodes := make([]pair, g.Size)
	for i := 0; i < g.Size; i++ {
		nodes[i] = pair{i, degree[i]}
	}
	used := make([]bool, g.Size)
	for range nodes {
		// Finde Minimum Degree Node
		minDeg := g.Size + 1
		minIdx := -1
		for i, n := range nodes {
			if !used[n.v] && n.d < minDeg {
				minDeg = n.d
				minIdx = i
			}
		}
		if minIdx == -1 {
			break
		}
		v := nodes[minIdx].v
		order = append(order, v)
		used[v] = true
		// Grad der Nachbarn verringern
		for u := range g.List[v] {
			for i := range nodes {
				if nodes[i].v == u {
					nodes[i].d--
				}
			}
		}
	}
	return order
}
