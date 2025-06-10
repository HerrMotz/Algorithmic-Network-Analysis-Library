package graphlib

import (
    "fmt"
    "math/rand"
    "sort"
    "time"
)

// ─────────────────────────────────────────────
// Hilfs‑Mengenoperationen
// ─────────────────────────────────────────────
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

// addClique – speichert eine Clique nur, wenn sie noch nicht gesehen wurde
func addClique(res *[][]int, seen map[string]struct{}, R map[int]struct{}) {
    clique := make([]int, 0, len(R))
    for v := range R {
        clique = append(clique, v)
    }
    sort.Ints(clique)              // kanonische Ordnung
    key := fmt.Sprint(clique)      // einfacher String‑Key
    if _, ok := seen[key]; !ok {
        *res = append(*res, clique)
        seen[key] = struct{}{}
    }
}

// ─────────────────────────────────────────────
// Bron‑Kerbosch – Varianten
// ─────────────────────────────────────────────

// BronKerboschSimple – ohne Pivot
func (g *AdjList) BronKerboschSimple() [][]int {
    var res [][]int
    seen := make(map[string]struct{})

    var bk func(R, P, X map[int]struct{})
    bk = func(R, P, X map[int]struct{}) {
        if len(P) == 0 && len(X) == 0 {
            addClique(&res, seen, R)
            return
        }
        for v := range copySet(P) { // über Kopie iterieren
            R2 := copySet(R)
            R2[v] = struct{}{}
            Nv := g.neighSet(v)
            bk(R2, intersect(P, Nv), intersect(X, Nv))
            delete(P, v)
            X[v] = struct{}{}
        }
    }

    P := make(map[int]struct{}, g.Size)
    for v := 0; v < g.Size; v++ {
        P[v] = struct{}{}
    }
    bk(make(map[int]struct{}), P, make(map[int]struct{}))
    return res
}

// BronKerboschSimpleIter – pivotlose Iteration ohne Rekursion
func (g *AdjList) BronKerboschSimpleIter() [][]int {
    // --- 0) Nachbarschaften einmal puffern
    nbr := make([]map[int]struct{}, g.Size)
    for i := 0; i < g.Size; i++ {
        nbr[i] = g.neighSet(i)
    }

    // --- 1) eigener Stack-Frame-Typ
    type frame struct {
        R        []int
        P, X     map[int]struct{}
        toVisit  []int   // verbleibende Kandidaten
    }

    // --- 2) Start-Frame erzeugen
    initP := make(map[int]struct{}, g.Size)
    for v := 0; v < g.Size; v++ {
        initP[v] = struct{}{}
    }
    stack := []frame{{R: nil, P: initP, X: make(map[int]struct{}), toVisit: nil}}

    var cliques [][]int
    seen := make(map[string]struct{})

    // --- 3) Tiefensuche per while-Schleife
    for len(stack) > 0 {
        f := &stack[len(stack)-1]

        // 3a) Initialisieren (nur beim ersten Betreten)
        if f.toVisit == nil {
            if len(f.P) == 0 && len(f.X) == 0 {
                // maximale Clique gefunden
                c := append([]int(nil), f.R...)
                sort.Ints(c)
                key := fmt.Sprint(c)
                if _, ok := seen[key]; !ok {
                    cliques = append(cliques, c)
                    seen[key] = struct{}{}
                }
            }
            if len(f.P) == 0 {
                // fertig – Frame von Stack poppen
                stack = stack[:len(stack)-1]
                continue
            }
            f.toVisit = make([]int, 0, len(f.P))
            for v := range f.P { f.toVisit = append(f.toVisit, v) }
        }

        // 3b) Keine Kandidaten mehr ⇒ Frame fertig
        if len(f.toVisit) == 0 {
            stack = stack[:len(stack)-1]
            continue
        }

        // 3c) Nächsten Kandidaten auswählen
        v := f.toVisit[len(f.toVisit)-1]
        f.toVisit = f.toVisit[:len(f.toVisit)-1]
        delete(f.P, v)

        // Neues R, P, X bilden
        Nv   := nbr[v]
        Pcap := intersect(f.P, Nv)
        Xcap := intersect(f.X, Nv)
        R2   := append(append(make([]int, 0, len(f.R)+1), f.R...), v)

        // Kind-Frame pushen
        stack = append(stack, frame{R: R2, P: Pcap, X: Xcap})

        // v nach X verschieben (entspricht Rückkehrschritt)
        f.X[v] = struct{}{}
    }
    return cliques
}


// BronKerboschPivot – Tomita‑Pivotregel
// BronKerboschPivot – Tomita-Pivotregel, jetzt ohne Seiteneffekte auf P/X
func (g *AdjList) BronKerboschPivot() [][]int {
    var res [][]int
    seen := make(map[string]struct{})

    var bk func(R, P, X map[int]struct{})
    bk = func(R, P, X map[int]struct{}) {
        if len(P) == 0 && len(X) == 0 {
            addClique(&res, seen, R)
            return
        }

        // --- Pivot u aus P ∪ X mit maximaler Schnittmenge P∩N(u)
        U := union(P, X)
        var pivot, maxDeg int
        first := true
        for u := range U {
            deg := len(intersect(P, g.neighSet(u)))
            if first || deg > maxDeg {
                pivot, maxDeg, first = u, deg, false
            }
        }

        // --- für alle v ∈ P \ N(pivot)
        for v := range difference(P, g.neighSet(pivot)) {

            // (1) neues R := R ∪ {v}
            R2 := copySet(R)
            R2[v] = struct{}{}

            // (2) frische Kandidaten- und Exklusionsmengen
            Nv  := g.neighSet(v)
            P2  := intersect(P, Nv) // NEU: eigene Kopien
            X2  := intersect(X, Nv)

            // (3) Rekursion
            bk(R2, P2, X2)

            // (4) v aus P entfernen und nach X verschieben
            delete(P, v)
            X[v] = struct{}{}
        }
    }

    // --- Initialaufruf
    P0 := make(map[int]struct{}, g.Size)
    for v := 0; v < g.Size; v++ {
        P0[v] = struct{}{}
    }
    bk(make(map[int]struct{}), P0, make(map[int]struct{}))
    return res
}


