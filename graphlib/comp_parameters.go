package graphlib

import (
    "errors"
    "sort"
)

// Computation of Parameters ÜS2T2

// Density: |A| / |V| * (|V| - 1), |A| = arcCount number of arcs, |V| = number of vertices
func (l *AdjList) Density() (float64, error) {
	if l.Size == 0 {
		return 0, errors.New("graph is empty")
	}
	if l.Size == 1 {
		return 0, nil
	}
	
	// count arcs
	var arcCount int
	for _, neighbours := range l.List {
		arcCount += len(neighbours)
	}

	// max number of arcs = |V| * (|V| - 1)
	maxArcs := l.Size * (l.Size - 1)
	return float64(arcCount) / float64(maxArcs), nil
}

// maximum degree: max(|N(u)|) für alle u
func (l *AdjList) MaxDegree() (int, error) {
    if l.Size == 0 {
        return 0, errors.New("graph is empty")
    }

	// get degrees of all vertices and find the maximum
    max := 0
    for i := 0; i < l.Size; i++ {
        deg := len(l.List[i])
        if deg > max {
            max = deg
        }
    }
    return max, nil
}

// Hindex
func (l *AdjList) HIndex() (int, error) {
    if l.Size == 0 {
        return 0, errors.New("graph is empty")
    }

	// get degrees of all vertices
    degrees := make([]int, 0, l.Size)
    for i := 0; i < l.Size; i++ {
        degrees = append(degrees, len(l.List[i]))
    }

	// sort degrees in descending order
    sort.Sort(sort.Reverse(sort.IntSlice(degrees)))

	// find h-index
    h := 0
    for i, d := range degrees {
        if d >= i+1 {
            h = i + 1
        } else {
            break
        }
    }
    return h, nil
}

func (l *AdjList) Degeneracy() (int, error) {
    // 1) get the degeneracy order σ
    sigma, err := l.DegeneracyOrder()
    if err != nil {
        return 0, err
    }

    // 2) track removed vertices
    removed := make([]bool, l.Size)
    maxDeg := 0

    // 3) for each v in removal order
    for _, v := range sigma {
        u := v.Id
        // count current degree (neighbors not yet removed)
        d := 0
        for w := range l.List[u] {
            if !removed[w] {
                d++
            }
        }
        // update degeneracy
        if d > maxDeg {
            maxDeg = d
        }
        // mark u removed
        removed[u] = true
    }

    return maxDeg, nil
}

// DegeneracyOrder implements the pseudocode from the lecture
func (l *AdjList) DegeneracyOrder() ([]Vertex, error) {
    n := l.Size
    if n == 0 {
        return nil, errors.New("graph is empty")
    }

    // σ ← empty list
    sigma := make([]Vertex, 0, n)

    // D ← array of length n, each an empty list
    D := make([][]int, n)
    for i := 0; i < n; i++ {
        D[i] = []int{}
    }

    // degree array
    deg := make([]int, n)

    // foreach v ∈ V(G)
    for v := 0; v < n; v++ {
        d := len(l.List[v])
        deg[v] = d
        D[d] = append(D[d], v)
    }

    // mark vertices already in σ
    inSigma := make([]bool, n)

    // repeat n times
    for i := 0; i < n; i++ {
        // find smallest j with D[j] non-empty
        j := 0
        for j < n && len(D[j]) == 0 {
            j++
        }
        if j >= n {
            break
        }

        // remove v ← D[j].remove()
        idx := len(D[j]) - 1
        v := D[j][idx]
        D[j] = D[j][:idx]

        // σ.append(v)
        sigma = append(sigma, Vertex{v})
        inSigma[v] = true

        // foreach u ∈ N(v)
        for u := range l.List[v] {
            if inSigma[u] {
                continue
            }
            du := deg[u]

            // remove u from D[du]
            for k, x := range D[du] {
                if x == u {
                    D[du] = append(D[du][:k], D[du][k+1:]...)
                    break
                }
            }

            // du ← du - 1
            deg[u] = du - 1

            // D[du-1].append(u)
            D[du-1] = append(D[du-1], u)
        }
    }
    return sigma, nil
}

// KCoreDecomp: L[k] enthält genau alle v mit deg⁺_D(v)=k
// KCoreDecomp implements the k-core decomposition as per the lecture pseudocode.
// Returns L (Buckets indexed by core level) and D (σ-oriented adjacency).
func (l *AdjList) KCoreDecomp() ([][]Vertex, [][]Vertex, error) {
    n := l.Size
    if n == 0 {
        return nil, nil, errors.New("graph is empty")
    }

    // 1) σ ← DegeneracyOrder(G)
    sigma, err := l.DegeneracyOrder()
    if err != nil {
        return nil, nil, err
    }

    // build rank[v] = position of v in σ
    rank := make([]int, n)
    for i, v := range sigma {
        rank[v.Id] = i
    }

    // 2) D ← directed graph oriented by σ
    D := make([][]Vertex, n)
    for u := 0; u < n; u++ {
        for w := range l.List[u] {
            if rank[u] < rank[w] {
                D[u] = append(D[u], Vertex{Id: w})
            }
        }
    }

    // 3) L ← array of length n for buckets
    L := make([][]Vertex, n)
    k := 0

    // 4–9) for each v in σ, update k and assign to L[k]
    for _, v := range sigma {
        dPlus := len(D[v.Id])
        if dPlus > k {
            k = dPlus
        }
        L[k] = append(L[k], v)
    }

    // 10–12) fill empty buckets with next non-empty
    for i := 0; i < n; i++ {
        if len(L[i]) == 0 {
            j := i + 1
            for j < n && len(L[j]) == 0 {
                j++
            }
            if j < n {
                L[i] = L[j]
            }
        }
    }

    return L, D, nil
}
