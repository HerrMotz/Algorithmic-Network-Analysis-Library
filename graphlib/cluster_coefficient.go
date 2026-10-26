package graphlib

import (
	"errors"
)

// local clustering coefficient
// (from lecture:) C(v) = |{{u,w} ⊆ N(v) : {u,w} ∈ E}| / (deg(v) über 2).
func (l *AdjList) LocalClustering(v Vertex) (float64, error) {
	if !l.HasVertex(v) {
		return 0, errors.New("vertex does not exist")
	}

	// neighbours of v as slices
	neighMap, _ := l.Neighbours(v)
	k := len(neighMap)
	if k < 2 {
		return 0, nil
	}

	neighs := make([]Vertex, 0, k)
	for uID := range neighMap {
		neighs = append(neighs, Vertex{uID})
	}

	// count edges between neighbours
	var connected int
	for i := 0; i < k; i++ {
		for j := i + 1; j < k; j++ {
			u, w := neighs[i], neighs[j]
			if l.HasEdge(Edge{U: u, V: w}) {
				connected++
			}
		}
	}

	possible := float64(k*(k-1)) / 2.0
	return float64(connected) / possible, nil
}

// global clustering coefficient using triangles (lecture)
// c(G) = 6·#triangles / # length-2 paths,
// where length-2 paths = ∑ₖ deg(k)·(deg(k)−1).
func (l *AdjList) GlobalClustering() (float64, error) {
	// 1) count all undirected triangles
	tris, err := l.TrianglesCNI()
	if err != nil {
		return 0, err
	}
	numTris := len(tris)

	// 2) count all 2-paths: ∑ₖ deg(k)*(deg(k)-1)
	var numPaths int
	for v := 0; v < l.Size; v++ {
		d := len(l.List[v])
		numPaths += d * (d - 1) // /2 bei ungerichteten?
	}

	if numPaths == 0 {
		return 0, nil
	}
	return 6 * float64(numTris) / float64(numPaths), nil
}