// BronKerboschDegeneracy – Pivot + Degeneracy‑Ordering
func (g *AdjList) BronKerboschDegeneracy() [][]int {
    order := g.degeneracyOrder()
    var res [][]int
    seenCliques := make(map[string]struct{})
    seenVertices := make(map[int]struct{})

    for _, v := range order {
        R := map[int]struct{}{v: {}}
        Nv := g.neighSet(v)
        P := difference(Nv, seenVertices)
        X := intersect(Nv, seenVertices)
        sub := g.bkPivotDegeneracy(R, P, X, seenCliques)
        res = append(res, sub...)
        seenVertices[v] = struct{}{}
    }
    return res
}

func (g *AdjList) bkPivotDegeneracy(R, P, X map[int]struct{}, seen map[string]struct{}) [][]int {
    var res [][]int
    var bk func(R, P, X map[int]struct{})
    bk = func(R, P, X map[int]struct{}) {
        if len(P) == 0 && len(X) == 0 {
            addClique(&res, seen, R)
            return
        }
        U := union(P, X)
        var pivot, maxNbr int
        first := true
        for u := range U {
            deg := len(intersect(P, g.neighSet(u)))
            if first || deg > maxNbr {
                pivot, maxNbr, first = u, deg, false
            }
        }

        PwoN := difference(P, g.neighSet(pivot))
        for v := range copySet(PwoN) {
            R2 := copySet(R)
            R2[v] = struct{}{}
            Nv := g.neighSet(v)
            bk(R2, intersect(P, Nv), intersect(X, Nv))
            delete(P, v)
            X[v] = struct{}{}
        }
    }
    bk(copySet(R), copySet(P), copySet(X))
    return res
}

// BronKerboschOwnPivotDegeneracy – randomisiertes Pivot
func (g *AdjList) BronKerboschOwnPivotDegeneracy() [][]int {
    order := g.degeneracyOrder()
    seenVertices := make(map[int]struct{})
    seenCliques := make(map[string]struct{})
    var out [][]int

    rand.Seed(time.Now().UnixNano())

    var bk func(R, P, X map[int]struct{})
    bk = func(R, P, X map[int]struct{}) {
        if len(P) == 0 && len(X) == 0 {
            addClique(&out, seenCliques, R)
            return
        }
        U := union(P, X)
        total := 0
        weights := make(map[int]int, len(U))
        for u := range U {
            w := len(intersect(P, g.neighSet(u)))
            if w == 0 {
                w = 1
            }
            weights[u] = w
            total += w
        }
        rnd := rand.Intn(total)
        run := 0
        pivot := -1
        for u, w := range weights {
            run += w
            if rnd < run {
                pivot = u
                break
            }
        }
        PwoN := difference(P, g.neighSet(pivot))
        for v := range copySet(PwoN) {
            R2 := copySet(R)
            R2[v] = struct{}{}
            Nv := g.neighSet(v)
            bk(R2, intersect(P, Nv), intersect(X, Nv))
            delete(P, v)
            X[v] = struct{}{}
        }
    }

    for _, v := range order {
        R := map[int]struct{}{v: {}}
        Nv := g.neighSet(v)
        P := difference(Nv, seenVertices)
        X := intersect(Nv, seenVertices)
        bk(R, P, X)
        seenVertices[v] = struct{}{}
    }
    return out
}

// ─────────────────────────────────────────────
// Nachbarn als Menge
// ─────────────────────────────────────────────
func (g *AdjList) neighSet(u int) map[int]struct{} {
    n, _ := g.Neighbours(Vertex{Id: u})
    res := make(map[int]struct{}, len(n))
    for v := range n {
        if v != u {               // ← Self-Loop unterdrücken
            res[v] = struct{}{}
        }
    }
    return res
}

// MakeUndirected stellt sicher, dass die Adjazenzliste symmetrisch ist.
func (l *AdjList) MakeUndirected() {
    for u, nbrs := range l.List {
        for v := range nbrs {
            if l.List[v] == nil {
                l.List[v] = make(VertexSet)
            }
            l.List[v][u] = Naught{}
        }
    }
}

// Degeneracy‑Ordering (einfache O(n²)‑Implementierung)
func (g *AdjList) degeneracyOrder() []int {
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
        minDeg := g.Size + 1
        minIdx := -1
        for i, n := range nodes {
            if !used[n.v] && n.d < minDeg {
                minDeg, minIdx = n.d, i
            }
        }
        if minIdx == -1 {
            break
        }
        v := nodes[minIdx].v
        order = append(order, v)
        used[v] = true
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
