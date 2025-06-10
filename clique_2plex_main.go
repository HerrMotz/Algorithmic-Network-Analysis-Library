package main

import (
    "flag"
    "fmt"
    "os"
    "sort"
    "time"

    "graphlib/graphlib"
)

// largestClique returns the clique with the maximum size from a slice of cliques.
func largestClique(cliques [][]int) []int {
    var max []int
    for _, c := range cliques {
        if len(c) > len(max) {
            max = c
        }
    }
    return max
}

// setToSortedSlice converts a VertexSet to a sorted slice of ints for pretty printing.
func setToSortedSlice(vs graphlib.VertexSet) []int {
    nodes := make([]int, 0, len(vs))
    for v := range vs {
        nodes = append(nodes, v)
    }
    sort.Ints(nodes)
    return nodes
}

// benchmarkBK runs one Bron–Kerbosch variant and prints the result immediately
// so that we can see which variant explodes (stack‑overflow etc.).
func benchmarkBK(name string, run func() [][]int) {
    start := time.Now()
    cliques := run()
    dur := time.Since(start)
    num := len(cliques)
    largest := largestClique(cliques)

    fmt.Printf("%-32s ⇒ Cliques: %d  |  Größte Clique: %d  |  Zeit: %v\n",
        name, num, len(largest), dur)
}

func main() {
    // 0) Gesamtzeit
    allStart := time.Now()

    // 1) CLI
    graphPath := flag.String("graph", "", "Pfad zur Graph-Datei (Adjazenzlistenformat)")
    flag.Parse()
    if *graphPath == "" {
        fmt.Println("Bitte gib den Pfad zum Graphen an mit -graph=...")
        os.Exit(1)
    }

    // 2) Einlesen
    readStart := time.Now()
    g, err := graphlib.ReadGraphAdjList(*graphPath, false)
    readDur := time.Since(readStart)
    if err != nil {
        fmt.Printf("Fehler beim Einlesen: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Graph eingelesen")

    // 3) ── Bron‑Kerbosch‑Tests VOR 2‑Plex ────────────────────────────
    fmt.Println("\n===== Bron–Kerbosch Varianten (laufen nacheinander) =====")
    variants := []struct {
        name string
        run  func() [][]int
    }{
        //{"BronKerboschSimple", func() [][]int { return g.BronKerboschSimple() }},
        {"BronKerboschSimpleIter", func() [][]int { return g.BronKerboschSimpleIter() }},
        {"BronKerboschPivot", func() [][]int { return g.BronKerboschPivot() }},
        {"BronKerboschDegeneracy", func() [][]int { return g.BronKerboschDegeneracy() }},
        {"BronKerboschOwnPivotDegeneracy", func() [][]int { return g.BronKerboschOwnPivotDegeneracy() }},
    }
    for _, v := range variants {
        benchmarkBK(v.name, v.run)
    }

    // 4) ── 2‑Plex erst jetzt ────────────────────────────────────────
    induced := graphlib.InducedGraph{
        AdjList:      g,
        GoodVertices: make(graphlib.VertexSet),
    }
    for i := 0; i < g.Size; i++ {
        induced.GoodVertices[i] = struct{}{}
    }
    pGraph := graphlib.PermanentInducedGraph{
        IGraph:            induced,
        PermanentVertices: make(graphlib.VertexSet),
    }
    fmt.Println("\n===== 2‑Plex Suche =====")
    plexStart := time.Now()
    plex, err := pGraph.TwoPlexPermanentParallel()
    plexDur := time.Since(plexStart)
    if err != nil {
        fmt.Printf("Fehler bei TwoPlexPermanentParallel: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Größter 2‑Plex: Größe %d  |  Zeit: %v\n",
        len(plex), plexDur)

    // 5) ── Zusammenfassung der I/O‑Zeit etc. ───────────────────────
    fmt.Println("\n===== Laufzeiten (I/O) =====")
    fmt.Printf("Einlesen des Graphen:  %v\n", readDur)
    fmt.Printf("Gesamtlaufzeit:       %v\n", time.Since(allStart))
}
